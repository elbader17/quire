package quire

import (
	"testing"
)

func TestMatchesOperator(t *testing.T) {
	tests := []struct {
		name     string
		cell     interface{}
		op       string
		value    interface{}
		expected bool
	}{
		{"equal strings", "hello", "=", "hello", true},
		{"not equal strings", "hello", "=", "world", false},
		{"equal double eq", "test", "==", "test", true},
		{"not equal", "hello", "!=", "world", true},
		{"not equal same", "hello", "!=", "hello", false},
		{"greater than true", 10.0, ">", 5.0, true},
		{"greater than false", 5.0, ">", 10.0, false},
		{"greater or equal true", 10.0, ">=", 10.0, true},
		{"greater or equal false", 5.0, ">=", 10.0, false},
		{"less than true", 5.0, "<", 10.0, true},
		{"less than false", 10.0, "<", 5.0, false},
		{"less or equal true", 10.0, "<=", 10.0, true},
		{"less or equal false", 10.0, "<=", 5.0, false},
		{"string greater", "20", ">", "10", true},
		{"string less", "10", "<", "20", true},
		{"contains true", "Hello World", "contains", "world", true},
		{"contains false", "Hello World", "contains", "foo", false},
		{"like true", "Hello World", "like", "hello", true},
		{"like case insensitive", "HELLO", "like", "hello", true},
		{"unknown operator", "test", "unknown", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesOperator(tt.cell, tt.op, tt.value)
			if result != tt.expected {
				t.Errorf("matchesOperator(%v, %s, %v) = %v, want %v",
					tt.cell, tt.op, tt.value, result, tt.expected)
			}
		})
	}
}

func TestCompareValues(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected int
	}{
		{"equal numbers", 10.0, 10.0, 0},
		{"a greater than b", 20.0, 10.0, 1},
		{"a less than b", 5.0, 10.0, -1},
		{"equal strings", "abc", "abc", 0},
		{"string a greater", "xyz", "abc", 1},
		{"string a less", "abc", "xyz", -1},
		{"numeric strings", "20", "10", 1},
		{"string vs number", "abc", 123.0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareValues(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("compareValues(%v, %v) = %d, want %d",
					tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestQuery_MatchesFilters(t *testing.T) {
	tests := []struct {
		name     string
		row      []interface{}
		headers  []interface{}
		filters  []Filter
		expected bool
	}{
		{
			name:     "no filters",
			row:      []interface{}{1.0, "Alice", 30.0},
			headers:  []interface{}{"ID", "Name", "Age"},
			filters:  []Filter{},
			expected: true,
		},
		{
			name:    "single filter match",
			row:     []interface{}{1.0, "Alice", 30.0},
			headers: []interface{}{"ID", "Name", "Age"},
			filters: []Filter{
				{Column: "Name", Operator: "=", Value: "Alice"},
			},
			expected: true,
		},
		{
			name:    "single filter no match",
			row:     []interface{}{1.0, "Alice", 30.0},
			headers: []interface{}{"ID", "Name", "Age"},
			filters: []Filter{
				{Column: "Name", Operator: "=", Value: "Bob"},
			},
			expected: false,
		},
		{
			name:    "multiple filters all match",
			row:     []interface{}{1.0, "Alice", 30.0},
			headers: []interface{}{"ID", "Name", "Age"},
			filters: []Filter{
				{Column: "Name", Operator: "=", Value: "Alice"},
				{Column: "Age", Operator: ">=", Value: 25.0},
			},
			expected: true,
		},
		{
			name:    "multiple filters one fails",
			row:     []interface{}{1.0, "Alice", 20.0},
			headers: []interface{}{"ID", "Name", "Age"},
			filters: []Filter{
				{Column: "Name", Operator: "=", Value: "Alice"},
				{Column: "Age", Operator: ">=", Value: 25.0},
			},
			expected: false,
		},
		{
			name:    "column not found",
			row:     []interface{}{1.0, "Alice"},
			headers: []interface{}{"ID", "Name"},
			filters: []Filter{
				{Column: "NonExistent", Operator: "=", Value: "test"},
			},
			expected: false,
		},
		{
			name:    "column index out of range",
			row:     []interface{}{1.0},
			headers: []interface{}{"ID", "Name", "Age"},
			filters: []Filter{
				{Column: "Age", Operator: "=", Value: 30.0},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Query{filters: tt.filters}
			result := q.matchesFilters(tt.row, tt.headers)
			if result != tt.expected {
				t.Errorf("matchesFilters() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestQuery_ApplyFilters(t *testing.T) {
	q := &Query{
		filters: []Filter{
			{Column: "Age", Operator: ">=", Value: 25.0},
		},
	}

	rows := [][]interface{}{
		{1.0, "Alice", 30.0},
		{2.0, "Bob", 20.0},
		{3.0, "Charlie", 35.0},
	}
	headers := []interface{}{"ID", "Name", "Age"}

	result := q.applyFilters(rows, headers)

	if len(result) != 2 {
		t.Errorf("applyFilters() returned %d rows, want 2", len(result))
	}

	if len(result) > 0 && result[0][1] != "Alice" {
		t.Errorf("First result should be Alice, got %v", result[0][1])
	}

	if len(result) > 1 && result[1][1] != "Charlie" {
		t.Errorf("Second result should be Charlie, got %v", result[1][1])
	}
}

func TestQuery_ApplyFilters_NoFilters(t *testing.T) {
	q := &Query{filters: []Filter{}}

	rows := [][]interface{}{
		{1.0, "Alice"},
		{2.0, "Bob"},
	}
	headers := []interface{}{"ID", "Name"}

	result := q.applyFilters(rows, headers)

	if len(result) != 2 {
		t.Errorf("applyFilters() with no filters should return all rows, got %d", len(result))
	}
}

func TestQuery_ApplyLimit(t *testing.T) {
	q := &Query{limit: 2}

	rows := [][]interface{}{
		{1.0},
		{2.0},
		{3.0},
		{4.0},
	}

	result := q.applyLimit(rows)

	if len(result) != 2 {
		t.Errorf("applyLimit(2) returned %d rows, want 2", len(result))
	}
}

func TestQuery_ApplyLimit_Zero(t *testing.T) {
	q := &Query{limit: 0}

	rows := [][]interface{}{
		{1.0},
		{2.0},
	}

	result := q.applyLimit(rows)

	if len(result) != 2 {
		t.Errorf("applyLimit(0) should return all rows, got %d", len(result))
	}
}

func TestQuery_ApplyLimit_GreaterThanLength(t *testing.T) {
	q := &Query{limit: 10}

	rows := [][]interface{}{
		{1.0},
		{2.0},
	}

	result := q.applyLimit(rows)

	if len(result) != 2 {
		t.Errorf("applyLimit(10) with 2 rows should return 2 rows, got %d", len(result))
	}
}

func TestQuery_Chaining(t *testing.T) {
	db := &DB{client: &MockSheetsClient{}}
	table := &Table{db: db, name: "Users"}

	query := table.Query().
		Where("Age", ">=", 18).
		Where("Status", "=", "active").
		Limit(10).
		OrderBy("Name", false)

	if len(query.filters) != 2 {
		t.Errorf("Chained Where() calls should add 2 filters, got %d", len(query.filters))
	}

	if query.limit != 10 {
		t.Errorf("Chained Limit() should set limit to 10, got %d", query.limit)
	}

	if query.orderBy != "Name" {
		t.Errorf("Chained OrderBy() should set orderBy to Name, got %s", query.orderBy)
	}
}
