// Code generated by MockGen. DO NOT EDIT.
// Source: cmd/shortener/main.go

// Package mocks is a generated GoMock package.
package mocks

import (
	http "net/http"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockAppHandlerInterface is a mock of AppHandlerInterface interface.
type MockAppHandlerInterface struct {
	ctrl     *gomock.Controller
	recorder *MockAppHandlerInterfaceMockRecorder
}

// MockAppHandlerInterfaceMockRecorder is the mock recorder for MockAppHandlerInterface.
type MockAppHandlerInterfaceMockRecorder struct {
	mock *MockAppHandlerInterface
}

// NewMockAppHandlerInterface creates a new mock instance.
func NewMockAppHandlerInterface(ctrl *gomock.Controller) *MockAppHandlerInterface {
	mock := &MockAppHandlerInterface{ctrl: ctrl}
	mock.recorder = &MockAppHandlerInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAppHandlerInterface) EXPECT() *MockAppHandlerInterfaceMockRecorder {
	return m.recorder
}

// APIDeleteUserURLs mocks base method.
func (m *MockAppHandlerInterface) APIDeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "APIDeleteUserURLs", w, r)
}

// APIDeleteUserURLs indicates an expected call of APIDeleteUserURLs.
func (mr *MockAppHandlerInterfaceMockRecorder) APIDeleteUserURLs(w, r interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "APIDeleteUserURLs", reflect.TypeOf((*MockAppHandlerInterface)(nil).APIDeleteUserURLs), w, r)
}

// APIGetInternalStats mocks base method.
func (m *MockAppHandlerInterface) APIGetInternalStats(w http.ResponseWriter, r *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "APIGetInternalStats", w, r)
}

// APIGetInternalStats indicates an expected call of APIGetInternalStats.
func (mr *MockAppHandlerInterfaceMockRecorder) APIGetInternalStats(w, r interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "APIGetInternalStats", reflect.TypeOf((*MockAppHandlerInterface)(nil).APIGetInternalStats), w, r)
}

// APIGetOrCreateURL mocks base method.
func (m *MockAppHandlerInterface) APIGetOrCreateURL(w http.ResponseWriter, r *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "APIGetOrCreateURL", w, r)
}

// APIGetOrCreateURL indicates an expected call of APIGetOrCreateURL.
func (mr *MockAppHandlerInterfaceMockRecorder) APIGetOrCreateURL(w, r interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "APIGetOrCreateURL", reflect.TypeOf((*MockAppHandlerInterface)(nil).APIGetOrCreateURL), w, r)
}

// APIGetOrCreateURLs mocks base method.
func (m *MockAppHandlerInterface) APIGetOrCreateURLs(w http.ResponseWriter, r *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "APIGetOrCreateURLs", w, r)
}

// APIGetOrCreateURLs indicates an expected call of APIGetOrCreateURLs.
func (mr *MockAppHandlerInterfaceMockRecorder) APIGetOrCreateURLs(w, r interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "APIGetOrCreateURLs", reflect.TypeOf((*MockAppHandlerInterface)(nil).APIGetOrCreateURLs), w, r)
}

// APIGetUserURLs mocks base method.
func (m *MockAppHandlerInterface) APIGetUserURLs(w http.ResponseWriter, r *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "APIGetUserURLs", w, r)
}

// APIGetUserURLs indicates an expected call of APIGetUserURLs.
func (mr *MockAppHandlerInterfaceMockRecorder) APIGetUserURLs(w, r interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "APIGetUserURLs", reflect.TypeOf((*MockAppHandlerInterface)(nil).APIGetUserURLs), w, r)
}

// GetOrCreateURL mocks base method.
func (m *MockAppHandlerInterface) GetOrCreateURL(w http.ResponseWriter, r *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "GetOrCreateURL", w, r)
}

// GetOrCreateURL indicates an expected call of GetOrCreateURL.
func (mr *MockAppHandlerInterfaceMockRecorder) GetOrCreateURL(w, r interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrCreateURL", reflect.TypeOf((*MockAppHandlerInterface)(nil).GetOrCreateURL), w, r)
}

// Ping mocks base method.
func (m *MockAppHandlerInterface) Ping(w http.ResponseWriter, r *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Ping", w, r)
}

// Ping indicates an expected call of Ping.
func (mr *MockAppHandlerInterfaceMockRecorder) Ping(w, r interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockAppHandlerInterface)(nil).Ping), w, r)
}

// RedirectToURL mocks base method.
func (m *MockAppHandlerInterface) RedirectToURL(w http.ResponseWriter, r *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RedirectToURL", w, r)
}

// RedirectToURL indicates an expected call of RedirectToURL.
func (mr *MockAppHandlerInterfaceMockRecorder) RedirectToURL(w, r interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RedirectToURL", reflect.TypeOf((*MockAppHandlerInterface)(nil).RedirectToURL), w, r)
}
