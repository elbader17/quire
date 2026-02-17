package quire

import (
	"context"
	"errors"
	"testing"
)

type TestUser struct {
	ID    int    `quire:"ID"`
	Name  string `quire:"Name"`
	Email string `quire:"Email"`
	Age   int    `quire:"Age"`
}

type TestProduct struct {
	SKU   string  `quire:"SKU"`
	Name  string  `quire:"Name"`
	Price float64 `quire:"Price"`
}

func TestTable_Insert(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		records    interface{}
		mockError  error
		wantErr    bool
		expectCall bool
	}{
		{
			name: "insert single struct",
			records: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@test.com", Age: 30},
			},
			expectCall: true,
		},
		{
			name: "insert multiple structs",
			records: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@test.com", Age: 30},
				{ID: 2, Name: "Bob", Email: "bob@test.com", Age: 25},
			},
			expectCall: true,
		},
		{
			name:       "insert non-slice",
			records:    TestUser{ID: 1, Name: "Alice"},
			wantErr:    true,
			expectCall: false,
		},
		{
			name: "insert with error",
			records: []TestUser{
				{ID: 1, Name: "Alice"},
			},
			mockError:  errors.New("append failed"),
			wantErr:    true,
			expectCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockSheetsClient{
				AppendFunc: func(ctx context.Context, range_ string, values [][]interface{}) error {
					return tt.mockError
				},
			}

			db := &DB{client: mock}
			table := &Table{db: db, name: "Users"}

			err := table.Insert(ctx, tt.records)

			if tt.wantErr {
				if err == nil {
					t.Error("Insert() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Insert() unexpected error = %v", err)
				return
			}

			if tt.expectCall && len(mock.AppendCalls) != 1 {
				t.Errorf("Insert() expected 1 append call, got %d", len(mock.AppendCalls))
			}

			if tt.expectCall && len(mock.AppendCalls) > 0 {
				if mock.AppendCalls[0].Range_ != "Users!A1" {
					t.Errorf("Insert() range = %v, want Users!A1", mock.AppendCalls[0].Range_)
				}
			}
		})
	}
}

func TestTable_Query(t *testing.T) {
	db := &DB{client: &MockSheetsClient{}}
	table := &Table{db: db, name: "Users"}

	query := table.Query()

	if query == nil {
		t.Fatal("Query() returned nil")
	}

	if query.table != table {
		t.Error("Query() table reference mismatch")
	}

	if len(query.filters) != 0 {
		t.Error("Query() should start with empty filters")
	}
}

func TestQuery_Where(t *testing.T) {
	db := &DB{client: &MockSheetsClient{}}
	table := &Table{db: db, name: "Users"}
	query := table.Query()

	result := query.Where("Age", ">=", 18)

	if result != query {
		t.Error("Where() should return the same query for chaining")
	}

	if len(query.filters) != 1 {
		t.Fatalf("Where() added %d filters, want 1", len(query.filters))
	}

	filter := query.filters[0]
	if filter.Column != "Age" {
		t.Errorf("Filter column = %v, want Age", filter.Column)
	}
	if filter.Operator != ">=" {
		t.Errorf("Filter operator = %v, want >=", filter.Operator)
	}
	if filter.Value != 18 {
		t.Errorf("Filter value = %v, want 18", filter.Value)
	}
}

func TestQuery_MultipleWheres(t *testing.T) {
	db := &DB{client: &MockSheetsClient{}}
	table := &Table{db: db, name: "Users"}
	query := table.Query()

	query.Where("Age", ">=", 18).Where("Name", "=", "Alice")

	if len(query.filters) != 2 {
		t.Fatalf("Expected 2 filters, got %d", len(query.filters))
	}

	if query.filters[0].Column != "Age" {
		t.Error("First filter should be Age")
	}

	if query.filters[1].Column != "Name" {
		t.Error("Second filter should be Name")
	}
}

