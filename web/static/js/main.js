// Main JavaScript for KolajAI Marketplace

// DOM Ready
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
});

// Initialize application
function initializeApp() {
    initSearch();
    initCart();
    initAuctions();
    initModals();
    initTooltips();
    initLazyLoading();
}

// Search functionality
function initSearch() {
    const searchForm = document.querySelector('form[action="/products"]');
    const searchInput = searchForm?.querySelector('input[name="search"]');
    
    if (searchInput) {
        let searchTimeout;
        
        searchInput.addEventListener('input', function() {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                if (this.value.length >= 3) {
                    performSearch(this.value);
                }
            }, 300);
        });
        
        // Hide search results when clicking outside
        document.addEventListener('click', function(e) {
            if (!searchForm.contains(e.target)) {
                hideSearchResults();
            }
        });
    }
}

// Perform search
async function performSearch(query) {
    try {
        const response = await fetch(`/api/search?q=${encodeURIComponent(query)}`);
        const products = await response.json();
        showSearchResults(products);
    } catch (error) {
        console.error('Search error:', error);
    }
}

// Show search results
function showSearchResults(products) {
    const searchForm = document.querySelector('form[action="/products"]');
    let resultsContainer = document.getElementById('search-results');
    
    if (!resultsContainer) {
        resultsContainer = document.createElement('div');
        resultsContainer.id = 'search-results';
        resultsContainer.className = 'search-results';
        searchForm.appendChild(resultsContainer);
    }
    
    if (products.length === 0) {
        resultsContainer.innerHTML = '<div class="search-result-item">Ürün bulunamadı</div>';
    } else {
        resultsContainer.innerHTML = products.slice(0, 5).map(product => `
            <div class="search-result-item" onclick="window.location.href='/product/${product.id}'">
                <div class="flex items-center space-x-3">
                    <img src="${product.image || '/static/images/placeholder.jpg'}" 
                         alt="${product.name}" class="w-10 h-10 object-cover rounded">
                    <div>
                        <div class="font-medium">${product.name}</div>
                        <div class="text-sm text-gray-600">${formatPrice(product.price)}</div>
                    </div>
                </div>
            </div>
        `).join('');
    }
    
    resultsContainer.style.display = 'block';
}

// Hide search results
function hideSearchResults() {
    const resultsContainer = document.getElementById('search-results');
    if (resultsContainer) {
        resultsContainer.style.display = 'none';
    }
}

// Cart functionality
function initCart() {
    // Add to cart buttons
    document.querySelectorAll('.add-to-cart').forEach(button => {
        button.addEventListener('click', function(e) {
            e.preventDefault();
            const productId = this.dataset.productId;
            const quantity = this.dataset.quantity || 1;
            addToCart(productId, quantity);
        });
    });
    
    // Cart quantity updates
    document.querySelectorAll('.cart-quantity').forEach(input => {
        input.addEventListener('change', function() {
            const itemId = this.dataset.itemId;
            const quantity = parseInt(this.value);
            updateCartItem(itemId, quantity);
        });
    });
    
    // Remove from cart
    document.querySelectorAll('.remove-from-cart').forEach(button => {
        button.addEventListener('click', function(e) {
            e.preventDefault();
            const itemId = this.dataset.itemId;
            removeFromCart(itemId);
        });
    });
}

// Add product to cart
async function addToCart(productId, quantity = 1) {
    try {
        const formData = new FormData();
        formData.append('product_id', productId);
        formData.append('quantity', quantity);
        
        const response = await fetch('/add-to-cart', {
            method: 'POST',
            body: formData
        });
        
        if (response.ok) {
            showNotification('Ürün sepete eklendi', 'success');
            updateCartCount();
        } else {
            showNotification('Ürün sepete eklenemedi', 'error');
        }
    } catch (error) {
        console.error('Add to cart error:', error);
        showNotification('Bir hata oluştu', 'error');
    }
}

// Update cart item
async function updateCartItem(itemId, quantity) {
    try {
        const response = await fetch('/api/cart/update', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                item_id: parseInt(itemId),
                quantity: quantity
            })
        });
        
        if (response.ok) {
            location.reload(); // Reload to update totals
        } else {
            showNotification('Sepet güncellenemedi', 'error');
        }
    } catch (error) {
        console.error('Update cart error:', error);
        showNotification('Bir hata oluştu', 'error');
    }
}

