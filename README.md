# grpcpool

## 特性

- 池化对象为逻辑连接，本质是逻辑连接池。
- 支持应用层自定义参数。
    - dial func(address string) (*grpc.ClientConn, error) 创建连接的函数，同时支持配置 grpc 连接自定义参数。
    - maxIdle int 连接池内最大空闲（物理）连接数。默认初始化数量与之相同。
    - maxActive int 连接池内最大活跃（物理）连接数。0 表示无限制。
    - maxConcurrentStreams int 每个物理连接内支持的最大并发流数。
    - reuse bool 如果 maxActive 已达上限，继续获取连接时，是否继续使用池内连接。否：会创建一个一次性连接（用完即销毁）返回。
- 根据参数自动扩、缩容。
- 根据参数执行池满后获取连接的策略。

## 基准测试

1. 每轮并发请求共用一个连接池，每次请求从池获取一个连接：

```shell
go test -run=none -parallel=2 -bench="^BenchmarkPoolRPC" -benchtime=5000x -count=3 -benchmem
goos: linux
goarch: amd64
pkg: github.com/chengyayu/grpcpool
cpu: AMD Ryzen 7 3700U with Radeon Vega Mobile Gfx  
BenchmarkPoolRPC-8          5000           4657010 ns/op         8404197 B/op        109 allocs/op
BenchmarkPoolRPC-8          5000           4654517 ns/op         8404151 B/op        109 allocs/op
BenchmarkPoolRPC-8          5000           4642664 ns/op         8404177 B/op        109 allocs/op
PASS
ok      github.com/chengyayu/grpcpool   69.855s
```

2. 每个请求创建一个新连接：

```shell
go test -run=none -parallel=2 -bench="^BenchmarkSingleRPC" -benchtime=5000x -count=3 -benchmem
goos: linux
goarch: amd64
pkg: github.com/chengyayu/grpcpool
cpu: AMD Ryzen 7 3700U with Radeon Vega Mobile Gfx  
BenchmarkSingleRPC-8        5000           5077726 ns/op         8523635 B/op        691 allocs/op
BenchmarkSingleRPC-8        5000           5094132 ns/op         8523640 B/op        691 allocs/op
BenchmarkSingleRPC-8        5000           5307944 ns/op         8523638 B/op        691 allocs/op
PASS
ok      github.com/chengyayu/grpcpool   77.499s
```

3. 每轮并发请求共用一个连接：

```shell
go test -run=none -parallel=2 -bench="^BenchmarkOnlyOneRPC" -benchtime=5000x -count=3 -benchmem
goos: linux
goarch: amd64
pkg: github.com/chengyayu/grpcpool
cpu: AMD Ryzen 7 3700U with Radeon Vega Mobile Gfx  
BenchmarkOnlyOneRPC-8               5000           6050398 ns/op         8403255 B/op        106 allocs/op
BenchmarkOnlyOneRPC-8               5000           5979344 ns/op         8403247 B/op        106 allocs/op
BenchmarkOnlyOneRPC-8               5000           6037154 ns/op         8403248 B/op        106 allocs/op
PASS
ok      github.com/chengyayu/grpcpool   90.410s
```
## 负载均衡

虽然 GRPC 负载均衡不在本库解决范围之内，但是由于 K8S+GRPC 组合的广泛应用，且由于众所周知的原因，K8S service 无法对 GRPC 请求进行负载。
`example/k8slb` 也给出了一个基于 [kuberesolver](https://github.com/sercand/kuberesolver) 的客户端负载方案 Demo，仅供参考。

## 特别感谢

本库基于 [github.com/shimingyah/pool](https://github.com/shimingyah/pool) 修改而成。

## 参考资料

- [https://github.com/grpc/grpc-go](https://github.com/grpc/grpc-go)
- [需要每次检查连接状态吗？stackoverflow.com](https://stackoverflow.com/questions/64484690/grpc-cpp-how-can-i-check-if-the-rpc-channel-connected-successfully)
- [GRPC doc 连接状态变化](https://grpc.github.io/grpc/core/md_doc_connectivity-semantics-and-api.html)
- [GRPC doc 负载均衡机制](https://github.com/grpc/grpc/blob/master/doc/load-balancing.md)
- [GRPC 在 K8S 中的负载均衡](https://www.cnblogs.com/BlueMountain-HaggenDazs/p/17071303.html#kuberesolver)
