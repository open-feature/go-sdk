// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/openfeature/provider.go

// Package openfeature is a generated GoMock package.
package openfeature

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockFeatureProvider is a mock of FeatureProvider interface.
type MockFeatureProvider struct {
	ctrl     *gomock.Controller
	recorder *MockFeatureProviderMockRecorder
}

// MockFeatureProviderMockRecorder is the mock recorder for MockFeatureProvider.
type MockFeatureProviderMockRecorder struct {
	mock *MockFeatureProvider
}

// NewMockFeatureProvider creates a new mock instance.
func NewMockFeatureProvider(ctrl *gomock.Controller) *MockFeatureProvider {
	mock := &MockFeatureProvider{ctrl: ctrl}
	mock.recorder = &MockFeatureProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFeatureProvider) EXPECT() *MockFeatureProviderMockRecorder {
	return m.recorder
}

// BooleanEvaluation mocks base method.
func (m *MockFeatureProvider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, evalCtx FlattenedContext) BoolResolutionDetail {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BooleanEvaluation", ctx, flag, defaultValue, evalCtx)
	ret0, _ := ret[0].(BoolResolutionDetail)
	return ret0
}

// BooleanEvaluation indicates an expected call of BooleanEvaluation.
func (mr *MockFeatureProviderMockRecorder) BooleanEvaluation(ctx, flag, defaultValue, evalCtx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BooleanEvaluation", reflect.TypeOf((*MockFeatureProvider)(nil).BooleanEvaluation), ctx, flag, defaultValue, evalCtx)
}

// FloatEvaluation mocks base method.
func (m *MockFeatureProvider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, evalCtx FlattenedContext) FloatResolutionDetail {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FloatEvaluation", ctx, flag, defaultValue, evalCtx)
	ret0, _ := ret[0].(FloatResolutionDetail)
	return ret0
}

// FloatEvaluation indicates an expected call of FloatEvaluation.
func (mr *MockFeatureProviderMockRecorder) FloatEvaluation(ctx, flag, defaultValue, evalCtx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FloatEvaluation", reflect.TypeOf((*MockFeatureProvider)(nil).FloatEvaluation), ctx, flag, defaultValue, evalCtx)
}

// Hooks mocks base method.
func (m *MockFeatureProvider) Hooks() []Hook {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Hooks")
	ret0, _ := ret[0].([]Hook)
	return ret0
}

// Hooks indicates an expected call of Hooks.
func (mr *MockFeatureProviderMockRecorder) Hooks() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Hooks", reflect.TypeOf((*MockFeatureProvider)(nil).Hooks))
}

// IntEvaluation mocks base method.
func (m *MockFeatureProvider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, evalCtx FlattenedContext) IntResolutionDetail {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IntEvaluation", ctx, flag, defaultValue, evalCtx)
	ret0, _ := ret[0].(IntResolutionDetail)
	return ret0
}

// IntEvaluation indicates an expected call of IntEvaluation.
func (mr *MockFeatureProviderMockRecorder) IntEvaluation(ctx, flag, defaultValue, evalCtx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IntEvaluation", reflect.TypeOf((*MockFeatureProvider)(nil).IntEvaluation), ctx, flag, defaultValue, evalCtx)
}

// Metadata mocks base method.
func (m *MockFeatureProvider) Metadata() Metadata {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Metadata")
	ret0, _ := ret[0].(Metadata)
	return ret0
}

// Metadata indicates an expected call of Metadata.
func (mr *MockFeatureProviderMockRecorder) Metadata() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Metadata", reflect.TypeOf((*MockFeatureProvider)(nil).Metadata))
}

// ObjectEvaluation mocks base method.
func (m *MockFeatureProvider) ObjectEvaluation(ctx context.Context, flag string, defaultValue interface{}, evalCtx FlattenedContext) InterfaceResolutionDetail {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ObjectEvaluation", ctx, flag, defaultValue, evalCtx)
	ret0, _ := ret[0].(InterfaceResolutionDetail)
	return ret0
}

// ObjectEvaluation indicates an expected call of ObjectEvaluation.
func (mr *MockFeatureProviderMockRecorder) ObjectEvaluation(ctx, flag, defaultValue, evalCtx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObjectEvaluation", reflect.TypeOf((*MockFeatureProvider)(nil).ObjectEvaluation), ctx, flag, defaultValue, evalCtx)
}

// StringEvaluation mocks base method.
func (m *MockFeatureProvider) StringEvaluation(ctx context.Context, flag, defaultValue string, evalCtx FlattenedContext) StringResolutionDetail {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StringEvaluation", ctx, flag, defaultValue, evalCtx)
	ret0, _ := ret[0].(StringResolutionDetail)
	return ret0
}

// StringEvaluation indicates an expected call of StringEvaluation.
func (mr *MockFeatureProviderMockRecorder) StringEvaluation(ctx, flag, defaultValue, evalCtx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StringEvaluation", reflect.TypeOf((*MockFeatureProvider)(nil).StringEvaluation), ctx, flag, defaultValue, evalCtx)
}

// MockStateHandler is a mock of StateHandler interface.
type MockStateHandler struct {
	ctrl     *gomock.Controller
	recorder *MockStateHandlerMockRecorder
}

// MockStateHandlerMockRecorder is the mock recorder for MockStateHandler.
type MockStateHandlerMockRecorder struct {
	mock *MockStateHandler
}

// NewMockStateHandler creates a new mock instance.
func NewMockStateHandler(ctrl *gomock.Controller) *MockStateHandler {
	mock := &MockStateHandler{ctrl: ctrl}
	mock.recorder = &MockStateHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStateHandler) EXPECT() *MockStateHandlerMockRecorder {
	return m.recorder
}

// Init mocks base method.
func (m *MockStateHandler) Init(evaluationContext EvaluationContext) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Init", evaluationContext)
}

// Init indicates an expected call of Init.
func (mr *MockStateHandlerMockRecorder) Init(evaluationContext interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Init", reflect.TypeOf((*MockStateHandler)(nil).Init), evaluationContext)
}

// Shutdown mocks base method.
func (m *MockStateHandler) Shutdown() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Shutdown")
}

// Shutdown indicates an expected call of Shutdown.
func (mr *MockStateHandlerMockRecorder) Shutdown() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*MockStateHandler)(nil).Shutdown))
}

// Status mocks base method.
func (m *MockStateHandler) Status() State {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Status")
	ret0, _ := ret[0].(State)
	return ret0
}

// Status indicates an expected call of Status.
func (mr *MockStateHandlerMockRecorder) Status() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Status", reflect.TypeOf((*MockStateHandler)(nil).Status))
}

// MockState is a mock of State interface.
type MockState struct {
	ctrl     *gomock.Controller
	recorder *MockStateMockRecorder
}

// MockStateMockRecorder is the mock recorder for MockState.
type MockStateMockRecorder struct {
	mock *MockState
}

// NewMockState creates a new mock instance.
func NewMockState(ctrl *gomock.Controller) *MockState {
	mock := &MockState{ctrl: ctrl}
	mock.recorder = &MockStateMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockState) EXPECT() *MockStateMockRecorder {
	return m.recorder
}

// get mocks base method.
func (m *MockState) get() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "get")
	ret0, _ := ret[0].(string)
	return ret0
}

// get indicates an expected call of get.
func (mr *MockStateMockRecorder) get() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "get", reflect.TypeOf((*MockState)(nil).get))
}
