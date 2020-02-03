module github.com/callistaenterprise/goblog/vipservice

go 1.13

replace github.com/callistaenterprise/goblog/common => ../common

require (
	github.com/callistaenterprise/goblog/common v0.0.0-00010101000000-000000000000
	github.com/gorilla/mux v1.7.3
	github.com/spf13/viper v1.6.2
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
)