// Remove from cart
async function removeFromCart(itemId) {
    if (!confirm('Bu ürünü sepetten çıkarmak istediğinizden emin misiniz?')) {
        return;
    }
    
    try {
        const response = await fetch('/api/cart/update', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                item_id: parseInt(itemId),
                quantity: 0
            })
        });
        
        if (response.ok) {
            location.reload();
        } else {
            showNotification('Ürün çıkarılamadı', 'error');
        }
    } catch (error) {
        console.error('Remove from cart error:', error);
        showNotification('Bir hata oluştu', 'error');
    }
}

// Update cart count in header
async function updateCartCount() {
    try {
        const response = await fetch('/api/cart/count');
        const data = await response.json();
        const cartBadge = document.querySelector('.cart-count');
        if (cartBadge) {
            cartBadge.textContent = data.count;
        }
    } catch (error) {
        console.error('Update cart count error:', error);
    }
}

// Auction functionality
function initAuctions() {
    // Place bid forms
    document.querySelectorAll('.bid-form').forEach(form => {
        form.addEventListener('submit', function(e) {
            e.preventDefault();
            const auctionId = this.dataset.auctionId;
            const amount = this.querySelector('input[name="amount"]').value;
            placeBid(auctionId, amount);
        });
    });
    
    // Watch auction buttons
    document.querySelectorAll('.watch-auction').forEach(button => {
        button.addEventListener('click', function(e) {
            e.preventDefault();
            const auctionId = this.dataset.auctionId;
            toggleWatchAuction(auctionId);
        });
    });
    
    // Initialize countdown timers
    initCountdownTimers();
}

// Place bid on auction
async function placeBid(auctionId, amount) {
    try {
        const formData = new FormData();
        formData.append('auction_id', auctionId);
        formData.append('amount', amount);
        
        const response = await fetch('/place-bid', {
            method: 'POST',
            body: formData
        });
        
        if (response.ok) {
            location.reload(); // Reload to show updated bid
        } else {
            const text = await response.text();
            showNotification(text || 'Teklif verilemedi', 'error');
        }
    } catch (error) {
        console.error('Place bid error:', error);
        showNotification('Bir hata oluştu', 'error');
    }
}

// Toggle watch auction
async function toggleWatchAuction(auctionId) {
    try {
        const response = await fetch(`/api/auction/${auctionId}/watch`, {
            method: 'POST'
        });
        
        if (response.ok) {
            const button = document.querySelector(`[data-auction-id="${auctionId}"]`);
            button.classList.toggle('watching');
            const isWatching = button.classList.contains('watching');
            button.textContent = isWatching ? 'Takipten Çıkar' : 'Takip Et';
        }
    } catch (error) {
        console.error('Toggle watch error:', error);
    }
}

// Initialize countdown timers
function initCountdownTimers() {
    document.querySelectorAll('.countdown-timer').forEach(timer => {
        const endTime = new Date(timer.dataset.endTime).getTime();
        updateCountdown(timer, endTime);
        
        setInterval(() => {
            updateCountdown(timer, endTime);
        }, 1000);
    });
}

