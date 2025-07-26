package main

import (
	"fmt"
	"log"

	"kolajAi/internal/database"
	"kolajAi/internal/services"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("Testing services...")

	// Connect to database
	db, err := database.NewSQLiteConnection("kolajAi.db")
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()
	fmt.Println("Database connected")

	// Create repositories
	mysqlRepo := database.NewMySQLRepository(db)
	repo := database.NewRepositoryWrapper(mysqlRepo)
	fmt.Println("Repository created")

	// Test services one by one
	fmt.Println("Creating VendorService...")
	vendorService := services.NewVendorService(repo)
	fmt.Printf("VendorService created: %v\n", vendorService != nil)

	fmt.Println("Creating ProductService...")
	productService := services.NewProductService(repo)
	fmt.Printf("ProductService created: %v\n", productService != nil)

	fmt.Println("Creating OrderService...")
	orderService := services.NewOrderService(repo)
	fmt.Printf("OrderService created: %v\n", orderService != nil)

	fmt.Println("Creating AuctionService...")
	auctionService := services.NewAuctionService(repo)
	fmt.Printf("AuctionService created: %v\n", auctionService != nil)

	fmt.Println("All services created successfully!")
}