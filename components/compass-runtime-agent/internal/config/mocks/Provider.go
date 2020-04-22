// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	config "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/config"
	mock "github.com/stretchr/testify/mock"
)

// Provider is an autogenerated mock type for the Provider type
type Provider struct {
	mock.Mock
}

// GetConnectionConfig provides a mock function with given fields:
func (_m *Provider) GetConnectionConfig() (config.ConnectionConfig, error) {
	ret := _m.Called()

	var r0 config.ConnectionConfig
	if rf, ok := ret.Get(0).(func() config.ConnectionConfig); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(config.ConnectionConfig)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetRuntimeConfig provides a mock function with given fields:
func (_m *Provider) GetRuntimeConfig() (config.RuntimeConfig, error) {
	ret := _m.Called()

	var r0 config.RuntimeConfig
	if rf, ok := ret.Get(0).(func() config.RuntimeConfig); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(config.RuntimeConfig)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
