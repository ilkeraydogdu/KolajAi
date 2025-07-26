/**
 * KolajAI Auth Module
 * Bu modül kimlik doğrulama ile ilgili tüm JavaScript işlevlerini içerir
 */

// Uygulama genelinde kullanılacak yardımcı fonksiyonlar
const AuthHelpers = {
  // Şifre görünürlüğünü değiştiren fonksiyon
  togglePasswordVisibility: function(containerId) {
    const passwordInput = $(containerId + ' input');
    const icon = $(containerId + ' i');
    
    if (passwordInput.attr('type') === 'password') {
      passwordInput.attr('type', 'text');
      icon.removeClass('bi-eye-slash-fill').addClass('bi-eye-fill');
    } else {
      passwordInput.attr('type', 'password');
      icon.removeClass('bi-eye-fill').addClass('bi-eye-slash-fill');
    }
  },
  
  // Bildirim gösterme fonksiyonu
  showNotification: function(type, title, message) {
    // Bildirim tiplerine göre renkler
    const typeClasses = {
      'success': 'bg-success text-white',
      'error': 'bg-danger text-white',
      'warning': 'bg-warning',
      'info': 'bg-info text-white'
    };
    
    // Toast elementini oluştur
    const toastId = 'toast-' + Date.now();
    const toastHtml = `
      <div id="${toastId}" class="toast ${typeClasses[type] || ''}" role="alert" aria-live="assertive" aria-atomic="true">
        <div class="toast-header">
          <strong class="me-auto">${title}</strong>
          <button type="button" class="btn-close" data-bs-dismiss="toast" aria-label="Close"></button>
        </div>
        <div class="toast-body">
          ${message}
        </div>
      </div>
    `;
    
    // Toast container'ı kontrol et veya oluştur
    let toastContainer = $('.toast-container');
    if (toastContainer.length === 0) {
      toastContainer = $('<div class="toast-container position-fixed top-0 end-0 p-3"></div>');
      $('body').append(toastContainer);
    }
    
    // Toast'u ekle ve göster
    const toastElement = $(toastHtml);
    toastContainer.append(toastElement);
    
    const toast = new bootstrap.Toast(toastElement[0], {
      autohide: true,
      delay: 5000
    });
    
    toast.show();
    
    // 5 saniye sonra otomatik kaldır
    setTimeout(function() {
      toastElement.remove();
    }, 5000);
  },
  
  // AJAX isteği gönderme fonksiyonu
  sendAjaxRequest: function(url, method, data, successCallback, errorCallback) {
    console.log("Sending AJAX request to:", url, "with data:", data);
    
    $.ajax({
      url: url,
      type: method,
      data: data,
      headers: {
        'X-Requested-With': 'XMLHttpRequest',
        'Accept': 'application/json'
      },
      crossDomain: true,
      xhrFields: {
        withCredentials: false
      },
      success: function(response) {
        console.log("AJAX response:", response);
        if (successCallback) {
          successCallback(response);
        }
      },
      error: function(xhr, status, error) {
        console.error("AJAX hatası:", {
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
};

// Şifre sıfırlama sayfası işlevleri
const ResetPassword = {
  // Değişkenler
  tempPasswordVerified: false,
  verificationInProgress: false,
  
  init: function() {
    console.log("Reset password page loaded");
    
    // DOM elementlerini seç
    this.tempPasswordInput = $('#inputTempPassword');
    this.newPasswordSection = $('#newPasswordSection');
    this.confirmPasswordSection = $('#confirmPasswordSection');
    this.submitButtonSection = $('#submitButtonSection');
    this.tempPasswordError = $('#tempPasswordError');
    this.resetPasswordForm = $('#resetPasswordForm');
    
    // E-posta parametresini logla
    const email = $('input[name="email"]').val();
    console.log("Email parameter:", email);
    
    // URL'deki parametreleri temizle
    if (window.history.replaceState) {
      window.history.replaceState(null, null, window.location.pathname);
    }

    // Başarı mesajını göster
    AuthHelpers.showNotification('success', 'Kayıt Başarılı', 
      'Hesabınız başarıyla oluşturuldu. Size e-posta ile gönderilen geçici şifre ile hemen yeni bir şifre belirleyebilirsiniz.');
    
    // Event listener'ları ekle
    this.setupEventListeners();
  },
  
  setupEventListeners: function() {
    const self = this;
    
    // Şifre alanında değişiklik olduğunda otomatik doğrulama
    this.tempPasswordInput.on('input', function() {
      self.verifyTempPassword($(this).val());
    });

    // Şifre göster/gizle fonksiyonları
    $("#show_hide_temp_password a").on('click', function(event) {
      event.preventDefault();
      AuthHelpers.togglePasswordVisibility('#show_hide_temp_password');
    });

    $("#show_hide_password a").on('click', function(event) {
      event.preventDefault();
      AuthHelpers.togglePasswordVisibility('#show_hide_password');
    });

    $("#show_hide_password2 a").on('click', function(event) {
      event.preventDefault();
      AuthHelpers.togglePasswordVisibility('#show_hide_password2');
    });

    // Form gönderimi
    this.resetPasswordForm.on('submit', function(e) {
      e.preventDefault(); // Form gönderimini engelle
      console.log("Form submit event triggered");
      
      if (!self.tempPasswordVerified) {
        console.error("Geçici şifre doğrulanmadı, form gönderimi engellendi");
        self.tempPasswordError.text('Lütfen önce geçici şifreyi doğrulayın');
        AuthHelpers.showNotification('warning', 'Uyarı', 'Lütfen önce geçici şifreyi doğrulayın');
        return false;
      }
      
      // Şifre ve şifre onayı eşleşiyor mu kontrol et
      const password = $('#inputChoosePassword').val();
      const confirmPassword = $('#inputChoosePassword2').val();
      const email = $('input[name="email"]').val();
      
      if (password !== confirmPassword) {
        $('#passwordError').text('Şifreler eşleşmiyor');
        $('#inputChoosePassword2').addClass('is-invalid');
        AuthHelpers.showNotification('error', 'Hata', 'Şifreler eşleşmiyor');
        return false;
      }
      
      // Şifre uzunluğu kontrolü
      if (password.length < 6) {
        $('#passwordError').text('Şifre en az 6 karakter olmalıdır');
        $('#inputChoosePassword').addClass('is-invalid');
        AuthHelpers.showNotification('error', 'Hata', 'Şifre en az 6 karakter olmalıdır');
        return false;
      }
      
      console.log("Şifre değiştirme işlemi başlatılıyor...");
      console.log("Form verileri:", {
        email: email,
        password: password,
        password_confirm: confirmPassword
      });
      
      // Butonun durumunu güncelle
      const submitButton = $('#submitButtonSection button');
      const originalText = submitButton.text();
      submitButton.prop('disabled', true).html('<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> İşleniyor...');
      
      // AJAX ile şifre değiştirme isteği gönder
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
        beforeSend: function(xhr) {
          // CORS için OPTIONS isteğine izin ver
          xhr.setRequestHeader('Access-Control-Allow-Methods', 'POST, OPTIONS');
          xhr.setRequestHeader('Access-Control-Allow-Headers', 'Content-Type, X-Requested-With');
        },
        success: function(response) {
          console.log("Şifre değiştirme başarılı:", response);
          // Başarılı mesajı göster
          AuthHelpers.showNotification('success', 'Başarılı', 'Şifreniz başarıyla değiştirildi');
          // Login sayfasına yönlendir
          setTimeout(function() {
            window.location.href = "/login?messageType=success&messageTitle=Başarılı&messageText=Şifreniz+başarıyla+değiştirildi.+Yeni+şifrenizle+giriş+yapabilirsiniz.";
          }, 2000);
        },
        error: function(xhr, status, error) {
          console.error("Şifre değiştirme hatası:", error);
          // Hatayı göster
          submitButton.prop('disabled', false).text(originalText);
          AuthHelpers.showNotification('error', 'Hata', xhr.responseJSON?.error || 'Şifre değiştirme işlemi başarısız oldu.');
        }
      });
    });
  },
  
  verifyTempPassword: function(tempPassword) {
    const self = this;
    
    if (tempPassword.length < 3) {
      return;
    }
    
    if (this.verificationInProgress) {
      return;
    }
    
    this.verificationInProgress = true;
    
    console.log("Temp password changed, length:", tempPassword.length);
    
    // Doğrulama için AJAX isteği
    AuthHelpers.sendAjaxRequest(
      'http://localhost:8080/verify-temp-password',
      'POST',
      {
        email: $('input[name="email"]').val(),
        temp_password: tempPassword
      },
      function(response) {
        self.verificationInProgress = false;
        
        console.log("AJAX başarılı:", response);
        if (response.success) {
          self.tempPasswordVerified = true;
          self.tempPasswordInput.removeClass('is-invalid').addClass('is-valid');
          self.tempPasswordError.text('');
          
          // Yeni şifre alanlarını göster
          self.newPasswordSection.removeClass('d-none');
          self.confirmPasswordSection.removeClass('d-none');
          self.submitButtonSection.removeClass('d-none');
          
          // Geçici şifre alanını devre dışı bırak
          self.tempPasswordInput.prop('disabled', true);
          
          // Bildirim göster
          AuthHelpers.showNotification('success', 'Doğrulama Başarılı', 'Geçici şifre doğrulandı. Lütfen yeni şifrenizi belirleyin.');
        } else {
          self.tempPasswordVerified = false;
          self.tempPasswordInput.removeClass('is-valid').addClass('is-invalid');
          
          // Hata mesajını göster
          const errorMessage = response.error || "Geçici şifre doğrulanamadı";
          self.tempPasswordError.text(errorMessage);
          console.error("Şifre doğrulama hatası:", errorMessage);
        }
      },
      function(error) {
        self.verificationInProgress = false;
        self.tempPasswordVerified = false;
        self.tempPasswordInput.removeClass('is-valid').addClass('is-invalid');
        
        // Hata mesajını göster
        console.error("AJAX hatası:", error);
        self.tempPasswordError.text("Bağlantı hatası. Lütfen tekrar deneyin.");
      }
    );
  }
};

// Login sayfası işlevleri
const Login = {
  init: function() {
    console.log("Login page loaded");
    this.setupEventListeners();
    
    // URL parametrelerinden mesaj göster
    this.showMessageFromURL();
  },
  
  setupEventListeners: function() {
    // Şifre göster/gizle
    $("#show_hide_password a").on('click', function(event) {
      event.preventDefault();
      AuthHelpers.togglePasswordVisibility('#show_hide_password');
    });
    
    // Form gönderimi
    $('#loginForm').on('submit', function(e) {
      console.log("Login form submit");
      
      // Form doğrulama
      const email = $('#inputEmailAddress').val();
      const password = $('#inputChoosePassword').val();
      
      if (!email || !password) {
        e.preventDefault();
        AuthHelpers.showNotification('warning', 'Uyarı', 'Lütfen tüm alanları doldurun');
        return false;
      }
    });
  },
  
  showMessageFromURL: function() {
    // URL'den mesaj parametrelerini al
    const urlParams = new URLSearchParams(window.location.search);
    const messageType = urlParams.get('messageType');
    const messageTitle = urlParams.get('messageTitle');
    const messageText = urlParams.get('messageText');
    
    // Mesaj varsa göster
    if (messageType && messageText) {
      AuthHelpers.showNotification(messageType, messageTitle || 'Bilgi', messageText);
      
      // URL'i temizle
      if (window.history.replaceState) {
        window.history.replaceState(null, null, window.location.pathname);
      }
    }
  }
};

// Register sayfası işlevleri
const Register = {
  // Form alanlarının doğrulama durumunu takip etmek için değişkenler
  formFields: {
    name: false,
    email: false,
    phone: false
  },
  
  init: function() {
    console.log("Register page loaded");
    this.setupEventListeners();
    
    // CAPTCHA oluştur (başlangıçta gizli)
    this.generateMathCaptcha();
  },
  
  setupEventListeners: function() {
    const self = this;
    
    // Form alanları değiştiğinde doğrulama yap
    $('#inputName').on('input blur', function() {
      const value = $(this).val();
      const isValid = value.length >= 5;
      self.formFields.name = isValid;
      self.updateFieldValidation($(this), isValid);
      self.checkAllFields();
    });
    
    $('#inputEmailAddress').on('input blur', function() {
      const value = $(this).val();
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      const isValid = emailRegex.test(value);
      self.formFields.email = isValid;
      self.updateFieldValidation($(this), isValid);
      self.checkAllFields();
    });
    
    $('#inputPhone').on('input blur', function() {
      const value = $(this).val();
      const phoneRegex = /^0[0-9 ]{10,14}$/;
      const isValid = phoneRegex.test(value);
      self.formFields.phone = isValid;
      self.updateFieldValidation($(this), isValid);
      self.checkAllFields();
      
      // Telefon numarası formatlama - Sadece başa 0 ekle
      if (value.length > 0 && value.charAt(0) !== '0') {
        $(this).val('0' + value);
      }
    });
    
    // Captcha alanı için doğrulama
    $('#captchaAnswer').on('input blur', function() {
      // Eğer captcha bölümü görünür değilse işlem yapma
      if ($('#captchaSection').is(':hidden')) {
        return;
      }
      
      const answerVal = $(this).val().trim();
      const expectedVal = $('#captchaExpected').val().trim();
      
      // Boş değer kontrolü
      if (!answerVal) {
        self.updateFieldValidation($(this), false);
        return;
      }
      
      // Sayısal değere dönüştür
      const answer = parseInt(answerVal);
      const expected = parseInt(expectedVal);
      
      console.log("Captcha kontrolü:", { answer, expected, isValid: answer === expected });
      
      // Sayısal karşılaştırma yap
      const isValid = !isNaN(answer) && !isNaN(expected) && answer === expected;
      self.updateFieldValidation($(this), isValid);
      
      // Doğrulama başarılı olduğunda captcha değerini form gönderilene kadar sakla
      if (isValid) {
        $(this).data('validatedAnswer', answer);
      }
    });
    
    // Form gönderimi
    $('#registerForm').on('submit', function(e) {
      console.log("Register form submit");
      
      // Form doğrulama - tüm alanları kontrol et
      if (!self.checkAllFields(true)) {
        e.preventDefault();
        AuthHelpers.showNotification('warning', 'Uyarı', 'Lütfen tüm alanları doğru şekilde doldurun');
        return false;
      }
      
      // Captcha kontrolü - değerleri tekrar al
      const captchaAnswerInput = $('#captchaAnswer');
      const captchaExpectedInput = $('#captchaExpected');
      
      if (captchaAnswerInput.length && captchaExpectedInput.length) {
        // Önce data-validatedAnswer'ı kontrol et (önceden doğrulanmış değer)
        const validatedAnswer = captchaAnswerInput.data('validatedAnswer');
        const captchaExpected = parseInt(captchaExpectedInput.val().trim());
        
        // Eğer daha önce doğrulanmış bir değer varsa ve bu değer beklenen değerle eşleşiyorsa
        if (!isNaN(validatedAnswer) && !isNaN(captchaExpected) && validatedAnswer === captchaExpected) {
          console.log("Form gönderilirken captcha kontrolü: Önceden doğrulanmış değer kullanılıyor", { 
            validatedAnswer, 
            captchaExpected
          });
          return true; // Form gönderimi devam etsin
        }
        
        // Önceden doğrulanmış değer yoksa veya eşleşmiyorsa, mevcut değeri kontrol et
        const captchaAnswer = parseInt(captchaAnswerInput.val().trim());
        
        console.log("Form gönderilirken captcha kontrolü:", { 
          captchaAnswer, 
          captchaExpected, 
          validatedAnswer,
          rawAnswer: captchaAnswerInput.val(),
          rawExpected: captchaExpectedInput.val()
        });
        
        // Sayısal karşılaştırma yap
        if (isNaN(captchaAnswer) || isNaN(captchaExpected) || captchaAnswer !== captchaExpected) {
          e.preventDefault();
          captchaAnswerInput.removeClass('is-valid').addClass('is-invalid');
          AuthHelpers.showNotification('warning', 'Uyarı', 'Doğrulama kodunu doğru giriniz');
          return false;
        }
      }
    });
    
    // Captcha yenileme butonu
    $('#refreshCaptcha').on('click', function() {
      self.generateMathCaptcha();
    });
  },
  
  // Alan doğrulama durumunu güncelle
  updateFieldValidation: function(field, isValid) {
    if (isValid) {
      field.removeClass('is-invalid').addClass('is-valid');
    } else {
      field.removeClass('is-valid').addClass('is-invalid');
    }
  },
  
  // Tüm alanların doğruluğunu kontrol et
  checkAllFields: function(showNotification = false) {
    const allValid = this.formFields.name && this.formFields.email && this.formFields.phone;
    
    // Tüm alanlar doğruysa captcha bölümünü göster
    if (allValid) {
      // Eğer captcha bölümü zaten görünür değilse göster ve yeni captcha oluştur
      if ($('#captchaSection').is(':hidden')) {
        $('#captchaSection').slideDown(300);
        this.generateMathCaptcha();
      }
    } else {
      $('#captchaSection').slideUp(300);
    }
    
    if (showNotification && !allValid) {
      // Hangi alanların eksik olduğunu belirle
      let missingFields = [];
      if (!this.formFields.name) missingFields.push("Ad Soyad");
      if (!this.formFields.email) missingFields.push("E-posta");
      if (!this.formFields.phone) missingFields.push("Telefon");
      
      AuthHelpers.showNotification('warning', 'Eksik Bilgi', 
        'Lütfen şu alanları doğru şekilde doldurun: ' + missingFields.join(', '));
    }
    
    return allValid;
  },
  
  // CAPTCHA oluşturma fonksiyonu
  generateMathCaptcha: function() {
    const challengeElement = document.getElementById('captchaChallenge');
    const answerElement = document.getElementById('captchaAnswer');
    const expectedElement = document.getElementById('captchaExpected');
    
    if (!challengeElement || !answerElement || !expectedElement) return;
    
    // Rastgele 1-10 arası sayılar
    const num1 = Math.floor(Math.random() * 10) + 1;
    const num2 = Math.floor(Math.random() * 10) + 1;
    
    // Toplama/çıkarma işlemi (çıkarma işleminde negatif sonuç çıkmaması için kontrol)
    let result, question;
    const isAddition = Math.random() > 0.5;
    
    if (isAddition) {
      result = num1 + num2;
      question = `${num1} + ${num2} = ?`;
    } else {
      // Büyük sayıdan küçük sayıyı çıkar
      if (num1 >= num2) {
        result = num1 - num2;
        question = `${num1} - ${num2} = ?`;
      } else {
        result = num2 - num1;
        question = `${num2} - ${num1} = ?`;
      }
    }
    
    // Soruyu ekrana yaz
    challengeElement.textContent = question;
    
    // Beklenen sonucu sakla
    expectedElement.value = result.toString();
    
    // Cevap alanını sıfırla
    answerElement.value = '';
    answerElement.classList.remove('is-valid', 'is-invalid');
    
    // Önceden doğrulanmış değeri temizle
    $(answerElement).removeData('validatedAnswer');
    
    console.log("Yeni captcha oluşturuldu:", { question, result });
  }
};

// Forgot Password sayfası işlevleri
const ForgotPassword = {
  init: function() {
    console.log("Forgot password page loaded");
    this.setupEventListeners();
  },
  
  setupEventListeners: function() {
    // Form gönderimi
    $('#forgotPasswordForm').on('submit', function(e) {
      console.log("Forgot password form submit");
      
      // Form doğrulama
      const email = $('#inputEmailAddress').val();
      
      if (!email) {
        e.preventDefault();
        AuthHelpers.showNotification('warning', 'Uyarı', 'Lütfen e-posta adresinizi girin');
        return false;
      }
    });
  }
};

// Sayfa yüklendiğinde doğru modülü başlat
$(document).ready(function() {
  // Sayfanın hangi auth sayfası olduğunu belirle
  const pageId = $('body').data('page-id');
  
  switch(pageId) {
    case 'login':
      Login.init();
      break;
    case 'register':
      Register.init();
      break;
    case 'forgot-password':
      ForgotPassword.init();
      break;
    case 'reset-password':
      ResetPassword.init();
      break;
  }
}); 