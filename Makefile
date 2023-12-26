.PHONY: proto
proto:
	protoc --proto_path=./example --go_out=paths=source_relative:./example --go-grpc_out=paths=source_relative:./example ./example/pb/hello.proto

.PHONY: benchmark
benchmark: benchmarkOnlyOneRPC benchmarkSingleRPC benchmarkPoolRPC

.PHONY: benchmarkOnlyOneRPC benchmarkSingleRPC benchmarkPoolRPC
benchmarkOnlyOneRPC:
	go test -run=none -parallel=2 -bench="^BenchmarkOnlyOneRPC" -benchtime=5000x -count=3 -benchmem
benchmarkSingleRPC:
	go test -run=none -parallel=2 -bench="^BenchmarkSingleRPC" -benchtime=5000x -count=3 -benchmem
benchmarkPoolRPC:
	go test -run=none -parallel=2 -bench="^BenchmarkPoolRPC" -benchtime=5000x -count=3 -benchmem


.PHONY: build deploy
build:
	go build -o ./example/client/client  ./example/client/main.go && \
	go build -o ./example/server/server ./example/server/main.go && \
	docker buildx build -t grpc-client:v1.1510 -f ./example/client/Dockerfile --load . && \
	docker buildx build -t grpc-server:v1.1452 -f ./example/server/Dockerfile --load .

deploy:
	kubectl apply -f ./example/rbac/service-account.yaml && \
	kubectl apply -f ./example/rbac/role.yaml && \
	kubectl apply -f ./example/rbac/role_binding.yaml && \
	kubectl apply -f ./example/server/test-server.yaml && \
	kubectl apply -f ./example/client/test-client.yaml