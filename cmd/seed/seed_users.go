package main

import (
	"log"
	"os"
	
	_ "github.com/mattn/go-sqlite3"
	"kolajAi/internal/database"
	"kolajAi/internal/repository"
	"kolajAi/internal/services"
	"kolajAi/internal/email"
	"kolajAi/internal/config"
)

func main() {
	log.Println("Test kullanıcıları oluşturuluyor...")
	
	// Veritabanı bağlantısı
	db, err := database.NewSQLiteConnection("kolajAi.db")
	if err != nil {
		log.Fatalf("Veritabanı bağlantısı kurulamadı: %v", err)
	}
	defer db.Close()
	
	// Config yükle
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("Config yüklenemedi, varsayılan değerler kullanılıyor: %v", err)
		cfg = config.GetDefaultConfig()
	}
	
	// Repository ve service'leri oluştur
	mysqlRepo := database.NewMySQLRepository(db)
	userRepo := repository.NewUserRepository(mysqlRepo)
	
	// Email service'i oluştur
	emailConfig := &email.Config{
		Host:     cfg.Email.SMTPHost,
		Port:     cfg.Email.SMTPPort,
		Username: cfg.Email.SMTPUser,
		Password: cfg.Email.SMTPPassword,
		FromName: cfg.Email.FromName,
		TLS:      cfg.Email.UseTLS,
	}
	emailService, err := email.NewService(emailConfig, "email_templates")
	if err != nil {
		log.Printf("Email service oluşturulamadı: %v", err)
		// Email service olmadan devam et
		emailService = nil
	}
	
	authService := services.NewAuthService(userRepo, emailService)
	
	// Test kullanıcıları
	testUsers := []map[string]string{
		{
			"name":     "Admin User",
			"email":    "admin@example.com",
			"password": "admin123",
			"phone":    "+90 555 111 1111",
			"role":     "admin",
		},
		{
			"name":     "Vendor User",
			"email":    "vendor@example.com",
			"password": "vendor123",
			"phone":    "+90 555 222 2222",
			"role":     "vendor",
		},
		{
			"name":     "Normal User",
			"email":    "user@example.com",
			"password": "user123",
			"phone":    "+90 555 333 3333",
			"role":     "user",
		},
		{
			"name":     "Test Admin",
			"email":    "test@admin.com",
			"password": "test123",
			"phone":    "+90 555 444 4444",
			"role":     "admin",
		},
	}
	
	// Kullanıcıları oluştur
	for _, userData := range testUsers {
		log.Printf("Kullanıcı oluşturuluyor: %s (%s)", userData["email"], userData["role"])
		
		// Önce kullanıcı var mı kontrol et
		exists, err := userRepo.EmailExists(userData["email"])
		if err != nil {
			log.Printf("Email kontrol hatası: %v", err)
			continue
		}
		
		if exists {
			log.Printf("Kullanıcı zaten mevcut: %s", userData["email"])
			
			// Mevcut kullanıcının şifresini güncelle
			hashedPassword, err := authService.CreateUserPassword(userData["password"])
			if err != nil {
				log.Printf("Şifre hash hatası: %v", err)
				continue
			}
			
			err = userRepo.ResetUserPassword(userData["email"], hashedPassword)
			if err != nil {
				log.Printf("Şifre güncelleme hatası: %v", err)
				continue
			}
			
			// Role güncelle
			query := `UPDATE users SET role = ? WHERE email = ?`
			_, err = db.Exec(query, userData["role"], userData["email"])
			if err != nil {
				log.Printf("Role güncelleme hatası: %v", err)
				continue
			}
			
			log.Printf("Kullanıcı güncellendi: %s", userData["email"])
			continue
		}
		
		// Yeni kullanıcı oluştur
		userID, err := authService.RegisterUser(userData)
		if err != nil {
			log.Printf("Kullanıcı oluşturma hatası: %v", err)
			continue
		}
		
		// Role güncelle (RegisterUser default olarak "user" role veriyor)
		if userData["role"] != "user" {
			query := `UPDATE users SET role = ? WHERE id = ?`
			_, err = db.Exec(query, userData["role"], userID)
			if err != nil {
				log.Printf("Role güncelleme hatası: %v", err)
				continue
			}
		}
		
		log.Printf("Kullanıcı başarıyla oluşturuldu: %s (ID: %d)", userData["email"], userID)
	}
	
	log.Println("\nTest kullanıcıları:")
	log.Println("==================")
	log.Println("Admin: admin@example.com / admin123")
	log.Println("Vendor: vendor@example.com / vendor123")
	log.Println("User: user@example.com / user123")
	log.Println("Test Admin: test@admin.com / test123")
	log.Println("==================")
	
	os.Exit(0)
}