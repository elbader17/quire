package quire

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Table represents a sheet (table) within the spreadsheet.
type Table struct {
	db   *DB
	name string
}

// Query builds a query for the table.
func (t *Table) Query() *Query {
	return &Query{
		table: t,
	}
}

// Insert adds new rows to the table.
func (t *Table) Insert(ctx context.Context, records interface{}) error {
	values, err := structSliceToValues(records)
	if err != nil {
		return fmt.Errorf("failed to convert records: %w", err)
	}

	range_ := t.name + "!A1"
	return t.db.client.Append(ctx, range_, values)
}

// Update modifies a specific row by its index (0-based, excluding header).
func (t *Table) Update(ctx context.Context, rowIndex int, record interface{}) error {
	if rowIndex < 0 {
		return fmt.Errorf("row index cannot be negative")
	}

	values, err := structToValues(record)
	if err != nil {
		return fmt.Errorf("failed to convert record: %w", err)
	}

	actualRow := rowIndex + 2
	colCount := len(values)
	endCol := columnIndexToLetter(colCount - 1)
	range_ := fmt.Sprintf("%s!A%d:%s%d", t.name, actualRow, endCol, actualRow)

	return t.db.client.Write(ctx, range_, [][]interface{}{values})
}

// UpdateWhere updates all rows matching the filter condition.
func (t *Table) UpdateWhere(ctx context.Context, column, operator string, value interface{}, record interface{}) error {
	data, err := t.db.client.Read(ctx, t.name)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	if len(data) < 2 {
		return nil
	}

	headers := data[0]
	rows := data[1:]

	filter := Filter{Column: column, Operator: operator, Value: value}
	indices := []int{}
	for i, row := range rows {
		if matchesFilter(row, headers, filter) {
			indices = append(indices, i)
		}
	}

	if len(indices) == 0 {
		return nil
	}

	values, err := structToValues(record)
	if err != nil {
		return fmt.Errorf("failed to convert record: %w", err)
	}

	colCount := len(values)
	endCol := columnIndexToLetter(colCount - 1)

	for _, idx := range indices {
		actualRow := idx + 2
		range_ := fmt.Sprintf("%s!A%d:%s%d", t.name, actualRow, endCol, actualRow)
		if err := t.db.client.Write(ctx, range_, [][]interface{}{values}); err != nil {
			return fmt.Errorf("failed to update row %d: %w", idx, err)
		}
	}

	return nil
}

// Delete removes a specific row by its index (0-based, excluding header).
func (t *Table) Delete(ctx context.Context, rowIndex int) error {
	if rowIndex < 0 {
		return fmt.Errorf("row index cannot be negative")
	}

	actualRow := rowIndex + 1
	return t.db.client.DeleteRows(ctx, t.name, []int{actualRow})
}

// DeleteWhere removes all rows matching the filter condition.
func (t *Table) DeleteWhere(ctx context.Context, column, operator string, value interface{}) error {
	data, err := t.db.client.Read(ctx, t.name)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	if len(data) < 2 {
		return nil
	}

	headers := data[0]
	rows := data[1:]

	filter := Filter{Column: column, Operator: operator, Value: value}
	indices := []int{}
	for i, row := range rows {
		if matchesFilter(row, headers, filter) {
			indices = append(indices, i+1)
		}
	}

	if len(indices) == 0 {
		return nil
	}

	sort.Sort(sort.Reverse(sort.IntSlice(indices)))

	return t.db.client.DeleteRows(ctx, t.name, indices)
}

func matchesFilter(row []interface{}, headers []interface{}, filter Filter) bool {
	colIdx := -1
	for i, h := range headers {
		if h == filter.Column {
			colIdx = i
			break
		}
	}
	if colIdx == -1 || colIdx >= len(row) {
		return false
	}

	return matchesOperator(row[colIdx], filter.Operator, filter.Value)
}

func columnIndexToLetter(index int) string {
	if index < 0 {
		return "A"
	}
	result := ""
	for index >= 0 {
		result = string(rune('A'+index%26)) + result
		index = index/26 - 1
	}
	return result
}

// Query provides a fluent interface for building queries.
type Query struct {
	table      *Table
	filters    []Filter
	limit      int
	orderBy    string
	descending bool
}

// Filter represents a WHERE condition.
type Filter struct {
	Column   string
	Operator string
	Value    interface{}
}

// Where adds a filter condition.
func (q *Query) Where(column, operator string, value interface{}) *Query {
	q.filters = append(q.filters, Filter{
		Column:   column,
		Operator: operator,
		Value:    value,
	})
	return q
}

// Limit sets the maximum number of results.
func (q *Query) Limit(n int) *Query {
	q.limit = n
	return q
}

// OrderBy sets the sort column and direction.
func (q *Query) OrderBy(column string, descending bool) *Query {
	q.orderBy = column
	q.descending = descending
	return q
}

