/**
 * Event Manager for KolajAI
 * Bu dosya event listener'ları yöneterek memory leak'leri önler
 */

class EventManager {
  constructor() {
    this.eventListeners = new Map();
    this.abortControllers = new Map();
    this.boundHandlers = new Map();
    
    // Cleanup on page unload
    this.setupUnloadHandler();
  }

  /**
   * Add event listener with automatic cleanup tracking
   * @param {Element|Window|Document} target - Event target
   * @param {string} type - Event type
   * @param {Function} handler - Event handler
   * @param {Object} options - Event options
   * @returns {string} - Event ID for manual removal
   */
  addEventListener(target, type, handler, options = {}) {
    const eventId = this.generateEventId();
    
    // Create abort controller for this event
    const controller = new AbortController();
    
    // Add abort signal to options
    const eventOptions = {
      ...options,
      signal: controller.signal
    };
    
    // Store event info for cleanup
    const eventInfo = {
      target,
      type,
      handler,
      options: eventOptions,
      controller
    };
    
    this.eventListeners.set(eventId, eventInfo);
    this.abortControllers.set(eventId, controller);
    
    // Add the actual event listener
    target.addEventListener(type, handler, eventOptions);
    
    return eventId;
  }

  /**
   * Remove specific event listener
   * @param {string} eventId - Event ID returned from addEventListener
   */
  removeEventListener(eventId) {
    const controller = this.abortControllers.get(eventId);
    if (controller) {
      controller.abort();
      this.abortControllers.delete(eventId);
      this.eventListeners.delete(eventId);
    }
  }

  /**
   * Remove all event listeners for a specific target
   * @param {Element|Window|Document} target - Event target
   */
  removeAllEventListeners(target) {
    const toRemove = [];
    
    for (const [eventId, eventInfo] of this.eventListeners) {
      if (eventInfo.target === target) {
        toRemove.push(eventId);
      }
    }
    
    toRemove.forEach(eventId => this.removeEventListener(eventId));
  }

  /**
   * Remove all event listeners of a specific type
   * @param {string} type - Event type
   */
  removeEventListenersByType(type) {
    const toRemove = [];
    
    for (const [eventId, eventInfo] of this.eventListeners) {
      if (eventInfo.type === type) {
        toRemove.push(eventId);
      }
    }
    
    toRemove.forEach(eventId => this.removeEventListener(eventId));
  }

  /**
   * Add delegated event listener (more memory efficient for dynamic content)
   * @param {Element} container - Container element
   * @param {string} selector - CSS selector for target elements
   * @param {string} type - Event type
   * @param {Function} handler - Event handler
   * @param {Object} options - Event options
   * @returns {string} - Event ID
   */
  addDelegatedEventListener(container, selector, type, handler, options = {}) {
    const delegatedHandler = (event) => {
      const target = event.target.closest(selector);
      if (target && container.contains(target)) {
        // Call handler with proper context
        handler.call(target, event);
      }
    };
    
    return this.addEventListener(container, type, delegatedHandler, options);
  }

  /**
   * Add throttled event listener
   * @param {Element|Window|Document} target - Event target
   * @param {string} type - Event type
   * @param {Function} handler - Event handler
   * @param {number} delay - Throttle delay in ms
   * @param {Object} options - Event options
   * @returns {string} - Event ID
   */
  addThrottledEventListener(target, type, handler, delay = 100, options = {}) {
    const throttledHandler = this.throttle(handler, delay);
    return this.addEventListener(target, type, throttledHandler, options);
  }

  /**
   * Add debounced event listener
   * @param {Element|Window|Document} target - Event target
   * @param {string} type - Event type
   * @param {Function} handler - Event handler
   * @param {number} delay - Debounce delay in ms
   * @param {Object} options - Event options
   * @returns {string} - Event ID
   */
  addDebouncedEventListener(target, type, handler, delay = 300, options = {}) {
    const debouncedHandler = this.debounce(handler, delay);
    return this.addEventListener(target, type, debouncedHandler, options);
  }

