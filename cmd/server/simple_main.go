package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"kolajAi/internal/database"
	"kolajAi/internal/models"
	"kolajAi/internal/services"
	"strings"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("KolajAI Basit Server başlatılıyor...")

	// Database connection
	db, err := database.NewSQLiteConnection("kolajAi.db")
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()
	fmt.Println("Database connected")

	// Repository
	mysqlRepo := database.NewMySQLRepository(db)
	repo := database.NewRepositoryWrapper(mysqlRepo)
	fmt.Println("Repository created")

	// Services
	productService := services.NewProductService(repo)
	vendorService := services.NewVendorService(repo)
	fmt.Println("Services created")

	// Routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/products", func(w http.ResponseWriter, r *http.Request) {
		apiProductsHandler(w, r, productService)
	})
	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		productsPageHandler(w, r, productService)
	})
	http.HandleFunc("/product/", func(w http.ResponseWriter, r *http.Request) {
		productDetailHandler(w, r, productService)
	})
	http.HandleFunc("/cart", cartHandler)
	http.HandleFunc("/api/cart/add", addToCartHandler)
	
	// Auth routes
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/api/login", apiLoginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/logout", logoutHandler)
	
	// Admin routes
	http.HandleFunc("/admin", adminDashboardHandler)
	http.HandleFunc("/admin/products", func(w http.ResponseWriter, r *http.Request) {
		adminProductsHandler(w, r, productService)
	})
	http.HandleFunc("/admin/vendors", func(w http.ResponseWriter, r *http.Request) {
		adminVendorsHandler(w, r, vendorService)
	})
	http.HandleFunc("/admin/api/products", func(w http.ResponseWriter, r *http.Request) {
		adminApiProductsHandler(w, r, productService)
	})

	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	fmt.Println("Server starting on :8082...")
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	userInfo := ""
	
	if session != nil {
		userName := session["name"].(string)
		isAdminUser := session["is_admin"].(bool)
		if isAdminUser {
			userInfo = fmt.Sprintf(`
				<div class="flex items-center space-x-4">
					<span class="text-sm text-gray-600">Hoş geldin, %s</span>
					<a href="/admin" class="bg-purple-600 text-white px-3 py-1 rounded text-sm hover:bg-purple-700">Admin Panel</a>
					<a href="/logout" class="bg-red-600 text-white px-3 py-1 rounded text-sm hover:bg-red-700">Çıkış</a>
				</div>`, userName)
		} else {
			userInfo = fmt.Sprintf(`
				<div class="flex items-center space-x-4">
					<span class="text-sm text-gray-600">Hoş geldin, %s</span>
					<a href="/logout" class="bg-red-600 text-white px-3 py-1 rounded text-sm hover:bg-red-700">Çıkış</a>
				</div>`, userName)
		}
	} else {
		userInfo = `
			<div class="flex items-center space-x-4">
				<a href="/login" class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">Giriş Yap</a>
				<a href="/register" class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">Kayıt Ol</a>
			</div>`
	}

	html := `
<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>KolajAI Marketplace</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <!-- Header -->
    <header class="bg-white shadow-sm border-b mb-8">
        <div class="container mx-auto px-4 py-4">
            <div class="flex justify-between items-center">
                <h1 class="text-2xl font-bold text-blue-600">KolajAI Marketplace</h1>
                %s
            </div>
        </div>
    </header>

    <div class="container mx-auto px-4 py-8">
        <div class="text-center mb-12">
            <h1 class="text-5xl font-bold text-gray-800 mb-4">KolajAI Marketplace</h1>
            <p class="text-gray-600 text-xl mb-8">Çoklu satıcı e-ticaret platformu</p>
            <div class="flex justify-center space-x-4">
                <a href="/products" class="bg-blue-600 text-white px-8 py-3 rounded-lg font-semibold hover:bg-blue-700 transition">
                    Ürünleri İncele
                </a>
                <a href="/cart" class="border-2 border-blue-600 text-blue-600 px-8 py-3 rounded-lg font-semibold hover:bg-blue-600 hover:text-white transition">
                    Sepeti Görüntüle
                </a>
            </div>
        </div>
        
        <div class="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-6xl mx-auto">
            <div class="bg-white p-8 rounded-lg shadow-lg text-center hover:shadow-xl transition">
                <div class="text-4xl mb-4">🛍️</div>
                <h3 class="text-2xl font-semibold mb-4">Ürünler</h3>
                <p class="text-gray-600 mb-6">Binlerce ürün arasından seçim yapın</p>
                <a href="/products" class="bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 inline-block">
                    Ürünleri Gör
                </a>
            </div>
            
            <div class="bg-white p-8 rounded-lg shadow-lg text-center hover:shadow-xl transition">
                <div class="text-4xl mb-4">🔧</div>
                <h3 class="text-2xl font-semibold mb-4">API Test</h3>
                <p class="text-gray-600 mb-6">Geliştiriciler için JSON API</p>
                <a href="/api/products" class="bg-green-600 text-white px-6 py-3 rounded-lg hover:bg-green-700 inline-block">
                    API Test
                </a>
            </div>
            
            <div class="bg-white p-8 rounded-lg shadow-lg text-center hover:shadow-xl transition">
                <div class="text-4xl mb-4">⚙️</div>
                <h3 class="text-2xl font-semibold mb-4">Yönetim</h3>
                <p class="text-gray-600 mb-6">Admin paneli ve yönetim araçları</p>
                <a href="/admin" class="bg-purple-600 text-white px-6 py-3 rounded-lg hover:bg-purple-700 inline-block">
                    Admin Panel
                </a>
            </div>
        </div>

        <!-- Features Section -->
        <div class="mt-16 text-center">
            <h2 class="text-3xl font-bold text-gray-800 mb-8">Platform Özellikleri</h2>
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <div class="bg-white p-6 rounded-lg shadow-md">
                    <div class="text-3xl mb-3">🔍</div>
                    <h4 class="font-semibold mb-2">Akıllı Arama</h4>
                    <p class="text-sm text-gray-600">Gelişmiş arama ve filtreleme</p>
                </div>
                <div class="bg-white p-6 rounded-lg shadow-md">
                    <div class="text-3xl mb-3">🛒</div>
                    <h4 class="font-semibold mb-2">Sepet Yönetimi</h4>
                    <p class="text-sm text-gray-600">Kolay sepet ve sipariş yönetimi</p>
                </div>
                <div class="bg-white p-6 rounded-lg shadow-md">
                    <div class="text-3xl mb-3">📱</div>
                    <h4 class="font-semibold mb-2">Responsive Tasarım</h4>
                    <p class="text-sm text-gray-600">Tüm cihazlarda mükemmel görünüm</p>
                </div>
                <div class="bg-white p-6 rounded-lg shadow-md">
                    <div class="text-3xl mb-3">🔐</div>
                    <h4 class="font-semibold mb-2">Güvenli Alışveriş</h4>
                    <p class="text-sm text-gray-600">SSL şifreleme ve güvenli ödeme</p>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`

	finalHTML := fmt.Sprintf(html, userInfo)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(finalHTML))
}

