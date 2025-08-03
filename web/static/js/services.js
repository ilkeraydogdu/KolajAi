/**
 * Frontend Services for KolajAI
 * Bu dosya frontend'de kullanılan service sınıflarını içerir
 */

// Base API Service
class ApiService {
  constructor(baseURL = '') {
    this.baseURL = baseURL || window.location.origin;
    this.defaultHeaders = {
      'Content-Type': 'application/json',
      'X-Requested-With': 'XMLHttpRequest'
    };
  }

  async request(method, endpoint, data = null, options = {}) {
    const url = `${this.baseURL}${endpoint}`;
    const config = {
      method: method.toUpperCase(),
      headers: { ...this.defaultHeaders, ...options.headers },
      credentials: 'same-origin',
      ...options
    };

    // Add CSRF token if available
    const csrfToken = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content');
    if (csrfToken) {
      config.headers['X-CSRF-Token'] = csrfToken;
    }

    // Add auth token if available
    const authToken = localStorage.getItem('auth_token');
    if (authToken) {
      config.headers['Authorization'] = `Bearer ${authToken}`;
    }

    if (data && ['POST', 'PUT', 'PATCH'].includes(config.method)) {
      config.body = JSON.stringify(data);
    }

    try {
      const response = await fetch(url, config);
      
      // Handle different response types
      const contentType = response.headers.get('content-type');
      let responseData;
      
      if (contentType && contentType.includes('application/json')) {
        responseData = await response.json();
      } else {
        responseData = await response.text();
      }

      if (!response.ok) {
        throw new Error(responseData.message || `HTTP ${response.status}: ${response.statusText}`);
      }

      return {
        data: responseData,
        status: response.status,
        headers: response.headers
      };
    } catch (error) {
      console.error('API Request failed:', error);
      throw error;
    }
  }

  async get(endpoint, options = {}) {
    return this.request('GET', endpoint, null, options);
  }

  async post(endpoint, data, options = {}) {
    return this.request('POST', endpoint, data, options);
  }

  async put(endpoint, data, options = {}) {
    return this.request('PUT', endpoint, data, options);
  }

  async patch(endpoint, data, options = {}) {
    return this.request('PATCH', endpoint, data, options);
  }

  async delete(endpoint, options = {}) {
    return this.request('DELETE', endpoint, null, options);
  }
}

// Authentication Service
class AuthService {
  constructor() {
    this.apiService = new ApiService();
    this.currentUser = null;
    this.isAuthenticated = false;
  }

  async init() {
    // Check if user is already authenticated
    const token = localStorage.getItem('auth_token');
    if (token) {
      try {
        await this.getCurrentUser();
      } catch (error) {
        // Token might be expired, remove it
        this.logout();
      }
    }
  }

  async login(email, password) {
    try {
      const response = await this.apiService.post('/api/auth/login', {
        email,
        password
      });

      if (response.data.token) {
        localStorage.setItem('auth_token', response.data.token);
        this.currentUser = response.data.user;
        this.isAuthenticated = true;
        
        // Trigger auth state change event
        window.dispatchEvent(new CustomEvent('authStateChanged', { 
          detail: { authenticated: true, user: this.currentUser } 
        }));
      }

      return response;
    } catch (error) {
      console.error('Login failed:', error);
      throw error;
    }
  }

  async register(userData) {
    try {
      const response = await this.apiService.post('/api/auth/register', userData);
      return response;
    } catch (error) {
      console.error('Registration failed:', error);
      throw error;
    }
  }

  async getCurrentUser() {
    try {
      const response = await this.apiService.get('/api/auth/me');
      this.currentUser = response.data;
      this.isAuthenticated = true;
      return response;
    } catch (error) {
      this.currentUser = null;
      this.isAuthenticated = false;
      throw error;
    }
  }

