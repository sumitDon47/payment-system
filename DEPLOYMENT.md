# Deployment Guide

## Pre-Deployment Checklist

Before deploying to production, ensure:

- ✅ All GitHub Actions workflows pass
- ✅ Code coverage > 60%
- ✅ No security vulnerabilities
- ✅ PR reviewed and approved
- ✅ Merged to `main` branch
- ✅ Docker images built and pushed
- ✅ Environment variables configured
- ✅ Database backups taken
- ✅ Rollback plan documented

---

## Automated Deployment (GitHub Actions)

### Option 1: Automatic on Main Branch Push

```
Push to main → GitHub Actions runs → All tests pass → Release created
```

The workflow automatically:
1. Runs tests
2. Builds Docker images
3. Pushes to container registry
4. Creates GitHub release
5. Notifies deployment team

**No manual steps required** — just push to main!

### Option 2: Manual Deployment

**Trigger manual deployment from GitHub Actions:**

1. Go to **GitHub → Actions tab**
2. Select **"Deploy to Docker Hub"** workflow
3. Click **"Run workflow"** dropdown
4. Choose environment:
   - `staging`: Test environment
   - `production`: Production environment
5. Click **"Run workflow"** green button
6. Watch the build in real-time

**Command line equivalent**:

```bash
# Trigger via GitHub CLI
gh workflow run deploy.yml -f environment=production
```

---

## Manual Local Deployment

For full control, deploy manually from your machine:

### Step 1: Pull Latest Code

```bash
cd /path/to/payment-system
git pull origin main
```

### Step 2: Build Docker Images Locally

```bash
docker-compose build
```

Or build specific services:

```bash
docker build -t payment-system-user-service:latest ./user-service
docker build -t payment-system-payment-service:latest ./payment-service
docker build -t payment-system-notification-service:latest ./notification-service
```

### Step 3: Stop Running Containers

```bash
docker-compose down
```

### Step 4: Start New Deployment

```bash
docker-compose up -d
```

### Step 5: Verify Deployment

```bash
# Wait 10 seconds for services to start
sleep 10

# Check User Service
curl http://localhost:8080/health
# Expected: {"status":"ok","service":"user-service","redis":"ok"}

# Check Payment Service (gRPC)
grpcurl -plaintext localhost:9090 list
# Expected: grpc.reflection.v1.ServerReflection, payment.PaymentService

# Check logs
docker-compose logs user-service
docker-compose logs payment-service
docker-compose logs notification-service
```

---

## Using Registry Images

### Pull Images from GitHub Container Registry

Instead of building locally, pull pre-built images:

```bash
# Authenticate (if private)
echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin

# Pull images
docker pull ghcr.io/<username>/payment-system-user-service:latest
docker pull ghcr.io/<username>/payment-system-payment-service:latest
docker pull ghcr.io/<username>/payment-system-notification-service:latest
```

### Update docker-compose.yml

```yaml
version: '3.8'

services:
  user-service:
    image: ghcr.io/<username>/payment-system-user-service:latest
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - DATABASE_URL=postgresql://user:password@postgres:5432/payment_system
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=your-secret-key
    depends_on:
      postgres:
        condition: service_healthy

  payment-service:
    image: ghcr.io/<username>/payment-system-payment-service:latest
    ports:
      - "9090:9090"
    environment:
      - GRPC_PORT=9090
      - DATABASE_URL=postgresql://user:password@postgres:5432/payment_system
      - KAFKA_BROKER=kafka:29092
      - KAFKA_TOPIC=payment-events
    depends_on:
      - postgres
      - kafka

  notification-service:
    image: ghcr.io/<username>/payment-system-notification-service:latest
    environment:
      - KAFKA_BROKER=kafka:29092
      - KAFKA_TOPIC=payment-events
      - SENDGRID_API_KEY=${SENDGRID_API_KEY}
    depends_on:
      - kafka

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=payment_system
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  kafka:
    image: confluentinc/cp-kafka:7.4.0
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:29092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
    depends_on:
      - zookeeper

  zookeeper:
    image: confluentinc/cp-zookeeper:7.4.0
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181

volumes:
  postgres_data:
```

Then deploy:

```bash
docker-compose up -d
```

---

## Zero-Downtime Deployment (Rolling Update)

For production with minimal downtime:

### 1. Deploy New Version in Parallel

```bash
# Terminal 1: Start new version
docker run -d --name payment-system-user-service-new \
  -p 8081:8080 \
  ghcr.io/<username>/payment-system-user-service:latest

# Wait for it to be healthy
sleep 5
curl http://localhost:8081/health
```

### 2. Switch Traffic (Using Reverse Proxy)

Use nginx or caddy to route traffic:

**nginx.conf**:
```nginx
upstream user_service {
    server user-service-old:8080;
    server user-service-new:8080;
}

server {
    listen 8080;
    location / {
        proxy_pass http://user_service;
    }
}
```

Or use docker-compose overlay networks:

```bash
# Create overlay network
docker network create payment-network

# Attach containers to same network
docker network connect payment-network user-service-old
docker network connect payment-network user-service-new

# Gradual traffic switch via load balancer
```

### 3. Stop Old Version

Once new version is stable:

```bash
docker stop payment-system-user-service-old
docker rm payment-system-user-service-old
```

---

## Health Checks & Monitoring

### Post-Deployment Verification

```bash
#!/bin/bash

echo "🔍 Checking service health..."

# User Service
USER_HEALTH=$(curl -s http://localhost:8080/health)
if [[ $USER_HEALTH == *"ok"* ]]; then
    echo "✅ User Service: Healthy"
else
    echo "❌ User Service: Unhealthy"
    exit 1
fi

# Payment Service (gRPC)
PAYMENT_HEALTH=$(grpcurl -plaintext localhost:9090 list 2>&1 | grep PaymentService)
if [[ -n $PAYMENT_HEALTH ]]; then
    echo "✅ Payment Service: Healthy"
else
    echo "❌ Payment Service: Unhealthy"
    exit 1
fi

# Database
DB_CHECK=$(docker exec payment-system-postgres-1 pg_isready -U user)
if [[ $DB_CHECK == *"accepting"* ]]; then
    echo "✅ PostgreSQL: Healthy"
else
    echo "❌ PostgreSQL: Unhealthy"
    exit 1
fi

echo ""
echo "✅ All services healthy - deployment successful!"
```

### Monitoring Tools

Add to docker-compose for production:

```yaml
prometheus:
  image: prom/prometheus:latest
  volumes:
    - ./prometheus.yml:/etc/prometheus/prometheus.yml
  ports:
    - "9090:9090"

grafana:
  image: grafana/grafana:latest
  ports:
    - "3000:3000"
  depends_on:
    - prometheus
```

---

## Rollback Procedure

If deployment fails:

### Option 1: Rollback via Docker Compose

```bash
# Stop current deployment
docker-compose down

# Update docker-compose.yml to use previous image tag
# (change `:latest` to `:previous-sha`)

# Restart with previous version
docker-compose up -d
```

### Option 2: Rollback via GitHub

```bash
# Revert commit
git revert <bad-commit-sha>
git push origin main

# GitHub Actions automatically:
# - Builds previous version
# - Runs tests
# - Pushes new images
# - Creates release
```

### Option 3: Keep Old Container

Don't delete old container immediately:

```bash
# Keep previous version running
docker rename payment-system-user-service user-service-backup

# Start new version
docker run -d --name payment-system-user-service ...

# If new version fails, switch back
docker stop payment-system-user-service
docker rename user-service-backup payment-system-user-service
docker start payment-system-user-service
```

---

## Production Checklist

### Before Going Live

- [ ] All GitHub Actions workflows passing
- [ ] Code coverage ≥ 60%
- [ ] Security scan clean
- [ ] Database backups taken
- [ ] Environment variables set securely
- [ ] HTTPS/TLS certificates ready
- [ ] Monitoring tools configured
- [ ] Alerting configured
- [ ] Runbooks written for common issues
- [ ] Team notified of deployment

### After Deployment

- [ ] Verify all services healthy
- [ ] Check application logs
- [ ] Monitor error rates (should be 0%)
- [ ] Monitor response times
- [ ] Check database performance
- [ ] Verify email notifications sending
- [ ] Test payment flow end-to-end
- [ ] Monitor CPU/memory usage
- [ ] Check disk space

