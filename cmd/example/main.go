package main

import (
	"context"
	"fmt"
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

	credentials, err := os.ReadFile("service-account.json")
	if err != nil {
		log.Fatalf("Failed to read credentials: %v", err)
	}

	db, err := quire.New(quire.Config{
		SpreadsheetID: "your-spreadsheet-id",
		Credentials:   credentials,
	})
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	users := db.Table("Users")

	newUsers := []User{
		{ID: 1, Name: "Alice", Email: "alice@example.com", Age: 30},
		{ID: 2, Name: "Bob", Email: "bob@example.com", Age: 25},
	}

	if err := users.Insert(ctx, newUsers); err != nil {
		log.Fatalf("Failed to insert users: %v", err)
	}

	var results []User
	err = users.Query().
		Where("Age", ">=", 25).
		Limit(10).
		Get(ctx, &results)
	if err != nil {
		log.Fatalf("Failed to query users: %v", err)
	}

	fmt.Printf("Found %d users:\n", len(results))
	for _, u := range results {
		fmt.Printf("  - %s (%s)\n", u.Name, u.Email)
	}
}
