// Copyright (c) TFG Co. All Rights Reserved.
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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/seanbit/nrpc/cluster"
	"github.com/seanbit/nrpc/component"
	"github.com/seanbit/nrpc/interfaces"
	"github.com/seanbit/nrpc/metrics"
	"github.com/seanbit/nrpc/mocks"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestStaticConfigure(t *testing.T) {
	Configure("frontendType", map[string]string{}, []*viper.Viper{}...)

	require.NotNil(t, DefaultApp)
}

func TestStaticGetDieChan(t *testing.T) {
	ctrl := gomock.NewController(t)

	expected := make(chan bool)

	app := mocks.NewMockNRpc(ctrl)
	app.EXPECT().GetDieChan().Return(expected)

	DefaultApp = app
	require.Equal(t, expected, GetDieChan())
}

func TestStaticSetDebug(t *testing.T) {
	ctrl := gomock.NewController(t)

	expected := true

	app := mocks.NewMockNRpc(ctrl)
	app.EXPECT().SetDebug(expected)

	DefaultApp = app
	SetDebug(expected)
}

func TestStaticSetHeartbeatTime(t *testing.T) {
	ctrl := gomock.NewController(t)

	expected := 2 * time.Second

	app := mocks.NewMockNRpc(ctrl)
	app.EXPECT().SetHeartbeatTime(expected)

	DefaultApp = app
	SetHeartbeatTime(expected)
}

func TestStaticGetServerID(t *testing.T) {
	ctrl := gomock.NewController(t)

	expected := uuid.New().String()

	app := mocks.NewMockNRpc(ctrl)
	app.EXPECT().GetServerID().Return(expected)

	DefaultApp = app
	require.Equal(t, expected, GetServerID())
}

func TestStaticGetMetricsReporters(t *testing.T) {
	ctrl := gomock.NewController(t)

	expected := []metrics.Reporter{}

	app := mocks.NewMockNRpc(ctrl)
	app.EXPECT().GetMetricsReporters().Return(expected)

	DefaultApp = app
	require.Equal(t, expected, GetMetricsReporters())
}

func TestStaticGetServer(t *testing.T) {
	ctrl := gomock.NewController(t)

	expected := &cluster.Server{}

	app := mocks.NewMockNRpc(ctrl)
	app.EXPECT().GetServer().Return(expected)

	DefaultApp = app
	require.Equal(t, expected, GetServer())
}

func TestStaticGetServerByID(t *testing.T) {
	tables := []struct {
		name   string
		id     string
		err    error
		server *cluster.Server
	}{
		{"Success", uuid.New().String(), nil, &cluster.Server{}},
		{"Error", uuid.New().String(), errors.New("error"), nil},
	}

	for _, row := range tables {
		t.Run(row.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := mocks.NewMockNRpc(ctrl)
			app.EXPECT().GetServerByID(row.id).Return(row.server, row.err)

			DefaultApp = app
			server, err := GetServerByID(row.id)
			require.Equal(t, row.err, err)
			require.Equal(t, row.server, server)
		})
	}
}

func TestStaticGetServersByType(t *testing.T) {
	tables := []struct {
		name   string
		typ    string
		err    error
		server map[string]*cluster.Server
	}{
		{"Success", "type", nil, map[string]*cluster.Server{"type": {}}},
		{"Error", "type", errors.New("error"), nil},
	}

	for _, row := range tables {
		t.Run(row.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := mocks.NewMockNRpc(ctrl)
			app.EXPECT().GetServersByType(row.typ).Return(row.server, row.err)

			DefaultApp = app
			server, err := GetServersByType(row.typ)
			require.Equal(t, row.err, err)
			require.Equal(t, row.server, server)
		})
	}
}

func TestStaticGetServers(t *testing.T) {
	ctrl := gomock.NewController(t)

	expected := []*cluster.Server{{}}

	app := mocks.NewMockNRpc(ctrl)
	app.EXPECT().GetServers().Return(expected)

	DefaultApp = app
	require.Equal(t, expected, GetServers())
}

func TestStaticStart(t *testing.T) {
	ctrl := gomock.NewController(t)

	app := mocks.NewMockNRpc(ctrl)
	app.EXPECT().Start()

	DefaultApp = app
	Start()
}

func TestStaticSetDictionary(t *testing.T) {
	ctrl := gomock.NewController(t)

	expected := map[string]uint16{"test": 1}

	app := mocks.NewMockNRpc(ctrl)
	app.EXPECT().SetDictionary(expected).Return(nil)

	DefaultApp = app
	SetDictionary(expected)
}

func TestStaticAddRoute(t *testing.T) {
	tables := []struct {
		name       string
		serverType string
		err        error
	}{
		{"Success", "type", nil},
		{"Error", "type", errors.New("error")},
	}

	for _, row := range tables {
		t.Run(row.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := mocks.NewMockNRpc(ctrl)
			app.EXPECT().AddRoute(row.serverType, nil).Return(row.err) // Note that functions can't be tested for equality

			DefaultApp = app
			err := AddRoute(row.serverType, nil)
			require.Equal(t, row.err, err)
		})
	}
}

func TestStaticShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)

	app := mocks.NewMockNRpc(ctrl)
	app.EXPECT().Shutdown()

	DefaultApp = app
	Shutdown()
}

func TestStaticIsRunning(t *testing.T) {
	tables := []struct {
		name     string
		returned bool
	}{
		{"Success", true},
		{"Error", false},
	}

	for _, row := range tables {
		t.Run(row.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := mocks.NewMockNRpc(ctrl)
			app.EXPECT().IsRunning().Return(row.returned)

			DefaultApp = app
			require.Equal(t, row.returned, IsRunning())
		})
	}
}

func TestStaticRPC(t *testing.T) {
	ctx := context.Background()
	routeStr := "route"
	var reply proto.Message
	var arg proto.Message

	tables := []struct {
		name     string
		returned error
	}{
		{"Success", nil},
		{"Error", errors.New("error")},
	}

	for _, row := range tables {
		t.Run(row.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := mocks.NewMockNRpc(ctrl)
			app.EXPECT().RPC(ctx, routeStr, reply, arg).Return(row.returned)

			DefaultApp = app
			require.Equal(t, row.returned, RPC(ctx, routeStr, reply, arg))
		})
	}
}

func TestStaticRPCTo(t *testing.T) {
	ctx := context.Background()
	routeStr := "route"
	serverId := uuid.New().String()
	var reply proto.Message
	var arg proto.Message

	tables := []struct {
		name     string
		returned error
	}{
		{"Success", nil},
		{"Error", errors.New("error")},
	}

	for _, row := range tables {
		t.Run(row.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := mocks.NewMockNRpc(ctrl)
			app.EXPECT().RPCTo(ctx, serverId, routeStr, reply, arg).Return(row.returned)

			DefaultApp = app
			require.Equal(t, row.returned, RPCTo(ctx, serverId, routeStr, reply, arg))
		})
	}
}

func TestStaticRegister(t *testing.T) {
	var c component.Component
	options := []component.Option{}
	ctrl := gomock.NewController(t)

	app := mocks.NewMockNRpc(ctrl)
	app.EXPECT().Register(c, options)

	DefaultApp = app
	Register(c, options...)
}

func TestStaticRegisterRemote(t *testing.T) {
	var c component.Component
	options := []component.Option{}
	ctrl := gomock.NewController(t)

	app := mocks.NewMockNRpc(ctrl)
	app.EXPECT().RegisterRemote(c, options)

	DefaultApp = app
	RegisterRemote(c, options...)
}

func TestStaticRegisterModule(t *testing.T) {
	var module interfaces.Module
	name := "name"

	tables := []struct {
		name     string
		returned error
	}{
		{"Success", nil},
		{"Error", errors.New("error")},
	}

	for _, row := range tables {
		t.Run(row.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := mocks.NewMockNRpc(ctrl)
			app.EXPECT().RegisterModule(module, name).Return(row.returned)

			DefaultApp = app
			require.Equal(t, row.returned, RegisterModule(module, name))
		})
	}
}

func TestStaticRegisterModuleAfter(t *testing.T) {
	var module interfaces.Module
	name := "name"

	tables := []struct {
		name     string
		returned error
	}{
		{"Success", nil},
		{"Error", errors.New("error")},
	}

	for _, row := range tables {
		t.Run(row.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := mocks.NewMockNRpc(ctrl)
			app.EXPECT().RegisterModuleAfter(module, name).Return(row.returned)

			DefaultApp = app
			require.Equal(t, row.returned, RegisterModuleAfter(module, name))
		})
	}
}

func TestStaticRegisterModuleBefore(t *testing.T) {
	var module interfaces.Module
	name := "name"

	tables := []struct {
		name     string
		returned error
	}{
		{"Success", nil},
		{"Error", errors.New("error")},
	}

	for _, row := range tables {
		t.Run(row.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := mocks.NewMockNRpc(ctrl)
			app.EXPECT().RegisterModuleBefore(module, name).Return(row.returned)

			DefaultApp = app
			require.Equal(t, row.returned, RegisterModuleBefore(module, name))
		})
	}
}

func TestStaticGetModule(t *testing.T) {
	tables := []struct {
		name       string
		moduleName string
		module     interfaces.Module
		err        error
	}{
		{"Success", "name", nil, nil},
		{"Error", "name", nil, errors.New("error")},
	}

	for _, row := range tables {
		t.Run(row.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			app := mocks.NewMockNRpc(ctrl)
			app.EXPECT().GetModule(row.moduleName).Return(row.module, row.err)

			DefaultApp = app
			module, err := GetModule(row.moduleName)
			require.Equal(t, row.err, err)
			require.Equal(t, row.module, module)
		})
	}
}
