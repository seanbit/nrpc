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

package cluster

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	nats "github.com/nats-io/nats.go"
	"github.com/seanbit/nrpc/config"
	"github.com/seanbit/nrpc/constants"
	"github.com/seanbit/nrpc/helpers"
	"github.com/seanbit/nrpc/metrics"
	metricsmocks "github.com/seanbit/nrpc/metrics/mocks"
	"github.com/seanbit/nrpc/protos"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

type funcPtrMatcher struct {
	ptr uintptr
}

func newFuncPtrMatcher(x interface{}) *funcPtrMatcher {
	if reflect.ValueOf(x).Kind() != reflect.Func {
		panic("funcPtrMatcher only accepts functions as arguments")
	}
	return &funcPtrMatcher{
		ptr: reflect.ValueOf(x).Pointer(),
	}
}

func (m funcPtrMatcher) Matches(x interface{}) bool {
	v := reflect.ValueOf(x)
	switch v.Kind() {
	case reflect.Func:
		ptr := reflect.ValueOf(x).Pointer()
		return ptr == m.ptr
	default:
		return false
	}
}

func (m funcPtrMatcher) String() string {
	return fmt.Sprintf("has address %d", m.ptr)
}

func TestNewNatsRPCServer(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMetricsReporter := metricsmocks.NewMockReporter(ctrl)
	mockMetricsReporters := []metrics.Reporter{mockMetricsReporter}

	cfg := config.NewDefaultConfig().Cluster.RPC.Server.Nats
	sv := getServer()
	n, err := NewNatsRPCServer(cfg, sv, mockMetricsReporters, nil)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.Equal(t, sv, n.server)
	assert.Equal(t, mockMetricsReporters, n.metricsReporters)
}

func TestNatsRPCServerConfigure(t *testing.T) {
	t.Parallel()
	tables := []struct {
		natsConnect        string
		messagesBufferSize int
		pushBufferSize     int
		err                error
	}{
		{"nats://localhost:2333", 10, 10, nil},
		{"nats://localhost:2333", 10, 0, constants.ErrNatsPushBufferSizeZero},
		{"nats://localhost:2333", 0, 10, constants.ErrNatsMessagesBufferSizeZero},
		{"", 10, 10, constants.ErrNoNatsConnectionString},
	}

	for _, table := range tables {
		t.Run(fmt.Sprintf("%s-%d-%d", table.natsConnect, table.messagesBufferSize, table.pushBufferSize), func(t *testing.T) {
			cfg := config.NewDefaultConfig().Cluster.RPC.Server.Nats
			cfg.Connect = table.natsConnect
			cfg.Buffer.Messages = table.messagesBufferSize
			_, err := NewNatsRPCServer(cfg, getServer(), nil, nil)
			assert.Equal(t, table.err, err)
		})
	}
}

func TestNatsRPCServerGetUnhandledRequestsChannel(t *testing.T) {
	t.Parallel()
	cfg := config.NewDefaultConfig().Cluster.RPC.Server.Nats
	sv := getServer()
	n, _ := NewNatsRPCServer(cfg, sv, nil, nil)
	assert.NotNil(t, n.GetUnhandledRequestsChannel())
	assert.IsType(t, make(chan *protos.Request), n.GetUnhandledRequestsChannel())
}

func TestNatsRPCServerSubscribe(t *testing.T) {
	cfg := config.NewDefaultConfig().Cluster.RPC.Server.Nats
	sv := getServer()
	rpcServer, _ := NewNatsRPCServer(cfg, sv, nil, nil)
	s := helpers.GetTestNatsServer(t)
	defer func() {
		s.Shutdown()
		s.WaitForShutdown()
	}()
	conn, err := setupNatsConn(fmt.Sprintf("nats://%s", s.Addr()), nil)
	assert.NoError(t, err)
	rpcServer.conn = conn
	tables := []struct {
		topic string
		msg   []byte
	}{
		{"user1/messages", []byte("msg1")},
		{"user2/messages", []byte("")},
		{"u/messages", []byte("000")},
	}

	for _, table := range tables {
		t.Run(table.topic, func(t *testing.T) {
			subs, err := rpcServer.subscribe(table.topic)
			assert.NoError(t, err)
			assert.Equal(t, true, subs.IsValid())
			conn.Publish(table.topic, table.msg)
			r := helpers.ShouldEventuallyReceive(t, rpcServer.subChan).(*nats.Msg)
			assert.Equal(t, table.msg, r.Data)
		})
	}
	conn.Close()
}

