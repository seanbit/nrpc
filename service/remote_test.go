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

package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/seanbit/nrpc/cluster"
	clustermocks "github.com/seanbit/nrpc/cluster/mocks"
	"github.com/seanbit/nrpc/component"
	"github.com/seanbit/nrpc/conn/codec"
	"github.com/seanbit/nrpc/conn/message"
	messagemocks "github.com/seanbit/nrpc/conn/message/mocks"
	"github.com/seanbit/nrpc/constants"
	e "github.com/seanbit/nrpc/errors"
	"github.com/seanbit/nrpc/pipeline"
	"github.com/seanbit/nrpc/protos"
	"github.com/seanbit/nrpc/protos/test"
	"github.com/seanbit/nrpc/route"
	"github.com/seanbit/nrpc/router"
	serializemocks "github.com/seanbit/nrpc/serialize/mocks"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"reflect"
	"testing"
)

const ctxModifiedResponse = "response"

type MyComp struct {
	component.Base
}

func (m *MyComp) Init()                        {}
func (m *MyComp) Shutdown()                    {}
func (m *MyComp) Handler1(ctx context.Context) {}
func (m *MyComp) Handler2(ctx context.Context, b []byte) ([]byte, error) {
	return nil, nil
}
func (m *MyComp) HandlerRawRaw(ctx context.Context, b []byte) ([]byte, error) {
	return b, nil
}

func (m *MyComp) Remote1(ctx context.Context, ss *test.SomeStruct) (*test.SomeStruct, error) {
	return &test.SomeStruct{B: "ack"}, nil
}
func (m *MyComp) Remote2(ctx context.Context) (*test.SomeStruct, error) {
	return nil, nil
}
func (m *MyComp) RemoteRes(ctx context.Context, b *test.SomeStruct) (*test.SomeStruct, error) {
	ctxRes := ctx.Value(ctxModifiedResponse) // used in hook tests
	if ctxRes != nil {
		return ctxRes.(*test.SomeStruct), nil
	}
	return b, nil
}
func (m *MyComp) RemoteErr(ctx context.Context) (*test.SomeStruct, error) {
	return nil, e.NewError(errors.New("remote err"), e.ErrUnknownCode)
}

type NoHandlerRemoteComp struct {
	component.Base
}

func (m *NoHandlerRemoteComp) Init()     {}
func (m *NoHandlerRemoteComp) Shutdown() {}

type unregisteredStruct struct{}

func TestNewRemoteService(t *testing.T) {
	packetEncoder := codec.NewPomeloPacketEncoder()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSerializer := serializemocks.NewMockSerializer(ctrl)
	mockSD := clustermocks.NewMockServiceDiscovery(ctrl)
	mockRPCClient := clustermocks.NewMockRPCClient(ctrl)
	mockRPCServer := clustermocks.NewMockRPCServer(ctrl)
	mockMessageEncoder := messagemocks.NewMockEncoder(ctrl)
	router := router.New()
	sv := &cluster.Server{}
	remoteHooks := pipeline.NewRemoteHooks()
	svc := NewRemoteService(mockRPCClient, mockRPCServer, mockSD, packetEncoder, mockSerializer, router, mockMessageEncoder, sv, remoteHooks)

	assert.NotNil(t, svc)
	assert.Empty(t, svc.services)
	assert.Equal(t, mockRPCClient, svc.rpcClient)
	assert.Equal(t, mockRPCServer, svc.rpcServer)
	assert.Equal(t, packetEncoder, svc.encoder)
	assert.Equal(t, mockSD, svc.serviceDiscovery)
	assert.Equal(t, mockSerializer, svc.serializer)
	assert.Equal(t, router, svc.router)
	assert.Equal(t, sv, svc.server)
	assert.Equal(t, remoteHooks, svc.remoteHooks)
}

