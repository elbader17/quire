package quire

import (
	"context"
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		cfg           Config
		wantErr       bool
		expectedError string
	}{
		{
			name: "missing spreadsheet id",
			cfg: Config{
				SpreadsheetID: "",
				Credentials:   []byte(`{"type":"service_account"}`),
			},
			wantErr:       true,
			expectedError: "spreadsheet ID is required",
		},
		{
			name: "missing credentials",
			cfg: Config{
				SpreadsheetID: "test-id",
				Credentials:   nil,
			},
			wantErr:       true,
			expectedError: "credentials are required",
		},
		{
			name: "empty credentials",
			cfg: Config{
				SpreadsheetID: "test-id",
				Credentials:   []byte{},
			},
			wantErr:       true,
			expectedError: "credentials are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.cfg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("New() expected error but got nil")
					return
				}
				if tt.expectedError != "" && err.Error() != tt.expectedError {
					t.Errorf("New() error = %v, want %v", err.Error(), tt.expectedError)
				}
				return
			}

			if err != nil {
				t.Errorf("New() unexpected error = %v", err)
				return
			}

			if db == nil {
				t.Error("New() returned nil db")
				return
			}

			if db.spreadsheetID != tt.cfg.SpreadsheetID {
				t.Errorf("New() spreadsheetID = %v, want %v", db.spreadsheetID, tt.cfg.SpreadsheetID)
			}

			err = db.Close()
			if err != nil {
				t.Errorf("Close() error = %v", err)
			}
		})
	}
}

func TestDB_Table(t *testing.T) {
	mockClient := &MockSheetsClient{}
	db := &DB{
		spreadsheetID: "test-id",
		client:        mockClient,
	}

	table := db.Table("Users")

	if table == nil {
		t.Fatal("Table() returned nil")
	}

	if table.name != "Users" {
		t.Errorf("Table() name = %v, want %v", table.name, "Users")
	}

	if table.db != db {
		t.Error("Table() db reference mismatch")
	}
}

func TestDB_Close(t *testing.T) {
	db := &DB{
		spreadsheetID: "test-id",
		client:        &MockSheetsClient{},
	}

	err := db.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestNew_WithInvalidCredentials(t *testing.T) {
	cfg := Config{
		SpreadsheetID: "test-id",
		Credentials:   []byte(`invalid json`),
	}

	_, err := New(cfg)
	if err == nil {
		t.Error("New() expected error for invalid credentials but got nil")
	}
}

func TestSheetsClientInterface(t *testing.T) {
	var _ SheetsClient = (*MockSheetsClient)(nil)
}

func TestMockSheetsClient_Methods(t *testing.T) {
	ctx := context.Background()
	mock := &MockSheetsClient{}

	t.Run("Read tracking", func(t *testing.T) {
		mock.Reset()
		mock.ReadFunc = func(ctx context.Context, range_ string) ([][]interface{}, error) {
			return [][]interface{}{{"test"}}, nil
		}

		_, _ = mock.Read(ctx, "Sheet1!A1")

		if len(mock.ReadCalls) != 1 {
			t.Errorf("Read calls = %d, want 1", len(mock.ReadCalls))
		}

		if mock.ReadCalls[0].Range_ != "Sheet1!A1" {
			t.Errorf("Read call range = %v, want Sheet1!A1", mock.ReadCalls[0].Range_)
		}
	})

	t.Run("Append tracking", func(t *testing.T) {
		mock.Reset()
		values := [][]interface{}{{"data"}}
		_ = mock.Append(ctx, "Sheet1!A1", values)

		if len(mock.AppendCalls) != 1 {
			t.Errorf("Append calls = %d, want 1", len(mock.AppendCalls))
		}

		if len(mock.AppendCalls[0].Values) != 1 {
			t.Errorf("Append values length = %d, want 1", len(mock.AppendCalls[0].Values))
		}
	})

	t.Run("default error", func(t *testing.T) {
		mock.Reset()
		mock.ReadFunc = nil

		_, err := mock.Read(ctx, "test")
		if err == nil {
			t.Error("Expected error when ReadFunc is nil")
		}
	})
}

func TestMockSheetsClient_Reset(t *testing.T) {
	ctx := context.Background()
	mock := &MockSheetsClient{
		ReadFunc: func(ctx context.Context, range_ string) ([][]interface{}, error) {
			return nil, nil
		},
	}

	_, _ = mock.Read(ctx, "test")
	_ = mock.Append(ctx, "test", nil)
	_ = mock.Write(ctx, "test", nil)
	_ = mock.Clear(ctx, "test")

	mock.Reset()

	if len(mock.ReadCalls) != 0 {
		t.Error("Reset did not clear ReadCalls")
	}
	if len(mock.AppendCalls) != 0 {
		t.Error("Reset did not clear AppendCalls")
	}
	if len(mock.WriteCalls) != 0 {
		t.Error("Reset did not clear WriteCalls")
	}
	if len(mock.ClearCalls) != 0 {
		t.Error("Reset did not clear ClearCalls")
	}
}

func TestErrorWrapping(t *testing.T) {
	mock := &MockSheetsClient{
		ReadFunc: func(ctx context.Context, range_ string) ([][]interface{}, error) {
			return nil, errors.New("network error")
		},
	}

	_, err := mock.Read(context.Background(), "Sheet1!A1")
	if err == nil {
		t.Fatal("Expected error")
	}

	if err.Error() != "network error" {
		t.Errorf("Error message = %v, want 'network error'", err.Error())
	}
}
