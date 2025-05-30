// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/seanbit/nrpc (interfaces: NRpc)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	cluster "github.com/seanbit/nrpc/cluster"
	component "github.com/seanbit/nrpc/component"
	interfaces "github.com/seanbit/nrpc/interfaces"
	metrics "github.com/seanbit/nrpc/metrics"
	router "github.com/seanbit/nrpc/router"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

// MockNRpc is a mock of NRpc interface.
type MockNRpc struct {
	ctrl     *gomock.Controller
	recorder *MockNRpcMockRecorder
}

// MockNRpcMockRecorder is the mock recorder for MockNRpc.
type MockNRpcMockRecorder struct {
	mock *MockNRpc
}

// NewMockNRpc creates a new mock instance.
func NewMockNRpc(ctrl *gomock.Controller) *MockNRpc {
	mock := &MockNRpc{ctrl: ctrl}
	mock.recorder = &MockNRpcMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNRpc) EXPECT() *MockNRpcMockRecorder {
	return m.recorder
}

// AddRoute mocks base method.
func (m *MockNRpc) AddRoute(arg0 string, arg1 router.RoutingFunc) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddRoute", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddRoute indicates an expected call of AddRoute.
func (mr *MockNRpcMockRecorder) AddRoute(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddRoute", reflect.TypeOf((*MockNRpc)(nil).AddRoute), arg0, arg1)
}

// GetDieChan mocks base method.
func (m *MockNRpc) GetDieChan() chan bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDieChan")
	ret0, _ := ret[0].(chan bool)
	return ret0
}

// GetDieChan indicates an expected call of GetDieChan.
func (mr *MockNRpcMockRecorder) GetDieChan() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDieChan", reflect.TypeOf((*MockNRpc)(nil).GetDieChan))
}

// GetMetricsReporters mocks base method.
func (m *MockNRpc) GetMetricsReporters() []metrics.Reporter {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetricsReporters")
	ret0, _ := ret[0].([]metrics.Reporter)
	return ret0
}

// GetMetricsReporters indicates an expected call of GetMetricsReporters.
func (mr *MockNRpcMockRecorder) GetMetricsReporters() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetricsReporters", reflect.TypeOf((*MockNRpc)(nil).GetMetricsReporters))
}

// GetModule mocks base method.
func (m *MockNRpc) GetModule(arg0 string) (interfaces.Module, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetModule", arg0)
	ret0, _ := ret[0].(interfaces.Module)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetModule indicates an expected call of GetModule.
func (mr *MockNRpcMockRecorder) GetModule(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetModule", reflect.TypeOf((*MockNRpc)(nil).GetModule), arg0)
}

// GetServer mocks base method.
func (m *MockNRpc) GetServer() *cluster.Server {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetServer")
	ret0, _ := ret[0].(*cluster.Server)
	return ret0
}

// GetServer indicates an expected call of GetServer.
func (mr *MockNRpcMockRecorder) GetServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetServer", reflect.TypeOf((*MockNRpc)(nil).GetServer))
}

// GetServerByID mocks base method.
func (m *MockNRpc) GetServerByID(arg0 string) (*cluster.Server, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetServerByID", arg0)
	ret0, _ := ret[0].(*cluster.Server)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetServerByID indicates an expected call of GetServerByID.
func (mr *MockNRpcMockRecorder) GetServerByID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetServerByID", reflect.TypeOf((*MockNRpc)(nil).GetServerByID), arg0)
}

// GetServerID mocks base method.
func (m *MockNRpc) GetServerID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetServerID")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetServerID indicates an expected call of GetServerID.
func (mr *MockNRpcMockRecorder) GetServerID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetServerID", reflect.TypeOf((*MockNRpc)(nil).GetServerID))
}

// GetServers mocks base method.
func (m *MockNRpc) GetServers() []*cluster.Server {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetServers")
	ret0, _ := ret[0].([]*cluster.Server)
	return ret0
}

// GetServers indicates an expected call of GetServers.
func (mr *MockNRpcMockRecorder) GetServers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetServers", reflect.TypeOf((*MockNRpc)(nil).GetServers))
}

// GetServersByType mocks base method.
func (m *MockNRpc) GetServersByType(arg0 string) (map[string]*cluster.Server, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetServersByType", arg0)
	ret0, _ := ret[0].(map[string]*cluster.Server)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetServersByType indicates an expected call of GetServersByType.
func (mr *MockNRpcMockRecorder) GetServersByType(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetServersByType", reflect.TypeOf((*MockNRpc)(nil).GetServersByType), arg0)
}

// IsRunning mocks base method.
func (m *MockNRpc) IsRunning() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsRunning")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsRunning indicates an expected call of IsRunning.
func (mr *MockNRpcMockRecorder) IsRunning() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsRunning", reflect.TypeOf((*MockNRpc)(nil).IsRunning))
}

