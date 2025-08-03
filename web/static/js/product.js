/**
 * Product related JavaScript functions
 */

// Global product management object
window.ProductManager = {
    
    // Change main product image
    changeMainImage: function(imageUrl) {
        const mainImage = document.getElementById('mainImage');
        if (mainImage && imageUrl) {
            mainImage.src = imageUrl;
            
            // Update active thumbnail
            const thumbnails = document.querySelectorAll('.thumbnail');
            thumbnails.forEach(thumb => {
                thumb.classList.remove('border-blue-500');
                thumb.classList.add('border-transparent');
                
                if (thumb.src === imageUrl) {
                    thumb.classList.remove('border-transparent');
                    thumb.classList.add('border-blue-500');
                }
            });
        }
    },
    
    // Add product to cart
    addToCart: function(productId, quantity = 1) {
        if (!productId) {
            console.error('Product ID is required');
            return;
        }
        
        // Show loading state
        const addButton = document.querySelector('.add-to-cart-btn');
        if (addButton) {
            const originalText = addButton.textContent;
            addButton.textContent = 'Ekleniyor...';
            addButton.disabled = true;
            
            // Use main app's cart service if available
            if (window.app && window.app.addToCart) {
                window.app.addToCart(productId, quantity)
                    .then(() => {
                        addButton.textContent = 'Sepete Eklendi!';
                        addButton.classList.add('bg-green-600');
                        
                        setTimeout(() => {
                            addButton.textContent = originalText;
                            addButton.classList.remove('bg-green-600');
                            addButton.disabled = false;
                        }, 2000);
                    })
                    .catch(error => {
                        console.error('Error adding to cart:', error);
                        addButton.textContent = 'Hata Oluştu';
                        addButton.classList.add('bg-red-600');
                        
                        setTimeout(() => {
                            addButton.textContent = originalText;
                            addButton.classList.remove('bg-red-600');
                            addButton.disabled = false;
                        }, 2000);
                    });
            } else {
                // Fallback to direct API call
                fetch('/api/cart/add', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
                    },
                    body: JSON.stringify({
                        product_id: productId,
                        quantity: quantity
                    })
                })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        addButton.textContent = 'Sepete Eklendi!';
                        addButton.classList.add('bg-green-600');
                    } else {
                        throw new Error(data.message || 'Sepete eklenirken hata oluştu');
                    }
                })
                .catch(error => {
                    console.error('Error adding to cart:', error);
                    addButton.textContent = 'Hata Oluştu';
                    addButton.classList.add('bg-red-600');
                })
                .finally(() => {
                    setTimeout(() => {
                        addButton.textContent = originalText;
                        addButton.classList.remove('bg-green-600', 'bg-red-600');
                        addButton.disabled = false;
                    }, 2000);
                });
            }
        }
    },
    
    // Update product quantity
    updateQuantity: function(change) {
        const quantityInput = document.getElementById('quantity');
        if (quantityInput) {
            let currentValue = parseInt(quantityInput.value) || 1;
            let newValue = currentValue + change;
            
            // Ensure minimum quantity is 1
            if (newValue < 1) {
                newValue = 1;
            }
            
            // Check maximum stock if available
            const maxStock = parseInt(quantityInput.getAttribute('max'));
            if (maxStock && newValue > maxStock) {
                newValue = maxStock;
            }
            
            quantityInput.value = newValue;
        }
    },
    
    // Apply product filters
    applyFilters: function() {
        const form = document.getElementById('productFilters');
        if (!form) return;
        
        const formData = new FormData(form);
        const params = new URLSearchParams();
        
        // Collect filter parameters
        for (let [key, value] of formData.entries()) {
            if (value) {
                params.append(key, value);
            }
        }
        
        // Collect checked categories
        const categoryCheckboxes = form.querySelectorAll('input[name="categories[]"]:checked');
        categoryCheckboxes.forEach(checkbox => {
            params.append('categories[]', checkbox.value);
        });
        
        // Collect price range
        const priceMin = document.getElementById('priceMin')?.value;
        const priceMax = document.getElementById('priceMax')?.value;
        
        if (priceMin) params.append('price_min', priceMin);
        if (priceMax) params.append('price_max', priceMax);
        
        // Update URL and reload
        const newUrl = window.location.pathname + '?' + params.toString();
        window.location.href = newUrl;
    },
    
    // Clear all filters
    clearFilters: function() {
        window.location.href = window.location.pathname;
    },
    
    // Toggle product favorite status
    toggleFavorite: function(productId) {
        if (!productId) return;
        
        const favoriteBtn = document.querySelector(`[data-product-id="${productId}"] .favorite-btn`);
        if (!favoriteBtn) return;
        
        const isFavorite = favoriteBtn.classList.contains('text-red-500');
        
        fetch('/api/favorites/toggle', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: JSON.stringify({ product_id: productId })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                if (data.is_favorite) {
                    favoriteBtn.classList.add('text-red-500');
                    favoriteBtn.classList.remove('text-gray-400');
                } else {
                    favoriteBtn.classList.remove('text-red-500');
                    favoriteBtn.classList.add('text-gray-400');
                }
            }
        })
        .catch(error => {
            console.error('Error toggling favorite:', error);
        });
    }
};

