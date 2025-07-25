// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/rudderlabs/rudder-server/services/streammanager/kinesis (interfaces: KinesisClientV1)
//
// Generated by this command:
//
//	mockgen -destination=../../../mocks/services/streammanager/kinesis_v1/mock_kinesis_v1.go -package mock_kinesis_v1 github.com/rudderlabs/rudder-server/services/streammanager/kinesis KinesisClientV1
//

// Package mock_kinesis_v1 is a generated GoMock package.
package mock_kinesis_v1

import (
	reflect "reflect"

	kinesis "github.com/aws/aws-sdk-go/service/kinesis"
	gomock "go.uber.org/mock/gomock"
)

// MockKinesisClientV1 is a mock of KinesisClientV1 interface.
type MockKinesisClientV1 struct {
	ctrl     *gomock.Controller
	recorder *MockKinesisClientV1MockRecorder
	isgomock struct{}
}

// MockKinesisClientV1MockRecorder is the mock recorder for MockKinesisClientV1.
type MockKinesisClientV1MockRecorder struct {
	mock *MockKinesisClientV1
}

// NewMockKinesisClientV1 creates a new mock instance.
func NewMockKinesisClientV1(ctrl *gomock.Controller) *MockKinesisClientV1 {
	mock := &MockKinesisClientV1{ctrl: ctrl}
	mock.recorder = &MockKinesisClientV1MockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockKinesisClientV1) EXPECT() *MockKinesisClientV1MockRecorder {
	return m.recorder
}

// PutRecord mocks base method.
func (m *MockKinesisClientV1) PutRecord(input *kinesis.PutRecordInput) (*kinesis.PutRecordOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutRecord", input)
	ret0, _ := ret[0].(*kinesis.PutRecordOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PutRecord indicates an expected call of PutRecord.
func (mr *MockKinesisClientV1MockRecorder) PutRecord(input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutRecord", reflect.TypeOf((*MockKinesisClientV1)(nil).PutRecord), input)
}