### Ongoing Monitoring

- [ ] Health check every 5 minutes
- [ ] Metric collection to Prometheus
- [ ] Anomaly detection enabled
- [ ] PagerDuty/Slack alerts active
- [ ] Weekly performance review
- [ ] Monthly security audit

---

## Troubleshooting Deployments

### Container Won't Start

```bash
docker-compose logs user-service
# Check for errors

# Common issues:
# - Port already in use: lsof -i :8080
# - Environment vars missing: docker-compose config
# - Image not found: docker images
```

### Database Connection Failed

```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check connection string
echo $DATABASE_URL

# Test connection
docker exec payment-system-postgres-1 psql -U user -d payment_system -c "SELECT 1"
```

### gRPC Service Not Responding

```bash
# Verify service is running
docker-compose ps payment-service

# Check logs
docker-compose logs payment-service

# Test connection
grpcurl -plaintext localhost:9090 list

# If port issue:
lsof -i :9090
netstat -tuln | grep 9090
```

### Email Notifications Not Sending

```bash
# Check SendGrid API key is set
echo $SENDGRID_API_KEY

# Check Kafka is running
docker-compose ps kafka

# Test Kafka connection
docker exec payment-system-kafka-1 kafka-broker-api-versions --bootstrap-server localhost:29092
```

---

## Performance Tuning for Production

### PostgreSQL

```sql
-- Increase connection pool
ALTER SYSTEM SET max_connections = 200;

-- Tune shared_buffers (1/4 of RAM)
ALTER SYSTEM SET shared_buffers = '4GB';

-- Increase work_mem for complex queries
ALTER SYSTEM SET work_mem = '256MB';

-- Enable query optimization
ALTER SYSTEM SET enable_partitionwise_join = on;
```

### Redis

```bash
# Increase max memory
redis-cli CONFIG SET maxmemory 2gb
redis-cli CONFIG SET maxmemory-policy allkeys-lru

# Enable persistence
redis-cli CONFIG SET save "900 1 300 10 60 10000"
```

### Kafka

```bash
# Increase partition replicas for durability
kafka-topics --alter --topic payment-events --replication-factor 3

# Increase retention
kafka-configs --entity-type topics --entity-name payment-events \
  --alter --add-config retention.ms=604800000  # 7 days
```

---

## Disaster Recovery

### Backup Strategy

```bash
# Daily PostgreSQL backups
docker exec payment-system-postgres-1 \
  pg_dump -U user payment_system > backup_$(date +%Y%m%d).sql

# Store off-site (S3, GCS, etc.)
aws s3 cp backup_$(date +%Y%m%d).sql s3://backup-bucket/

# Verify backup
pg_restore -l backup_$(date +%Y%m%d).sql
```

### Recovery Procedure

```bash
# 1. Stop services
docker-compose down

# 2. Restore database
docker run --rm -v $(pwd):/backup postgres:15-alpine \
  psql -h postgres -U user payment_system < /backup/backup.sql

# 3. Start services
docker-compose up -d

# 4. Verify
curl http://localhost:8080/health
```

---

## Deployment Frequency

### Recommended Schedule

- **Development**: Multiple times daily
- **Staging**: Weekly
- **Production**: Weekly (Tuesday morning for safety)

### Change Window

- **Preferred**: 10am - 12pm (business hours, team available)
- **Avoid**: Friday evening, holidays, maintenance windows
- **Duration**: 15-30 minutes

---

## Team Communication

### Deployment Announcement

```
🚀 DEPLOYING TO PRODUCTION
   Time: 10:00 AM EST
   Changes: Add rate limiting to API
   Risk Level: Low (fully tested)
   Rollback: Available (1 minute)
   Contact: @devops-team in #deployments
```

### Success Notification

```
✅ DEPLOYMENT COMPLETE
   Services: All healthy
   Tests: All passed
   Metrics: Normal
   Duration: 5 minutes
   No incidents reported
```

---

## References

- [GitHub Actions Docs](https://docs.github.com/en/actions)
- [Docker Deployment Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Release Engineering](https://landing.google.com/sre/books/)
