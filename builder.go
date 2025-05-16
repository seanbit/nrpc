package nrpc

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/seanbit/nrpc/cluster"
	"github.com/seanbit/nrpc/config"
	"github.com/seanbit/nrpc/conn/codec"
	"github.com/seanbit/nrpc/conn/message"
	"github.com/seanbit/nrpc/logger"
	"github.com/seanbit/nrpc/metrics"
	"github.com/seanbit/nrpc/metrics/models"
	"github.com/seanbit/nrpc/pipeline"
	"github.com/seanbit/nrpc/router"
	"github.com/seanbit/nrpc/serialize"
	"github.com/seanbit/nrpc/service"
)

// Builder holds dependency instances for a pitaya App
type Builder struct {
	postBuildHooks   []func(app NRpc)
	Config           config.Config
	DieChan          chan bool
	PacketDecoder    codec.PacketDecoder
	PacketEncoder    codec.PacketEncoder
	MessageEncoder   *message.MessagesEncoder
	Serializer       serialize.Serializer
	Router           *router.Router
	RPCClient        cluster.RPCClient
	RPCServer        cluster.RPCServer
	MetricsReporters []metrics.Reporter
	Server           *cluster.Server
	ServiceDiscovery cluster.ServiceDiscovery
	RemoteHooks      *pipeline.RemoteHooks
}

// NBuilder Builder interface
type NBuilder interface {
	// AddPostBuildHook adds a post-build hook to the builder, a function receiving a Pitaya instance as parameter.
	AddPostBuildHook(hook func(app NRpc))
	Build() NRpc
}

// NewBuilderWithConfigs return a builder instance with default dependency instances for a nrpc App
// with configs defined by a config file (config.Config) and default paths (see documentation).
func NewBuilderWithConfigs(
	serverType string,
	serverMetadata map[string]string,
	conf *config.ViperConfig,
) *Builder {
	appConfig := config.NewConfig(conf)
	return NewBuilder(
		serverType,
		serverMetadata,
		*appConfig,
	)
}

// NewDefaultBuilder return a builder instance with default dependency instances for a nrpc App,
// with default configs
func NewDefaultBuilder(serverType string, serverMetadata map[string]string, appConfig config.Config) *Builder {
	return NewBuilder(
		serverType,
		serverMetadata,
		appConfig,
	)
}

// NewBuilder return a builder instance with default dependency instances for a nrpc App,
// with configs explicitly defined
func NewBuilder(
	serverType string,
	serverMetadata map[string]string,
	config config.Config,
) *Builder {
	serverId := ""
	if config.ServerID > 0 {
		serverId = fmt.Sprintf("%s:%d", serverType, config.ServerID)
	}
	if serverId == "" {
		serverId = uuid.New().String()
	}
	if serverMetadata == nil {
		serverMetadata = map[string]string{}
	}
	serverMetadata["id"] = serverId
	server := cluster.NewServer(serverId, serverType, serverMetadata)
	dieChan := make(chan bool)

	metricsReporters := []metrics.Reporter{}
	if config.Metrics.Prometheus.Enabled {
		metricsReporters = addDefaultPrometheus(config.Metrics, config.Metrics.Custom, metricsReporters, serverType)
	}

	if config.Metrics.Statsd.Enabled {
		metricsReporters = addDefaultStatsd(config.Metrics, metricsReporters, serverType)
	}

	var serviceDiscovery cluster.ServiceDiscovery
	var rpcServer cluster.RPCServer
	var rpcClient cluster.RPCClient

	var err error
	serviceDiscovery, err = cluster.NewEtcdServiceDiscovery(config.Cluster.SD.Etcd, server, dieChan)
	if err != nil {
		logger.Log.Fatalf("error creating default cluster service discovery component: %s", err.Error())
	}

	workerPool := cluster.NewWorkerPool(
		config.Concurrency.PoolSize,
		config.Concurrency.QueueSize,
		config.Concurrency.ClearInterval,
		config.Concurrency.ClearTimePass,
	)
	natsServer, err := cluster.NewNatsRPCServer(config.Cluster.RPC.Server.Nats, server, metricsReporters, dieChan, workerPool)
	if err != nil {
		logger.Log.Fatalf("error setting default cluster rpc server component: %s", err.Error())
	}
	rpcServer = natsServer

	rpcClient, err = cluster.NewNatsRPCClient(config.Cluster.RPC.Client.Nats, server, metricsReporters, dieChan)
	if err != nil {
		logger.Log.Fatalf("error setting default cluster rpc client component: %s", err.Error())
	}

	serializer, err := serialize.NewSerializer(serialize.Type(config.SerializerType))
	if err != nil {
		logger.Log.Fatalf("error creating serializer: %s", err.Error())
	}

	return &Builder{
		postBuildHooks:   make([]func(app NRpc), 0),
		Config:           config,
		DieChan:          dieChan,
		PacketDecoder:    codec.NewPomeloPacketDecoder(),
		PacketEncoder:    codec.NewPomeloPacketEncoder(),
		MessageEncoder:   message.NewMessagesEncoder(config.Handler.Messages.Compression),
		Serializer:       serializer,
		Router:           router.New(),
		RPCClient:        rpcClient,
		RPCServer:        rpcServer,
		MetricsReporters: metricsReporters,
		Server:           server,
		RemoteHooks:      pipeline.NewRemoteHooks(),
		ServiceDiscovery: serviceDiscovery,
	}
}

