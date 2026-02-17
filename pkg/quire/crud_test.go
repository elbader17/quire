package quire

import (
	"context"
	"errors"
	"testing"
)

func TestTable_Update(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		rowIndex int
		record   interface{}
		wantErr  bool
		expected string
	}{
		{
			name:     "update valid row",
			rowIndex: 0,
			record:   TestUser{ID: 1, Name: "Updated Alice", Email: "updated@example.com", Age: 31},
			wantErr:  false,
		},
		{
			name:     "update negative index",
			rowIndex: -1,
			record:   TestUser{ID: 1, Name: "Test"},
			wantErr:  true,
		},
		{
			name:     "update non-struct",
			rowIndex: 0,
			record:   "not a struct",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockSheetsClient{
				WriteFunc: func(ctx context.Context, range_ string, values [][]interface{}) error {
					return nil
				},
			}

			db := &DB{client: mock}
			table := &Table{db: db, name: "Users"}

			err := table.Update(ctx, tt.rowIndex, tt.record)

			if tt.wantErr {
				if err == nil {
					t.Error("Update() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Update() unexpected error = %v", err)
				return
			}

			if len(mock.WriteCalls) != 1 {
				t.Errorf("Update() expected 1 write call, got %d", len(mock.WriteCalls))
			}
		})
	}
}

func TestTable_UpdateWhere(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		mockData     [][]interface{}
		column       string
		operator     string
		value        interface{}
		record       interface{}
		expectedRows int
		wantErr      bool
	}{
		{
			name: "update matching rows",
			mockData: [][]interface{}{
				{"ID", "Name", "Status"},
				{1.0, "Alice", "pending"},
				{2.0, "Bob", "active"},
				{3.0, "Charlie", "pending"},
			},
			column:       "Status",
			operator:     "=",
			value:        "pending",
			record:       TestUser{ID: 99, Name: "Updated", Email: "test@test.com", Age: 25},
			expectedRows: 2,
			wantErr:      false,
		},
		{
			name: "update no matching rows",
			mockData: [][]interface{}{
				{"ID", "Name", "Status"},
				{1.0, "Alice", "active"},
				{2.0, "Bob", "active"},
			},
			column:       "Status",
			operator:     "=",
			value:        "deleted",
			record:       TestUser{ID: 99, Name: "Updated"},
			expectedRows: 0,
			wantErr:      false,
		},
		{
			name:     "read error",
			mockData: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeCount := 0
			mock := &MockSheetsClient{
				ReadFunc: func(ctx context.Context, range_ string) ([][]interface{}, error) {
					if tt.mockData == nil {
						return nil, errors.New("read error")
					}
					return tt.mockData, nil
				},
				WriteFunc: func(ctx context.Context, range_ string, values [][]interface{}) error {
					writeCount++
					return nil
				},
			}

			db := &DB{client: mock}
			table := &Table{db: db, name: "Users"}

			err := table.UpdateWhere(ctx, tt.column, tt.operator, tt.value, tt.record)

			if tt.wantErr {
				if err == nil {
					t.Error("UpdateWhere() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateWhere() unexpected error = %v", err)
				return
			}

			if writeCount != tt.expectedRows {
				t.Errorf("UpdateWhere() expected %d write calls, got %d", tt.expectedRows, writeCount)
			}
		})
	}
}

func TestTable_Delete(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		rowIndex int
		wantErr  bool
	}{
		{
			name:     "delete valid row",
			rowIndex: 0,
			wantErr:  false,
		},
		{
			name:     "delete negative index",
			rowIndex: -1,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockSheetsClient{
				DeleteRowsFunc: func(ctx context.Context, sheetName string, rowIndices []int) error {
					return nil
				},
			}

			db := &DB{client: mock}
			table := &Table{db: db, name: "Users"}

			err := table.Delete(ctx, tt.rowIndex)

			if tt.wantErr {
				if err == nil {
					t.Error("Delete() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Delete() unexpected error = %v", err)
				return
			}

			if len(mock.DeleteRowsCalls) != 1 {
				t.Errorf("Delete() expected 1 delete call, got %d", len(mock.DeleteRowsCalls))
			}

			if len(mock.DeleteRowsCalls) > 0 {
				if len(mock.DeleteRowsCalls[0].RowIndices) != 1 || mock.DeleteRowsCalls[0].RowIndices[0] != 1 {
					t.Errorf("Delete() expected row index 1 (0+1), got %v", mock.DeleteRowsCalls[0].RowIndices)
				}
			}
		})
	}
}