func TestRemoteServiceRegister(t *testing.T) {
	svc := NewRemoteService(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	err := svc.Register(&MyComp{}, []component.Option{})
	assert.NoError(t, err)
	defer func() { svc.remotes = make(map[string]*component.Remote, 0) }()
	assert.Len(t, svc.services, 1)
	val, ok := svc.services["MyComp"]
	assert.True(t, ok)
	assert.NotNil(t, val)
	val2, ok := svc.remotes["MyComp.Remote1"]
	assert.True(t, ok)
	assert.NotNil(t, val2)
	val2, ok = svc.remotes["MyComp.Remote2"]
	assert.True(t, ok)
	assert.NotNil(t, val2)
	val2, ok = svc.remotes["MyComp.RemoteErr"]
	assert.True(t, ok)
	assert.NotNil(t, val)
	val2, ok = svc.remotes["MyComp.RemoteRes"]
	assert.True(t, ok)
	assert.NotNil(t, val)
}

func TestRemoteServiceRegisterFailsIfRegisterTwice(t *testing.T) {
	svc := NewRemoteService(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	err := svc.Register(&MyComp{}, []component.Option{})
	assert.NoError(t, err)
	err = svc.Register(&MyComp{}, []component.Option{})
	assert.Contains(t, err.Error(), "remote: service already defined")
}

func TestRemoteServiceRegisterFailsIfNoRemoteMethods(t *testing.T) {
	svc := NewRemoteService(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	err := svc.Register(&NoHandlerRemoteComp{}, []component.Option{})
	assert.Equal(t, errors.New("type NoHandlerRemoteComp has no exported methods of remote type"), err)
}

func TestRemoteServiceRemoteCallWithDifferentServerArguments(t *testing.T) {
	route := route.NewRoute("sv", "svc", "method")
	table := []struct {
		name           string
		serverArg      *cluster.Server
		routeServer    *cluster.Server
		expectedServer *cluster.Server
	}{
		{
			name:           "should use server argument if provided",
			serverArg:      &cluster.Server{Type: "sv"},
			routeServer:    &cluster.Server{Type: "sv2"},
			expectedServer: &cluster.Server{Type: "sv"},
		},
		{
			name:           "should use route's returned server if server argument is nil",
			serverArg:      nil,
			routeServer:    &cluster.Server{Type: "sv"},
			expectedServer: &cluster.Server{Type: "sv"},
		},
	}

	for _, row := range table {
		t.Run(row.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRPCClient := clustermocks.NewMockRPCClient(ctrl)
			mockServiceDiscovery := clustermocks.NewMockServiceDiscovery(ctrl)
			router := router.New()
			router.SetServiceDiscovery(mockServiceDiscovery)
			mockServiceDiscovery.EXPECT().GetServersByType(gomock.Any()).Return(map[string]*cluster.Server{row.routeServer.Type: row.routeServer}, nil).AnyTimes()

			msg := &message.Message{}
			ctx := context.Background()
			mockRPCClient.EXPECT().Call(ctx, gomock.Any(), msg, row.expectedServer).Return(nil, nil).AnyTimes()

			svc := NewRemoteService(mockRPCClient, nil, nil, nil, nil, router, nil, nil, nil)
			assert.NotNil(t, svc)

			_, err := svc.remoteCall(ctx, row.serverArg, route, msg)
			assert.NoError(t, err)
		})
	}
}

func TestRemoteServiceRemoteCall(t *testing.T) {
	tables := []struct {
		name        string
		route       route.Route
		serverArg   *cluster.Server
		routeErr    error
		callRes     *protos.Response
		callErr     error
		expectedRes *protos.Response
		expectedErr error
	}{
		{
			name:        "should return internal error for routing generic error",
			route:       *route.NewRoute("sv", "svc", "method"),
			serverArg:   nil,
			routeErr:    assert.AnError,
			callRes:     nil,
			callErr:     nil,
			expectedRes: nil,
			expectedErr: e.NewError(assert.AnError, e.ErrInternalCode),
		},
		{
			name:        "should propagate error for routing pitaya error",
			route:       *route.NewRoute("sv", "svc", "method"),
			serverArg:   nil,
			routeErr:    e.NewError(assert.AnError, "CUSTOM-123"),
			callRes:     nil,
			callErr:     nil,
			expectedRes: nil,
			expectedErr: e.NewError(assert.AnError, "CUSTOM-123"),
		},
		{
			name:        "should propagate error for routing wrapped pitaya error",
			route:       *route.NewRoute("sv", "svc", "method"),
			serverArg:   nil,
			routeErr:    fmt.Errorf("wrapper error: %w", e.NewError(assert.AnError, "CUSTOM-123")),
			callRes:     nil,
			callErr:     nil,
			expectedRes: nil,
			expectedErr: e.NewError(assert.AnError, "CUSTOM-123"),
		},
		{
			name:        "should return error for rpc call error",
			route:       *route.NewRoute("sv", "svc", "method"),
			serverArg:   &cluster.Server{Type: "sv"},
			routeErr:    nil,
			callRes:     nil,
			callErr:     assert.AnError,
			expectedRes: nil,
			expectedErr: assert.AnError,
		},
		{
			name:        "should succeed",
			route:       *route.NewRoute("sv", "svc", "method"),
			serverArg:   &cluster.Server{Type: "sv"},
			routeErr:    nil,
			callRes:     &protos.Response{Data: []byte("ok")},
			callErr:     nil,
			expectedRes: &protos.Response{Data: []byte("ok")},
			expectedErr: nil,
		},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRPCClient := clustermocks.NewMockRPCClient(ctrl)
			mockServiceDiscovery := clustermocks.NewMockServiceDiscovery(ctrl)
			router := router.New()
			router.SetServiceDiscovery(mockServiceDiscovery)
			mockServiceDiscovery.EXPECT().GetServersByType(table.route.SvType).Return(map[string]*cluster.Server{"sv": {Type: "sv"}}, nil).AnyTimes()

			router.AddRoute(table.route.SvType, func(ctx context.Context, route *route.Route, payload []byte, servers map[string]*cluster.Server) (context.Context, *cluster.Server, error) {
				return ctx, &cluster.Server{}, table.routeErr
			})
			svc := NewRemoteService(mockRPCClient, nil, nil, nil, nil, router, nil, nil, nil)
			assert.NotNil(t, svc)

			mockRPCClient.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(table.callRes, table.callErr).AnyTimes()

			ctx := context.Background()
			msg := &message.Message{}
			res, err := svc.remoteCall(ctx, table.serverArg, &table.route, msg)
			assert.Equal(t, table.expectedErr, err)
			assert.Equal(t, table.expectedRes, res)
		})
	}
}

