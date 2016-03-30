.PHONY: gen fmt

gen:
	protoc  --go_out=plugins=grpc:. proto/helloworld.proto

fmt:
	gofmt -w -l .
