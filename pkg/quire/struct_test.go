package quire

import (
	"reflect"
	"testing"
)

func TestStructSliceToValues(t *testing.T) {
	tests := []struct {
		name        string
		records     interface{}
		wantErr     bool
		expectedErr string
	}{
		{
			name: "valid struct slice",
			records: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@test.com", Age: 30},
			},
			wantErr: false,
		},
		{
			name:        "non-slice",
			records:     TestUser{ID: 1, Name: "Alice"},
			wantErr:     true,
			expectedErr: "records must be a slice",
		},
		{
			name:    "empty slice",
			records: []TestUser{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := structSliceToValues(tt.records)

			if tt.wantErr {
				if err == nil {
					t.Error("structSliceToValues() expected error but got nil")
					return
				}
				if tt.expectedErr != "" && err.Error() != tt.expectedErr {
					t.Errorf("structSliceToValues() error = %v, want %v", err.Error(), tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("structSliceToValues() unexpected error = %v", err)
				return
			}

			if !tt.wantErr {
				t.Logf("structSliceToValues() returned %d rows", len(values))
			}
		})
	}
}

func TestStructToValues(t *testing.T) {
	tests := []struct {
		name        string
		record      interface{}
		wantErr     bool
		expectedErr string
		expectedLen int
	}{
		{
			name:        "valid struct",
			record:      TestUser{ID: 1, Name: "Alice", Email: "alice@test.com", Age: 30},
			wantErr:     false,
			expectedLen: 4,
		},
		{
			name:        "pointer to struct",
			record:      &TestUser{ID: 1, Name: "Alice"},
			wantErr:     false,
			expectedLen: 4,
		},
		{
			name:        "non-struct",
			record:      "not a struct",
			wantErr:     true,
			expectedErr: "record must be a struct",
		},
		{
			name: "struct with ignored field",
			record: struct {
				ID   int    `quire:"ID"`
				Name string `quire:"-"`
			}{ID: 1, Name: "Alice"},
			wantErr:     false,
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := structToValues(tt.record)

			if tt.wantErr {
				if err == nil {
					t.Error("structToValues() expected error but got nil")
					return
				}
				if tt.expectedErr != "" && err.Error() != tt.expectedErr {
					t.Errorf("structToValues() error = %v, want %v", err.Error(), tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("structToValues() unexpected error = %v", err)
				return
			}

			if len(values) != tt.expectedLen {
				t.Errorf("structToValues() returned %d values, want %d", len(values), tt.expectedLen)
			}
		})
	}
}

func TestScanIntoSlice(t *testing.T) {
	tests := []struct {
		name     string
		rows     [][]interface{}
		headers  []interface{}
		dest     interface{}
		wantErr  bool
		validate func(t *testing.T, dest interface{})
	}{
		{
			name: "scan into user slice",
			rows: [][]interface{}{
				{1.0, "Alice", "alice@test.com", 30.0},
			},
			headers: []interface{}{"ID", "Name", "Email", "Age"},
			dest:    &[]TestUser{},
			wantErr: false,
			validate: func(t *testing.T, dest interface{}) {
				users := dest.(*[]TestUser)
				if len(*users) != 1 {
					t.Errorf("Expected 1 user, got %d", len(*users))
					return
				}
				if (*users)[0].Name != "Alice" {
					t.Errorf("Expected Name=Alice, got %s", (*users)[0].Name)
				}
			},
		},
		{
			name:    "non-pointer dest",
			rows:    [][]interface{}{},
			headers: []interface{}{"ID"},
			dest:    []TestUser{},
			wantErr: true,
		},
		{
			name:    "pointer to non-slice",
			rows:    [][]interface{}{},
			headers: []interface{}{"ID"},
			dest:    new(int),
			wantErr: true,
		},
		{
			name:    "empty rows",
			rows:    [][]interface{}{},
			headers: []interface{}{"ID", "Name"},
			dest:    &[]TestUser{},
			wantErr: false,
			validate: func(t *testing.T, dest interface{}) {
				users := dest.(*[]TestUser)
				if len(*users) != 0 {
					t.Errorf("Expected 0 users, got %d", len(*users))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := scanIntoSlice(tt.rows, tt.headers, tt.dest)

			if tt.wantErr {
				if err == nil {
					t.Error("scanIntoSlice() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("scanIntoSlice() unexpected error = %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, tt.dest)
			}
		})
	}
}

func TestScanRow(t *testing.T) {
	tests := []struct {
		name     string
		row      []interface{}
		headers  []interface{}
		dest     interface{}
		wantErr  bool
		validate func(t *testing.T, dest interface{})
	}{
		{
			name:    "scan valid row",
			row:     []interface{}{1.0, "Alice", "alice@test.com", 30.0},
			headers: []interface{}{"ID", "Name", "Email", "Age"},
			dest:    &TestUser{},
			wantErr: false,
			validate: func(t *testing.T, dest interface{}) {
				user := dest.(*TestUser)
				if user.ID != 1 {
					t.Errorf("Expected ID=1, got %d", user.ID)
				}
				if user.Name != "Alice" {
					t.Errorf("Expected Name=Alice, got %s", user.Name)
				}
				if user.Age != 30 {
					t.Errorf("Expected Age=30, got %d", user.Age)
				}
			},
		},
		{
			name:    "scan with missing columns",
			row:     []interface{}{1.0, "Alice"},
			headers: []interface{}{"ID", "Name", "Email", "Age"},
			dest:    &TestUser{},
			wantErr: false,
			validate: func(t *testing.T, dest interface{}) {
				user := dest.(*TestUser)
				if user.ID != 1 {
					t.Errorf("Expected ID=1, got %d", user.ID)
				}
			},
		},
		{
			name:    "non-struct dest",
			row:     []interface{}{1.0},
			headers: []interface{}{"ID"},
			dest:    new(int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destVal := reflect.ValueOf(tt.dest)
			err := scanRow(tt.row, tt.headers, destVal.Elem())

			if tt.wantErr {
				if err == nil {
					t.Error("scanRow() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("scanRow() unexpected error = %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, tt.dest)
			}
		})
	}
}

func TestSetField(t *testing.T) {
	tests := []struct {
		name      string
		field     reflect.Value
		value     interface{}
		expected  interface{}
		expectSet bool
	}{
		{
			name:      "set string field",
			field:     reflect.ValueOf(new(string)).Elem(),
			value:     "test",
			expected:  "test",
			expectSet: true,
		},
		{
			name:      "set int field from float",
			field:     reflect.ValueOf(new(int)).Elem(),
			value:     42.0,
			expected:  42,
			expectSet: true,
		},
		{
			name:      "set int field from string",
			field:     reflect.ValueOf(new(int)).Elem(),
			value:     "42",
			expected:  42,
			expectSet: true,
		},
		{
			name:      "set float field",
			field:     reflect.ValueOf(new(float64)).Elem(),
			value:     "3.14",
			expected:  3.14,
			expectSet: true,
		},
		{
			name:      "set bool field",
			field:     reflect.ValueOf(new(bool)).Elem(),
			value:     "true",
			expected:  true,
			expectSet: true,
		},
		{
			name:      "set uint field",
			field:     reflect.ValueOf(new(uint)).Elem(),
			value:     "100",
			expected:  uint(100),
			expectSet: true,
		},
		{
			name:      "invalid int value",
			field:     reflect.ValueOf(new(int)).Elem(),
			value:     "not-a-number",
			expected:  0,
			expectSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := tt.field.Interface()
			err := setField(tt.field, tt.value)

			if err != nil {
				t.Errorf("setField() unexpected error = %v", err)
				return
			}

			if tt.expectSet {
				actual := tt.field.Interface()
				if actual != tt.expected {
					t.Errorf("setField() = %v, want %v", actual, tt.expected)
				}
			} else {
				if tt.field.Interface() != original {
					t.Error("setField() should not have changed the field")
				}
			}
		})
	}
}

func TestSetField_CannotSet(t *testing.T) {
	type TestStruct struct {
		unexported string
	}

	s := TestStruct{}
	field := reflect.ValueOf(s).FieldByName("unexported")

	err := setField(field, "value")
	if err != nil {
		t.Errorf("setField() unexpected error = %v", err)
	}
}