// Global functions for backward compatibility
function changeMainImage(imageUrl) {
    window.ProductManager.changeMainImage(imageUrl);
}

function addToCart(productId, quantity) {
    window.ProductManager.addToCart(productId, quantity);
}

function updateQuantity(change) {
    window.ProductManager.updateQuantity(change);
}

// Initialize product page functionality
document.addEventListener('DOMContentLoaded', function() {
    
    // Initialize quantity controls
    const quantityControls = document.querySelectorAll('.quantity-control');
    quantityControls.forEach(control => {
        const decreaseBtn = control.querySelector('.decrease-btn');
        const increaseBtn = control.querySelector('.increase-btn');
        
        if (decreaseBtn) {
            decreaseBtn.addEventListener('click', () => window.ProductManager.updateQuantity(-1));
        }
        
        if (increaseBtn) {
            increaseBtn.addEventListener('click', () => window.ProductManager.updateQuantity(1));
        }
    });
    
    // Initialize filter form
    const filterForm = document.getElementById('productFilters');
    if (filterForm) {
        // Auto-submit on filter changes
        const filterInputs = filterForm.querySelectorAll('input[type="checkbox"], input[type="radio"], select');
        filterInputs.forEach(input => {
            input.addEventListener('change', function() {
                // Debounce the filter application
                clearTimeout(window.filterTimeout);
                window.filterTimeout = setTimeout(() => {
                    window.ProductManager.applyFilters();
                }, 500);
            });
        });
        
        // Manual filter application
        const applyBtn = document.getElementById('applyFilters');
        if (applyBtn) {
            applyBtn.addEventListener('click', window.ProductManager.applyFilters);
        }
        
        const clearBtn = document.getElementById('clearFilters');
        if (clearBtn) {
            clearBtn.addEventListener('click', window.ProductManager.clearFilters);
        }
    }
    
    // Initialize favorite buttons
    const favoriteButtons = document.querySelectorAll('.favorite-btn');
    favoriteButtons.forEach(btn => {
        btn.addEventListener('click', function(e) {
            e.preventDefault();
            const productId = this.closest('[data-product-id]')?.getAttribute('data-product-id');
            if (productId) {
                window.ProductManager.toggleFavorite(productId);
            }
        });
    });
    
    // Initialize add to cart buttons
    const addToCartButtons = document.querySelectorAll('.add-to-cart-btn');
    addToCartButtons.forEach(btn => {
        btn.addEventListener('click', function(e) {
            e.preventDefault();
            const productId = this.getAttribute('data-product-id') || 
                             this.closest('[data-product-id]')?.getAttribute('data-product-id');
            const quantity = document.getElementById('quantity')?.value || 1;
            
            if (productId) {
                window.ProductManager.addToCart(productId, parseInt(quantity));
            }
        });
    });
});

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = window.ProductManager;
}