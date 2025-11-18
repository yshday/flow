# Production Deployment Guide

This guide explains how to deploy the Issue Tracker application in production using Docker and Docker Compose.

## Prerequisites

- Docker Engine 20.10 or higher
- Docker Compose v2.0 or higher
- At least 2GB of available RAM
- 10GB of available disk space

## Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd flow
```

### 2. Configure Environment Variables

Copy the production environment template and configure it:

```bash
cp .env.production .env
```

Edit the `.env` file and update the following REQUIRED values:

```bash
# Database password (REQUIRED - CHANGE THIS)
DB_PASSWORD=your_secure_database_password

# Redis password (REQUIRED - CHANGE THIS)
REDIS_PASSWORD=your_secure_redis_password

# JWT secrets (REQUIRED - GENERATE RANDOM STRINGS)
# Generate with: openssl rand -base64 32
JWT_SECRET=$(openssl rand -base64 32)
JWT_REFRESH_SECRET=$(openssl rand -base64 32)
```

### 3. Create Storage Directory

```bash
mkdir -p storage
```

### 4. Build and Start Services

```bash
# Build the Docker images
docker-compose build

# Start all services in detached mode
docker-compose up -d
```

### 5. Run Database Migrations

The application will automatically run migrations on startup. Check the logs to ensure they completed successfully:

```bash
docker-compose logs app
```

### 6. Verify Deployment

Check that all services are running:

```bash
docker-compose ps
```

You should see three services running:
- `issue-tracker-db` (PostgreSQL)
- `issue-tracker-redis` (Redis)
- `issue-tracker-app` (Application)

Test the health endpoint:

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"ok"}
```

## Service Architecture

The deployment consists of three Docker containers:

### 1. PostgreSQL Database (`postgres`)
- Stores all application data
- Data persisted in Docker volume `postgres_data`
- Default port: 5432

### 2. Redis Cache (`redis`)
- Handles caching and rate limiting
- Data persisted in Docker volume `redis_data`
- Default port: 6379
- Password-protected

### 3. Application Server (`app`)
- Go-based REST API server
- Built from source using multi-stage Docker build
- Runs as non-root user for security
- Default port: 8080

## Environment Variables Reference

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_PASSWORD` | PostgreSQL password | `SecurePassword123!` |
| `REDIS_PASSWORD` | Redis password | `RedisPass456!` |
| `JWT_SECRET` | JWT signing secret | `<generated>` |
| `JWT_REFRESH_SECRET` | JWT refresh token secret | `<generated>` |

### Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_NAME` | `issuetracker` | Database name |
| `DB_USER` | `postgres` | Database username |
| `DB_PORT` | `5432` | PostgreSQL port |
| `REDIS_PORT` | `6379` | Redis port |
| `SERVER_PORT` | `8080` | Application port |
| `JWT_ACCESS_TTL` | `900` | Access token TTL (seconds) |
| `JWT_REFRESH_TTL` | `604800` | Refresh token TTL (seconds) |
| `STORAGE_MAX_FILE_SIZE` | `10485760` | Max file size (10MB) |
| `RATE_LIMIT_ENABLED` | `true` | Enable rate limiting |
| `RATE_LIMIT_PER_MINUTE` | `60` | Requests per minute |
| `SMTP_HOST` | - | SMTP server (optional) |
| `SMTP_PORT` | `587` | SMTP port |
| `SMTP_USERNAME` | - | SMTP username |
| `SMTP_PASSWORD` | - | SMTP password |
| `SMTP_FROM` | `noreply@issuetracker.com` | From email address |

## Common Operations

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f app
docker-compose logs -f postgres
docker-compose logs -f redis
```

### Restart Services

```bash
# Restart all services
docker-compose restart

# Restart specific service
docker-compose restart app
```

### Stop Services

```bash
# Stop all services (keeps containers)
docker-compose stop

# Stop and remove containers
docker-compose down

# Stop and remove containers + volumes (WARNING: deletes data)
docker-compose down -v
```

### Update Application

```bash
# Pull latest changes
git pull

# Rebuild and restart
docker-compose build app
docker-compose up -d app
```

### Database Backup

```bash
# Create backup
docker-compose exec postgres pg_dump -U postgres issuetracker > backup.sql

# Restore from backup
cat backup.sql | docker-compose exec -T postgres psql -U postgres issuetracker
```

### Scale Application

To run multiple application instances behind a load balancer:

```bash
docker-compose up -d --scale app=3
```

Note: You'll need to configure a reverse proxy (nginx, traefik) to load balance between instances.

## Security Considerations

### Production Checklist

- [ ] Changed all default passwords (DB_PASSWORD, REDIS_PASSWORD)
- [ ] Generated strong JWT secrets (32+ characters)
- [ ] Configured HTTPS/TLS (use reverse proxy)
- [ ] Set up firewall rules (only expose necessary ports)
- [ ] Regular backups configured
- [ ] Log monitoring configured
- [ ] Resource limits set in docker-compose.yml
- [ ] Updated CORS_ALLOWED_ORIGINS if using frontend

### Recommended: Reverse Proxy Setup

For production, use a reverse proxy like nginx or Traefik:

```yaml
# Example nginx configuration
server {
    listen 80;
    server_name your-domain.com;

    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Monitoring and Health Checks

### Health Check Endpoints

- Application: `GET http://localhost:8080/health`
- PostgreSQL: Built-in Docker healthcheck
- Redis: Built-in Docker healthcheck

### Monitor Resources

```bash
# View resource usage
docker stats

# View container health
docker-compose ps
```

## Troubleshooting

### Application Won't Start

```bash
# Check logs
docker-compose logs app

# Common issues:
# 1. Database not ready - wait for postgres healthcheck
# 2. Missing environment variables - check .env file
# 3. Port already in use - change SERVER_PORT
```

### Database Connection Issues

```bash
# Test database connection
docker-compose exec postgres psql -U postgres -d issuetracker -c "SELECT 1"

# Check database logs
docker-compose logs postgres
```

### Redis Connection Issues

```bash
# Test Redis connection
docker-compose exec redis redis-cli -a your_redis_password ping

# Should return: PONG
```

### Storage Permission Issues

```bash
# Fix storage permissions
chmod 755 storage
chown -R 1000:1000 storage
```

## Upgrading

### Minor Version Upgrades

```bash
git pull
docker-compose build app
docker-compose up -d app
```

### Major Version Upgrades

1. Backup database
2. Review migration notes
3. Pull latest code
4. Run migrations in staging first
5. Rebuild and restart

```bash
# Backup first
docker-compose exec postgres pg_dump -U postgres issuetracker > backup-$(date +%Y%m%d).sql

# Upgrade
git pull
docker-compose build
docker-compose down
docker-compose up -d
```

## Support

For issues and questions:
- Check logs: `docker-compose logs`
- Review this documentation
- Check GitHub issues
- Contact support team
