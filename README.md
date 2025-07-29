# KolajAI Enterprise Marketplace

## Overview

KolajAI Enterprise Marketplace is an advanced, AI-powered e-commerce platform built with Go. This enterprise-level application includes real marketplace integrations with production-ready API implementations for major Turkish e-commerce platforms like Trendyol and Hepsiburada.

## Enterprise Features

### 🔐 Advanced Security System
- **Multi-layer Security**: IP whitelisting/blacklisting, rate limiting, input validation
- **Vulnerability Scanning**: Real-time threat detection for SQL injection, XSS, CSRF
- **Security Headers**: Comprehensive HTTP security headers
- **Two-Factor Authentication**: Optional 2FA support
- **Audit Logging**: Complete security event logging and monitoring

### 🚀 Performance Optimization
- **Advanced Caching**: Multi-store cache system (Memory, Redis, Database)
- **Compression**: Automatic gzip compression for responses
- **Cache Invalidation**: Tag-based and TTL-based cache management
- **Load Balancing Ready**: Designed for horizontal scaling

### 📊 Dynamic Reporting System
- **Configurable Reports**: Custom report generation with filters and grouping
- **User Behavior Analysis**: Detailed user activity and purchasing pattern analysis
- **Real-time Analytics**: Live dashboard with performance metrics
- **Export Capabilities**: Multiple format support (JSON, CSV, PDF)

### 🔧 Session & Cookie Management
- **Database-backed Sessions**: Persistent session storage
- **Session Analytics**: User activity tracking and device information
- **Secure Cookies**: HttpOnly, Secure, SameSite configuration
- **Session Cleanup**: Automatic cleanup of expired sessions

### 🌐 SEO & Multi-language Support
- **Dynamic Sitemap**: Auto-generated XML sitemaps
- **Multi-language Content**: Full internationalization support
- **Meta Tag Management**: Dynamic meta tags and structured data
- **Search Engine Optimization**: Built-in SEO tools and analytics

### 📢 Notification System
- **Multi-channel Notifications**: Email, SMS, push notifications
- **Template System**: Customizable notification templates
- **Scheduling**: Delayed and recurring notifications
- **User Preferences**: Opt-out and quiet hours management

### 🧪 Advanced Testing Framework
- **Multiple Test Types**: Unit, integration, API, UI, performance, security tests
- **Code Coverage**: Detailed coverage reporting
- **Parallel Execution**: Concurrent test running
- **Test Reporting**: Comprehensive test result analysis

### ❌ Error Management
- **Centralized Error Handling**: Structured error logging and management
- **Error Grouping**: Similar error aggregation
- **Notification Rules**: Configurable error alerting
- **Stack Trace Analysis**: Detailed error context and debugging information

### 🎛️ Advanced Admin Panel
- **Real-time Dashboard**: Live statistics and metrics
- **User Management**: Detailed user profiles and behavior analysis
- **Content Management**: Product, vendor, and order management
- **System Health**: Server monitoring and performance tracking
- **Configuration Management**: Dynamic system settings

## Technical Architecture

### Core Technologies
- **Backend**: Go 1.23+
- **Database**: SQLite (with MySQL support)
- **Caching**: In-memory with Redis support
- **Templates**: Go HTML templates
- **Security**: Custom security middleware stack

### Advanced Systems
- **Middleware Stack**: Layered security, caching, and logging
- **Router System**: Advanced routing with group support
- **Configuration Management**: YAML-based configuration
- **Dependency Injection**: Service-oriented architecture

## Installation & Setup

### Prerequisites
- Go 1.23 or higher
- SQLite3 (or MySQL for production)
- Optional: Redis for advanced caching

### Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd kolajAi
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Configure the application**
   ```bash
   cp config.yaml.example config.yaml
   # Edit config.yaml with your settings
   ```

4. **Run database migrations**
   ```bash
   go run cmd/server/main.go
   ```

5. **Start the server**
   ```bash
   go run cmd/server/main.go
   ```

The application will be available at `http://localhost:8081`

## Configuration

The application uses a YAML configuration file (`config.yaml`) with the following sections:

- **Server**: Port, host, timeouts
- **Database**: Connection settings
- **Security**: Encryption keys, authentication settings
- **Cache**: Cache configuration and limits
- **Email**: SMTP settings for notifications
- **SEO**: Site metadata and search engine settings
- **Logging**: Log levels and output configuration

