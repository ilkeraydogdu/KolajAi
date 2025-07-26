/**
 * KolajAI Form Doğrulama Sistemi
 * Tüm formlar için merkezi doğrulama işlemlerini yönetir
 */

class FormValidator {
  constructor() {
    this.forms = {};
    this.init();
  }

  /**
   * Sayfa yüklendiğinde formları başlat
   */
  init() {
    // Sayfadaki tüm data-validate formlarını bul
    document.querySelectorAll('form[data-validate]').forEach(form => {
      this.initializeForm(form);
    });
  }

  /**
   * Form doğrulama sistemini başlat
   * @param {HTMLFormElement} form - Doğrulanacak form elementi
   */
  initializeForm(form) {
    const formId = form.id;
    
    // Form içindeki tüm input alanlarını işle
    form.querySelectorAll('input, select, textarea').forEach(field => {
      // data-validation attributelarını kontrol et
      if (field.dataset.validate) {
        this.setupFieldValidation(field);
      }
      
      // HTML5 required, pattern, minlength gibi özellikleri de kontrol et
      if (field.required || field.pattern || field.minLength || field.maxLength) {
        this.setupFieldValidation(field);
      }
    });
    
    // Form gönderimini kontrol et
    form.addEventListener('submit', (e) => {
      let isValid = this.validateForm(form);
      if (!isValid) {
        e.preventDefault();
      }
    });
  }
  
  /**
   * Input alanı için doğrulama olaylarını ayarla
   * @param {HTMLElement} field - Doğrulanacak form alanı
   */
  setupFieldValidation(field) {
    // Input değeri değiştiğinde doğrula
    field.addEventListener('input', () => {
      this.validateField(field);
    });
    
    // Alan kaybettiğinde doğrula (blur)
    field.addEventListener('blur', () => {
      this.validateField(field);
    });
  }
  
  /**
   * Bir form alanını doğrula
   * @param {HTMLElement} field - Doğrulanacak form alanı
   * @returns {boolean} - Doğrulama sonucu
   */
  validateField(field) {
    let isValid = true;
    let errorMessage = '';
    
    // Boş değer kontrolü
    if (field.required && !field.value.trim()) {
      isValid = false;
      errorMessage = field.dataset.errorRequired || 'Bu alan zorunludur';
    }
    
    // Minimum uzunluk kontrolü
    else if (field.minLength && field.value.length < field.minLength) {
      isValid = false;
      errorMessage = field.dataset.errorMinlength || `En az ${field.minLength} karakter gereklidir`;
    }
    
    // Maksimum uzunluk kontrolü
    else if (field.getAttribute('maxlength') && field.value.length > parseInt(field.getAttribute('maxlength'))) {
      isValid = false;
      errorMessage = field.dataset.errorMaxlength || `En fazla ${field.getAttribute('maxlength')} karakter olmalıdır`;
    }
    
    // Pattern kontrolü
    else if (field.pattern && field.value) {
      const regex = new RegExp(field.pattern);
      if (!regex.test(field.value)) {
        isValid = false;
        errorMessage = field.dataset.errorPattern || 'Geçersiz format';
      }
    }
    
    // Özel doğrulama kuralları
    else if (field.dataset.validate) {
      switch (field.dataset.validate) {
        case 'email':
          const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
          if (!emailRegex.test(field.value)) {
            isValid = false;
            errorMessage = field.dataset.errorEmail || 'Geçerli bir e-posta adresi giriniz';
          }
          break;
          
        case 'phone':
          const phoneRegex = /^0[0-9 ]{10,14}$/;
          if (!phoneRegex.test(field.value)) {
            isValid = false;
            errorMessage = field.dataset.errorPhone || 'Telefon numarası 0 ile başlamalıdır';
          }
          break;
      }
    }
    
    // Captcha alanı için özel kontrol - ID'ye göre kontrol et
    if (field.id === 'captchaAnswer') {
      // Captcha doğrulamasını auth.js'deki Register modülüne bırak
      return true;
    }
    
    // Sonucu göster
    this.showValidationResult(field, isValid, errorMessage);
    return isValid;
  }
  