  /**
   * Add one-time event listener that auto-removes after first trigger
   * @param {Element|Window|Document} target - Event target
   * @param {string} type - Event type
   * @param {Function} handler - Event handler
   * @param {Object} options - Event options
   * @returns {string} - Event ID
   */
  addOneTimeEventListener(target, type, handler, options = {}) {
    const eventId = this.generateEventId();
    
    const oneTimeHandler = (event) => {
      handler(event);
      this.removeEventListener(eventId);
    };
    
    return this.addEventListener(target, type, oneTimeHandler, options);
  }

  /**
   * Throttle function
   */
  throttle(func, delay) {
    let timeoutId;
    let lastExecTime = 0;
    
    return function (...args) {
      const currentTime = Date.now();
      
      if (currentTime - lastExecTime > delay) {
        func.apply(this, args);
        lastExecTime = currentTime;
      } else {
        clearTimeout(timeoutId);
        timeoutId = setTimeout(() => {
          func.apply(this, args);
          lastExecTime = Date.now();
        }, delay - (currentTime - lastExecTime));
      }
    };
  }

  /**
   * Debounce function
   */
  debounce(func, delay) {
    let timeoutId;
    
    return function (...args) {
      clearTimeout(timeoutId);
      timeoutId = setTimeout(() => func.apply(this, args), delay);
    };
  }

  /**
   * Generate unique event ID
   */
  generateEventId() {
    return `event_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * Setup page unload handler for cleanup
   */
  setupUnloadHandler() {
    const cleanup = () => {
      this.cleanup();
    };
    
    // Use multiple events to ensure cleanup
    window.addEventListener('beforeunload', cleanup);
    window.addEventListener('unload', cleanup);
    window.addEventListener('pagehide', cleanup);
    
    // For SPAs, also listen to custom events
    window.addEventListener('app:cleanup', cleanup);
  }

  /**
   * Manual cleanup - removes all event listeners
   */
  cleanup() {
    // Abort all controllers
    for (const controller of this.abortControllers.values()) {
      try {
        controller.abort();
      } catch (error) {
        // Ignore errors during cleanup
      }
    }
    
    // Clear all maps
    this.eventListeners.clear();
    this.abortControllers.clear();
    this.boundHandlers.clear();
    
    if (window.logger) {
      window.logger.debug('EventManager: All event listeners cleaned up');
    }
  }

  /**
   * Get statistics about registered events
   */
  getStats() {
    const stats = {
      totalEvents: this.eventListeners.size,
      eventsByType: {},
      eventsByTarget: {}
    };
    
    for (const eventInfo of this.eventListeners.values()) {
      // Count by type
      stats.eventsByType[eventInfo.type] = (stats.eventsByType[eventInfo.type] || 0) + 1;
      
      // Count by target type
      const targetType = eventInfo.target.constructor.name;
      stats.eventsByTarget[targetType] = (stats.eventsByTarget[targetType] || 0) + 1;
    }
    
    return stats;
  }

  /**
   * Check for potential memory leaks
   */
  checkForLeaks() {
    const stats = this.getStats();
    const warnings = [];
    
    if (stats.totalEvents > 1000) {
      warnings.push(`High number of event listeners: ${stats.totalEvents}`);
    }
    
    // Check for excessive listeners of same type
    for (const [type, count] of Object.entries(stats.eventsByType)) {
      if (count > 100) {
        warnings.push(`Excessive ${type} listeners: ${count}`);
      }
    }
    
    if (warnings.length > 0 && window.logger) {
      window.logger.warn('EventManager: Potential memory leaks detected:', warnings);
    }
    
    return warnings;
  }
}

// Create global event manager instance
const eventManager = new EventManager();

// Export for global access
if (typeof window !== 'undefined') {
  window.eventManager = eventManager;
  
  // Provide convenient global functions
  window.addEvent = (target, type, handler, options) => 
    eventManager.addEventListener(target, type, handler, options);
  
  window.removeEvent = (eventId) => 
    eventManager.removeEventListener(eventId);
  
  window.addDelegatedEvent = (container, selector, type, handler, options) =>
    eventManager.addDelegatedEventListener(container, selector, type, handler, options);
}

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
  module.exports = EventManager;
}

export default EventManager;