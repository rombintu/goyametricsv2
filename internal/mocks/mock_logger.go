package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	"go.uber.org/zap"
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

// Error mocks the Error method.
func (m *MockLogger) Error(msg string, fields ...zap.Field) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Error", msg, fields)
}

// Error indicates an expected call of Error.
func (mr *MockLoggerMockRecorder) Error(msg interface{}, fields ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{msg}, fields...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockLogger)(nil).Error), varargs...)
}

// Debug mocks the Debug method.
func (m *MockLogger) Debug(msg string, fields ...zap.Field) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Debug", msg, fields)
}

// Debug indicates an expected call of Debug.
func (mr *MockLoggerMockRecorder) Debug(msg interface{}, fields ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{msg}, fields...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockLogger)(nil).Debug), varargs...)
}
