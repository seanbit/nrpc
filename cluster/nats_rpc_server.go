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
	pcontext "github.com/seanbit/nrpc/context"
	"golang.org/x/sync/semaphore"
	"math"
	"strconv"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/seanbit/nrpc/config"
	"github.com/seanbit/nrpc/constants"
	e "github.com/seanbit/nrpc/errors"
	"github.com/seanbit/nrpc/logger"
	"github.com/seanbit/nrpc/metrics"
	"github.com/seanbit/nrpc/protos"
	"github.com/seanbit/nrpc/util"
	"google.golang.org/protobuf/proto"
)

// NatsRPCServer struct
type NatsRPCServer struct {
	service                int
	connString             string
	connectionTimeout      time.Duration
	maxReconnectionRetries int
	server                 *Server
	conn                   *nats.Conn
	messagesBufferSize     int
	stopChan               chan bool
	subChan                chan *nats.Msg // subChan is the channel used by the server to receive network messages addressed to itself
	unhandledReqCh         chan *protos.Request
	responses              []*protos.Response
	requests               []*protos.Request
	sub                    *nats.Subscription
	dropped                int
	handler                IHandler
	metricsReporters       []metrics.Reporter
	appDieChan             chan bool
	websocketCompression   bool
	reconnectJitter        time.Duration
	reconnectJitterTLS     time.Duration
	reconnectWait          time.Duration
	pingInterval           time.Duration
	maxPingsOutstanding    int
	workerPool             *WorkerPool
	sempWeighted           *semaphore.Weighted
}

// NewNatsRPCServer ctor
func NewNatsRPCServer(
	config config.NatsRPCServerConfig,
	server *Server,
	metricsReporters []metrics.Reporter,
	appDieChan chan bool,
	workerPool *WorkerPool,
) (*NatsRPCServer, error) {
	ns := &NatsRPCServer{
		server:            server,
		stopChan:          make(chan bool),
		unhandledReqCh:    make(chan *protos.Request),
		dropped:           0,
		metricsReporters:  metricsReporters,
		appDieChan:        appDieChan,
		connectionTimeout: nats.DefaultTimeout,
		workerPool:        workerPool,
	}
	if err := ns.configure(config); err != nil {
		return nil, err
	}

	return ns, nil
}

func (ns *NatsRPCServer) configure(config config.NatsRPCServerConfig) error {
	ns.service = config.Services
	ns.connString = config.Connect
	if ns.connString == "" {
		return constants.ErrNoNatsConnectionString
	}
	ns.connectionTimeout = config.ConnectionTimeout
	ns.maxReconnectionRetries = config.MaxReconnectionRetries
	ns.messagesBufferSize = config.Buffer.Messages
	if ns.messagesBufferSize == 0 {
		return constants.ErrNatsMessagesBufferSizeZero
	}
	ns.subChan = make(chan *nats.Msg, ns.messagesBufferSize)
	// the reason this channel is buffered is that we can achieve more performance by not
	// blocking producers on a massive push
	ns.responses = make([]*protos.Response, ns.service)
	ns.requests = make([]*protos.Request, ns.service)
	ns.websocketCompression = config.WebsocketCompression
	ns.reconnectJitter = config.ReconnectJitter
	ns.reconnectJitterTLS = config.ReconnectJitterTLS
	ns.reconnectWait = config.ReconnectWait
	ns.pingInterval = config.PingInterval
	ns.maxPingsOutstanding = config.MaxPingsOutstanding
	if config.Semaphore > 0 {
		ns.sempWeighted = semaphore.NewWeighted(int64(config.Semaphore))
	}
	return nil
}

// SetHandler sets the handler
func (ns *NatsRPCServer) SetHandler(handler IHandler) {
	ns.handler = handler
}

// SetWorkerPool sets the work pool
func (ns *NatsRPCServer) SetWorkerPool(pool *WorkerPool) {
	ns.workerPool = pool
}

