// Copyright (c) nano Author and TFG Co. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package nrpc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/seanbit/nrpc/cluster"
	"github.com/seanbit/nrpc/config"
	"github.com/seanbit/nrpc/conn/message"
	"github.com/seanbit/nrpc/constants"
	e "github.com/seanbit/nrpc/errors"
	"github.com/seanbit/nrpc/helpers"
	"github.com/seanbit/nrpc/route"
	"github.com/seanbit/nrpc/router"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var (
	tables = []struct {
		serverType     string
		serverMetadata map[string]string
		cfg            *viper.Viper
	}{
		{"sv1", map[string]string{"name": "bla"}, viper.New()},
		{"sv2", map[string]string{}, viper.New()},
	}
)

func TestMain(m *testing.M) {
	exit := m.Run()
	os.Exit(exit)
}

func TestNewApp(t *testing.T) {
	for _, table := range tables {
		t.Run(table.serverType, func(t *testing.T) {
			builderConfig := config.NewDefaultConfig()
			app := NewDefaultApp(table.serverType, table.serverMetadata, *builderConfig).(*App)
			assert.Equal(t, table.serverType, app.server.Type)
			assert.Equal(t, table.serverMetadata, app.server.Metadata)
		})
	}
}

func TestSetDebug(t *testing.T) {
	builderConfig := config.NewDefaultConfig()
	app := NewDefaultApp("testtype", map[string]string{}, *builderConfig).(*App)
	app.SetDebug(true)
	assert.Equal(t, true, app.debug)
	app.SetDebug(false)
	assert.Equal(t, false, app.debug)
}

func TestGetDieChan(t *testing.T) {
	builderConfig := config.NewDefaultConfig()
	app := NewDefaultApp("testtype", map[string]string{}, *builderConfig).(*App)
	assert.Equal(t, app.dieChan, app.GetDieChan())
}

func TestGetSever(t *testing.T) {
	builderConfig := config.NewDefaultConfig()
	app := NewDefaultApp("testtype", map[string]string{}, *builderConfig).(*App)
	assert.Equal(t, app.server, app.GetServer())
}

func TestGetMetricsReporters(t *testing.T) {
	builderConfig := config.NewDefaultConfig()
	app := NewDefaultApp("testtype", map[string]string{}, *builderConfig).(*App)
	assert.Equal(t, app.metricsReporters, app.GetMetricsReporters())
}
func TestGetServerByID(t *testing.T) {
	builderConfig := config.NewDefaultConfig()
	app := NewDefaultApp("testtype", map[string]string{}, *builderConfig)
	s, err := app.GetServerByID("id")
	assert.Nil(t, s)
	assert.EqualError(t, constants.ErrNoServerWithID, err.Error())
}

func TestGetServersByType(t *testing.T) {
	builderConfig := config.NewDefaultConfig()
	app := NewDefaultApp("testtype", map[string]string{}, *builderConfig)
	s, err := app.GetServersByType("id")
	assert.Nil(t, s)
	assert.EqualError(t, constants.ErrNoServersAvailableOfType, err.Error())
}

func TestSetHeartbeatInterval(t *testing.T) {
	inter := 35 * time.Millisecond
	builderConfig := config.NewDefaultConfig()
	app := NewDefaultApp("testtype", map[string]string{}, *builderConfig).(*App)
	app.SetHeartbeatTime(inter)
	assert.Equal(t, inter, app.heartbeat)
}

func TestSetDictionary(t *testing.T) {
	builderConfig := config.NewDefaultConfig()
	app := NewDefaultApp("testtype", map[string]string{}, *builderConfig).(*App)

	dict := map[string]uint16{"someroute": 12}
	err := app.SetDictionary(dict)
	assert.NoError(t, err)
	assert.Equal(t, dict, message.GetDictionary())

	app.running = true
	err = app.SetDictionary(dict)
	assert.EqualError(t, constants.ErrChangeDictionaryWhileRunning, err.Error())
}

