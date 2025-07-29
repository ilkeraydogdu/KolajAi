# KolajAI Marketplace Integration Guide

## Table of Contents
1. [Overview](#overview)
2. [Supported Marketplaces](#supported-marketplaces)
3. [Getting Started](#getting-started)
4. [Marketplace Integrations](#marketplace-integrations)
   - [Trendyol](#trendyol-integration)
   - [Hepsiburada](#hepsiburada-integration)
   - [N11](#n11-integration)
   - [Amazon](#amazon-integration)
   - [Çiçeksepeti](#ciceksepeti-integration)
5. [Payment Integrations](#payment-integrations)
   - [Iyzico](#iyzico-integration)
6. [Best Practices](#best-practices)
7. [Troubleshooting](#troubleshooting)

## Overview

KolajAI provides seamless integration with major Turkish e-commerce marketplaces and payment providers. This guide will help you set up and manage these integrations effectively.

## Supported Marketplaces

| Marketplace | Features | Status |
|-------------|----------|---------|
| Trendyol | Product sync, Order management, Stock updates | ✅ Active |
| Hepsiburada | Product sync, Order management, Stock updates | ✅ Active |
| N11 | Product sync, Order management, Stock updates | ✅ Active |
| Amazon TR | Product sync, Order management, FBA support | ✅ Active |
| Çiçeksepeti | Product sync, Order management | ✅ Active |

## Getting Started

### Prerequisites
1. Active vendor account on KolajAI
2. API credentials for each marketplace
3. Products ready for synchronization

### Initial Setup
1. Navigate to **Settings > Integrations**
2. Select the marketplace you want to integrate
3. Enter your API credentials
4. Configure integration settings
5. Test the connection

## Marketplace Integrations

### Trendyol Integration

#### Required Credentials
- **API Key**: Your Trendyol API key
- **API Secret**: Your Trendyol API secret
- **Supplier ID**: Your Trendyol supplier ID

#### Configuration Steps

1. **Obtain API Credentials**
   - Log in to Trendyol Partner Panel
   - Navigate to Settings > API Management
   - Generate new API credentials

2. **Configure in KolajAI**
   ```json
   {
     "api_key": "your_api_key",
     "api_secret": "your_api_secret",
     "supplier_id": "your_supplier_id",
     "environment": "production"
   }
   ```

3. **Product Mapping**
   - Map your categories to Trendyol categories
   - Set up brand mappings
   - Configure attribute mappings

#### Product Synchronization

**Sync All Products:**
```bash
POST /api/integrations/trendyol/sync-products
{
  "action": "create",
  "product_ids": [] // Empty array syncs all products
}
```

**Sync Specific Products:**
```bash
POST /api/integrations/trendyol/sync-products
{
  "action": "update",
  "product_ids": [1, 2, 3]
}
```

#### Order Management

Orders are automatically imported every 15 minutes. You can also trigger manual import:

```bash
GET /api/integrations/trendyol/import-orders
```

#### Stock Updates

Stock updates are synchronized in real-time. You can configure buffer stock:

```json
{
  "stock_buffer": 5,
  "auto_update": true
}
```

### Hepsiburada Integration

#### Required Credentials
- **Merchant ID**: Your Hepsiburada merchant ID
- **Username**: API username
- **Password**: API password

#### Configuration Steps

1. **API Access Setup**
   - Contact Hepsiburada support for API access
   - Receive your merchant credentials
   - Enable API access in merchant panel

2. **Configure in KolajAI**
   ```json
   {
     "merchant_id": "your_merchant_id",
     "username": "api_username",
     "password": "api_password",
     "environment": "production"
   }
   ```

3. **Category Mapping**
   - Download Hepsiburada category tree
   - Map your categories appropriately
   - Set up required attributes

#### Product Requirements

Hepsiburada has specific requirements:
- High-quality images (minimum 1000x1000)
- Detailed product descriptions
- Valid barcodes (EAN/UPC)
- Accurate stock information

### N11 Integration

#### Required Credentials
- **API Key**: N11 API key
- **API Secret**: N11 API secret

#### SOAP API Configuration

N11 uses SOAP API. Configure endpoints:

```json
{
  "api_key": "your_api_key",
  "api_secret": "your_api_secret",
  "wsdl_url": "https://api.n11.com/ws/ProductService.wsdl"
}
```

#### Product Upload Process

1. **Category Selection**
   - Use N11 category service to find appropriate categories
   - Each product must have valid N11 category ID

2. **Product Attributes**
   - Mandatory attributes vary by category
   - Use attribute service to get required fields

3. **Image Requirements**
   - Maximum 8 images per product
   - Minimum resolution: 800x800
   - Maximum file size: 2MB

### Amazon Integration

#### Required Credentials
- **Merchant ID**: Your Amazon merchant ID
- **Access Key**: MWS access key
- **Secret Key**: MWS secret key
- **Marketplace ID**: Turkish marketplace ID

#### FBA (Fulfillment by Amazon) Support

Configure FBA settings:
```json
{
  "fulfillment_channel": "FBA",
  "prep_instructions": "PREP_NOT_REQUIRED",
  "ship_to_fc": true
}
```

#### Feed Management

Amazon uses feed-based updates:
1. Product feeds
2. Inventory feeds
3. Price feeds
4. Image feeds

Monitor feed status:
```bash
GET /api/integrations/amazon/feed-status/{feedId}
```

### Çiçeksepeti Integration

#### Special Considerations
- Products must be appropriate for gifting
- Delivery date selection is mandatory
- Special occasion tags improve visibility

#### Configuration
```json
{
  "api_key": "your_api_key",
  "branch_code": "your_branch_code",
  "delivery_options": {
    "same_day": true,
    "next_day": true,
    "scheduled": true
  }
}
```

## Payment Integrations

### Iyzico Integration

#### Setup Process

1. **Create Iyzico Account**
   - Register at iyzico.com
   - Complete merchant verification
   - Obtain API credentials

2. **Configure in KolajAI**
   ```json
   {
     "api_key": "your_api_key",
     "secret_key": "your_secret_key",
     "base_url": "https://api.iyzipay.com",
     "enable_3d_secure": true,
     "enable_installments": true
   }
   ```

3. **Test Environment**
   - Use sandbox credentials for testing
   - Test with provided test cards
   - Verify webhook integration

#### Supported Features
- ✅ Credit/Debit card payments
- ✅ 3D Secure
- ✅ Installments (up to 12 months)
- ✅ Refunds
- ✅ BIN checking
- ✅ Fraud protection

#### Webhook Configuration

Configure webhook endpoint:
```
https://your-domain.com/api/webhooks/iyzico
```

Handle webhook events:
- Payment success
- Payment failure
- Refund completion
- Fraud alerts

## Best Practices

### 1. Inventory Management
- Maintain buffer stock to prevent overselling
- Use automatic stock synchronization
- Set up low stock alerts
- Regular inventory audits

### 2. Pricing Strategy
- Consider marketplace commissions
- Set competitive prices using analytics
- Use dynamic pricing where appropriate
- Monitor competitor prices

### 3. Product Information
- Use high-quality images
- Write detailed descriptions
- Include all specifications
- Optimize for marketplace search

### 4. Order Processing
- Enable automatic order import
- Set up order status mapping
- Configure shipping templates
- Implement tracking updates

### 5. Error Handling
- Monitor integration logs
- Set up error notifications
- Implement retry mechanisms
- Keep credentials secure

### 6. Performance Optimization
- Use bulk operations when possible
- Implement caching for frequently accessed data
- Schedule heavy operations during off-peak hours
- Monitor API rate limits

## Troubleshooting

### Common Issues

#### 1. Authentication Failures
**Problem:** "Invalid credentials" error
**Solution:**
- Verify API credentials are correct
- Check if credentials are for correct environment
- Ensure API access is enabled
- Regenerate credentials if necessary

#### 2. Product Sync Failures
**Problem:** Products not appearing on marketplace
**Solution:**
- Check category mappings
- Verify required attributes are filled
- Ensure images meet requirements
- Review marketplace-specific rules

#### 3. Stock Discrepancies
**Problem:** Stock levels don't match
**Solution:**
- Check stock buffer settings
- Verify automatic sync is enabled
- Look for failed update logs
- Manually trigger stock sync

#### 4. Order Import Issues
**Problem:** Orders not importing
**Solution:**
- Check webhook configuration
- Verify order status mappings
- Review import logs for errors
- Test manual import

### Debug Mode

Enable debug mode for detailed logs:
```json
{
  "debug_mode": true,
  "log_level": "verbose",
  "log_requests": true
}
```

### Support Channels

**Technical Support:**
- Email: integration-support@kolajAi.com
- Documentation: docs.kolajAi.com/integrations
- API Status: status.kolajAi.com

**Marketplace Support:**
- Trendyol: entegrasyon@trendyol.com
- Hepsiburada: api-support@hepsiburada.com
- N11: api@n11.com
- Amazon: seller-support@amazon.com.tr

## Appendix

### Rate Limits

| Marketplace | Requests/Min | Requests/Hour | Requests/Day |
|-------------|--------------|---------------|--------------|
| Trendyol | 60 | 3,600 | 86,400 |
| Hepsiburada | 100 | 6,000 | 144,000 |
| N11 | 30 | 1,800 | 43,200 |
| Amazon | 30 | 1,800 | 43,200 |
| Çiçeksepeti | 60 | 3,600 | 86,400 |

### Error Codes

| Code | Description | Action |
|------|-------------|--------|
| 401 | Unauthorized | Check credentials |
| 403 | Forbidden | Verify permissions |
| 429 | Rate limit exceeded | Implement backoff |
| 500 | Server error | Retry with backoff |
| 503 | Service unavailable | Wait and retry |