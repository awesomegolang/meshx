package client

import (
	"context"

	"github.com/smallnest/meshx/client/adapter"
	"github.com/smallnest/meshx/util"
	"github.com/smallnest/rpcx/client"
)

// Agent is the client agent (sidecar) in the client side.
// It redirects requests to the meshx server agent.
type Agent struct {
	op         util.Option
	adapter    adapter.Adapter
	rpcxClient client.XClient
}

// New creates an agent.
func New(op util.Option) *Agent {
	rpcName := op.Get("rpc-adapter")
	ad := adapter.NewAdapter(rpcName, op)
	return &Agent{
		op:      op,
		adapter: ad,
	}
}

// Start starts this agent.
func (a *Agent) Start() error {
	rpcxBasepath := a.op.GetWithDefault("rpcx-basepath", "/rpcx")
	etcdAddr := a.op.Get("etcd-address")

	d := client.NewEtcdDiscovery(rpcxBasepath, "MeshXService", []string{etcdAddr}, nil)
	a.rpcxClient = client.NewXClient("MeshXService", client.Failover, client.RoundRobin, d, client.DefaultOption)

	a.adapter.SetCallback(a)
	return a.adapter.Start(a.op)
}

// Call sends a request.
func (a *Agent) Call(service string, method string, frame []byte, extra ...interface{}) ([]byte, error) {
	args := &Args{
		Service: service,
		Method:  method,
		Frame:   frame,
	}

	reply := &Reply{}
	err := a.rpcxClient.Call(context.Background(), "Mul", args, reply)
	if err != nil {
		return nil, err
	}
	return reply.Data, nil
}
