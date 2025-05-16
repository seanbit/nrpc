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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func setup() {
	os.Setenv("NRPC_HEARTBEAT_INTERVAL", "123s")
	os.Setenv("NRPC_CONCURRENCY_TEST", "42")
	os.Setenv("NRPC_BUFFER_TEST", "14")
}

func shutdown() {
	os.Unsetenv("NRPC_HEARTBEAT_INTERVAL")
	os.Unsetenv("NRPC_CONCURRENCY_TEST")
	os.Unsetenv("NRPC_BUFFER_TEST")
}

func TestNewConfig(t *testing.T) {
	t.Parallel()

	cfg := viper.New()
	cfg.SetDefault("nrpc.buffer.agent.messages", 20)
	cfg.Set("nrpc.concurrency.handler.dispatch", 23)
	cfg.SetDefault("nrpc.no.default", "custom")

	tables := []struct {
		in  []*viper.Viper
		key string
		val interface{}
	}{
		{[]*viper.Viper{}, "nrpc.buffer.agent.messages", 100},
		{[]*viper.Viper{cfg}, "nrpc.buffer.agent.messages", 20},
		{[]*viper.Viper{}, "nrpc.no.default", nil},
		{[]*viper.Viper{cfg}, "nrpc.no.default", "custom"},
		{[]*viper.Viper{}, "nrpc.concurrency.handler.dispatch", 25},
		{[]*viper.Viper{cfg}, "nrpc.concurrency.handler.dispatch", 23},
		{[]*viper.Viper{}, "nrpc.heartbeat.interval", "123s"},
		{[]*viper.Viper{cfg}, "nrpc.heartbeat.interval", "123s"},
		{[]*viper.Viper{}, "nrpc.concurrency.test", "42"},
		{[]*viper.Viper{cfg}, "nrpc.concurrency.test", "42"},
		{[]*viper.Viper{}, "nrpc.buffer.test", "14"},
		{[]*viper.Viper{cfg}, "nrpc.buffer.test", "14"},
	}

	for _, table := range tables {
		t.Run(fmt.Sprintf("arguments:%d", len(table.in)), func(t *testing.T) {
			c := NewViperConfig(table.in...)
			assert.Equal(t, table.val, c.Get(table.key))
		})
	}
}

func TestGetDuration(t *testing.T) {
	t.Parallel()

	c := NewViperConfig()
	tables := []struct {
		key string
		val time.Duration
	}{
		{"nrpc.heartbeat.interval", 123 * time.Second},
		{"nrpc.cluster.sd.etcd.dialtimeout", 5 * time.Second},
		{"unexistent", time.Duration(0)},
	}

	for _, table := range tables {
		t.Run(fmt.Sprintf("key:%s val:%d", table.key, table.val), func(t *testing.T) {
			assert.Equal(t, table.val, c.GetDuration(table.key))
		})
	}
}

func TestGetString(t *testing.T) {
	t.Parallel()

	c := NewViperConfig()
	tables := []struct {
		key string
		val []string
	}{
		{"nrpc.cluster.sd.etcd.endpoints", []string{"localhost:2379"}},
		{"unexistent", nil},
	}

	for _, table := range tables {
		t.Run(fmt.Sprintf("key:%s val:%s", table.key, table.val), func(t *testing.T) {
			assert.Equal(t, table.val, c.GetStringSlice(table.key))
		})
	}
}

func TestGetInt(t *testing.T) {
	t.Parallel()

	c := NewViperConfig()
	tables := []struct {
		key string
		val int
	}{
		{"nrpc.buffer.agent.messages", 100},
		{"unexistent", 0},
	}

	for _, table := range tables {
		t.Run(fmt.Sprintf("key:%s val:%d", table.key, table.val), func(t *testing.T) {
			assert.Equal(t, table.val, c.GetInt(table.key))
		})
	}
}

func TestGetStringSlice(t *testing.T) {
	t.Parallel()

	c := NewViperConfig()
	tables := []struct {
		key string
		val []string
	}{
		{"nrpc.cluster.sd.etcd.endpoints", []string{"localhost:2379"}},
		{"unexistent", nil},
	}

	for _, table := range tables {
		t.Run(fmt.Sprintf("key:%s val:%s", table.key, table.val), func(t *testing.T) {
			assert.Equal(t, table.val, c.GetStringSlice(table.key))
		})
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	c := NewViperConfig()
	tables := []struct {
		key string
		val interface{}
	}{
		{"nrpc.cluster.rpc.client.nats.connectiontimeout", time.Duration(2000000000)},
		{"unexistent", nil},
	}

	for _, table := range tables {
		t.Run(fmt.Sprintf("key:%s val:%v", table.key, table.val), func(t *testing.T) {
			assert.Equal(t, table.val, c.Get(table.key))
		})
	}
}
