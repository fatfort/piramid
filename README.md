# PIRAMID
**P**redictive, **I**ntegrated, **R**eal-time **A**daptive **M**onitoring **I**DS

A modern Network Intrusion Detection System (NIDS) built with Go, React, and PostgreSQL. PIRAMID provides real-time monitoring, threat detection, and automated response capabilities for network security.

## Features

- 🔍 **Real-time Network Monitoring** - Live event streaming and analysis
- 🌍 **Geographic Threat Visualization** - World map showing attack origins
- 🚫 **Automated IP Blocking** - Dynamic threat response and mitigation
- 📊 **Interactive Dashboard** - Modern web interface for security monitoring
- 🏢 **Multi-tenant Architecture** - Support for multiple organizations
- 📈 **Advanced Analytics** - SSH brute-force detection and statistics
- 🔐 **Secure Authentication** - JWT-based user management

## Architecture

- **Backend**: Go with Gin framework
- **Frontend**: React with TypeScript and Tailwind CSS
- **Database**: PostgreSQL with GORM
- **Message Queue**: NATS with JetStream
- **Network Analysis**: Suricata IDS integration
- **Deployment**: Docker with Docker Compose

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Git

### 1. Clone the Repository

```bash
git clone <repository-url>
cd piramid
```

### 2. Configure Environment

Copy the example environment file and customize it:

```bash
cp .env.example .env
```

Edit `.env` and update these key values:

```env
# Change these for security!
POSTGRES_PASSWORD=your-secure-database-password
JWT_SECRET=your-super-secret-jwt-key
ADMIN_EMAIL=your-email@domain.com
ADMIN_PASSWORD=your-secure-admin-password
ACME_EMAIL=your-email@domain.com
```

### 3. Start the Application

```bash
# Start all services
docker compose -f deploy/docker-compose.yml up -d

# Wait for services to start (about 30 seconds)
docker compose -f deploy/docker-compose.yml logs -f

# Create admin user account
./scripts/create-admin-user.sh
```

### 4. Access the Application

- **Web Interface**: http://localhost:8080
- **API**: http://localhost:8001
- **Grafana** (optional): http://localhost:3000

Login with the credentials you set in `.env`:
- Email: `ADMIN_EMAIL` value
- Password: `ADMIN_PASSWORD` value

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `POSTGRES_PASSWORD` | Database password | `piramid-db-secure-2024` | ✅ |
| `JWT_SECRET` | JWT signing key | `piramid-jwt-secret-key-production-2024` | ✅ |
| `ADMIN_EMAIL` | Initial admin email | `admin@piramid.local` | ✅ |
| `ADMIN_PASSWORD` | Initial admin password | `piramid-admin-2024` | ✅ |
| `ENVIRONMENT` | Application environment | `production` | ❌ |
| `HOME_NET` | Network range to monitor | `any` | ❌ |
| `INTERFACE` | Network interface for Suricata | `eth0` | ❌ |
| `ACME_EMAIL` | Email for SSL certificates | `admin@piramid.local` | ❌ |
| `GRAFANA_PASSWORD` | Grafana admin password | `piramid-grafana-2024` | ❌ |

## Services

The application consists of several services:

- **postgres**: PostgreSQL database
- **nats**: NATS message broker with JetStream
- **api**: PIRAMID REST API server
- **frontend**: Nginx serving React application
- **suricata**: Network IDS (production profile)
- **eve2nats**: Log processor (production profile)
- **grafana**: Analytics dashboard (observability profile)
- **traefik**: Reverse proxy with SSL (production profile)

## Production Deployment

### With Suricata IDS

```bash
# Start with network monitoring
docker compose -f deploy/docker-compose.yml --profile production up -d
```

### With Observability

```bash
# Start with Grafana dashboard
docker compose -f deploy/docker-compose.yml --profile observability up -d
```

### With SSL/HTTPS

```bash
# Start with Traefik reverse proxy
docker compose -f deploy/docker-compose.yml --profile production up -d
```

## Development

### Local Development Setup

```bash
# Start only core services for development
docker compose -f deploy/docker-compose.yml up -d postgres nats

# Set environment for development
export ENVIRONMENT=development
export DATABASE_URL=postgres://piramid:$(cat .env | grep POSTGRES_PASSWORD | cut -d'=' -f2)@localhost:5432/piramid?sslmode=disable

# Run API locally
go run cmd/api/main.go

# Run frontend locally (in another terminal)
cd web
npm install
npm run dev
```

### Building Images

```bash
# Build API image
docker build -f deploy/Dockerfile.api -t piramid-api .

# Build eve2nats image
docker build -f deploy/Dockerfile.eve2nats -t piramid-eve2nats .
```

## API Endpoints

### Authentication
- `POST /auth/login` - User login
- `POST /auth/logout` - User logout

### Events
- `GET /api/events` - Get paginated events
- `GET /api/events/stream` - Real-time event stream (SSE)

### Statistics
- `GET /api/stats/overview` - General statistics
- `GET /api/stats/ssh` - SSH brute-force statistics

### IP Management
- `POST /api/ban` - Ban IP address
- `GET /api/bans` - List banned IPs
- `DELETE /api/bans/{id}` - Unban IP address

## Monitoring

### Health Checks

```bash
# Check API health
curl http://localhost:8080/health

# Check all services
docker compose -f deploy/docker-compose.yml ps
```

### Logs

```bash
# View all logs
docker compose -f deploy/docker-compose.yml logs -f

# View specific service logs
docker compose -f deploy/docker-compose.yml logs -f api
docker compose -f deploy/docker-compose.yml logs -f postgres
```

## Security Considerations

1. **Change Default Passwords**: Always modify default passwords in `.env`
2. **JWT Secret**: Use a strong, unique JWT secret key
3. **Network Security**: Configure `HOME_NET` to match your network range
4. **SSL Certificates**: Use valid SSL certificates in production
5. **Firewall Rules**: Restrict access to necessary ports only

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   ```bash
   # Check if postgres is running
   docker compose -f deploy/docker-compose.yml ps postgres
   
   # Check postgres logs
   docker compose -f deploy/docker-compose.yml logs postgres
   ```

2. **API Container Restarting**
   ```bash
   # Check API logs for errors
   docker compose -f deploy/docker-compose.yml logs api
   ```

3. **Frontend Not Loading**
   ```bash
   # Check if nginx is running
   docker compose -f deploy/docker-compose.yml ps frontend
   
   # Test API connectivity
   curl http://localhost:8080/health
   ```

### Reset Everything

```bash
# Stop all services and remove volumes
docker compose -f deploy/docker-compose.yml down -v

# Remove all images
docker rmi $(docker images "piramid*" -q)

# Start fresh
docker compose -f deploy/docker-compose.yml up -d
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support and questions:
- Create an issue on GitHub
- Check the troubleshooting section above
- Review the logs for error messages 