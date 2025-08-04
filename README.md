# KolajAI Enterprise Marketplace

A modern, AI-powered e-commerce marketplace platform built with Go and MySQL.

## Features

- **MySQL Database**: Production-ready MySQL database integration
- **Redis Caching**: High-performance caching layer
- **Modern Frontend**: Built with TailwindCSS and Alpine.js
- **AI Integration**: Support for various AI services (OpenAI, Anthropic, etc.)
- **Marketplace Integration**: Connect with major Turkish marketplaces
- **Secure Authentication**: JWT-based authentication with 2FA support
- **RESTful API**: Comprehensive REST API
- **Docker Support**: Full Docker containerization

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.23+ (for development)
- Node.js 18+ (for frontend development)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd kolajAi
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start with Docker**
   ```bash
   docker-compose up -d
   ```

The application will be available at `http://localhost:8081`

### Development Setup

1. **Install Go dependencies**
   ```bash
   go mod download
   ```

2. **Install Node.js dependencies**
   ```bash
   npm install
   ```

3. **Build frontend assets**
   ```bash
   npm run build
   ```

4. **Run the application**
   ```bash
   go run cmd/server/main.go
   ```

## Configuration

The application uses environment variables for configuration. Key variables include:

- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`: MySQL connection
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`: Redis connection
- `JWT_SECRET`: JWT signing secret
- `ENCRYPTION_KEY`: Data encryption key
- `SMTP_*`: Email configuration

See `.env.example` for all available options.

## Database

The application uses MySQL with automatic migrations. The database schema is created automatically on first run.

## API Documentation

The API follows RESTful conventions. Key endpoints:

- `/api/auth/*` - Authentication
- `/api/users/*` - User management
- `/api/products/*` - Product management
- `/api/orders/*` - Order management
- `/health` - Health check

## Docker Services

- **app**: Main application (port 8081)
- **mysql**: MySQL database (port 3306)
- **redis**: Redis cache (port 6379)

## License

MIT License - see LICENSE file for details.

## Support

For support and questions, please contact the development team.