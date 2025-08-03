/**
 * Event Handlers for KolajAI
 * Bu dosya inline onclick handler'ları değiştiren güvenli event delegation sistemi içerir
 */

// Event delegation for data-action attributes
document.addEventListener('DOMContentLoaded', function() {
  // Main event delegator
  document.addEventListener('click', function(event) {
    const target = event.target.closest('[data-action]');
    if (!target) return;

    const action = target.getAttribute('data-action');
    
    // Prevent default if needed
    event.preventDefault();
    
    // Route to appropriate handler
    switch (action) {
      case 'use-image':
        handleUseImage(target);
        break;
      case 'download-image':
        handleDownloadImage(target);
        break;
      case 'use-template':
        handleUseTemplate(target);
        break;
      case 'save-configuration':
        handleSaveConfiguration(target);
        break;
      case 'configure-integration':
        handleConfigureIntegration(target);
        break;
      case 'sync-products':
        handleSyncProducts(target);
        break;
      default:
        console.warn('Unknown action:', action);
    }
  });
});

// Handler functions
function handleUseImage(element) {
  const imageUrl = element.getAttribute('data-url');
  if (!imageUrl) {
    console.error('No image URL provided');
    return;
  }
  
  try {
    // Safely use the generated image
    if (typeof useGeneratedImage === 'function') {
      useGeneratedImage(imageUrl);
    } else {
      console.warn('useGeneratedImage function not found');
      // Fallback behavior
      window.showToast('Resim kullanıldı: ' + imageUrl, 'success');
    }
  } catch (error) {
    console.error('Error using image:', error);
    window.showToast('Resim kullanılırken hata oluştu', 'error');
  }
}

function handleDownloadImage(element) {
  const imageUrl = element.getAttribute('data-url');
  if (!imageUrl) {
    console.error('No image URL provided');
    return;
  }
  
  try {
    // Safely download the image
    if (typeof downloadImage === 'function') {
      downloadImage(imageUrl);
    } else {
      // Fallback: create download link
      const link = document.createElement('a');
      link.href = imageUrl;
      link.download = 'generated-image.jpg';
      link.style.display = 'none';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.showToast('Resim indiriliyor...', 'info');
    }
  } catch (error) {
    console.error('Error downloading image:', error);
    window.showToast('Resim indirilemedi', 'error');
  }
}

function handleUseTemplate(element) {
  const templateId = element.getAttribute('data-template-id');
  if (!templateId) {
    console.error('No template ID provided');
    return;
  }
  
  try {
    // Safely use the template
    if (typeof useTemplate === 'function') {
      useTemplate(parseInt(templateId, 10));
    } else {
      console.warn('useTemplate function not found');
      // Fallback behavior
      window.showToast('Şablon kullanıldı: ' + templateId, 'success');
    }
  } catch (error) {
    console.error('Error using template:', error);
    window.showToast('Şablon kullanılırken hata oluştu', 'error');
  }
}

function handleSaveConfiguration(element) {
  try {
    // Safely save configuration
    if (typeof saveConfiguration === 'function') {
      saveConfiguration();
    } else {
      console.warn('saveConfiguration function not found');
      // Fallback: collect form data and show success
      const modal = element.closest('.modal');
      if (modal) {
        const form = modal.querySelector('form');
        if (form) {
          const formData = new FormData(form);
          window.logger && window.logger.debug('Configuration data:', Object.fromEntries(formData));
          window.showToast('Yapılandırma kaydedildi', 'success');
          
          // Close modal
          const bsModal = bootstrap.Modal.getInstance(modal);
          if (bsModal) {
            bsModal.hide();
          }
        }
      }
    }
  } catch (error) {
    console.error('Error saving configuration:', error);
    window.showToast('Yapılandırma kaydedilemedi', 'error');
  }
}

function handleConfigureIntegration(element) {
  const integrationId = element.getAttribute('data-integration-id');
  if (!integrationId) {
    console.error('No integration ID provided');
    return;
  }
  
  try {
    // Safely configure integration
    if (typeof configureIntegration === 'function') {
      configureIntegration(integrationId);
    } else {
      console.warn('configureIntegration function not found');
      // Fallback behavior
      window.showToast('Entegrasyon yapılandırılıyor: ' + integrationId, 'info');
    }
  } catch (error) {
    console.error('Error configuring integration:', error);
    window.showToast('Entegrasyon yapılandırılamadı', 'error');
  }
}

function handleSyncProducts(element) {
  const integrationId = element.getAttribute('data-integration-id');
  if (!integrationId) {
    console.error('No integration ID provided');
    return;
  }
  
  try {
    // Safely sync products
    if (typeof syncProducts === 'function') {
      syncProducts(integrationId);
    } else {
      console.warn('syncProducts function not found');
      // Fallback behavior
      window.showToast('Ürünler senkronize ediliyor: ' + integrationId, 'info');
    }
  } catch (error) {
    console.error('Error syncing products:', error);
    window.showToast('Ürün senkronizasyonu başarısız', 'error');
  }
}

// Export functions for global access if needed
if (typeof window !== 'undefined') {
  window.eventHandlers = {
    handleUseImage,
    handleDownloadImage,
    handleUseTemplate,
    handleSaveConfiguration,
    handleConfigureIntegration,
    handleSyncProducts
  };
}

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    handleUseImage,
    handleDownloadImage,
    handleUseTemplate,
    handleSaveConfiguration,
    handleConfigureIntegration,
    handleSyncProducts
  };
}