func TestTable_DeleteWhere(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		mockData     [][]interface{}
		column       string
		operator     string
		value        interface{}
		expectedRows int
		wantErr      bool
	}{
		{
			name: "delete matching rows",
			mockData: [][]interface{}{
				{"ID", "Name", "Status"},
				{1.0, "Alice", "deleted"},
				{2.0, "Bob", "active"},
				{3.0, "Charlie", "deleted"},
			},
			column:       "Status",
			operator:     "=",
			value:        "deleted",
			expectedRows: 2,
			wantErr:      false,
		},
		{
			name: "delete no matching rows",
			mockData: [][]interface{}{
				{"ID", "Name", "Status"},
				{1.0, "Alice", "active"},
				{2.0, "Bob", "active"},
			},
			column:       "Status",
			operator:     "=",
			value:        "deleted",
			expectedRows: 0,
			wantErr:      false,
		},
		{
			name:     "read error",
			mockData: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockSheetsClient{
				ReadFunc: func(ctx context.Context, range_ string) ([][]interface{}, error) {
					if tt.mockData == nil {
						return nil, errors.New("read error")
					}
					return tt.mockData, nil
				},
				DeleteRowsFunc: func(ctx context.Context, sheetName string, rowIndices []int) error {
					return nil
				},
			}

			db := &DB{client: mock}
			table := &Table{db: db, name: "Users"}

			err := table.DeleteWhere(ctx, tt.column, tt.operator, tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("DeleteWhere() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("DeleteWhere() unexpected error = %v", err)
				return
			}

			expectedCalls := 0
			if tt.expectedRows > 0 {
				expectedCalls = 1
			}
			if len(mock.DeleteRowsCalls) != expectedCalls {
				t.Errorf("DeleteWhere() expected %d delete call(s), got %d", expectedCalls, len(mock.DeleteRowsCalls))
				return
			}

			if tt.expectedRows > 0 && len(mock.DeleteRowsCalls) > 0 {
				deletedCount := len(mock.DeleteRowsCalls[0].RowIndices)
				if deletedCount != tt.expectedRows {
					t.Errorf("DeleteWhere() expected %d rows deleted, got %d", tt.expectedRows, deletedCount)
				}
			}
		})
	}
}

func TestColumnIndexToLetter(t *testing.T) {
	tests := []struct {
		index    int
		expected string
	}{
		{0, "A"},
		{1, "B"},
		{25, "Z"},
		{26, "AA"},
		{27, "AB"},
		{51, "AZ"},
		{52, "BA"},
		{701, "ZZ"},
		{702, "AAA"},
		{-1, "A"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := columnIndexToLetter(tt.index)
			if result != tt.expected {
				t.Errorf("columnIndexToLetter(%d) = %s, want %s", tt.index, result, tt.expected)
			}
		})
	}
}

func TestMatchesFilter(t *testing.T) {
	headers := []interface{}{"ID", "Name", "Status"}
	row := []interface{}{1.0, "Alice", "active"}

	tests := []struct {
		name     string
		filter   Filter
		expected bool
	}{
		{
			name:     "match equal",
			filter:   Filter{Column: "Name", Operator: "=", Value: "Alice"},
			expected: true,
		},
		{
			name:     "no match equal",
			filter:   Filter{Column: "Name", Operator: "=", Value: "Bob"},
			expected: false,
		},
		{
			name:     "match contains",
			filter:   Filter{Column: "Name", Operator: "contains", Value: "lic"},
			expected: true,
		},
		{
			name:     "column not found",
			filter:   Filter{Column: "NonExistent", Operator: "=", Value: "test"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesFilter(row, headers, tt.filter)
			if result != tt.expected {
				t.Errorf("matchesFilter() = %v, want %v", result, tt.expected)
			}
		})
	}
}
