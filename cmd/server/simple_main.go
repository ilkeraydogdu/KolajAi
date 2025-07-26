package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"kolajAi/internal/database"
	"kolajAi/internal/services"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"kolajAi/internal/models"
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
	fmt.Println("ProductService created")

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

	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	fmt.Println("Server starting on :8082...")
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
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
    <div class="container mx-auto px-4 py-8">
        <div class="text-center mb-8">
            <h1 class="text-4xl font-bold text-blue-600 mb-4">KolajAI Marketplace</h1>
            <p class="text-gray-600 text-lg">Çoklu satıcı e-ticaret platformu</p>
        </div>
        
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-4xl mx-auto">
            <div class="bg-white p-6 rounded-lg shadow-md text-center">
                <h3 class="text-xl font-semibold mb-3">Ürünler</h3>
                <p class="text-gray-600 mb-4">Tüm ürünleri görüntüle</p>
                <a href="/products" class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">
                    Ürünleri Gör
                </a>
            </div>
            
            <div class="bg-white p-6 rounded-lg shadow-md text-center">
                <h3 class="text-xl font-semibold mb-3">API Test</h3>
                <p class="text-gray-600 mb-4">JSON API'yi test et</p>
                <a href="/api/products" class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">
                    API Test
                </a>
            </div>
            
            <div class="bg-white p-6 rounded-lg shadow-md text-center">
                <h3 class="text-xl font-semibold mb-3">Admin</h3>
                <p class="text-gray-600 mb-4">Yönetim paneli</p>
                <a href="/admin" class="bg-purple-600 text-white px-4 py-2 rounded hover:bg-purple-700">
                    Admin Panel
                </a>
            </div>
        </div>
    </div>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
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

// Simple in-memory cart
var cart = make(map[string]int) // productID -> quantity

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