func TestRemoteServiceHandleRPC(t *testing.T) {

	tObj := &MyComp{}
	m, ok := reflect.TypeOf(tObj).MethodByName("Remote1")
	assert.True(t, ok)
	assert.NotNil(t, m)
	rt := route.NewRoute("", uuid.New().String(), uuid.New().String())
	comp := &component.Remote{Receiver: reflect.ValueOf(tObj), Method: m, HasArgs: m.Type.NumIn() > 2}

	m, ok = reflect.TypeOf(tObj).MethodByName("RemoteErr")
	assert.True(t, ok)
	assert.NotNil(t, m)
	rtErr := route.NewRoute("", uuid.New().String(), uuid.New().String())
	compErr := &component.Remote{Receiver: reflect.ValueOf(tObj), Method: m, HasArgs: m.Type.NumIn() > 2}

	m, ok = reflect.TypeOf(tObj).MethodByName("Remote2")
	assert.True(t, ok)
	assert.NotNil(t, m)
	rtStr := route.NewRoute("", uuid.New().String(), uuid.New().String())
	compStr := &component.Remote{Receiver: reflect.ValueOf(tObj), Method: m, HasArgs: m.Type.NumIn() > 2}

	m, ok = reflect.TypeOf(tObj).MethodByName("RemoteRes")
	assert.True(t, ok)
	assert.NotNil(t, m)
	rtRes := route.NewRoute("", uuid.New().String(), uuid.New().String())
	compRes := &component.Remote{Receiver: reflect.ValueOf(tObj), Method: m, HasArgs: m.Type.NumIn() > 2, Type: reflect.TypeOf(&test.SomeStruct{B: "aa"})}

	b, err := proto.Marshal(&test.SomeStruct{B: "aa"})
	assert.NoError(t, err)
	tables := []struct {
		name         string
		req          *protos.Request
		rt           *route.Route
		errSubstring string
	}{
		{"remote_not_found", &protos.Request{Msg: &protos.Msg{}}, route.NewRoute("bla", "bla", "bla"), "route not found"},
		{"failed_unmarshal", &protos.Request{Msg: &protos.Msg{Data: []byte("dd")}}, rt, "reflect: Call using zero Value argument"},
		{"failed_pcall", &protos.Request{Msg: &protos.Msg{}}, rtErr, "remote err"},
		{"success_nil_response", &protos.Request{Msg: &protos.Msg{}}, rtStr, ""},
		{"success_response", &protos.Request{Msg: &protos.Msg{Data: b}}, rtRes, ""},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			packetEncoder := codec.NewPomeloPacketEncoder()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockSerializer := serializemocks.NewMockSerializer(ctrl)
			mockSD := clustermocks.NewMockServiceDiscovery(ctrl)
			mockRPCClient := clustermocks.NewMockRPCClient(ctrl)
			mockRPCServer := clustermocks.NewMockRPCServer(ctrl)
			messageEncoder := message.NewMessagesEncoder(false)
			router := router.New()
			svc := NewRemoteService(mockRPCClient, mockRPCServer, mockSD, packetEncoder, mockSerializer, router, messageEncoder, &cluster.Server{}, pipeline.NewRemoteHooks())

			svc.remotes[rt.Short()] = comp
			svc.remotes[rtErr.Short()] = compErr
			svc.remotes[rtStr.Short()] = compStr
			svc.remotes[rtRes.Short()] = compRes

			assert.NotNil(t, svc)
			res := svc.handleRPC(context.Background(), table.req, table.rt)
			assert.NoError(t, err)
			if table.errSubstring != "" {
				assert.Contains(t, res.Error.Msg, table.errSubstring)
			} else if table.req.Msg.Data != nil {
				assert.NotNil(t, res.Data)
			}
		})
	}
}

