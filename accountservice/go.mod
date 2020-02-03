module github.com/callistaenterprise/goblog/accountservice

go 1.13

// replace (
// 	github.com/callistaenterprise/goblog => ../
// 	github.com/callistaenterprise/goblog/accountservice => ./
// )

require (
	github.com/boltdb/bolt v1.3.1
	github.com/gorilla/mux v1.7.3
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/viper v1.6.2
	github.com/stretchr/testify v1.4.0
	golang.org/x/sys v0.0.0-20200124204421-9fbb57f87de9 // indirect
	gopkg.in/h2non/gock.v1 v1.0.15
)
