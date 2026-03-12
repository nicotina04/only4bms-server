.PHONY: build run clean build-arm

BINARY=only4bms-server

build:
	go build -o $(BINARY) .

run: build
	./$(BINARY)

build-arm:
	GOOS=linux GOARCH=arm64 go build -o $(BINARY)-arm64 .

clean:
	rm -f $(BINARY) $(BINARY)-arm64
