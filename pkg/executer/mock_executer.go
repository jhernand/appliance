// Code generated by MockGen. DO NOT EDIT.
// Source: executer.go

// Package executer is a generated GoMock package.
package executer

import (
	os "os"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockExecuter is a mock of Executer interface.
type MockExecuter struct {
	ctrl     *gomock.Controller
	recorder *MockExecuterMockRecorder
}

// MockExecuterMockRecorder is the mock recorder for MockExecuter.
type MockExecuterMockRecorder struct {
	mock *MockExecuter
}

// NewMockExecuter creates a new mock instance.
func NewMockExecuter(ctrl *gomock.Controller) *MockExecuter {
	mock := &MockExecuter{ctrl: ctrl}
	mock.recorder = &MockExecuterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExecuter) EXPECT() *MockExecuterMockRecorder {
	return m.recorder
}

// Execute mocks base method.
func (m *MockExecuter) Execute(command Command) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", command)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Execute indicates an expected call of Execute.
func (mr *MockExecuterMockRecorder) Execute(command interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockExecuter)(nil).Execute), command)
}

// TempFile mocks base method.
func (m *MockExecuter) TempFile(dir, pattern string) (*os.File, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TempFile", dir, pattern)
	ret0, _ := ret[0].(*os.File)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TempFile indicates an expected call of TempFile.
func (mr *MockExecuterMockRecorder) TempFile(dir, pattern interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TempFile", reflect.TypeOf((*MockExecuter)(nil).TempFile), dir, pattern)
}
