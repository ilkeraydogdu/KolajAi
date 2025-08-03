/**
 * User Addresses Management JavaScript
 */

// Global user addresses manager
window.UserAddressesManager = {
    currentEditingId: null,
    
    // Initialize user addresses functionality
    init: function() {
        this.initEventListeners();
    },
    
    // Initialize event listeners
    initEventListeners: function() {
        // Modal close on outside click
        document.addEventListener('click', (e) => {
            if (e.target.matches('#addressModal, #deleteModal')) {
                this.closeAddressModal();
                this.closeDeleteModal();
            }
        });
        
        // Escape key to close modals
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.closeAddressModal();
                this.closeDeleteModal();
            }
        });
    },
    
    // Show add address modal
    showAddAddressModal: function() {
        this.currentEditingId = null;
        document.getElementById('modalTitle').textContent = 'Yeni Adres Ekle';
        document.getElementById('formAction').value = 'add';
        this.clearForm();
        document.getElementById('addressModal').classList.remove('hidden');
    },
    
    // Close address modal
    closeAddressModal: function() {
        document.getElementById('addressModal').classList.add('hidden');
        this.clearForm();
        this.currentEditingId = null;
    },
    
    // Clear form
    clearForm: function() {
        const form = document.getElementById('addressForm');
        form.reset();
        document.getElementById('addressId').value = '';
    },
    
    // Edit address
    editAddress: function(addressId) {
        this.currentEditingId = addressId;
        document.getElementById('modalTitle').textContent = 'Adresi Düzenle';
        document.getElementById('formAction').value = 'edit';
        document.getElementById('addressId').value = addressId;
        
        // Fetch address data
        fetch(`/api/user/addresses/${addressId}`)
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                this.fillForm(data.address);
                document.getElementById('addressModal').classList.remove('hidden');
            } else {
                alert(data.message || 'Adres bilgileri alınamadı.');
            }
        })
        .catch(error => {
            console.error('Edit address error:', error);
            alert('Bir hata oluştu. Lütfen tekrar deneyin.');
        });
    },
    
    // Fill form with address data
    fillForm: function(address) {
        document.getElementById('title').value = address.title || '';
        document.getElementById('fullName').value = address.full_name || '';
        document.getElementById('phone').value = address.phone || '';
        document.getElementById('addressLine1').value = address.address_line_1 || '';
        document.getElementById('addressLine2').value = address.address_line_2 || '';
        document.getElementById('city').value = address.city || '';
        document.getElementById('state').value = address.state || '';
        document.getElementById('postalCode').value = address.postal_code || '';
        document.getElementById('country').value = address.country || 'Türkiye';
        document.getElementById('isDefault').checked = address.is_default || false;
    },
    
    // Save address
    saveAddress: function() {
        const form = document.getElementById('addressForm');
        
        if (!this.validateForm()) {
            return;
        }
        
        const formData = new FormData(form);
        const isEdit = this.currentEditingId !== null;
        
        // Show loading state
        const saveBtn = document.querySelector('#addressModal button[onclick="saveAddress()"]');
        const originalText = saveBtn.innerHTML;
        saveBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-2"></i>Kaydediliyor...';
        saveBtn.disabled = true;
        
        fetch('/user/addresses', {
            method: 'POST',
            headers: {
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: formData
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert(isEdit ? 'Adres başarıyla güncellendi!' : 'Adres başarıyla eklendi!');
                window.location.reload();
            } else {
                alert(data.message || 'Adres kaydedilemedi.');
            }
        })
        .catch(error => {
            console.error('Save address error:', error);
            alert('Bir hata oluştu. Lütfen tekrar deneyin.');
        })
        .finally(() => {
            saveBtn.innerHTML = originalText;
            saveBtn.disabled = false;
        });
    },
    
    // Validate form
    validateForm: function() {
        const title = document.getElementById('title').value.trim();
        const fullName = document.getElementById('fullName').value.trim();
        const addressLine1 = document.getElementById('addressLine1').value.trim();
        const city = document.getElementById('city').value.trim();
        const state = document.getElementById('state').value.trim();
        const country = document.getElementById('country').value.trim();
        
        if (!title) {
            alert('Adres başlığı gereklidir.');
            document.getElementById('title').focus();
            return false;
        }
        
        if (!fullName) {
            alert('Ad Soyad gereklidir.');
            document.getElementById('fullName').focus();
            return false;
        }
        
        if (!addressLine1) {
            alert('Adres satır 1 gereklidir.');
            document.getElementById('addressLine1').focus();
            return false;
        }
        
        if (!city) {
            alert('Şehir gereklidir.');
            document.getElementById('city').focus();
            return false;
        }
        
        if (!state) {
            alert('İlçe gereklidir.');
            document.getElementById('state').focus();
            return false;
        }
        
        if (!country) {
            alert('Ülke gereklidir.');
            document.getElementById('country').focus();
            return false;
        }
        
        // Phone validation (if provided)
        const phone = document.getElementById('phone').value.trim();
        if (phone) {
            const phoneRegex = /^[0-9+\-\s\(\)]+$/;
            if (!phoneRegex.test(phone)) {
                alert('Geçerli bir telefon numarası girin.');
                document.getElementById('phone').focus();
                return false;
            }
        }
        
        return true;
    },
    
    // Delete address
    deleteAddress: function(addressId) {
        this.currentDeletingId = addressId;
        document.getElementById('deleteModal').classList.remove('hidden');
    },
    
    // Close delete modal
    closeDeleteModal: function() {
        document.getElementById('deleteModal').classList.add('hidden');
        this.currentDeletingId = null;
    },
    
    // Confirm delete address
    confirmDeleteAddress: function() {
        if (!this.currentDeletingId) return;
        
        const deleteBtn = document.querySelector('#deleteModal button[onclick="confirmDeleteAddress()"]');
        const originalText = deleteBtn.innerHTML;
        deleteBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-2"></i>Siliniyor...';
        deleteBtn.disabled = true;
        
        const formData = new FormData();
        formData.append('action', 'delete');
        formData.append('address_id', this.currentDeletingId);
        
        fetch('/user/addresses', {
            method: 'POST',
            headers: {
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: formData
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('Adres başarıyla silindi!');
                window.location.reload();
            } else {
                alert(data.message || 'Adres silinemedi.');
            }
        })
        .catch(error => {
            console.error('Delete address error:', error);
            alert('Bir hata oluştu. Lütfen tekrar deneyin.');
        })
        .finally(() => {
            deleteBtn.innerHTML = originalText;
            deleteBtn.disabled = false;
            this.closeDeleteModal();
        });
    },
    
    // Set default address
    setDefaultAddress: function(addressId) {
        if (!confirm('Bu adresi varsayılan adres olarak ayarlamak istediğinizden emin misiniz?')) return;
        
        const formData = new FormData();
        formData.append('action', 'set_default');
        formData.append('address_id', addressId);
        
        fetch('/user/addresses', {
            method: 'POST',
            headers: {
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: formData
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('Varsayılan adres güncellendi!');
                window.location.reload();
            } else {
                alert(data.message || 'Varsayılan adres ayarlanamadı.');
            }
        })
        .catch(error => {
            console.error('Set default address error:', error);
            alert('Bir hata oluştu. Lütfen tekrar deneyin.');
        });
    },
    
    // Validate address with external service
    validateAddress: function(addressId) {
        fetch(`/api/user/addresses/${addressId}/validate`, {
            method: 'POST',
            headers: {
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            }
        })
        .then(response => response.json())
        .then(data => {
            if (data.valid) {
                alert('Adres geçerli ve doğrulandı!');
            } else {
                alert(data.message || 'Adres doğrulanamadı.');
            }
        })
        .catch(error => {
            console.error('Validate address error:', error);
            alert('Doğrulama sırasında bir hata oluştu.');
        });
    },
    
    // Export addresses
    exportAddresses: function() {
        const format = prompt('Dışa aktarma formatı seçin:\n1. PDF\n2. Excel\n3. CSV\n\nLütfen 1, 2 veya 3 girin:', '1');
        
        let exportFormat = 'pdf';
        if (format === '2') exportFormat = 'xlsx';
        else if (format === '3') exportFormat = 'csv';
        
        const link = document.createElement('a');
        link.href = `/api/user/addresses/export?format=${exportFormat}`;
        link.download = `adreslerim.${exportFormat}`;
        link.target = '_blank';
        link.click();
    },
    
    // Import addresses
    importAddresses: function() {
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = '.csv,.xlsx,.json';
        input.onchange = (e) => {
            const file = e.target.files[0];
            if (!file) return;
            
            const formData = new FormData();
            formData.append('file', file);
            
            fetch('/api/user/addresses/import', {
                method: 'POST',
                headers: {
                    'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
                },
                body: formData
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert(`${data.imported_count} adres başarıyla içe aktarıldı!`);
                    window.location.reload();
                } else {
                    alert(data.message || 'İçe aktarma başarısız.');
                }
            })
            .catch(error => {
                console.error('Import addresses error:', error);
                alert('İçe aktarma sırasında bir hata oluştu.');
            });
        };
        input.click();
    },
    
    // Copy address to clipboard
    copyAddress: function(addressId) {
        fetch(`/api/user/addresses/${addressId}`)
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                const address = data.address;
                const addressText = `${address.full_name}\n${address.address_line_1}\n${address.address_line_2 ? address.address_line_2 + '\n' : ''}${address.city}, ${address.state} ${address.postal_code}\n${address.country}${address.phone ? '\nTel: ' + address.phone : ''}`;
                
                navigator.clipboard.writeText(addressText).then(() => {
                    alert('Adres panoya kopyalandı!');
                }).catch(() => {
                    // Fallback for older browsers
                    const textArea = document.createElement('textarea');
                    textArea.value = addressText;
                    document.body.appendChild(textArea);
                    textArea.select();
                    document.execCommand('copy');
                    document.body.removeChild(textArea);
                    alert('Adres panoya kopyalandı!');
                });
            } else {
                alert('Adres kopyalanamadı.');
            }
        })
        .catch(error => {
            console.error('Copy address error:', error);
            alert('Bir hata oluştu.');
        });
    },
    
    // Share address
    shareAddress: function(addressId) {
        fetch(`/api/user/addresses/${addressId}`)
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                const address = data.address;
                const addressText = `${address.full_name}\n${address.address_line_1}\n${address.address_line_2 ? address.address_line_2 + '\n' : ''}${address.city}, ${address.state} ${address.postal_code}\n${address.country}${address.phone ? '\nTel: ' + address.phone : ''}`;
                
                if (navigator.share) {
                    navigator.share({
                        title: 'Adres Paylaşımı',
                        text: addressText
                    });
                } else {
                    // Fallback - copy to clipboard
                    this.copyAddress(addressId);
                }
            } else {
                alert('Adres paylaşılamadı.');
            }
        })
        .catch(error => {
            console.error('Share address error:', error);
            alert('Bir hata oluştu.');
        });
    },
    
    // Use address for current order
    useForOrder: function(addressId) {
        sessionStorage.setItem('selectedAddressId', addressId);
        alert('Adres seçildi! Sepete gidebilirsiniz.');
        
        // Redirect to cart if user wants
        if (confirm('Sepete gitmek ister misiniz?')) {
            window.location.href = '/cart';
        }
    },
    
    // Get directions to address
    getDirections: function(addressId) {
        fetch(`/api/user/addresses/${addressId}`)
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                const address = data.address;
                const addressString = `${address.address_line_1}, ${address.city}, ${address.state}, ${address.country}`;
                const mapsUrl = `https://www.google.com/maps/dir/?api=1&destination=${encodeURIComponent(addressString)}`;
                window.open(mapsUrl, '_blank');
            } else {
                alert('Adres bilgileri alınamadı.');
            }
        })
        .catch(error => {
            console.error('Get directions error:', error);
            alert('Bir hata oluştu.');
        });
    },
    
    // Search addresses
    searchAddresses: function(query) {
        const addressCards = document.querySelectorAll('.bg-white.rounded-lg.shadow-md');
        
        addressCards.forEach(card => {
            const title = card.querySelector('h3')?.textContent?.toLowerCase() || '';
            const fullName = card.querySelector('.font-medium')?.textContent?.toLowerCase() || '';
            const addressText = card.querySelector('.text-sm.text-gray-700')?.textContent?.toLowerCase() || '';
            
            const matches = title.includes(query.toLowerCase()) || 
                           fullName.includes(query.toLowerCase()) || 
                           addressText.includes(query.toLowerCase());
            
            card.style.display = matches ? 'block' : 'none';
        });
    },
    
    // Bulk operations
    selectAllAddresses: function() {
        const checkboxes = document.querySelectorAll('.address-checkbox');
        checkboxes.forEach(cb => cb.checked = true);
        this.updateBulkActions();
    },
    
    unselectAllAddresses: function() {
        const checkboxes = document.querySelectorAll('.address-checkbox');
        checkboxes.forEach(cb => cb.checked = false);
        this.updateBulkActions();
    },
    
    updateBulkActions: function() {
        const selectedCount = document.querySelectorAll('.address-checkbox:checked').length;
        const bulkActions = document.getElementById('bulkActions');
        
        if (bulkActions) {
            bulkActions.style.display = selectedCount > 0 ? 'block' : 'none';
            const countSpan = bulkActions.querySelector('.selected-count');
            if (countSpan) {
                countSpan.textContent = selectedCount;
            }
        }
    },
    
    bulkDelete: function() {
        const selectedCheckboxes = document.querySelectorAll('.address-checkbox:checked');
        const selectedIds = Array.from(selectedCheckboxes).map(cb => cb.value);
        
        if (selectedIds.length === 0) {
            alert('Lütfen silmek için adres seçin.');
            return;
        }
        
        if (!confirm(`${selectedIds.length} adresi silmek istediğinizden emin misiniz?`)) return;
        
        Promise.all(selectedIds.map(id => 
            fetch(`/api/user/addresses/${id}`, { method: 'DELETE' })
        ))
        .then(responses => Promise.all(responses.map(r => r.json())))
        .then(results => {
            const successCount = results.filter(r => r.success).length;
            alert(`${successCount} adres başarıyla silindi.`);
            window.location.reload();
        })
        .catch(error => {
            console.error('Bulk delete error:', error);
            alert('Silme işlemi sırasında bir hata oluştu.');
        });
    }
};

