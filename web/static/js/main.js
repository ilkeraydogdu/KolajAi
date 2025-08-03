// Main application entry point
// Import styles
import '../sass/main.scss';

import Alpine from 'alpinejs';
import axios from 'axios';
import { format, formatDistanceToNow } from 'date-fns';
import { tr } from 'date-fns/locale';

// Basic utilities (will create these if needed)
// import { debounce, throttle } from './utils/performance';
// import { showToast, showModal, showConfirm } from './utils/ui';
// import { formatCurrency, formatNumber } from './utils/formatters';
// import { validateEmail, validatePhone } from './utils/validators';

class KolajAIApp {
  constructor() {
    // Initialize services
    this.initializeServices();
    this.init();
  }

  initializeServices() {
    // Check if services are available globally first
    if (window.app && window.app.apiService) {
      this.apiService = window.app.apiService;
      this.authService = window.app.authService;
      this.cartService = window.app.cartService;
      this.notificationService = window.app.notificationService;
    } else {
      // Fallback: create basic service stubs
      console.warn('Services not found globally, creating stubs');
      this.apiService = {
        get: () => Promise.reject(new Error('API service not initialized')),
        post: () => Promise.reject(new Error('API service not initialized'))
      };
      this.authService = {
        getCurrentUser: () => Promise.reject(new Error('Auth service not initialized')),
        isLoggedIn: () => false
      };
      this.cartService = {
        getItemCount: () => 0,
        getTotal: () => 0
      };
      this.notificationService = {
        getUnreadCount: () => 0
      };
    }
  }

  async init() {
    try {
      // Initialize Alpine.js
      this.initAlpine();
      
      // Setup global axios configuration
      this.setupAxios();
      
      // Initialize services
      await this.initServices();
      
      // Setup event listeners
      this.setupEventListeners();
      
      // Initialize PWA features
      this.initPWA();
      
      // Setup performance monitoring
      this.setupPerformanceMonitoring();
      
      // Only log in development
      if (process.env.NODE_ENV === 'development') {
        window.logger && window.logger.debug('KolajAI App initialized successfully');
      }
      
    } catch (error) {
      // Always log errors but sanitize in production
      if (process.env.NODE_ENV === 'development') {
        console.error('Failed to initialize KolajAI App:', error);
      } else {
        console.error('Application initialization failed');
      }
      this.handleInitError(error);
    }
  }