// AddPostBuildHook adds a post-build hook to the builder, a function receiving a Pitaya instance as parameter.
func (builder *Builder) AddPostBuildHook(hook func(app NRpc)) {
	builder.postBuildHooks = append(builder.postBuildHooks, hook)
}

// Build returns a valid App instance
func (builder *Builder) Build() NRpc {
	var remoteService *service.RemoteService

	if !(builder.ServiceDiscovery != nil && builder.RPCClient != nil && builder.RPCServer != nil) {
		panic("Cluster mode must have RPC and service discovery instances")
	}

	builder.Router.SetServiceDiscovery(builder.ServiceDiscovery)

	remoteService = service.NewRemoteService(
		builder.RPCClient,
		builder.RPCServer,
		builder.ServiceDiscovery,
		builder.PacketEncoder,
		builder.Serializer,
		builder.Router,
		builder.MessageEncoder,
		builder.Server,
		builder.RemoteHooks,
	)

	builder.RPCServer.SetHandler(remoteService)

	app := NewApp(
		builder.Serializer,
		builder.DieChan,
		builder.Router,
		builder.Server,
		builder.RPCClient,
		builder.RPCServer,
		builder.ServiceDiscovery,
		remoteService,
		builder.MetricsReporters,
		builder.Config,
	)

	for _, postBuildHook := range builder.postBuildHooks {
		postBuildHook(app)
	}

	return app
}

// NewDefaultApp returns a default pitaya app instance
func NewDefaultApp(serverType string, serverMetadata map[string]string, config config.Config) NRpc {
	builder := NewDefaultBuilder(serverType, serverMetadata, config)
	return builder.Build()
}

func addDefaultPrometheus(config config.MetricsConfig, customMetrics models.CustomMetricsSpec, reporters []metrics.Reporter, serverType string) []metrics.Reporter {
	prometheus, err := CreatePrometheusReporter(serverType, config, customMetrics)
	if err != nil {
		logger.Log.Errorf("failed to start prometheus metrics reporter, skipping %v", err)
	} else {
		reporters = append(reporters, prometheus)
	}
	return reporters
}

func addDefaultStatsd(config config.MetricsConfig, reporters []metrics.Reporter, serverType string) []metrics.Reporter {
	statsd, err := CreateStatsdReporter(serverType, config)
	if err != nil {
		logger.Log.Errorf("failed to start statsd metrics reporter, skipping %v", err)
	} else {
		reporters = append(reporters, statsd)
	}
	return reporters
}
