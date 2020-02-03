package main // import "github.com/callistaenterprise/goblog/vipservice"

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/callistaenterprise/goblog/common/config"
	"github.com/callistaenterprise/goblog/common/messaging"
	"github.com/callistaenterprise/goblog/vipservice/service"
	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

var appName = "vipservice"

var messagingClient messaging.IMessagingClient

func init() {
	configServerUrl := flag.String("configServerUrl", "http://192.168.6.190:8888", "Address to config server")
	profile := flag.String("profile", "dev", "Environment profile, something similar to spring profiles")
	configBranch := flag.String("configBranch", "P9", "git branch to fetch configuration from")
	flag.Parse()

	viper.Set("profile", *profile)
	viper.Set("configServerUrl", *configServerUrl)
	viper.Set("configBranch", *configBranch)
}

func main() {
	logrus.Infoln("Starting " + appName + "...")

	config.LoadConfigurationFromBranch(viper.GetString("configServerUrl"), appName, viper.GetString("profile"), viper.GetString("configBranch"))
	initializeMessaging()

	// Makes sure connection is closed when service exits.
	handleSigterm(func() {
		if messagingClient != nil {
			messagingClient.Close()
		}
	})
	service.StartWebServer(viper.GetString("server_port"))
}

// The callback function that's invoked whenever we get a message on the "vipQueue"
func onMessage(delivery amqp.Delivery) {
	logrus.Infof("Got a message: %v\n", string(delivery.Body))
}

func initializeMessaging() {
	if !viper.IsSet("broker_url") {
		panic("No 'broker_url' set in configuration, cannot start")
	}
	messagingClient = &messaging.MessagingClient{}
	messagingClient.ConnectToBroker(viper.GetString("broker_url"))

	// Call the subscribe method with queue name and callback function
	err := messagingClient.SubscribeToQueue("vip_queue", appName, onMessage)
	failOnError(err, "Could not start subscribe to vip_queue")

	err = messagingClient.Subscribe(viper.GetString("config_event_bus"), "topic", appName, config.HandleRefreshEvent)
	failOnError(err, "Could not start subscribe to "+viper.GetString("config_event_bus")+" topic")
}

// Handles Ctrl+C or most other means of "controlled" shutdown gracefully. Invokes the supplied func before exiting.
func handleSigterm(handleExit func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		handleExit()
		os.Exit(1)
	}()
}

func failOnError(err error, msg string) {
	if err != nil {
		logrus.Infof("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
