# KolajAI Marketplace

KolajAI is a modern, AI-powered e-commerce marketplace built with Go, featuring advanced analytics, intelligent recommendations, and comprehensive vendor management.

## 🚀 Features

### Core Marketplace Features
- **Product Management**: Full CRUD operations for products with categories, variants, and images
- **Vendor System**: Multi-vendor marketplace with approval workflows
- **Order Management**: Complete order lifecycle with status tracking
- **Auction System**: Real-time bidding functionality
- **User Authentication**: Secure session-based authentication with role management

### AI-Powered Features
- **Smart Recommendations**: AI-driven product recommendations based on user behavior
- **Price Optimization**: Intelligent pricing suggestions for vendors
- **Market Analytics**: Advanced market trend analysis and insights
- **Customer Segmentation**: AI-powered customer behavior analysis
- **Smart Search**: Enhanced search with AI-powered relevance

### Technical Features
- **SQLite Database**: Lightweight, embedded database with migration system
- **Template Engine**: Server-side rendering with Go templates
- **RESTful API**: Clean API design for frontend integration
- **Comprehensive Testing**: Unit, integration, and end-to-end tests
- **Development Tools**: Database seeding, migration tools, and debugging utilities

## 📋 Prerequisites

- Go 1.23.0 or higher
- SQLite3
- Make (for using the Makefile)

## 🛠️ Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd kolajAi
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Build the application**
   ```bash
   make build
   ```

4. **Set up the database**
   ```bash
   make seed
   ```

5. **Run the application**
   ```bash
   make run
   ```

The application will be available at `http://localhost:8081`

## 🧪 Testing

### Run All Tests
```bash
make all-tests
```

### Run Specific Test Suites
```bash
# Unit tests only
make unit-test

# Integration tests only
make integration-test

# Generate coverage report
make test-coverage
```

### Run Comprehensive Test Suite
```bash
./test_runner.sh
```

## 📁 Project Structure

```
kolajAi/
├── cmd/                    # Application entry points
│   ├── server/            # Main web server
│   ├── seed/              # Database seeding tool
│   └── db-tools/          # Database management tools
├── internal/              # Private application code
│   ├── database/          # Database layer
│   ├── handlers/          # HTTP handlers
│   ├── models/            # Data models
│   ├── services/          # Business logic
│   ├── middleware/        # HTTP middleware
│   ├── config/            # Configuration management
│   └── validation/        # Input validation
├── web/                   # Frontend assets
│   ├── static/            # Static files (CSS, JS, images)
│   └── templates/         # HTML templates
├── Makefile              # Build automation
├── test_runner.sh        # Comprehensive test script
└── integration_test.go   # Integration tests
```

## 🔧 Development

### Available Make Commands

```bash
make help                 # Show all available commands
make build               # Build the main server
make build-tools         # Build command-line tools
make run                 # Build and run the server
make test                # Run all tests
make test-coverage       # Generate test coverage report
make clean               # Clean build artifacts
make fmt                 # Format code
make vet                 # Run go vet
make dev-setup           # Setup development environment
make ci                  # Run full CI pipeline
```

### Database Tools

```bash
# Get database information
make db-info

# Run custom queries (example)
./db-tools query "SELECT COUNT(*) FROM users"
```

### Development Workflow

1. **Setup development environment**
   ```bash
   make dev-setup
   ```

2. **Make changes and test**
   ```bash
   make fmt vet all-tests
   ```

3. **Build and run**
   ```bash
   make run
   ```

## 🏗️ Architecture

### Database Layer
- **Connection Management**: Pooled connections with configurable limits
- **Migration System**: Version-controlled database schema changes
- **Repository Pattern**: Clean separation between data access and business logic
- **Caching Layer**: Optional caching for improved performance

### Service Layer
- **Product Service**: Product management and catalog operations
- **Order Service**: Order processing and fulfillment
- **Vendor Service**: Vendor onboarding and management
- **AI Service**: Machine learning and analytics features
- **Auction Service**: Real-time bidding functionality

### Handler Layer
- **Authentication**: Session-based user authentication
- **Authorization**: Role-based access control
- **Template Rendering**: Server-side HTML generation
- **API Endpoints**: RESTful API for frontend integration

## 🧪 Testing Strategy

### Unit Tests
- **Models**: Data validation and business rules
- **Services**: Business logic and edge cases
- **Database**: Connection and query functionality

### Integration Tests
- **Component Integration**: Service interactions
- **Database Integration**: End-to-end data flow
- **API Integration**: HTTP endpoint functionality

### Test Coverage
Current test coverage: **90.9%** for models, with comprehensive coverage across core components.

## 📊 Monitoring and Logging

### Logging
- **Structured Logging**: Detailed logging with context
- **Debug Logs**: Separate debug log files for troubleshooting
- **Error Tracking**: Comprehensive error logging and handling

### Performance
- **Database Optimization**: Indexed queries and connection pooling
- **Template Caching**: Compiled template caching
- **Static Asset Serving**: Efficient static file serving

## 🚀 Deployment

### Production Build
```bash
make ci                   # Run full CI pipeline
```

### Environment Configuration
The application supports environment-based configuration for:
- Database connections
- Session secrets
- Logging levels
- Performance tuning

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run the test suite: `make all-tests`
5. Submit a pull request

### Code Standards
- Follow Go conventions and best practices
- Maintain test coverage above 80%
- Use meaningful commit messages
- Document public APIs

## 📝 API Documentation

### Authentication Endpoints
- `POST /login` - User authentication
- `POST /register` - User registration
- `POST /logout` - User logout

### Product Endpoints
- `GET /products` - List products
- `GET /product/{id}` - Get product details
- `POST /products` - Create product (vendor only)

### AI Endpoints
- `GET /api/ai/recommendations` - Get AI recommendations
- `POST /api/ai/price-optimize/{id}` - Get price optimization
- `GET /api/ai/market-trends` - Get market analytics

## 🐛 Troubleshooting

### Common Issues

1. **Database Connection Errors**
   ```bash
   # Check database file permissions
   ls -la *.db
   
   # Recreate database
   make clean && make seed
   ```

2. **Template Errors**
   ```bash
   # Check template syntax
   go run cmd/server/main.go --check-templates
   ```

3. **Build Issues**
   ```bash
   # Clean and rebuild
   make clean && make build
   ```

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Built with Go and the Go standard library
- Uses SQLite for embedded database functionality
- Inspired by modern e-commerce platforms and AI-driven analytics

---

For more information, please refer to the documentation in the `/docs` directory or contact the development team.