executable := accountservice-linux-amd64

$(executable):
	@echo Building $(executable)
	GO111MODULE=on go build -o bin/$@ cmd/$@/main.go

build:
	@echo Building $(executable)
	GOOS=linux GO111MODULE=on CGO_ENABLED=0 go build -o $(executable)
	# docker build -t someprefix/$(executable) -f Dockerfile .