func TestAddRoute(t *testing.T) {
	builderConfig := config.NewDefaultConfig()
	app := NewDefaultApp("testtype", map[string]string{}, *builderConfig).(*App)
	app.router = nil
	err := app.AddRoute("somesv", func(ctx context.Context, route *route.Route, payload []byte, servers map[string]*cluster.Server) (context.Context, *cluster.Server, error) {
		return ctx, nil, nil
	})
	assert.EqualError(t, constants.ErrRouterNotInitialized, err.Error())

	app.router = router.New()
	err = app.AddRoute("somesv", func(ctx context.Context, route *route.Route, payload []byte, servers map[string]*cluster.Server) (context.Context, *cluster.Server, error) {
		return ctx, nil, nil
	})
	assert.NoError(t, err)

	app.running = true
	err = app.AddRoute("somesv", func(ctx context.Context, route *route.Route, payload []byte, servers map[string]*cluster.Server) (context.Context, *cluster.Server, error) {
		return ctx, nil, nil
	})
	assert.EqualError(t, constants.ErrChangeRouteWhileRunning, err.Error())
}

func TestShutdown(t *testing.T) {
	builderConfig := config.NewDefaultConfig()
	app := NewDefaultApp("testtype", map[string]string{}, *builderConfig).(*App)
	go func() {
		app.Shutdown()
	}()
	<-app.dieChan
}

func TestConfigureDefaultMetricsReporter(t *testing.T) {
	tables := []struct {
		enabled bool
	}{
		{true},
		{false},
	}

	for _, table := range tables {
		t.Run(fmt.Sprintf("%t", table.enabled), func(t *testing.T) {
			builderConfig := config.NewDefaultConfig()
			builderConfig.Metrics.Prometheus.Enabled = table.enabled
			builderConfig.Metrics.Statsd.Enabled = table.enabled
			app := NewDefaultApp("testtype", map[string]string{}, *builderConfig).(*App)
			// if statsd is enabled there are 2 metricsReporters, prometheus and statsd
			assert.Equal(t, table.enabled, len(app.metricsReporters) == 2)
		})
	}
}

func TestDefaultSD(t *testing.T) {
	builderConfig := config.NewDefaultConfig()
	app := NewDefaultApp("testtype", map[string]string{}, *builderConfig).(*App)
	assert.NotNil(t, app.serviceDiscovery)

	etcdSD, err := cluster.NewEtcdServiceDiscovery(config.NewDefaultConfig().Cluster.SD.Etcd, app.server, app.dieChan)
	assert.NoError(t, err)
	typeOfetcdSD := reflect.TypeOf(etcdSD)

	assert.Equal(t, typeOfetcdSD, reflect.TypeOf(app.serviceDiscovery))
}

func TestDefaultRPCServer(t *testing.T) {
	builderConfig := config.NewDefaultConfig()
	app := NewDefaultApp("testtype", map[string]string{}, *builderConfig).(*App)
	assert.NotNil(t, app.rpcServer)

	natsRPCServer, err := cluster.NewNatsRPCServer(config.NewDefaultConfig().Cluster.RPC.Server.Nats, app.server, nil, app.dieChan, nil)
	assert.NoError(t, err)
	typeOfNatsRPCServer := reflect.TypeOf(natsRPCServer)

	assert.Equal(t, typeOfNatsRPCServer, reflect.TypeOf(app.rpcServer))
}

func TestDefaultRPCClient(t *testing.T) {
	builderConfig := config.NewDefaultConfig()
	app := NewDefaultApp("testtype", map[string]string{}, *builderConfig).(*App)
	assert.NotNil(t, app.rpcClient)

	natsRPCClient, err := cluster.NewNatsRPCClient(config.NewDefaultConfig().Cluster.RPC.Client.Nats, app.server, nil, app.dieChan)
	assert.NoError(t, err)
	typeOfNatsRPCClient := reflect.TypeOf(natsRPCClient)

	assert.Equal(t, typeOfNatsRPCClient, reflect.TypeOf(app.rpcClient))
}

