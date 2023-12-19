.PHONY: proto
proto:
	protoc --proto_path=./example --go_out=paths=source_relative:./example --go-grpc_out=paths=source_relative:./example ./example/pb/hello.proto

.PHONY: benchmark
benchmark: benchmarkOnlyOneRPC benchmarkSingleRPC benchmarkPoolRPC

.PHONY: benchmarkOnlyOneRPC benchmarkSingleRPC benchmarkPoolRPC
benchmarkOnlyOneRPC:
	go test -bench BenchmarkOnlyOneRPC -benchtime=500x -count=3 -benchmem .

benchmarkSingleRPC:
	go test -bench BenchmarkSingleRPC -benchtime=500x -count=3 -benchmem .

benchmarkPoolRPC:
	go test -bench BenchmarkPoolRPC -benchtime=500x -count=3 -benchmem .
	#go test -bench BenchmarkPoolRPC -run onoe -benchmem -cpuprofile cpuprofile.out -memprofile memprofile.out