func apiProductsHandler(w http.ResponseWriter, r *http.Request, productService *services.ProductService) {
	products, err := productService.GetFeaturedProducts(10, 0)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    products,
		"count":   len(products),
	})
}

func productsPageHandler(w http.ResponseWriter, r *http.Request, productService *services.ProductService) {
	// Get query parameters
	searchTerm := r.URL.Query().Get("search")

	var products []models.Product
	var err error

	if searchTerm != "" {
		// Search products
		products, err = productService.SearchProducts(searchTerm, 20, 0)
	} else {
		// Get all featured products
		products, err = productService.GetFeaturedProducts(20, 0)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Ürünler yüklenirken hata: %v", err), http.StatusInternalServerError)
		return
	}

	categories, _ := productService.GetAllCategories()

	html := `
<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Ürünler - KolajAI</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="flex justify-between items-center mb-8">
            <h1 class="text-3xl font-bold text-gray-800">Ürünler</h1>
            <a href="/" class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">Ana Sayfa</a>
        </div>
        
        <!-- Search and Filters -->
        <div class="bg-white rounded-lg shadow-md p-6 mb-8">
            <div class="md:flex md:items-center md:space-x-4">
                <div class="flex-1 mb-4 md:mb-0">
                    <form method="GET" class="flex">
                        <input type="text" name="search" value="%s" 
                               placeholder="Ürün ara..." 
                               class="flex-1 px-4 py-2 border border-gray-300 rounded-l-lg focus:outline-none focus:ring-2 focus:ring-blue-500">
                        <button type="submit" 
                                class="bg-blue-600 text-white px-6 py-2 rounded-r-lg hover:bg-blue-700 transition">
                            <span class="hidden md:inline">Ara</span>
                            <span class="md:hidden">🔍</span>
                        </button>
                    </form>
                </div>
                <div class="flex space-x-2">
                    <a href="/products" class="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600 transition text-sm">
                        Filtreleri Temizle
                    </a>
                </div>
            </div>
        </div>
        
        <!-- Categories -->
        <div class="mb-8">
            <h2 class="text-xl font-semibold mb-4">Kategoriler</h2>
            <div class="flex flex-wrap gap-2">
                %s
            </div>
        </div>
        
        <!-- Products -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            %s
        </div>
        
        %s
    </div>
    
    <script>
        function quickAddToCart(productId) {
            fetch('/api/cart/add', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    product_id: productId,
                    quantity: 1
                })
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('Ürün sepete eklendi!');
                } else {
                    alert('Hata: ' + data.message);
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert('Bir hata oluştu!');
            });
        }
    </script>
</body>
</html>`

	// Categories HTML
	categoriesHTML := ""
	for _, cat := range categories {
		categoriesHTML += fmt.Sprintf(`
            <a href="/products?category=%d" class="bg-blue-100 text-blue-800 px-3 py-1 rounded-full text-sm hover:bg-blue-200 transition">
                %s
            </a>`, cat.ID, cat.Name)
	}

	// Products HTML
	productsHTML := ""
	for _, product := range products {
		productsHTML += fmt.Sprintf(`
            <div class="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition">
                <div class="h-48 bg-gray-200 flex items-center justify-center">
                    <span class="text-gray-500 text-4xl">📱</span>
                </div>
                <div class="p-4">
                    <h3 class="font-semibold text-lg mb-2">
                        <a href="/product/%d" class="hover:text-blue-600">%s</a>
                    </h3>
                    <p class="text-gray-600 text-sm mb-3">%s</p>
                    <div class="flex justify-between items-center mb-3">
                        <span class="text-xl font-bold text-blue-600">%.2f TL</span>
                        <span class="text-xs text-gray-500">Stok: %d</span>
                    </div>
                    <div class="flex space-x-2">
                        <a href="/product/%d" class="flex-1 bg-blue-600 text-white px-3 py-1 rounded text-sm text-center hover:bg-blue-700">
                            Detay
                        </a>
                        <button onclick="quickAddToCart(%d)" class="bg-green-600 text-white px-3 py-1 rounded text-sm hover:bg-green-700">
                            Sepet
                        </button>
                    </div>
                </div>
            </div>`, product.ID, product.Name, product.ShortDesc, product.Price, product.Stock, product.ID, product.ID)
	}

	// Status message
	statusHTML := ""
	if len(products) == 0 {
		message := "Henüz ürün bulunmuyor."
		if searchTerm != "" {
			message = fmt.Sprintf("'%s' için ürün bulunamadı.", searchTerm)
		}
		statusHTML = fmt.Sprintf(`<div class="text-center py-12">
            <p class="text-gray-500 text-lg">%s</p>
            <a href="/products" class="inline-block mt-4 bg-blue-600 text-white px-6 py-2 rounded hover:bg-blue-700">
                Tüm Ürünleri Görüntüle
            </a>
        </div>`, message)
	} else {
		statusHTML = fmt.Sprintf(`<div class="text-center mt-8">
            <p class="text-gray-600">Toplam %d ürün görüntüleniyor</p>
        </div>`, len(products))
	}

	finalHTML := fmt.Sprintf(html, searchTerm, categoriesHTML, productsHTML, statusHTML)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(finalHTML))
}

