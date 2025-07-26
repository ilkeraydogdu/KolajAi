/**
 * KolajAI Notifications Module
 * Bu modül, tüm uygulama genelinde kullanılabilecek bildirim işlevlerini sağlar.
 */

// Hemen çalıştırılacak anonim fonksiyon (IIFE)
(function(window) {
  'use strict';
  
  // KolajAI namespace oluştur veya mevcut olanı kullan
  window.KolajAI = window.KolajAI || {};
  
  // Bildirim ayarları
  const notificationDefaults = {
    position: 'top right',
    showClass: 'fadeInRight',
    hideClass: 'fadeOutRight',
    delay: 5000,
    rounded: true,
    delayIndicator: true,
    sound: false
  };
  
  // Bildirim tiplerine göre ikonlar
  const notificationIcons = {
    'success': 'bi bi-check-circle',
    'error': 'bi bi-x-circle',
    'warning': 'bi bi-exclamation-triangle',
    'info': 'bi bi-info-circle',
    'default': 'bi bi-bell'
  };
  
  /**
   * Bildirim gösterme fonksiyonu
   * @param {string} type - Bildirim tipi (success, error, warning, info)
   * @param {string} title - Bildirim başlığı
   * @param {string} message - Bildirim mesajı
   * @param {object} options - Ek ayarlar
   */
  function showNotification(type, title, message, options = {}) {
    // Geçerli bir tip değilse varsayılanı kullan
    if (!['success', 'error', 'warning', 'info', 'default'].includes(type)) {
      type = 'info';
    }
    
    // Varsayılan ayarları ve kullanıcı ayarlarını birleştir
    const settings = Object.assign({}, notificationDefaults, options);
    
    // Bildirim ikonunu ayarla
    settings.icon = options.icon || notificationIcons[type] || notificationIcons.default;
    
    // Bildirim göster
    if (window.Lobibox && window.Lobibox.notify) {
      Lobibox.notify(type, {
        title: title,
        msg: message,
        sound: false, // Ses devre dışı
        soundPath: null, // Ses dosyası yolu yok
        ...settings
      });
      
      // Debug için konsola bilgi yazdır
      if (window.KolajAI.debug) {
        console.debug(`Notification shown: ${type} - ${title}`);
      }
    } else {
      // Lobibox yoksa basit bir alert göster
      console.warn('Lobibox notification library not available');
      alert(`${title}: ${message}`);
    }
  }
  
  /**
   * URL parametrelerinden bildirim gösterme
   * Örnek: ?messageType=success&messageTitle=Başarılı&messageText=İşlem+başarıyla+tamamlandı
   */
  function showNotificationsFromURL() {
    const urlParams = new URLSearchParams(window.location.search);
    const messageType = urlParams.get('messageType');
    const messageTitle = urlParams.get('messageTitle');
    const messageText = urlParams.get('messageText');
    
    if (messageType && messageText) {
      showNotification(
        messageType,
        messageTitle || getDefaultTitle(messageType),
        messageText
      );
      
      // Bildirim gösterildikten sonra URL'i temizle (tarayıcı geçmişini etkilemeden)
      const url = new URL(window.location.href);
      url.searchParams.delete('messageType');
      url.searchParams.delete('messageTitle');
      url.searchParams.delete('messageText');
      window.history.replaceState({}, document.title, url.toString());
    }
  }
  
  /**
   * Sayfa yüklendiğinde otomatik bildirimler göster
   */
  function showAutoNotifications() {
    if (window.KolajAI && window.KolajAI.notifications) {
      const { success, error, info, warning } = window.KolajAI.notifications;
      
      if (success) {
        showNotification('success', 'Başarılı', success);
      }
      
      if (error) {
        showNotification('error', 'Hata', error);
      }
      
      if (info) {
        showNotification('info', 'Bilgi', info);
      }
      
      if (warning) {
        showNotification('warning', 'Uyarı', warning);
      }
    }
  }
  
  /**
   * Varsayılan bildirim başlıkları
   * @param {string} type - Bildirim tipi
   * @returns {string} Varsayılan başlık
   */
  function getDefaultTitle(type) {
    switch (type) {
      case 'success': return 'Başarılı';
      case 'error': return 'Hata';
      case 'warning': return 'Uyarı';
      case 'info': return 'Bilgi';
      default: return 'Bildirim';
    }
  }
  
  /**
   * Bildirim ile başka bir sayfaya yönlendirme
   * @param {string} url - Yönlendirilecek URL
   * @param {string} type - Bildirim tipi
   * @param {string} title - Bildirim başlığı
   * @param {string} message - Bildirim mesajı
   */
  function redirectWithNotification(url, type, title, message) {
    const redirectUrl = new URL(url, window.location.origin);
    redirectUrl.searchParams.set('messageType', type);
    if (title) redirectUrl.searchParams.set('messageTitle', title);
    redirectUrl.searchParams.set('messageText', message);
    window.location.href = redirectUrl.toString();
  }
  
  // Sayfa yüklendiğinde bildirimleri göster
  document.addEventListener('DOMContentLoaded', function() {
    showNotificationsFromURL();
    showAutoNotifications();
  });
  
  // Bildirim API'sini dışa aktar
  window.KolajAI.notify = {
    show: showNotification,
    success: (message, title = 'Başarılı', options = {}) => showNotification('success', title, message, options),
    error: (message, title = 'Hata', options = {}) => showNotification('error', title, message, options),
    info: (message, title = 'Bilgi', options = {}) => showNotification('info', title, message, options),
    warning: (message, title = 'Uyarı', options = {}) => showNotification('warning', title, message, options),
    redirectWith: redirectWithNotification
  };
  
})(window); 