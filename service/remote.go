//
// Copyright (c) TFG Co. All Rights Reserved.
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
	"github.com/seanbit/nrpc/cluster"
	"github.com/seanbit/nrpc/component"
	"github.com/seanbit/nrpc/conn/codec"
	"github.com/seanbit/nrpc/conn/message"
	"github.com/seanbit/nrpc/constants"
	e "github.com/seanbit/nrpc/errors"
	"github.com/seanbit/nrpc/logger"
	"github.com/seanbit/nrpc/pipeline"
	"github.com/seanbit/nrpc/protos"
	"github.com/seanbit/nrpc/route"
	"github.com/seanbit/nrpc/router"
	"github.com/seanbit/nrpc/serialize"
	"github.com/seanbit/nrpc/tracing"
	"github.com/seanbit/nrpc/util"
	"google.golang.org/protobuf/proto"
	"reflect"
)

// RemoteService struct
type RemoteService struct {
	rpcServer        cluster.RPCServer
	serviceDiscovery cluster.ServiceDiscovery
	serializer       serialize.Serializer
	encoder          codec.PacketEncoder
	rpcClient        cluster.RPCClient
	services         map[string]*component.Service // all registered service
	router           *router.Router
	messageEncoder   message.Encoder
	server           *cluster.Server // server obj
	remoteHooks      *pipeline.RemoteHooks
	remotes          map[string]*component.Remote // all remote method
}

// NewRemoteService creates and return a new RemoteService
func NewRemoteService(
	rpcClient cluster.RPCClient,
	rpcServer cluster.RPCServer,
	sd cluster.ServiceDiscovery,
	encoder codec.PacketEncoder,
	serializer serialize.Serializer,
	router *router.Router,
	messageEncoder message.Encoder,
	server *cluster.Server,
	remoteHooks *pipeline.RemoteHooks,
) *RemoteService {
	remote := &RemoteService{
		services:         make(map[string]*component.Service),
		rpcClient:        rpcClient,
		rpcServer:        rpcServer,
		encoder:          encoder,
		serviceDiscovery: sd,
		serializer:       serializer,
		router:           router,
		messageEncoder:   messageEncoder,
		server:           server,
		remotes:          make(map[string]*component.Remote),
	}

	remote.remoteHooks = remoteHooks

	return remote
}

// Call processes a remote call
func (r *RemoteService) Call(ctx context.Context, req *protos.Request) (*protos.Response, error) {
	c, err := util.GetContextFromRequest(req)
	c = util.StartSpanFromRequest(c, r.server.ID, req.GetMsg().GetRoute())
	defer tracing.FinishSpan(c, err)
	var res *protos.Response

	if err == nil {
		return processRemoteMessage(c, req, r), nil
	}
	res = &protos.Response{
		Error: &protos.Error{
			Code: e.ErrInternalCode,
			Msg:  err.Error(),
		},
	}
	if res.Error != nil {
		err = errors.New(res.Error.Msg)
	}
	return res, err
}

//// Call processes a remote call
//func (r *RemoteService) Call(ctx context.Context, req *protos.Request) (*protos.Response, error) {
//	c, err := util.GetContextFromRequest(req, r.server.ID)
//	c = util.StartSpanFromRequest(c, r.server.ID, req.GetMsg().GetRoute())
//	defer tracing.FinishSpan(c, err)
//	var res *protos.Response
//
//	if err == nil {
//		result := make(chan *protos.Response, 1)
//		go func() {
//			result <- processRemoteMessage(c, req, r)
//		}()
//
//		reqTimeout := pcontext.GetFromPropagateCtx(ctx, constants.RequestTimeout)
//		if reqTimeout != nil {
//			var timeout time.Duration
//			timeout, err = time.ParseDuration(reqTimeout.(string))
//			if err == nil {
//				select {
//				case <-time.After(timeout):
//					err = constants.ErrRPCRequestTimeout
//				case res := <-result:
//					return res, nil
//				}
//			}
//		} else {
//			res := <-result
//			return res, nil
//		}
//	}
//
//	if err != nil {
//		res = &protos.Response{
//			Error: &protos.Error{
//				Code: e.ErrInternalCode,
//				Msg:  err.Error(),
//			},
//		}
//	}
//
//	if res.Error != nil {
//		err = errors.New(res.Error.Msg)
//	}
//
//	return res, err
//}

// DoRPC do rpc and get answer
func (r *RemoteService) DoRPC(ctx context.Context, serverID string, mt message.Type, route *route.Route, protoData []byte) (*protos.Response, error) {
	msg := &message.Message{
		Type:  mt,
		Route: route.Short(),
		Data:  protoData,
	}

	if serverID == "" {
		return r.remoteCall(ctx, nil, route, msg)
	}

	target, _ := r.serviceDiscovery.GetServer(serverID)
	if target == nil {
		return nil, constants.ErrServerNotFound
	}

	return r.remoteCall(ctx, target, route, msg)
}

