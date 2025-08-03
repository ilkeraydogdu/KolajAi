/**
 * Auction functionality JavaScript
 */

// Global auction manager
window.AuctionManager = {
    
    // Active countdowns
    countdowns: new Map(),
    
    // WebSocket connection for real-time updates
    socket: null,
    
    // Initialize auction functionality
    init: function() {
        this.initCountdowns();
        this.initEventListeners();
        this.initWebSocket();
        this.initFilters();
    },
    
    // Initialize countdown timers
    initCountdowns: function() {
        const countdownElements = document.querySelectorAll('.countdown');
        
        countdownElements.forEach(element => {
            const endTime = element.getAttribute('data-end-time');
            if (endTime) {
                this.startCountdown(element, new Date(endTime));
            }
        });
    },
    
    // Start individual countdown
    startCountdown: function(element, endTime) {
        const countdownId = 'countdown_' + Math.random().toString(36).substr(2, 9);
        
        const updateCountdown = () => {
            const now = new Date().getTime();
            const distance = endTime.getTime() - now;
            
            if (distance < 0) {
                // Auction ended
                element.innerHTML = '<span class="text-red-600 font-bold">Müzayede Bitti</span>';
                this.onAuctionEnd(element);
                clearInterval(this.countdowns.get(countdownId));
                this.countdowns.delete(countdownId);
                return;
            }
            
            // Calculate time units
            const days = Math.floor(distance / (1000 * 60 * 60 * 24));
            const hours = Math.floor((distance % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
            const minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60));
            const seconds = Math.floor((distance % (1000 * 60)) / 1000);
            
            // Update display
            const daysEl = element.querySelector('.days');
            const hoursEl = element.querySelector('.hours');
            const minutesEl = element.querySelector('.minutes');
            const secondsEl = element.querySelector('.seconds');
            
            if (daysEl) daysEl.textContent = days;
            if (hoursEl) hoursEl.textContent = hours.toString().padStart(2, '0');
            if (minutesEl) minutesEl.textContent = minutes.toString().padStart(2, '0');
            if (secondsEl) secondsEl.textContent = seconds.toString().padStart(2, '0');
            
            // Change color when time is running out
            if (distance < 3600000) { // Less than 1 hour
                element.classList.add('text-red-600');
                element.classList.remove('text-orange-600', 'text-green-600');
            } else if (distance < 86400000) { // Less than 1 day
                element.classList.add('text-orange-600');
                element.classList.remove('text-red-600', 'text-green-600');
            }
        };
        
        // Initial update
        updateCountdown();
        
        // Set interval and store reference
        const intervalId = setInterval(updateCountdown, 1000);
        this.countdowns.set(countdownId, intervalId);
    },
    
    // Handle auction end
    onAuctionEnd: function(element) {
        const auctionCard = element.closest('[data-auction-id]');
        if (auctionCard) {
            const auctionId = auctionCard.getAttribute('data-auction-id');
            
            // Update UI
            const statusBadge = auctionCard.querySelector('.bg-green-500, .bg-orange-500');
            if (statusBadge) {
                statusBadge.className = 'bg-gray-500 text-white px-2 py-1 rounded-full text-xs font-medium';
                statusBadge.innerHTML = '<i class="fas fa-stop mr-1"></i>Bitti';
            }
            
            // Update action buttons
            const bidBtn = auctionCard.querySelector('.bid-btn');
            if (bidBtn) {
                bidBtn.className = 'flex-1 bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-lg font-medium text-center transition-colors';
                bidBtn.innerHTML = '<i class="fas fa-eye mr-2"></i>Detayları Gör';
                bidBtn.onclick = () => window.location.href = `/auction/${auctionId}`;
            }
            
            // Refresh auction data
            this.refreshAuctionData(auctionId);
        }
    },
    
    // Initialize event listeners
    initEventListeners: function() {
        // Bid buttons
        document.addEventListener('click', (e) => {
            if (e.target.matches('.bid-btn') || e.target.closest('.bid-btn')) {
                e.preventDefault();
                const btn = e.target.matches('.bid-btn') ? e.target : e.target.closest('.bid-btn');
                const auctionId = btn.getAttribute('data-auction-id');
                if (auctionId) {
                    this.showBidModal(auctionId);
                }
            }
        });
        
        // Quick bid buttons
        document.addEventListener('click', (e) => {
            if (e.target.matches('.quick-bid')) {
                e.preventDefault();
                const amount = e.target.getAttribute('data-amount');
                const bidInput = document.getElementById('bidAmount');
                if (bidInput && amount) {
                    bidInput.value = amount;
                }
            }
        });
        
        // Watch/Favorite buttons
        document.addEventListener('click', (e) => {
            if (e.target.matches('.favorite-btn') || e.target.closest('.favorite-btn')) {
                e.preventDefault();
                const btn = e.target.matches('.favorite-btn') ? e.target : e.target.closest('.favorite-btn');
                const auctionCard = btn.closest('[data-auction-id]');
                if (auctionCard) {
                    const auctionId = auctionCard.getAttribute('data-auction-id');
                    this.toggleFavorite(auctionId, btn);
                }
            }
            
            if (e.target.matches('.watchlist-btn') || e.target.closest('.watchlist-btn')) {
                e.preventDefault();
                const btn = e.target.matches('.watchlist-btn') ? e.target : e.target.closest('.watchlist-btn');
                const auctionId = btn.getAttribute('data-auction-id');
                if (auctionId) {
                    this.toggleWatchlist(auctionId, btn);
                }
            }
        });
        
        // Modal close buttons
        document.addEventListener('click', (e) => {
            if (e.target.matches('#closeBidModal')) {
                this.closeBidModal();
            }
        });
        
        // Filter changes
        document.addEventListener('change', (e) => {
            if (e.target.matches('#categoryFilter, #statusFilter, #sortBy')) {
                this.applyFilters();
            }
        });
    },
    
    // Initialize WebSocket for real-time updates
    initWebSocket: function() {
        if (!window.WebSocket) return;
        
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws/auctions`;
        
        try {
            this.socket = new WebSocket(wsUrl);
            
            this.socket.onopen = () => {
                console.log('Auction WebSocket connected');
            };
            
            this.socket.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    this.handleWebSocketMessage(data);
                } catch (error) {
                    console.error('WebSocket message parse error:', error);
                }
            };
            
            this.socket.onclose = () => {
                console.log('Auction WebSocket disconnected');
                // Reconnect after 5 seconds
                setTimeout(() => this.initWebSocket(), 5000);
            };
            
            this.socket.onerror = (error) => {
                console.error('WebSocket error:', error);
            };
        } catch (error) {
            console.error('WebSocket initialization error:', error);
        }
    },
    
    // Handle WebSocket messages
    handleWebSocketMessage: function(data) {
        switch (data.type) {
            case 'bid_update':
                this.updateBidDisplay(data.auction_id, data.current_bid, data.bid_count);
                break;
            case 'auction_ended':
                this.handleAuctionEndedMessage(data.auction_id);
                break;
            case 'new_bid':
                this.showNewBidNotification(data);
                break;
            default:
                console.log('Unknown WebSocket message type:', data.type);
        }
    },
    
    // Update bid display in real-time
    updateBidDisplay: function(auctionId, currentBid, bidCount) {
        const auctionCard = document.querySelector(`[data-auction-id="${auctionId}"]`);
        if (!auctionCard) return;
        
        // Update current bid amount
        const bidAmountEl = auctionCard.querySelector('.text-green-600');
        if (bidAmountEl && bidAmountEl.textContent.includes('₺')) {
            bidAmountEl.textContent = `₺${parseFloat(currentBid).toFixed(2)}`;
        }
        
        // Update bid count
        const bidCountEl = auctionCard.querySelector('.text-gray-500');
        if (bidCountEl && bidCountEl.textContent.includes('teklif')) {
            bidCountEl.textContent = `${bidCount} teklif`;
        }
        
        // Update minimum bid in modal if open
        const bidInput = document.getElementById('bidAmount');
        if (bidInput) {
            const minBid = parseFloat(currentBid) + 1;
            bidInput.min = minBid;
            bidInput.placeholder = minBid.toString();
        }
    },
    
    // Show bid modal
    showBidModal: function(auctionId) {
        const modal = document.getElementById('bidModal');
        if (!modal) return;
        
        // Load auction details
        fetch(`/api/auctions/${auctionId}`)
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    this.renderBidModal(data.auction);
                    modal.classList.remove('hidden');
                } else {
                    alert('Müzayede bilgileri yüklenemedi.');
                }
            })
            .catch(error => {
                console.error('Error loading auction:', error);
                alert('Bir hata oluştu. Lütfen tekrar deneyin.');
            });
    },
    
    // Render bid modal content
    renderBidModal: function(auction) {
        const content = document.getElementById('bidModalContent');
        if (!content) return;
        
        const minBid = parseFloat(auction.current_bid) + 1;
        
        content.innerHTML = `
            <div class="text-center mb-4">
                <img src="${auction.images[0] || '/static/images/no-image.jpg'}" 
                     alt="${auction.title}" class="w-20 h-20 object-cover rounded mx-auto mb-3">
                <h4 class="font-semibold">${auction.title}</h4>
                <p class="text-sm text-gray-600">Mevcut Teklif: ₺${parseFloat(auction.current_bid).toFixed(2)}</p>
            </div>
            
            <form id="bidForm" class="space-y-4">
                <input type="hidden" name="auction_id" value="${auction.id}">
                
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-2">Teklif Miktarı (₺)</label>
                    <input type="number" 
                           name="bid_amount" 
                           id="bidAmount"
                           min="${minBid}"
                           step="0.01"
                           placeholder="${minBid}"
                           class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                           required>
                    <div class="text-xs text-gray-500 mt-1">
                        Minimum: ₺${minBid.toFixed(2)}
                    </div>
                </div>
                
                <div class="grid grid-cols-3 gap-2">
                    <button type="button" class="quick-bid px-3 py-2 text-sm bg-gray-100 hover:bg-gray-200 rounded-lg" 
                            data-amount="${minBid + 10}">
                        +₺10
                    </button>
                    <button type="button" class="quick-bid px-3 py-2 text-sm bg-gray-100 hover:bg-gray-200 rounded-lg" 
                            data-amount="${minBid + 50}">
                        +₺50
                    </button>
                    <button type="button" class="quick-bid px-3 py-2 text-sm bg-gray-100 hover:bg-gray-200 rounded-lg" 
                            data-amount="${minBid + 100}">
                        +₺100
                    </button>
                </div>
                
                <button type="submit" class="w-full bg-blue-600 hover:bg-blue-700 text-white py-3 px-4 rounded-lg font-medium transition-colors">
                    <i class="fas fa-gavel mr-2"></i>Teklif Ver
                </button>
            </form>
        `;
        
        // Add form submit handler
        const form = document.getElementById('bidForm');
        if (form) {
            form.addEventListener('submit', (e) => {
                e.preventDefault();
                this.submitBid(form);
            });
        }
    },
    
    // Submit bid
    submitBid: function(form) {
        const formData = new FormData(form);
        const submitBtn = form.querySelector('button[type="submit"]');
        
        // Show loading state
        const originalText = submitBtn.textContent;
        submitBtn.textContent = 'Teklif Veriliyor...';
        submitBtn.disabled = true;
        
        fetch('/api/auctions/bid', {
            method: 'POST',
            headers: {
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: formData
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                alert('Teklifiniz başarıyla verildi!');
                this.closeBidModal();
                // Refresh page or update UI
                window.location.reload();
            } else {
                alert(data.message || 'Teklif verilemedi. Lütfen tekrar deneyin.');
            }
        })
        .catch(error => {
            console.error('Bid submission error:', error);
            alert('Bir hata oluştu. Lütfen tekrar deneyin.');
        })
        .finally(() => {
            submitBtn.textContent = originalText;
            submitBtn.disabled = false;
        });
    },
    
    // Close bid modal
    closeBidModal: function() {
        const modal = document.getElementById('bidModal');
        if (modal) {
            modal.classList.add('hidden');
        }
    },
    
    // Toggle favorite status
    toggleFavorite: function(auctionId, button) {
        const icon = button.querySelector('i');
        const isFavorite = icon.classList.contains('text-red-500');
        
        fetch('/api/auctions/favorite', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: JSON.stringify({ auction_id: auctionId })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                if (data.is_favorite) {
                    icon.classList.add('text-red-500');
                    icon.classList.remove('text-gray-400');
                } else {
                    icon.classList.remove('text-red-500');
                    icon.classList.add('text-gray-400');
                }
            }
        })
        .catch(error => {
            console.error('Error toggling favorite:', error);
        });
    },
    
    // Toggle watchlist status
    toggleWatchlist: function(auctionId, button) {
        fetch('/api/auctions/watchlist', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
            },
            body: JSON.stringify({ auction_id: auctionId })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                const icon = button.querySelector('i');
                const text = button.querySelector('span') || button;
                
                if (data.is_watching) {
                    if (icon) icon.className = 'fas fa-eye-slash mr-2';
                    text.textContent = 'Takipten Çıkar';
                    button.classList.add('bg-red-200', 'text-red-700');
                    button.classList.remove('bg-gray-200', 'text-gray-700');
                } else {
                    if (icon) icon.className = 'fas fa-eye mr-2';
                    text.textContent = 'Takip Listesine Ekle';
                    button.classList.remove('bg-red-200', 'text-red-700');
                    button.classList.add('bg-gray-200', 'text-gray-700');
                }
            }
        })
        .catch(error => {
            console.error('Error toggling watchlist:', error);
        });
    },
    
    // Apply filters
    applyFilters: function() {
        const categoryFilter = document.getElementById('categoryFilter');
        const statusFilter = document.getElementById('statusFilter');
        const sortBy = document.getElementById('sortBy');
        
        const params = new URLSearchParams(window.location.search);
        
        if (categoryFilter && categoryFilter.value) {
            params.set('category', categoryFilter.value);
        } else {
            params.delete('category');
        }
        
        if (statusFilter && statusFilter.value) {
            params.set('status', statusFilter.value);
        } else {
            params.delete('status');
        }
        
        if (sortBy && sortBy.value) {
            params.set('sort', sortBy.value);
        } else {
            params.delete('sort');
        }
        
        // Update URL and reload
        const newUrl = window.location.pathname + '?' + params.toString();
        window.location.href = newUrl;
    },
    
    // Initialize filters from URL
    initFilters: function() {
        const params = new URLSearchParams(window.location.search);
        
        const categoryFilter = document.getElementById('categoryFilter');
        if (categoryFilter && params.get('category')) {
            categoryFilter.value = params.get('category');
        }
        
        const statusFilter = document.getElementById('statusFilter');
        if (statusFilter && params.get('status')) {
            statusFilter.value = params.get('status');
        }
        
        const sortBy = document.getElementById('sortBy');
        if (sortBy && params.get('sort')) {
            sortBy.value = params.get('sort');
        }
    },
    
    // Refresh auction data
    refreshAuctionData: function(auctionId) {
        fetch(`/api/auctions/${auctionId}`)
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    // Update auction card with new data
                    this.updateAuctionCard(data.auction);
                }
            })
            .catch(error => {
                console.error('Error refreshing auction data:', error);
            });
    },
    
    // Update auction card
    updateAuctionCard: function(auction) {
        const card = document.querySelector(`[data-auction-id="${auction.id}"]`);
        if (!card) return;
        
        // Update current bid
        const bidEl = card.querySelector('.text-green-600');
        if (bidEl) {
            bidEl.textContent = `₺${parseFloat(auction.current_bid).toFixed(2)}`;
        }
        
        // Update bid count
        const countEl = card.querySelector('.text-gray-500');
        if (countEl && countEl.textContent.includes('teklif')) {
            countEl.textContent = `${auction.bid_count} teklif`;
        }
    },
    
    // Show new bid notification
    showNewBidNotification: function(data) {
        // Only show if not current user's bid
        if (data.bidder_id === window.currentUserId) return;
        
        const notification = document.createElement('div');
        notification.className = 'fixed top-4 right-4 bg-blue-600 text-white px-6 py-3 rounded-lg shadow-lg z-50 transition-all duration-300';
        notification.innerHTML = `
            <div class="flex items-center">
                <i class="fas fa-gavel mr-2"></i>
                <div>
                    <div class="font-medium">Yeni Teklif!</div>
                    <div class="text-sm">₺${parseFloat(data.amount).toFixed(2)} - ${data.auction_title}</div>
                </div>
            </div>
        `;
        
        document.body.appendChild(notification);
        
        // Auto remove after 5 seconds
        setTimeout(() => {
            notification.style.opacity = '0';
            notification.style.transform = 'translateX(100%)';
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        }, 5000);
    },
    
    // Handle auction ended message
    handleAuctionEndedMessage: function(auctionId) {
        const card = document.querySelector(`[data-auction-id="${auctionId}"]`);
        if (card) {
            const countdownEl = card.querySelector('.countdown');
            if (countdownEl) {
                this.onAuctionEnd(countdownEl);
            }
        }
    },
    
    // Cleanup function
    destroy: function() {
        // Clear all countdowns
        this.countdowns.forEach(intervalId => {
            clearInterval(intervalId);
        });
        this.countdowns.clear();
        
        // Close WebSocket
        if (this.socket) {
            this.socket.close();
            this.socket = null;
        }
    }
};

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    window.AuctionManager.init();
});

// Cleanup on page unload
window.addEventListener('beforeunload', function() {
    window.AuctionManager.destroy();
});

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = window.AuctionManager;
}