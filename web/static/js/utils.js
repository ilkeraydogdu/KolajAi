/**
 * Utility Functions for KolajAI
 * Bu dosya uygulama genelinde kullanılan yardımcı fonksiyonları içerir
 */

// Toast notification system
function showToast(message, type = 'info', duration = 5000) {
  // Toast container'ı kontrol et veya oluştur
  let toastContainer = document.querySelector('.toast-container');
  if (!toastContainer) {
    toastContainer = document.createElement('div');
    toastContainer.className = 'toast-container position-fixed top-0 end-0 p-3';
    toastContainer.style.zIndex = '9999';
    document.body.appendChild(toastContainer);
  }

  // Toast element oluştur
  const toastId = 'toast-' + Date.now();
  const toast = document.createElement('div');
  toast.id = toastId;
  toast.className = `toast align-items-center text-white bg-${getToastColor(type)} border-0`;
  toast.setAttribute('role', 'alert');
  toast.setAttribute('aria-live', 'assertive');
  toast.setAttribute('aria-atomic', 'true');

  toast.innerHTML = `
    <div class="d-flex">
      <div class="toast-body">
        ${getToastIcon(type)} ${message}
      </div>
      <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast" aria-label="Close"></button>
    </div>
  `;

  toastContainer.appendChild(toast);

  // Bootstrap toast initialize et
  const bsToast = new bootstrap.Toast(toast, {
    autohide: true,
    delay: duration
  });

  // Toast'u göster
  bsToast.show();

  // Toast kapandığında DOM'dan kaldır
  toast.addEventListener('hidden.bs.toast', () => {
    toast.remove();
  });

  return toast;
}

// Toast renk yardımcı fonksiyonu
function getToastColor(type) {
  const colors = {
    'success': 'success',
    'error': 'danger',
    'warning': 'warning',
    'info': 'info',
    'primary': 'primary'
  };
  return colors[type] || 'info';
}

// Toast icon yardımcı fonksiyonu
function getToastIcon(type) {
  const icons = {
    'success': '<i class="bi bi-check-circle-fill me-2"></i>',
    'error': '<i class="bi bi-exclamation-triangle-fill me-2"></i>',
    'warning': '<i class="bi bi-exclamation-triangle me-2"></i>',
    'info': '<i class="bi bi-info-circle-fill me-2"></i>',
    'primary': '<i class="bi bi-info-circle me-2"></i>'
  };
  return icons[type] || icons['info'];
}

// Currency formatting
function formatCurrency(amount, currency = 'TRY', locale = 'tr-TR') {
  if (amount === null || amount === undefined || isNaN(amount)) {
    return '0,00 ₺';
  }

  try {
    const formatter = new Intl.NumberFormat(locale, {
      style: 'currency',
      currency: currency,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    });
    return formatter.format(amount);
  } catch (error) {
    // Fallback formatting
    const formatted = parseFloat(amount).toFixed(2).replace('.', ',');
    return `${formatted} ₺`;
  }
}

// Number formatting
function formatNumber(number, locale = 'tr-TR') {
  if (number === null || number === undefined || isNaN(number)) {
    return '0';
  }

  try {
    return new Intl.NumberFormat(locale).format(number);
  } catch (error) {
    return number.toString();
  }
}

// Percentage formatting
function formatPercentage(value, decimals = 1) {
  if (value === null || value === undefined || isNaN(value)) {
    return '0%';
  }
  return `${parseFloat(value).toFixed(decimals)}%`;
}

// Date formatting utilities
function formatDate(date, format = 'short', locale = 'tr-TR') {
  if (!date) return '';
  
  const d = new Date(date);
  if (isNaN(d.getTime())) return '';

  const options = {
    'short': { day: '2-digit', month: '2-digit', year: 'numeric' },
    'long': { day: 'numeric', month: 'long', year: 'numeric' },
    'datetime': { 
      day: '2-digit', 
      month: '2-digit', 
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    }
  };

  try {
    return d.toLocaleDateString(locale, options[format] || options.short);
  } catch (error) {
    return d.toLocaleDateString();
  }
}

// Time ago formatting
function formatTimeAgo(date, locale = 'tr') {
  if (!date) return '';
  
  const now = new Date();
  const past = new Date(date);
  const diffMs = now - past;
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  const translations = {
    'tr': {
      'now': 'şimdi',
      'seconds': 'saniye önce',
      'minutes': 'dakika önce',
      'hours': 'saat önce',
      'days': 'gün önce',
      'weeks': 'hafta önce',
      'months': 'ay önce',
      'years': 'yıl önce'
    },
    'en': {
      'now': 'now',
      'seconds': 'seconds ago',
      'minutes': 'minutes ago',
      'hours': 'hours ago',
      'days': 'days ago',
      'weeks': 'weeks ago',
      'months': 'months ago',
      'years': 'years ago'
    }
  };

  const t = translations[locale] || translations['tr'];

  if (diffSecs < 60) return t.now;
  if (diffMins < 60) return `${diffMins} ${t.minutes}`;
  if (diffHours < 24) return `${diffHours} ${t.hours}`;
  if (diffDays < 7) return `${diffDays} ${t.days}`;
  if (diffDays < 30) return `${Math.floor(diffDays / 7)} ${t.weeks}`;
  if (diffDays < 365) return `${Math.floor(diffDays / 30)} ${t.months}`;
  return `${Math.floor(diffDays / 365)} ${t.years}`;
}

