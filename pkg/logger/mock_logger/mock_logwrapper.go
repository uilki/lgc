// Code generated by MockGen. DO NOT EDIT.
// Source: logwrapper.go

// Package mock_logger is a generated GoMock package.
package mock_logger

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockLogger is a mock of Logger interface.
type MockLogger struct {
	ctrl     *gomock.Controller
	recorder *MockLoggerMockRecorder
}

// MockLoggerMockRecorder is the mock recorder for MockLogger.
type MockLoggerMockRecorder struct {
	mock *MockLogger
}

// NewMockLogger creates a new mock instance.
func NewMockLogger(ctrl *gomock.Controller) *MockLogger {
	mock := &MockLogger{ctrl: ctrl}
	mock.recorder = &MockLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLogger) EXPECT() *MockLoggerMockRecorder {
	return m.recorder
}

// LogDebug mocks base method.
func (m *MockLogger) LogDebug(message string, arg ...any) {
	m.ctrl.T.Helper()
	varargs := []interface{}{message}
	for _, a := range arg {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "LogDebug", varargs...)
}

// LogDebug indicates an expected call of LogDebug.
func (mr *MockLoggerMockRecorder) LogDebug(message interface{}, arg ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{message}, arg...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LogDebug", reflect.TypeOf((*MockLogger)(nil).LogDebug), varargs...)
}

// LogError mocks base method.
func (m *MockLogger) LogError(e interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "LogError", e)
}

// LogError indicates an expected call of LogError.
func (mr *MockLoggerMockRecorder) LogError(e interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LogError", reflect.TypeOf((*MockLogger)(nil).LogError), e)
}

// LogInfo mocks base method.
func (m *MockLogger) LogInfo(info interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "LogInfo", info)
}

// LogInfo indicates an expected call of LogInfo.
func (mr *MockLoggerMockRecorder) LogInfo(info interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LogInfo", reflect.TypeOf((*MockLogger)(nil).LogInfo), info)
}