  async logout() {
    try {
      await this.apiService.post('/api/auth/logout');
    } catch (error) {
      console.error('Logout request failed:', error);
    } finally {
      // Clear local state regardless of API call result
      localStorage.removeItem('auth_token');
      this.currentUser = null;
      this.isAuthenticated = false;
      
      // Trigger auth state change event
      window.dispatchEvent(new CustomEvent('authStateChanged', { 
        detail: { authenticated: false, user: null } 
      }));
      
      // Redirect to login page
      window.location.href = '/login';
    }
  }

  async forgotPassword(email) {
    try {
      const response = await this.apiService.post('/api/auth/forgot-password', { email });
      return response;
    } catch (error) {
      console.error('Forgot password failed:', error);
      throw error;
    }
  }

  async resetPassword(token, password, passwordConfirm) {
    try {
      const response = await this.apiService.post('/api/auth/reset-password', {
        token,
        password,
        password_confirm: passwordConfirm
      });
      return response;
    } catch (error) {
      console.error('Reset password failed:', error);
      throw error;
    }
  }

  async updateProfile(userData) {
    try {
      const response = await this.apiService.put('/api/auth/profile', userData);
      this.currentUser = response.data;
      return response;
    } catch (error) {
      console.error('Profile update failed:', error);
      throw error;
    }
  }

  async changePassword(currentPassword, newPassword) {
    try {
      const response = await this.apiService.post('/api/auth/change-password', {
        current_password: currentPassword,
        new_password: newPassword
      });
      return response;
    } catch (error) {
      console.error('Password change failed:', error);
      throw error;
    }
  }

  getUser() {
    return this.currentUser;
  }

  isLoggedIn() {
    return this.isAuthenticated;
  }
}

// Cart Service
class CartService {
  constructor() {
    this.apiService = new ApiService();
    this.cart = {
      items: [],
      total: 0,
      count: 0,
      subtotal: 0,
      tax: 0,
      shipping: 0
    };
    this.storageKey = 'kolajAI_cart';
  }

  async init() {
    // Load cart from localStorage first (for offline support)
    this.loadFromStorage();
    
    // Then sync with server if user is authenticated
    if (window.app?.authService?.isLoggedIn()) {
      try {
        await this.syncWithServer();
      } catch (error) {
        console.warn('Cart sync failed:', error);
      }
    }
  }

  async getCart() {
    try {
      const response = await this.apiService.get('/api/cart');
      this.updateCart(response.data);
      return this.cart;
    } catch (error) {
      console.error('Get cart failed:', error);
      return this.cart;
    }
  }

  async addItem(productId, quantity = 1, options = {}) {
    try {
      const response = await this.apiService.post('/api/cart/add', {
        product_id: productId,
        quantity,
        ...options
      });

      this.updateCart(response.data);
      this.saveToStorage();
      this.triggerCartUpdate();
      
      return response;
    } catch (error) {
      console.error('Add to cart failed:', error);
      // Fallback to local storage for offline support
      this.addItemLocally(productId, quantity, options);
      throw error;
    }
  }

  async updateItem(itemId, quantity) {
    try {
      const response = await this.apiService.put(`/api/cart/item/${itemId}`, {
        quantity
      });

      this.updateCart(response.data);
      this.saveToStorage();
      this.triggerCartUpdate();
      
      return response;
    } catch (error) {
      console.error('Update cart item failed:', error);
      throw error;
    }
  }

  async removeItem(itemId) {
    try {
      const response = await this.apiService.delete(`/api/cart/item/${itemId}`);
      
      this.updateCart(response.data);
      this.saveToStorage();
      this.triggerCartUpdate();
      
      return response;
    } catch (error) {
      console.error('Remove cart item failed:', error);
      throw error;
    }
  }

  async clearCart() {
    try {
      const response = await this.apiService.delete('/api/cart');
      
      this.cart = {
        items: [],
        total: 0,
        count: 0,
        subtotal: 0,
        tax: 0,
        shipping: 0
      };
      
      this.saveToStorage();
      this.triggerCartUpdate();
      
      return response;
    } catch (error) {
      console.error('Clear cart failed:', error);
      throw error;
    }
  }

