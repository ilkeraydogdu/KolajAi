# Frontend Build Guide - KolajAI

Bu dokümanda, KolajAI projesinde frontend sayfalarında HTML kodlarının direk görünmesi sorununu çözen build sistemi açıklanmaktadır.

## Sorun

Proje frontend sayfalarında tasarım olarak gelmiyor, direk HTML kodları ekrana geliyor. Bu sorun şu nedenlerden kaynaklanıyordu:

1. **Webpack build edilmemişti** - CSS ve JS dosyaları compile edilmemişti
2. **Eksik bağımlılıklar** - NPM paketleri ve Babel pluginleri eksikti
3. **Yanlış asset referansları** - Template dosyalarında yanlış CSS/JS yolları
4. **Server konfigürasyonu** - Go server webpack build dosyalarını serve etmiyordu

## Çözüm

### 1. Bağımlılık Yönetimi

Tüm NPM bağımlılıkları yüklendi ve webpack konfigürasyonu düzeltildi:

```bash
npm install
npm install --save-dev @babel/plugin-transform-class-properties @babel/plugin-transform-optional-chaining @babel/plugin-transform-nullish-coalescing-operator
npm install web-vitals
npm install --save-dev webpack-manifest-plugin
```

### 2. Webpack Konfigürasyonu

`webpack.config.js` dosyasında şu düzeltmeler yapıldı:

- **Workbox plugin import** düzeltildi
- **Babel plugin isimleri** modern versiyonlarına güncellendi
- **Entry points** mevcut dosyalara göre düzenlendi
- **Manifest plugin** eklendi asset versioning için

### 3. SCSS Dosya Yolları

`web/static/sass/main.scss` dosyasında eksik olan background image dosyaları mevcut dosyalarla eşleştirildi:

- `error-bg.png` → `login1.png`
- `bg-login.png` → `login1.png`
- `bg-register.png` → `register1.png`
- `bg-forgot.png` → `forgot-password1.png`
- `bg-reset-password.png` → `reset-password1.png`

### 4. Build Sistemi

#### Build Komutu
```bash
npm run build
```

Bu komut şu dosyaları oluşturur:
- `dist/css/styles.[hash].css` - Tüm CSS kodları
- `dist/js/runtime.[hash].js` - Webpack runtime
- `dist/js/vendors.[hash].js` - Üçüncü parti kütüphaneler
- `dist/js/main.[hash].js` - Ana uygulama kodu
- `dist/manifest.json` - Asset mapping dosyası

#### Development Modu
```bash
npm run dev
```

Geliştirme için hot reload ile çalışır.

### 5. Server Konfigürasyonu

`cmd/server/main.go` dosyasında webpack build dosyalarını serve etmek için yeni route'lar eklendi:

```go
// Webpack built assets
appRouter.Handle("/static/css/", http.StripPrefix("/static/", http.FileServer(http.Dir("dist"))))
appRouter.Handle("/static/js/", http.StripPrefix("/static/", http.FileServer(http.Dir("dist"))))
appRouter.Handle("/static/images/", http.StripPrefix("/static/", http.FileServer(http.Dir("dist"))))
```

### 6. Template Entegrasyonu

`web/templates/layout/base.gohtml` dosyası güncellenerek webpack build dosyaları dahil edildi:

```html
<!-- Webpack Built CSS -->
{{if .Assets}}
  {{range .Assets.CSS}}
  <link href="{{.}}" rel="stylesheet">
  {{end}}
{{else}}
  <link href="/static/css/styles.d2c431ed.css" rel="stylesheet">
{{end}}

<!-- Webpack Built JS -->
{{if .Assets}}
  {{range .Assets.JS}}
  <script src="{{.}}" type="text/javascript"></script>
  {{end}}
{{else}}
  <!-- Fallback static references -->
{{end}}
```

### 7. Asset Management

`internal/utils/assets.go` dosyası oluşturularak asset yönetimi için utility fonksiyonları eklendi:

```go
// AssetManager kullanımı
assetManager := utils.NewAssetManager("dist/manifest.json")

// Template data'ya asset bilgilerini ekleme
templateData := map[string]interface{}{
    "Assets": map[string]interface{}{
        "CSS": assetManager.GetCSSAssets(),
        "JS":  assetManager.GetJSAssets(),
    },
}
```

## Kullanım

### 1. İlk Kurulum

```bash
# Bağımlılıkları yükle
npm install

# Assets'i build et
npm run build

# Go server'ı compile et
go build -o server cmd/server/main.go
```

### 2. Geliştirme

```bash
# Frontend geliştirme için
npm run dev

# Go server'ı ayrı terminalde çalıştır
./server
```

### 3. Production

```bash
# Production build
npm run build

# Server'ı çalıştır
./server
```

## Dosya Yapısı

```
project/
├── web/
│   ├── static/
│   │   ├── js/           # Kaynak JS dosyaları
│   │   ├── sass/         # Kaynak SCSS dosyaları
│   │   └── assets/       # Statik dosyalar (images, fonts, etc.)
│   └── templates/
│       └── layout/
│           └── base.gohtml
├── dist/                 # Webpack build çıktıları
│   ├── css/
│   ├── js/
│   ├── images/
│   └── manifest.json
├── internal/
│   └── utils/
│       └── assets.go     # Asset management utility
├── webpack.config.js
├── package.json
└── cmd/
    └── server/
        └── main.go
```

## Önemli Notlar

1. **Cache Busting**: Webpack otomatik olarak dosya isimlerine hash ekler, böylece browser cache problemi olmaz.

2. **Manifest Dosyası**: `dist/manifest.json` dosyası asset isimlerini hash'li versiyonlarıyla eşleştirir.

3. **Fallback**: Template'lerde asset yükleme başarısız olursa hardcoded yollar kullanılır.

4. **Hot Reload**: Development modunda değişiklikler otomatik olarak browser'da güncellenir.

5. **Production Optimizasyonu**: Build sırasında CSS/JS dosyaları minify edilir ve optimize edilir.

## Sorun Giderme

### Build Hataları

```bash
# Cache temizle
npm run clean

# Node modules'u yeniden yükle
rm -rf node_modules package-lock.json
npm install

# Build'i tekrar dene
npm run build
```

### Asset Yükleme Sorunları

1. `dist/manifest.json` dosyasının var olduğunu kontrol edin
2. Server'ın `/static/` route'larını doğru serve ettiğini kontrol edin
3. Template'lerde `.Assets` değişkeninin set edildiğini kontrol edin

### Server Sorunları

```bash
# Server loglarını kontrol et
./server 2>&1 | tee server.log

# Port'un açık olduğunu kontrol et
netstat -tlnp | grep :8081
```

## Sonuç

Bu çözümle birlikte:
- ✅ Frontend sayfaları artık düzgün stillendirilmiş olarak görünüyor
- ✅ CSS ve JS dosyaları optimize edilmiş şekilde yükleniyor
- ✅ Asset versioning sistemi çalışıyor
- ✅ Development ve production ortamları ayrı şekilde yapılandırılmış
- ✅ Otomatik build sistemi kurulmuş

Artık proje frontend sayfalarında HTML kodları değil, tasarlanmış arayüz görünecektir.