// Simple in-memory cart and sessions
var cart = make(map[string]int) // productID -> quantity
var sessions = make(map[string]map[string]interface{}) // sessionID -> user data

// Simple session management
func getSession(r *http.Request) map[string]interface{} {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil
	}
	return sessions[cookie.Value]
}

func createSession(w http.ResponseWriter, userData map[string]interface{}) string {
	sessionID := fmt.Sprintf("sess_%d", time.Now().UnixNano())
	sessions[sessionID] = userData
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
	})
	return sessionID
}

func destroySession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		delete(sessions, cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func isAdmin(r *http.Request) bool {
	session := getSession(r)
	if session == nil {
		return false
	}
	isAdmin, ok := session["is_admin"].(bool)
	return ok && isAdmin
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Giriş - KolajAI</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen flex items-center justify-center">
        <div class="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
            <div class="text-center mb-8">
                <h1 class="text-3xl font-bold text-gray-800">KolajAI</h1>
                <p class="text-gray-600 mt-2">Marketplace'e Giriş</p>
            </div>
            
            <form id="loginForm" class="space-y-6">
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-2">Email</label>
                    <input type="email" id="email" required
                           class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                           placeholder="admin@kolajAi.com">
                </div>
                
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-2">Şifre</label>
                    <input type="password" id="password" required
                           class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                           placeholder="admin123">
                </div>
                
                <button type="submit" 
                        class="w-full bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 transition">
                    Giriş Yap
                </button>
            </form>
            
            <div class="mt-6 text-center">
                <p class="text-sm text-gray-600">
                    Hesabınız yok mu? 
                    <a href="/register" class="text-blue-600 hover:text-blue-800">Kayıt Ol</a>
                </p>
                <p class="text-xs text-gray-500 mt-4">
                    Test: admin@kolajAi.com / admin123
                </p>
            </div>
            
            <div class="mt-4 text-center">
                <a href="/" class="text-sm text-gray-600 hover:text-gray-800">← Ana Sayfaya Dön</a>
            </div>
        </div>
    </div>
    
    <script>
        document.getElementById('loginForm').addEventListener('submit', function(e) {
            e.preventDefault();
            
            const email = document.getElementById('email').value;
            const password = document.getElementById('password').value;
            
            fetch('/api/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    email: email,
                    password: password
                })
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    if (data.is_admin) {
                        window.location.href = '/admin';
                    } else {
                        window.location.href = '/';
                    }
                } else {
                    alert('Hata: ' + data.message);
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert('Bir hata oluştu!');
            });
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func apiLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request",
		})
		return
	}

	// Simple authentication (in real app, check against database)
	isValidUser := false
	isAdminUser := false
	userName := ""

	if req.Email == "admin@kolajAi.com" && req.Password == "admin123" {
		isValidUser = true
		isAdminUser = true
		userName = "Admin User"
	} else if req.Email == "user@kolajAi.com" && req.Password == "user123" {
		isValidUser = true
		isAdminUser = false
		userName = "Normal User"
	}

	if isValidUser {
		createSession(w, map[string]interface{}{
			"email":    req.Email,
			"name":     userName,
			"is_admin": isAdminUser,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"message":  "Login successful",
			"is_admin": isAdminUser,
			"name":     userName,
		})
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid email or password",
		})
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Kayıt - KolajAI</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen flex items-center justify-center">
        <div class="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
            <div class="text-center mb-8">
                <h1 class="text-3xl font-bold text-gray-800">KolajAI</h1>
                <p class="text-gray-600 mt-2">Yeni Hesap Oluştur</p>
            </div>
            
            <div class="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
                <p class="text-blue-800 text-sm">
                    🚧 Kayıt özelliği henüz geliştirme aşamasında. 
                    Test için giriş yapın.
                </p>
            </div>
            
            <div class="text-center space-y-4">
                <a href="/login" class="block w-full bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 transition">
                    Giriş Sayfasına Dön
                </a>
                <a href="/" class="block text-sm text-gray-600 hover:text-gray-800">← Ana Sayfaya Dön</a>
            </div>
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	destroySession(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func productDetailHandler(w http.ResponseWriter, r *http.Request, productService *services.ProductService) {
	// Extract product ID from URL
	path := r.URL.Path
	idStr := path[len("/product/"):]
	id := 0
	fmt.Sscanf(idStr, "%d", &id)

	if id == 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := productService.GetProductByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Product not found: %v", err), http.StatusNotFound)
		return
	}

	html := `
<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - KolajAI</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="mb-4">
            <a href="/products" class="text-blue-600 hover:text-blue-800">← Ürünlere Dön</a>
        </div>
        
        <div class="bg-white rounded-lg shadow-lg overflow-hidden">
            <div class="md:flex">
                <div class="md:w-1/2">
                    <div class="h-96 bg-gray-200 flex items-center justify-center">
                        <span class="text-gray-500 text-8xl">📱</span>
                    </div>
                </div>
                <div class="md:w-1/2 p-8">
                    <h1 class="text-3xl font-bold text-gray-800 mb-4">%s</h1>
                    <p class="text-gray-600 mb-6">%s</p>
                    
                    <div class="mb-6">
                        <div class="flex items-center mb-2">
                            <span class="text-3xl font-bold text-blue-600">%.2f TL</span>
                            <span class="text-lg text-gray-500 line-through ml-4">%.2f TL</span>
                        </div>
                        <p class="text-sm text-gray-500">SKU: %s</p>
                    </div>
                    
                    <div class="mb-6">
                        <p class="text-sm text-gray-600">Stok: <span class="font-semibold">%d adet</span></p>
                        <p class="text-sm text-gray-600">Ağırlık: <span class="font-semibold">%.2f kg</span></p>
                        <p class="text-sm text-gray-600">Boyutlar: <span class="font-semibold">%s</span></p>
                    </div>
                    
                    <div class="flex items-center space-x-4">
                        <input type="number" id="quantity" value="1" min="1" max="%d" 
                               class="w-20 px-3 py-2 border border-gray-300 rounded">
                        <button onclick="addToCart(%d)" 
                                class="bg-green-600 text-white px-6 py-2 rounded-lg hover:bg-green-700 transition">
                            Sepete Ekle
                        </button>
                        <a href="/cart" class="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition">
                            Sepeti Görüntüle
                        </a>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="mt-8 bg-white rounded-lg shadow-lg p-6">
            <h2 class="text-2xl font-bold mb-4">Ürün Açıklaması</h2>
            <p class="text-gray-700 leading-relaxed">%s</p>
            
            <div class="mt-6">
                <h3 class="text-lg font-semibold mb-2">Etiketler</h3>
                <div class="flex flex-wrap gap-2">
                    %s
                </div>
            </div>
        </div>
    </div>
    
    <script>
        function addToCart(productId) {
            const quantity = document.getElementById('quantity').value;
            
            fetch('/api/cart/add', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    product_id: productId,
                    quantity: parseInt(quantity)
                })
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('Ürün sepete eklendi!');
                } else {
                    alert('Hata: ' + data.message);
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert('Bir hata oluştu!');
            });
        }
    </script>
</body>
</html>`

	// Create tags HTML
	tagsHTML := ""
	if product.Tags != "" {
		tags := strings.Split(product.Tags, ",")
		for _, tag := range tags {
			tagsHTML += fmt.Sprintf(`<span class="bg-blue-100 text-blue-800 px-2 py-1 rounded text-sm">%s</span>`, strings.TrimSpace(tag))
		}
	}

	finalHTML := fmt.Sprintf(html, 
		product.Name, product.Name, product.ShortDesc, 
		product.Price, product.ComparePrice, product.SKU,
		product.Stock, product.Weight, product.Dimensions,
		product.Stock, product.ID, product.Description, tagsHTML)
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(finalHTML))
}

