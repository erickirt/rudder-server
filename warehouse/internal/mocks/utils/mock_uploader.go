// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/rudderlabs/rudder-server/warehouse/utils (interfaces: Uploader)
//
// Generated by this command:
//
//	mockgen -destination=../internal/mocks/utils/mock_uploader.go -package mock_uploader github.com/rudderlabs/rudder-server/warehouse/utils Uploader
//

// Package mock_uploader is a generated GoMock package.
package mock_uploader

import (
	context "context"
	reflect "reflect"

	model "github.com/rudderlabs/rudder-server/warehouse/internal/model"
	warehouseutils "github.com/rudderlabs/rudder-server/warehouse/utils"
	gomock "go.uber.org/mock/gomock"
)

// MockUploader is a mock of Uploader interface.
type MockUploader struct {
	ctrl     *gomock.Controller
	recorder *MockUploaderMockRecorder
	isgomock struct{}
}

// MockUploaderMockRecorder is the mock recorder for MockUploader.
type MockUploaderMockRecorder struct {
	mock *MockUploader
}

// NewMockUploader creates a new mock instance.
func NewMockUploader(ctrl *gomock.Controller) *MockUploader {
	mock := &MockUploader{ctrl: ctrl}
	mock.recorder = &MockUploaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUploader) EXPECT() *MockUploaderMockRecorder {
	return m.recorder
}

// CanAppend mocks base method.
func (m *MockUploader) CanAppend() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CanAppend")
	ret0, _ := ret[0].(bool)
	return ret0
}

// CanAppend indicates an expected call of CanAppend.
func (mr *MockUploaderMockRecorder) CanAppend() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CanAppend", reflect.TypeOf((*MockUploader)(nil).CanAppend))
}

// GetLoadFileType mocks base method.
func (m *MockUploader) GetLoadFileType() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLoadFileType")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetLoadFileType indicates an expected call of GetLoadFileType.
func (mr *MockUploaderMockRecorder) GetLoadFileType() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLoadFileType", reflect.TypeOf((*MockUploader)(nil).GetLoadFileType))
}

// GetLoadFilesMetadata mocks base method.
func (m *MockUploader) GetLoadFilesMetadata(ctx context.Context, options warehouseutils.GetLoadFilesOptions) ([]warehouseutils.LoadFile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLoadFilesMetadata", ctx, options)
	ret0, _ := ret[0].([]warehouseutils.LoadFile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLoadFilesMetadata indicates an expected call of GetLoadFilesMetadata.
func (mr *MockUploaderMockRecorder) GetLoadFilesMetadata(ctx, options any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLoadFilesMetadata", reflect.TypeOf((*MockUploader)(nil).GetLoadFilesMetadata), ctx, options)
}

// GetLocalSchema mocks base method.
func (m *MockUploader) GetLocalSchema(ctx context.Context) (model.Schema, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLocalSchema", ctx)
	ret0, _ := ret[0].(model.Schema)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLocalSchema indicates an expected call of GetLocalSchema.
func (mr *MockUploaderMockRecorder) GetLocalSchema(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLocalSchema", reflect.TypeOf((*MockUploader)(nil).GetLocalSchema), ctx)
}

// GetSampleLoadFileLocation mocks base method.
func (m *MockUploader) GetSampleLoadFileLocation(ctx context.Context, tableName string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSampleLoadFileLocation", ctx, tableName)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSampleLoadFileLocation indicates an expected call of GetSampleLoadFileLocation.
func (mr *MockUploaderMockRecorder) GetSampleLoadFileLocation(ctx, tableName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSampleLoadFileLocation", reflect.TypeOf((*MockUploader)(nil).GetSampleLoadFileLocation), ctx, tableName)
}

// GetSingleLoadFile mocks base method.
func (m *MockUploader) GetSingleLoadFile(ctx context.Context, tableName string) (warehouseutils.LoadFile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSingleLoadFile", ctx, tableName)
	ret0, _ := ret[0].(warehouseutils.LoadFile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSingleLoadFile indicates an expected call of GetSingleLoadFile.
func (mr *MockUploaderMockRecorder) GetSingleLoadFile(ctx, tableName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSingleLoadFile", reflect.TypeOf((*MockUploader)(nil).GetSingleLoadFile), ctx, tableName)
}

// GetTableSchemaInUpload mocks base method.
func (m *MockUploader) GetTableSchemaInUpload(tableName string) model.TableSchema {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTableSchemaInUpload", tableName)
	ret0, _ := ret[0].(model.TableSchema)
	return ret0
}

// GetTableSchemaInUpload indicates an expected call of GetTableSchemaInUpload.
func (mr *MockUploaderMockRecorder) GetTableSchemaInUpload(tableName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTableSchemaInUpload", reflect.TypeOf((*MockUploader)(nil).GetTableSchemaInUpload), tableName)
}

// GetTableSchemaInWarehouse mocks base method.
func (m *MockUploader) GetTableSchemaInWarehouse(tableName string) model.TableSchema {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTableSchemaInWarehouse", tableName)
	ret0, _ := ret[0].(model.TableSchema)
	return ret0
}

// GetTableSchemaInWarehouse indicates an expected call of GetTableSchemaInWarehouse.
func (mr *MockUploaderMockRecorder) GetTableSchemaInWarehouse(tableName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTableSchemaInWarehouse", reflect.TypeOf((*MockUploader)(nil).GetTableSchemaInWarehouse), tableName)
}

// IsWarehouseSchemaEmpty mocks base method.
func (m *MockUploader) IsWarehouseSchemaEmpty() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsWarehouseSchemaEmpty")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsWarehouseSchemaEmpty indicates an expected call of IsWarehouseSchemaEmpty.
func (mr *MockUploaderMockRecorder) IsWarehouseSchemaEmpty() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsWarehouseSchemaEmpty", reflect.TypeOf((*MockUploader)(nil).IsWarehouseSchemaEmpty))
}

// ShouldOnDedupUseNewRecord mocks base method.
func (m *MockUploader) ShouldOnDedupUseNewRecord() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ShouldOnDedupUseNewRecord")
	ret0, _ := ret[0].(bool)
	return ret0
}

// ShouldOnDedupUseNewRecord indicates an expected call of ShouldOnDedupUseNewRecord.
func (mr *MockUploaderMockRecorder) ShouldOnDedupUseNewRecord() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShouldOnDedupUseNewRecord", reflect.TypeOf((*MockUploader)(nil).ShouldOnDedupUseNewRecord))
}

// UpdateLocalSchema mocks base method.
func (m *MockUploader) UpdateLocalSchema(ctx context.Context, schema model.Schema) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateLocalSchema", ctx, schema)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateLocalSchema indicates an expected call of UpdateLocalSchema.
func (mr *MockUploaderMockRecorder) UpdateLocalSchema(ctx, schema any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateLocalSchema", reflect.TypeOf((*MockUploader)(nil).UpdateLocalSchema), ctx, schema)
}

// UseRudderStorage mocks base method.
func (m *MockUploader) UseRudderStorage() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UseRudderStorage")
	ret0, _ := ret[0].(bool)
	return ret0
}

// UseRudderStorage indicates an expected call of UseRudderStorage.
func (mr *MockUploaderMockRecorder) UseRudderStorage() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UseRudderStorage", reflect.TypeOf((*MockUploader)(nil).UseRudderStorage))
}