// RPC makes rpcs
func (r *RemoteService) RPC(ctx context.Context, serverID string, mt message.Type, route *route.Route, reply proto.Message, arg proto.Message) error {
	var data []byte
	var err error
	if arg != nil {
		data, err = proto.Marshal(arg)
		if err != nil {
			return err
		}
	}
	res, err := r.DoRPC(ctx, serverID, mt, route, data)
	if err != nil {
		return err
	}

	if res.Error != nil {
		return &e.Error{
			Code:     res.Error.Code,
			Message:  res.Error.Msg,
			Metadata: res.Error.Metadata,
		}
	}

	if reply != nil {
		err = proto.Unmarshal(res.GetData(), reply)
		if err != nil {
			return err
		}
	}
	return nil
}

// Register registers components
func (r *RemoteService) Register(comp component.Component, opts []component.Option) error {
	s := component.NewService(comp, opts)

	if _, ok := r.services[s.Name]; ok {
		return fmt.Errorf("remote: service already defined: %s", s.Name)
	}

	if err := s.ExtractRemote(); err != nil {
		return err
	}

	r.services[s.Name] = s
	// register all remotes
	for name, remote := range s.Remotes {
		r.remotes[fmt.Sprintf("%s.%s", s.Name, name)] = remote
	}

	return nil
}

func processRemoteMessage(ctx context.Context, req *protos.Request, r *RemoteService) *protos.Response {
	rt, err := route.Decode(req.GetMsg().GetRoute())
	if err != nil {
		response := &protos.Response{
			Error: &protos.Error{
				Code: e.ErrBadRequestCode,
				Msg:  "cannot decode route",
				Metadata: map[string]string{
					"route": req.GetMsg().GetRoute(),
				},
			},
		}
		return response
	}
	return r.handleRPC(ctx, req, rt)
}

func (r *RemoteService) handleRPC(ctx context.Context, req *protos.Request, rt *route.Route) *protos.Response {
	remote, ok := r.remotes[rt.Short()]
	if !ok {
		logger.Log.Warnf("pitaya/remote: %s not found", rt.Short())
		response := &protos.Response{
			Error: &protos.Error{
				Code: e.ErrNotFoundCode,
				Msg:  "route not found",
				Metadata: map[string]string{
					"route": rt.Short(),
				},
			},
		}
		return response
	}

	var ret interface{}
	var arg interface{}
	var err error

	if remote.HasArgs {
		arg, err = unmarshalRemoteArg(remote, req.GetMsg().GetData())
		if err != nil {
			response := &protos.Response{
				Error: &protos.Error{
					Code: e.ErrBadRequestCode,
					Msg:  err.Error(),
				},
			}
			return response
		}
	}

	ctx, arg, err = r.remoteHooks.BeforeHandler.ExecuteBeforePipeline(ctx, arg)
	if err != nil {
		response := &protos.Response{
			Error: &protos.Error{
				Code: e.ErrInternalCode,
				Msg:  err.Error(),
			},
		}
		return response
	}

	params := []reflect.Value{remote.Receiver, reflect.ValueOf(ctx)}
	if remote.HasArgs {
		params = append(params, reflect.ValueOf(arg))
	}
	ret, err = util.Pcall(remote.Method, params)

	ret, err = r.remoteHooks.AfterHandler.ExecuteAfterPipeline(ctx, ret, err)
	if err != nil {
		response := &protos.Response{
			Error: &protos.Error{
				Code: e.ErrUnknownCode,
				Msg:  err.Error(),
			},
		}
		if val, ok := err.(*e.Error); ok {
			response.Error.Code = val.Code
			if val.Metadata != nil {
				response.Error.Metadata = val.Metadata
			}
		}
		return response
	}

	var b []byte
	if ret != nil {
		pb, ok := ret.(proto.Message)
		if !ok {
			response := &protos.Response{
				Error: &protos.Error{
					Code: e.ErrUnknownCode,
					Msg:  constants.ErrWrongValueType.Error(),
				},
			}
			return response
		}
		if b, err = proto.Marshal(pb); err != nil {
			response := &protos.Response{
				Error: &protos.Error{
					Code: e.ErrUnknownCode,
					Msg:  err.Error(),
				},
			}
			return response
		}
	}

	response := &protos.Response{}
	response.Data = b
	return response
}

func (r *RemoteService) remoteCall(
	ctx context.Context,
	server *cluster.Server,
	route *route.Route,
	msg *message.Message,
) (*protos.Response, error) {
	svType := route.SvType

	var err error
	target := server

	if target == nil {
		ctx, target, err = r.router.Route(ctx, svType, route, msg)
		if err != nil {
			logger.Log.Errorf("error making call for route %s: %w", route.String(), err)
			return nil, e.NewError(err, e.ErrInternalCode)
		}
	}

	res, err := r.rpcClient.Call(ctx, route, msg, target)
	if err != nil {
		logger.Log.Errorf("error making call to target with id %s, route %s and host %s: %w", target.ID, route.String(), target.Hostname, err)
		return nil, err
	}
	return res, err
}

// DumpServices outputs all registered services
func (r *RemoteService) DumpServices() {
	for name := range r.remotes {
		logger.Log.Infof("registered remote %s", name)
	}
}
