package config

import (
	"time"

	"github.com/seanbit/nrpc/metrics/models"
)

const DefaultWriteTimeout = 10 * time.Second

// Config provides all the configuration for a nrpc app
type Config struct {
	ServerID         int    `mapstructure:"server_id"`
	SerializerType   uint16 `mapstructure:"serializertype"`
	DefaultPipelines struct {
		StructValidation struct {
			Enabled bool `mapstructure:"enabled"`
		} `mapstructure:"structvalidation"`
	} `mapstructure:"defaultpipelines"`
	Modules   ModulesConfig
	Heartbeat struct {
		Interval time.Duration `mapstructure:"interval"`
	} `mapstructure:"heartbeat"`
	Handler struct {
		Messages struct {
			Compression bool `mapstructure:"compression"`
		} `mapstructure:"messages"`
	} `mapstructure:"handler"`
	Concurrency struct {
		PoolSize      int           `mapstructure:"poolsize"`
		QueueSize     int           `mapstructure:"queuesize"`
		ClearInterval time.Duration `mapstructure:"clearinterval"`
		ClearTimePass time.Duration `mapstructure:"cleartimepass"`
	} `mapstructure:"concurrency"`
	Conn struct {
		RateLimiting RateLimitingConfig `mapstructure:"rateLimiting"`
	} `mapstructure:"conn"`
	Metrics MetricsConfig `mapstructure:"metrics"`
	Cluster ClusterConfig `mapstructure:"cluster"`
	Groups  GroupsConfig  `mapstructure:"groups"`
	Worker  WorkerConfig  `mapstructure:"worker"`
}

// NewDefaultConfig provides default configuration for nrpc App
func NewDefaultConfig() *Config {
	return &Config{
		SerializerType: 1,
		DefaultPipelines: struct {
			StructValidation struct {
				Enabled bool `mapstructure:"enabled"`
			} `mapstructure:"structvalidation"`
		}{
			StructValidation: struct {
				Enabled bool `mapstructure:"enabled"`
			}{
				Enabled: false,
			},
		},
		Heartbeat: struct {
			Interval time.Duration `mapstructure:"interval"`
		}{
			Interval: time.Duration(30 * time.Second),
		},
		Handler: struct {
			Messages struct {
				Compression bool `mapstructure:"compression"`
			} `mapstructure:"messages"`
		}{
			Messages: struct {
				Compression bool `mapstructure:"compression"`
			}{
				Compression: true,
			},
		},
		Concurrency: struct {
			PoolSize      int           `mapstructure:"poolsize"`
			QueueSize     int           `mapstructure:"queuesize"`
			ClearInterval time.Duration `mapstructure:"clearinterval"`
			ClearTimePass time.Duration `mapstructure:"cleartimepass"`
		}{
			PoolSize:      128,
			QueueSize:     256,
			ClearInterval: 30 * time.Second,
			ClearTimePass: 300 * time.Second,
		},
		Metrics: *newDefaultMetricsConfig(),
		Cluster: *newDefaultClusterConfig(),
		Groups:  *newDefaultGroupsConfig(),
		Worker:  *newDefaultWorkerConfig(),
		Modules: *newDefaultModulesConfig(),
		Conn: struct {
			RateLimiting RateLimitingConfig `mapstructure:"rateLimiting"`
		}{
			RateLimiting: *newDefaultRateLimitingConfig(),
		},
	}
}

// NewConfig returns a config instance with values extracted from default config paths
func NewConfig(config *ViperConfig) *Config {
	conf := NewDefaultConfig()
	if err := config.UnmarshalKey("nrpc", &conf); err != nil {
		panic(err)
	}
	return conf
}

// NatsRPCClientConfig provides nats client configuration
type NatsRPCClientConfig struct {
	Connect                string        `mapstructure:"connect"`
	MaxReconnectionRetries int           `mapstructure:"maxreconnectionretries"`
	RequestTimeout         time.Duration `mapstructure:"requesttimeout"`
	ConnectionTimeout      time.Duration `mapstructure:"connectiontimeout"`
	WebsocketCompression   bool          `mapstructure:"websocketcompression"`
	ReconnectJitter        time.Duration `mapstructure:"reconnectjitter"`
	ReconnectJitterTLS     time.Duration `mapstructure:"reconnectjittertls"`
	ReconnectWait          time.Duration `mapstructure:"reconnectwait"`
	PingInterval           time.Duration `mapstructure:"pinginterval"`
	MaxPingsOutstanding    int           `mapstructure:"maxpingsoutstanding"`
}

