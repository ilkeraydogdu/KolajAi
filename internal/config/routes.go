package config

import "net/http"

// RouteConfig represents a route configuration
type RouteConfig struct {
	Path     string        `json:"path"`
	Handler  string        `json:"handler"`
	Methods  []string      `json:"methods"`
	Template string        `json:"template"`
	Layout   string        `json:"layout,omitempty"`
	SEOMeta  SEOMetaConfig `json:"seo_meta"`
}

// SEOMetaConfig represents SEO metadata for a route
type SEOMetaConfig struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords,omitempty"`
	Robots      string   `json:"robots,omitempty"`
}

// GetAuthRoutes returns the authentication routes configuration
func GetAuthRoutes() []RouteConfig {
	return []RouteConfig{
		{
			Path:     "/login",
			Handler:  "Login",
			Methods:  []string{http.MethodGet, http.MethodPost},
			Template: "auth/login",
			Layout:   "auth",
			SEOMeta: SEOMetaConfig{
				Title:       "Giriş Yap",
				Description: "Hesabınıza giriş yapın",
				Keywords:    []string{"giriş", "login", "hesap", "kullanıcı"},
				Robots:      "noindex, nofollow",
			},
		},
		{
			Path:     "/register",
			Handler:  "Register",
			Methods:  []string{http.MethodGet, http.MethodPost},
			Template: "auth/register",
			Layout:   "auth",
			SEOMeta: SEOMetaConfig{
				Title:       "Kayıt Ol",
				Description: "Yeni bir hesap oluşturun",
				Keywords:    []string{"kayıt", "register", "hesap", "kullanıcı"},
				Robots:      "noindex, nofollow",
			},
		},
		{
			Path:     "/forgot-password",
			Handler:  "ForgotPassword",
			Methods:  []string{http.MethodGet, http.MethodPost},
			Template: "auth/forgot-password",
			Layout:   "auth",
			SEOMeta: SEOMetaConfig{
				Title:       "Şifremi Unuttum",
				Description: "Şifrenizi sıfırlayın",
				Keywords:    []string{"şifre", "unuttum", "sıfırlama"},
				Robots:      "noindex, nofollow",
			},
		},
		{
			Path:     "/reset-password",
			Handler:  "ResetPassword",
			Methods:  []string{http.MethodGet, http.MethodPost},
			Template: "auth/reset-password",
			Layout:   "auth",
			SEOMeta: SEOMetaConfig{
				Title:       "Şifre Sıfırlama",
				Description: "Yeni şifrenizi belirleyin",
				Keywords:    []string{"şifre", "sıfırlama", "yeni şifre"},
				Robots:      "noindex, nofollow",
			},
		},
	}
}

// GetMainRoutes returns the main routes configuration
func GetMainRoutes() []RouteConfig {
	return []RouteConfig{
		{
			Path:     "/",
			Handler:  "Index",
			Methods:  []string{http.MethodGet},
			Template: "index",
			Layout:   "main",
			SEOMeta: SEOMetaConfig{
				Title:       "Ana Sayfa",
				Description: "KolajAI'ya Hoş Geldiniz",
				Keywords:    []string{"anasayfa", "kolajAI", "yapay zeka"},
			},
		},
		{
			Path:     "/dashboard",
			Handler:  "Dashboard",
			Methods:  []string{http.MethodGet},
			Template: "dashboard",
			Layout:   "main",
			SEOMeta: SEOMetaConfig{
				Title:       "Kontrol Paneli",
				Description: "Kullanıcı kontrol paneli",
				Keywords:    []string{"dashboard", "panel", "kontrol"},
			},
		},
		{
			Path:     "/settings",
			Handler:  "Settings",
			Methods:  []string{http.MethodGet, http.MethodPost},
			Template: "settings",
			Layout:   "main",
			SEOMeta: SEOMetaConfig{
				Title:       "Ayarlar",
				Description: "Hesap ayarlarınızı yönetin",
				Keywords:    []string{"ayarlar", "hesap", "profil"},
			},
		},
		{
			Path:     "/components",
			Handler:  "ComponentsExample",
			Methods:  []string{http.MethodGet},
			Template: "components/example",
			Layout:   "main",
			SEOMeta: SEOMetaConfig{
				Title:       "Bileşen Örnekleri",
				Description: "Yeniden kullanılabilir UI bileşenleri örnekleri",
				Keywords:    []string{"bileşenler", "components", "UI", "arayüz"},
			},
		},
	}
}
