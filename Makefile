download:
	go mod download

build: lint download
	go build

lint:
	golangci-lint run