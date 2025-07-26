/**
 * KolajAI Login Module
 * Bu modül, giriş sayfası için özel işlevleri içerir.
 */

// Hemen çalıştırılacak anonim fonksiyon (IIFE)
(function(window) {
  'use strict';
  
  // KolajAI namespace'ini kullan
  const auth = window.KolajAI && window.KolajAI.auth ? window.KolajAI.auth : {};
  const notify = window.KolajAI && window.KolajAI.notify ? window.KolajAI.notify : null;
  
  // Modül değişkenleri
  let loginForm;
  let emailInput;
  let passwordInput;
  let rememberMeCheckbox;
  let submitButton;
  
  /**
   * Modülü başlat
   */
  function init() {
    console.log('Login module initialized');
    
    // DOM elementlerini seç
    loginForm = $('#loginForm');
    emailInput = $('#inputEmailAddress');
    passwordInput = $('#inputChoosePassword');
    rememberMeCheckbox = $('#flexSwitchCheckChecked');
    submitButton = $('.btn-grd-primary');
    
    // URL'deki parametreleri kontrol et
    checkURLParameters();
    
    // Event listener'ları ekle
    setupEventListeners();
  }
  
  /**
   * Event listener'ları ayarla
   */
  function setupEventListeners() {
    // Form gönderimi
    loginForm.on('submit', handleFormSubmit);
    
    // Enter tuşu ile gönderim
    passwordInput.on('keypress', function(e) {
      if (e.which === 13) {
        e.preventDefault();
        loginForm.submit();
      }
    });
  }
  
  /**
   * Form gönderimini işle
   * @param {Event} e - Form submit olayı
   */
  function handleFormSubmit(e) {
    // Form verilerini al
    const email = emailInput.val();
    const password = passwordInput.val();
    const rememberMe = rememberMeCheckbox.is(':checked');
    
    // Basit doğrulama
    if (!email || !password) {
      e.preventDefault();
      if (notify) {
        notify.warning('Lütfen tüm alanları doldurun', 'Uyarı');
      }
      return false;
    }
    
    // AJAX ile giriş yapma örneği (varsayılan form gönderimi yerine)
    // Bu kısmı aktif etmek için e.preventDefault() ekleyin
    /*
    e.preventDefault();
    
    // Butonun durumunu güncelle
    if (auth.core) {
      auth.core.updateSubmitButton('.btn-grd-primary', true, 'Giriş Yapılıyor...');
    } else {
      submitButton.prop('disabled', true).html('<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> Giriş Yapılıyor...');
    }
    
    // AJAX isteği gönder
    if (auth.core) {
      auth.core.sendAjaxRequest('/login', 'POST', {
        email: email,
        password: password,
        remember_me: rememberMe
      }, handleLoginSuccess, handleLoginError);
    } else {
      $.ajax({
        url: '/login',
        type: 'POST',
        data: JSON.stringify({
          email: email,
          password: password,
          remember_me: rememberMe
        }),
        contentType: 'application/json',
        success: handleLoginSuccess,
        error: handleLoginError
      });
    }
    */
  }
  
  /**
   * Başarılı giriş işlemini ele al
   * @param {object} response - Sunucu yanıtı
   */
  function handleLoginSuccess(response) {
    console.log('Login successful:', response);
    
    // Butonun durumunu geri al
    if (auth.core) {
      auth.core.updateSubmitButton('.btn-grd-primary', false, null, 'Giriş Yap');
    } else {
      submitButton.prop('disabled', false).text('Giriş Yap');
    }
    
    // Başarılı bildirim göster
    if (notify) {
      notify.success('Giriş başarılı. Yönlendiriliyorsunuz...', 'Başarılı');
    }
    
    // Yönlendirme
    setTimeout(function() {
      window.location.href = response.redirect || '/dashboard';
    }, 1000);
  }
  
  /**
   * Giriş hatasını ele al
   * @param {object} error - Hata bilgisi
   */
  function handleLoginError(error) {
    console.error('Login error:', error);
    
    // Butonun durumunu geri al
    if (auth.core) {
      auth.core.updateSubmitButton('.btn-grd-primary', false, null, 'Giriş Yap');
    } else {
      submitButton.prop('disabled', false).text('Giriş Yap');
    }
    
    // Hata mesajını göster
    let errorMessage = 'Giriş yapılamadı. Lütfen bilgilerinizi kontrol edin.';
    
    if (error.responseJSON && error.responseJSON.error) {
      errorMessage = error.responseJSON.error;
    } else if (error.responseText) {
      try {
        const parsedError = JSON.parse(error.responseText);
        if (parsedError.error) {
          errorMessage = parsedError.error;
        }
      } catch (e) {
        // JSON ayrıştırma hatası, varsayılan mesajı kullan
      }
    }
    
    // Hata bildirimini göster
    if (notify) {
      notify.error(errorMessage, 'Giriş Hatası');
    }
  }
  
  /**
   * URL parametrelerini kontrol et
   */
  function checkURLParameters() {
    const urlParams = new URLSearchParams(window.location.search);
    const redirect = urlParams.get('redirect');
    
    if (redirect) {
      // Yönlendirme URL'ini sakla
      sessionStorage.setItem('loginRedirect', redirect);
      
      // URL'i temizle
      const url = new URL(window.location.href);
      url.searchParams.delete('redirect');
      window.history.replaceState({}, document.title, url.toString());
    }
  }
  
  // Sayfa yüklendiğinde modülü başlat
  $(document).ready(init);
  
  // Modül API'sini dışa aktar
  window.KolajAI = window.KolajAI || {};
  window.KolajAI.auth = window.KolajAI.auth || {};
  window.KolajAI.auth.login = {
    init: init
  };
  
})(window); 