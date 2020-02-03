package main // import "github.com/callistaenterprise/goblog/accountservice"

import (
	"flag"
	"fmt"

	"github.com/callistaenterprise/goblog/accountservice/config"
	"github.com/callistaenterprise/goblog/accountservice/dbclient"
	"github.com/callistaenterprise/goblog/accountservice/service"

	"github.com/spf13/viper"
)

var appName = "accountservice"

// Init function, runs before main()
func init() {
	// Read command line flags
	profile := flag.String("profile", "test", "Environment profile, something similar to spring profiles")
	configServerUrl := flag.String("configServerUrl", "http://configserver:8888", "Address to config server")
	configBranch := flag.String("configBranch", "master", "git branch to fetch configuration from")
	flag.Parse()

	// Pass the flag values into viper.
	viper.Set("profile", *profile)
	viper.Set("configServerUrl", *configServerUrl)
	viper.Set("configBranch", *configBranch)
}

func main() {
	fmt.Printf("Starting %v\n", appName)

	config.LoadConfigurationFromBranch(
		viper.GetString("configServerUrl"),
		appName,
		viper.GetString("profile"),
		viper.GetString("configBranch"))

	initializeBoltClient() // NEW
	go config.StartListener(appName, viper.GetString("amqp_server_url"), viper.GetString("config_event_bus"))
	service.StartWebServer(viper.GetString("server_port"))
}

// Creates instance and calls the OpenBoltDb and Seed funcs
func initializeBoltClient() {
	service.DBClient = &dbclient.BoltClient{}
	service.DBClient.OpenBoltDb()
	service.DBClient.Seed()
}
