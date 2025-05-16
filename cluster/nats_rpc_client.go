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
	"github.com/nats-io/nuid"
	e "github.com/seanbit/nrpc/errors"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/opentracing/opentracing-go"
	"github.com/seanbit/nrpc/config"
	"github.com/seanbit/nrpc/conn/message"
	"github.com/seanbit/nrpc/constants"
	pcontext "github.com/seanbit/nrpc/context"
	"github.com/seanbit/nrpc/logger"
	"github.com/seanbit/nrpc/metrics"
	"github.com/seanbit/nrpc/protos"
	"github.com/seanbit/nrpc/route"
	"github.com/seanbit/nrpc/tracing"
	"google.golang.org/protobuf/proto"
)

// NatsRPCClient struct
type NatsRPCClient struct {
	conn                   *nats.Conn
	connString             string
	connectionTimeout      time.Duration
	maxReconnectionRetries int
	reqTimeout             time.Duration
	running                bool
	server                 *Server
	metricsReporters       []metrics.Reporter
	appDieChan             chan bool
	websocketCompression   bool
	reconnectJitter        time.Duration
	reconnectJitterTLS     time.Duration
	reconnectWait          time.Duration
	pingInterval           time.Duration
	maxPingsOutstanding    int
}

// NewNatsRPCClient ctor
func NewNatsRPCClient(
	config config.NatsRPCClientConfig,
	server *Server,
	metricsReporters []metrics.Reporter,
	appDieChan chan bool,
) (*NatsRPCClient, error) {
	ns := &NatsRPCClient{
		server:            server,
		running:           false,
		metricsReporters:  metricsReporters,
		appDieChan:        appDieChan,
		connectionTimeout: nats.DefaultTimeout,
	}
	if err := ns.configure(config); err != nil {
		return nil, err
	}
	return ns, nil
}

func (ns *NatsRPCClient) configure(config config.NatsRPCClientConfig) error {
	ns.connString = config.Connect
	if ns.connString == "" {
		return constants.ErrNoNatsConnectionString
	}
	ns.connectionTimeout = config.ConnectionTimeout
	ns.maxReconnectionRetries = config.MaxReconnectionRetries
	ns.reqTimeout = config.RequestTimeout
	if ns.reqTimeout == 0 {
		return constants.ErrNatsNoRequestTimeout
	}
	ns.websocketCompression = config.WebsocketCompression
	ns.reconnectJitter = config.ReconnectJitter
	ns.reconnectJitterTLS = config.ReconnectJitterTLS
	ns.reconnectWait = config.ReconnectWait
	ns.pingInterval = config.PingInterval
	ns.maxPingsOutstanding = config.MaxPingsOutstanding
	return nil
}

// Send publishes a message in a given topic
func (ns *NatsRPCClient) Send(topic string, data []byte) error {
	if !ns.running {
		return constants.ErrRPCClientNotInitialized
	}
	return ns.conn.Publish(topic, data)
}

