/**
 * Admin Orders Management JavaScript
 */

// Global admin orders manager
window.AdminOrdersManager = {
    
    // Selected orders for bulk operations
    selectedOrders: new Set(),
    
    // Initialize admin orders functionality
    init: function() {
        this.initEventListeners();
        this.initFilters();
        this.initBulkSelection();
        this.initAutoRefresh();
    },
    
    // Initialize event listeners
    initEventListeners: function() {
        // Search input with debounce
        const searchInput = document.getElementById('searchInput');
        if (searchInput) {
            let searchTimeout;
            searchInput.addEventListener('input', (e) => {
                clearTimeout(searchTimeout);
                searchTimeout = setTimeout(() => {
                    this.applyFilters();
                }, 500);
            });
        }
        
        // Filter changes
        document.addEventListener('change', (e) => {
            if (e.target.matches('#statusFilter, #dateFilter, #sortFilter')) {
                this.applyFilters();
            }
        });
        
        // Modal close buttons
        document.addEventListener('click', (e) => {
            if (e.target.matches('#closeOrderDetailsModal')) {
                this.closeOrderDetailsModal();
            }
            if (e.target.matches('#closeStatusUpdateModal, #cancelStatusUpdate')) {
                this.closeStatusUpdateModal();
            }
        });
        
        // Status update form
        const statusUpdateForm = document.getElementById('statusUpdateForm');
        if (statusUpdateForm) {
            statusUpdateForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.submitStatusUpdate(statusUpdateForm);
            });
        }
        
        // Click outside dropdowns to close
        document.addEventListener('click', (e) => {
            if (!e.target.closest('.relative')) {
                document.querySelectorAll('[id^="dropdown-"]').forEach(dropdown => {
                    dropdown.classList.add('hidden');
                });
            }
        });
    },
    
    // Initialize filters from URL parameters
    initFilters: function() {
        const params = new URLSearchParams(window.location.search);
        
        const searchInput = document.getElementById('searchInput');
        if (searchInput && params.get('search')) {
            searchInput.value = params.get('search');
        }
        
        const statusFilter = document.getElementById('statusFilter');
        if (statusFilter && params.get('status')) {
            statusFilter.value = params.get('status');
        }
        
        const dateFilter = document.getElementById('dateFilter');
        if (dateFilter && params.get('date')) {
            dateFilter.value = params.get('date');
        }
        
        const sortFilter = document.getElementById('sortFilter');
        if (sortFilter && params.get('sort')) {
            sortFilter.value = params.get('sort');
        }
    },
    
    // Initialize bulk selection functionality
    initBulkSelection: function() {
        const selectAllCheckbox = document.getElementById('selectAll');
        if (selectAllCheckbox) {
            selectAllCheckbox.addEventListener('change', (e) => {
                this.toggleSelectAll(e.target.checked);
            });
        }
        
        // Individual checkboxes
        document.addEventListener('change', (e) => {
            if (e.target.matches('.order-checkbox')) {
                this.toggleOrderSelection(e.target.value, e.target.checked);
            }
        });
    },
    
    // Initialize auto-refresh
    initAutoRefresh: function() {
        // Auto-refresh every 30 seconds
        setInterval(() => {
            this.refreshOrderStatuses();
        }, 30000);
    },
    
    // Apply filters
    applyFilters: function() {
        const params = new URLSearchParams();
        
        const searchInput = document.getElementById('searchInput');
        if (searchInput && searchInput.value.trim()) {
            params.set('search', searchInput.value.trim());
        }
        
        const statusFilter = document.getElementById('statusFilter');
        if (statusFilter && statusFilter.value) {
            params.set('status', statusFilter.value);
        }
        
        const dateFilter = document.getElementById('dateFilter');
        if (dateFilter && dateFilter.value) {
            params.set('date', dateFilter.value);
        }
        
        const sortFilter = document.getElementById('sortFilter');
        if (sortFilter && sortFilter.value) {
            params.set('sort', sortFilter.value);
        }
        
        // Update URL and reload
        const newUrl = window.location.pathname + (params.toString() ? '?' + params.toString() : '');
        window.location.href = newUrl;
    },
    
    // Clear all filters
    clearFilters: function() {
        document.getElementById('searchInput').value = '';
        document.getElementById('statusFilter').value = '';
        document.getElementById('dateFilter').value = '';
        document.getElementById('sortFilter').value = 'created_at_desc';
        
        window.location.href = window.location.pathname;
    },
    
    // View order details
    viewOrderDetails: function(orderId) {
        const modal = document.getElementById('orderDetailsModal');
        const content = document.getElementById('orderDetailsContent');
        
        if (!modal || !content) return;
        
        // Show loading state
        content.innerHTML = `
            <div class="flex items-center justify-center py-12">
                <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
                <span class="ml-3 text-gray-600">Sipariş detayları yükleniyor...</span>
            </div>
        `;
        
        modal.classList.remove('hidden');
        
        // Load order details
        fetch(`/api/admin/orders/${orderId}`)
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    this.renderOrderDetails(data.order);
                } else {
                    content.innerHTML = `
                        <div class="text-center py-12">
                            <i class="fas fa-exclamation-triangle text-4xl text-red-500 mb-4"></i>
                            <p class="text-red-600">${data.message || 'Sipariş detayları yüklenemedi.'}</p>
                        </div>
                    `;
                }
            })
            .catch(error => {
                console.error('Error loading order details:', error);
                content.innerHTML = `
                    <div class="text-center py-12">
                        <i class="fas fa-exclamation-triangle text-4xl text-red-500 mb-4"></i>
                        <p class="text-red-600">Bir hata oluştu. Lütfen tekrar deneyin.</p>
                    </div>
                `;
            });
    },
    
    // Render order details in modal
    renderOrderDetails: function(order) {
        const content = document.getElementById('orderDetailsContent');
        if (!content) return;
        
        const statusBadge = this.getStatusBadgeHTML(order.status);
        
        content.innerHTML = `
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
                <!-- Order Information -->
                <div class="space-y-6">
                    <div>
                        <h4 class="text-lg font-semibold text-gray-900 mb-4">Sipariş Bilgileri</h4>
                        <div class="bg-gray-50 rounded-lg p-4 space-y-3">
                            <div class="flex justify-between">
                                <span class="text-sm font-medium text-gray-500">Sipariş No:</span>
                                <span class="text-sm text-gray-900">#${order.order_number}</span>
                            </div>
                            <div class="flex justify-between">
                                <span class="text-sm font-medium text-gray-500">Durum:</span>
                                <div>${statusBadge}</div>
                            </div>
                            <div class="flex justify-between">
                                <span class="text-sm font-medium text-gray-500">Tarih:</span>
                                <span class="text-sm text-gray-900">${new Date(order.created_at).toLocaleDateString('tr-TR')}</span>
                            </div>
                            <div class="flex justify-between">
                                <span class="text-sm font-medium text-gray-500">Toplam:</span>
                                <span class="text-sm font-bold text-gray-900">₺${parseFloat(order.total).toFixed(2)}</span>
                            </div>
                            ${order.tracking_number ? `
                            <div class="flex justify-between">
                                <span class="text-sm font-medium text-gray-500">Kargo Takip:</span>
                                <span class="text-sm text-blue-600">${order.tracking_number}</span>
                            </div>
                            ` : ''}
                        </div>
                    </div>
                    
                    <!-- Customer Information -->
                    <div>
                        <h4 class="text-lg font-semibold text-gray-900 mb-4">Müşteri Bilgileri</h4>
                        <div class="bg-gray-50 rounded-lg p-4">
                            <div class="flex items-center mb-4">
                                ${order.customer.avatar ? 
                                    `<img src="${order.customer.avatar}" alt="${order.customer.name}" class="w-12 h-12 rounded-full mr-4">` :
                                    `<div class="w-12 h-12 rounded-full bg-gray-300 flex items-center justify-center mr-4">
                                        <i class="fas fa-user text-gray-600"></i>
                                    </div>`
                                }
                                <div>
                                    <div class="font-medium text-gray-900">${order.customer.name}</div>
                                    <div class="text-sm text-gray-500">${order.customer.email}</div>
                                </div>
                            </div>
                            ${order.customer.phone ? `
                            <div class="text-sm text-gray-600">
                                <i class="fas fa-phone mr-2"></i>${order.customer.phone}
                            </div>
                            ` : ''}
                        </div>
                    </div>
                    
                    <!-- Shipping Address -->
                    ${order.shipping_address ? `
                    <div>
                        <h4 class="text-lg font-semibold text-gray-900 mb-4">Teslimat Adresi</h4>
                        <div class="bg-gray-50 rounded-lg p-4">
                            <div class="text-sm text-gray-900">
                                ${order.shipping_address.full_name}<br>
                                ${order.shipping_address.address_line_1}<br>
                                ${order.shipping_address.address_line_2 ? order.shipping_address.address_line_2 + '<br>' : ''}
                                ${order.shipping_address.city}, ${order.shipping_address.state} ${order.shipping_address.postal_code}<br>
                                ${order.shipping_address.country}
                            </div>
                        </div>
                    </div>
                    ` : ''}
                </div>
                
                <!-- Order Items -->
                <div>
                    <h4 class="text-lg font-semibold text-gray-900 mb-4">Sipariş Ürünleri</h4>
                    <div class="space-y-4">
                        ${order.items.map(item => `
                            <div class="flex items-center p-4 bg-gray-50 rounded-lg">
                                <img src="${item.product.image || '/static/images/no-image.jpg'}" 
                                     alt="${item.product.name}" class="w-16 h-16 object-cover rounded mr-4">
                                <div class="flex-1">
                                    <div class="font-medium text-gray-900">${item.product.name}</div>
                                    <div class="text-sm text-gray-500">SKU: ${item.product.sku || 'N/A'}</div>
                                    <div class="text-sm text-gray-500">Adet: ${item.quantity}</div>
                                </div>
                                <div class="text-right">
                                    <div class="font-medium text-gray-900">₺${parseFloat(item.price).toFixed(2)}</div>
                                    <div class="text-sm text-gray-500">Toplam: ₺${(parseFloat(item.price) * item.quantity).toFixed(2)}</div>
                                </div>
                            </div>
                        `).join('')}
                    </div>
                    
                    <!-- Order Summary -->
                    <div class="mt-6 bg-gray-50 rounded-lg p-4">
                        <div class="space-y-2">
                            <div class="flex justify-between text-sm">
                                <span>Ara Toplam:</span>
                                <span>₺${parseFloat(order.subtotal || order.total).toFixed(2)}</span>
                            </div>
                            ${order.tax_amount ? `
                            <div class="flex justify-between text-sm">
                                <span>KDV:</span>
                                <span>₺${parseFloat(order.tax_amount).toFixed(2)}</span>
                            </div>
                            ` : ''}
                            ${order.shipping_cost ? `
                            <div class="flex justify-between text-sm">
                                <span>Kargo:</span>
                                <span>₺${parseFloat(order.shipping_cost).toFixed(2)}</span>
                            </div>
                            ` : ''}
                            ${order.discount_amount ? `
                            <div class="flex justify-between text-sm text-green-600">
                                <span>İndirim:</span>
                                <span>-₺${parseFloat(order.discount_amount).toFixed(2)}</span>
                            </div>
                            ` : ''}
                            <div class="flex justify-between font-bold text-lg border-t pt-2">
                                <span>Toplam:</span>
                                <span>₺${parseFloat(order.total).toFixed(2)}</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            
            <!-- Action Buttons -->
            <div class="mt-8 flex justify-end space-x-4 border-t pt-6">
                <button onclick="AdminOrdersManager.editOrderStatus(${order.id}, '${order.status}')" 
                        class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-medium transition-colors">
                    <i class="fas fa-edit mr-2"></i>Durumu Değiştir
                </button>
                
                <button onclick="AdminOrdersManager.printInvoice(${order.id})" 
                        class="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded-lg font-medium transition-colors">
                    <i class="fas fa-print mr-2"></i>Fatura Yazdır
                </button>
                
                <button onclick="AdminOrdersManager.closeOrderDetailsModal()" 
                        class="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-lg font-medium transition-colors">
                    Kapat
                </button>
            </div>
        `;
    },
    
    // Get status badge HTML
    getStatusBadgeHTML: function(status) {
        const statusConfig = {
            'pending': { class: 'bg-yellow-100 text-yellow-800', icon: 'fas fa-clock', text: 'Bekliyor' },
            'confirmed': { class: 'bg-blue-100 text-blue-800', icon: 'fas fa-check', text: 'Onaylandı' },
            'processing': { class: 'bg-indigo-100 text-indigo-800', icon: 'fas fa-cog', text: 'İşleniyor' },
            'shipped': { class: 'bg-purple-100 text-purple-800', icon: 'fas fa-truck', text: 'Kargoda' },
            'delivered': { class: 'bg-green-100 text-green-800', icon: 'fas fa-check-circle', text: 'Teslim Edildi' },
            'cancelled': { class: 'bg-red-100 text-red-800', icon: 'fas fa-times', text: 'İptal Edildi' },
            'refunded': { class: 'bg-gray-100 text-gray-800', icon: 'fas fa-undo', text: 'İade Edildi' }
        };
        
        const config = statusConfig[status] || statusConfig['pending'];
        return `<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${config.class}">
            <i class="${config.icon} mr-1"></i>${config.text}
        </span>`;
    },
    
    // Close order details modal
    closeOrderDetailsModal: function() {
        const modal = document.getElementById('orderDetailsModal');
        if (modal) {
            modal.classList.add('hidden');
        }
    },
    
    // Edit order status
    editOrderStatus: function(orderId, currentStatus) {
        const modal = document.getElementById('statusUpdateModal');
        const orderIdInput = document.getElementById('updateOrderId');
        const statusSelect = document.getElementById('newStatus');
        
        if (!modal || !orderIdInput || !statusSelect) return;
        
        orderIdInput.value = orderId;
        statusSelect.value = currentStatus;
        
        // Clear other fields
        document.getElementById('trackingNumber').value = '';
        document.getElementById('statusNote').value = '';
        document.getElementById('notifyCustomer').checked = true;
        
        modal.classList.remove('hidden');
    },
    
    // Close status update modal
    closeStatusUpdateModal: function() {
        const modal = document.getElementById('statusUpdateModal');
        if (modal) {
            modal.classList.add('hidden');
        }
    },
    
    // Submit status update
    submitStatusUpdate: function(form) {
        const formData = new FormData(form);
        const submitBtn = form.querySelector('button[type="submit"]');
        
        // Show loading state
        const originalText = submitBtn.textContent;
        submitBtn.textContent = 'Güncelleniyor...';
        submitBtn.disabled = true;
        
        fetch('/api/admin/orders/update-status', {
            method: 'POST',
            headers: {
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: formData
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('Sipariş durumu başarıyla güncellendi!');
                this.closeStatusUpdateModal();
                window.location.reload();
            } else {
                alert(data.message || 'Durum güncellenemedi. Lütfen tekrar deneyin.');
            }
        })
        .catch(error => {
            console.error('Status update error:', error);
            alert('Bir hata oluştu. Lütfen tekrar deneyin.');
        })
        .finally(() => {
            submitBtn.textContent = originalText;
            submitBtn.disabled = false;
        });
    },
    
    // Print invoice
    printInvoice: function(orderId) {
        const printWindow = window.open(`/api/admin/orders/${orderId}/invoice?format=pdf`, '_blank');
        if (!printWindow) {
            alert('Pop-up engelleyici nedeniyle fatura açılamadı. Lütfen pop-up engelleyiciyi devre dışı bırakın.');
        }
    },
    
    // Toggle dropdown menu
    toggleDropdown: function(orderId) {
        const dropdown = document.getElementById(`dropdown-${orderId}`);
        if (!dropdown) return;
        
        // Close all other dropdowns
        document.querySelectorAll('[id^="dropdown-"]').forEach(d => {
            if (d !== dropdown) {
                d.classList.add('hidden');
            }
        });
        
        dropdown.classList.toggle('hidden');
    },
    
    // Send notification to customer
    sendNotification: function(orderId) {
        const message = prompt('Müşteriye gönderilecek mesaj:');
        if (!message) return;
        
        fetch('/api/admin/orders/send-notification', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: JSON.stringify({
                order_id: orderId,
                message: message
            })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('Bildirim başarıyla gönderildi!');
            } else {
                alert(data.message || 'Bildirim gönderilemedi.');
            }
        })
        .catch(error => {
            console.error('Notification error:', error);
            alert('Bir hata oluştu.');
        });
    },
    
    // Add note to order
    addNote: function(orderId) {
        const note = prompt('Sipariş notu:');
        if (!note) return;
        
        fetch('/api/admin/orders/add-note', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: JSON.stringify({
                order_id: orderId,
                note: note
            })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('Not başarıyla eklendi!');
            } else {
                alert(data.message || 'Not eklenemedi.');
            }
        })
        .catch(error => {
            console.error('Add note error:', error);
            alert('Bir hata oluştu.');
        });
    },
    
    // Refund order
    refundOrder: function(orderId) {
        if (!confirm('Bu siparişi iade etmek istediğinizden emin misiniz?')) return;
        
        const reason = prompt('İade nedeni (isteğe bağlı):');
        
        fetch('/api/admin/orders/refund', {
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
                alert('Sipariş başarıyla iade edildi!');
                window.location.reload();
            } else {
                alert(data.message || 'İade işlemi başarısız.');
            }
        })
        .catch(error => {
            console.error('Refund error:', error);
            alert('Bir hata oluştu.');
        });
    },
    
    // Cancel order
    cancelOrder: function(orderId) {
        if (!confirm('Bu siparişi iptal etmek istediğinizden emin misiniz?')) return;
        
        const reason = prompt('İptal nedeni (isteğe bağlı):');
        
        fetch('/api/admin/orders/cancel', {
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
                alert(data.message || 'İptal işlemi başarısız.');
            }
        })
        .catch(error => {
            console.error('Cancel error:', error);
            alert('Bir hata oluştu.');
        });
    },
    
    // Toggle select all
    toggleSelectAll: function(checked) {
        const checkboxes = document.querySelectorAll('.order-checkbox');
        checkboxes.forEach(checkbox => {
            checkbox.checked = checked;
            this.toggleOrderSelection(checkbox.value, checked);
        });
    },
    
    // Toggle individual order selection
    toggleOrderSelection: function(orderId, selected) {
        if (selected) {
            this.selectedOrders.add(orderId);
        } else {
            this.selectedOrders.delete(orderId);
        }
        
        this.updateBulkActionsVisibility();
        this.updateSelectAllState();
    },
    
    // Update bulk actions visibility
    updateBulkActionsVisibility: function() {
        const bulkActions = document.getElementById('bulkActions');
        const selectedCount = document.getElementById('selectedCount');
        
        if (!bulkActions || !selectedCount) return;
        
        if (this.selectedOrders.size > 0) {
            bulkActions.classList.remove('hidden');
            selectedCount.textContent = `${this.selectedOrders.size} sipariş seçildi`;
        } else {
            bulkActions.classList.add('hidden');
        }
    },
    
    // Update select all checkbox state
    updateSelectAllState: function() {
        const selectAllCheckbox = document.getElementById('selectAll');
        const checkboxes = document.querySelectorAll('.order-checkbox');
        
        if (!selectAllCheckbox || checkboxes.length === 0) return;
        
        const checkedCount = Array.from(checkboxes).filter(cb => cb.checked).length;
        
        if (checkedCount === 0) {
            selectAllCheckbox.checked = false;
            selectAllCheckbox.indeterminate = false;
        } else if (checkedCount === checkboxes.length) {
            selectAllCheckbox.checked = true;
            selectAllCheckbox.indeterminate = false;
        } else {
            selectAllCheckbox.checked = false;
            selectAllCheckbox.indeterminate = true;
        }
    },
    
    // Bulk update status
    bulkUpdateStatus: function() {
        if (this.selectedOrders.size === 0) return;
        
        const newStatus = prompt('Yeni durum (pending, confirmed, processing, shipped, delivered, cancelled, refunded):');
        if (!newStatus) return;
        
        const validStatuses = ['pending', 'confirmed', 'processing', 'shipped', 'delivered', 'cancelled', 'refunded'];
        if (!validStatuses.includes(newStatus)) {
            alert('Geçersiz durum. Geçerli durumlar: ' + validStatuses.join(', '));
            return;
        }
        
        if (!confirm(`${this.selectedOrders.size} siparişin durumunu '${newStatus}' olarak değiştirmek istediğinizden emin misiniz?`)) return;
        
        fetch('/api/admin/orders/bulk-update-status', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: JSON.stringify({
                order_ids: Array.from(this.selectedOrders),
                status: newStatus
            })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert(`${data.updated_count} sipariş başarıyla güncellendi!`);
                window.location.reload();
            } else {
                alert(data.message || 'Toplu güncelleme başarısız.');
            }
        })
        .catch(error => {
            console.error('Bulk update error:', error);
            alert('Bir hata oluştu.');
        });
    },
    
    // Bulk export
    bulkExport: function() {
        if (this.selectedOrders.size === 0) return;
        
        const format = prompt('Dışa aktarma formatı (csv, excel, pdf):') || 'csv';
        
        const form = document.createElement('form');
        form.method = 'POST';
        form.action = '/api/admin/orders/bulk-export';
        form.target = '_blank';
        
        // Add CSRF token
        const csrfInput = document.createElement('input');
        csrfInput.type = 'hidden';
        csrfInput.name = '_token';
        csrfInput.value = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || '';
        form.appendChild(csrfInput);
        
        // Add order IDs
        this.selectedOrders.forEach(orderId => {
            const input = document.createElement('input');
            input.type = 'hidden';
            input.name = 'order_ids[]';
            input.value = orderId;
            form.appendChild(input);
        });
        
        // Add format
        const formatInput = document.createElement('input');
        formatInput.type = 'hidden';
        formatInput.name = 'format';
        formatInput.value = format;
        form.appendChild(formatInput);
        
        document.body.appendChild(form);
        form.submit();
        document.body.removeChild(form);
    },
    
    // Clear selection
    clearSelection: function() {
        this.selectedOrders.clear();
        document.querySelectorAll('.order-checkbox').forEach(cb => cb.checked = false);
        document.getElementById('selectAll').checked = false;
        this.updateBulkActionsVisibility();
    },
    
    // Export all orders
    exportOrders: function() {
        const format = prompt('Dışa aktarma formatı (csv, excel, pdf):') || 'csv';
        window.open(`/api/admin/orders/export?format=${format}`, '_blank');
    },
    
    // Refresh orders
    refreshOrders: function() {
        window.location.reload();
    },
    
    // Refresh order statuses (for auto-refresh)
    refreshOrderStatuses: function() {
        const orderRows = document.querySelectorAll('[data-order-id]');
        if (orderRows.length === 0) return;
        
        const orderIds = Array.from(orderRows).map(row => row.getAttribute('data-order-id'));
        
        fetch('/api/admin/orders/status-check', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: JSON.stringify({ order_ids: orderIds })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success && data.orders) {
                data.orders.forEach(order => {
                    this.updateOrderRowStatus(order.id, order.status);
                });
            }
        })
        .catch(error => {
            console.error('Status refresh error:', error);
        });
    },
    
    // Update order row status
    updateOrderRowStatus: function(orderId, newStatus) {
        const row = document.querySelector(`[data-order-id="${orderId}"]`);
        if (!row) return;
        
        const statusCell = row.querySelector('td:nth-child(6)'); // Status column
        if (statusCell) {
            statusCell.innerHTML = this.getStatusBadgeHTML(newStatus);
        }
    }
};

