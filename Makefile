.PHONY: proto
proto:
	protoc --proto_path=./example/single --go_out=paths=source_relative:./example/single --go-grpc_out=paths=source_relative:./example/single ./example/single/pb/hello.proto

.PHONY: build
build: proto
	go build -o ./example/single/client/client  ./example/single/client/main.go && \
	go build -o ./example/single/server/server ./example/single/server/main.go

.PHONY: run
run: build
	./example/single/server/server

.PHONY: benchmark
benchmark: benchmarkOnlyOneRPC benchmarkSingleRPC benchmarkPoolRPC

.PHONY: benchmarkOnlyOneRPC benchmarkSingleRPC benchmarkPoolRPC
benchmarkOnlyOneRPC:
	go test -run=none -parallel=2 -bench="^BenchmarkOnlyOneRPC" -benchtime=5000x -count=3 -benchmem
benchmarkSingleRPC:
	go test -run=none -parallel=2 -bench="^BenchmarkSingleRPC" -benchtime=5000x -count=3 -benchmem
benchmarkPoolRPC:
	go test -run=none -parallel=2 -bench="^BenchmarkPoolRPC" -benchtime=5000x -count=3 -benchmem