// Get executes the query and scans results into the provided slice.
func (q *Query) Get(ctx context.Context, dest interface{}) error {
	range_ := q.table.name
	data, err := q.table.db.client.Read(ctx, range_)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	if len(data) < 2 {
		return nil
	}

	headers := data[0]
	rows := data[1:]

	filtered := q.applyFilters(rows, headers)

	if q.orderBy != "" {
		filtered = q.applySort(filtered, headers)
	}

	filtered = q.applyLimit(filtered)

	return scanIntoSlice(filtered, headers, dest)
}

func (q *Query) applyFilters(rows [][]interface{}, headers []interface{}) [][]interface{} {
	if len(q.filters) == 0 {
		return rows
	}

	var result [][]interface{}
	for _, row := range rows {
		if q.matchesFilters(row, headers) {
			result = append(result, row)
		}
	}
	return result
}

func (q *Query) matchesFilters(row []interface{}, headers []interface{}) bool {
	for _, f := range q.filters {
		colIdx := -1
		for i, h := range headers {
			if h == f.Column {
				colIdx = i
				break
			}
		}
		if colIdx == -1 || colIdx >= len(row) {
			return false
		}

		if !matchesOperator(row[colIdx], f.Operator, f.Value) {
			return false
		}
	}
	return true
}

func matchesOperator(cell interface{}, op string, value interface{}) bool {
	cellStr := fmt.Sprintf("%v", cell)
	valueStr := fmt.Sprintf("%v", value)

	switch op {
	case "=", "==":
		return cellStr == valueStr
	case "!=":
		return cellStr != valueStr
	case ">":
		return compareValues(cell, value) > 0
	case ">=":
		return compareValues(cell, value) >= 0
	case "<":
		return compareValues(cell, value) < 0
	case "<=":
		return compareValues(cell, value) <= 0
	case "contains", "like":
		return strings.Contains(strings.ToLower(cellStr), strings.ToLower(valueStr))
	default:
		return false
	}
}

func compareValues(a, b interface{}) int {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	// Try numeric comparison
	aNum, aErr := strconv.ParseFloat(aStr, 64)
	bNum, bErr := strconv.ParseFloat(bStr, 64)

	if aErr == nil && bErr == nil {
		if aNum < bNum {
			return -1
		}
		if aNum > bNum {
			return 1
		}
		return 0
	}

	if aStr < bStr {
		return -1
	}
	if aStr > bStr {
		return 1
	}
	return 0
}

func (q *Query) applySort(rows [][]interface{}, headers []interface{}) [][]interface{} {
	return rows
}

func (q *Query) applyLimit(rows [][]interface{}) [][]interface{} {
	if q.limit > 0 && q.limit < len(rows) {
		return rows[:q.limit]
	}
	return rows
}

func structSliceToValues(records interface{}) ([][]interface{}, error) {
	v := reflect.ValueOf(records)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("records must be a slice")
	}

	var result [][]interface{}
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		row, err := structToValues(elem.Interface())
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, nil
}

func structToValues(record interface{}) ([]interface{}, error) {
	v := reflect.ValueOf(record)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("record must be a struct")
	}

	t := v.Type()
	var result []interface{}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		tag := fieldType.Tag.Get("quire")
		if tag == "-" {
			continue
		}

		result = append(result, field.Interface())
	}

	return result, nil
}

func scanIntoSlice(rows [][]interface{}, headers []interface{}, dest interface{}) error {
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr || destVal.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to a slice")
	}

	sliceVal := destVal.Elem()
	elemType := sliceVal.Type().Elem()

	for _, row := range rows {
		elem := reflect.New(elemType).Elem()
		if err := scanRow(row, headers, elem); err != nil {
			return err
		}
		sliceVal = reflect.Append(sliceVal, elem)
	}

	destVal.Elem().Set(sliceVal)
	return nil
}

func scanRow(row []interface{}, headers []interface{}, dest reflect.Value) error {
	if dest.Kind() == reflect.Ptr {
		dest = dest.Elem()
	}
	if dest.Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a struct")
	}

	t := dest.Type()
	for i := 0; i < dest.NumField(); i++ {
		field := dest.Field(i)
		fieldType := t.Field(i)

		tag := fieldType.Tag.Get("quire")
		if tag == "-" {
			continue
		}

		colName := fieldType.Name
		if tag != "" {
			colName = tag
		}

		colIdx := -1
		for j, h := range headers {
			if h == colName {
				colIdx = j
				break
			}
		}

		if colIdx == -1 || colIdx >= len(row) {
			continue
		}

		if err := setField(field, row[colIdx]); err != nil {
			return fmt.Errorf("failed to set field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

func setField(field reflect.Value, value interface{}) error {
	if !field.CanSet() {
		return nil
	}

	valueStr := fmt.Sprintf("%v", value)

	switch field.Kind() {
	case reflect.String:
		field.SetString(valueStr)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
			field.SetInt(i)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if i, err := strconv.ParseUint(valueStr, 10, 64); err == nil {
			field.SetUint(i)
		}
	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(valueStr, 64); err == nil {
			field.SetFloat(f)
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(valueStr); err == nil {
			field.SetBool(b)
		}
	default:
		if field.Kind() == reflect.Struct || field.Kind() == reflect.Slice {
			data, _ := json.Marshal(value)
			json.Unmarshal(data, field.Addr().Interface())
		}
	}

	return nil
}