func TestStartAndListenCluster(t *testing.T) {
	es, cli := helpers.GetTestEtcd(t)
	defer es.Terminate(t)

	ns := helpers.GetTestNatsServer(t)
	nsAddr := ns.Addr().String()

	builder := NewDefaultBuilder("testtype", map[string]string{}, *config.NewDefaultConfig())

	var err error
	natsClientConfig := config.NewDefaultConfig().Cluster.RPC.Client.Nats
	natsClientConfig.Connect = fmt.Sprintf("nats://%s", nsAddr)
	builder.RPCClient, err = cluster.NewNatsRPCClient(natsClientConfig, builder.Server, builder.MetricsReporters, builder.DieChan)
	if err != nil {
		panic(err.Error())
	}

	natsServerConfig := config.NewDefaultConfig().Cluster.RPC.Server.Nats
	natsServerConfig.Connect = fmt.Sprintf("nats://%s", nsAddr)
	builder.RPCServer, err = cluster.NewNatsRPCServer(natsServerConfig, builder.Server, builder.MetricsReporters, builder.DieChan, nil)
	if err != nil {
		panic(err.Error())
	}

	etcdSD, err := cluster.NewEtcdServiceDiscovery(config.NewDefaultConfig().Cluster.SD.Etcd, builder.Server, builder.DieChan, cli)
	builder.ServiceDiscovery = etcdSD
	assert.NoError(t, err)
	app := builder.Build().(*App)

	go func() {
		app.Start()
	}()
	helpers.ShouldEventuallyReturn(t, func() bool {
		return app.running
	}, true)
}

func TestError(t *testing.T) {
	t.Parallel()

	tables := []struct {
		name     string
		err      error
		code     string
		metadata map[string]string
	}{
		{"nil_metadata", errors.New(uuid.New().String()), uuid.New().String(), nil},
		{"empty_metadata", errors.New(uuid.New().String()), uuid.New().String(), map[string]string{}},
		{"non_empty_metadata", errors.New(uuid.New().String()), uuid.New().String(), map[string]string{"key": uuid.New().String()}},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			var err *e.Error
			if table.metadata != nil {
				err = Error(table.err, table.code, table.metadata)
			} else {
				err = Error(table.err, table.code)
			}
			assert.NotNil(t, err)
			assert.Equal(t, table.code, err.Code)
			assert.Equal(t, table.err.Error(), err.Message)
			assert.Equal(t, table.metadata, err.Metadata)
		})
	}
}

func TestAddMetricTagsToPropagateCtx(t *testing.T) {
	ctx := AddMetricTagsToPropagateCtx(context.Background(), map[string]string{
		"key": "value",
	})
	val := ctx.Value(constants.PropagateCtxKey)
	assert.Equal(t, map[string]interface{}{
		constants.MetricTagsKey: map[string]string{
			"key": "value",
		},
	}, val)
}

func TestAddToPropagateCtx(t *testing.T) {
	ctx := AddToPropagateCtx(context.Background(), "key", "val")
	val := ctx.Value(constants.PropagateCtxKey)
	assert.Equal(t, map[string]interface{}{"key": "val"}, val)
}

func TestGetFromPropagateCtx(t *testing.T) {
	ctx := AddToPropagateCtx(context.Background(), "key", "val")
	val := GetFromPropagateCtx(ctx, "key")
	assert.Equal(t, "val", val)
}

func TestExtractSpan(t *testing.T) {
	span := opentracing.StartSpan("op", opentracing.ChildOf(nil))
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	spanCtx, err := ExtractSpan(ctx)
	assert.NoError(t, err)
	assert.Equal(t, span.Context(), spanCtx)
}