func TestRemoteServiceHandleRPCUserWithHooks(t *testing.T) {

	tObj := &MyComp{}
	m, ok := reflect.TypeOf(tObj).MethodByName("Remote1")
	assert.True(t, ok)
	assert.NotNil(t, m)
	rt := route.NewRoute("", uuid.New().String(), uuid.New().String())
	comp := &component.Remote{Receiver: reflect.ValueOf(tObj), Method: m, HasArgs: m.Type.NumIn() > 2}

	m, ok = reflect.TypeOf(tObj).MethodByName("RemoteErr")
	assert.True(t, ok)
	assert.NotNil(t, m)
	rtErr := route.NewRoute("", uuid.New().String(), uuid.New().String())
	compErr := &component.Remote{Receiver: reflect.ValueOf(tObj), Method: m, HasArgs: m.Type.NumIn() > 2}

	m, ok = reflect.TypeOf(tObj).MethodByName("Remote2")
	assert.True(t, ok)
	assert.NotNil(t, m)
	rtStr := route.NewRoute("", uuid.New().String(), uuid.New().String())
	compStr := &component.Remote{Receiver: reflect.ValueOf(tObj), Method: m, HasArgs: m.Type.NumIn() > 2}

	m, ok = reflect.TypeOf(tObj).MethodByName("RemoteRes")
	assert.True(t, ok)
	assert.NotNil(t, m)
	rtRes := route.NewRoute("", uuid.New().String(), uuid.New().String())
	compRes := &component.Remote{Receiver: reflect.ValueOf(tObj), Method: m, HasArgs: m.Type.NumIn() > 2, Type: reflect.TypeOf(&test.SomeStruct{B: "aa"})}

	b, err := proto.Marshal(&test.SomeStruct{B: "aa"})
	assert.NoError(t, err)

	modifiedInput := &test.SomeStruct{B: "cc"}

	modifiedResponse, err := proto.Marshal(modifiedInput)
	assert.NoError(t, err)

	modifiedCtx := context.WithValue(context.Background(), ctxModifiedResponse, modifiedInput)
	tables := []struct {
		name                string
		req                 *protos.Request
		rt                  *route.Route
		expectedOutput      []byte
		errSubstring        string
		shouldRunBeforeHook bool
		shouldRunAfterHook  bool
		modifiedInput       interface{}
		modifiedCtx         context.Context
		modifiedInputError  error
		modifiedOutput      interface{}
		modifiedOutputError error
	}{
		{"remote_not_found", &protos.Request{Msg: &protos.Msg{}}, route.NewRoute("bla", "bla", "bla"), nil, "route not found", false, false, nil, nil, nil, nil, nil},
		{"failed_unmarshal", &protos.Request{Msg: &protos.Msg{Data: []byte("dd")}}, rt, nil, "reflect: Call using zero Value argument", true, true, nil, nil, nil, nil, nil},
		{"failed_pcall", &protos.Request{Msg: &protos.Msg{}}, rtErr, nil, "remote err", true, true, nil, nil, nil, nil, nil},
		{"failed_before_hook", &protos.Request{Msg: &protos.Msg{}}, rtErr, nil, "before hook err", true, false, nil, nil, fmt.Errorf("before hook err"), nil, nil},
		{"failed_pcall_modified_err", &protos.Request{Msg: &protos.Msg{}}, rtErr, nil, "remote err modified output", true, true, nil, nil, nil, nil, fmt.Errorf("remote err modified output")},
		{"success_nil_response", &protos.Request{Msg: &protos.Msg{}}, rtStr, nil, "", true, true, nil, nil, nil, nil, nil},
		{"success_response", &protos.Request{Msg: &protos.Msg{Data: b}}, rtRes, b, "", true, true, nil, nil, nil, nil, nil},
		{"success_response_modified_ctx", &protos.Request{Msg: &protos.Msg{Data: b}}, rtRes, modifiedResponse, "", true, true, nil, modifiedCtx, nil, nil, nil},
		{"success_response_modified_input", &protos.Request{Msg: &protos.Msg{Data: b}}, rtRes, modifiedResponse, "", true, true, modifiedInput, nil, nil, nil, nil},
		{"success_response_modified_input_ctx", &protos.Request{Msg: &protos.Msg{Data: b}}, rtRes, modifiedResponse, "", true, true, modifiedInput, modifiedCtx, nil, nil, nil},
		{"success_response_modified_output", &protos.Request{Msg: &protos.Msg{Data: b}}, rtRes, modifiedResponse, "", true, true, nil, nil, nil, modifiedInput, nil},
		{"failed_after_hook", &protos.Request{Msg: &protos.Msg{Data: b}}, rtRes, nil, "after hook err", true, true, nil, nil, nil, nil, fmt.Errorf("after hook err")},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			packetEncoder := codec.NewPomeloPacketEncoder()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockSerializer := serializemocks.NewMockSerializer(ctrl)
			mockSD := clustermocks.NewMockServiceDiscovery(ctrl)
			mockRPCClient := clustermocks.NewMockRPCClient(ctrl)
			mockRPCServer := clustermocks.NewMockRPCServer(ctrl)
			messageEncoder := message.NewMessagesEncoder(false)
			router := router.New()

			beforeHookInvoked := false
			afterHookInvoked := false

			remoteHooks := pipeline.NewRemoteHooks()
			remoteHooks.BeforeHandler.PushFront(func(ctx context.Context, in interface{}) (context.Context, interface{}, error) {
				if beforeHookInvoked {
					assert.FailNow(t, "BeforeHandler hook invoked twice")
				}
				if afterHookInvoked {
					assert.FailNow(t, "BeforeHandler and AfterHandler hooks running out of order")
				}

				var err error
				if table.modifiedInput != nil {
					in = table.modifiedInput
				}
				if table.modifiedCtx != nil {
					ctx = table.modifiedCtx
				}
				if table.modifiedInputError != nil {
					err = table.modifiedInputError
				}

				beforeHookInvoked = true
				return ctx, in, err
			})
			remoteHooks.AfterHandler.PushFront(func(ctx context.Context, out interface{}, err error) (interface{}, error) {
				if afterHookInvoked {
					assert.FailNow(t, "AfterHandler hook invoked twice")
				}
				if !beforeHookInvoked {
					assert.FailNow(t, "BeforeHandler and AfterHandler hooks running out of order")
				}

				if table.modifiedOutput != nil {
					out = table.modifiedOutput
				}
				if table.modifiedOutputError != nil {
					err = table.modifiedOutputError
				}

				afterHookInvoked = true
				return out, err
			})

			svc := NewRemoteService(mockRPCClient, mockRPCServer, mockSD, packetEncoder, mockSerializer, router, messageEncoder, &cluster.Server{}, remoteHooks)

			svc.remotes[rt.Short()] = comp
			svc.remotes[rtErr.Short()] = compErr
			svc.remotes[rtStr.Short()] = compStr
			svc.remotes[rtRes.Short()] = compRes

			assert.NotNil(t, svc)

			assert.False(t, beforeHookInvoked, "Before hook invoked before RPC")
			assert.False(t, afterHookInvoked, "After hook invoked before RPC")

			res := svc.handleRPC(context.Background(), table.req, table.rt)

			if table.shouldRunBeforeHook {
				assert.True(t, beforeHookInvoked, "After hook was never invoked")
			} else {
				assert.False(t, beforeHookInvoked, "After hook should not have run")
			}
			if table.shouldRunAfterHook {
				assert.True(t, afterHookInvoked, "After hook was never invoked")
			} else {
				assert.False(t, afterHookInvoked, "After hook should not have run")
			}

			assert.NoError(t, err)
			if table.errSubstring != "" {
				assert.Contains(t, res.Error.Msg, table.errSubstring)
			} else if table.req.Msg.Data != nil {
				assert.NotNil(t, res.Data)
			}

			assert.Equal(t, res.Data, table.expectedOutput)
		})
	}
}