  initAlpine() {
    // Global Alpine data
    Alpine.data('app', () => ({
      // Global state
      user: null,
      cart: {
        items: [],
        total: 0,
        count: 0
      },
      notifications: [],
      isLoading: false,
      
      // UI state
      mobileMenuOpen: false,
      searchOpen: false,
      cartOpen: false,
      notificationsOpen: false,
      
      // Search
      searchQuery: '',
      searchResults: [],
      searchLoading: false,
      
      // Filters
      filters: {
        category: '',
        priceRange: [0, 1000],
        rating: 0,
        availability: 'all'
      },
      
      // Methods
      async init() {
        await this.loadUser();
        await this.loadCart();
        await this.loadNotifications();
      },
      
      async loadUser() {
        try {
          const response = await window.app.authService.getCurrentUser();
          this.user = response.data;
        } catch (error) {
          window.logger && window.logger.debug('User not authenticated');
        }
      },
      
      async loadCart() {
        try {
          this.cart = await window.app.cartService.getCart();
        } catch (error) {
          console.error('Failed to load cart:', error);
        }
      },
      
      async loadNotifications() {
        if (!this.user) return;
        
        try {
          const response = await window.app.notificationService.getNotifications();
          this.notifications = response.data;
        } catch (error) {
          console.error('Failed to load notifications:', error);
        }
      },
      
      // Search functionality
      async search() {
        if (!this.searchQuery.trim()) {
          this.searchResults = [];
          return;
        }
        
        this.searchLoading = true;
        
        try {
          const response = await window.app.apiService.get('/api/search', {
            params: { q: this.searchQuery, ...this.filters }
          });
          this.searchResults = response.data.results;
        } catch (error) {
          console.error('Search failed:', error);
          showToast('Arama sırasında bir hata oluştu', 'error');
        } finally {
          this.searchLoading = false;
        }
      },
      
      // Cart functionality
      async addToCart(productId, quantity = 1) {
        try {
          await window.app.cartService.addItem(productId, quantity);
          await this.loadCart();
          showToast('Ürün sepete eklendi', 'success');
        } catch (error) {
          console.error('Failed to add to cart:', error);
          showToast('Ürün sepete eklenemedi', 'error');
        }
      },
      
      async removeFromCart(itemId) {
        try {
          await window.app.cartService.removeItem(itemId);
          await this.loadCart();
          showToast('Ürün sepetten kaldırıldı', 'info');
        } catch (error) {
          console.error('Failed to remove from cart:', error);
          showToast('Ürün sepetten kaldırılamadı', 'error');
        }
      },
      
      // Utility methods
      formatCurrency,
      formatNumber,
      formatDate: (date) => format(new Date(date), 'dd MMMM yyyy', { locale: tr }),
      formatTimeAgo: (date) => formatDistanceToNow(new Date(date), { 
        addSuffix: true, 
        locale: tr 
      }),
      
      // Navigation
      toggleMobileMenu() {
        this.mobileMenuOpen = !this.mobileMenuOpen;
      },
      
      toggleSearch() {
        this.searchOpen = !this.searchOpen;
        if (this.searchOpen) {
          this.$nextTick(() => {
            this.$refs.searchInput?.focus();
          });
        }
      },
      
      toggleCart() {
        this.cartOpen = !this.cartOpen;
      },
      
      toggleNotifications() {
        this.notificationsOpen = !this.notificationsOpen;
      }
    }));

    // Global Alpine stores
    Alpine.store('theme', {
      current: localStorage.getItem('theme') || 'light',
      toggle() {
        this.current = this.current === 'light' ? 'dark' : 'light';
        localStorage.setItem('theme', this.current);
        document.documentElement.classList.toggle('dark', this.current === 'dark');
      }
    });

    Alpine.store('preferences', {
      language: localStorage.getItem('language') || 'tr',
      currency: localStorage.getItem('currency') || 'TRY',
      notifications: JSON.parse(localStorage.getItem('notifications') || 'true'),
      
      setLanguage(lang) {
        this.language = lang;
        localStorage.setItem('language', lang);
        // Reload page to apply language changes
        window.location.reload();
      },
      
      setCurrency(currency) {
        this.currency = currency;
        localStorage.setItem('currency', currency);
      },
      
      toggleNotifications() {
        this.notifications = !this.notifications;
        localStorage.setItem('notifications', JSON.stringify(this.notifications));
      }
    });

    // Start Alpine
    Alpine.start();
    
    // Make Alpine globally available
    window.Alpine = Alpine;
  }

  setupAxios() {
    // Set default base URL
    axios.defaults.baseURL = window.location.origin;
    
    // Add CSRF token to all requests
    const csrfToken = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content');
    if (csrfToken) {
      axios.defaults.headers.common['X-CSRF-Token'] = csrfToken;
    }
    
    // Add auth token if available
    const token = localStorage.getItem('auth_token');
    if (token) {
      axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    }
    
    // Request interceptor
    axios.interceptors.request.use(
      (config) => {
        // Show loading indicator for non-background requests
        if (!config.background) {
          document.body.classList.add('loading');
        }
        return config;
      },
      (error) => {
        document.body.classList.remove('loading');
        return Promise.reject(error);
      }
    );
    
    // Response interceptor
    axios.interceptors.response.use(
      (response) => {
        document.body.classList.remove('loading');
        return response;
      },
      (error) => {
        document.body.classList.remove('loading');
        
        // Handle common errors
        if (error.response?.status === 401) {
          this.authService.logout();
          showToast('Oturum süreniz doldu, lütfen tekrar giriş yapın', 'warning');
        } else if (error.response?.status === 403) {
          showToast('Bu işlem için yetkiniz bulunmuyor', 'error');
        } else if (error.response?.status >= 500) {
          showToast('Sunucu hatası oluştu, lütfen daha sonra tekrar deneyin', 'error');
        }
        
        return Promise.reject(error);
      }
    );
  }

