# grpcpool

修改自 [github.com/shimingyah/pool](https://github.com/shimingyah/pool)

## 特性

- 池化对象为逻辑连接，本质是逻辑连接（stream）池。
- 支持应用层自定义参数。
    - dial func(address string) (*grpc.ClientConn, error) 创建连接的函数，同时支持配置 grpc 连接自定义参数。
    - maxIdle int 连接池内最大空闲（物理）连接数。默认初始化数量与之相同。
    - maxActive int 连接池内最大活跃（物理）连接数。0 表示无限制。
    - maxConcurrentStreams int 每个物理连接内支持的最大并发流数。
    - reuse bool 如果 maxActive 已达上限，继续获取连接时，是否继续使用池内连接。否：会创建一个一次性连接（用完即销毁）返回。
- 根据参数自动扩、缩容。
- 池满后获取连接的策略。

## 基准测试

1. 预先初始化一个连接池，每次请求从池获取一个连接。

```shell
go test -bench='BenchmarkPoolRPC' -benchtime=5000x -count=3 -benchmem .
goos: linux
goarch: amd64
pkg: github.com/chengyayu/grpcpool
cpu: AMD Ryzen 7 3700U with Radeon Vega Mobile Gfx  
BenchmarkPoolRPC-8          5000           4895632 ns/op         8404161 B/op        109 allocs/op
BenchmarkPoolRPC-8          5000           4903488 ns/op         8404146 B/op        109 allocs/op
BenchmarkPoolRPC-8          5000           5306860 ns/op         8404156 B/op        109 allocs/op
PASS
ok      github.com/chengyayu/grpcpool   75.850s
```

2. 每次请求创建一个新连接。

```shell
go test -bench='BenchmarkSingleRPC' -benchtime=5000x -count=3 -benchmem .
goos: linux
goarch: amd64
pkg: github.com/chengyayu/grpcpool
cpu: AMD Ryzen 7 3700U with Radeon Vega Mobile Gfx  
BenchmarkSingleRPC-8        5000           6532750 ns/op         8519701 B/op        587 allocs/op
BenchmarkSingleRPC-8        5000           6114239 ns/op         8519594 B/op        594 allocs/op
BenchmarkSingleRPC-8        5000           6623537 ns/op         8519824 B/op        605 allocs/op
PASS
ok      github.com/chengyayu/grpcpool   97.918s
```

3. 全局公用一个连接

```shell
go test -bench='BenchmarkOnlyOneRPC' -benchtime=5000x -count=3 -benchmem .
goos: linux
goarch: amd64
pkg: github.com/chengyayu/grpcpool
cpu: AMD Ryzen 7 3700U with Radeon Vega Mobile Gfx  
BenchmarkOnlyOneRPC-8               5000           6984967 ns/op         8403239 B/op        106 allocs/op
BenchmarkOnlyOneRPC-8               5000           6873981 ns/op         8403239 B/op        106 allocs/op
BenchmarkOnlyOneRPC-8               5000           6604527 ns/op         8403238 B/op        106 allocs/op
PASS
ok      github.com/chengyayu/grpcpool   102.623s
```