  async applyCoupon(couponCode) {
    try {
      const response = await this.apiService.post('/api/cart/coupon', {
        coupon_code: couponCode
      });

      this.updateCart(response.data);
      this.saveToStorage();
      this.triggerCartUpdate();
      
      return response;
    } catch (error) {
      console.error('Apply coupon failed:', error);
      throw error;
    }
  }

  async removeCoupon() {
    try {
      const response = await this.apiService.delete('/api/cart/coupon');
      
      this.updateCart(response.data);
      this.saveToStorage();
      this.triggerCartUpdate();
      
      return response;
    } catch (error) {
      console.error('Remove coupon failed:', error);
      throw error;
    }
  }

  addItemLocally(productId, quantity, options) {
    // Find existing item
    const existingItem = this.cart.items.find(item => item.product_id === productId);
    
    if (existingItem) {
      existingItem.quantity += quantity;
    } else {
      this.cart.items.push({
        id: Date.now(), // Temporary ID
        product_id: productId,
        quantity,
        ...options
      });
    }
    
    this.calculateTotals();
    this.saveToStorage();
    this.triggerCartUpdate();
  }

  updateCart(cartData) {
    this.cart = {
      ...this.cart,
      ...cartData
    };
    this.calculateTotals();
  }

  calculateTotals() {
    this.cart.count = this.cart.items.reduce((sum, item) => sum + item.quantity, 0);
    this.cart.subtotal = this.cart.items.reduce((sum, item) => sum + (item.price * item.quantity), 0);
    this.cart.total = this.cart.subtotal + this.cart.tax + this.cart.shipping;
  }

  saveToStorage() {
    try {
      localStorage.setItem(this.storageKey, JSON.stringify(this.cart));
    } catch (error) {
      console.error('Failed to save cart to storage:', error);
    }
  }

  loadFromStorage() {
    try {
      const stored = localStorage.getItem(this.storageKey);
      if (stored) {
        this.cart = { ...this.cart, ...JSON.parse(stored) };
      }
    } catch (error) {
      console.error('Failed to load cart from storage:', error);
    }
  }

  async syncWithServer() {
    // Sync local cart with server
    if (this.cart.items.length > 0) {
      for (const item of this.cart.items) {
        if (!item.synced) {
          try {
            await this.addItem(item.product_id, item.quantity, item.options);
            item.synced = true;
          } catch (error) {
            console.warn('Failed to sync cart item:', error);
          }
        }
      }
    }
    
    // Get latest cart from server
    await this.getCart();
  }

  async savePendingData() {
    // Save any pending changes before page unload
    this.saveToStorage();
  }

  async syncOfflineData() {
    // Sync offline changes when connection is restored
    await this.syncWithServer();
  }

  triggerCartUpdate() {
    // Trigger cart update event
    window.dispatchEvent(new CustomEvent('cartUpdated', { 
      detail: this.cart 
    }));
  }

  getItemCount() {
    return this.cart.count;
  }

  getTotal() {
    return this.cart.total;
  }

  getItems() {
    return this.cart.items;
  }
}

// Notification Service
class NotificationService {
  constructor() {
    this.apiService = new ApiService();
    this.notifications = [];
    this.unreadCount = 0;
    this.pollInterval = null;
    this.pollFrequency = 30000; // 30 seconds
  }

  async init() {
    // Load initial notifications
    await this.getNotifications();
    
    // Start polling for new notifications if user is authenticated
    if (window.app?.authService?.isLoggedIn()) {
      this.startPolling();
    }

    // Listen for auth state changes
    window.addEventListener('authStateChanged', (event) => {
      if (event.detail.authenticated) {
        this.startPolling();
      } else {
        this.stopPolling();
        this.notifications = [];
        this.unreadCount = 0;
      }
    });
  }

  async getNotifications(page = 1, limit = 20) {
    try {
      const response = await this.apiService.get(`/api/notifications?page=${page}&limit=${limit}`);
      
      if (page === 1) {
        this.notifications = response.data.notifications || [];
      } else {
        this.notifications = [...this.notifications, ...(response.data.notifications || [])];
      }
      
      this.unreadCount = response.data.unread_count || 0;
      this.triggerNotificationUpdate();
      
      return response;
    } catch (error) {
      console.error('Get notifications failed:', error);
      return { data: { notifications: [], unread_count: 0 } };
    }
  }