func TestNatsRPCServerHandleMessages(t *testing.T) {
	cfg := config.NewDefaultConfig().Cluster.RPC.Server.Nats
	sv := getServer()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMetricsReporter := metricsmocks.NewMockReporter(ctrl)
	mockMetricsReporters := []metrics.Reporter{mockMetricsReporter}

	rpcServer, _ := NewNatsRPCServer(cfg, sv, mockMetricsReporters, nil)
	s := helpers.GetTestNatsServer(t)
	defer func() {
		s.Shutdown()
		s.WaitForShutdown()
	}()
	conn, err := setupNatsConn(fmt.Sprintf("nats://%s", s.Addr()), nil)
	assert.NoError(t, err)
	rpcServer.conn = conn
	tables := []struct {
		topic string
		req   *protos.Request
	}{
		{"user1/messages", &protos.Request{Msg: &protos.Msg{Id: 1, Reply: "ae"}}},
		{"user2/messages", &protos.Request{Msg: &protos.Msg{Id: 1}}},
	}

	go rpcServer.handleMessages()

	for _, table := range tables {
		t.Run(table.topic, func(t *testing.T) {
			subs, err := rpcServer.subscribe(table.topic)
			assert.NoError(t, err)
			assert.Equal(t, true, subs.IsValid())
			b, err := proto.Marshal(table.req)
			assert.NoError(t, err)

			mockMetricsReporter.EXPECT().ReportGauge(metrics.DroppedMessages, gomock.Any(), float64(0))
			mockMetricsReporter.EXPECT().ReportGauge(metrics.ChannelCapacity, gomock.Any(), gomock.Any()).Times(1)

			conn.Publish(table.topic, b)
			r := helpers.ShouldEventuallyReceive(t, rpcServer.unhandledReqCh).(*protos.Request)
			assert.Equal(t, table.req.Msg.Id, r.Msg.Id)
		})
	}
	conn.Close()
}

func TestNatsRPCServerInitShouldFailIfConnFails(t *testing.T) {
	t.Parallel()
	cfg := config.NewDefaultConfig().Cluster.RPC.Server.Nats
	cfg.Connect = "nats://localhost:1"
	sv := getServer()

	rpcServer, _ := NewNatsRPCServer(cfg, sv, nil, nil)
	//mockSessionPool.EXPECT().OnSessionBind(rpcServer.onSessionBind)
	err := rpcServer.Init()
	assert.Error(t, err)
}

func TestNatsRPCServerInit(t *testing.T) {
	s := helpers.GetTestNatsServer(t)
	defer func() {
		s.Shutdown()
		s.WaitForShutdown()
	}()
	cfg := config.NewDefaultConfig().Cluster.RPC.Server.Nats
	cfg.Connect = fmt.Sprintf("nats://%s", s.Addr())
	sv := getServer()

	rpcServer, _ := NewNatsRPCServer(cfg, sv, nil, nil)
	err := rpcServer.Init()
	assert.NoError(t, err)
	// should setup conn
	assert.NotNil(t, rpcServer.conn)
	assert.True(t, rpcServer.conn.IsConnected())
	// should subscribe
	assert.True(t, rpcServer.sub.IsValid())
	//should handle messages
	tables := []struct {
		name  string
		topic string
		req   *protos.Request
	}{
		{"test1", getChannel(sv.Type, sv.ID), &protos.Request{Msg: &protos.Msg{Id: 1, Reply: "ae"}}},
		{"test2", getChannel(sv.Type, sv.ID), &protos.Request{Msg: &protos.Msg{Id: 1, Reply: "boa"}}},
	}
	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			c := make(chan *nats.Msg)
			rpcServer.conn.ChanSubscribe(table.req.Msg.Reply, c)
			rpcServer.unhandledReqCh <- table.req
			r := helpers.ShouldEventuallyReceive(t, c).(*nats.Msg)
			assert.NotNil(t, r.Data)
		})
	}
}

func TestNatsRPCServerReportMetrics(t *testing.T) {
	cfg := config.NewDefaultConfig().Cluster.RPC.Server.Nats
	sv := getServer()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMetricsReporter := metricsmocks.NewMockReporter(ctrl)
	mockMetricsReporters := []metrics.Reporter{mockMetricsReporter}

	rpcServer, _ := NewNatsRPCServer(cfg, sv, mockMetricsReporters, nil)
	rpcServer.dropped = 100
	rpcServer.messagesBufferSize = 100

	rpcServer.subChan <- &nats.Msg{}

	mockMetricsReporter.EXPECT().ReportGauge(metrics.DroppedMessages, gomock.Any(), float64(rpcServer.dropped))
	mockMetricsReporter.EXPECT().ReportGauge(metrics.ChannelCapacity, gomock.Any(), float64(99)).Times(1)
	rpcServer.reportMetrics()
}
