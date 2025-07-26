/**
 * KolajAI Auth Core Module
 * Bu modül, kimlik doğrulama ile ilgili tüm sayfalarda kullanılan ortak fonksiyonları içerir.
 */

// Hemen çalıştırılacak anonim fonksiyon (IIFE)
(function(window) {
  'use strict';
  
  // KolajAI namespace oluştur veya mevcut olanı kullan
  window.KolajAI = window.KolajAI || {};
  window.KolajAI.auth = window.KolajAI.auth || {};
  
  /**
   * Şifre görünürlüğünü değiştiren fonksiyon
   * @param {string} containerId - Şifre alanını içeren container ID'si
   */
  function togglePasswordVisibility(containerId) {
    const passwordInput = $(containerId + ' input');
    const icon = $(containerId + ' i');
    
    if (passwordInput.attr('type') === 'password') {
      passwordInput.attr('type', 'text');
      icon.removeClass('bi-eye-slash-fill').addClass('bi-eye-fill');
    } else {
      passwordInput.attr('type', 'password');
      icon.removeClass('bi-eye-fill').addClass('bi-eye-slash-fill');
    }
  }
  
  /**
   * AJAX isteği gönderme fonksiyonu
   * @param {string} url - İstek URL'i
   * @param {string} method - HTTP metodu (GET, POST, PUT, DELETE)
   * @param {object|string} data - Gönderilecek veri
   * @param {function} successCallback - Başarılı yanıt için callback
   * @param {function} errorCallback - Hata durumunda callback
   * @param {object} options - Ek ayarlar
   */
  function sendAjaxRequest(url, method, data, successCallback, errorCallback, options = {}) {
    console.log("Sending AJAX request to:", url, "with data:", data);
    
    // Varsayılan ayarlar
    const defaultOptions = {
      contentType: 'application/json',
      processData: false,
      crossDomain: true,
      xhrFields: {
        withCredentials: false
      }
    };
    
    // Ayarları birleştir
    const ajaxOptions = Object.assign({}, defaultOptions, options);
    
    // JSON veri formatı için
    let processedData = data;
    if (ajaxOptions.contentType === 'application/json' && typeof data === 'object') {
      processedData = JSON.stringify(data);
    }
    
    // AJAX isteği gönder
    $.ajax({
      url: url,
      type: method,
      data: processedData,
      contentType: ajaxOptions.contentType,
      processData: ajaxOptions.processData,
      headers: {
        'X-Requested-With': 'XMLHttpRequest',
        'Accept': 'application/json'
      },
      crossDomain: ajaxOptions.crossDomain,
      xhrFields: ajaxOptions.xhrFields,
      success: function(response) {
        console.log("AJAX response:", response);
        if (successCallback) {
          successCallback(response);
        }
      },
      error: function(xhr, status, error) {
        console.error("AJAX error:", {
          status: xhr.status,
          statusText: xhr.statusText,
          responseText: xhr.responseText,
          error: error
        });
        
        if (errorCallback) {
          errorCallback({
            status: xhr.status,
            statusText: xhr.statusText,
            responseText: xhr.responseText,
            error: error
          });
        }
      }
    });
  }
  
  /**
   * Form validasyonu başlat
   * @param {string} formSelector - Form seçicisi
   * @param {object} options - Validasyon ayarları
   */
  function initFormValidation(formSelector, options = {}) {
    const form = $(formSelector);
    
    if (!form.length) {
      console.warn(`Form not found: ${formSelector}`);
      return;
    }
    
    if (!$.fn.validate) {
      console.warn('jQuery Validate plugin not loaded');
      return;
    }
    
    // Form üzerinde data-rules özniteliği varsa, JSON olarak ayrıştır
    let validationRules = {};
    const rulesAttr = form.attr('data-rules');
    
    if (rulesAttr) {
      try {
        validationRules = JSON.parse(rulesAttr);
      } catch (e) {
        console.error("Error parsing validation rules:", e);
      }
    }
    
    // Kullanıcı tarafından sağlanan ayarları birleştir
    const validationOptions = Object.assign({}, validationRules, options);
    
    // Validasyonu başlat
    form.validate(validationOptions);
    
    return form;
  }
  
  /**
   * Form alanlarını temizle
   * @param {string} formSelector - Form seçicisi
   */
  function clearForm(formSelector) {
    const form = $(formSelector);
    
    if (!form.length) {
      console.warn(`Form not found: ${formSelector}`);
      return;
    }
    
    form[0].reset();
    form.find('.is-invalid').removeClass('is-invalid');
    form.find('.is-valid').removeClass('is-valid');
    form.find('.invalid-feedback').text('');
    form.find('.valid-feedback').text('');
  }
  
  /**
   * Form gönderim durumunu güncelle
   * @param {string} buttonSelector - Buton seçicisi
   * @param {boolean} isLoading - Yükleniyor durumu
   * @param {string} loadingText - Yükleniyor metni
   * @param {string} originalText - Orijinal buton metni
   */
  function updateSubmitButton(buttonSelector, isLoading, loadingText = 'İşleniyor...', originalText = null) {
    const button = $(buttonSelector);
    
    if (!button.length) {
      console.warn(`Button not found: ${buttonSelector}`);
      return;
    }
    
    if (isLoading) {
      // Orijinal metni data özniteliğine kaydet
      if (!button.data('original-text')) {
        button.data('original-text', button.html());
      }
      
      // Yükleniyor durumunu göster
      button.prop('disabled', true).html(`<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> ${loadingText}`);
    } else {
      // Orijinal metni geri yükle veya parametre olarak verilen metni kullan
      const textToRestore = originalText || button.data('original-text') || 'Gönder';
      button.prop('disabled', false).html(textToRestore);
    }
  }
  
  /**
   * Form verilerini JSON nesnesine dönüştür
   * @param {string} formSelector - Form seçicisi
   * @returns {object} Form verileri
   */
  function getFormData(formSelector) {
    const form = $(formSelector);
    
    if (!form.length) {
      console.warn(`Form not found: ${formSelector}`);
      return {};
    }
    
    // Form verilerini topla
    const formArray = form.serializeArray();
    const formData = {};
    
    // Her alan için değeri ekle
    formArray.forEach(item => {
      formData[item.name] = item.value;
    });
    
    return formData;
  }
  
  // Ortak fonksiyonları dışa aktar
  window.KolajAI.auth.core = {
    togglePasswordVisibility,
    sendAjaxRequest,
    initFormValidation,
    clearForm,
    updateSubmitButton,
    getFormData
  };
  
  // Sayfa yüklendiğinde ortak işlemleri başlat
  $(document).ready(function() {
    // Şifre göster/gizle işlevselliği
    $("[id^='show_hide_password'] a").on('click', function(event) {
      event.preventDefault();
      const containerId = '#' + $(this).closest('.input-group').attr('id');
      togglePasswordVisibility(containerId);
    });
    
    // Form validasyonu
    $("form[data-validate='true']").each(function() {
      initFormValidation('#' + $(this).attr('id'));
    });
  });
  
})(window); 