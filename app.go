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
	"os"
	"os/signal"
	"syscall"

	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/seanbit/nrpc/cluster"
	"github.com/seanbit/nrpc/component"
	"github.com/seanbit/nrpc/config"
	"github.com/seanbit/nrpc/conn/message"
	"github.com/seanbit/nrpc/constants"
	pcontext "github.com/seanbit/nrpc/context"
	"github.com/seanbit/nrpc/errors"
	"github.com/seanbit/nrpc/interfaces"
	"github.com/seanbit/nrpc/logger"
	logging "github.com/seanbit/nrpc/logger/interfaces"
	"github.com/seanbit/nrpc/metrics"
	"github.com/seanbit/nrpc/router"
	"github.com/seanbit/nrpc/serialize"
	"github.com/seanbit/nrpc/service"
	"github.com/seanbit/nrpc/tracing"
	"google.golang.org/protobuf/proto"
)

// NRpc interface
type NRpc interface {
	GetDieChan() chan bool
	SetDebug(debug bool)
	SetHeartbeatTime(interval time.Duration)
	GetServerID() string
	GetMetricsReporters() []metrics.Reporter
	GetServer() *cluster.Server
	GetServerByID(id string) (*cluster.Server, error)
	GetServersByType(t string) (map[string]*cluster.Server, error)
	GetServers() []*cluster.Server
	Start()
	SetDictionary(dict map[string]uint16) error
	AddRoute(serverType string, routingFunction router.RoutingFunc) error
	Shutdown()
	IsRunning() bool

	RPC(ctx context.Context, routeStr string, reply proto.Message, arg proto.Message) error
	RPCTo(ctx context.Context, serverID, routeStr string, reply proto.Message, arg proto.Message) error
	Notify(ctx context.Context, routeStr string, arg proto.Message) error
	NotifyTo(ctx context.Context, serverID, routeStr string, arg proto.Message) error

	Register(c component.Component, options ...component.Option)

	RegisterModule(module interfaces.Module, name string) error
	RegisterModuleAfter(module interfaces.Module, name string) error
	RegisterModuleBefore(module interfaces.Module, name string) error
	GetModule(name string) (interfaces.Module, error)
}

// App is the base nrpc server struct
type App struct {
	config           config.Config
	debug            bool
	dieChan          chan bool
	heartbeat        time.Duration
	router           *router.Router
	rpcClient        cluster.RPCClient
	rpcServer        cluster.RPCServer
	metricsReporters []metrics.Reporter
	running          bool
	serializer       serialize.Serializer
	server           *cluster.Server
	serviceDiscovery cluster.ServiceDiscovery
	startAt          time.Time
	remoteService    *service.RemoteService
	remoteComp       []regComp
	modulesMap       map[string]interfaces.Module
	modulesArr       []moduleWrapper
}

// NewApp is the base constructor for a nrpc server instance
func NewApp(
	serializer serialize.Serializer,
	dieChan chan bool,
	router *router.Router,
	server *cluster.Server,
	rpcClient cluster.RPCClient,
	rpcServer cluster.RPCServer,
	serviceDiscovery cluster.ServiceDiscovery,
	remoteService *service.RemoteService,
	metricsReporters []metrics.Reporter,
	config config.Config,
) *App {
	app := &App{
		server:           server,
		config:           config,
		rpcClient:        rpcClient,
		rpcServer:        rpcServer,
		serviceDiscovery: serviceDiscovery,
		remoteService:    remoteService,
		debug:            false,
		startAt:          time.Now(),
		dieChan:          dieChan,
		metricsReporters: metricsReporters,
		running:          false,
		serializer:       serializer,
		router:           router,
		remoteComp:       make([]regComp, 0),
		modulesMap:       make(map[string]interfaces.Module),
		modulesArr:       []moduleWrapper{},
	}
	if app.heartbeat == time.Duration(0) {
		app.heartbeat = config.Heartbeat.Interval
	}
	return app
}

// GetDieChan gets the channel that the app sinalizes when its going to die
func (app *App) GetDieChan() chan bool {
	return app.dieChan
}

// SetDebug toggles debug on/off
func (app *App) SetDebug(debug bool) {
	app.debug = debug
}

// SetHeartbeatTime sets the heartbeat time
func (app *App) SetHeartbeatTime(interval time.Duration) {
	app.heartbeat = interval
}

// GetServerID returns the generated server id
func (app *App) GetServerID() string {
	return app.server.ID
}

// GetMetricsReporters gets registered metrics reporters
func (app *App) GetMetricsReporters() []metrics.Reporter {
	return app.metricsReporters
}

// GetServer gets the local server instance
func (app *App) GetServer() *cluster.Server {
	return app.server
}

