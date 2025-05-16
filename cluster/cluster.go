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

	"github.com/seanbit/nrpc/conn/message"
	"github.com/seanbit/nrpc/constants"
	pcontext "github.com/seanbit/nrpc/context"
	"github.com/seanbit/nrpc/interfaces"
	"github.com/seanbit/nrpc/logger"
	"github.com/seanbit/nrpc/protos"
	"github.com/seanbit/nrpc/route"
	"github.com/seanbit/nrpc/tracing"
)

type IHandler interface {
	Call(ctx context.Context, in *protos.Request) (*protos.Response, error)
}

// RPCServer interface
type RPCServer interface {
	interfaces.Module
	SetHandler(handler IHandler)
}

// RPCClient interface
type RPCClient interface {
	Send(route string, data []byte) error
	Call(ctx context.Context, route *route.Route, msg *message.Message, server *Server) (*protos.Response, error)
	interfaces.Module
}

// SDListener interface
type SDListener interface {
	AddServer(*Server)
	RemoveServer(*Server)
}

// InfoRetriever gets cluster info
// It can be implemented, for exemple, by reading
// env var, config or by accessing the cluster API
type InfoRetriever interface {
	Region() string
}

// Action type for enum
type Action int

// Action values
const (
	ADD Action = iota
	DEL
)

func buildRequest(
	ctx context.Context,
	route *route.Route,
	msg *message.Message,
	thisServer *Server,
) (protos.Request, error) {
	req := protos.Request{
		Msg: &protos.Msg{
			Route: route.String(),
			Data:  msg.Data,
		},
	}
	ctx, err := tracing.InjectSpan(ctx)
	if err != nil {
		logger.Log.Errorf("failed to inject span: %s", err)
	}
	ctx = pcontext.AddToPropagateCtx(ctx, constants.PeerIDKey, thisServer.ID)
	ctx = pcontext.AddToPropagateCtx(ctx, constants.PeerServiceKey, thisServer.Type)
	req.Metadata, err = pcontext.Encode(ctx)
	if err != nil {
		return req, err
	}

	switch msg.Type {
	case message.Request:
		req.Msg.Type = protos.MsgType_MsgRequest
	case message.Notify:
		req.Msg.Type = protos.MsgType_MsgNotify
	}
	return req, nil
}