// newDefaultNatsRPCClientConfig provides default nats client configuration
func newDefaultNatsRPCClientConfig() *NatsRPCClientConfig {
	return &NatsRPCClientConfig{
		Connect:                "nats://localhost:4222",
		MaxReconnectionRetries: 15,
		RequestTimeout:         time.Duration(5 * time.Second),
		ConnectionTimeout:      time.Duration(2 * time.Second),
		WebsocketCompression:   true,
		ReconnectJitter:        time.Duration(100 * time.Millisecond),
		ReconnectJitterTLS:     time.Duration(1 * time.Second),
		ReconnectWait:          time.Duration(time.Second),
		PingInterval:           time.Duration(2 * time.Minute),
		MaxPingsOutstanding:    3,
	}
}

// NatsRPCServerConfig provides nats server configuration
type NatsRPCServerConfig struct {
	Connect                string `mapstructure:"connect"`
	MaxReconnectionRetries int    `mapstructure:"maxreconnectionretries"`
	Buffer                 struct {
		Messages int `mapstructure:"messages"`
	} `mapstructure:"buffer"`
	Services             int           `mapstructure:"services"`
	Semaphore            int           `mapstructure:"semaphore"`
	ConnectionTimeout    time.Duration `mapstructure:"connectiontimeout"`
	WebsocketCompression bool          `mapstructure:"websocketcompression"`
	ReconnectJitter      time.Duration `mapstructure:"reconnectjitter"`
	ReconnectJitterTLS   time.Duration `mapstructure:"reconnectjittertls"`
	ReconnectWait        time.Duration `mapstructure:"reconnectwait"`
	PingInterval         time.Duration `mapstructure:"pinginterval"`
	MaxPingsOutstanding  int           `mapstructure:"maxpingsoutstanding"`
}

// newDefaultNatsRPCServerConfig provides default nats server configuration
func newDefaultNatsRPCServerConfig() *NatsRPCServerConfig {
	return &NatsRPCServerConfig{
		Connect:                "nats://localhost:4222",
		MaxReconnectionRetries: 15,
		Buffer: struct {
			Messages int `mapstructure:"messages"`
		}{
			Messages: 75,
		},
		Services:             30,
		Semaphore:            256,
		ConnectionTimeout:    time.Duration(2 * time.Second),
		WebsocketCompression: true,
		ReconnectJitter:      time.Duration(100 * time.Millisecond),
		ReconnectJitterTLS:   time.Duration(1 * time.Second),
		ReconnectWait:        time.Duration(time.Second),
		PingInterval:         time.Duration(2 * time.Minute),
		MaxPingsOutstanding:  3,
	}
}

// InfoRetrieverConfig provides InfoRetriever configuration
type InfoRetrieverConfig struct {
	Region string `mapstructure:"region"`
}

// newDefaultInfoRetrieverConfig provides default configuration for InfoRetriever
func newDefaultInfoRetrieverConfig() *InfoRetrieverConfig {
	return &InfoRetrieverConfig{
		Region: "",
	}
}

// EtcdServiceDiscoveryConfig Etcd service discovery config
type EtcdServiceDiscoveryConfig struct {
	Endpoints   []string      `mapstructure:"endpoints"`
	User        string        `mapstructure:"user"`
	Pass        string        `mapstructure:"pass"`
	DialTimeout time.Duration `mapstructure:"dialtimeout"`
	Prefix      string        `mapstructure:"prefix"`
	Heartbeat   struct {
		TTL time.Duration `mapstructure:"ttl"`
		Log bool          `mapstructure:"log"`
	} `mapstructure:"heartbeat"`
	SyncServers struct {
		Interval    time.Duration `mapstructure:"interval"`
		Parallelism int           `mapstructure:"parallelism"`
	} `mapstructure:"syncservers"`
	Revoke struct {
		Timeout time.Duration `mapstructure:"timeout"`
	} `mapstructure:"revoke"`
	GrantLease struct {
		Timeout       time.Duration `mapstructure:"timeout"`
		MaxRetries    int           `mapstructure:"maxretries"`
		RetryInterval time.Duration `mapstructure:"retryinterval"`
	} `mapstructure:"grantlease"`
	Shutdown struct {
		Delay time.Duration `mapstructure:"delay"`
	} `mapstructure:"shutdown"`
	ServerTypesBlacklist []string `mapstructure:"servertypesblacklist"`
}

