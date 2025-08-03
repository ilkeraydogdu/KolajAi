/**
 * User Profile Management JavaScript
 */

// Global user profile manager
window.UserProfileManager = {
    
    // Initialize user profile functionality
    init: function() {
        this.initEditMode();
        this.initAvatarUpload();
        this.initFormValidation();
        this.initSecurityFeatures();
    },
    
    // Initialize edit mode functionality
    initEditMode: function() {
        const editBtn = document.getElementById('editProfileBtn');
        const cancelBtn = document.getElementById('cancelEditBtn');
        const submitSection = document.getElementById('submitSection');
        const profileFields = document.querySelectorAll('.profile-field');
        
        if (editBtn) {
            editBtn.addEventListener('click', () => {
                this.enableEditMode();
            });
        }
        
        if (cancelBtn) {
            cancelBtn.addEventListener('click', () => {
                this.disableEditMode();
            });
        }
        
        // Store original values for cancel functionality
        this.originalValues = {};
        profileFields.forEach(field => {
            if (field.type === 'checkbox') {
                this.originalValues[field.name] = field.checked;
            } else {
                this.originalValues[field.name] = field.value;
            }
        });
    },
    
    // Enable edit mode
    enableEditMode: function() {
        const editBtn = document.getElementById('editProfileBtn');
        const submitSection = document.getElementById('submitSection');
        const profileFields = document.querySelectorAll('.profile-field');
        
        // Hide edit button, show submit section
        if (editBtn) editBtn.classList.add('hidden');
        if (submitSection) submitSection.classList.remove('hidden');
        
        // Enable all form fields
        profileFields.forEach(field => {
            field.removeAttribute('readonly');
            field.removeAttribute('disabled');
            
            // Add visual indication of editable fields
            if (field.type !== 'file' && field.type !== 'button') {
                field.classList.add('ring-2', 'ring-blue-200');
            }
        });
    },
    
    // Disable edit mode
    disableEditMode: function() {
        const editBtn = document.getElementById('editProfileBtn');
        const submitSection = document.getElementById('submitSection');
        const profileFields = document.querySelectorAll('.profile-field');
        
        // Show edit button, hide submit section
        if (editBtn) editBtn.classList.remove('hidden');
        if (submitSection) submitSection.classList.add('hidden');
        
        // Disable all form fields and restore original values
        profileFields.forEach(field => {
            field.setAttribute('readonly', 'readonly');
            if (field.tagName === 'SELECT') {
                field.setAttribute('disabled', 'disabled');
            }
            
            // Restore original values
            if (this.originalValues.hasOwnProperty(field.name)) {
                if (field.type === 'checkbox') {
                    field.checked = this.originalValues[field.name];
                } else {
                    field.value = this.originalValues[field.name];
                }
            }
            
            // Remove visual indication
            field.classList.remove('ring-2', 'ring-blue-200');
        });
    },
    
    // Initialize avatar upload functionality
    initAvatarUpload: function() {
        const avatarInput = document.getElementById('avatarInput');
        const avatarPreview = document.getElementById('avatarPreview');
        
        if (avatarInput) {
            avatarInput.addEventListener('change', (e) => {
                const file = e.target.files[0];
                if (file) {
                    this.previewAvatar(file, avatarPreview);
                }
            });
        }
    },
    
    // Preview avatar before upload
    previewAvatar: function(file, previewElement) {
        // Validate file
        if (!this.validateAvatarFile(file)) {
            return;
        }
        
        const reader = new FileReader();
        reader.onload = (e) => {
            if (previewElement.tagName === 'IMG') {
                previewElement.src = e.target.result;
            } else {
                // Replace div with img
                const img = document.createElement('img');
                img.id = 'avatarPreview';
                img.src = e.target.result;
                img.alt = 'Avatar';
                img.className = 'w-16 h-16 rounded-full object-cover';
                previewElement.parentNode.replaceChild(img, previewElement);
            }
        };
        reader.readAsDataURL(file);
    },
    
    // Validate avatar file
    validateAvatarFile: function(file) {
        // Check file type
        const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif'];
        if (!allowedTypes.includes(file.type)) {
            alert('Lütfen geçerli bir resim dosyası seçin (JPG, PNG, GIF)');
            return false;
        }
        
        // Check file size (2MB max)
        const maxSize = 2 * 1024 * 1024; // 2MB in bytes
        if (file.size > maxSize) {
            alert('Dosya boyutu 2MB\'dan küçük olmalıdır');
            return false;
        }
        
        return true;
    },
    
    // Initialize form validation
    initFormValidation: function() {
        const profileForm = document.getElementById('profileForm');
        
        if (profileForm) {
            profileForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.submitProfile();
            });
        }
    },
    
    // Submit profile form
    submitProfile: function() {
        const form = document.getElementById('profileForm');
        const submitBtn = form.querySelector('button[type="submit"]');
        
        if (!this.validateProfileForm()) {
            return;
        }
        
        // Show loading state
        const originalText = submitBtn.textContent;
        submitBtn.textContent = 'Kaydediliyor...';
        submitBtn.disabled = true;
        
        const formData = new FormData(form);
        
        fetch('/user/profile', {
            method: 'POST',
            headers: {
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: formData
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('Profil başarıyla güncellendi!');
                this.disableEditMode();
                // Update original values
                this.updateOriginalValues();
            } else {
                alert(data.message || 'Profil güncellenirken bir hata oluştu.');
            }
        })
        .catch(error => {
            console.error('Profile update error:', error);
            alert('Bir hata oluştu. Lütfen tekrar deneyin.');
        })
        .finally(() => {
            submitBtn.textContent = originalText;
            submitBtn.disabled = false;
        });
    },
    
    // Validate profile form
    validateProfileForm: function() {
        const nameField = document.querySelector('input[name="name"]');
        const emailField = document.querySelector('input[name="email"]');
        
        // Name validation
        if (!nameField.value.trim()) {
            alert('Ad Soyad alanı boş bırakılamaz.');
            nameField.focus();
            return false;
        }
        
        // Email validation
        if (!emailField.value.trim()) {
            alert('E-posta alanı boş bırakılamaz.');
            emailField.focus();
            return false;
        }
        
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(emailField.value)) {
            alert('Geçerli bir e-posta adresi girin.');
            emailField.focus();
            return false;
        }
        
        // Phone validation (if provided)
        const phoneField = document.querySelector('input[name="phone"]');
        if (phoneField.value.trim()) {
            const phoneRegex = /^[0-9+\-\s\(\)]+$/;
            if (!phoneRegex.test(phoneField.value)) {
                alert('Geçerli bir telefon numarası girin.');
                phoneField.focus();
                return false;
            }
        }
        
        return true;
    },
    
    // Update original values after successful save
    updateOriginalValues: function() {
        const profileFields = document.querySelectorAll('.profile-field');
        profileFields.forEach(field => {
            if (field.type === 'checkbox') {
                this.originalValues[field.name] = field.checked;
            } else {
                this.originalValues[field.name] = field.value;
            }
        });
    },
    
    // Initialize security features
    initSecurityFeatures: function() {
        // These would be implemented as needed
        window.changePassword = this.changePassword.bind(this);
        window.enable2FA = this.enable2FA.bind(this);
        window.viewLoginHistory = this.viewLoginHistory.bind(this);
        window.exportData = this.exportData.bind(this);
        window.deleteAccount = this.deleteAccount.bind(this);
        window.verifyEmail = this.verifyEmail.bind(this);
    },
    
    // Change password
    changePassword: function() {
        const currentPassword = prompt('Mevcut şifrenizi girin:');
        if (!currentPassword) return;
        
        const newPassword = prompt('Yeni şifrenizi girin:');
        if (!newPassword) return;
        
        const confirmPassword = prompt('Yeni şifrenizi tekrar girin:');
        if (newPassword !== confirmPassword) {
            alert('Şifreler eşleşmiyor!');
            return;
        }
        
        if (newPassword.length < 8) {
            alert('Şifre en az 8 karakter olmalıdır!');
            return;
        }
        
        fetch('/api/user/change-password', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: JSON.stringify({
                current_password: currentPassword,
                new_password: newPassword,
                confirm_password: confirmPassword
            })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('Şifreniz başarıyla değiştirildi!');
            } else {
                alert(data.message || 'Şifre değiştirilemedi.');
            }
        })
        .catch(error => {
            console.error('Password change error:', error);
            alert('Bir hata oluştu.');
        });
    },
    
    // Enable 2FA
    enable2FA: function() {
        fetch('/api/user/2fa/enable', {
            method: 'POST',
            headers: {
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            }
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                this.show2FASetup(data.qr_code, data.secret);
            } else {
                alert(data.message || '2FA etkinleştirilemedi.');
            }
        })
        .catch(error => {
            console.error('2FA enable error:', error);
            alert('Bir hata oluştu.');
        });
    },
    
    // Show 2FA setup modal
    show2FASetup: function(qrCode, secret) {
        const modal = document.createElement('div');
        modal.className = 'fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4';
        modal.innerHTML = `
            <div class="bg-white rounded-lg max-w-md w-full p-6">
                <div class="text-center">
                    <h3 class="text-lg font-semibold mb-4">İki Faktörlü Kimlik Doğrulama</h3>
                    <p class="text-gray-600 mb-4">QR kodunu Google Authenticator veya benzer bir uygulamayla tarayın:</p>
                    <div class="mb-4">
                        <img src="data:image/png;base64,${qrCode}" alt="QR Code" class="mx-auto">
                    </div>
                    <p class="text-sm text-gray-500 mb-4">Manuel giriş için kod: <code class="bg-gray-100 px-2 py-1 rounded">${secret}</code></p>
                    <input type="text" id="verificationCode" placeholder="6 haneli kodu girin" 
                           class="w-full px-3 py-2 border border-gray-300 rounded-lg mb-4">
                    <div class="flex gap-3">
                        <button onclick="this.closest('.fixed').remove()" 
                                class="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50">
                            İptal
                        </button>
                        <button onclick="UserProfileManager.verify2FA()" 
                                class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
                            Doğrula
                        </button>
                    </div>
                </div>
            </div>
        `;
        document.body.appendChild(modal);
    },
    
    // Verify 2FA setup
    verify2FA: function() {
        const code = document.getElementById('verificationCode').value;
        if (!code) {
            alert('Lütfen doğrulama kodunu girin.');
            return;
        }
        
        fetch('/api/user/2fa/verify', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: JSON.stringify({ code: code })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('İki faktörlü kimlik doğrulama başarıyla etkinleştirildi!');
                document.querySelector('.fixed').remove();
                window.location.reload();
            } else {
                alert(data.message || 'Doğrulama kodu geçersiz.');
            }
        })
        .catch(error => {
            console.error('2FA verification error:', error);
            alert('Bir hata oluştu.');
        });
    },
    
    // View login history
    viewLoginHistory: function() {
        fetch('/api/user/login-history')
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                this.showLoginHistoryModal(data.history);
            } else {
                alert('Giriş geçmişi yüklenemedi.');
            }
        })
        .catch(error => {
            console.error('Login history error:', error);
            alert('Bir hata oluştu.');
        });
    },
    
    // Show login history modal
    showLoginHistoryModal: function(history) {
        const modal = document.createElement('div');
        modal.className = 'fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4';
        
        const historyHTML = history.map(entry => `
            <div class="flex justify-between items-center py-3 border-b border-gray-100">
                <div>
                    <div class="font-medium">${entry.location || 'Bilinmeyen Konum'}</div>
                    <div class="text-sm text-gray-500">${entry.user_agent || 'Bilinmeyen Cihaz'}</div>
                </div>
                <div class="text-right">
                    <div class="text-sm">${new Date(entry.created_at).toLocaleDateString('tr-TR')}</div>
                    <div class="text-xs text-gray-500">${new Date(entry.created_at).toLocaleTimeString('tr-TR')}</div>
                </div>
            </div>
        `).join('');
        
        modal.innerHTML = `
            <div class="bg-white rounded-lg max-w-2xl w-full max-h-96 overflow-y-auto">
                <div class="p-6 border-b">
                    <h3 class="text-lg font-semibold">Giriş Geçmişi</h3>
                </div>
                <div class="p-6">
                    ${history.length > 0 ? historyHTML : '<p class="text-gray-500 text-center">Giriş geçmişi bulunamadı.</p>'}
                </div>
                <div class="p-6 border-t">
                    <button onclick="this.closest('.fixed').remove()" 
                            class="w-full px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700">
                        Kapat
                    </button>
                </div>
            </div>
        `;
        document.body.appendChild(modal);
    },
    
    // Export user data
    exportData: function() {
        if (!confirm('Tüm verilerinizi dışa aktarmak istediğinizden emin misiniz?')) return;
        
        const link = document.createElement('a');
        link.href = '/api/user/export-data';
        link.download = 'user-data.json';
        link.click();
    },
    
    // Delete account
    deleteAccount: function() {
        const confirmation = prompt('Hesabınızı silmek için "SİL" yazın:');
        if (confirmation !== 'SİL') {
            alert('İşlem iptal edildi.');
            return;
        }
        
        if (!confirm('Bu işlem geri alınamaz! Hesabınızı kalıcı olarak silmek istediğinizden emin misiniz?')) return;
        
        fetch('/api/user/delete-account', {
            method: 'DELETE',
            headers: {
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            }
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('Hesabınız başarıyla silindi. Yönlendiriliyorsunuz...');
                window.location.href = '/';
            } else {
                alert(data.message || 'Hesap silinemedi.');
            }
        })
        .catch(error => {
            console.error('Account deletion error:', error);
            alert('Bir hata oluştu.');
        });
    },
    
    // Verify email
    verifyEmail: function() {
        fetch('/api/user/verify-email', {
            method: 'POST',
            headers: {
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            }
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('Doğrulama e-postası gönderildi! Lütfen e-posta adresinizi kontrol edin.');
            } else {
                alert(data.message || 'E-posta gönderilemedi.');
            }
        })
        .catch(error => {
            console.error('Email verification error:', error);
            alert('Bir hata oluştu.');
        });
    }
};

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    window.UserProfileManager.init();
});

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = window.UserProfileManager;
}