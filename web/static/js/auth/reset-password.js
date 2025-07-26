/**
 * KolajAI Reset Password Module
 * Bu modül, şifre sıfırlama sayfası için özel işlevleri içerir.
 */

// Hemen çalıştırılacak anonim fonksiyon (IIFE)
(function(window) {
  'use strict';
  
  // KolajAI namespace'ini kullan
  const auth = window.KolajAI && window.KolajAI.auth ? window.KolajAI.auth : {};
  const notify = window.KolajAI && window.KolajAI.notify ? window.KolajAI.notify : null;
  
  // Modül değişkenleri
  let resetPasswordForm;
  let tempPasswordInput;
  let newPasswordInput;
  let confirmPasswordInput;
  let newPasswordSection;
  let confirmPasswordSection;
  let submitButtonSection;
  let tempPasswordError;
  let tempPasswordVerified = false;
  let verificationInProgress = false;
  
  /**
   * Modülü başlat
   */
  function init() {
    console.log("Reset password module initialized");
    
    // DOM elementlerini seç
    resetPasswordForm = $('#resetPasswordForm');
    tempPasswordInput = $('#inputTempPassword');
    newPasswordInput = $('#inputChoosePassword');
    confirmPasswordInput = $('#inputChoosePassword2');
    newPasswordSection = $('#newPasswordSection');
    confirmPasswordSection = $('#confirmPasswordSection');
    submitButtonSection = $('#submitButtonSection');
    tempPasswordError = $('#tempPasswordError');
    
    // E-posta parametresini logla
    const email = $('input[name="email"]').val();
    console.log("Email parameter:", email);
    
    // URL'deki parametreleri temizle
    if (window.history.replaceState) {
      window.history.replaceState(null, null, window.location.pathname);
    }

    // Başarı mesajını göster
    if (notify) {
      notify.success('Hesabınız başarıyla oluşturuldu. Size e-posta ile gönderilen geçici şifre ile hemen yeni bir şifre belirleyebilirsiniz.', 'Kayıt Başarılı');
    }
    
    // Event listener'ları ekle
    setupEventListeners();
  }
  
  /**
   * Event listener'ları ayarla
   */
  function setupEventListeners() {
    // Şifre alanında değişiklik olduğunda otomatik doğrulama
    tempPasswordInput.on('input', function() {
      verifyTempPassword($(this).val());
    });
    
    // Form gönderimi
    resetPasswordForm.on('submit', handleFormSubmit);
  }
  
  /**
   * Geçici şifreyi doğrula
   * @param {string} password - Doğrulanacak geçici şifre
   */
  function verifyTempPassword(password) {
    // Şifre boşsa işlemi atla
    if (!password || password.length < 6) {
      tempPasswordInput.removeClass('is-valid').addClass('is-invalid');
      tempPasswordError.text('Geçici şifre en az 6 karakter olmalıdır');
      tempPasswordVerified = false;
      return;
    }
    
    // Zaten doğrulama işlemi devam ediyorsa tekrar istek gönderme
    if (verificationInProgress) {
      return;
    }
    
    // Doğrulama işlemini başlat
    verificationInProgress = true;
    tempPasswordInput.removeClass('is-invalid').removeClass('is-valid');
    
    // E-posta adresini al
    const email = $('input[name="email"]').val();
    
    // AJAX isteği gönder
    if (auth.core) {
      auth.core.sendAjaxRequest('/validate-temp-password', 'POST', {
        email: email,
        temp_password: password
      }, handleVerificationSuccess, handleVerificationError);
    } else {
      $.ajax({
        url: '/validate-temp-password',
        type: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({
          email: email,
          temp_password: password
        }),
        success: handleVerificationSuccess,
        error: handleVerificationError
      });
    }
  }
  
  /**
   * Başarılı doğrulama işlemini ele al
   * @param {object} response - Sunucu yanıtı
   */
  function handleVerificationSuccess(response) {
    console.log("Verification successful:", response);
    verificationInProgress = false;
    
    if (response.success) {
      tempPasswordVerified = true;
      tempPasswordInput.removeClass('is-invalid').addClass('is-valid');
      
      // Yeni şifre alanlarını göster
      newPasswordSection.removeClass('d-none');
      confirmPasswordSection.removeClass('d-none');
      submitButtonSection.removeClass('d-none');
      
      // Bildirim göster
      if (notify) {
        notify.success('Geçici şifre doğrulandı. Lütfen yeni şifrenizi belirleyin.', 'Doğrulama Başarılı');
      }
    } else {
      tempPasswordVerified = false;
      tempPasswordInput.removeClass('is-valid').addClass('is-invalid');
      tempPasswordError.text(response.error || 'Geçici şifre doğrulanamadı');
    }
  }
  
  /**
   * Doğrulama hatasını ele al
   * @param {object} error - Hata bilgisi
   */
  function handleVerificationError(error) {
    console.error("Verification error:", error);
    verificationInProgress = false;
    tempPasswordVerified = false;
    
    tempPasswordInput.removeClass('is-valid').addClass('is-invalid');
    let errorMessage = 'Geçici şifre doğrulanamadı';
    
    if (error.responseJSON && error.responseJSON.error) {
      errorMessage = error.responseJSON.error;
    }
    
    tempPasswordError.text(errorMessage);
  }
  
  /**
   * Form gönderimini işle
   * @param {Event} e - Form submit olayı
   */
  function handleFormSubmit(e) {
    e.preventDefault();
    console.log("Form submit event triggered");
    
    // Geçici şifre doğrulanmadıysa engelle
    if (!tempPasswordVerified) {
      console.error("Geçici şifre doğrulanmadı, form gönderimi engellendi");
      tempPasswordError.text('Lütfen önce geçici şifreyi doğrulayın');
      if (notify) {
        notify.warning('Lütfen önce geçici şifreyi doğrulayın', 'Uyarı');
      }
      return false;
    }
    
    // Şifre ve şifre onayı eşleşiyor mu kontrol et
    const password = newPasswordInput.val();
    const confirmPassword = confirmPasswordInput.val();
    const email = $('input[name="email"]').val();
    
    if (password !== confirmPassword) {
      $('#passwordError').text('Şifreler eşleşmiyor');
      confirmPasswordInput.addClass('is-invalid');
      if (notify) {
        notify.error('Şifreler eşleşmiyor', 'Hata');
      }
      return false;
    }
    
    // Şifre uzunluğu kontrolü
    if (password.length < 6) {
      $('#passwordError').text('Şifre en az 6 karakter olmalıdır');
      newPasswordInput.addClass('is-invalid');
      if (notify) {
        notify.error('Şifre en az 6 karakter olmalıdır', 'Hata');
      }
      return false;
    }
    
    console.log("Şifre değiştirme işlemi başlatılıyor...");
    
    // Butonun durumunu güncelle
    const submitButton = $('#submitButtonSection button');
    if (auth.core) {
      auth.core.updateSubmitButton('#submitButtonSection button', true, 'İşleniyor...');
    } else {
      const originalText = submitButton.text();
      submitButton.prop('disabled', true).html('<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> İşleniyor...');
      submitButton.data('original-text', originalText);
    }
    
    // AJAX ile şifre değiştirme isteği gönder
    if (auth.core) {
      auth.core.sendAjaxRequest('/reset-password', 'POST', {
        email: email,
        password: password,
        password_confirm: confirmPassword
      }, handleResetSuccess, handleResetError);
    } else {
      $.ajax({
        url: '/reset-password',
        type: 'POST',
        contentType: 'application/json',
        dataType: 'json',
        data: JSON.stringify({
          email: email,
          password: password,
          password_confirm: confirmPassword
        }),
        headers: {
          'X-Requested-With': 'XMLHttpRequest',
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        },
        success: handleResetSuccess,
        error: handleResetError
      });
    }
  }
  
  /**
   * Başarılı şifre sıfırlama işlemini ele al
   * @param {object} response - Sunucu yanıtı
   */
  function handleResetSuccess(response) {
    console.log("Şifre değiştirme başarılı:", response);
    
    // Butonun durumunu geri al
    const submitButton = $('#submitButtonSection button');
    if (auth.core) {
      auth.core.updateSubmitButton('#submitButtonSection button', false);
    } else {
      submitButton.prop('disabled', false).text(submitButton.data('original-text') || 'Şifreyi Değiştir');
    }
    
    // Başarılı bildirim göster
    if (notify) {
      notify.success('Şifreniz başarıyla değiştirildi', 'Başarılı');
    }
    
    // Login sayfasına yönlendir
    setTimeout(function() {
      if (notify && notify.redirectWith) {
        notify.redirectWith('/login', 'success', 'Başarılı', 'Şifreniz başarıyla değiştirildi. Yeni şifrenizle giriş yapabilirsiniz.');
      } else {
        window.location.href = "/login?messageType=success&messageTitle=Başarılı&messageText=Şifreniz+başarıyla+değiştirildi.+Yeni+şifrenizle+giriş+yapabilirsiniz.";
      }
    }, 2000);
  }
  
  /**
   * Şifre sıfırlama hatasını ele al
   * @param {object} error - Hata bilgisi
   */
  function handleResetError(error) {
    console.error("Şifre değiştirme hatası:", error);
    
    // Butonun durumunu geri al
    const submitButton = $('#submitButtonSection button');
    if (auth.core) {
      auth.core.updateSubmitButton('#submitButtonSection button', false);
    } else {
      submitButton.prop('disabled', false).text(submitButton.data('original-text') || 'Şifreyi Değiştir');
    }
    
    // Hata mesajını göster
    let errorMessage = 'Şifre değiştirme işlemi başarısız oldu.';
    
    if (error.responseJSON && error.responseJSON.error) {
      errorMessage = error.responseJSON.error;
    }
    
    // Hata bildirimini göster
    if (notify) {
      notify.error(errorMessage, 'Hata');
    }
  }
  
  // Sayfa yüklendiğinde modülü başlat
  $(document).ready(init);
  
  // Modül API'sini dışa aktar
  window.KolajAI = window.KolajAI || {};
  window.KolajAI.auth = window.KolajAI.auth || {};
  window.KolajAI.auth.resetPassword = {
    init: init,
    verifyTempPassword: verifyTempPassword
  };
  
})(window); 