// RPC mocks base method.
func (m *MockNRpc) RPC(arg0 context.Context, arg1 string, arg2, arg3 protoreflect.ProtoMessage) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RPC", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// RPC indicates an expected call of RPC.
func (mr *MockNRpcMockRecorder) RPC(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RPC", reflect.TypeOf((*MockNRpc)(nil).RPC), arg0, arg1, arg2, arg3)
}

// RPCTo mocks base method.
func (m *MockNRpc) RPCTo(arg0 context.Context, arg1, arg2 string, arg3, arg4 protoreflect.ProtoMessage) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RPCTo", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// RPCTo indicates an expected call of RPCTo.
func (mr *MockNRpcMockRecorder) RPCTo(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RPCTo", reflect.TypeOf((*MockNRpc)(nil).RPCTo), arg0, arg1, arg2, arg3, arg4)
}

// Register mocks base method.
func (m *MockNRpc) Register(arg0 component.Component, arg1 ...component.Option) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Register", varargs...)
}

// Register indicates an expected call of Register.
func (mr *MockNRpcMockRecorder) Register(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockNRpc)(nil).Register), varargs...)
}

// RegisterModule mocks base method.
func (m *MockNRpc) RegisterModule(arg0 interfaces.Module, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterModule", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegisterModule indicates an expected call of RegisterModule.
func (mr *MockNRpcMockRecorder) RegisterModule(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterModule", reflect.TypeOf((*MockNRpc)(nil).RegisterModule), arg0, arg1)
}

// RegisterModuleAfter mocks base method.
func (m *MockNRpc) RegisterModuleAfter(arg0 interfaces.Module, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterModuleAfter", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegisterModuleAfter indicates an expected call of RegisterModuleAfter.
func (mr *MockNRpcMockRecorder) RegisterModuleAfter(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterModuleAfter", reflect.TypeOf((*MockNRpc)(nil).RegisterModuleAfter), arg0, arg1)
}

// RegisterModuleBefore mocks base method.
func (m *MockNRpc) RegisterModuleBefore(arg0 interfaces.Module, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterModuleBefore", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegisterModuleBefore indicates an expected call of RegisterModuleBefore.
func (mr *MockNRpcMockRecorder) RegisterModuleBefore(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterModuleBefore", reflect.TypeOf((*MockNRpc)(nil).RegisterModuleBefore), arg0, arg1)
}

// RegisterRemote mocks base method.
func (m *MockNRpc) RegisterRemote(arg0 component.Component, arg1 ...component.Option) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "RegisterRemote", varargs...)
}

// RegisterRemote indicates an expected call of RegisterRemote.
func (mr *MockNRpcMockRecorder) RegisterRemote(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterRemote", reflect.TypeOf((*MockNRpc)(nil).RegisterRemote), varargs...)
}

// SetDebug mocks base method.
func (m *MockNRpc) SetDebug(arg0 bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetDebug", arg0)
}

// SetDebug indicates an expected call of SetDebug.
func (mr *MockNRpcMockRecorder) SetDebug(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetDebug", reflect.TypeOf((*MockNRpc)(nil).SetDebug), arg0)
}

// SetDictionary mocks base method.
func (m *MockNRpc) SetDictionary(arg0 map[string]uint16) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetDictionary", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetDictionary indicates an expected call of SetDictionary.
func (mr *MockNRpcMockRecorder) SetDictionary(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetDictionary", reflect.TypeOf((*MockNRpc)(nil).SetDictionary), arg0)
}

// SetHeartbeatTime mocks base method.
func (m *MockNRpc) SetHeartbeatTime(arg0 time.Duration) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetHeartbeatTime", arg0)
}

// SetHeartbeatTime indicates an expected call of SetHeartbeatTime.
func (mr *MockNRpcMockRecorder) SetHeartbeatTime(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeartbeatTime", reflect.TypeOf((*MockNRpc)(nil).SetHeartbeatTime), arg0)
}

// Shutdown mocks base method.
func (m *MockNRpc) Shutdown() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Shutdown")
}

// Shutdown indicates an expected call of Shutdown.
func (mr *MockNRpcMockRecorder) Shutdown() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*MockNRpc)(nil).Shutdown))
}

// Start mocks base method.
func (m *MockNRpc) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start.
func (mr *MockNRpcMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockNRpc)(nil).Start))
}