func (ns *NatsRPCServer) handleMessages() {
	defer (func() {
		ns.conn.Drain()
		close(ns.unhandledReqCh)
		close(ns.subChan)
	})()
	maxPending := float64(0)
	for {
		select {
		case msg := <-ns.subChan:
			ns.reportMetrics()
			dropped, err := ns.sub.Dropped()
			if err != nil {
				logger.Log.Errorf("error getting number of dropped messages: %s", err.Error())
			}
			if dropped > ns.dropped {
				logger.Log.Warnf("[rpc server] some messages were dropped! numDropped: %d", dropped)
				ns.dropped = dropped
			}
			subsChanLen := float64(len(ns.subChan))
			maxPending = math.Max(float64(maxPending), subsChanLen)
			logger.Log.Debugf("subs channel size: %v, max: %v, dropped: %v", subsChanLen, maxPending, dropped)
			req := &protos.Request{}
			// TODO: Add tracing here to report delay to start processing message in spans
			err = proto.Unmarshal(msg.Data, req)
			if err != nil {
				// should answer rpc with an error
				logger.Log.Error("error unmarshalling rpc message:", err.Error())
				continue
			}
			req.Msg.Reply = msg.Reply
			ns.unhandledReqCh <- req
		case <-ns.stopChan:
			return
		}
	}
}

// GetUnhandledRequestsChannel gets the unhandled requests channel from nats rpc server
func (ns *NatsRPCServer) GetUnhandledRequestsChannel() chan *protos.Request {
	return ns.unhandledReqCh
}

func (ns *NatsRPCServer) marshalResponse(res *protos.Response) ([]byte, error) {
	p, err := proto.Marshal(res)
	if err != nil {
		logger.Log.Errorf("error marshaling response: %s", err.Error())

		res := &protos.Response{
			Error: &protos.Error{
				Code: e.ErrUnknownCode,
				Msg:  err.Error(),
			},
		}
		p, _ = proto.Marshal(res)
	}

	if err == nil && res.Error != nil {
		err = errors.New(res.Error.Msg)
	}
	return p, err
}

func (ns *NatsRPCServer) processMessages(threadID int) {
	for ns.requests[threadID] = range ns.GetUnhandledRequestsChannel() {
		logger.Log.Debugf("(%d) processing message %v", threadID, ns.requests[threadID].GetMsg().GetId())
		ctx, err := util.GetContextFromRequest(ns.requests[threadID])
		if err != nil {
			ns.responses[threadID] = &protos.Response{
				Error: &protos.Error{
					Code: e.ErrInternalCode,
					Msg:  err.Error(),
				},
			}
			logger.Log.Errorf("error getting context from request: %s", err)
			ns.processResponse(ns.requests[threadID], ns.responses[threadID], err)
			continue
		}

		reqStartTime, timeout, err := ns.parseRequestTimeout(ctx)
		if err != nil {
			ns.responses[threadID] = &protos.Response{
				Error: &protos.Error{
					Code: e.ErrInternalCode,
					Msg:  err.Error(),
				},
			}
			logger.Log.Errorf("error getting time info from request: %s", err)
			ns.processResponse(ns.requests[threadID], ns.responses[threadID], err)
			continue
		}

		shareIDVal := pcontext.GetFromPropagateCtx(ctx, constants.ShareIDKey)
		shareID, ok := shareIDVal.(string)
		if ns.workerPool != nil && ok && shareID != "" {
			req := ns.requests[threadID]
			ns.workerPool.Submit(shareID, func() {
				resp, err := ns.processRequest(ctx, req, reqStartTime, timeout)
				ns.processResponse(req, resp, err)
			})
			continue
		}
		// 信号量控制并发，替代service常驻goroutine
		go func(req *protos.Request) {
			sempCtx, cancel := context.WithTimeout(context.Background(), timeout-time.Since(reqStartTime))
			defer cancel()
			// 获取信号量
			if err := ns.sempWeighted.Acquire(sempCtx, 1); err != nil {
				resp := &protos.Response{
					Error: &protos.Error{
						Code: e.ErrTooManyRequestCode,
						Msg:  err.Error(),
					},
				}
				logger.Log.Errorf("error semaphore acquire for request: %s", err)
				ns.processResponse(req, resp, err)
				return
			}
			resp, err := ns.processRequest(ctx, req, reqStartTime, timeout)
			ns.processResponse(req, resp, err)
			ns.sempWeighted.Release(1) // 释放信号量
		}(ns.requests[threadID])
	}
}

