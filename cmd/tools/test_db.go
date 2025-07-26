package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "kolajAi.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Test basic connection
	if err := db.Ping(); err != nil {
		log.Fatal("Ping failed:", err)
	}

	// Count products
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		log.Fatal("Count query failed:", err)
	}
	fmt.Printf("Total products: %d\n", count)

	// Count categories
	err = db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if err != nil {
		log.Fatal("Categories count failed:", err)
	}
	fmt.Printf("Total categories: %d\n", count)

	// Test a simple select
	rows, err := db.Query("SELECT id, name, price FROM products LIMIT 5")
	if err != nil {
		log.Fatal("Products query failed:", err)
	}
	defer rows.Close()

	fmt.Println("\nProducts:")
	for rows.Next() {
		var id int
		var name string
		var price float64
		
		if err := rows.Scan(&id, &name, &price); err != nil {
			log.Fatal("Scan failed:", err)
		}
		fmt.Printf("ID: %d, Name: %s, Price: %.2f\n", id, name, price)
	}

	if err := rows.Err(); err != nil {
		log.Fatal("Rows error:", err)
	}
}