// newDefaultEtcdServiceDiscoveryConfig Etcd service discovery default config
func newDefaultEtcdServiceDiscoveryConfig() *EtcdServiceDiscoveryConfig {
	return &EtcdServiceDiscoveryConfig{
		Endpoints:   []string{"localhost:2379"},
		User:        "",
		Pass:        "",
		DialTimeout: time.Duration(5 * time.Second),
		Prefix:      "nrpc/",
		Heartbeat: struct {
			TTL time.Duration `mapstructure:"ttl"`
			Log bool          `mapstructure:"log"`
		}{
			TTL: time.Duration(60 * time.Second),
			Log: false,
		},
		SyncServers: struct {
			Interval    time.Duration `mapstructure:"interval"`
			Parallelism int           `mapstructure:"parallelism"`
		}{
			Interval:    time.Duration(120 * time.Second),
			Parallelism: 10,
		},
		Revoke: struct {
			Timeout time.Duration `mapstructure:"timeout"`
		}{
			Timeout: time.Duration(5 * time.Second),
		},
		GrantLease: struct {
			Timeout       time.Duration `mapstructure:"timeout"`
			MaxRetries    int           `mapstructure:"maxretries"`
			RetryInterval time.Duration `mapstructure:"retryinterval"`
		}{
			Timeout:       time.Duration(60 * time.Second),
			MaxRetries:    15,
			RetryInterval: time.Duration(5 * time.Second),
		},
		Shutdown: struct {
			Delay time.Duration `mapstructure:"delay"`
		}{
			Delay: time.Duration(300 * time.Millisecond),
		},
		ServerTypesBlacklist: nil,
	}
}

// NewDefaultCustomMetricsSpec returns an empty *CustomMetricsSpec
func NewDefaultCustomMetricsSpec() *models.CustomMetricsSpec {
	return &models.CustomMetricsSpec{
		Summaries: []*models.Summary{},
		Gauges:    []*models.Gauge{},
		Counters:  []*models.Counter{},
	}
}

// NewCustomMetricsSpec returns a *CustomMetricsSpec by reading config key (DEPRECATED)
func NewCustomMetricsSpec(config *ViperConfig) *models.CustomMetricsSpec {
	spec := &models.CustomMetricsSpec{}

	if err := config.UnmarshalKey("nrpc.metrics.custom", &spec); err != nil {
		return NewDefaultCustomMetricsSpec()
	}

	return spec
}

// Metrics provides configuration for all metrics related configurations
type MetricsConfig struct {
	App              string                   `mapstructure:"app"`
	Period           time.Duration            `mapstructure:"period"`
	AdditionalLabels map[string]string        `mapstructure:"additionallabels"`
	ConstLabels      map[string]string        `mapstructure:"constlabels"`
	Custom           models.CustomMetricsSpec `mapstructure:"custom"`
	Prometheus       *PrometheusConfig        `mapstructure:"prometheus"`
	Statsd           *StatsdConfig            `mapstructure:"statsd"`
}

// newDefaultPrometheusConfig provides default configuration for PrometheusReporter
func newDefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		Period:           time.Duration(15 * time.Second),
		ConstLabels:      map[string]string{},
		AdditionalLabels: map[string]string{},
		Custom:           *NewDefaultCustomMetricsSpec(),
		Prometheus:       newDefaultPrometheusConfig(),
		Statsd:           newDefaultStatsdConfig(),
	}
}