// Global functions for backward compatibility
function applyFilters() {
    window.AdminOrdersManager.applyFilters();
}

function clearFilters() {
    window.AdminOrdersManager.clearFilters();
}

function viewOrderDetails(orderId) {
    window.AdminOrdersManager.viewOrderDetails(orderId);
}

function editOrderStatus(orderId, currentStatus) {
    window.AdminOrdersManager.editOrderStatus(orderId, currentStatus);
}

function printInvoice(orderId) {
    window.AdminOrdersManager.printInvoice(orderId);
}

function toggleDropdown(orderId) {
    window.AdminOrdersManager.toggleDropdown(orderId);
}

function sendNotification(orderId) {
    window.AdminOrdersManager.sendNotification(orderId);
}

function addNote(orderId) {
    window.AdminOrdersManager.addNote(orderId);
}

function refundOrder(orderId) {
    window.AdminOrdersManager.refundOrder(orderId);
}

function cancelOrder(orderId) {
    window.AdminOrdersManager.cancelOrder(orderId);
}

function bulkUpdateStatus() {
    window.AdminOrdersManager.bulkUpdateStatus();
}

function bulkExport() {
    window.AdminOrdersManager.bulkExport();
}

function clearSelection() {
    window.AdminOrdersManager.clearSelection();
}

function exportOrders() {
    window.AdminOrdersManager.exportOrders();
}

function refreshOrders() {
    window.AdminOrdersManager.refreshOrders();
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    window.AdminOrdersManager.init();
});

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = window.AdminOrdersManager;
}