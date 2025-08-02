package main

import (
	"database/sql"
	"fmt"
	"log"

	"kolajAi/internal/database"
)

func main() {
	fmt.Println("KolajAI Veritabanı Bilgi Aracı")
	fmt.Println("==============================")

	// Veritabanı yapılandırması
	dbConfig := database.DefaultConfig()
	fmt.Printf("Veritabanı: %s@%s:%d/%s\n\n",
		dbConfig.User, dbConfig.Host, dbConfig.Port, dbConfig.DatabaseName)

	// Veritabanı bağlantısı
	db, err := database.InitDB(dbConfig)
	if err != nil {
		log.Printf("Veritabanı bağlantısı yapılamadı: %v", err)
		return
	}
	defer db.Close()

	fmt.Println("Veritabanı bağlantısı başarılı!")

	// Tabloları listele
	tables, err := getTables(db)
	if err != nil {
		log.Printf("Tablolar listelenirken hata oluştu: %v", err)
		return
	}

	fmt.Println("\nVeritabanı Tabloları:")
	fmt.Println("====================")
	for _, table := range tables {
		fmt.Println(table)
	}

	// Önce users tablosunu göster
	fmt.Println("\n*** USERS TABLOSU ***")
	fmt.Println("====================")
	usersColumns, err := getTableStructure(db, "users")
	if err != nil {
		log.Printf("Hata: users tablosu yapısı alınamadı: %v", err)
	} else {
		for _, col := range usersColumns {
			fmt.Printf("%-20s %-20s %-10s %-10s %s\n",
				col.Field, col.Type, col.Null, col.Key, col.Extra)
		}
	}

	// Diğer tabloların yapısını göster
	for _, table := range tables {
		if table == "users" {
			continue // Users tablosunu atla, zaten gösterdik
		}

		fmt.Printf("\nTablo Yapısı: %s\n", table)
		fmt.Println("====================")
		columns, err := getTableStructure(db, table)
		if err != nil {
			log.Printf("Hata: %s tablosu yapısı alınamadı: %v", table, err)
			continue
		}

		for _, col := range columns {
			fmt.Printf("%-20s %-20s %s\n", col.Field, col.Type, col.Extra)
		}
	}

	// İstatistikler
	fmt.Println("\nTablo İstatistikleri:")
	fmt.Println("====================")
	for _, table := range tables {
		count, err := getRowCount(db, table)
		if err != nil {
			log.Printf("Hata: %s tablosu satır sayısı alınamadı: %v", table, err)
			continue
		}
		fmt.Printf("%-20s: %d kayıt\n", table, count)
	}
}

// getTables veritabanındaki tüm tabloları listeler
func getTables(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// ColumnInfo tablo sütun bilgilerini temsil eder
type ColumnInfo struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default sql.NullString
	Extra   string
}

// getTableStructure tablo yapısını döndürür
func getTableStructure(db *sql.DB, table string) ([]ColumnInfo, error) {
	rows, err := db.Query(fmt.Sprintf("DESCRIBE %s", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		if err := rows.Scan(&col.Field, &col.Type, &col.Null, &col.Key, &col.Default, &col.Extra); err != nil {
			return nil, err
		}
		columns = append(columns, col)
	}

	return columns, nil
}

// getRowCount tablodaki satır sayısını döndürür
func getRowCount(db *sql.DB, table string) (int, error) {
	var count int
	err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