## API Endpoints

### Public Endpoints
- `GET /` - Homepage/Marketplace
- `GET /products` - Product listings
- `GET /product/{id}` - Product details
- `POST /login` - User authentication
- `POST /register` - User registration

### Protected Endpoints
- `GET /dashboard` - User dashboard
- `GET /cart` - Shopping cart
- `POST /add-to-cart` - Add items to cart
- `GET /vendor/dashboard` - Vendor management

### Admin Endpoints
- `GET /admin/dashboard` - Admin dashboard
- `GET /admin/users` - User management
- `GET /admin/products` - Product management
- `GET /admin/reports` - Reporting system
- `GET /admin/seo` - SEO management
- `GET /admin/system` - System health

### API Endpoints
- `GET /api/search` - Product search
- `POST /api/cart/update` - Cart updates
- `GET /api/ai/recommendations` - AI recommendations
- `GET /health` - Health check
- `GET /metrics` - Application metrics

## Security Features

### Authentication & Authorization
- Session-based authentication
- Role-based access control (Admin, Vendor, User)
- CSRF protection
- XSS prevention

### Data Protection
- Input validation and sanitization
- SQL injection prevention
- Secure password hashing
- Data encryption at rest

### Network Security
- Rate limiting
- IP-based access control
- HTTPS enforcement
- Security headers

## Performance Features

### Caching Strategy
- Page-level caching for static content
- Database query result caching
- Session data caching
- Asset caching with versioning

### Optimization
- Gzip compression
- Minified assets
- Database connection pooling
- Efficient query optimization

## Monitoring & Analytics

### System Monitoring
- Real-time performance metrics
- Error tracking and alerting
- Resource usage monitoring
- Health check endpoints

### Business Analytics
- User behavior tracking
- Sales analytics
- Product performance metrics
- Custom report generation

## Development

### Project Structure
```
kolajAi/
├── cmd/                    # Application entry points
│   ├── server/            # Main server application
│   ├── seed/              # Database seeding tools
│   └── db-tools/          # Database utilities
├── internal/              # Private application code
│   ├── cache/             # Caching system
│   ├── config/            # Configuration management
│   ├── database/          # Database layer
│   ├── errors/            # Error management
│   ├── handlers/          # HTTP handlers
│   ├── middleware/        # HTTP middleware
│   ├── models/            # Data models
│   ├── notifications/     # Notification system
│   ├── reporting/         # Reporting system
│   ├── router/            # Routing system
│   ├── security/          # Security management
│   ├── seo/               # SEO management
│   ├── services/          # Business logic
│   ├── session/           # Session management
│   └── testing/           # Testing framework
├── web/                   # Web assets
│   ├── static/            # Static files
│   └── templates/         # HTML templates
├── config.yaml            # Configuration file
└── go.mod                 # Go modules
```

### Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test types
go test -tags=integration ./...
go test -tags=security ./...
```

### Building for Production

```bash
# Build the application
go build -o kolajAi cmd/server/main.go

# Build with optimizations
go build -ldflags="-s -w" -o kolajAi cmd/server/main.go
```

## Deployment

### Docker Deployment
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o kolajAi cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/kolajAi .
COPY --from=builder /app/web ./web
COPY --from=builder /app/config.yaml .
CMD ["./kolajAi"]
```

### Environment Variables
Key environment variables for production:
- `APP_ENV=production`
- `DB_DRIVER=mysql`
- `DB_HOST=your-db-host`
- `ENCRYPTION_KEY=your-32-byte-key`
- `JWT_SECRET=your-jwt-secret`

## Contributing

1. Fork the repository
2. Create a feature branch
3. Implement your changes with tests
4. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support and questions:
- Create an issue on GitHub
- Check the documentation
- Review the configuration examples

## Changelog

### Version 2.0.0 (Enterprise)
- ✅ Advanced security system with threat detection
- ✅ Performance optimization with multi-layer caching
- ✅ Dynamic reporting and analytics system
- ✅ Advanced session and cookie management
- ✅ SEO and multi-language support
- ✅ Centralized notification system
- ✅ Comprehensive testing framework
- ✅ Advanced error management
- ✅ Enterprise-grade admin panel
- ✅ Configuration management system
- ✅ Middleware and routing system

### Version 1.0.0 (Basic)
- Basic e-commerce functionality
- User authentication
- Product management
- Order processing
- AI features