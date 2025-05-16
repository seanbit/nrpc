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
	"github.com/golang/mock/gomock"
	"github.com/seanbit/nrpc/cluster"
	clustermocks "github.com/seanbit/nrpc/cluster/mocks"
	"github.com/seanbit/nrpc/config"
	"github.com/seanbit/nrpc/conn/codec"
	"github.com/seanbit/nrpc/conn/message"
	"github.com/seanbit/nrpc/constants"
	"github.com/seanbit/nrpc/pipeline"
	"github.com/seanbit/nrpc/protos"
	"github.com/seanbit/nrpc/protos/test"
	"github.com/seanbit/nrpc/route"
	"github.com/seanbit/nrpc/router"
	serializemocks "github.com/seanbit/nrpc/serialize/mocks"
	"github.com/seanbit/nrpc/service"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"testing"
)

func TestDoSendRPCNotInitialized(t *testing.T) {
	config := config.NewDefaultConfig()
	app := NewDefaultApp("testtype", map[string]string{}, *config).(*App)
	err := app.doSendRPC(nil, "", "", nil, nil)
	assert.Equal(t, constants.ErrReplyShouldBePtr, err)
}

func TestDoSendRPC(t *testing.T) {
	config := config.NewDefaultConfig()
	app := NewDefaultApp("testtype", map[string]string{}, *config).(*App)
	app.server.ID = "myserver"
	app.rpcServer = &cluster.NatsRPCServer{}
	tables := []struct {
		name     string
		routeStr string
		reply    proto.Message
		arg      proto.Message
		err      error
	}{
		{"bad_route", "badroute", &test.SomeStruct{}, nil, route.ErrInvalidRoute},
		{"no_server_type", "bla.bla", &test.SomeStruct{}, nil, constants.ErrNoServerTypeChosenForRPC},
		{"nonsense_rpc", "mytype.bla.bla", &test.SomeStruct{}, nil, constants.ErrNonsenseRPC},
		{"success", "bla.bla.bla", &test.SomeStruct{}, &test.SomeStruct{A: 1}, nil},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			ctx := context.Background()
			if table.err == nil {
				packetEncoder := codec.NewPomeloPacketEncoder()
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				mockSerializer := serializemocks.NewMockSerializer(ctrl)
				mockSD := clustermocks.NewMockServiceDiscovery(ctrl)
				mockRPCClient := clustermocks.NewMockRPCClient(ctrl)
				mockRPCServer := clustermocks.NewMockRPCServer(ctrl)
				messageEncoder := message.NewMessagesEncoder(false)
				router := router.New()
				svc := service.NewRemoteService(mockRPCClient, mockRPCServer, mockSD, packetEncoder, mockSerializer, router, messageEncoder, &cluster.Server{}, pipeline.NewRemoteHooks())
				assert.NotNil(t, svc)
				app.remoteService = svc
				app.server.ID = "notmyserver"
				b, err := proto.Marshal(&test.SomeStruct{A: 1})
				assert.NoError(t, err)
				mockSD.EXPECT().GetServer("myserver").Return(&cluster.Server{}, nil)
				mockRPCClient.EXPECT().Call(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(&protos.Response{Data: b}, nil)
			}
			err := app.RPCTo(ctx, "myserver", table.routeStr, table.reply, table.arg)
			assert.Equal(t, table.err, err)
		})
	}
}