// Global functions for backward compatibility
function showAddAddressModal() {
    window.UserAddressesManager.showAddAddressModal();
}

function closeAddressModal() {
    window.UserAddressesManager.closeAddressModal();
}

function editAddress(addressId) {
    window.UserAddressesManager.editAddress(addressId);
}

function saveAddress() {
    window.UserAddressesManager.saveAddress();
}

function deleteAddress(addressId) {
    window.UserAddressesManager.deleteAddress(addressId);
}

function closeDeleteModal() {
    window.UserAddressesManager.closeDeleteModal();
}

function confirmDeleteAddress() {
    window.UserAddressesManager.confirmDeleteAddress();
}

function setDefaultAddress(addressId) {
    window.UserAddressesManager.setDefaultAddress(addressId);
}

function validateAddress(addressId) {
    window.UserAddressesManager.validateAddress(addressId);
}

function exportAddresses() {
    window.UserAddressesManager.exportAddresses();
}

function importAddresses() {
    window.UserAddressesManager.importAddresses();
}

function copyAddress(addressId) {
    window.UserAddressesManager.copyAddress(addressId);
}

function shareAddress(addressId) {
    window.UserAddressesManager.shareAddress(addressId);
}

function useForOrder(addressId) {
    window.UserAddressesManager.useForOrder(addressId);
}

function getDirections(addressId) {
    window.UserAddressesManager.getDirections(addressId);
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    window.UserAddressesManager.init();
});

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = window.UserAddressesManager;
}