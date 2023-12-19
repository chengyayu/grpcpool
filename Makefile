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


