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
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	nats "github.com/nats-io/nats.go"
	"github.com/seanbit/nrpc/config"
	"github.com/seanbit/nrpc/conn/message"
	"github.com/seanbit/nrpc/constants"
	e "github.com/seanbit/nrpc/errors"
	"github.com/seanbit/nrpc/helpers"
	"github.com/seanbit/nrpc/metrics"
	metricsmocks "github.com/seanbit/nrpc/metrics/mocks"
	"github.com/seanbit/nrpc/protos"
	"github.com/seanbit/nrpc/route"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestNewNatsRPCClient(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMetricsReporter := metricsmocks.NewMockReporter(ctrl)
	mockMetricsReporters := []metrics.Reporter{mockMetricsReporter}

	cfg := config.NewDefaultConfig().Cluster.RPC.Client.Nats
	sv := getServer()
	n, err := NewNatsRPCClient(cfg, sv, mockMetricsReporters, nil)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.Equal(t, sv, n.server)
	assert.Equal(t, mockMetricsReporters, n.metricsReporters)
	assert.False(t, n.running)
}

func TestNatsRPCClientConfigure(t *testing.T) {
	t.Parallel()
	tables := []struct {
		natsConnect string
		reqTimeout  time.Duration
		err         error
	}{
		{"nats://localhost:2333", time.Duration(10 * time.Second), nil},
		{"nats://localhost:2333", time.Duration(0), constants.ErrNatsNoRequestTimeout},
		{"", time.Duration(10 * time.Second), constants.ErrNoNatsConnectionString},
	}

	for _, table := range tables {
		t.Run(fmt.Sprintf("%s-%s", table.natsConnect, table.reqTimeout), func(t *testing.T) {
			cfg := config.NewDefaultConfig().Cluster.RPC.Client.Nats
			cfg.Connect = table.natsConnect
			cfg.RequestTimeout = table.reqTimeout
			_, err := NewNatsRPCClient(cfg, getServer(), nil, nil)
			assert.Equal(t, table.err, err)
		})
	}
}

func TestNatsRPCClientGetSubscribeChannel(t *testing.T) {
	t.Parallel()
	cfg := config.NewDefaultConfig().Cluster.RPC.Client.Nats
	sv := getServer()
	n, _ := NewNatsRPCClient(cfg, sv, nil, nil)
	assert.Equal(t, fmt.Sprintf("nrpc/servers/%s/%s", n.server.Type, n.server.ID), n.getSubscribeChannel())
}

func TestNatsRPCClientStop(t *testing.T) {
	t.Parallel()
	cfg := config.NewDefaultConfig().Cluster.RPC.Client.Nats
	sv := getServer()
	n, _ := NewNatsRPCClient(cfg, sv, nil, nil)
	// change it to true to ensure it goes to false
	n.running = true
	n.stop()
	assert.False(t, n.running)
}

func TestNatsRPCClientInitShouldFailIfConnFails(t *testing.T) {
	t.Parallel()
	sv := getServer()
	cfg := config.NewDefaultConfig().Cluster.RPC.Client.Nats
	cfg.Connect = "nats://localhost:1"
	rpcClient, _ := NewNatsRPCClient(cfg, sv, nil, nil)
	err := rpcClient.Init()
	assert.Error(t, err)
}

func TestNatsRPCClientInit(t *testing.T) {
	s := helpers.GetTestNatsServer(t)
	defer s.Shutdown()
	cfg := config.NewDefaultConfig().Cluster.RPC.Client.Nats
	cfg.Connect = fmt.Sprintf("nats://%s", s.Addr())
	sv := getServer()

	rpcClient, _ := NewNatsRPCClient(cfg, sv, nil, nil)
	err := rpcClient.Init()
	assert.NoError(t, err)
	assert.True(t, rpcClient.running)

	// should setup conn
	assert.NotNil(t, rpcClient.conn)
	assert.True(t, rpcClient.conn.IsConnected())
}

func TestNatsRPCClientSendShouldFailIfNotRunning(t *testing.T) {
	config := config.NewDefaultConfig().Cluster.RPC.Client.Nats
	sv := getServer()
	rpcClient, _ := NewNatsRPCClient(config, sv, nil, nil)
	err := rpcClient.Send("topic", []byte("data"))
	assert.Equal(t, constants.ErrRPCClientNotInitialized, err)
}

func TestNatsRPCClientSend(t *testing.T) {
	s := helpers.GetTestNatsServer(t)
	defer s.Shutdown()
	cfg := config.NewDefaultConfig().Cluster.RPC.Client.Nats
	cfg.Connect = fmt.Sprintf("nats://%s", s.Addr())
	sv := getServer()

	rpcClient, _ := NewNatsRPCClient(cfg, sv, nil, nil)
	rpcClient.Init()

	tables := []struct {
		name  string
		topic string
		data  []byte
	}{
		{"test1", getChannel(sv.Type, sv.ID), []byte("test1")},
		{"test2", getChannel(sv.Type, sv.ID), []byte("test2")},
	}
	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			subChan := make(chan *nats.Msg)
			subs, err := rpcClient.conn.ChanSubscribe(table.topic, subChan)
			assert.NoError(t, err)
			// TODO this is ugly, can lead to flaky tests and we could probably do it better
			time.Sleep(50 * time.Millisecond)

			err = rpcClient.Send(table.topic, table.data)
			assert.NoError(t, err)

			r := helpers.ShouldEventuallyReceive(t, subChan).(*nats.Msg)
			assert.Equal(t, table.data, r.Data)
			subs.Unsubscribe()
		})
	}
}