  async markAsRead(notificationId) {
    try {
      const response = await this.apiService.post(`/api/notifications/${notificationId}/read`);
      
      // Update local notification
      const notification = this.notifications.find(n => n.id === notificationId);
      if (notification) {
        notification.read_at = new Date().toISOString();
        this.unreadCount = Math.max(0, this.unreadCount - 1);
        this.triggerNotificationUpdate();
      }
      
      return response;
    } catch (error) {
      console.error('Mark notification as read failed:', error);
      throw error;
    }
  }

  async markAllAsRead() {
    try {
      const response = await this.apiService.post('/api/notifications/read-all');
      
      // Update local notifications
      this.notifications.forEach(notification => {
        if (!notification.read_at) {
          notification.read_at = new Date().toISOString();
        }
      });
      
      this.unreadCount = 0;
      this.triggerNotificationUpdate();
      
      return response;
    } catch (error) {
      console.error('Mark all notifications as read failed:', error);
      throw error;
    }
  }

  async deleteNotification(notificationId) {
    try {
      const response = await this.apiService.delete(`/api/notifications/${notificationId}`);
      
      // Remove from local notifications
      const index = this.notifications.findIndex(n => n.id === notificationId);
      if (index !== -1) {
        const notification = this.notifications[index];
        if (!notification.read_at) {
          this.unreadCount = Math.max(0, this.unreadCount - 1);
        }
        this.notifications.splice(index, 1);
        this.triggerNotificationUpdate();
      }
      
      return response;
    } catch (error) {
      console.error('Delete notification failed:', error);
      throw error;
    }
  }

  async sendNotification(userId, title, message, type = 'info', data = {}) {
    try {
      const response = await this.apiService.post('/api/notifications/send', {
        user_id: userId,
        title,
        message,
        type,
        data
      });
      
      return response;
    } catch (error) {
      console.error('Send notification failed:', error);
      throw error;
    }
  }

  startPolling() {
    if (this.pollInterval) {
      clearInterval(this.pollInterval);
    }
    
    this.pollInterval = setInterval(async () => {
      try {
        await this.getNotifications();
      } catch (error) {
        console.warn('Notification polling failed:', error);
      }
    }, this.pollFrequency);
  }

  stopPolling() {
    if (this.pollInterval) {
      clearInterval(this.pollInterval);
      this.pollInterval = null;
    }
  }

  async syncOfflineData() {
    // Sync any offline notification actions when connection is restored
    await this.getNotifications();
  }

  triggerNotificationUpdate() {
    // Trigger notification update event
    window.dispatchEvent(new CustomEvent('notificationsUpdated', { 
      detail: { 
        notifications: this.notifications, 
        unreadCount: this.unreadCount 
      } 
    }));
  }

  getNotifications() {
    return this.notifications;
  }

  getUnreadCount() {
    return this.unreadCount;
  }

  // Browser notification support
  async requestPermission() {
    if ('Notification' in window) {
      const permission = await Notification.requestPermission();
      return permission === 'granted';
    }
    return false;
  }

  showBrowserNotification(title, options = {}) {
    if ('Notification' in window && Notification.permission === 'granted') {
      const notification = new Notification(title, {
        icon: '/static/assets/images/icon-192x192.png',
        badge: '/static/assets/images/icon-72x72.png',
        ...options
      });
      
      // Auto close after 5 seconds
      setTimeout(() => notification.close(), 5000);
      
      return notification;
    }
  }
}

// Initialize services when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  // Make services globally available
  window.app = window.app || {};
  window.app.apiService = new ApiService();
  window.app.authService = new AuthService();
  window.app.cartService = new CartService();
  window.app.notificationService = new NotificationService();
});

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    ApiService,
    AuthService,
    CartService,
    NotificationService
  };
}