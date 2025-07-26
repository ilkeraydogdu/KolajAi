package main

import (
	"fmt"
	"log"

	"kolajAi/internal/database"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("Testing repository...")

	// Connect to database
	db, err := database.NewSQLiteConnection("kolajAi.db")
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()
	fmt.Println("Database connected")

	// Create repositories
	mysqlRepo := database.NewMySQLRepository(db)
	fmt.Println("MySQLRepository created")

	repo := database.NewRepositoryWrapper(mysqlRepo)
	fmt.Println("RepositoryWrapper created")

	// Test basic operations
	var count int64
	count, err = repo.Count("products", map[string]interface{}{})
	if err != nil {
		fmt.Printf("Count error: %v\n", err)
	} else {
		fmt.Printf("Product count: %d\n", count)
	}

	fmt.Println("Repository test completed successfully!")
}