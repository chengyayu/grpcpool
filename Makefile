.PHONY: proto
proto:
	protoc --proto_path=./example --go_out=paths=source_relative:./example --go-grpc_out=paths=source_relative:./example ./example/pb/hello.proto

.PHONY: benchmark
benchmark: benchmarkOnlyOneRPC benchmarkSingleRPC benchmarkPoolRPC

.PHONY: benchmarkOnlyOneRPC benchmarkSingleRPC benchmarkPoolRPC
benchmarkOnlyOneRPC:
	go test -bench='BenchmarkOnlyOneRPC' -benchtim=5000x -count=3 -benchmem .
benchmarkSingleRPC:
	go test -bench='BenchmarkSingleRPC' -benchtime=5000x -count=3 -benchmem .
benchmarkPoolRPC:
	go test -bench='BenchmarkPoolRPC' -benchtime=5000x -count=3 -benchmem .


