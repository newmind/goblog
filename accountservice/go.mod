module github.com/callistaenterprise/goblog/accountservice

go 1.13

replace (
	github.com/callistaenterprise/goblog => ../
	github.com/callistaenterprise/goblog/accountservice => ./
)

require github.com/gorilla/mux v1.7.3
