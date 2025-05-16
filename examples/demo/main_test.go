package main

import (
	"github.com/seanbit/nrpc"
	"github.com/seanbit/nrpc/protos/test"
	"testing"
)

func TestClientRpcCall(t *testing.T) {
	StartServer("test_client", true)
	defer nrpc.Shutdown()
	ctx := nrpc.NewShareKeyContext(nil, "uid001")
	route := "test_server.serv.requesthandle"
	reply := &test.SomeResponse{}
	args := &test.SomeRequest{
		A: 1,
		B: "test call params",
	}
	if err := nrpc.RPC(ctx, route, reply, args); err != nil {
		t.Fatal(err)
	}
	t.Logf("reply:%v", reply.String())
}

func TestClientNotifyCall(t *testing.T) {
	StartServer("test_client", true)
	defer nrpc.Shutdown()
	ctx := nrpc.NewShareKeyContext(nil, "uid001")
	route := "test_server.serv.notifyhandle"
	args := &test.SomeRequest{
		A: 2,
		B: "test notify params",
	}
	if err := nrpc.Notify(ctx, route, args); err != nil {
		t.Fatal(err)
	}
	t.Log("notify success")
}
