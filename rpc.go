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
	"github.com/seanbit/nrpc/conn/message"
	"github.com/seanbit/nrpc/constants"
	"github.com/seanbit/nrpc/route"
	"google.golang.org/protobuf/proto"
	"reflect"
)

// RPC calls a method in a different server
func (app *App) RPC(ctx context.Context, routeStr string, reply proto.Message, arg proto.Message) error {
	return app.doSendRPC(ctx, "", routeStr, message.Request, reply, arg)
}

// RPCTo send a rpc to a specific server
func (app *App) RPCTo(ctx context.Context, serverID, routeStr string, reply proto.Message, arg proto.Message) error {
	return app.doSendRPC(ctx, serverID, routeStr, message.Request, reply, arg)
}

// Notify calls a method in a different server
func (app *App) Notify(ctx context.Context, routeStr string, arg proto.Message) error {
	return app.doSendRPC(ctx, "", routeStr, message.Notify, nil, arg)
}

// NotifyTo send a rpc to a specific server
func (app *App) NotifyTo(ctx context.Context, serverID, routeStr string, arg proto.Message) error {
	return app.doSendRPC(ctx, serverID, routeStr, message.Notify, nil, arg)
}

func (app *App) doSendRPC(ctx context.Context, serverID, routeStr string, mt message.Type, reply proto.Message, arg proto.Message) error {
	if app.rpcServer == nil {
		return constants.ErrRPCServerNotInitialized
	}

	if mt == message.Request {
		if reply == nil || reflect.TypeOf(reply).Kind() != reflect.Ptr {
			return constants.ErrReplyShouldBePtr
		}
	}

	r, err := route.Decode(routeStr)
	if err != nil {
		return err
	}

	if r.SvType == "" {
		return constants.ErrNoServerTypeChosenForRPC
	}

	if (r.SvType == app.server.Type && serverID == "") || serverID == app.server.ID {
		return constants.ErrNonsenseRPC
	}

	return app.remoteService.RPC(ctx, serverID, mt, r, reply, arg)
}
