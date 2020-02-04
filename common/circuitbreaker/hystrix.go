package circuitbreaker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/callistaenterprise/goblog/common/messaging"
	"github.com/callistaenterprise/goblog/common/util"
	"github.com/eapache/go-resiliency/retrier"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

var Client http.Client

var RETRIES = 3

func CallUsingCircuitBreaker(breakerName string, url string, method string) ([]byte, error) {
	output := make(chan []byte, 1) // Declare the channel where the hystrix goroutine will put success responses.

	errors := hystrix.Go(breakerName, // Pass the name of the circuit breaker as first parameter.

		// 2nd parameter, the inlined func to run inside the breaker.
		func() error {
			// Create the request. Omitted err handling for brevity
			req, _ := http.NewRequest(method, url, nil)

			// For hystrix, forward the err from the retrier. It's nil if successful.
			return callWithRetries(req, output)
		},

		// 3rd parameter, the fallback func. In this case, we just do a bit of logging and return the error.
		func(err error) error {
			logrus.Errorf("In fallback function for breaker %v, error: %v", breakerName, err.Error())
			circuit, _, _ := hystrix.GetCircuit(breakerName)
			logrus.Errorf("Circuit state is: %v", circuit.IsOpen())
			return err
		})

	// Response and error handling. If the call was successful, the output channel gets the response. Otherwise,
	// the errors channel gives us the error.
	select {
	case out := <-output:
		logrus.Debugf("Call in breaker %v successful", breakerName)
		return out, nil

	case err := <-errors:
		return nil, err
	}
}

func callWithRetries(req *http.Request, output chan []byte) error {

	r := retrier.New(retrier.ConstantBackoff(RETRIES, 100*time.Millisecond), nil)
	attempt := 0
	err := r.Run(func() error {
		attempt++
		resp, err := Client.Do(req)
		if err == nil && resp.StatusCode < 299 {
			responseBody, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				output <- responseBody
				return nil
			}
			return err
		} else if err == nil {
			err = fmt.Errorf("Status was %v", resp.StatusCode)
		}

		logrus.Errorf("Retrier failed, attempt %v", attempt)

		return err
	})
	return err
}

func ConfigureHystrix(commands []string, amqpClient messaging.IMessagingClient) {

	for _, command := range commands {
		hystrix.ConfigureCommand(command, hystrix.CommandConfig{
			Timeout:                resolveProperty(command, "Timeout"),
			MaxConcurrentRequests:  resolveProperty(command, "MaxConcurrentRequests"),
			ErrorPercentThreshold:  resolveProperty(command, "ErrorPercentThreshold"),
			RequestVolumeThreshold: resolveProperty(command, "RequestVolumeThreshold"),
			SleepWindow:            resolveProperty(command, "SleepWindow"),
		})
		logrus.Printf("Circuit %v settings: %v", command, hystrix.GetCircuitSettings()[command])
	}

	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	go http.ListenAndServe(net.JoinHostPort("", "8181"), hystrixStreamHandler)
	logrus.Infoln("Launched hystrixStreamHandler at 8181")

	// Publish presence on RabbitMQ
	publishDiscoveryToken(amqpClient)
}

func Deregister(amqpClient messaging.IMessagingClient) {
	ip, err := util.ResolveIpFromHostsFile()
	if err != nil {
		ip = util.GetIPWithPrefix("10.0.")
	}
	token := DiscoveryToken{
		State:   "DOWN",
		Address: ip,
	}
	bytes, _ := json.Marshal(token)
	amqpClient.PublishOnQueue(bytes, "discovery")
}

func publishDiscoveryToken(amqpClient messaging.IMessagingClient) {
	ip, err := util.ResolveIpFromHostsFile()
	if err != nil {
		ip = util.GetIPWithPrefix("10.0.")
	}
	token := DiscoveryToken{
		State:   "UP",
		Address: ip,
	}
	bytes, _ := json.Marshal(token)
	go func() {
		for {
			amqpClient.PublishOnQueue(bytes, "discovery")
			amqpClient.PublishOnQueue(bytes, "discovery")
			time.Sleep(time.Second * 30)
		}
	}()
}

func resolveProperty(command string, prop string) int {
	if viper.IsSet("hystrix.command." + command + "." + prop) {
		return viper.GetInt("hystrix.command." + command + "." + prop)
	} else {
		return getDefaultHystrixConfigPropertyValue(prop)
	}
}
func getDefaultHystrixConfigPropertyValue(prop string) int {
	switch prop {
	case "Timeout":
		return hystrix.DefaultTimeout
	case "MaxConcurrentRequests":
		return hystrix.DefaultMaxConcurrent
	case "RequestVolumeThreshold":
		return hystrix.DefaultVolumeThreshold
	case "SleepWindow":
		return hystrix.DefaultSleepWindow
	case "ErrorPercentThreshold":
		return hystrix.DefaultErrorPercentThreshold
	}
	panic("Got unknown hystrix property: " + prop + ". Panicing!")
}

type DiscoveryToken struct {
	State   string `json:"state"` // UP, RUNNING, DOWN ??
	Address string `json:"address"`
}
