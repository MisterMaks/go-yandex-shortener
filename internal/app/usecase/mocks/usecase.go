// Code generated by MockGen. DO NOT EDIT.
// Source: internal/app/usecase/usecase.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	app "github.com/MisterMaks/go-yandex-shortener/internal/app"
	gomock "github.com/golang/mock/gomock"
)

// MockAppRepoInterface is a mock of AppRepoInterface interface.
type MockAppRepoInterface struct {
	ctrl     *gomock.Controller
	recorder *MockAppRepoInterfaceMockRecorder
}

// MockAppRepoInterfaceMockRecorder is the mock recorder for MockAppRepoInterface.
type MockAppRepoInterfaceMockRecorder struct {
	mock *MockAppRepoInterface
}

// NewMockAppRepoInterface creates a new mock instance.
func NewMockAppRepoInterface(ctrl *gomock.Controller) *MockAppRepoInterface {
	mock := &MockAppRepoInterface{ctrl: ctrl}
	mock.recorder = &MockAppRepoInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAppRepoInterface) EXPECT() *MockAppRepoInterfaceMockRecorder {
	return m.recorder
}

// CheckIDExistence mocks base method.
func (m *MockAppRepoInterface) CheckIDExistence(id string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckIDExistence", id)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckIDExistence indicates an expected call of CheckIDExistence.
func (mr *MockAppRepoInterfaceMockRecorder) CheckIDExistence(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckIDExistence", reflect.TypeOf((*MockAppRepoInterface)(nil).CheckIDExistence), id)
}

// Close mocks base method.
func (m *MockAppRepoInterface) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockAppRepoInterfaceMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockAppRepoInterface)(nil).Close))
}

// DeleteUserURLs mocks base method.
func (m *MockAppRepoInterface) DeleteUserURLs(urls []*app.URL) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUserURLs", urls)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUserURLs indicates an expected call of DeleteUserURLs.
func (mr *MockAppRepoInterfaceMockRecorder) DeleteUserURLs(urls interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUserURLs", reflect.TypeOf((*MockAppRepoInterface)(nil).DeleteUserURLs), urls)
}

// GetOrCreateURL mocks base method.
func (m *MockAppRepoInterface) GetOrCreateURL(id, rawURL string, userID uint) (*app.URL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrCreateURL", id, rawURL, userID)
	ret0, _ := ret[0].(*app.URL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrCreateURL indicates an expected call of GetOrCreateURL.
func (mr *MockAppRepoInterfaceMockRecorder) GetOrCreateURL(id, rawURL, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrCreateURL", reflect.TypeOf((*MockAppRepoInterface)(nil).GetOrCreateURL), id, rawURL, userID)
}

// GetOrCreateURLs mocks base method.
func (m *MockAppRepoInterface) GetOrCreateURLs(urls []*app.URL) ([]*app.URL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrCreateURLs", urls)
	ret0, _ := ret[0].([]*app.URL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrCreateURLs indicates an expected call of GetOrCreateURLs.
func (mr *MockAppRepoInterfaceMockRecorder) GetOrCreateURLs(urls interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrCreateURLs", reflect.TypeOf((*MockAppRepoInterface)(nil).GetOrCreateURLs), urls)
}

// GetURL mocks base method.
func (m *MockAppRepoInterface) GetURL(id string) (*app.URL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetURL", id)
	ret0, _ := ret[0].(*app.URL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetURL indicates an expected call of GetURL.
func (mr *MockAppRepoInterfaceMockRecorder) GetURL(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetURL", reflect.TypeOf((*MockAppRepoInterface)(nil).GetURL), id)
}

// GetUserURLs mocks base method.
func (m *MockAppRepoInterface) GetUserURLs(userID uint) ([]*app.URL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserURLs", userID)
	ret0, _ := ret[0].([]*app.URL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserURLs indicates an expected call of GetUserURLs.
func (mr *MockAppRepoInterfaceMockRecorder) GetUserURLs(userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserURLs", reflect.TypeOf((*MockAppRepoInterface)(nil).GetUserURLs), userID)
}