// Update countdown display
function updateCountdown(element, endTime) {
    const now = new Date().getTime();
    const distance = endTime - now;
    
    if (distance < 0) {
        element.innerHTML = 'Süresi Doldu';
        element.classList.add('expired');
        return;
    }
    
    const days = Math.floor(distance / (1000 * 60 * 60 * 24));
    const hours = Math.floor((distance % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    const minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60));
    const seconds = Math.floor((distance % (1000 * 60)) / 1000);
    
    let display = '';
    if (days > 0) display += `${days}g `;
    display += `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    
    element.innerHTML = display;
}

// Modal functionality
function initModals() {
    // Open modal buttons
    document.querySelectorAll('[data-modal-target]').forEach(button => {
        button.addEventListener('click', function(e) {
            e.preventDefault();
            const modalId = this.dataset.modalTarget;
            openModal(modalId);
        });
    });
    
    // Close modal buttons
    document.querySelectorAll('[data-modal-close]').forEach(button => {
        button.addEventListener('click', function(e) {
            e.preventDefault();
            const modalId = this.dataset.modalClose;
            closeModal(modalId);
        });
    });
    
    // Close modal on backdrop click
    document.querySelectorAll('.modal-backdrop').forEach(backdrop => {
        backdrop.addEventListener('click', function(e) {
            if (e.target === this) {
                const modal = this.closest('.modal');
                closeModal(modal.id);
            }
        });
    });
}

// Open modal
function openModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.remove('hidden');
        document.body.style.overflow = 'hidden';
    }
}

// Close modal
function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.add('hidden');
        document.body.style.overflow = 'auto';
    }
}

// Tooltip functionality
function initTooltips() {
    document.querySelectorAll('[data-tooltip]').forEach(element => {
        element.addEventListener('mouseenter', showTooltip);
        element.addEventListener('mouseleave', hideTooltip);
    });
}

// Show tooltip
function showTooltip(e) {
    const text = e.target.dataset.tooltip;
    const tooltip = document.createElement('div');
    tooltip.className = 'tooltip absolute bg-gray-800 text-white px-2 py-1 rounded text-sm z-50';
    tooltip.textContent = text;
    tooltip.id = 'tooltip';
    
    document.body.appendChild(tooltip);
    
    const rect = e.target.getBoundingClientRect();
    tooltip.style.left = rect.left + (rect.width / 2) - (tooltip.offsetWidth / 2) + 'px';
    tooltip.style.top = rect.top - tooltip.offsetHeight - 5 + 'px';
}

// Hide tooltip
function hideTooltip() {
    const tooltip = document.getElementById('tooltip');
    if (tooltip) {
        tooltip.remove();
    }
}

// Lazy loading for images
function initLazyLoading() {
    const images = document.querySelectorAll('img[data-src]');
    
    if ('IntersectionObserver' in window) {
        const imageObserver = new IntersectionObserver((entries, observer) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    const img = entry.target;
                    img.src = img.dataset.src;
                    img.classList.remove('lazy');
                    imageObserver.unobserve(img);
                }
            });
        });
        
        images.forEach(img => imageObserver.observe(img));
    } else {
        // Fallback for older browsers
        images.forEach(img => {
            img.src = img.dataset.src;
        });
    }
}

// Notification system
function showNotification(message, type = 'info', duration = 5000) {
    const notification = document.createElement('div');
    notification.className = `notification fixed top-4 right-4 px-6 py-3 rounded-lg shadow-lg z-50 transition-all duration-300 transform translate-x-full`;
    
    // Set notification style based on type
    switch (type) {
        case 'success':
            notification.classList.add('bg-green-500', 'text-white');
            break;
        case 'error':
            notification.classList.add('bg-red-500', 'text-white');
            break;
        case 'warning':
            notification.classList.add('bg-yellow-500', 'text-white');
            break;
        default:
            notification.classList.add('bg-blue-500', 'text-white');
    }
    
    notification.innerHTML = `
        <div class="flex items-center space-x-2">
            <span>${message}</span>
            <button onclick="this.parentElement.parentElement.remove()" class="ml-4 text-white hover:text-gray-200">
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path>
                </svg>
            </button>
        </div>
    `;
    
    document.body.appendChild(notification);
    
    // Animate in
    setTimeout(() => {
        notification.classList.remove('translate-x-full');
    }, 100);
    
    // Auto remove
    setTimeout(() => {
        notification.classList.add('translate-x-full');
        setTimeout(() => {
            notification.remove();
        }, 300);
    }, duration);
}

// Utility functions
function formatPrice(price) {
    return new Intl.NumberFormat('tr-TR', {
        style: 'currency',
        currency: 'TRY'
    }).format(price);
}

function formatDate(date) {
    return new Intl.DateTimeFormat('tr-TR', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    }).format(new Date(date));
}

function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// Export functions for global use
window.KolajAI = {
    addToCart,
    updateCartItem,
    removeFromCart,
    placeBid,
    showNotification,
    formatPrice,
    formatDate
};