// PrometheusConfig provides configuration for PrometheusReporter
type PrometheusConfig struct {
	Port    int  `mapstructure:"port"`
	Enabled bool `mapstructure:"enabled"`
}

// newDefaultPrometheusConfig provides default configuration for PrometheusReporter
func newDefaultPrometheusConfig() *PrometheusConfig {
	return &PrometheusConfig{
		Port:    9090,
		Enabled: false,
	}
}

// StatsdConfig provides configuration for statsd
type StatsdConfig struct {
	Enabled bool    `mapstructure:"enabled"`
	Host    string  `mapstructure:"host"`
	Prefix  string  `mapstructure:"prefix"`
	Rate    float64 `mapstructure:"rate"`
}

// newDefaultStatsdConfig provides default configuration for statsd
func newDefaultStatsdConfig() *StatsdConfig {
	return &StatsdConfig{
		Enabled: false,
		Host:    "localhost:9125",
		Prefix:  "nrpc.",
		Rate:    1,
	}
}

// newDefaultStatsdConfig provides default configuration for statsd
func newDefaultClusterConfig() *ClusterConfig {
	return &ClusterConfig{
		Info: *newDefaultInfoRetrieverConfig(),
		RPC:  *newDefaultClusterRPCConfig(),
		SD:   *newDefaultClusterSDConfig(),
	}
}

type ClusterConfig struct {
	Info InfoRetrieverConfig `mapstructure:"info"`
	RPC  ClusterRPCConfig    `mapstructure:"rpc"`
	SD   ClusterSDConfig     `mapstructure:"sd"`
}

type ClusterRPCConfig struct {
	Client struct {
		Nats NatsRPCClientConfig `mapstructure:"nats"`
	} `mapstructure:"client"`
	Server struct {
		Nats NatsRPCServerConfig `mapstructure:"nats"`
	} `mapstructure:"server"`
}

func newDefaultClusterRPCConfig() *ClusterRPCConfig {
	return &ClusterRPCConfig{
		Client: struct {
			Nats NatsRPCClientConfig `mapstructure:"nats"`
		}{
			Nats: *newDefaultNatsRPCClientConfig(),
		},
		Server: struct {
			Nats NatsRPCServerConfig `mapstructure:"nats"`
		}{
			Nats: *newDefaultNatsRPCServerConfig(),
		},
	}

}

type ClusterSDConfig struct {
	Etcd EtcdServiceDiscoveryConfig `mapstructure:"etcd"`
}

func newDefaultClusterSDConfig() *ClusterSDConfig {
	return &ClusterSDConfig{Etcd: *newDefaultEtcdServiceDiscoveryConfig()}
}

// WorkerConfig provides worker configuration
type WorkerConfig struct {
	Redis struct {
		ServerURL string `mapstructure:"serverurl"`
		Pool      string `mapstructure:"pool"`
		Password  string `mapstructure:"password"`
	} `mapstructure:"redis"`
	Namespace   string      `mapstructure:"namespace"`
	Concurrency int         `mapstructure:"concurrency"`
	Retry       EnqueueOpts `mapstructure:"retry"`
}

// newDefaultWorkerConfig provides worker default configuration
func newDefaultWorkerConfig() *WorkerConfig {
	return &WorkerConfig{
		Redis: struct {
			ServerURL string `mapstructure:"serverurl"`
			Pool      string `mapstructure:"pool"`
			Password  string `mapstructure:"password"`
		}{
			ServerURL: "localhost:6379",
			Pool:      "10",
		},
		Concurrency: 1,
		Retry:       *newDefaultEnqueueOpts(),
	}
}

// EnqueueOpts has retry options for worker
type EnqueueOpts struct {
	Enabled     bool `mapstructure:"enabled"`
	Max         int  `mapstructure:"max"`
	Exponential int  `mapstructure:"exponential"`
	MinDelay    int  `mapstructure:"mindelay"`
	MaxDelay    int  `mapstructure:"maxdelay"`
	MaxRandom   int  `mapstructure:"maxrandom"`
}

// newDefaultEnqueueOpts provides default EnqueueOpts
func newDefaultEnqueueOpts() *EnqueueOpts {
	return &EnqueueOpts{
		Enabled:     true,
		Max:         2,
		Exponential: 5,
		MinDelay:    10,
		MaxDelay:    10,
		MaxRandom:   0,
	}
}

