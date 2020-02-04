package circuitbreaker

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/eapache/go-resiliency/retrier"
	"github.com/sirupsen/logrus"
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
