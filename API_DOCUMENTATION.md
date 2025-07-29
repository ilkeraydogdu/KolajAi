# KolajAI Enterprise Marketplace API Documentation

## Table of Contents
1. [Introduction](#introduction)
2. [Authentication](#authentication)
3. [Base URL](#base-url)
4. [Error Handling](#error-handling)
5. [Rate Limiting](#rate-limiting)
6. [API Endpoints](#api-endpoints)
   - [Authentication](#authentication-endpoints)
   - [Users](#user-endpoints)
   - [Products](#product-endpoints)
   - [Orders](#order-endpoints)
   - [Marketplace Integration](#marketplace-integration-endpoints)
   - [Payment](#payment-endpoints)
   - [AI Services](#ai-service-endpoints)
7. [Webhooks](#webhooks)
8. [Examples](#examples)

## Introduction

The KolajAI Enterprise Marketplace API provides programmatic access to all marketplace features including product management, order processing, marketplace integrations, and AI services.

### API Version
Current version: `v1`

### Request/Response Format
- All requests and responses are in JSON format
- UTF-8 encoding is used throughout
- Dates are in ISO 8601 format

## Authentication

The API uses session-based authentication with CSRF protection.

### Login
```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "secure_password"
}
```

**Response:**
```json
{
  "success": true,
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "role": "user"
  },
  "session_token": "session_token_here"
}
```

### Headers Required for Authenticated Requests
```http
Cookie: session=session_token_here
X-CSRF-Token: csrf_token_here
```

## Base URL

- Development: `http://localhost:8081/api`
- Production: `https://api.kolajAi.com/v1`

## Error Handling

### Error Response Format
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": {
      "field": "Additional error details"
    }
  },
  "request_id": "unique_request_id"
}
```

### Common Error Codes
| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Authentication required |
| `FORBIDDEN` | 403 | Access denied |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Invalid request data |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Server error |

## Rate Limiting

- Rate limits are applied per user/IP
- Default limits: 100 requests per minute
- Headers included in response:
  - `X-RateLimit-Limit`: Request limit
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Reset timestamp

## API Endpoints

### Authentication Endpoints

#### Register User
```http
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "secure_password",
  "name": "John Doe",
  "phone": "+905551234567"
}
```

#### Logout
```http
POST /api/auth/logout
```

#### Get Current User
```http
GET /api/auth/me
```

### User Endpoints

#### Get User Profile
```http
GET /api/users/{userId}
```

#### Update User Profile
```http
PUT /api/users/{userId}
Content-Type: application/json

{
  "name": "Updated Name",
  "phone": "+905551234567",
  "address": "New Address"
}
```

#### Get User Orders
```http
GET /api/users/{userId}/orders?page=1&limit=20&status=completed
```

### Product Endpoints

#### List Products
```http
GET /api/products?page=1&limit=20&category=electronics&sort=price_asc
```

**Query Parameters:**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 20, max: 100)
- `category` (string): Filter by category
- `vendor_id` (int): Filter by vendor
- `min_price` (float): Minimum price
- `max_price` (float): Maximum price
- `sort` (string): Sort order (price_asc, price_desc, created_at_desc)
- `search` (string): Search in name and description

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "name": "Product Name",
      "description": "Product description",
      "price": 99.99,
      "stock": 100,
      "category": "electronics",
      "vendor": {
        "id": 1,
        "name": "Vendor Name"
      },
      "images": ["url1", "url2"],
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

#### Get Product Details
```http
GET /api/products/{productId}
```

#### Create Product (Vendor Only)
```http
POST /api/products
Content-Type: application/json

{
  "name": "New Product",
  "description": "Product description",
  "price": 99.99,
  "stock": 100,
  "category": "electronics",
  "images": ["base64_image_data"],
  "attributes": {
    "color": "Black",
    "size": "Large"
  }
}
```

#### Update Product (Vendor Only)
```http
PUT /api/products/{productId}
Content-Type: application/json

{
  "name": "Updated Product Name",
  "price": 89.99,
  "stock": 150
}
```

#### Delete Product (Vendor Only)
```http
DELETE /api/products/{productId}
```

### Order Endpoints

#### Create Order
```http
POST /api/orders
Content-Type: application/json

{
  "items": [
    {
      "product_id": 1,
      "quantity": 2
    }
  ],
  "shipping_address": {
    "name": "John Doe",
    "address": "123 Main St",
    "city": "Istanbul",
    "postal_code": "34000",
    "country": "Turkey"
  },
  "payment_method": "credit_card"
}
```

#### Get Order Details
```http
GET /api/orders/{orderId}
```

#### Update Order Status (Vendor Only)
```http
PUT /api/orders/{orderId}/status
Content-Type: application/json

{
  "status": "shipped",
  "tracking_number": "TR123456789"
}
```

#### Cancel Order
```http
POST /api/orders/{orderId}/cancel
Content-Type: application/json

{
  "reason": "Customer request"
}
```

### Marketplace Integration Endpoints

#### List Integrations
```http
GET /api/integrations
```

#### Get Integration Details
```http
GET /api/integrations/{integrationId}
```

#### Sync Products to Marketplace
```http
POST /api/integrations/{integrationId}/sync-products
Content-Type: application/json

{
  "product_ids": [1, 2, 3],
  "action": "create" // create, update, delete
}
```

#### Get Marketplace Orders
```http
GET /api/integrations/{integrationId}/orders?status=pending&date_from=2024-01-01
```

#### Update Integration Settings
```http
PUT /api/integrations/{integrationId}/settings
Content-Type: application/json

{
  "auto_sync": true,
  "sync_interval": 3600,
  "price_markup": 10
}
```

### Payment Endpoints

#### Process Payment
```http
POST /api/payments/process
Content-Type: application/json

{
  "order_id": 123,
  "payment_method": "credit_card",
  "card": {
    "number": "4111111111111111",
    "holder_name": "John Doe",
    "exp_month": "12",
    "exp_year": "2025",
    "cvv": "123"
  }
}
```

#### Get Payment Status
```http
GET /api/payments/{paymentId}
```

#### Refund Payment
```http
POST /api/payments/{paymentId}/refund
Content-Type: application/json

{
  "amount": 50.00,
  "reason": "Customer request"
}
```

### AI Service Endpoints

#### Generate Product Description
```http
POST /api/ai/generate-description
Content-Type: application/json

{
  "product_name": "Wireless Headphones",
  "category": "Electronics",
  "features": ["Bluetooth 5.0", "40-hour battery", "Noise cancellation"],
  "target_audience": "Music enthusiasts"
}
```

#### Generate Product Image
```http
POST /api/ai/generate-image
Content-Type: application/json

{
  "prompt": "Modern wireless headphones, product photography, white background",
  "style": "product_photo",
  "size": "1024x1024"
}
```

#### Analyze Product Image
```http
POST /api/ai/analyze-image
Content-Type: application/json

{
  "image": "base64_encoded_image_data",
  "analysis_type": "product_quality"
}
```

#### Get AI Recommendations
```http
GET /api/ai/recommendations?user_id=123&type=products&limit=10
```

## Webhooks

### Webhook Configuration
```http
POST /api/webhooks
Content-Type: application/json

{
  "url": "https://your-domain.com/webhook",
  "events": ["order.created", "order.updated", "product.stock_low"],
  "secret": "your_webhook_secret"
}
```

### Webhook Events

#### Order Created
```json
{
  "event": "order.created",
  "timestamp": "2024-01-01T00:00:00Z",
  "data": {
    "order_id": 123,
    "total": 99.99,
    "items": [...]
  }
}
```

#### Product Stock Low
```json
{
  "event": "product.stock_low",
  "timestamp": "2024-01-01T00:00:00Z",
  "data": {
    "product_id": 1,
    "current_stock": 5,
    "threshold": 10
  }
}
```

### Webhook Security

All webhooks include a signature header:
```http
X-Webhook-Signature: sha256=signature_here
```

Verify the signature using:
```javascript
const crypto = require('crypto');

function verifyWebhookSignature(payload, signature, secret) {
  const hash = crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
  
  return `sha256=${hash}` === signature;
}
```

## Examples

### cURL Example - Create Product
```bash
curl -X POST https://api.kolajAi.com/v1/products \
  -H "Content-Type: application/json" \
  -H "Cookie: session=your_session_token" \
  -H "X-CSRF-Token: your_csrf_token" \
  -d '{
    "name": "Wireless Mouse",
    "price": 29.99,
    "stock": 50,
    "category": "electronics"
  }'
```

### JavaScript Example - Fetch Products
```javascript
const fetchProducts = async () => {
  const response = await fetch('https://api.kolajAi.com/v1/products?page=1&limit=10', {
    method: 'GET',
    headers: {
      'Accept': 'application/json'
    },
    credentials: 'include'
  });
  
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  
  const data = await response.json();
  return data;
};
```

### Python Example - Process Payment
```python
import requests

def process_payment(order_id, card_details):
    url = "https://api.kolajAi.com/v1/payments/process"
    
    payload = {
        "order_id": order_id,
        "payment_method": "credit_card",
        "card": card_details
    }
    
    headers = {
        "Content-Type": "application/json",
        "X-CSRF-Token": "your_csrf_token"
    }
    
    response = requests.post(
        url, 
        json=payload, 
        headers=headers,
        cookies={"session": "your_session_token"}
    )
    
    return response.json()
```

## SDK Libraries

Official SDKs are available for:
- JavaScript/TypeScript: `npm install @kolajAi/marketplace-sdk`
- Python: `pip install kolajAi-marketplace`
- Go: `go get github.com/kolajAi/marketplace-sdk-go`

## Support

For API support:
- Email: api-support@kolajAi.com
- Documentation: https://docs.kolajAi.com
- Status Page: https://status.kolajAi.com