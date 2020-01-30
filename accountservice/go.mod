module github.com/callistaenterprise/goblog/accountservice

go 1.13

// replace (
// 	github.com/callistaenterprise/goblog => ../
// 	github.com/callistaenterprise/goblog/accountservice => ./
// )

require (
	github.com/boltdb/bolt v1.3.1
	github.com/gorilla/mux v1.7.3
	golang.org/x/sys v0.0.0-20200124204421-9fbb57f87de9 // indirect
)
