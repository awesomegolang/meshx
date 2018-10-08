package adapter

import (
	"github.com/smallnest/meshx/util"
	"github.com/smallnest/rpcx/log"
)

var (
	adapters = make(map[string]NewFuncForAdapter)
)

// NewFuncForAdapter is the adapter factory.
type NewFuncForAdapter func(util.Option) Adapter

// Register registers adapters such as grpc, rpcx, dubbo, motan, thrift.
func Register(name string, adapter NewFuncForAdapter) {
	if _, ok := adapters[name]; ok {
		log.Infof("client adapter '%v' has alrealy registered, the old adapter will be replaced", name)
	}
	adapters[name] = adapter
}

// NewAdapter creates an adapter.
func NewAdapter(name string, op util.Option) Adapter {
	fn, ok := adapters[name]
	if !ok {
		log.Fatalf("no client adapter registered for '%v'", name)
	}
	return fn(op)
}

// Adapter is the client agent interface of sidecar which implements protocol of rpc frameworks.
// It is used to listen requests from the local rpc clients.
type Adapter interface {
	Start(op util.Option) error
	Close() error
	SetCallback(callback Callback)
}

// Callback is invoked when this agent has received a request.
type Callback interface {
	Call(service string, method string, frame []byte, extra ...interface{})
}
