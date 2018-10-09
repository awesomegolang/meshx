package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/smallnest/meshx/client/adapter"
	"github.com/smallnest/meshx/util"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/log"
)

const (
	RPCAdaptor     = "rpc-adapter"
	RPCXBasepath   = "rpcx-basepath"
	RPCXRegistry   = "rpcx-registry"
	RPCXFailmode   = "rpcx-failmode"
	RPCXSelectmode = "rpcx-selectmode"

	ServiceMethodName = "MeshXService.Invoke"
)

// Agent is the client agent (sidecar) in the client side.
// It redirects requests to the meshx server agent.
type Agent struct {
	op          util.Option
	adapter     adapter.Adapter
	rpcxClients map[string]client.XClient

	basePath       string
	registry       string
	rpcxFailmode   client.FailMode
	rpcxSelectmode client.SelectMode
	sd             client.ServiceDiscovery
}

// New creates an agent.
func New(op util.Option) *Agent {
	rpcName := op.Get(RPCAdaptor)
	ad := adapter.NewAdapter(rpcName, op)

	rpcxBasepath := op.GetWithDefault(RPCXBasepath, "/rpcx")
	rpcxRegistry := op.Get(RPCXRegistry)
	rpcxFailmode := op.GetWithDefault(RPCXFailmode, "Failover")
	rpcxSelectmode := op.GetWithDefault(RPCXSelectmode, "RoundRobin")

	rpcxFailm, err := client.FailModeString(rpcxFailmode)
	if err != nil {
		log.Fatal(err)
	}

	rpcxSelectm, err := client.SelectModeString(rpcxSelectmode)
	if err != nil {
		log.Fatal(err)
	}

	d, err := createServiceDiscovery(rpcxRegistry, rpcxBasepath)
	if err != nil {
		log.Fatal(err)
	}

	return &Agent{
		op:             op,
		adapter:        ad,
		rpcxClients:    make(map[string]client.XClient),
		basePath:       rpcxBasepath,
		registry:       rpcxRegistry,
		rpcxFailmode:   rpcxFailm,
		rpcxSelectmode: rpcxSelectm,
		sd:             d,
	}
}

// Start starts this agent.
func (a *Agent) Start() error {
	a.adapter.SetCallback(a)
	return a.adapter.Start(a.op)
}

// Close closes all xclients.
func (a *Agent) Close() error {
	var err error
	for k, v := range a.rpcxClients {
		err = v.Close()
		if err != nil {
			log.Errorf("failed to close client for %s: %v", k, err)
		}
	}

	return nil
}

func createServiceDiscovery(regAddr, basePath string) (client.ServiceDiscovery, error) {
	i := strings.Index(regAddr, "://")
	if i < 0 {
		return nil, errors.New("wrong format registry address. The right fotmat is [registry_type://address]")
	}

	regType := regAddr[:i]
	regAddr = regAddr[i+3:]

	switch regType {
	case "peer2peer": //peer2peer://127.0.0.1:8972
		return client.NewPeer2PeerDiscovery("tcp@"+regAddr, ""), nil
	case "multiple":
		var pairs []*client.KVPair
		pp := strings.Split(regAddr, ",")
		for _, v := range pp {
			pairs = append(pairs, &client.KVPair{Key: v})
		}
		return client.NewMultipleServersDiscovery(pairs), nil
	case "zookeeper":
		return client.NewZookeeperDiscoveryTemplate(basePath, []string{regAddr}, nil), nil
	case "etcd":
		return client.NewEtcdDiscoveryTemplate(basePath, []string{regAddr}, nil), nil
	case "consul":
		return client.NewConsulDiscoveryTemplate(basePath, []string{regAddr}, nil), nil
	case "mdns":
		client.NewMDNSDiscoveryTemplate(10*time.Second, 10*time.Second, "")
	default:
		return nil, fmt.Errorf("wrong registry type %s. only support peer2peer,multiple, zookeeper, etcd, consul and mdns", regType)
	}

	return nil, errors.New("wrong registry type. only support peer2peer,multiple, zookeeper, etcd, consul and mdns")
}

// Call sends a request.
func (a *Agent) Call(service string, method string, frame []byte, extra ...interface{}) ([]byte, error) {
	args := &Args{
		Service: service,
		Method:  method,
		Frame:   frame,
	}

	reply := &Reply{}
	xc, err := a.getXClient(service)
	if err != nil {
		return nil, err
	}
	err = xc.Call(context.Background(), ServiceMethodName, args, reply)
	if err != nil {
		return nil, err
	}
	return reply.Data, nil
}

func (a *Agent) getXClient(service string) (xc client.XClient, err error) {
	defer func() {
		if e := recover(); e != nil {
			if ee, ok := e.(error); ok {
				err = ee
				return
			}

			err = fmt.Errorf("failed to get xclient: %v", e)
		}
	}()

	if a.rpcxClients[service] == nil {
		a.rpcxClients[service] = client.NewXClient(service, a.rpcxFailmode, a.rpcxSelectmode, a.sd.Clone(service), client.DefaultOption)
	}
	xc = a.rpcxClients[service]

	return xc, err
}