func TestQuery_Limit(t *testing.T) {
	db := &DB{client: &MockSheetsClient{}}
	table := &Table{db: db, name: "Users"}
	query := table.Query()

	result := query.Limit(10)

	if result != query {
		t.Error("Limit() should return the same query for chaining")
	}

	if query.limit != 10 {
		t.Errorf("Limit() = %v, want 10", query.limit)
	}
}

func TestQuery_OrderBy(t *testing.T) {
	db := &DB{client: &MockSheetsClient{}}
	table := &Table{db: db, name: "Users"}
	query := table.Query()

	result := query.OrderBy("Age", true)

	if result != query {
		t.Error("OrderBy() should return the same query for chaining")
	}

	if query.orderBy != "Age" {
		t.Errorf("OrderBy() column = %v, want Age", query.orderBy)
	}

	if !query.descending {
		t.Error("OrderBy() descending should be true")
	}
}

func TestQuery_Get(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		mockData      [][]interface{}
		mockError     error
		setupQuery    func(*Query)
		wantErr       bool
		expectedCount int
	}{
		{
			name: "empty sheet",
			mockData: [][]interface{}{
				{"ID", "Name", "Email", "Age"},
			},
			expectedCount: 0,
		},
		{
			name: "single row",
			mockData: [][]interface{}{
				{"ID", "Name", "Email", "Age"},
				{1.0, "Alice", "alice@test.com", 30.0},
			},
			expectedCount: 1,
		},
		{
			name: "multiple rows",
			mockData: [][]interface{}{
				{"ID", "Name", "Email", "Age"},
				{1.0, "Alice", "alice@test.com", 30.0},
				{2.0, "Bob", "bob@test.com", 25.0},
			},
			expectedCount: 2,
		},
		{
			name:      "read error",
			mockError: errors.New("read failed"),
			wantErr:   true,
		},
		{
			name: "with filter",
			mockData: [][]interface{}{
				{"ID", "Name", "Email", "Age"},
				{1.0, "Alice", "alice@test.com", 30.0},
				{2.0, "Bob", "bob@test.com", 25.0},
			},
			setupQuery: func(q *Query) {
				q.Where("Age", ">=", 26)
			},
			expectedCount: 1,
		},
		{
			name: "with limit",
			mockData: [][]interface{}{
				{"ID", "Name", "Email", "Age"},
				{1.0, "Alice", "alice@test.com", 30.0},
				{2.0, "Bob", "bob@test.com", 25.0},
				{3.0, "Charlie", "charlie@test.com", 35.0},
			},
			setupQuery: func(q *Query) {
				q.Limit(2)
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockSheetsClient{
				ReadFunc: func(ctx context.Context, range_ string) ([][]interface{}, error) {
					return tt.mockData, tt.mockError
				},
			}

			db := &DB{client: mock}
			table := &Table{db: db, name: "Users"}
			query := table.Query()

			if tt.setupQuery != nil {
				tt.setupQuery(query)
			}

			var results []TestUser
			err := query.Get(ctx, &results)

			if tt.wantErr {
				if err == nil {
					t.Error("Get() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Get() unexpected error = %v", err)
				return
			}

			if len(results) != tt.expectedCount {
				t.Errorf("Get() returned %d results, want %d", len(results), tt.expectedCount)
			}
		})
	}
}

func TestQuery_Get_InvalidDest(t *testing.T) {
	mock := &MockSheetsClient{
		ReadFunc: func(ctx context.Context, range_ string) ([][]interface{}, error) {
			return [][]interface{}{
				{"ID", "Name"},
				{1.0, "Alice"},
			}, nil
		},
	}

	db := &DB{client: mock}
	table := &Table{db: db, name: "Users"}
	query := table.Query()

	var notASlice int
	err := query.Get(context.Background(), &notASlice)
	if err == nil {
		t.Error("Get() expected error for non-slice destination")
	}
}
