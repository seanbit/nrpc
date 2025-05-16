package main

import (
	"github.com/seanbit/nrpc"
	"github.com/seanbit/nrpc/component"
	"github.com/seanbit/nrpc/config"
	"github.com/seanbit/nrpc/pipeline/plugin"
	"strings"
	"sync"
	"time"
)

func main() {
	StartServer("test_server", false)
}

func StartServer(serverType string, async bool) {
	cfg := config.NewDefaultConfig()
	serverMetadata := map[string]string{}

	builder := nrpc.NewBuilder(serverType, serverMetadata, *cfg)
	rateLimiter := plugin.NewRateLimiter(builder.MetricsReporters, 10, time.Second, true)
	builder.RemoteHooks.BeforeHandler.PushFront(rateLimiter.HandleBefore)
	nrpc.DefaultApp = builder.Build()

	if !async {
		defer nrpc.Shutdown()
	}
	nrpc.Register(NewService(),
		component.WithName("serv"),
		component.WithNameFunc(strings.ToLower),
	)
	if async {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		if err := nrpc.RegisterModule(&sysModule{wg: wg}, "sys"); err != nil {
			panic(err)
		}
		go nrpc.Start()
		wg.Wait()
	} else {
		nrpc.Start()
	}
}

type sysModule struct {
	wg *sync.WaitGroup
}

func (m *sysModule) Init() error     { return nil }
func (m *sysModule) AfterInit()      { m.wg.Done() }
func (m *sysModule) BeforeShutdown() {}
func (m *sysModule) Shutdown() error { return nil }
