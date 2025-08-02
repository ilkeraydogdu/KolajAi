package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"kolajAi/internal/database"
)

func dbQuery() {
	if len(os.Args) < 2 {
		fmt.Println("Kullanım: go run cmd/tools/dbquery/main.go \"SQL SORGUSU\"")
		os.Exit(1)
	}

	// Sorguyu al
	query := strings.Join(os.Args[1:], " ")
	fmt.Printf("Çalıştırılacak sorgu: %s\n\n", query)

	// Veritabanı yapılandırması
	dbConfig := database.DefaultConfig()

	// Veritabanı bağlantısı
	db, err := database.InitDB(dbConfig)
	if err != nil {
		log.Printf("Veritabanı bağlantısı yapılamadı: %v", err)
		return
	}
	defer db.Close()

	// Sorgu türünü belirle (SELECT, INSERT, UPDATE, DELETE)
	queryType := strings.ToUpper(strings.Split(strings.TrimSpace(query), " ")[0])

	// Sorguyu çalıştır
	switch queryType {
	case "SELECT":
		executeSelectQuery(db, query)
	case "INSERT", "UPDATE", "DELETE":
		executeUpdateQuery(db, query)
	default:
		fmt.Printf("Desteklenmeyen sorgu türü: %s\n", queryType)
	}
}

func main() {
	dbQuery()
}

// executeSelectQuery SELECT sorgusunu çalıştırır ve sonuçları gösterir
func executeSelectQuery(db *sql.DB, query string) {
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Sorgu çalıştırılırken hata oluştu: %v", err)
		return
	}
	defer rows.Close()

	// Sütun bilgilerini al
	columns, err := rows.Columns()
	if err != nil {
		log.Printf("Sütun bilgileri alınamadı: %v", err)
		return
	}

	// Sonuçları depolamak için değişkenler
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Sonuçları JSON formatında göstermek için
	var results []map[string]interface{}

	// Satırları oku
	for rows.Next() {
		// Satırı oku
		err := rows.Scan(valuePtrs...)
		if err != nil {
			log.Printf("Satır okunurken hata oluştu: %v", err)
		}

		// Satırı map'e dönüştür
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// byte array'i string'e dönüştür
			b, ok := val.([]byte)
			if ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}

		results = append(results, row)
	}

	// Sonuçları göster
	if len(results) == 0 {
		fmt.Println("Sonuç bulunamadı.")
		return
	}

	// JSON formatında göster
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Printf("JSON formatına dönüştürülürken hata oluştu: %v", err)
	}
	fmt.Println(string(jsonData))
}

// executeUpdateQuery INSERT, UPDATE veya DELETE sorgusunu çalıştırır
func executeUpdateQuery(db *sql.DB, query string) {
	result, err := db.Exec(query)
	if err != nil {
		log.Printf("Sorgu çalıştırılırken hata oluştu: %v", err)
	}

	// Etkilenen satır sayısını göster
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Etkilenen satır sayısı alınamadı: %v", err)
	}

	fmt.Printf("Etkilenen satır sayısı: %d\n", rowsAffected)

	// Eğer INSERT ise son eklenen ID'yi göster
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "INSERT") {
		lastInsertID, err := result.LastInsertId()
		if err != nil {
			log.Printf("Son eklenen ID alınamadı: %v", err)
		}
		fmt.Printf("Son eklenen ID: %d\n", lastInsertID)
	}
}