// Call calls a method remotely
func (ns *NatsRPCClient) Call(
	ctx context.Context,
	route *route.Route,
	msg *message.Message,
	server *Server,
) (*protos.Response, error) {
	parent, err := tracing.ExtractSpan(ctx)
	if err != nil {
		logger.Log.Warnf("failed to retrieve parent span: %s", err.Error())
	}
	tags := opentracing.Tags{
		"span.kind":       "client",
		"local.id":        ns.server.ID,
		"peer.serverType": server.Type,
		"peer.id":         server.ID,
	}
	ctx = tracing.StartSpan(ctx, "NATS RPC Call", tags, parent)
	defer tracing.FinishSpan(ctx, err)

	if !ns.running {
		err = constants.ErrRPCClientNotInitialized
		return nil, err
	}

	requestID := pcontext.GetFromPropagateCtx(ctx, constants.RequestIDKey)
	if requestID == nil {
		requestID = nuid.Next()
		ctx = pcontext.AddToPropagateCtx(ctx, constants.RequestIDKey, requestID)
	} else if rID, ok := requestID.(string); ok {
		if rID == "" {
			requestID = nuid.Next()
		}
	}

	reqTimeout := pcontext.GetFromPropagateCtx(ctx, constants.RequestTimeout)
	if reqTimeout == nil {
		reqTimeout = ns.reqTimeout.String()
		ctx = pcontext.AddToPropagateCtx(ctx, constants.RequestTimeout, reqTimeout)
	}
	logger.Log.Debugf("[rpc_client] sending remote nats request for route %s with timeout of %s", route, reqTimeout)

	startTime := time.Now()
	ctx = pcontext.AddToPropagateCtx(ctx, constants.StartTimeKey, strconv.FormatInt(startTime.UnixNano(), 10))
	ctx = pcontext.AddToPropagateCtx(ctx, constants.RouteKey, route.String())

	req, err := buildRequest(ctx, route, msg, ns.server)
	if err != nil {
		return nil, err
	}
	marshalledData, err := proto.Marshal(&req)
	if err != nil {
		return nil, err
	}

	var m *nats.Msg

	if ns.metricsReporters != nil {
		defer func() {
			typ := "rpc"
			metrics.ReportTimingFromCtx(ctx, ns.metricsReporters, typ, err)
		}()
	}

	var timeout time.Duration
	switch msg.Type {
	case message.Request:
		timeout, _ = time.ParseDuration(reqTimeout.(string))
		m, err = ns.conn.Request(getChannel(server.Type, server.ID), marshalledData, timeout)
	case message.Notify:
		err = ns.conn.Publish(getChannel(server.Type, server.ID), marshalledData)
	}
	if err != nil {
		if errors.Is(err, nats.ErrTimeout) {
			err = e.NewError(constants.ErrRPCRequestTimeout, "PIT-408", map[string]string{
				"timeout": timeout.String(),
				"route":   route.String(),
				"server":  ns.server.ID,
				"peer.id": server.ID,
			})
		}
		return nil, err
	}
	res := &protos.Response{}
	if m == nil {
		return res, nil
	}
	err = proto.Unmarshal(m.Data, res)
	if err != nil {
		return nil, err
	}
	if res.Error != nil {
		if res.Error.Code == "" {
			res.Error.Code = e.ErrUnknownCode
		}
		err = &e.Error{
			Code:     res.Error.Code,
			Message:  res.Error.Msg,
			Metadata: res.Error.Metadata,
		}
		return nil, err
	}
	return res, nil
}

// Init inits nats rpc client
func (ns *NatsRPCClient) Init() error {
	ns.running = true
	logger.Log.Debugf("connecting to nats (client) with timeout of %s", ns.connectionTimeout)
	conn, err := setupNatsConn(
		ns.connString,
		ns.appDieChan,
		nats.RetryOnFailedConnect(false),
		nats.MaxReconnects(ns.maxReconnectionRetries),
		nats.Timeout(ns.connectionTimeout),
		nats.Compression(ns.websocketCompression),
		nats.ReconnectJitter(ns.reconnectJitter, ns.reconnectJitterTLS),
		nats.ReconnectWait(ns.reconnectWait),
		nats.PingInterval(ns.pingInterval),
		nats.MaxPingsOutstanding(ns.maxPingsOutstanding),
	)
	if err != nil {
		return err
	}
	ns.conn = conn
	return nil
}

// AfterInit runs after initialization
func (ns *NatsRPCClient) AfterInit() {}

// BeforeShutdown runs before shutdown
func (ns *NatsRPCClient) BeforeShutdown() {}

// Shutdown stops nats rpc server
func (ns *NatsRPCClient) Shutdown() error {
	return nil
}

func (ns *NatsRPCClient) stop() {
	ns.running = false
}

func (ns *NatsRPCClient) getSubscribeChannel() string {
	return fmt.Sprintf("nrpc/servers/%s/%s", ns.server.Type, ns.server.ID)
}
