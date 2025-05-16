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
	"github.com/seanbit/nrpc/constants"
	pcontext "github.com/seanbit/nrpc/context"
	"time"

	"github.com/seanbit/nrpc/cluster"
	"github.com/seanbit/nrpc/component"
	"github.com/seanbit/nrpc/config"
	"github.com/seanbit/nrpc/interfaces"
	"github.com/seanbit/nrpc/metrics"
	"github.com/seanbit/nrpc/router"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"
)

var DefaultApp NRpc

// Configure configures the app
func Configure(
	serverType string,
	serverMetadata map[string]string,
	cfgs ...*viper.Viper,
) {
	builder := NewBuilderWithConfigs(
		serverType,
		serverMetadata,
		config.NewViperConfig(cfgs...),
	)
	DefaultApp = builder.Build()
}

func GetDieChan() chan bool {
	return DefaultApp.GetDieChan()
}

func SetDebug(debug bool) {
	DefaultApp.SetDebug(debug)
}

func SetHeartbeatTime(interval time.Duration) {
	DefaultApp.SetHeartbeatTime(interval)
}

func GetServerID() string {
	return DefaultApp.GetServerID()
}

func GetMetricsReporters() []metrics.Reporter {
	return DefaultApp.GetMetricsReporters()
}

func GetServer() *cluster.Server {
	return DefaultApp.GetServer()
}

func GetServerByID(id string) (*cluster.Server, error) {
	return DefaultApp.GetServerByID(id)
}

func GetServersByType(t string) (map[string]*cluster.Server, error) {
	return DefaultApp.GetServersByType(t)
}

func GetServers() []*cluster.Server {
	return DefaultApp.GetServers()
}

func Start() {
	DefaultApp.Start()
}

func SetDictionary(dict map[string]uint16) error {
	return DefaultApp.SetDictionary(dict)
}

func AddRoute(serverType string, routingFunction router.RoutingFunc) error {
	return DefaultApp.AddRoute(serverType, routingFunction)
}

func Shutdown() {
	DefaultApp.Shutdown()
}

func IsRunning() bool {
	return DefaultApp.IsRunning()
}

func RPC(ctx context.Context, routeStr string, reply proto.Message, arg proto.Message) error {
	return DefaultApp.RPC(ctx, routeStr, reply, arg)
}

func RPCTo(ctx context.Context, serverID, routeStr string, reply proto.Message, arg proto.Message) error {
	return DefaultApp.RPCTo(ctx, serverID, routeStr, reply, arg)
}

// Notify calls a method in a different server
func Notify(ctx context.Context, routeStr string, arg proto.Message) error {
	return DefaultApp.Notify(ctx, routeStr, arg)
}

// NotifyTo send a rpc to a specific server
func NotifyTo(ctx context.Context, serverID, routeStr string, arg proto.Message) error {
	return DefaultApp.NotifyTo(ctx, serverID, routeStr, arg)
}

func Register(c component.Component, options ...component.Option) {
	DefaultApp.Register(c, options...)
}

func RegisterModule(module interfaces.Module, name string) error {
	return DefaultApp.RegisterModule(module, name)
}

func RegisterModuleAfter(module interfaces.Module, name string) error {
	return DefaultApp.RegisterModuleAfter(module, name)
}

func RegisterModuleBefore(module interfaces.Module, name string) error {
	return DefaultApp.RegisterModuleBefore(module, name)
}

func GetModule(name string) (interfaces.Module, error) {
	return DefaultApp.GetModule(name)
}

func NewShareKeyContext(ctx context.Context, shareID string) context.Context {
	if ctx == nil {
		ctx = context.TODO()
	}
	return pcontext.AddToPropagateCtx(ctx, constants.ShareIDKey, shareID)
}
