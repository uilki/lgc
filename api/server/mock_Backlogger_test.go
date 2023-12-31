// Code generated by mockery v2.30.1. DO NOT EDIT.

package server

import (
	mock "github.com/stretchr/testify/mock"
	generated "github.com/uilki/lgc/api/server/generated"
)

// MockBacklogger is an autogenerated mock type for the Backlogger type
type MockBacklogger struct {
	mock.Mock
}

type MockBacklogger_Expecter struct {
	mock *mock.Mock
}

func (_m *MockBacklogger) EXPECT() *MockBacklogger_Expecter {
	return &MockBacklogger_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with given fields:
func (_m *MockBacklogger) Close() {
	_m.Called()
}

// MockBacklogger_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type MockBacklogger_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *MockBacklogger_Expecter) Close() *MockBacklogger_Close_Call {
	return &MockBacklogger_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *MockBacklogger_Close_Call) Run(run func()) *MockBacklogger_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockBacklogger_Close_Call) Return() *MockBacklogger_Close_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockBacklogger_Close_Call) RunAndReturn(run func()) *MockBacklogger_Close_Call {
	_c.Call.Return(run)
	return _c
}

// GetHistory provides a mock function with given fields:
func (_m *MockBacklogger) GetHistory() ([]generated.Message, error) {
	ret := _m.Called()

	var r0 []generated.Message
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]generated.Message, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []generated.Message); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]generated.Message)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockBacklogger_GetHistory_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetHistory'
type MockBacklogger_GetHistory_Call struct {
	*mock.Call
}

// GetHistory is a helper method to define mock.On call
func (_e *MockBacklogger_Expecter) GetHistory() *MockBacklogger_GetHistory_Call {
	return &MockBacklogger_GetHistory_Call{Call: _e.mock.On("GetHistory")}
}

func (_c *MockBacklogger_GetHistory_Call) Run(run func()) *MockBacklogger_GetHistory_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockBacklogger_GetHistory_Call) Return(_a0 []generated.Message, _a1 error) *MockBacklogger_GetHistory_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockBacklogger_GetHistory_Call) RunAndReturn(run func() ([]generated.Message, error)) *MockBacklogger_GetHistory_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: _a0
func (_m *MockBacklogger) Update(_a0 *generated.Message) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*generated.Message) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockBacklogger_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type MockBacklogger_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - _a0 *generated.Message
func (_e *MockBacklogger_Expecter) Update(_a0 interface{}) *MockBacklogger_Update_Call {
	return &MockBacklogger_Update_Call{Call: _e.mock.On("Update", _a0)}
}

func (_c *MockBacklogger_Update_Call) Run(run func(_a0 *generated.Message)) *MockBacklogger_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*generated.Message))
	})
	return _c
}

func (_c *MockBacklogger_Update_Call) Return(_a0 error) *MockBacklogger_Update_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockBacklogger_Update_Call) RunAndReturn(run func(*generated.Message) error) *MockBacklogger_Update_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockBacklogger creates a new instance of MockBacklogger. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockBacklogger(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockBacklogger {
	mock := &MockBacklogger{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