func TestNatsRPCClientBuildRequest(t *testing.T) {
	config := config.NewDefaultConfig().Cluster.RPC.Client.Nats
	sv := getServer()
	rpcClient, _ := NewNatsRPCClient(config, sv, nil, nil)

	rt := route.NewRoute("sv", "svc", "method")

	data := []byte("data")
	messageID := uint(123)
	tables := []struct {
		name           string
		frontendServer bool
		route          *route.Route
		msg            *message.Message
		expected       protos.Request
	}{
		{
			"test-frontend-request", true, rt,
			&message.Message{Type: message.Request, ID: messageID, Data: data},
			protos.Request{
				Msg: &protos.Msg{
					Route: rt.String(),
					Data:  data,
					Type:  protos.MsgType_MsgRequest,
					Id:    uint64(messageID),
				},
			},
		},
		{
			"test-rpc-sys-request", false, rt,
			&message.Message{Type: message.Request, ID: messageID, Data: data},
			protos.Request{
				Msg: &protos.Msg{
					Route: rt.String(),
					Data:  data,
					Type:  protos.MsgType_MsgRequest,
					Id:    uint64(messageID),
				},
			},
		},
		{
			"test-rpc-user-request", false, rt,
			&message.Message{Type: message.Request, ID: messageID, Data: data},
			protos.Request{
				Msg: &protos.Msg{
					Route: rt.String(),
					Data:  data,
					Type:  protos.MsgType_MsgRequest,
				},
			},
		},
		{
			"test-notify", false, rt,
			&message.Message{Type: message.Notify, ID: messageID, Data: data},
			protos.Request{
				Msg: &protos.Msg{
					Route: rt.String(),
					Data:  data,
					Type:  protos.MsgType_MsgNotify,
					Id:    0,
				},
			},
		},
	}
	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			req, err := buildRequest(context.Background(), table.route, table.msg, rpcClient.server)
			assert.NoError(t, err)
			assert.NotNil(t, req.Metadata)
			req.Metadata = nil
			assert.Equal(t, table.expected, req)
		})
	}
}

func TestNatsRPCClientCallShouldFailIfNotRunning(t *testing.T) {
	config := config.NewDefaultConfig().Cluster.RPC.Client.Nats
	sv := getServer()
	rpcClient, _ := NewNatsRPCClient(config, sv, nil, nil)
	res, err := rpcClient.Call(context.Background(), nil, nil, sv)
	assert.Equal(t, constants.ErrRPCClientNotInitialized, err)
	assert.Nil(t, res)
}

func TestNatsRPCClientCall(t *testing.T) {
	s := helpers.GetTestNatsServer(t)
	sv := getServer()
	defer s.Shutdown()
	cfg := config.NewDefaultConfig().Cluster.RPC.Client.Nats
	cfg.Connect = fmt.Sprintf("nats://%s", s.Addr())
	cfg.RequestTimeout = time.Duration(300 * time.Millisecond)
	rpcClient, _ := NewNatsRPCClient(cfg, sv, nil, nil)
	rpcClient.Init()

	rt := route.NewRoute("sv", "svc", "method")

	msg := &message.Message{
		Type: message.Request,
		ID:   uint(123),
		Data: []byte("data"),
	}

	tables := []struct {
		name     string
		response interface{}
		expected *protos.Response
		err      error
	}{
		{"test_error", &protos.Response{Data: []byte("nok"), Error: &protos.Error{Msg: "nok"}}, nil, e.NewError(errors.New("nok"), e.ErrUnknownCode)},
		{"test_ok", &protos.Response{Data: []byte("ok")}, &protos.Response{Data: []byte("ok")}, nil},
		{"test_bad_response", []byte("invalid"), nil, errors.New("cannot parse invalid wire-format data")},
		{"test_bad_proto", &protos.Request{Metadata: []byte("bad proto")}, nil, errors.New("cannot parse invalid wire-format data")},
		{"test_no_response", nil, nil, constants.ErrRPCRequestTimeout},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			conn, err := setupNatsConn(fmt.Sprintf("nats://%s", s.Addr()), nil)
			defer conn.Close()
			assert.NoError(t, err)

			sv2 := getServer()
			sv2.Type = uuid.New().String()
			sv2.ID = uuid.New().String()
			subs, err := conn.Subscribe(getChannel(sv2.Type, sv2.ID), func(m *nats.Msg) {
				if table.response != nil {
					if val, ok := table.response.(*protos.Response); ok {
						b, _ := proto.Marshal(val)
						conn.Publish(m.Reply, b)
					} else if val, ok := table.response.(*protos.Request); ok {
						b, _ := proto.Marshal(val)
						conn.Publish(m.Reply, b)
					} else {
						conn.Publish(m.Reply, table.response.([]byte))
					}
				}
			})
			assert.NoError(t, err)
			// TODO this is ugly, can lead to flaky tests and we could probably do it better
			time.Sleep(50 * time.Millisecond)

			res, err := rpcClient.Call(context.Background(), rt, msg, sv2)
			assert.Equal(t, table.expected, res)
			if table.err != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), table.err.Error())
			}
			err = subs.Unsubscribe()
			assert.NoError(t, err)
			conn.Close()
		})
	}
}
