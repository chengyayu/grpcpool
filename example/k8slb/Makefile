.PHONY: proto
proto:
	protoc --proto_path=./pb --go_out=paths=source_relative:./pb --go-grpc_out=paths=source_relative:./pb ./pb/hello.proto

.PHONY: build deploy
build: proto
	go build -o ./client/client  ./client/main.go && \
	go build -o ./server/server ./server/main.go && \
	docker buildx build -t grpc-client:v0.1 -f ./client/Dockerfile --load . && \
	docker buildx build -t grpc-server:v0.1 -f ./server/Dockerfile --load .

deploy:
	kubectl apply -f ./rbac/service-account.yaml && \
	kubectl apply -f ./rbac/role.yaml && \
	kubectl apply -f ./rbac/role_binding.yaml && \
	kubectl apply -f ./server/test-server.yaml && \
	kubectl apply -f ./client/test-client.yaml