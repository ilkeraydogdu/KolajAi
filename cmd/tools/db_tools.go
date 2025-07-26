package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "info":
		runDBInfo()
	case "query":
		runDBQuery()
	default:
		fmt.Printf("Bilinmeyen komut: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Kullanım: go run cmd/tools/db_tools.go KOMUT [PARAMETRELER]")
	fmt.Println("\nKomutlar:")
	fmt.Println("  info    Veritabanı yapısı hakkında bilgi gösterir")
	fmt.Println("  query   SQL sorgusu çalıştırır (örn: \"SELECT * FROM users\")")
}

func runDBInfo() {
	cmd := exec.Command("go", "run", filepath.Join("cmd", "tools", "dbinfo", "main.go"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Hata: %v\n", err)
		os.Exit(1)
	}
}

func runDBQuery() {
	if len(os.Args) < 3 {
		fmt.Println("Hata: Sorgu belirtilmedi")
		fmt.Println("Kullanım: go run cmd/tools/db_tools.go query \"SQL SORGUSU\"")
		os.Exit(1)
	}

	args := []string{"run", filepath.Join("cmd", "tools", "dbquery", "main.go")}
	args = append(args, os.Args[2:]...)

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Hata: %v\n", err)
		os.Exit(1)
	}
}
