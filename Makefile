compile_broker_proto_file:
	protoc --go_out=. --go_opt=paths=source_relative \
 	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
 	api/proto/broker.proto