// GetServerByID returns the server with the specified id
func (app *App) GetServerByID(id string) (*cluster.Server, error) {
	return app.serviceDiscovery.GetServer(id)
}

// GetServersByType get all servers of type
func (app *App) GetServersByType(t string) (map[string]*cluster.Server, error) {
	return app.serviceDiscovery.GetServersByType(t)
}

// GetServers get all servers
func (app *App) GetServers() []*cluster.Server {
	return app.serviceDiscovery.GetServers()
}

// IsRunning indicates if the Pitaya app has been initialized. Note: This
// doesn't cover acceptors, only the pitaya internal registration and modules
// initialization.
func (app *App) IsRunning() bool {
	return app.running
}

// SetLogger logger setter
func SetLogger(l logging.Logger) {
	logger.Log = l
}

func (app *App) periodicMetrics() {
	period := app.config.Metrics.Period
	go metrics.ReportSysMetrics(app.metricsReporters, period)
}

// Start starts the app
func (app *App) Start() {

	if err := app.RegisterModuleBefore(app.rpcServer, "rpcServer"); err != nil {
		logger.Log.Fatal("failed to register rpc server module: %s", err.Error())
	}
	if err := app.RegisterModuleBefore(app.rpcClient, "rpcClient"); err != nil {
		logger.Log.Fatal("failed to register rpc client module: %s", err.Error())
	}
	// set the service discovery as the last module to be started to ensure
	// all modules have been properly initialized before the server starts
	// receiving requests from other nrpc servers
	if err := app.RegisterModuleAfter(app.serviceDiscovery, "serviceDiscovery"); err != nil {
		logger.Log.Fatal("failed to register service discovery module: %s", err.Error())
	}

	app.periodicMetrics()

	app.start()

	defer func() {
		app.running = false
	}()

	sg := make(chan os.Signal, 1)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	// stop server
	select {
	case <-app.dieChan:
		logger.Log.Warn("the app will shutdown in a few seconds")
	case s := <-sg:
		logger.Log.Warn("got signal: ", s, ", shutting down...")
		close(app.dieChan)
	}

	logger.Log.Warn("server is stopping...")

	app.shutdownModules()
	app.shutdownComponents()
}

func (app *App) start() {
	app.startupComponents()

	logger.Log.Infof("starting server %s:%s", app.server.Type, app.server.ID)

	app.startModules()

	logger.Log.Info("all modules started!")

	app.running = true
}

// SetDictionary sets routes map
func (app *App) SetDictionary(dict map[string]uint16) error {
	if app.running {
		return constants.ErrChangeDictionaryWhileRunning
	}
	return message.SetDictionary(dict)
}

// AddRoute adds a routing function to a server type
func (app *App) AddRoute(
	serverType string,
	routingFunction router.RoutingFunc,
) error {
	if app.router != nil {
		if app.running {
			return constants.ErrChangeRouteWhileRunning
		}
		app.router.AddRoute(serverType, routingFunction)
	} else {
		return constants.ErrRouterNotInitialized
	}
	return nil
}

// Shutdown send a signal to let 'pitaya' shutdown itself.
func (app *App) Shutdown() {
	select {
	case <-app.dieChan: // prevent closing closed channel
	default:
		close(app.dieChan)
	}
}

// Error creates a new error with a code, message and metadata
func Error(err error, code string, metadata ...map[string]string) *errors.Error {
	return errors.NewError(err, code, metadata...)
}

// GetDefaultLoggerFromCtx returns the default logger from the given context
func GetDefaultLoggerFromCtx(ctx context.Context) logging.Logger {
	l := ctx.Value(constants.LoggerCtxKey)
	if l == nil {
		return logger.Log
	}

	return l.(logging.Logger)
}

// AddMetricTagsToPropagateCtx adds a key and metric tags that will
// be propagated through RPC calls. Use the same tags that are at
// 'pitaya.metrics.additionalLabels' config
func AddMetricTagsToPropagateCtx(
	ctx context.Context,
	tags map[string]string,
) context.Context {
	return pcontext.AddToPropagateCtx(ctx, constants.MetricTagsKey, tags)
}

// AddToPropagateCtx adds a key and value that will be propagated through RPC calls
func AddToPropagateCtx(ctx context.Context, key string, val interface{}) context.Context {
	return pcontext.AddToPropagateCtx(ctx, key, val)
}

// GetFromPropagateCtx adds a key and value that came through RPC calls
func GetFromPropagateCtx(ctx context.Context, key string) interface{} {
	return pcontext.GetFromPropagateCtx(ctx, key)
}

// ExtractSpan retrieves an opentracing span context from the given context
// The span context can be received directly or via an RPC call
func ExtractSpan(ctx context.Context) (opentracing.SpanContext, error) {
	return tracing.ExtractSpan(ctx)
}
