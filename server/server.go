package server

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/meshx/server/adapter"
	"github.com/smallnest/meshx/util"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
)

const (
	RPCAdaptor   = "rpc-adapter"
	RPCXBasepath = "rpcx-basepath"
	RPCXRegistry = "rpcx-registry"
	RPCXNetwork  = "rpcx-network" // agent服务的网络类型
	RPCXAddress  = "rpcx-address" // agent服务的地址

	ServiceName = "MeshXService"

	RPCXMDNSDomain = "rpcx-mdns-domain"
	RPCXMDNSPort   = "rpcx-mdns-port"
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
	rpcName := op.Get(RPCAdaptor)
	ad := adapter.NewAdapter(rpcName, op)
	return &Agent{
		op:      op,
		adapter: ad,
	}
}

// Start starts this agent.
func (a *Agent) Start() error {
	rpcxNetwork := a.op.GetWithDefault(RPCXNetwork, "tcp")
	rpcxAddr := a.op.Get(RPCXAddress)
	rpcxRegAddr := a.op.Get(RPCXRegistry)
	rpcxBasepath := a.op.Get(RPCXBasepath)

	rpcxServer := server.NewServer()
	addRegistryPlugin(rpcxServer, rpcxRegAddr, rpcxBasepath, rpcxNetwork, rpcxAddr, a.op)
	rpcxServer.RegisterName(ServiceName, &MeshxService{agent: a}, "")
	a.rpcxServer = rpcxServer

	go func() {
		err := a.rpcxServer.Serve(rpcxNetwork, rpcxAddr)
		if err != nil {
			log.Fatalf("failed to start rpcx server: %v", err)
		}
	}()

	return a.adapter.Start(a.op)
}

func addRegistryPlugin(s *server.Server, regAddr, basePath, rpcxNetwork, listenAddr string, op util.Option) error {
	i := strings.Index(regAddr, "://")
	if i < 0 {
		return errors.New("wrong format registry address. The right fotmat is [registry_type://address]")
	}

	regType := regAddr[:i]
	regAddr = regAddr[i+3:]

	switch regType {
	case "peer2peer": //peer2peer://127.0.0.1:8972
		return nil
	case "multiple":
		return nil
	case "zookeeper":
		r := &serverplugin.ZooKeeperRegisterPlugin{
			ServiceAddress:   rpcxNetwork + "@" + listenAddr,
			ZooKeeperServers: []string{regAddr},
			BasePath:         basePath,
			Metrics:          metrics.NewRegistry(),
			UpdateInterval:   time.Minute,
		}
		err := r.Start()
		if err != nil {
			log.Fatal(err)
		}
		s.Plugins.Add(r)
		return nil
	case "etcd":
		r := &serverplugin.EtcdRegisterPlugin{
			ServiceAddress: rpcxNetwork + "@" + listenAddr,
			EtcdServers:    []string{regAddr},
			BasePath:       basePath,
			Metrics:        metrics.NewRegistry(),
			UpdateInterval: time.Minute,
		}
		err := r.Start()
		if err != nil {
			log.Fatal(err)
		}
		s.Plugins.Add(r)
		return nil
	case "consul":
		r := &serverplugin.ConsulRegisterPlugin{
			ServiceAddress: rpcxNetwork + "@" + listenAddr,
			ConsulServers:  []string{regAddr},
			BasePath:       basePath,
			Metrics:        metrics.NewRegistry(),
			UpdateInterval: time.Minute,
		}
		err := r.Start()
		if err != nil {
			log.Fatal(err)
		}
		s.Plugins.Add(r)
		return nil
	case "mdns":
		domain := op.Get(RPCXMDNSDomain)
		portStr := op.Get(RPCXMDNSPort)
		port, err := strconv.Atoi(portStr)

		r := serverplugin.NewMDNSRegisterPlugin(rpcxNetwork+"@"+regAddr, port, metrics.NewRegistry(), time.Minute, domain)
		err = r.Start()
		if err != nil {
			log.Fatal(err)
		}
		s.Plugins.Add(r)
		return nil
	default:
		return fmt.Errorf("wrong registry type %s. only support peer2peer,multiple, zookeeper, etcd, consul and mdns", regType)
	}
}