// Modal utilities
function showModal(title, content, options = {}) {
  const modalId = 'dynamic-modal-' + Date.now();
  const modal = document.createElement('div');
  modal.className = 'modal fade';
  modal.id = modalId;
  modal.setAttribute('tabindex', '-1');
  modal.setAttribute('aria-hidden', 'true');

  const size = options.size || '';
  const sizeClass = size ? `modal-${size}` : '';

  modal.innerHTML = `
    <div class="modal-dialog ${sizeClass}">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">${title}</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
        </div>
        <div class="modal-body">
          ${content}
        </div>
        ${options.footer ? `<div class="modal-footer">${options.footer}</div>` : ''}
      </div>
    </div>
  `;

  document.body.appendChild(modal);

  const bsModal = new bootstrap.Modal(modal);
  bsModal.show();

  // Modal kapandığında DOM'dan kaldır
  modal.addEventListener('hidden.bs.modal', () => {
    modal.remove();
  });

  return bsModal;
}

// Confirmation dialog
function showConfirm(message, title = 'Onay', options = {}) {
  return new Promise((resolve) => {
    const confirmText = options.confirmText || 'Evet';
    const cancelText = options.cancelText || 'Hayır';
    const type = options.type || 'warning';

    const footer = `
      <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">${cancelText}</button>
      <button type="button" class="btn btn-${type === 'danger' ? 'danger' : 'primary'}" id="confirm-btn">${confirmText}</button>
    `;

    const modal = showModal(title, message, { footer });

    const confirmBtn = document.getElementById('confirm-btn');
    confirmBtn.addEventListener('click', () => {
      modal.hide();
      resolve(true);
    });

    // Modal kapandığında false döndür
    document.getElementById(modal._element.id).addEventListener('hidden.bs.modal', () => {
      resolve(false);
    });
  });
}

// String utilities
function truncateText(text, maxLength = 100, suffix = '...') {
  if (!text || text.length <= maxLength) return text;
  return text.substring(0, maxLength - suffix.length) + suffix;
}

function capitalizeFirst(text) {
  if (!text) return text;
  return text.charAt(0).toUpperCase() + text.slice(1);
}

function slugify(text) {
  if (!text) return '';
  return text
    .toString()
    .toLowerCase()
    .trim()
    .replace(/\s+/g, '-')
    .replace(/[^\w\-]+/g, '')
    .replace(/\-\-+/g, '-')
    .replace(/^-+/, '')
    .replace(/-+$/, '');
}

// URL utilities
function getQueryParam(param) {
  const urlParams = new URLSearchParams(window.location.search);
  return urlParams.get(param);
}

function updateQueryParam(param, value) {
  const url = new URL(window.location);
  if (value) {
    url.searchParams.set(param, value);
  } else {
    url.searchParams.delete(param);
  }
  window.history.replaceState({}, '', url);
}

// Storage utilities
function setLocalStorage(key, value) {
  try {
    localStorage.setItem(key, JSON.stringify(value));
    return true;
  } catch (error) {
    console.error('LocalStorage error:', error);
    return false;
  }
}

function getLocalStorage(key, defaultValue = null) {
  try {
    const item = localStorage.getItem(key);
    return item ? JSON.parse(item) : defaultValue;
  } catch (error) {
    console.error('LocalStorage error:', error);
    return defaultValue;
  }
}

function removeLocalStorage(key) {
  try {
    localStorage.removeItem(key);
    return true;
  } catch (error) {
    console.error('LocalStorage error:', error);
    return false;
  }
}

// Validation utilities
function validateEmail(email) {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

function validatePhone(phone) {
  const phoneRegex = /^[\+]?[1-9][\d]{0,15}$/;
  return phoneRegex.test(phone.replace(/[\s\-\(\)]/g, ''));
}

function validateURL(url) {
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
}

// Performance utilities
function debounce(func, wait, immediate = false) {
  let timeout;
  return function executedFunction(...args) {
    const later = () => {
      timeout = null;
      if (!immediate) func.apply(this, args);
    };
    const callNow = immediate && !timeout;
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
    if (callNow) func.apply(this, args);
  };
}

function throttle(func, limit) {
  let inThrottle;
  return function executedFunction(...args) {
    if (!inThrottle) {
      func.apply(this, args);
      inThrottle = true;
      setTimeout(() => inThrottle = false, limit);
    }
  };
}

// Export functions to global scope
if (typeof window !== 'undefined') {
  window.showToast = showToast;
  window.formatCurrency = formatCurrency;
  window.formatNumber = formatNumber;
  window.formatPercentage = formatPercentage;
  window.formatDate = formatDate;
  window.formatTimeAgo = formatTimeAgo;
  window.showModal = showModal;
  window.showConfirm = showConfirm;
  window.truncateText = truncateText;
  window.capitalizeFirst = capitalizeFirst;
  window.slugify = slugify;
  window.getQueryParam = getQueryParam;
  window.updateQueryParam = updateQueryParam;
  window.setLocalStorage = setLocalStorage;
  window.getLocalStorage = getLocalStorage;
  window.removeLocalStorage = removeLocalStorage;
  window.validateEmail = validateEmail;
  window.validatePhone = validatePhone;
  window.validateURL = validateURL;
  window.debounce = debounce;
  window.throttle = throttle;
}

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    showToast,
    formatCurrency,
    formatNumber,
    formatPercentage,
    formatDate,
    formatTimeAgo,
    showModal,
    showConfirm,
    truncateText,
    capitalizeFirst,
    slugify,
    getQueryParam,
    updateQueryParam,
    setLocalStorage,
    getLocalStorage,
    removeLocalStorage,
    validateEmail,
    validatePhone,
    validateURL,
    debounce,
    throttle
  };
}