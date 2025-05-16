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

package config

import (
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"

	"github.com/spf13/viper"
)

// ViperConfig is a wrapper around a viper config
type ViperConfig struct {
	viper.Viper
}

// NewViperConfig creates a new config with a given viper config if given
func NewViperConfig(cfgs ...*viper.Viper) *ViperConfig {
	var cfg *viper.Viper
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	} else {
		cfg = viper.New()
	}

	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	cfg.AutomaticEnv()
	c := &ViperConfig{*cfg}
	c.fillDefaultValues()
	return c
}

func (c *ViperConfig) fillDefaultValues() {
	nrpcConfig := NewDefaultConfig()

	defaultsMap := map[string]interface{}{
		"nrpc.serializertype":                                 nrpcConfig.SerializerType,
		"nrpc.cluster.info.region":                            nrpcConfig.Cluster.Info.Region,
		"nrpc.cluster.rpc.client.nats.connect":                nrpcConfig.Cluster.RPC.Client.Nats.Connect,
		"nrpc.cluster.rpc.client.nats.connectiontimeout":      nrpcConfig.Cluster.RPC.Client.Nats.ConnectionTimeout,
		"nrpc.cluster.rpc.client.nats.maxreconnectionretries": nrpcConfig.Cluster.RPC.Client.Nats.MaxReconnectionRetries,
		"nrpc.cluster.rpc.client.nats.websocketcompression":   nrpcConfig.Cluster.RPC.Client.Nats.WebsocketCompression,
		"nrpc.cluster.rpc.client.nats.reconnectjitter":        nrpcConfig.Cluster.RPC.Client.Nats.ReconnectJitter,
		"nrpc.cluster.rpc.client.nats.reconnectjittertls":     nrpcConfig.Cluster.RPC.Client.Nats.ReconnectJitterTLS,
		"nrpc.cluster.rpc.client.nats.reconnectwait":          nrpcConfig.Cluster.RPC.Client.Nats.ReconnectWait,
		"nrpc.cluster.rpc.client.nats.pinginterval":           nrpcConfig.Cluster.RPC.Client.Nats.PingInterval,
		"nrpc.cluster.rpc.client.nats.maxpingsoutstanding":    nrpcConfig.Cluster.RPC.Client.Nats.MaxPingsOutstanding,
		"nrpc.cluster.rpc.client.nats.requesttimeout":         nrpcConfig.Cluster.RPC.Client.Nats.RequestTimeout,
		"nrpc.cluster.rpc.server.nats.connect":                nrpcConfig.Cluster.RPC.Server.Nats.Connect,
		"nrpc.cluster.rpc.server.nats.connectiontimeout":      nrpcConfig.Cluster.RPC.Server.Nats.ConnectionTimeout,
		"nrpc.cluster.rpc.server.nats.maxreconnectionretries": nrpcConfig.Cluster.RPC.Server.Nats.MaxReconnectionRetries,
		"nrpc.cluster.rpc.server.nats.websocketcompression":   nrpcConfig.Cluster.RPC.Server.Nats.WebsocketCompression,
		"nrpc.cluster.rpc.server.nats.reconnectjitter":        nrpcConfig.Cluster.RPC.Server.Nats.ReconnectJitter,
		"nrpc.cluster.rpc.server.nats.reconnectjittertls":     nrpcConfig.Cluster.RPC.Server.Nats.ReconnectJitterTLS,
		"nrpc.cluster.rpc.server.nats.reconnectwait":          nrpcConfig.Cluster.RPC.Server.Nats.ReconnectWait,
		"nrpc.cluster.rpc.server.nats.pinginterval":           nrpcConfig.Cluster.RPC.Server.Nats.PingInterval,
		"nrpc.cluster.rpc.server.nats.maxpingsoutstanding":    nrpcConfig.Cluster.RPC.Server.Nats.MaxPingsOutstanding,
		"nrpc.cluster.rpc.server.nats.services":               nrpcConfig.Cluster.RPC.Server.Nats.Services,
		"nrpc.cluster.rpc.server.nats.buffer.messages":        nrpcConfig.Cluster.RPC.Server.Nats.Buffer.Messages,
		"nrpc.cluster.sd.etcd.dialtimeout":                    nrpcConfig.Cluster.SD.Etcd.DialTimeout,
		"nrpc.cluster.sd.etcd.endpoints":                      nrpcConfig.Cluster.SD.Etcd.Endpoints,
		"nrpc.cluster.sd.etcd.prefix":                         nrpcConfig.Cluster.SD.Etcd.Prefix,
		"nrpc.cluster.sd.etcd.grantlease.maxretries":          nrpcConfig.Cluster.SD.Etcd.GrantLease.MaxRetries,
		"nrpc.cluster.sd.etcd.grantlease.retryinterval":       nrpcConfig.Cluster.SD.Etcd.GrantLease.RetryInterval,
		"nrpc.cluster.sd.etcd.grantlease.timeout":             nrpcConfig.Cluster.SD.Etcd.GrantLease.Timeout,
		"nrpc.cluster.sd.etcd.heartbeat.log":                  nrpcConfig.Cluster.SD.Etcd.Heartbeat.Log,
		"nrpc.cluster.sd.etcd.heartbeat.ttl":                  nrpcConfig.Cluster.SD.Etcd.Heartbeat.TTL,
		"nrpc.cluster.sd.etcd.revoke.timeout":                 nrpcConfig.Cluster.SD.Etcd.Revoke.Timeout,
		"nrpc.cluster.sd.etcd.syncservers.interval":           nrpcConfig.Cluster.SD.Etcd.SyncServers.Interval,
		"nrpc.cluster.sd.etcd.syncservers.parallelism":        nrpcConfig.Cluster.SD.Etcd.SyncServers.Parallelism,
		"nrpc.cluster.sd.etcd.shutdown.delay":                 nrpcConfig.Cluster.SD.Etcd.Shutdown.Delay,
		"nrpc.cluster.sd.etcd.servertypeblacklist":            nrpcConfig.Cluster.SD.Etcd.ServerTypesBlacklist,
		// the sum of this config among all the frontend servers should always be less than
		// the sum of nrpc.buffer.cluster.rpc.server.nats.messages, for covering the worst case scenario
		// a single backend server should have the config nrpc.buffer.cluster.rpc.server.nats.messages bigger
		// than the sum of the config nrpc.concurrency.handler.dispatch among all frontend servers
		"nrpc.defaultpipelines.structvalidation.enabled": nrpcConfig.DefaultPipelines.StructValidation.Enabled,
		"nrpc.groups.etcd.dialtimeout":                   nrpcConfig.Groups.Etcd.DialTimeout,
		"nrpc.groups.etcd.endpoints":                     nrpcConfig.Groups.Etcd.Endpoints,
		"nrpc.groups.etcd.prefix":                        nrpcConfig.Groups.Etcd.Prefix,
		"nrpc.groups.etcd.transactiontimeout":            nrpcConfig.Groups.Etcd.TransactionTimeout,
		"nrpc.groups.memory.tickduration":                nrpcConfig.Groups.Memory.TickDuration,
		"nrpc.handler.messages.compression":              nrpcConfig.Handler.Messages.Compression,
		"nrpc.heartbeat.interval":                        nrpcConfig.Heartbeat.Interval,
		"nrpc.metrics.additionalLabels":                  nrpcConfig.Metrics.AdditionalLabels,
		"nrpc.metrics.constLabels":                       nrpcConfig.Metrics.ConstLabels,
		"nrpc.metrics.custom":                            nrpcConfig.Metrics.Custom,
		"nrpc.metrics.period":                            nrpcConfig.Metrics.Period,
		"nrpc.metrics.prometheus.enabled":                nrpcConfig.Metrics.Prometheus.Enabled,
		"nrpc.metrics.prometheus.port":                   nrpcConfig.Metrics.Prometheus.Port,
		"nrpc.metrics.statsd.enabled":                    nrpcConfig.Metrics.Statsd.Enabled,
		"nrpc.metrics.statsd.host":                       nrpcConfig.Metrics.Statsd.Host,
		"nrpc.metrics.statsd.prefix":                     nrpcConfig.Metrics.Statsd.Prefix,
		"nrpc.metrics.statsd.rate":                       nrpcConfig.Metrics.Statsd.Rate,
		"nrpc.modules.bindingstorage.etcd.dialtimeout":   nrpcConfig.Modules.BindingStorage.Etcd.DialTimeout,
		"nrpc.modules.bindingstorage.etcd.endpoints":     nrpcConfig.Modules.BindingStorage.Etcd.Endpoints,
		"nrpc.modules.bindingstorage.etcd.leasettl":      nrpcConfig.Modules.BindingStorage.Etcd.LeaseTTL,
		"nrpc.modules.bindingstorage.etcd.prefix":        nrpcConfig.Modules.BindingStorage.Etcd.Prefix,
		"nrpc.conn.ratelimiting.limit":                   nrpcConfig.Conn.RateLimiting.Limit,
		"nrpc.conn.ratelimiting.interval":                nrpcConfig.Conn.RateLimiting.Interval,
		"nrpc.conn.ratelimiting.forcedisable":            nrpcConfig.Conn.RateLimiting.ForceDisable,
		"nrpc.worker.concurrency":                        nrpcConfig.Worker.Concurrency,
		"nrpc.worker.redis.pool":                         nrpcConfig.Worker.Redis.Pool,
		"nrpc.worker.redis.url":                          nrpcConfig.Worker.Redis.ServerURL,
		"nrpc.worker.retry.enabled":                      nrpcConfig.Worker.Retry.Enabled,
		"nrpc.worker.retry.exponential":                  nrpcConfig.Worker.Retry.Exponential,
		"nrpc.worker.retry.max":                          nrpcConfig.Worker.Retry.Max,
		"nrpc.worker.retry.maxDelay":                     nrpcConfig.Worker.Retry.MaxDelay,
		"nrpc.worker.retry.maxRandom":                    nrpcConfig.Worker.Retry.MaxRandom,
		"nrpc.worker.retry.minDelay":                     nrpcConfig.Worker.Retry.MinDelay,
	}

	for param := range defaultsMap {
		val := c.Get(param)
		if val == nil {
			c.SetDefault(param, defaultsMap[param])
		} else {
			c.SetDefault(param, val)
			c.Set(param, val)
		}

	}
}