  async initServices() {
    // Initialize all services
    await Promise.all([
      this.authService.init(),
      this.cartService.init(),
      this.notificationService.init()
    ]);
  }

  setupEventListeners() {
    // Handle online/offline status
    window.addEventListener('online', () => {
      showToast('İnternet bağlantısı yeniden kuruldu', 'success');
      this.syncOfflineData();
    });
    
    window.addEventListener('offline', () => {
      showToast('İnternet bağlantısı kesildi', 'warning');
    });
    
    // Handle visibility change
    document.addEventListener('visibilitychange', () => {
      if (!document.hidden) {
        // Refresh data when page becomes visible
        this.refreshData();
      }
    });
    
    // Handle beforeunload
    window.addEventListener('beforeunload', (e) => {
      // Save any pending data
      this.savePendingData();
    });
    
    // Keyboard shortcuts
    document.addEventListener('keydown', (e) => {
      // Ctrl/Cmd + K for search
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault();
        Alpine.store('app').toggleSearch();
      }
      
      // Escape to close modals
      if (e.key === 'Escape') {
        // Close any open modals
        document.querySelectorAll('[x-show]').forEach(el => {
          if (el._x_dataStack?.[0]?.open) {
            el._x_dataStack[0].open = false;
          }
        });
      }
    });
  }

  initPWA() {
    // Register service worker
    if ('serviceWorker' in navigator) {
      window.addEventListener('load', async () => {
        try {
          const registration = await navigator.serviceWorker.register('/sw.js');
          window.logger && window.logger.debug('SW registered: ', registration);
          
          // Handle updates
          registration.addEventListener('updatefound', () => {
            const newWorker = registration.installing;
            newWorker.addEventListener('statechange', () => {
              if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
                // Show update available notification
                this.showUpdateAvailable();
              }
            });
          });
        } catch (error) {
          window.logger && window.logger.debug('SW registration failed: ', error);
        }
      });
    }
    
    // Handle app installation
    let deferredPrompt;
    window.addEventListener('beforeinstallprompt', (e) => {
      e.preventDefault();
      deferredPrompt = e;
      this.showInstallPrompt(deferredPrompt);
    });
    
    // Handle app installed
    window.addEventListener('appinstalled', () => {
      showToast('Uygulama başarıyla yüklendi!', 'success');
      deferredPrompt = null;
    });
  }

  setupPerformanceMonitoring() {
    // Monitor Core Web Vitals
    if ('web-vital' in window) {
      import('web-vitals').then(({ getCLS, getFID, getFCP, getLCP, getTTFB }) => {
        getCLS(console.log);
        getFID(console.log);
        getFCP(console.log);
        getLCP(console.log);
        getTTFB(console.log);
      });
    }
    
    // Monitor long tasks
    if ('PerformanceObserver' in window) {
      try {
        const observer = new PerformanceObserver((list) => {
          for (const entry of list.getEntries()) {
            if (entry.duration > 50) {
              console.warn('Long task detected:', entry);
            }
          }
        });
        observer.observe({ entryTypes: ['longtask'] });
      } catch (e) {
        // PerformanceObserver not supported
      }
    }
  }

  handleInitError(error) {
    // Show user-friendly error message
    const errorDiv = document.createElement('div');
    errorDiv.className = 'fixed inset-0 bg-red-50 flex items-center justify-center z-50';
    // Create safe error structure to prevent XSS
    const containerDiv = document.createElement('div');
    containerDiv.className = 'text-center p-8';
    
    const iconDiv = document.createElement('div');
    iconDiv.className = 'text-red-600 text-6xl mb-4';
    iconDiv.textContent = '⚠️';
    
    const titleH1 = document.createElement('h1');
    titleH1.className = 'text-2xl font-bold text-red-800 mb-2';
    titleH1.textContent = 'Uygulama Başlatılamadı';
    
    const messageP = document.createElement('p');
    messageP.className = 'text-red-600 mb-4';
    messageP.textContent = 'Bir hata oluştu. Lütfen sayfayı yenileyin.';
    
    const reloadButton = document.createElement('button');
    reloadButton.className = 'bg-red-600 text-white px-6 py-2 rounded-lg hover:bg-red-700';
    reloadButton.textContent = 'Sayfayı Yenile';
    reloadButton.addEventListener('click', () => window.location.reload());
    
    containerDiv.appendChild(iconDiv);
    containerDiv.appendChild(titleH1);
    containerDiv.appendChild(messageP);
    containerDiv.appendChild(reloadButton);
    errorDiv.appendChild(containerDiv);
    document.body.appendChild(errorDiv);
  }

  async syncOfflineData() {
    // Sync any offline data when connection is restored
    try {
      await this.cartService.syncOfflineData();
      await this.notificationService.syncOfflineData();
    } catch (error) {
      console.error('Failed to sync offline data:', error);
    }
  }

  async refreshData() {
    // Refresh data when page becomes visible
    try {
      const app = Alpine.store('app');
      if (app) {
        await Promise.all([
          app.loadCart(),
          app.loadNotifications()
        ]);
      }
    } catch (error) {
      console.error('Failed to refresh data:', error);
    }
  }

  savePendingData() {
    // Save any pending data before page unload
    try {
      this.cartService.savePendingData();
    } catch (error) {
      console.error('Failed to save pending data:', error);
    }
  }

  showUpdateAvailable() {
    const updateBanner = document.createElement('div');
    updateBanner.className = 'fixed top-0 left-0 right-0 bg-blue-600 text-white p-4 z-50';
    updateBanner.innerHTML = `
      <div class="flex items-center justify-between max-w-7xl mx-auto">
        <span>Yeni bir sürüm mevcut!</span>
        <button onclick="window.location.reload()" 
                class="bg-blue-700 px-4 py-2 rounded hover:bg-blue-800">
          Güncelle
        </button>
      </div>
    `;
    document.body.appendChild(updateBanner);
  }

  showInstallPrompt(deferredPrompt) {
    // Show install prompt after some user interaction
    setTimeout(() => {
      const installBanner = document.createElement('div');
      installBanner.className = 'fixed bottom-4 right-4 bg-white shadow-lg rounded-lg p-4 max-w-sm z-50';
      installBanner.innerHTML = `
        <div class="flex items-center space-x-3">
          <div class="flex-1">
            <h3 class="font-semibold">Uygulamayı Yükle</h3>
            <p class="text-sm text-gray-600">Daha iyi deneyim için uygulamayı yükleyin</p>
          </div>
          <div class="flex space-x-2">
            <button onclick="this.parentElement.parentElement.parentElement.remove()" 
                    class="text-gray-400 hover:text-gray-600">×</button>
            <button onclick="window.app.installApp()" 
                    class="bg-blue-600 text-white px-3 py-1 rounded text-sm hover:bg-blue-700">
              Yükle
            </button>
          </div>
        </div>
      `;
      document.body.appendChild(installBanner);
    }, 5000);
  }

  async installApp() {
    if (this.deferredPrompt) {
      this.deferredPrompt.prompt();
      const { outcome } = await this.deferredPrompt.userChoice;
      window.logger && window.logger.debug(`User response to the install prompt: ${outcome}`);
      this.deferredPrompt = null;
    }
  }
}

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  window.app = new KolajAIApp();
});

// Export for use in other modules
export default KolajAIApp;