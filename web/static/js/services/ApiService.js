import axios from 'axios';

class ApiService {
  constructor() {
    this.baseURL = window.location.origin;
    this.timeout = 10000;
    this.retryAttempts = 3;
    this.retryDelay = 1000;
    
    this.setupInterceptors();
  }

  setupInterceptors() {
    // Request interceptor
    axios.interceptors.request.use(
      (config) => {
        // Add timestamp to prevent caching
        if (config.method === 'get') {
          config.params = {
            ...config.params,
            _t: Date.now()
          };
        }
        
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Response interceptor with retry logic
    axios.interceptors.response.use(
      (response) => response,
      async (error) => {
        const config = error.config;
        
        // Retry logic for network errors
        if (
          error.code === 'NETWORK_ERROR' ||
          error.code === 'ECONNABORTED' ||
          (error.response && error.response.status >= 500)
        ) {
          config.__retryCount = config.__retryCount || 0;
          
          if (config.__retryCount < this.retryAttempts) {
            config.__retryCount++;
            
            // Wait before retry
            await this.delay(this.retryDelay * config.__retryCount);
            
            return axios(config);
          }
        }
        
        return Promise.reject(error);
      }
    );
  }

  async delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  // GET request
  async get(url, config = {}) {
    try {
      const response = await axios.get(url, {
        timeout: this.timeout,
        ...config
      });
      return response;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  // POST request
  async post(url, data = {}, config = {}) {
    try {
      const response = await axios.post(url, data, {
        timeout: this.timeout,
        ...config
      });
      return response;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  // PUT request
  async put(url, data = {}, config = {}) {
    try {
      const response = await axios.put(url, data, {
        timeout: this.timeout,
        ...config
      });
      return response;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  // PATCH request
  async patch(url, data = {}, config = {}) {
    try {
      const response = await axios.patch(url, data, {
        timeout: this.timeout,
        ...config
      });
      return response;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  // DELETE request
  async delete(url, config = {}) {
    try {
      const response = await axios.delete(url, {
        timeout: this.timeout,
        ...config
      });
      return response;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  // Upload file
  async upload(url, file, options = {}) {
    const formData = new FormData();
    formData.append('file', file);
    
    // Add additional fields
    if (options.fields) {
      Object.keys(options.fields).forEach(key => {
        formData.append(key, options.fields[key]);
      });
    }

    try {
      const response = await axios.post(url, formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        },
        timeout: 60000, // Longer timeout for uploads
        onUploadProgress: options.onProgress || null,
        ...options.config
      });
      return response;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  // Download file
  async download(url, filename, config = {}) {
    try {
      const response = await axios.get(url, {
        responseType: 'blob',
        timeout: 60000, // Longer timeout for downloads
        ...config
      });

      // Create download link
      const blob = new Blob([response.data]);
      const downloadUrl = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = downloadUrl;
      link.download = filename || 'download';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(downloadUrl);

      return response;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  // Batch requests
  async batch(requests) {
    try {
      const promises = requests.map(request => {
        const { method, url, data, config } = request;
        return this[method.toLowerCase()](url, data, config);
      });
      
      const responses = await Promise.allSettled(promises);
      return responses;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  // Paginated GET request
  async paginate(url, options = {}) {
    const {
      page = 1,
      limit = 20,
      params = {},
      ...config
    } = options;

    try {
      const response = await this.get(url, {
        params: {
          page,
          limit,
          ...params
        },
        ...config
      });

      return {
        data: response.data.data || response.data,
        pagination: {
          page: response.data.page || page,
          limit: response.data.limit || limit,
          total: response.data.total || 0,
          totalPages: response.data.totalPages || Math.ceil((response.data.total || 0) / limit),
          hasNext: response.data.hasNext || false,
          hasPrev: response.data.hasPrev || false
        }
      };
    } catch (error) {
      throw this.handleError(error);
    }
  }

  // Search with debouncing
  async search(url, query, options = {}) {
    const {
      debounceMs = 300,
      minLength = 2,
      params = {},
      ...config
    } = options;

    // Clear previous search timeout
    if (this.searchTimeout) {
      clearTimeout(this.searchTimeout);
    }

    return new Promise((resolve, reject) => {
      this.searchTimeout = setTimeout(async () => {
        if (query.length < minLength) {
          resolve({ data: [] });
          return;
        }

        try {
          const response = await this.get(url, {
            params: {
              q: query,
              ...params
            },
            ...config
          });
          resolve(response);
        } catch (error) {
          reject(this.handleError(error));
        }
      }, debounceMs);
    });
  }

  // Cache management
  setupCache(options = {}) {
    const {
      maxAge = 5 * 60 * 1000, // 5 minutes
      maxEntries = 100
    } = options;

    this.cache = new Map();
    this.cacheOptions = { maxAge, maxEntries };
  }

  getCacheKey(config) {
    return `${config.method}_${config.url}_${JSON.stringify(config.params || {})}`;
  }

  getFromCache(key) {
    if (!this.cache) return null;
    
    const cached = this.cache.get(key);
    if (!cached) return null;
    
    if (Date.now() - cached.timestamp > this.cacheOptions.maxAge) {
      this.cache.delete(key);
      return null;
    }
    
    return cached.data;
  }

  setCache(key, data) {
    if (!this.cache) return;
    
    // Remove oldest entries if cache is full
    if (this.cache.size >= this.cacheOptions.maxEntries) {
      const firstKey = this.cache.keys().next().value;
      this.cache.delete(firstKey);
    }
    
    this.cache.set(key, {
      data,
      timestamp: Date.now()
    });
  }

  // Request with caching
  async getWithCache(url, config = {}) {
    const cacheKey = this.getCacheKey({ method: 'GET', url, ...config });
    const cached = this.getFromCache(cacheKey);
    
    if (cached) {
      return cached;
    }
    
    const response = await this.get(url, config);
    this.setCache(cacheKey, response);
    
    return response;
  }

  // Clear cache
  clearCache() {
    if (this.cache) {
      this.cache.clear();
    }
  }

  // Health check
  async healthCheck() {
    try {
      const response = await this.get('/api/health', {
        timeout: 5000,
        background: true
      });
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  // Error handling
  handleError(error) {
    const errorInfo = {
      message: 'An error occurred',
      status: null,
      code: null,
      data: null,
      timestamp: new Date().toISOString()
    };

    if (error.response) {
      // Server responded with error status
      errorInfo.status = error.response.status;
      errorInfo.data = error.response.data;
      errorInfo.message = error.response.data?.message || `HTTP ${error.response.status}`;
    } else if (error.request) {
      // Request was made but no response received
      errorInfo.code = 'NETWORK_ERROR';
      errorInfo.message = 'Network error - please check your connection';
    } else {
      // Something happened in setting up the request
      errorInfo.message = error.message;
    }

    // Log error for debugging
    console.error('API Error:', errorInfo);

    return errorInfo;
  }

  // Set auth token
  setAuthToken(token) {
    if (token) {
      axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
      localStorage.setItem('auth_token', token);
    } else {
      delete axios.defaults.headers.common['Authorization'];
      localStorage.removeItem('auth_token');
    }
  }

  // Get auth token
  getAuthToken() {
    return localStorage.getItem('auth_token');
  }

  // Check if user is authenticated
  isAuthenticated() {
    return !!this.getAuthToken();
  }

  // Set base URL
  setBaseURL(url) {
    this.baseURL = url;
    axios.defaults.baseURL = url;
  }

  // Create cancel token
  createCancelToken() {
    return axios.CancelToken.source();
  }

  // Check if request was cancelled
  isCancel(error) {
    return axios.isCancel(error);
  }
}

export default ApiService;