func TestRemoteServiceRPC(t *testing.T) {
	rt := route.NewRoute("sv", "svc", "method")
	tables := []struct {
		name        string
		serverID    string
		reply       proto.Message
		arg         proto.Message
		foundServer bool
		err         error
	}{
		{"server_id_and_no_target", "serverId", nil, &test.SomeStruct{}, false, constants.ErrServerNotFound},
		{"failed_remote_call", "serverId", nil, &test.SomeStruct{}, true, errors.New("rpc failed")},
		{"success", "serverId", &test.SomeStruct{}, &test.SomeStruct{}, true, nil},
		{"success_nil_reply", "serverId", nil, &test.SomeStruct{}, true, nil},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			packetEncoder := codec.NewPomeloPacketEncoder()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockSerializer := serializemocks.NewMockSerializer(ctrl)
			mockSD := clustermocks.NewMockServiceDiscovery(ctrl)
			mockRPCClient := clustermocks.NewMockRPCClient(ctrl)
			mockRPCServer := clustermocks.NewMockRPCServer(ctrl)
			messageEncoder := message.NewMessagesEncoder(false)
			router := router.New()
			svc := NewRemoteService(mockRPCClient, mockRPCServer, mockSD, packetEncoder, mockSerializer, router, messageEncoder, &cluster.Server{}, nil)
			assert.NotNil(t, svc)

			if table.serverID != "" {
				var sdRet *cluster.Server
				if table.foundServer {
					sdRet = &cluster.Server{}
				}
				mockSD.EXPECT().GetServer(table.serverID).Return(sdRet, nil)
			}

			var expected *test.SomeStruct
			ctx := context.Background()
			if table.foundServer {
				expectedData, _ := proto.Marshal(table.arg)
				expectedMsg := &message.Message{
					Type:  message.Request,
					Route: rt.Short(),
					Data:  expectedData,
				}

				expected = &test.SomeStruct{}
				b, err := proto.Marshal(expected)
				assert.NoError(t, err)
				mockRPCClient.EXPECT().Call(ctx, rt, expectedMsg, gomock.Any()).Return(&protos.Response{Data: b}, table.err)
			}
			err := svc.RPC(ctx, table.serverID, rt, table.reply, table.arg)
			assert.Equal(t, table.err, err)
			if table.reply != nil {
				//assert.Equal(t, table.reply, expected)
				assert.True(t, proto.Equal(table.reply, expected))
			}
		})
	}
}
