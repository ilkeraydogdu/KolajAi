/**
 * Logger Utility for KolajAI
 * Bu dosya production-safe logging sistemi sağlar
 */

class Logger {
  constructor() {
    // Determine environment
    this.isDevelopment = this.detectEnvironment();
    this.logLevel = this.getLogLevel();
  }

  detectEnvironment() {
    // Check various indicators for development environment
    if (typeof process !== 'undefined' && process.env && process.env.NODE_ENV) {
      return process.env.NODE_ENV === 'development';
    }
    
    // Check for localhost or development domains
    if (typeof window !== 'undefined') {
      const hostname = window.location.hostname;
      return hostname === 'localhost' || 
             hostname === '127.0.0.1' || 
             hostname.includes('.local') ||
             hostname.includes('dev.') ||
             hostname.includes('staging.');
    }
    
    // Default to production for safety
    return false;
  }

  getLogLevel() {
    if (typeof localStorage !== 'undefined') {
      const level = localStorage.getItem('kolajAI_log_level');
      if (level) return level;
    }
    
    return this.isDevelopment ? 'debug' : 'error';
  }

  setLogLevel(level) {
    this.logLevel = level;
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('kolajAI_log_level', level);
    }
  }

  shouldLog(level) {
    const levels = {
      'debug': 0,
      'info': 1,
      'warn': 2,
      'error': 3
    };
    
    return levels[level] >= levels[this.logLevel];
  }

  debug(...args) {
    if (this.shouldLog('debug')) {
      window.logger && window.logger.debug('[DEBUG]', ...args);
    }
  }

  info(...args) {
    if (this.shouldLog('info')) {
      console.info('[INFO]', ...args);
    }
  }

  warn(...args) {
    if (this.shouldLog('warn')) {
      console.warn('[WARN]', ...args);
    }
  }

  error(...args) {
    if (this.shouldLog('error')) {
      console.error('[ERROR]', ...args);
    }
  }

  // Special method for AJAX responses (only in development)
  ajax(message, data) {
    if (this.isDevelopment) {
      window.logger && window.logger.debug('[AJAX]', message, data);
    }
  }

  // Special method for form operations (only in development)
  form(message, data) {
    if (this.isDevelopment) {
      window.logger && window.logger.debug('[FORM]', message, data);
    }
  }

  // Special method for authentication operations (only in development)
  auth(message, data) {
    if (this.isDevelopment) {
      window.logger && window.logger.debug('[AUTH]', message, data);
    }
  }

  // Method to log user actions for analytics (always enabled but sanitized)
  userAction(action, details = {}) {
    // Sanitize sensitive data
    const sanitizedDetails = this.sanitizeData(details);
    
    // Send to analytics service if available
    if (typeof window !== 'undefined' && window.analytics) {
      window.analytics.track(action, sanitizedDetails);
    }
    
    // Log in development
    if (this.isDevelopment) {
      window.logger && window.logger.debug('[USER_ACTION]', action, sanitizedDetails);
    }
  }

  sanitizeData(data) {
    const sensitiveKeys = ['password', 'token', 'secret', 'key', 'auth'];
    const sanitized = {};
    
    for (const [key, value] of Object.entries(data)) {
      const lowerKey = key.toLowerCase();
      const isSensitive = sensitiveKeys.some(sensitive => lowerKey.includes(sensitive));
      
      if (isSensitive) {
        sanitized[key] = '[REDACTED]';
      } else if (typeof value === 'object' && value !== null) {
        sanitized[key] = this.sanitizeData(value);
      } else {
        sanitized[key] = value;
      }
    }
    
    return sanitized;
  }

  // Performance logging
  time(label) {
    if (this.isDevelopment) {
      console.time(label);
    }
  }

  timeEnd(label) {
    if (this.isDevelopment) {
      console.timeEnd(label);
    }
  }

  // Group logging for better organization
  group(label) {
    if (this.isDevelopment) {
      console.group(label);
    }
  }

  groupEnd() {
    if (this.isDevelopment) {
      console.groupEnd();
    }
  }
}

// Create global logger instance
const logger = new Logger();

// Export for global access
if (typeof window !== 'undefined') {
  window.logger = logger;
  
  // Backward compatibility - replace console.log in development
  if (logger.isDevelopment) {
    // Store original console methods
    window._originalConsole = {
      log: console.log,
      info: console.info,
      warn: console.warn,
      error: console.error
    };
  }
}

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
  module.exports = Logger;
}

export default Logger;