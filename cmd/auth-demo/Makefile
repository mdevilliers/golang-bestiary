.PHONY: gen rotate_auth_certs fmt

gen:
	protoc  --go_out=plugins=grpc:. proto/helloworld.proto

fmt:
	gofmt -w -l .

rotate_auth_certs:
	# generate signing key for JWT tokens
	openssl genrsa -out keys/auth.rsa 2048
	openssl rsa -in keys/auth.rsa -pubout > keys/auth.rsa.pub
	