// MemoryGroupConfig provides configuration for MemoryGroup
type MemoryGroupConfig struct {
	TickDuration time.Duration `mapstructure:"tickduration"`
}

// newDefaultMemoryGroupConfig returns a new, default group instance
func newDefaultMemoryGroupConfig() *MemoryGroupConfig {
	return &MemoryGroupConfig{TickDuration: time.Duration(30 * time.Second)}
}

// EtcdGroupServiceConfig provides ETCD configuration
type EtcdGroupServiceConfig struct {
	DialTimeout        time.Duration `mapstructure:"dialtimeout"`
	Endpoints          []string      `mapstructure:"endpoints"`
	Prefix             string        `mapstructure:"prefix"`
	TransactionTimeout time.Duration `mapstructure:"transactiontimeout"`
}

// newDefaultEtcdGroupServiceConfig provides default ETCD configuration
func newDefaultEtcdGroupServiceConfig() *EtcdGroupServiceConfig {
	return &EtcdGroupServiceConfig{
		DialTimeout:        time.Duration(5 * time.Second),
		Endpoints:          []string{"localhost:2379"},
		Prefix:             "nrpc/",
		TransactionTimeout: time.Duration(5 * time.Second),
	}
}

// NewEtcdGroupServiceConfig reads from config to build ETCD configuration
func newEtcdGroupServiceConfig(config *ViperConfig) *EtcdGroupServiceConfig {
	conf := newDefaultEtcdGroupServiceConfig()
	if err := config.UnmarshalKey("nrpc.groups.etcd", &conf); err != nil {
		panic(err)
	}
	return conf
}

type GroupsConfig struct {
	Etcd   EtcdGroupServiceConfig `mapstructure:"etcd"`
	Memory MemoryGroupConfig      `mapstructure:"memory"`
}

// NewDefaultGroupConfig provides default ETCD configuration
func newDefaultGroupsConfig() *GroupsConfig {
	return &GroupsConfig{
		Etcd:   *newDefaultEtcdGroupServiceConfig(),
		Memory: *newDefaultMemoryGroupConfig(),
	}
}

// ETCDBindingConfig provides configuration for ETCDBindingStorage
type ETCDBindingConfig struct {
	DialTimeout time.Duration `mapstructure:"dialtimeout"`
	Endpoints   []string      `mapstructure:"endpoints"`
	Prefix      string        `mapstructure:"prefix"`
	LeaseTTL    time.Duration `mapstructure:"leasettl"`
}

// NewDefaultETCDBindingConfig provides default configuration for ETCDBindingStorage
func newDefaultETCDBindingConfig() *ETCDBindingConfig {
	return &ETCDBindingConfig{
		DialTimeout: time.Duration(5 * time.Second),
		Endpoints:   []string{"localhost:2379"},
		Prefix:      "nrpc/",
		LeaseTTL:    time.Duration(5 * time.Hour),
	}
}

// ModulesConfig provides configuration for nrpc Modules
type ModulesConfig struct {
	BindingStorage struct {
		Etcd ETCDBindingConfig `mapstructure:"etcd"`
	} `mapstructure:"bindingstorage"`
}

// NewDefaultModulesConfig provides default configuration for nrpc Modules
func newDefaultModulesConfig() *ModulesConfig {
	return &ModulesConfig{
		BindingStorage: struct {
			Etcd ETCDBindingConfig `mapstructure:"etcd"`
		}{
			Etcd: *newDefaultETCDBindingConfig(),
		},
	}
}

// RateLimitingConfig rate limits config
type RateLimitingConfig struct {
	Limit        int           `mapstructure:"limit"`
	Interval     time.Duration `mapstructure:"interval"`
	ForceDisable bool          `mapstructure:"forcedisable"`
}

// newDefaultRateLimitingConfig rate limits default config
func newDefaultRateLimitingConfig() *RateLimitingConfig {
	return &RateLimitingConfig{
		Limit:        20,
		Interval:     time.Duration(time.Second),
		ForceDisable: false,
	}
}