  /**
   * Doğrulama sonucunu göster
   * @param {HTMLElement} field - Doğrulanan alan
   * @param {boolean} isValid - Doğrulama sonucu
   * @param {string} errorMessage - Hata mesajı
   */
  showValidationResult(field, isValid, errorMessage) {
    // Bootstrap sınıflarını kullan
    if (isValid) {
      field.classList.remove('is-invalid');
      field.classList.add('is-valid');
    } else {
      field.classList.remove('is-valid');
      field.classList.add('is-invalid');
    }
    
    // Geri bildirim elementini bul veya oluştur
    let feedback;
    if (isValid) {
      feedback = field.parentElement.querySelector('.valid-feedback');
      if (!feedback) {
        feedback = document.createElement('div');
        feedback.className = 'valid-feedback';
        field.parentElement.appendChild(feedback);
      }
      feedback.textContent = field.dataset.successMessage || 'Geçerli!';
    } else {
      feedback = field.parentElement.querySelector('.invalid-feedback');
      if (!feedback) {
        feedback = document.createElement('div');
        feedback.className = 'invalid-feedback';
        field.parentElement.appendChild(feedback);
      }
      feedback.textContent = errorMessage;
    }
  }
  
  /**
   * Tüm formu doğrula
   * @param {HTMLFormElement} form - Doğrulanacak form
   * @returns {boolean} - Doğrulama sonucu
   */
  validateForm(form) {
    let isValid = true;
    
    // Tüm gerekli alanları doğrula
    form.querySelectorAll('input, select, textarea').forEach(field => {
      if (!this.validateField(field)) {
        isValid = false;
      }
    });
    
    // Submit butonunu güncelle
    const submitButton = form.querySelector('button[type="submit"]');
    if (submitButton) {
      submitButton.disabled = !isValid;
    }
    
    return isValid;
  }
  
  /**
   * CAPTCHA oluştur
   * @param {string} questionElementId - Soru gösterilecek element ID
   * @param {string} answerElementId - Cevap girilecek element ID
   * @param {string} expectedElementId - Beklenen cevap saklanacak element ID
   */
  generateCaptcha(questionElementId, answerElementId, expectedElementId) {
    const questionElement = document.getElementById(questionElementId);
    const answerElement = document.getElementById(answerElementId);
    const expectedElement = document.getElementById(expectedElementId);
    
    if (!questionElement || !answerElement || !expectedElement) return;
    
    const num1 = Math.floor(Math.random() * 10) + 1;
    const num2 = Math.floor(Math.random() * 10) + 1;
    const operator = Math.random() > 0.5 ? '+' : '-';
    let result;
    
    if (operator === '+') {
      result = num1 + num2;
      questionElement.textContent = `${num1} + ${num2} = ?`;
    } else {
      // Negatif sonuçları önle
      if (num1 >= num2) {
        result = num1 - num2;
        questionElement.textContent = `${num1} - ${num2} = ?`;
      } else {
        result = num2 - num1;
        questionElement.textContent = `${num2} - ${num1} = ?`;
      }
    }
    
    // Beklenen sonucu sakla
    expectedElement.value = result;
    
    // Cevap alanını sıfırla
    answerElement.value = '';
    answerElement.classList.remove('is-valid', 'is-invalid');
    
    // data-expected attributu da ekle
    answerElement.dataset.expected = result.toString();
  }
}

// Sayfa yüklendiğinde doğrulayıcıyı başlat
document.addEventListener('DOMContentLoaded', () => {
  window.formValidator = new FormValidator();
  
  // CAPTCHA'yı başlat
  const refreshButton = document.getElementById('refreshCaptcha');
  if (refreshButton) {
    // İlk yüklemede CAPTCHA oluştur
    window.formValidator.generateCaptcha('captchaQuestion', 'captchaAnswer', 'captchaExpected');
    
    // Yenile butonuna tıklandığında CAPTCHA yenile
    refreshButton.addEventListener('click', () => {
      window.formValidator.generateCaptcha('captchaQuestion', 'captchaAnswer', 'captchaExpected');
    });
  }
}); 