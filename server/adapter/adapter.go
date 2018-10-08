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
		log.Infof("server adapter '%v' has alrealy registered, the old adapter will be replaced", name)
	}
	adapters[name] = adapter
}

// NewAdapter creates an adapter.
func NewAdapter(name string, op util.Option) Adapter {
	fn, ok := adapters[name]
	if !ok {
		log.Fatalf("no server adapter registered for '%v'", name)
	}
	return fn(op)
}

// Adapter is the server agent interface of sidecar which implements protocol of rpc frameworks.
// It is used to connect the local rpc server.
type Adapter interface {
	Start(op util.Option) error
	Close() error
	Send(data []byte) ([]byte, error)
}
