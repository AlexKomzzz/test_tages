protoc:
	protoc -I=grpc_api/proto --go_out=. --go-grpc_out=. grpc_api/proto/app.proto

rm:
	rm ./pkg/api/*.go

gorun:
	go run ./cmd/server/main.go