func (ns *NatsRPCServer) parseRequestTimeout(ctx context.Context) (reqStartTime time.Time, reqTimeout time.Duration, err error) {
	reqStart := pcontext.GetFromPropagateCtx(ctx, constants.StartTimeKey)
	if reqStart == nil {
		err = errors.New("rpc server process lost start time request")
		return
	}
	startTimestamp, err := strconv.ParseInt(reqStart.(string), 10, 64)
	if err != nil {
		err = errors.New("rpc server process invalid start time request")
		return
	}
	reqTimeoutStr := pcontext.GetFromPropagateCtx(ctx, constants.RequestTimeout)
	if reqTimeoutStr == nil {
		err = errors.New("rpc server process lost timeout request")
		return
	}
	reqTimeout, err = time.ParseDuration(reqTimeoutStr.(string))
	if err != nil {
		err = fmt.Errorf("rpc server process invalid timeout request: %s", err.Error())
		return
	}
	reqStartTime = time.Unix(0, startTimestamp)
	return
}

func (ns *NatsRPCServer) processRequest(ctx context.Context, req *protos.Request,
	reqStartTime time.Time, reqTimeout time.Duration) (*protos.Response, error) {
	res := &protos.Response{}
	// 已经超过1/2的时间才进行处理，可以预设为超时了
	if time.Since(reqStartTime) > reqTimeout/2 {
		res.Error = &protos.Error{
			Code: e.ErrRequestTimeoutCode,
			Msg:  "rpc server process timeout",
		}
		return res, errors.New(res.Error.Msg)
	}
	// 超时检测
	select {
	case <-time.After(reqTimeout - time.Since(reqStartTime)):
		res.Error = &protos.Error{
			Code: e.ErrRequestTimeoutCode,
			Msg:  "rpc server process timeout",
		}
		return res, errors.New(res.Error.Msg)
	default:
		res, err := ns.handler.Call(ctx, req)
		if err != nil {
			logger.Log.Errorf("error processing route %s: %s", req.GetMsg().GetRoute(), err)
		}
		return res, err
	}
}

func (ns *NatsRPCServer) processResponse(req *protos.Request, resp *protos.Response, err error) {
	if req.Msg.Type == protos.MsgType_MsgNotify {
		return
	}
	p, err := ns.marshalResponse(resp)
	err = ns.conn.Publish(req.GetMsg().GetReply(), p)
	if err != nil {
		logger.Log.Errorf("error sending message response: %s", err.Error())
	}
}

// Init inits nats rpc server
func (ns *NatsRPCServer) Init() error {
	// TODO should we have concurrency here? it feels like we should
	go ns.handleMessages()

	logger.Log.Debugf("connecting to nats (server) with timeout of %s", ns.connectionTimeout)
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
	if ns.sub, err = ns.subscribe(getChannel(ns.server.Type, ns.server.ID)); err != nil {
		return err
	}

	//
	if ns.workerPool != nil {
		ns.workerPool.Start()
	}

	// this handles remote messages
	for i := 0; i < ns.service; i++ {
		go ns.processMessages(i)
	}
	return nil
}

// AfterInit runs after initialization
func (ns *NatsRPCServer) AfterInit() {}

// BeforeShutdown runs before shutdown
func (ns *NatsRPCServer) BeforeShutdown() {}

// Shutdown stops nats rpc server
func (ns *NatsRPCServer) Shutdown() error {
	close(ns.stopChan)
	return nil
}

func (ns *NatsRPCServer) subscribe(topic string) (*nats.Subscription, error) {
	return ns.conn.ChanSubscribe(topic, ns.subChan)
}

func (ns *NatsRPCServer) stop() {
}

func (ns *NatsRPCServer) reportMetrics() {
	if ns.metricsReporters != nil {
		for _, mr := range ns.metricsReporters {
			if err := mr.ReportGauge(metrics.DroppedMessages, map[string]string{}, float64(ns.dropped)); err != nil {
				logger.Log.Warnf("failed to report dropped message: %s", err.Error())
			}

			// subchan
			subChanCapacity := ns.messagesBufferSize - len(ns.subChan)
			if subChanCapacity == 0 {
				logger.Log.Warn("subChan is at maximum capacity")
			}
			if err := mr.ReportGauge(metrics.ChannelCapacity, map[string]string{"channel": "rpc_server_subchan"}, float64(subChanCapacity)); err != nil {
				logger.Log.Warnf("failed to report subChan queue capacity: %s", err.Error())
			}
		}
	}
}