// UnmarshalKey unmarshals key into v
func (c *ViperConfig) UnmarshalKey(key string, rawVal interface{}) error {
	key = strings.ToLower(key)
	delimiter := "."
	prefix := key + delimiter

	i := c.Get(key)
	if i == nil {
		return nil
	}
	if isStringMapInterface(i) {
		val := i.(map[string]interface{})
		keys := c.AllKeys()
		for _, k := range keys {
			if !strings.HasPrefix(k, prefix) {
				continue
			}
			mk := strings.TrimPrefix(k, prefix)
			mk = strings.Split(mk, delimiter)[0]
			if _, exists := val[mk]; exists {
				continue
			}
			mv := c.Get(key + delimiter + mk)
			if mv == nil {
				continue
			}
			val[mk] = mv
		}
		i = val
	}
	return decode(i, defaultDecoderConfig(rawVal))
}

func isStringMapInterface(val interface{}) bool {
	vt := reflect.TypeOf(val)
	return vt.Kind() == reflect.Map &&
		vt.Key().Kind() == reflect.String &&
		vt.Elem().Kind() == reflect.Interface
}

// A wrapper around mapstructure.Decode that mimics the WeakDecode functionality
func decode(input interface{}, config *mapstructure.DecoderConfig) error {
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}

// defaultDecoderConfig returns default mapstructure.DecoderConfig with support
// of time.Duration values & string slices
func defaultDecoderConfig(output interface{}, opts ...viper.DecoderConfigOption) *mapstructure.DecoderConfig {
	c := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           output,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c

}
