# Quire

A Go library that provides a database-like interface for Google Sheets. Turn your spreadsheets into a document database with CRUD operations, filtered queries, and struct mapping.

[![Go Report Card](https://goreportcard.com/badge/github.com/elbader17/quire)](https://goreportcard.com/report/github.com/elbader17/quire)
[![GoDoc](https://godoc.org/github.com/elbader17/quire?status.svg)](https://godoc.org/github.com/elbader17/quire)

## Features

- **ORM-like API**: Familiar database operations (CRUD) on Google Sheets
- **Type-safe mapping**: Convert Go structs to sheet rows using tags
- **Query builder**: Build fluent queries with filters, limits, and ordering
- **Secure authentication**: Use Service Accounts for server-to-server authentication
- **Schema-less**: No need to define tables or columns beforehand
- **Concurrent**: Full support for concurrent operations with `context.Context`

## Table of Contents

- [Installation](#installation)
- [Google Cloud Configuration](#google-cloud-configuration)
- [Quick Start Guide](#quick-start-guide)
- [Core Concepts](#core-concepts)
- [Complete API](#complete-api)
  - [Configuration](#configuration)
  - [Connection](#connection)
  - [Inserting Data](#inserting-data)
  - [Updating Data](#updating-data)
  - [Deleting Data](#deleting-data)
  - [Queries](#queries)
  - [Filters](#filters)
  - [Struct Mapping](#struct-mapping)
- [Advanced Examples](#advanced-examples)
- [Best Practices](#best-practices)
- [Limitations](#limitations)
- [Troubleshooting](#troubleshooting)

## Installation

```bash
go get github.com/elbader17/quire
```

Requires Go 1.21 or higher.

## Google Cloud Configuration

Before using Quire, you need to set up access to the Google Sheets API:

### 1. Create a Project in Google Cloud Console

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the **Google Sheets API**:
   - Navigate to "APIs & Services" > "Library"
   - Search for "Google Sheets API"
   - Click "Enable"

### 2. Create a Service Account

1. Go to "IAM & Admin" > "Service Accounts"
2. Click "Create Service Account"
3. Assign a name and description
4. In "Grant this service account access to project", select "Editor" or a custom role with Sheets permissions
5. Create the account

### 3. Generate Keys

1. In the Service Accounts list, click on the created account
2. Go to the "Keys" tab
3. Click "Add Key" > "Create new key"
4. Select JSON format
5. Download the file (e.g., `service-account.json`)

**IMPORTANT**: Keep this file secure. Do not commit it to your repository.

### 4. Share Your Spreadsheet

1. Open your Google Sheet
2. Click "Share"
3. Add the Service Account email (found in the JSON, field `client_email`)
4. Assign "Editor" permissions

### 5. Get the Spreadsheet ID

The ID is in your Google Sheet URL:

```
https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
                                        ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                                              This is your Spreadsheet ID
```

## Quick Start Guide

### Minimal Example

```go
package main

import (
    "context"
    "log"
    "os"
    
    "github.com/elbader17/quire/pkg/quire"
)

type User struct {
    ID    int    `quire:"ID"`
    Name  string `quire:"Name"`
    Email string `quire:"Email"`
    Age   int    `quire:"Age"`
}

func main() {
    ctx := context.Background()
    
    // Read credentials
    credentials, err := os.ReadFile("service-account.json")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create connection
    db, err := quire.New(quire.Config{
        SpreadsheetID: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
        Credentials:   credentials,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Insert data
    users := db.Table("Users")
    err = users.Insert(ctx, []User{
        {ID: 1, Name: "Alice", Email: "alice@example.com", Age: 30},
        {ID: 2, Name: "Bob", Email: "bob@example.com", Age: 25},
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Query data
    var results []User
    err = users.Query().
        Where("Age", ">=", 25).
        Limit(10).
        Get(ctx, &results)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Found %d users", len(results))
    
    // Update data
    if len(results) > 0 {
        results[0].Name = "Alice Updated"
        err = users.Update(ctx, 0, results[0])
        if err != nil {
            log.Fatal(err)
        }
    }
    
    // Delete data
    err = users.DeleteWhere(ctx, "Status", "=", "inactive")
    if err != nil {
        log.Fatal(err)
    }
}
```

## Core Concepts

### Sheet Structure

Quire expects your Google Sheet to have the following structure:

| ID | Name  | Email             | Age |
|----|-------|-------------------|-----|
| 1  | Alice | alice@example.com | 30  |
| 2  | Bob   | bob@example.com   | 25  |

- **First row**: Headers matching your struct's `quire` tags
- **Following rows**: Data
- Each sheet (tab) represents a "table"

### Type Mapping

Quire automatically converts between Go types and Sheet values:

| Go Type | Sheet Example | Conversion |
|---------|---------------|------------|
| `string` | "Alice" | Direct |
| `int` | 42 or "42" | Parsing from string or float |
| `float64` | 3.14 or "3.14" | Parsing from string |
| `bool` | "true", "TRUE", "1" | Case-insensitive parsing |
| `uint` | 100 or "100" | Parsing with validation |

## Complete API

### Configuration

```go
type Config struct {
    // SpreadsheetID is your Google Sheet ID (required)
    // Found in URL: .../d/{SPREADSHEET_ID}/...
    SpreadsheetID string
    
    // Credentials is the content of the Service Account JSON file (required)
    // Use os.ReadFile() to load the file
    Credentials []byte
}
```

### Connection

```go
// Create a new database instance
db, err := quire.New(quire.Config{
    SpreadsheetID: "your-spreadsheet-id",
    Credentials:   credentials,
})
if err != nil {
    // Handle error (invalid credentials, wrong format, etc.)
}
defer db.Close()

// Get a table handle
users := db.Table("Users")  // "Users" is the sheet tab name
products := db.Table("Products")
```

### Inserting Data

```go
// Insert accepts a slice of structs
users := []User{
    {ID: 1, Name: "Alice", Email: "alice@example.com"},
    {ID: 2, Name: "Bob", Email: "bob@example.com"},
}

err := db.Table("Users").Insert(ctx, users)
if err != nil {
    log.Fatal(err)
}
```

**Notes:**
- Data is appended to the end of the sheet
- Duplicates are not checked automatically
- Fields with tag `quire:"-"` are ignored

### Updating Data

#### Update by Index

Update a specific row using its index (0-based, excluding header):

```go
// Update the first data row (index 0)
updatedUser := User{
    ID:    1,
    Name:  "Alice Updated",
    Email: "alice.new@example.com",
    Age:   31,
}

err := db.Table("Users").Update(ctx, 0, updatedUser)
if err != nil {
    log.Fatal(err)
}
```

#### Update with Filter

Update all rows matching a condition:

```go
// Update all users with Status = "pending"
updatedUser := User{
    ID:     99,
    Name:   "Updated Name",
    Email:  "updated@example.com",
    Age:    25,
    Status: "active",
}

err := db.Table("Users").UpdateWhere(ctx, "Status", "=", "pending", updatedUser)
if err != nil {
    log.Fatal(err)
}
```

### Deleting Data

#### Delete by Index

Delete a specific row using its index (0-based, excluding header):

```go
// Delete the first data row (index 0)
err := db.Table("Users").Delete(ctx, 0)
if err != nil {
    log.Fatal(err)
}
```

#### Delete with Filter

Delete all rows matching a condition:

```go
// Delete all users with Status = "deleted"
err := db.Table("Users").DeleteWhere(ctx, "Status", "=", "deleted")
if err != nil {
    log.Fatal(err)
}

// Delete inactive users over 60
err = db.Table("Users").DeleteWhere(ctx, "IsActive", "=", false)
if err != nil {
    log.Fatal(err)
}
```

**Notes:**
- `DeleteWhere` physically removes rows from the spreadsheet
- Rows are deleted in reverse order to maintain correct indices
- If no rows match, no error is returned

### Queries

#### Basic Query

```go
var users []User
err := db.Table("Users").Query().Get(ctx, &users)
```

#### With Limit

```go
var users []User
err := db.Table("Users").Query().
    Limit(10).
    Get(ctx, &users)
```

#### With Ordering (placeholder)

```go
var users []User
err := db.Table("Users").Query().
    OrderBy("Age", true).  // true = descending
    Get(ctx, &users)
```

**Note:** Ordering is currently a placeholder and doesn't actually sort results.

### Filters

#### Supported Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `=`, `==` | Equality | `Where("Status", "=", "active")` |
| `!=` | Inequality | `Where("Status", "!=", "deleted")` |
| `>` | Greater than | `Where("Age", ">", 18)` |
| `>=` | Greater or equal | `Where("Score", ">=", 90)` |
| `<` | Less than | `Where("Price", "<", 100)` |
| `<=` | Less or equal | `Where("Stock", "<=", 10)` |
| `contains`, `like` | Contains substring (case-insensitive) | `Where("Name", "contains", "john")` |

#### Multiple Filters (AND)

```go
var users []User
err := db.Table("Users").Query().
    Where("Age", ">=", 18).
    Where("Status", "=", "active").
    Where("Country", "=", "US").
    Get(ctx, &users)
// WHERE Age >= 18 AND Status = 'active' AND Country = 'US'
```

#### String Filters

```go
// Case-insensitive search
var users []User
err := db.Table("Users").Query().
    Where("Name", "contains", "alice").
    Get(ctx, &users)
// Matches "Alice", "ALICE", "alice smith", etc.
```

### Struct Mapping

#### Basic Tags

```go
type Product struct {
    ID          int     `quire:"ID"`          // Maps to "ID" column
    Name        string  `quire:"Name"`        // Maps to "Name" column
    Price       float64 `quire:"Price"`       // Maps to "Price" column
    Description string  `quire:"Description"` // Maps to "Description" column
}
```

#### Ignore Fields

```go
type User struct {
    ID        int       `quire:"ID"`
    Name      string    `quire:"Name"`
    Password  string    `quire:"-"`  // This field is ignored
    CreatedAt time.Time `quire:"-"`  // This field is also ignored
}
```

#### Mapping by Field Name

If you don't specify a tag, the field name is used:

```go
type User struct {
    ID   int    // Maps to "ID" column
    Name string // Maps to "Name" column
}
```

## Advanced Examples

### Complete CRUD

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/elbader17/quire/pkg/quire"
)

type Task struct {
    ID          int    `quire:"ID"`
    Title       string `quire:"Title"`
    Description string `quire:"Description"`
    Status      string `quire:"Status"`
    Priority    int    `quire:"Priority"`
}

func main() {
    ctx := context.Background()
    
    credentials, _ := os.ReadFile("service-account.json")
    db, _ := quire.New(quire.Config{
        SpreadsheetID: "your-spreadsheet-id",
        Credentials:   credentials,
    })
    defer db.Close()
    
    tasks := db.Table("Tasks")
    
    // CREATE
    newTasks := []Task{
        {ID: 1, Title: "Implement login", Status: "pending", Priority: 1},
        {ID: 2, Title: "Create tests", Status: "pending", Priority: 2},
        {ID: 3, Title: "Document API", Status: "done", Priority: 3},
    }
    
    if err := tasks.Insert(ctx, newTasks); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Tasks created")
    
    // READ - All pending high priority tasks
    var pendingTasks []Task
    err := tasks.Query().
        Where("Status", "=", "pending").
        Where("Priority", "<=", 2).
        Get(ctx, &pendingTasks)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Pending high priority tasks: %d\n", len(pendingTasks))
    for _, task := range pendingTasks {
        fmt.Printf("  - %s (Priority: %d)\n", task.Title, task.Priority)
    }
    
    // READ - Completed tasks
    var doneTasks []Task
    err = tasks.Query().
        Where("Status", "=", "done").
        Get(ctx, &doneTasks)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Completed tasks: %d\n", len(doneTasks))
    
    // UPDATE - Mark a task as completed
    if len(pendingTasks) > 0 {
        pendingTasks[0].Status = "done"
        err = tasks.Update(ctx, 0, pendingTasks[0])
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Task '%s' updated to completed\n", pendingTasks[0].Title)
    }
    
    // BULK UPDATE - Change priority of all pending tasks
    err = tasks.UpdateWhere(ctx, "Status", "=", "pending", Task{
        ID:     0,
        Title:  "Updated",
        Status: "pending",
        Priority: 5,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Priority of pending tasks updated")
    
    // DELETE - Remove cancelled tasks
    err = tasks.DeleteWhere(ctx, "Status", "=", "cancelled")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Cancelled tasks deleted")
}
```

### Inventory System

```go
type Product struct {
    SKU       string  `quire:"SKU"`
    Name      string  `quire:"Name"`
    Category  string  `quire:"Category"`
    Price     float64 `quire:"Price"`
    Stock     int     `quire:"Stock"`
}

func getLowStockProducts(ctx context.Context, db *quire.DB) ([]Product, error) {
    var products []Product
    err := db.Table("Inventory").Query().
        Where("Stock", "<", 10).
        OrderBy("Stock", false).  // Sort by stock ascending
        Get(ctx, &products)
    return products, err
}

func getProductsByCategory(ctx context.Context, db *quire.DB, category string) ([]Product, error) {
    var products []Product
    err := db.Table("Inventory").Query().
        Where("Category", "=", category).
        Where("Stock", ">", 0).  // Only available products
        Get(ctx, &products)
    return products, err
}

func searchProducts(ctx context.Context, db *quire.DB, query string) ([]Product, error) {
    var products []Product
    err := db.Table("Inventory").Query().
        Where("Name", "contains", query).
        Limit(20).
        Get(ctx, &products)
    return products, err
}
```

### User Management

```go
type User struct {
    ID        int    `quire:"ID"`
    Username  string `quire:"Username"`
    Email     string `quire:"Email"`
    Age       int    `quire:"Age"`
    Country   string `quire:"Country"`
    IsActive  bool   `quire:"IsActive"`
}

func getActiveUsersByCountry(ctx context.Context, db *quire.DB, country string) ([]User, error) {
    var users []User
    err := db.Table("Users").Query().
        Where("Country", "=", country).
        Where("IsActive", "=", true).
        Where("Age", ">=", 18).
        Get(ctx, &users)
    return users, err
}

func getAdultUsers(ctx context.Context, db *quire.DB, limit int) ([]User, error) {
    var users []User
    err := db.Table("Users").Query().
        Where("Age", ">=", 18).
        Limit(limit).
        Get(ctx, &users)
    return users, err
}

func findUserByEmail(ctx context.Context, db *quire.DB, email string) (*User, error) {
    var users []User
    err := db.Table("Users").Query().
        Where("Email", "=", email).
        Limit(1).
        Get(ctx, &users)
    if err != nil {
        return nil, err
    }
    if len(users) == 0 {
        return nil, fmt.Errorf("user not found")
    }
    return &users[0], nil
}
```

### Using Context with Timeouts

```go
func getUsersWithTimeout(db *quire.DB) ([]User, error) {
    // Create context with 5 second timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    var users []User
    err := db.Table("Users").Query().
        Limit(100).
        Get(ctx, &users)
    
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return nil, fmt.Errorf("timeout querying users")
        }
        return nil, err
    }
    
    return users, nil
}
```

## Best Practices

### 1. Error Handling

```go
if err != nil {
    // Don't use log.Fatal in production code
    // Better return the error to handle it up the stack
    return fmt.Errorf("failed to query users: %w", err)
}
```

### 2. Closing Connections

```go
db, err := quire.New(config)
if err != nil {
    return err
}
defer db.Close()  // Always use defer
```

### 3. Data Validation

```go
// Validate before inserting
if user.Age < 0 {
    return fmt.Errorf("age cannot be negative")
}

if user.Email == "" {
    return fmt.Errorf("email is required")
}

err := db.Table("Users").Insert(ctx, []User{user})
```

### 4. Manual Pagination

```go
func getUsersPaginated(ctx context.Context, db *quire.DB, page, pageSize int) ([]User, error) {
    // Note: This loads all data and then filters
    // For large volumes, consider other strategies
    var allUsers []User
    err := db.Table("Users").Query().Get(ctx, &allUsers)
    if err != nil {
        return nil, err
    }
    
    start := (page - 1) * pageSize
    end := start + pageSize
    
    if start >= len(allUsers) {
        return []User{}, nil
    }
    
    if end > len(allUsers) {
        end = len(allUsers)
    }
    
    return allUsers[start:end], nil
}
```

### 5. Sensitive Fields

```go
type User struct {
    ID       int    `quire:"ID"`
    Username string `quire:"Username"`
    Email    string `quire:"Email"`
    Password string `quire:"-"`  // Never store in Sheets
}
```

### 6. Connection Caching

```go
type UserRepository struct {
    db *quire.DB
}

func NewUserRepository(credentials []byte, spreadsheetID string) (*UserRepository, error) {
    db, err := quire.New(quire.Config{
        SpreadsheetID: spreadsheetID,
        Credentials:   credentials,
    })
    if err != nil {
        return nil, err
    }
    
    return &UserRepository{db: db}, nil
}

func (r *UserRepository) Close() error {
    return r.db.Close()
}

func (r *UserRepository) GetUsers(ctx context.Context) ([]User, error) {
    var users []User
    err := r.db.Table("Users").Query().Get(ctx, &users)
    return users, err
}
```

## Limitations

1. **No transactions**: Operations are not atomic. If an operation fails halfway, some changes may have been applied while others haven't.

2. **Google Rate Limits**: Google Sheets API has quotas:
   - 500 requests per 100 seconds per project
   - 100 requests per 100 seconds per user

3. **Row limit**: Google Sheets supports up to 10 million cells per spreadsheet.

4. **Sorting**: The `OrderBy` function is currently a placeholder and doesn't actually sort results.

5. **Concurrency**: While Quire supports `context.Context`, there's no row-level concurrency control.

## Troubleshooting

### Error: "spreadsheet ID is required"

Verify you're passing the SpreadsheetID correctly:

```go
db, err := quire.New(quire.Config{
    SpreadsheetID: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms", // No spaces
    Credentials:   credentials,
})
```

### Error: "credentials are required"

Make sure the credentials file exists and isn't empty:

```go
credentials, err := os.ReadFile("service-account.json")
if err != nil {
    log.Fatal("Could not read credentials file:", err)
}

if len(credentials) == 0 {
    log.Fatal("Credentials file is empty")
}
```

### Error: "failed to read data"

Possible causes:
1. Service Account doesn't have access to the spreadsheet
2. Spreadsheet doesn't exist
3. Sheet (tab) doesn't exist

Solution:
- Verify you shared the spreadsheet with the Service Account email
- Verify the table name matches the sheet tab name

### Error: "googleapi: Error 403: Forbidden"

Service Account doesn't have sufficient permissions:
- Go to Google Sheets > Share
- Make sure to give "Editor" permissions
- Check for domain restrictions

### Data doesn't map correctly

Verify that:
1. Headers in the first row match exactly with `quire` tags or field names
2. Data types are compatible
3. No extra spaces in headers

```go
// Sheet: | ID | Name  | Email |
// Struct:  ID   Name    Email   <- Exact match

type User struct {
    ID    int    `quire:"ID"`
    Name  string `quire:"Name"`
    Email string `quire:"Email"`
}
```

### Filters don't work

- Check column name (case-sensitive)
- Make sure operators are valid: `=`, `!=`, `>`, `>=`, `<`, `<=`, `contains`
- For strings, use `contains` for partial matches

### Slow performance

For large data volumes:
1. Use `Limit()` to paginate results
2. Consider splitting data into multiple sheets
3. Implement local cache if data doesn't change frequently
4. Keep in mind Google's API rate limits

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome. Please:

1. Fork the repository
2. Create a branch for your feature (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Acknowledgments

- [Google Sheets API](https://developers.google.com/sheets/api)
- [Google API Go Client](https://github.com/googleapis/google-api-go-client)
