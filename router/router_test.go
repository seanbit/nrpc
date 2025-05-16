package router

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/seanbit/nrpc/cluster"
	"github.com/seanbit/nrpc/cluster/mocks"
	"github.com/seanbit/nrpc/conn/message"
	"github.com/seanbit/nrpc/constants"
	"github.com/seanbit/nrpc/route"
	"github.com/stretchr/testify/assert"
)

var (
	serverID   = "id"
	serverType = "serverType"
	server     = cluster.NewServer(serverID, serverType)
	servers    = map[string]*cluster.Server{
		serverID: server,
	}

	routingFunction = func(
		ctx context.Context,
		route *route.Route,
		payload []byte,
		servers map[string]*cluster.Server,
	) (context.Context, *cluster.Server, error) {
		return ctx, server, nil
	}
)

var routerTables = map[string]struct {
	server     *cluster.Server
	serverType string
	err        error
}{
	"test_server_has_route_func":   {server, serverType, nil},
	"test_server_use_default_func": {server, "notRegisteredType", nil},
	"test_user_use_default_func":   {server, serverType, nil},
	"test_error_on_service_disc":   {nil, serverType, errors.New("sd error")},
}

var addRouteRouterTables = map[string]struct {
	serverType string
}{
	"test_overrige_server_type": {serverType},
	"test_new_server_type":      {"notRegisteredType"},
}

func TestNew(t *testing.T) {
	t.Parallel()
	router := New()
	assert.NotNil(t, router)
}

func TestDefaultRoute(t *testing.T) {
	t.Parallel()

	router := New()

	retServer := router.defaultRoute(servers)
	assert.Equal(t, server, retServer)
}

func TestRoute(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	route := route.NewRoute(serverType, "service", "method")

	for name, table := range routerTables {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockServiceDiscovery := mocks.NewMockServiceDiscovery(ctrl)
			mockServiceDiscovery.EXPECT().
				GetServersByType(table.serverType).
				Return(servers, table.err)

			router := New()
			router.AddRoute(serverType, routingFunction)
			router.SetServiceDiscovery(mockServiceDiscovery)

			_, retServer, err := router.Route(ctx, table.serverType, route, &message.Message{
				Data: []byte{0x01},
			})
			assert.Equal(t, table.server, retServer)
			assert.Equal(t, table.err, err)
		})
	}
}

func TestAddRoute(t *testing.T) {
	t.Parallel()

	for name, table := range addRouteRouterTables {
		t.Run(name, func(t *testing.T) {
			router := New()
			router.AddRoute(table.serverType, routingFunction)

			assert.NotNil(t, router.routesMap[table.serverType])
			assert.Nil(t, router.routesMap["anotherServerType"])
		})
	}
}

func TestRouteFailIfNullServiceDiscovery(t *testing.T) {
	t.Parallel()

	router := New()
	_, _, err := router.Route(context.Background(), serverType, route.NewRoute(serverType, "service", "method"), &message.Message{
		Data: []byte{0x01},
	})
	assert.Equal(t, constants.ErrServiceDiscoveryNotInitialized, err)
}