func cartHandler(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sepet - KolajAI</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <div class="flex justify-between items-center mb-8">
            <h1 class="text-3xl font-bold text-gray-800">Alışveriş Sepeti</h1>
            <a href="/products" class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">Alışverişe Devam</a>
        </div>
        
        <div class="bg-white rounded-lg shadow-lg p-6">
            %s
        </div>
    </div>
</body>
</html>`

	cartHTML := ""
	if len(cart) == 0 {
		cartHTML = `
            <div class="text-center py-12">
                <p class="text-gray-500 text-lg">Sepetiniz boş</p>
                <a href="/products" class="inline-block mt-4 bg-blue-600 text-white px-6 py-2 rounded hover:bg-blue-700">
                    Alışverişe Başla
                </a>
            </div>`
	} else {
		cartHTML = `
            <div class="space-y-4">
                <h2 class="text-xl font-semibold">Sepetinizdeki Ürünler</h2>`
		
		for productID, quantity := range cart {
			cartHTML += fmt.Sprintf(`
                <div class="flex justify-between items-center border-b pb-4">
                    <div>
                        <p class="font-semibold">Ürün ID: %s</p>
                        <p class="text-gray-600">Adet: %d</p>
                    </div>
                    <button onclick="removeFromCart('%s')" class="text-red-600 hover:text-red-800">
                        Kaldır
                    </button>
                </div>`, productID, quantity, productID)
		}
		
		cartHTML += `
                <div class="mt-6 pt-4 border-t">
                    <button class="bg-green-600 text-white px-6 py-2 rounded hover:bg-green-700">
                        Siparişi Tamamla
                    </button>
                </div>
            </div>`
	}

	finalHTML := fmt.Sprintf(html, cartHTML)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(finalHTML))
}

func addToCartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request",
		})
		return
	}

	// Add to cart
	productKey := fmt.Sprintf("%d", req.ProductID)
	cart[productKey] += req.Quantity

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Product added to cart",
		"cart_count": len(cart),
	})
}

func adminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	if !isAdmin(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	session := getSession(r)
	userName := session["name"].(string)

	html := `
<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Admin Panel - KolajAI</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <!-- Header -->
    <header class="bg-white shadow-sm border-b">
        <div class="container mx-auto px-4 py-4">
            <div class="flex justify-between items-center">
                <div class="flex items-center space-x-4">
                    <h1 class="text-2xl font-bold text-gray-800">KolajAI Admin</h1>
                    <span class="text-sm text-gray-500">Yönetim Paneli</span>
                </div>
                <div class="flex items-center space-x-4">
                    <span class="text-sm text-gray-600">Hoş geldin, %s</span>
                    <a href="/" class="text-blue-600 hover:text-blue-800 text-sm">Ana Sayfa</a>
                    <a href="/logout" class="bg-red-600 text-white px-3 py-1 rounded text-sm hover:bg-red-700">Çıkış</a>
                </div>
            </div>
        </div>
    </header>

    <div class="container mx-auto px-4 py-8">
        <!-- Stats Cards -->
        <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            <div class="bg-white rounded-lg shadow-md p-6">
                <div class="flex items-center">
                    <div class="p-3 bg-blue-100 rounded-full">
                        <span class="text-blue-600 text-xl">📦</span>
                    </div>
                    <div class="ml-4">
                        <p class="text-sm text-gray-500">Toplam Ürün</p>
                        <p class="text-2xl font-bold text-gray-800" id="totalProducts">-</p>
                    </div>
                </div>
            </div>
            
            <div class="bg-white rounded-lg shadow-md p-6">
                <div class="flex items-center">
                    <div class="p-3 bg-green-100 rounded-full">
                        <span class="text-green-600 text-xl">🏪</span>
                    </div>
                    <div class="ml-4">
                        <p class="text-sm text-gray-500">Aktif Satıcı</p>
                        <p class="text-2xl font-bold text-gray-800" id="totalVendors">-</p>
                    </div>
                </div>
            </div>
            
            <div class="bg-white rounded-lg shadow-md p-6">
                <div class="flex items-center">
                    <div class="p-3 bg-yellow-100 rounded-full">
                        <span class="text-yellow-600 text-xl">🛒</span>
                    </div>
                    <div class="ml-4">
                        <p class="text-sm text-gray-500">Sepetteki Ürün</p>
                        <p class="text-2xl font-bold text-gray-800" id="cartItems">%d</p>
                    </div>
                </div>
            </div>
            
            <div class="bg-white rounded-lg shadow-md p-6">
                <div class="flex items-center">
                    <div class="p-3 bg-purple-100 rounded-full">
                        <span class="text-purple-600 text-xl">💰</span>
                    </div>
                    <div class="ml-4">
                        <p class="text-sm text-gray-500">Toplam Değer</p>
                        <p class="text-2xl font-bold text-gray-800">₺45.000</p>
                    </div>
                </div>
            </div>
        </div>

        <!-- Quick Actions -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            <div class="bg-white rounded-lg shadow-md p-6">
                <h3 class="text-lg font-semibold mb-4">Ürün Yönetimi</h3>
                <div class="space-y-3">
                    <a href="/admin/products" class="block w-full bg-blue-600 text-white py-2 px-4 rounded hover:bg-blue-700 text-center">
                        Ürünleri Yönet
                    </a>
                    <button class="w-full bg-green-600 text-white py-2 px-4 rounded hover:bg-green-700">
                        Yeni Ürün Ekle
                    </button>
                </div>
            </div>
            
            <div class="bg-white rounded-lg shadow-md p-6">
                <h3 class="text-lg font-semibold mb-4">Satıcı Yönetimi</h3>
                <div class="space-y-3">
                    <a href="/admin/vendors" class="block w-full bg-purple-600 text-white py-2 px-4 rounded hover:bg-purple-700 text-center">
                        Satıcıları Yönet
                    </a>
                    <button class="w-full bg-orange-600 text-white py-2 px-4 rounded hover:bg-orange-700">
                        Onay Bekleyenler
                    </button>
                </div>
            </div>
            
            <div class="bg-white rounded-lg shadow-md p-6">
                <h3 class="text-lg font-semibold mb-4">Sistem</h3>
                <div class="space-y-3">
                    <button class="w-full bg-gray-600 text-white py-2 px-4 rounded hover:bg-gray-700">
                        Sistem Ayarları
                    </button>
                    <button class="w-full bg-red-600 text-white py-2 px-4 rounded hover:bg-red-700">
                        Raporlar
                    </button>
                </div>
            </div>
        </div>

        <!-- Recent Activity -->
        <div class="bg-white rounded-lg shadow-md p-6">
            <h3 class="text-lg font-semibold mb-4">Son Aktiviteler</h3>
            <div class="space-y-4">
                <div class="flex items-center justify-between border-b pb-3">
                    <div class="flex items-center space-x-3">
                        <span class="text-green-600">✓</span>
                        <span class="text-sm">Yeni ürün eklendi: iPhone 15 Pro</span>
                    </div>
                    <span class="text-xs text-gray-500">2 saat önce</span>
                </div>
                <div class="flex items-center justify-between border-b pb-3">
                    <div class="flex items-center space-x-3">
                        <span class="text-blue-600">📦</span>
                        <span class="text-sm">Ürün stoku güncellendi: Samsung Galaxy S24</span>
                    </div>
                    <span class="text-xs text-gray-500">5 saat önce</span>
                </div>
                <div class="flex items-center justify-between">
                    <div class="flex items-center space-x-3">
                        <span class="text-yellow-600">⚠️</span>
                        <span class="text-sm">Düşük stok uyarısı: Nike Air Max</span>
                    </div>
                    <span class="text-xs text-gray-500">1 gün önce</span>
                </div>
            </div>
        </div>
    </div>

    <script>
        // Load stats
        fetch('/api/products')
            .then(response => response.json())
            .then(data => {
                document.getElementById('totalProducts').textContent = data.count || 0;
            });
    </script>
</body>
</html>`

	finalHTML := fmt.Sprintf(html, userName, len(cart))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(finalHTML))
}

func adminProductsHandler(w http.ResponseWriter, r *http.Request, productService *services.ProductService) {
	if !isAdmin(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	html := `
<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Ürün Yönetimi - KolajAI Admin</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <!-- Header -->
    <header class="bg-white shadow-sm border-b">
        <div class="container mx-auto px-4 py-4">
            <div class="flex justify-between items-center">
                <div class="flex items-center space-x-4">
                    <a href="/admin" class="text-blue-600 hover:text-blue-800">← Admin Panel</a>
                    <h1 class="text-2xl font-bold text-gray-800">Ürün Yönetimi</h1>
                </div>
                <div class="flex items-center space-x-4">
                    <button class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">
                        + Yeni Ürün
                    </button>
                    <a href="/logout" class="bg-red-600 text-white px-3 py-1 rounded text-sm hover:bg-red-700">Çıkış</a>
                </div>
            </div>
        </div>
    </header>

    <div class="container mx-auto px-4 py-8">
        <!-- Filters -->
        <div class="bg-white rounded-lg shadow-md p-6 mb-6">
            <div class="flex flex-wrap items-center gap-4">
                <input type="text" placeholder="Ürün ara..." 
                       class="px-4 py-2 border border-gray-300 rounded-lg flex-1 min-w-64">
                <select class="px-4 py-2 border border-gray-300 rounded-lg">
                    <option>Tüm Kategoriler</option>
                    <option>Elektronik</option>
                    <option>Giyim</option>
                    <option>Ev & Bahçe</option>
                </select>
                <select class="px-4 py-2 border border-gray-300 rounded-lg">
                    <option>Tüm Durumlar</option>
                    <option>Aktif</option>
                    <option>Pasif</option>
                    <option>Stokta Yok</option>
                </select>
                <button class="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700">
                    Filtrele
                </button>
            </div>
        </div>

        <!-- Products Table -->
        <div class="bg-white rounded-lg shadow-md overflow-hidden">
            <div class="px-6 py-4 border-b border-gray-200">
                <h3 class="text-lg font-semibold">Ürün Listesi</h3>
            </div>
            <div class="overflow-x-auto">
                <table class="w-full">
                    <thead class="bg-gray-50">
                        <tr>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Ürün</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Kategori</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Fiyat</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Stok</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Durum</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">İşlemler</th>
                        </tr>
                    </thead>
                    <tbody id="productsTableBody" class="bg-white divide-y divide-gray-200">
                        <!-- Products will be loaded here -->
                    </tbody>
                </table>
            </div>
        </div>
    </div>

    <script>
        // Load products
        fetch('/admin/api/products')
            .then(response => response.json())
            .then(data => {
                const tbody = document.getElementById('productsTableBody');
                tbody.innerHTML = '';
                
                data.products.forEach(product => {
                    const row = document.createElement('tr');
                    row.innerHTML = 
                        '<td class="px-6 py-4 whitespace-nowrap">' +
                            '<div class="flex items-center">' +
                                '<div class="flex-shrink-0 h-10 w-10 bg-gray-200 rounded-lg flex items-center justify-center">' +
                                    '<span class="text-gray-500">📱</span>' +
                                '</div>' +
                                '<div class="ml-4">' +
                                    '<div class="text-sm font-medium text-gray-900">' + product.name + '</div>' +
                                    '<div class="text-sm text-gray-500">' + product.sku + '</div>' +
                                '</div>' +
                            '</div>' +
                        '</td>' +
                        '<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">Kategori ' + product.category_id + '</td>' +
                        '<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">' + product.price.toFixed(2) + ' TL</td>' +
                        '<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">' + product.stock + '</td>' +
                        '<td class="px-6 py-4 whitespace-nowrap">' +
                            '<span class="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">' +
                                product.status +
                            '</span>' +
                        '</td>' +
                        '<td class="px-6 py-4 whitespace-nowrap text-sm font-medium space-x-2">' +
                            '<button class="text-blue-600 hover:text-blue-900">Düzenle</button>' +
                            '<button class="text-red-600 hover:text-red-900">Sil</button>' +
                        '</td>';
                    tbody.appendChild(row);
                });
            })
            .catch(error => {
                console.error('Error loading products:', error);
            });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func adminVendorsHandler(w http.ResponseWriter, r *http.Request, vendorService *services.VendorService) {
	if !isAdmin(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	html := `
<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Satıcı Yönetimi - KolajAI Admin</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <!-- Header -->
    <header class="bg-white shadow-sm border-b">
        <div class="container mx-auto px-4 py-4">
            <div class="flex justify-between items-center">
                <div class="flex items-center space-x-4">
                    <a href="/admin" class="text-blue-600 hover:text-blue-800">← Admin Panel</a>
                    <h1 class="text-2xl font-bold text-gray-800">Satıcı Yönetimi</h1>
                </div>
                <a href="/logout" class="bg-red-600 text-white px-3 py-1 rounded text-sm hover:bg-red-700">Çıkış</a>
            </div>
        </div>
    </header>

    <div class="container mx-auto px-4 py-8">
        <!-- Stats -->
        <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            <div class="bg-white rounded-lg shadow-md p-6 text-center">
                <div class="text-3xl font-bold text-green-600">1</div>
                <div class="text-sm text-gray-500">Aktif Satıcı</div>
            </div>
            <div class="bg-white rounded-lg shadow-md p-6 text-center">
                <div class="text-3xl font-bold text-yellow-600">0</div>
                <div class="text-sm text-gray-500">Onay Bekleyen</div>
            </div>
            <div class="bg-white rounded-lg shadow-md p-6 text-center">
                <div class="text-3xl font-bold text-red-600">0</div>
                <div class="text-sm text-gray-500">Askıda</div>
            </div>
            <div class="bg-white rounded-lg shadow-md p-6 text-center">
                <div class="text-3xl font-bold text-blue-600">3</div>
                <div class="text-sm text-gray-500">Toplam Ürün</div>
            </div>
        </div>

        <!-- Vendors List -->
        <div class="bg-white rounded-lg shadow-md">
            <div class="px-6 py-4 border-b border-gray-200">
                <h3 class="text-lg font-semibold">Satıcı Listesi</h3>
            </div>
            <div class="p-6">
                <div class="space-y-4">
                    <div class="border border-gray-200 rounded-lg p-4">
                        <div class="flex items-center justify-between">
                            <div class="flex items-center space-x-4">
                                <div class="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center">
                                    <span class="text-blue-600 font-bold">TC</span>
                                </div>
                                <div>
                                    <h4 class="font-semibold">Test Company</h4>
                                    <p class="text-sm text-gray-500">vendor@test.com</p>
                                    <p class="text-xs text-gray-400">İstanbul, Turkey</p>
                                </div>
                            </div>
                            <div class="flex items-center space-x-4">
                                <span class="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">
                                    Onaylandı
                                </span>
                                <div class="text-right">
                                    <div class="text-sm font-medium">3 Ürün</div>
                                    <div class="text-xs text-gray-500">Son aktivite: Bugün</div>
                                </div>
                                <div class="flex space-x-2">
                                    <button class="bg-blue-600 text-white px-3 py-1 rounded text-sm hover:bg-blue-700">
                                        Detay
                                    </button>
                                    <button class="bg-red-600 text-white px-3 py-1 rounded text-sm hover:bg-red-700">
                                        Askıya Al
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func adminApiProductsHandler(w http.ResponseWriter, r *http.Request, productService *services.ProductService) {
	if !isAdmin(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	products, err := productService.GetFeaturedProducts(50, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"products": products,
		"count":    len(products),
	})
}