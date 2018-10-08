## meshx

[meshx](https://github.com/smallnest/meshx)是一个通用的微服务组件，为rpc服务提供无侵入的服务治理的功能，
它相当于service mesh中的sidecar组件，但是相当的简单易用。

- 简单,无配置文件、无需kubernetes
- 不依赖第三方组件。(服务发现可选etcd、zookeeper、consul,或者不使用这些应用)
- 支持go的标准rpc、rpcx、grpc、dubbo、motan、thrift等rpc框架
- 性能优异，底层通讯使用高性能的rpcx框架


**注意**: 这个项目正在孵化之中，还没有实现正式的功能。