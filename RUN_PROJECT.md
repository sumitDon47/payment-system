# Running the Payment System Project

**Complete Guide** - April 29, 2026

---

## 📋 System Requirements

### Hardware
- RAM: 8GB minimum (16GB recommended)
- Disk Space: 10GB
- CPU: Dual-core or better

### Software Required
- **Docker & Docker Compose** (https://www.docker.com/products/docker-desktop)
- **Go 1.25** (https://golang.org/dl)
- **Node.js 18+** (https://nodejs.org)
- **Git** (https://git-scm.com)

### Installation Check
```bash
# Verify installations
docker --version        # Docker 20.10+
docker-compose --version # Docker Compose 2.0+
go version             # go 1.25
node --version         # v18 or higher
npm --version          # npm 9 or higher
git --version          # git 2.x
```

---

## 🚀 QUICK START (5 minutes)

### Option 1: Run Everything with Docker (Easiest)

```bash
cd /path/to/payment-system

# Start all services
docker compose up -d --build

# Wait 60 seconds for services to initialize
# Then test:
curl http://localhost:8080/health        # User service
curl http://localhost:9090/health        # Payment service (gRPC)

# View logs
docker compose logs -f user-service
docker compose logs -f payment-service
docker compose logs -f notification-service

# Stop everything
docker compose down
```

**Success Indicators:**
- ✅ All containers running (`docker compose ps`)
- ✅ User service health: 200 OK
- ✅ No error logs

---

### Option 2: Run Locally (Development)

```bash
# Terminal 1: Start Infrastructure (Postgres, Redis, Kafka)
docker compose up postgres redis kafka zookeeper

# Terminal 2: Start User Service
cd user-service
go mod download
go run main.go

# Terminal 3: Start Payment Service
cd payment-service
go mod download
go run main.go

# Terminal 4: Start Notification Service
cd notification-service
go mod download
go run main.go

# Terminal 5: Start Frontend
cd payment-app
npm install
npm start
```

---

## 📚 DETAILED SETUP

### Step 1: Clone/Navigate to Project

```bash
# If cloning fresh
git clone https://github.com/sumitDon47/payment-system.git
cd payment-system

# If you already have it
cd c:\Users\Raja\Desktop\payment-system
```

### Step 2: Verify Docker is Running

```bash
# Start Docker Desktop (if on Windows/Mac)
# Or ensure Docker daemon is running on Linux

# Check Docker is accessible
docker ps
# Should show running containers or empty list
```

### Step 3: Choose Your Setup Method

---

## ⚙️ METHOD 1: DOCKER COMPOSE (Recommended for Testing)

**Best for:** Testing all services together, CI/CD, production-like environment

### Setup

```bash
cd payment-system

# Start all services in background
docker compose up -d --build

# Watch startup progress
docker compose logs -f

# Wait until you see:
# user-service       | 🚀 User Service running on port 8080
# payment-service    | 🚀 Payment Service running on port 9090
# notification-service | 🚀 Notification Service running
```

### Test Services

```bash
# Test User Service
curl -X GET http://localhost:8080/health

# Expected response:
# {"status":"ok","message":"User service is healthy"}

# Test via HTTP client
# Import: body.json into Postman or VS Code REST Client
# POST http://localhost:8080/register
# Body:
# {
#   "name": "Test User",
#   "email": "test@example.com",
#   "password": "SecurePass123!"
# }
```

### View Real-Time Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f user-service
docker compose logs -f payment-service
docker compose logs -f notification-service

# Specific service with 50 last lines
docker compose logs -f --tail 50 user-service
```

### Stop Services

```bash
# Stop but keep containers
docker compose stop

# Stop and remove containers
docker compose down

# Stop and remove everything including volumes
docker compose down -v
```

---

## ⚙️ METHOD 2: DOCKER COMPOSE + LOCAL DEVELOPMENT

**Best for:** Development, debugging, hot reload

### Step 1: Start Infrastructure Only

```bash
cd payment-system

# Start supporting services
docker compose up -d postgres redis kafka zookeeper

# Verify they're healthy
docker compose ps
# All should show "healthy" or "up"

# Wait 30 seconds for databases to initialize
```

### Step 2: Start Go Services Locally

#### Terminal 2: User Service

```bash
cd payment-system/user-service

# Download dependencies
go mod download

# Run the service
go run main.go

# Expected output:
# 🚀 User Service running on port 8080
# Database connected
# ✅ Redis cache enabled
```

#### Terminal 3: Payment Service

```bash
cd payment-system/payment-service

go mod download
go run main.go

# Expected output:
# 🚀 Payment Service running on port 9090
# Database connected
# Listening on gRPC port 9090
```

#### Terminal 4: Notification Service

```bash
cd payment-system/notification-service

go mod download
go run main.go

# Expected output:
# 🚀 Notification Service running
# Connected to Kafka
# Listening for messages
```

### Step 3: Start Frontend

#### Terminal 5: React Native App

```bash
cd payment-system/payment-app

# Install dependencies (first time only)
npm install
# This restores node_modules that were deleted

# Start the development server
npm start

# Choose platform:
# Press 'a' for Android
# Press 'i' for iOS
# Press 'w' for web

# Opens in emulator/simulator/browser
```

---

## ⚙️ METHOD 3: FULL NATIVE SETUP (Advanced)

**Best for:** Pure development, understanding architecture

### Prerequisites

```bash
# Install Go
# https://golang.org/dl

# Install PostgreSQL 15
# https://www.postgresql.org/download

# Install Redis
# https://redis.io/download

# Install Kafka
# https://kafka.apache.org/downloads

# Install Node.js
# https://nodejs.org
```

### Configuration

Create `.env` file in project root:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=payment_db

# Services
USER_SERVICE_PORT=8080
PAYMENT_SERVICE_GRPC_PORT=9090

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Kafka
KAFKA_BROKER=localhost:9092
KAFKA_CONSUMER_GROUP=notification-service

# JWT
JWT_SECRET=your-secret-key-here
```

### Start Services

```bash
# Terminal 1: PostgreSQL
# Linux/Mac
postgres -D /usr/local/var/postgres

# Windows
# Start from Services or PostgreSQL installer

# Terminal 2: Redis
redis-server

# Terminal 3: Kafka + Zookeeper
# Extract Kafka, then:
bin/zookeeper-server-start.sh config/zookeeper.properties
bin/kafka-server-start.sh config/server.properties

# Terminal 4: User Service
cd payment-system/user-service
go run main.go

# Terminal 5: Payment Service
cd payment-system/payment-service
go run main.go

# Terminal 6: Notification Service
cd payment-system/notification-service
go run main.go

# Terminal 7: Frontend
cd payment-system/payment-app
npm install
npm start
```

---

## 🧪 TEST THE SYSTEM

### Manual Testing

#### 1. Register a User

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'

# Expected response:
# {
#   "token": "eyJhbGciOiJIUzI1NiIs...",
#   "user": {
#     "id": "550e8400-e29b-41d4-a716-446655440000",
#     "name": "John Doe",
#     "email": "john@example.com"
#   }
# }
```

#### 2. Login

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'

# Save the token from response
```

#### 3. Get Wallet Balance

```bash
curl -X GET http://localhost:8080/wallet \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Expected response:
# {
#   "data": {
#     "NPR": 0,
#     "USD": 0
#   }
# }
```

#### 4. Test via Postman/VS Code REST Client

Create file `test.http`:

```http
### Register User
POST http://localhost:8080/register
Content-Type: application/json

{
  "name": "Test User",
  "email": "test@example.com",
  "password": "SecurePass123!"
}

### Login
POST http://localhost:8080/login
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "SecurePass123!"
}

### Get Wallet (replace with actual token)
GET http://localhost:8080/wallet
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

---

## 🐛 TROUBLESHOOTING

### Issue: Containers won't start

```bash
# Check Docker is running
docker ps

# Check logs
docker compose logs

# Rebuild without cache
docker compose up -d --build --no-cache

# Full reset
docker compose down -v
docker system prune -a
docker compose up -d --build
```

### Issue: Port already in use

```bash
# Find process using port 8080
# Windows
netstat -ano | findstr :8080

# Mac/Linux
lsof -i :8080

# Kill process
# Windows
taskkill /PID <PID> /F

# Mac/Linux
kill -9 <PID>
```

### Issue: Database connection failed

```bash
# Verify PostgreSQL is running
docker compose ps postgres
# Should show "healthy"

# Check logs
docker compose logs postgres

# Reset database
docker compose down -v
docker compose up postgres redis kafka zookeeper
```

### Issue: npm start fails

```bash
# Rebuild node_modules
cd payment-app
rm -rf node_modules package-lock.json
npm install

# Clear cache
npm cache clean --force

# Try again
npm start
```

### Issue: Go services won't compile

```bash
# Update dependencies
go mod tidy
go mod download

# Clear build cache
go clean -cache

# Try again
go run main.go
```

### Issue: Kafka not working

```bash
# Check Zookeeper is running first
docker compose ps zookeeper
# Must be running before Kafka

# Restart both
docker compose down
docker compose up -d zookeeper kafka
docker compose logs -f kafka
```

---

## 📊 VERIFY EVERYTHING IS RUNNING

```bash
# Check all containers
docker compose ps

# Expected output:
# NAME                    STATUS
# postgres               healthy
# redis                  healthy  
# zookeeper              healthy
# kafka                  healthy
# user-service           up
# payment-service        up
# notification-service   up

# Check all services respond
curl http://localhost:8080/health  # User service
curl http://localhost:8080/metrics # Prometheus metrics

# Check logs for errors
docker compose logs | grep -i error
```

---

## 🔄 COMMON WORKFLOWS

### Development Loop

```bash
# 1. Start all services once
docker compose up -d

# 2. Make code changes
# Edit files in user-service/handler/user.go, etc.

# 3. Rebuild services
docker compose restart user-service

# 4. Test changes
curl http://localhost:8080/health

# 5. View logs
docker compose logs -f user-service
```

### Testing Payments

```bash
# 1. Register 2 users
curl -X POST http://localhost:8080/register \
  -d '{"name":"Sender","email":"sender@test.com","password":"Pass123!"}'

curl -X POST http://localhost:8080/register \
  -d '{"name":"Receiver","email":"receiver@test.com","password":"Pass123!"}'

# 2. Set balances (via database or API)
# Use payment-service gRPC client

# 3. Send payment
# Use payment-service gRPC call

# 4. Verify in notification-service logs
docker compose logs -f notification-service
```

### Database Management

```bash
# Access PostgreSQL
docker compose exec postgres psql -U postgres -d payment_db

# Common queries
SELECT * FROM users;
SELECT * FROM transactions;

# Exit
\q

# Backup database
docker compose exec postgres pg_dump -U postgres payment_db > backup.sql

# Restore database
docker compose exec -T postgres psql -U postgres payment_db < backup.sql
```

---

## 🚀 PRODUCTION DEPLOYMENT

### Build Docker Images

```bash
# Build all images
docker compose build

# Push to registry (e.g., Docker Hub)
docker tag payment-system/user-service:latest your-registry/user-service:latest
docker push your-registry/user-service:latest

# Similar for payment-service and notification-service
```

### Deploy to Cloud

```bash
# DigitalOcean App Platform
# 1. Connect GitHub repo
# 2. Select docker-compose.yml
# 3. Configure environment variables
# 4. Deploy

# AWS ECS
# 1. Create task definitions from Dockerfiles
# 2. Create service
# 3. Configure load balancer
# 4. Deploy

# Kubernetes
# Use provided docker-compose.yml
# Or write Helm charts
```

---

## 📝 QUICK REFERENCE

| Task | Command |
|------|---------|
| Start all | `docker compose up -d --build` |
| Stop all | `docker compose down` |
| View logs | `docker compose logs -f` |
| Restart service | `docker compose restart user-service` |
| Run tests | `go test ./... -v` |
| Install deps | `npm install` or `go mod download` |
| Build frontend | `npm run build` |
| Build backend | `go build .` |
| Access DB | `docker compose exec postgres psql -U postgres` |
| Reset DB | `docker compose down -v && docker compose up` |

---

## ✅ SUCCESS CHECKLIST

- [ ] Docker is running
- [ ] All containers started (`docker compose ps`)
- [ ] User service responds to health check
- [ ] Database is connected
- [ ] Redis is connected
- [ ] Kafka is connected
- [ ] Can register user via API
- [ ] Can login via API
- [ ] Frontend app starts
- [ ] No errors in logs

---

## 🆘 NEED HELP?

**Check logs first:**
```bash
docker compose logs service-name
```

**Common issues:**
1. **Port in use:** Change port in docker-compose.yml or kill process
2. **Database error:** `docker compose down -v && docker compose up -d`
3. **Memory issue:** Increase Docker memory allocation (Preferences/Settings)
4. **Permission denied:** Run with `sudo` or add user to docker group

---

**You're ready to run the project!** 🎉
