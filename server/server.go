package server

import (
	"log"

	"github.com/smallnest/meshx/server/adapter"
	"github.com/smallnest/meshx/util"
	"github.com/smallnest/rpcx/server"
)

// Agent is the server agent (sidecar) in the server side.
// It accepts requests from meshx client agents.
type Agent struct {
	op         util.Option
	adapter    adapter.Adapter
	rpcxServer *server.Server
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
	rpcxNetwork := a.op.GetWithDefault("rpcx-network", "tcp")
	rpcxAddr := a.op.Get("rpcx-address")

	rpcxServer := server.NewServer()
	addRegistryPlugin(rpcxServer, a.op)
	rpcxServer.RegisterName("MeshXService", &MeshxService{agent: a}, "")
	a.rpcxServer = rpcxServer

	go func() {
		err := a.rpcxServer.Serve(rpcxNetwork, rpcxAddr)
		if err != nil {
			log.Fatalf("failed to start rpcx server: %v", err)
		}
	}()

	return a.adapter.Start(a.op)
}

func addRegistryPlugin(s *server.Server, op util.Option) {

	// r := &serverplugin.EtcdRegisterPlugin{
	// 	ServiceAddress: "tcp@" + *addr,
	// 	EtcdServers:    []string{*etcdAddr},
	// 	BasePath:       *basePath,
	// 	Metrics:        metrics.NewRegistry(),
	// 	UpdateInterval: time.Minute,
	// }
	// err := r.Start()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// s.Plugins.Add(r)
}
