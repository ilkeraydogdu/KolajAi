/**
 * User Orders Management JavaScript
 */

// Global user orders manager
window.UserOrdersManager = {
    
    // Initialize user orders functionality
    init: function() {
        this.initEventListeners();
        this.initFilters();
    },
    
    // Initialize event listeners
    initEventListeners: function() {
        // Filter changes
        document.addEventListener('change', (e) => {
            if (e.target.matches('#statusFilter, #dateRangeFilter')) {
                // Auto-apply filters on change (optional)
                // this.applyFilters();
            }
        });
    },
    
    // Initialize filters from URL
    initFilters: function() {
        const params = new URLSearchParams(window.location.search);
        
        const statusFilter = document.getElementById('statusFilter');
        if (statusFilter && params.get('status')) {
            statusFilter.value = params.get('status');
        }
        
        const dateRangeFilter = document.getElementById('dateRangeFilter');
        if (dateRangeFilter && params.get('date_range')) {
            dateRangeFilter.value = params.get('date_range');
        }
    },
    
    // Apply filters
    applyFilters: function() {
        const params = new URLSearchParams();
        
        const statusFilter = document.getElementById('statusFilter');
        if (statusFilter && statusFilter.value) {
            params.set('status', statusFilter.value);
        }
        
        const dateRangeFilter = document.getElementById('dateRangeFilter');
        if (dateRangeFilter && dateRangeFilter.value) {
            params.set('date_range', dateRangeFilter.value);
        }
        
        // Update URL and reload
        const newUrl = window.location.pathname + (params.toString() ? '?' + params.toString() : '');
        window.location.href = newUrl;
    },
    
    // Toggle order details
    toggleOrderDetails: function(orderId) {
        const detailsDiv = document.getElementById(`order-details-${orderId}`);
        const chevron = document.getElementById(`chevron-${orderId}`);
        
        if (!detailsDiv || !chevron) return;
        
        if (detailsDiv.classList.contains('hidden')) {
            detailsDiv.classList.remove('hidden');
            chevron.classList.add('rotate-180');
        } else {
            detailsDiv.classList.add('hidden');
            chevron.classList.remove('rotate-180');
        }
    },
    
    // Track order
    trackOrder: function(trackingNumber) {
        // You can customize this URL based on your shipping provider
        const trackingUrls = {
            'PTT': `https://gonderitakip.ptt.gov.tr/Track/Verify?code=${trackingNumber}`,
            'MNG': `https://www.mngkargo.com.tr/track?code=${trackingNumber}`,
            'YURTICI': `https://www.yurticikargo.com/tr/online-servisler/gonderi-sorgula?code=${trackingNumber}`,
            'ARAS': `https://kargotakip.araskargo.com.tr/Track/Verify?code=${trackingNumber}`
        };
        
        // Default to PTT if no specific provider is detected
        let trackingUrl = trackingUrls['PTT'];
        
        // Try to detect provider from tracking number format
        if (trackingNumber.startsWith('MNG')) {
            trackingUrl = trackingUrls['MNG'];
        } else if (trackingNumber.startsWith('YK')) {
            trackingUrl = trackingUrls['YURTICI'];
        } else if (trackingNumber.startsWith('ARS')) {
            trackingUrl = trackingUrls['ARAS'];
        }
        
        window.open(trackingUrl, '_blank');
    },
    
    // Review order
    reviewOrder: function(orderId) {
        // Redirect to review page
        window.location.href = `/user/orders/${orderId}/review`;
    },
    
    // Cancel order
    cancelOrder: function(orderId) {
        if (!confirm('Bu siparişi iptal etmek istediğinizden emin misiniz?')) return;
        
        const reason = prompt('İptal nedeni (isteğe bağlı):');
        
        fetch('/api/user/orders/cancel', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: JSON.stringify({
                order_id: orderId,
                reason: reason
            })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('Sipariş başarıyla iptal edildi!');
                window.location.reload();
            } else {
                alert(data.message || 'Sipariş iptal edilemedi.');
            }
        })
        .catch(error => {
            console.error('Cancel order error:', error);
            alert('Bir hata oluştu. Lütfen tekrar deneyin.');
        });
    },
    
    // Reorder items
    reorderItems: function(orderId) {
        if (!confirm('Bu siparişin ürünlerini tekrar sepete eklemek istediğinizden emin misiniz?')) return;
        
        fetch('/api/user/orders/reorder', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: JSON.stringify({ order_id: orderId })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('Ürünler sepete eklendi!');
                if (confirm('Sepete gitmek ister misiniz?')) {
                    window.location.href = '/cart';
                }
            } else {
                alert(data.message || 'Ürünler sepete eklenemedi.');
            }
        })
        .catch(error => {
            console.error('Reorder error:', error);
            alert('Bir hata oluştu. Lütfen tekrar deneyin.');
        });
    },
    
    // Download invoice
    downloadInvoice: function(orderId) {
        const link = document.createElement('a');
        link.href = `/api/user/orders/${orderId}/invoice?format=pdf`;
        link.download = `fatura-${orderId}.pdf`;
        link.target = '_blank';
        link.click();
    },
    
    // Export orders
    exportOrders: function() {
        const format = prompt('Dışa aktarma formatı seçin:\n1. Excel (xlsx)\n2. CSV\n3. PDF\n\nLütfen 1, 2 veya 3 girin:', '1');
        
        let exportFormat = 'xlsx';
        if (format === '2') exportFormat = 'csv';
        else if (format === '3') exportFormat = 'pdf';
        
        // Get current filters
        const params = new URLSearchParams(window.location.search);
        params.set('format', exportFormat);
        
        const link = document.createElement('a');
        link.href = `/api/user/orders/export?${params.toString()}`;
        link.download = `siparislerim.${exportFormat}`;
        link.target = '_blank';
        link.click();
    },
    
    // Return/Refund request
    requestReturn: function(orderId) {
        const reason = prompt('İade nedeni:');
        if (!reason) return;
        
        const description = prompt('Detaylı açıklama (isteğe bağlı):');
        
        fetch('/api/user/orders/return-request', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: JSON.stringify({
                order_id: orderId,
                reason: reason,
                description: description
            })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('İade talebi başarıyla oluşturuldu. En kısa sürede değerlendirilecek.');
                window.location.reload();
            } else {
                alert(data.message || 'İade talebi oluşturulamadı.');
            }
        })
        .catch(error => {
            console.error('Return request error:', error);
            alert('Bir hata oluştu. Lütfen tekrar deneyin.');
        });
    },
    
    // Contact support about order
    contactSupport: function(orderId) {
        const message = prompt('Destek ekibine göndermek istediğiniz mesaj:');
        if (!message) return;
        
        fetch('/api/user/support/ticket', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: JSON.stringify({
                order_id: orderId,
                subject: `Sipariş #${orderId} hakkında`,
                message: message,
                category: 'order_inquiry'
            })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('Destek talebiniz oluşturuldu. Ticket numaranız: ' + data.ticket_number);
            } else {
                alert(data.message || 'Destek talebi oluşturulamadı.');
            }
        })
        .catch(error => {
            console.error('Support ticket error:', error);
            alert('Bir hata oluştu. Lütfen tekrar deneyin.');
        });
    },
    
    // Add to favorites (for reordering)
    addToFavorites: function(orderId) {
        fetch('/api/user/orders/add-to-favorites', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: JSON.stringify({ order_id: orderId })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('Sipariş favorilere eklendi!');
            } else {
                alert(data.message || 'Favorilere eklenemedi.');
            }
        })
        .catch(error => {
            console.error('Add to favorites error:', error);
            alert('Bir hata oluştu.');
        });
    },
    
    // Share order (for gift orders)
    shareOrder: function(orderId) {
        const modal = document.createElement('div');
        modal.className = 'fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4';
        modal.innerHTML = `
            <div class="bg-white rounded-lg max-w-md w-full p-6">
                <div class="text-center">
                    <h3 class="text-lg font-semibold mb-4">Siparişi Paylaş</h3>
                    <p class="text-gray-600 mb-4">Bu siparişi nasıl paylaşmak istiyorsunuz?</p>
                    
                    <div class="space-y-3">
                        <button onclick="UserOrdersManager.shareViaWhatsApp(${orderId})" 
                                class="w-full flex items-center justify-center py-3 px-4 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors">
                            <i class="fab fa-whatsapp mr-2"></i>WhatsApp
                        </button>
                        
                        <button onclick="UserOrdersManager.shareViaEmail(${orderId})" 
                                class="w-full flex items-center justify-center py-3 px-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
                            <i class="fas fa-envelope mr-2"></i>E-posta
                        </button>
                        
                        <button onclick="UserOrdersManager.copyOrderLink(${orderId})" 
                                class="w-full flex items-center justify-center py-3 px-4 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition-colors">
                            <i class="fas fa-copy mr-2"></i>Link Kopyala
                        </button>
                    </div>
                    
                    <button onclick="this.closest('.fixed').remove()" 
                            class="mt-4 w-full px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50">
                        İptal
                    </button>
                </div>
            </div>
        `;
        document.body.appendChild(modal);
    },
    
    // Share via WhatsApp
    shareViaWhatsApp: function(orderId) {
        const message = `Siparişimi paylaşıyorum: ${window.location.origin}/user/orders/${orderId}`;
        const whatsappUrl = `https://wa.me/?text=${encodeURIComponent(message)}`;
        window.open(whatsappUrl, '_blank');
        document.querySelector('.fixed').remove();
    },
    
    // Share via email
    shareViaEmail: function(orderId) {
        const subject = 'Sipariş Paylaşımı';
        const body = `Siparişimi sizinle paylaşıyorum: ${window.location.origin}/user/orders/${orderId}`;
        const mailtoUrl = `mailto:?subject=${encodeURIComponent(subject)}&body=${encodeURIComponent(body)}`;
        window.location.href = mailtoUrl;
        document.querySelector('.fixed').remove();
    },
    
    // Copy order link
    copyOrderLink: function(orderId) {
        const link = `${window.location.origin}/user/orders/${orderId}`;
        navigator.clipboard.writeText(link).then(() => {
            alert('Link kopyalandı!');
            document.querySelector('.fixed').remove();
        }).catch(() => {
            // Fallback for older browsers
            const textArea = document.createElement('textarea');
            textArea.value = link;
            document.body.appendChild(textArea);
            textArea.select();
            document.execCommand('copy');
            document.body.removeChild(textArea);
            alert('Link kopyalandı!');
            document.querySelector('.fixed').remove();
        });
    }
};

// Global functions for backward compatibility
function applyFilters() {
    window.UserOrdersManager.applyFilters();
}

function toggleOrderDetails(orderId) {
    window.UserOrdersManager.toggleOrderDetails(orderId);
}

function trackOrder(trackingNumber) {
    window.UserOrdersManager.trackOrder(trackingNumber);
}

function reviewOrder(orderId) {
    window.UserOrdersManager.reviewOrder(orderId);
}

function cancelOrder(orderId) {
    window.UserOrdersManager.cancelOrder(orderId);
}

function reorderItems(orderId) {
    window.UserOrdersManager.reorderItems(orderId);
}

function downloadInvoice(orderId) {
    window.UserOrdersManager.downloadInvoice(orderId);
}

function exportOrders() {
    window.UserOrdersManager.exportOrders();
}

function requestReturn(orderId) {
    window.UserOrdersManager.requestReturn(orderId);
}

function contactSupport(orderId) {
    window.UserOrdersManager.contactSupport(orderId);
}

function shareOrder(orderId) {
    window.UserOrdersManager.shareOrder(orderId);
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    window.UserOrdersManager.init();
});

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = window.UserOrdersManager;
}