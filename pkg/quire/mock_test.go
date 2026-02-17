package quire

import (
	"context"
	"fmt"
)

type MockSheetsClient struct {
	ReadFunc       func(ctx context.Context, range_ string) ([][]interface{}, error)
	WriteFunc      func(ctx context.Context, range_ string, values [][]interface{}) error
	AppendFunc     func(ctx context.Context, range_ string, values [][]interface{}) error
	ClearFunc      func(ctx context.Context, range_ string) error
	DeleteRowsFunc func(ctx context.Context, sheetName string, rowIndices []int) error

	ReadCalls       []MockCall
	WriteCalls      []MockCall
	AppendCalls     []MockCall
	ClearCalls      []MockCall
	DeleteRowsCalls []DeleteRowsCall
}

type DeleteRowsCall struct {
	SheetName  string
	RowIndices []int
}

type MockCall struct {
	Range_ string
	Values [][]interface{}
}

func (m *MockSheetsClient) Read(ctx context.Context, range_ string) ([][]interface{}, error) {
	m.ReadCalls = append(m.ReadCalls, MockCall{Range_: range_})
	if m.ReadFunc != nil {
		return m.ReadFunc(ctx, range_)
	}
	return nil, fmt.Errorf("Read not implemented")
}

func (m *MockSheetsClient) Write(ctx context.Context, range_ string, values [][]interface{}) error {
	m.WriteCalls = append(m.WriteCalls, MockCall{Range_: range_, Values: values})
	if m.WriteFunc != nil {
		return m.WriteFunc(ctx, range_, values)
	}
	return fmt.Errorf("Write not implemented")
}

func (m *MockSheetsClient) Append(ctx context.Context, range_ string, values [][]interface{}) error {
	m.AppendCalls = append(m.AppendCalls, MockCall{Range_: range_, Values: values})
	if m.AppendFunc != nil {
		return m.AppendFunc(ctx, range_, values)
	}
	return fmt.Errorf("Append not implemented")
}

func (m *MockSheetsClient) Clear(ctx context.Context, range_ string) error {
	m.ClearCalls = append(m.ClearCalls, MockCall{Range_: range_})
	if m.ClearFunc != nil {
		return m.ClearFunc(ctx, range_)
	}
	return fmt.Errorf("Clear not implemented")
}

func (m *MockSheetsClient) DeleteRows(ctx context.Context, sheetName string, rowIndices []int) error {
	m.DeleteRowsCalls = append(m.DeleteRowsCalls, DeleteRowsCall{SheetName: sheetName, RowIndices: rowIndices})
	if m.DeleteRowsFunc != nil {
		return m.DeleteRowsFunc(ctx, sheetName, rowIndices)
	}
	return nil
}

func (m *MockSheetsClient) Reset() {
	m.ReadCalls = nil
	m.WriteCalls = nil
	m.AppendCalls = nil
	m.ClearCalls = nil
	m.DeleteRowsCalls = nil
}
