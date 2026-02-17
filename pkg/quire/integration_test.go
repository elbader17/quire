package quire

import (
	"context"
	"errors"
	"testing"
)

func TestIntegration_FullWorkflow(t *testing.T) {
	ctx := context.Background()
	mock := &MockSheetsClient{
		ReadFunc: func(ctx context.Context, range_ string) ([][]interface{}, error) {
			return [][]interface{}{
				{"ID", "Name", "Email", "Age"},
				{1.0, "Alice", "alice@example.com", 30.0},
				{2.0, "Bob", "bob@example.com", 25.0},
				{3.0, "Charlie", "charlie@example.com", 35.0},
			}, nil
		},
		AppendFunc: func(ctx context.Context, range_ string, values [][]interface{}) error {
			return nil
		},
	}

	db := &DB{client: mock}
	users := db.Table("Users")

	newUsers := []TestUser{
		{ID: 4, Name: "Diana", Email: "diana@example.com", Age: 28},
	}

	err := users.Insert(ctx, newUsers)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	if len(mock.AppendCalls) != 1 {
		t.Errorf("Expected 1 append call, got %d", len(mock.AppendCalls))
	}

	var results []TestUser
	err = users.Query().
		Where("Age", ">=", 28).
		Limit(2).
		Get(ctx, &results)

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestIntegration_EmptyTable(t *testing.T) {
	ctx := context.Background()
	mock := &MockSheetsClient{
		ReadFunc: func(ctx context.Context, range_ string) ([][]interface{}, error) {
			return [][]interface{}{
				{"ID", "Name"},
			}, nil
		},
	}

	db := &DB{client: mock}
	users := db.Table("Users")

	var results []TestUser
	err := users.Query().Get(ctx, &results)

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty table, got %d", len(results))
	}
}

func TestIntegration_NoMatchingFilters(t *testing.T) {
	ctx := context.Background()
	mock := &MockSheetsClient{
		ReadFunc: func(ctx context.Context, range_ string) ([][]interface{}, error) {
			return [][]interface{}{
				{"ID", "Name", "Age"},
				{1.0, "Alice", 30.0},
				{2.0, "Bob", 25.0},
			}, nil
		},
	}

	db := &DB{client: mock}
	users := db.Table("Users")

	var results []TestUser
	err := users.Query().
		Where("Age", ">", 100).
		Get(ctx, &results)

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestIntegration_ReadError(t *testing.T) {
	ctx := context.Background()
	mock := &MockSheetsClient{
		ReadFunc: func(ctx context.Context, range_ string) ([][]interface{}, error) {
			return nil, errors.New("network error")
		},
	}

	db := &DB{client: mock}
	users := db.Table("Users")

	var results []TestUser
	err := users.Query().Get(ctx, &results)

	if err == nil {
		t.Error("Expected error for read failure")
	}
}

func TestIntegration_InsertError(t *testing.T) {
	ctx := context.Background()
	mock := &MockSheetsClient{
		AppendFunc: func(ctx context.Context, range_ string, values [][]interface{}) error {
			return errors.New("quota exceeded")
		},
	}

	db := &DB{client: mock}
	users := db.Table("Users")

	err := users.Insert(ctx, []TestUser{{ID: 1, Name: "Test"}})

	if err == nil {
		t.Error("Expected error for insert failure")
	}
}

func TestIntegration_MultipleOperations(t *testing.T) {
	ctx := context.Background()

	readCount := 0
	mock := &MockSheetsClient{
		ReadFunc: func(ctx context.Context, range_ string) ([][]interface{}, error) {
			readCount++
			return [][]interface{}{
				{"ID", "Name"},
				{float64(readCount), "User"},
			}, nil
		},
	}

	db := &DB{client: mock}

	users1 := db.Table("Users1")
	users2 := db.Table("Users2")

	var results1, results2 []TestUser
	_ = users1.Query().Get(ctx, &results1)
	_ = users2.Query().Get(ctx, &results2)

	if readCount != 2 {
		t.Errorf("Expected 2 read calls, got %d", readCount)
	}
}

func TestIntegration_MixedTypes(t *testing.T) {
	ctx := context.Background()
	mock := &MockSheetsClient{
		ReadFunc: func(ctx context.Context, range_ string) ([][]interface{}, error) {
			return [][]interface{}{
				{"ID", "Name", "Active", "Score"},
				{1.0, "Alice", "true", "95.5"},
				{2.0, "Bob", "false", "87.3"},
			}, nil
		},
	}

	db := &DB{client: mock}
	users := db.Table("Users")

	type MixedUser struct {
		ID     int     `quire:"ID"`
		Name   string  `quire:"Name"`
		Active bool    `quire:"Active"`
		Score  float64 `quire:"Score"`
	}

	var results []MixedUser
	err := users.Query().Get(ctx, &results)

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if len(results) > 0 {
		if !results[0].Active {
			t.Error("Expected Active=true for Alice")
		}
		if results[0].Score != 95.5 {
			t.Errorf("Expected Score=95.5, got %f", results[0].Score)
		}
	}
}

func TestIntegration_FilterWithDifferentOperators(t *testing.T) {
	ctx := context.Background()
	mock := &MockSheetsClient{
		ReadFunc: func(ctx context.Context, range_ string) ([][]interface{}, error) {
			return [][]interface{}{
				{"ID", "Name", "Age"},
				{1.0, "Alice", 30.0},
				{2.0, "Bob", 25.0},
				{3.0, "Charlie", 35.0},
				{4.0, "Diana", 28.0},
			}, nil
		},
	}

	db := &DB{client: mock}
	users := db.Table("Users")

	tests := []struct {
		name          string
		column        string
		op            string
		value         interface{}
		expectedCount int
	}{
		{"equal", "Name", "=", "Alice", 1},
		{"not equal", "Name", "!=", "Alice", 3},
		{"greater than", "Age", ">", 28.0, 2},
		{"greater or equal", "Age", ">=", 30.0, 2},
		{"less than", "Age", "<", 30.0, 2},
		{"less or equal", "Age", "<=", 28.0, 2},
		{"contains", "Name", "contains", "li", 2},
		{"like", "Name", "like", "BOB", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var results []TestUser
			err := users.Query().
				Where(tt.column, tt.op, tt.value).
				Get(ctx, &results)

			if err != nil {
				t.Fatalf("Query failed: %v", err)
			}

			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results for %s %s %v, got %d",
					tt.expectedCount, tt.column, tt.op, tt.value, len(results))
			}
		})
	}
}
