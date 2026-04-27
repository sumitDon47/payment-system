# 🚀 Complete Payment System - Frontend & Backend Setup Guide

## Overview

This payment system consists of:
- **Backend**: Docker-based microservices (User, Payment, Notification services)
- **Frontend**: React Native app with Expo (Web/Mobile compatible)
- **Database**: PostgreSQL, Redis cache, Kafka messaging

---

## ✅ Backend Services Status

All services are now running properly! The backend includes:

| Service | Port | Protocol | Purpose |
|---------|------|----------|---------|
| **User Service** | 8082 | HTTP/REST | Authentication, profiles, wallet |
| **Payment Service** | 50051 | gRPC | Payment processing, transactions |
| **Notification Service** | Internal | Kafka | Email notifications |
| **PostgreSQL** | 5432 | SQL | Data storage |
| **Redis** | 6379 | Cache | Balance caching |
| **Kafka** | 9092 | Messaging | Event streaming |
| **Prometheus** | 9091 | Metrics | Monitoring |
| **Grafana** | 3000 | Dashboard | Visualization |

---

## 🛠️ Running the System

### Step 1: Start Backend Services

```bash
cd c:\Users\Raja\Desktop\payment-system

# Start all Docker containers
docker-compose up -d

# Verify services are running
docker-compose ps

# Check user-service logs
docker-compose logs user-service
```

### Step 2: Start Frontend Application

```bash
cd c:\Users\Raja\Desktop\payment-system\payment-app

# Install dependencies (if not already done)
npm install

# Start the Expo development server
npm start

# Choose to run on web (press 'w') or mobile emulator
```

### Step 3: Access the Application

- **Frontend**: `http://localhost:19006` (Expo web)
- **Grafana**: `http://localhost:3000` (user: admin, password: admin)
- **Prometheus**: `http://localhost:9091`

---

## 📱 Frontend Features

### 1. **Login Screen**
- Email and password authentication
- Form validation
- Secure JWT token storage
- Link to sign-up page

### 2. **Sign-Up Screen**
- User registration with name, email, password
- Password confirmation
- Input validation
- Auto-login after successful signup

### 3. **Wallet Dashboard** (New!)
- View current balance
- Send money to other users
- Transaction notes
- Balance refresh
- Logout functionality

**Validation Features:**
- ✅ Email format validation
- ✅ Password strength (minimum 8 characters)
- ✅ Amount validation (must be > 0)
- ✅ Sufficient balance check
- ✅ Form error messages

---

## 🔄 Complete User Flow

### New User Registration Flow
```
1. Launch App → Login Screen
2. Click "Sign up"
3. Enter Name, Email, Password, Confirm Password
4. Backend creates user account with initial balance
5. Auto-login and navigate to Wallet Dashboard
6. User can immediately start sending money
```

### Sending Money Flow
```
1. Wallet Dashboard → Click "Send Money"
2. Enter Receiver Email, Amount, Optional Note
3. Form validates all inputs
4. Confirm Transfer
5. Backend processes payment:
   - Verify sender has funds
   - Deduct from sender
   - Add to receiver
   - Create transaction record
   - Publish Kafka event
   - Send notification email
6. Balance updates instantly (cached via Redis)
7. Success alert shows transaction details
```

---

## 📊 Database Schema

### Users Table
```sql
id              UUID (Primary Key)
name            VARCHAR(100)
email           VARCHAR(150) UNIQUE
password        VARCHAR(255) (bcrypt hashed)
balance         NUMERIC(15, 2) DEFAULT 0.00
created_at      TIMESTAMP
updated_at      TIMESTAMP
```

### Transactions Table
```sql
id              UUID (Primary Key)
sender_id       UUID (FK to users)
receiver_id     UUID (FK to users)
amount          NUMERIC(15, 2)
currency        VARCHAR(10) DEFAULT 'NPR'
status          VARCHAR(20) DEFAULT 'pending'
note            TEXT (optional)
created_at      TIMESTAMP
```

---

## 🔐 Security Features

### Authentication
- JWT tokens with expiration
- Secure password hashing (bcrypt)
- Token stored in secure storage

### API Security
- CORS enabled for frontend origins
- Rate limiting:
  - Auth endpoints: 5 requests/minute
  - API endpoints: 100 requests/minute
- Request validation

### Database
- ACID-compliant transactions
- Foreign key constraints
- Password encryption

---

## 🐛 Troubleshooting

### Issue: Connection Refused Error
**Solution**: Make sure all Docker containers are running
```bash
docker-compose logs user-service
docker-compose restart user-service
```

### Issue: Port Already in Use
**Solution**: Change the port in `.env` file and restart services
```bash
# .env
USER_SERVICE_PORT=8082  # Change this if 8082 is in use
```

### Issue: Frontend Can't Reach Backend
**Solution**: Verify the API base URL in `payment-app/src/api/axios.ts`
```typescript
const BASE_URL = 'http://localhost:8082'  // Must match USER_SERVICE_PORT
```

### Issue: Database Connection Error
**Solution**: Check database credentials in `.env` file
```bash
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=Sumit@123
DB_NAME=payment_db
```

---

## 📚 API Endpoints

### User Service (HTTP/REST)

**POST /register**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

**POST /login**
```json
{
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

**GET /wallet** (Requires JWT)
```
Response: { "balance": 1000.50 }
```

**POST /transfer** (Requires JWT)
```json
{
  "receiver_id": "receiver-uuid",
  "amount": 50.00,
  "currency": "NPR",
  "note": "Payment for services"
}
```

---

## 🔧 Environment Variables

Create `.env` file in the project root:

```bash
# PostgreSQL
POSTGRES_USER=postgres
POSTGRES_PASSWORD=Sumit@123
POSTGRES_DB=payment_db
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=Sumit@123
DB_NAME=payment_db

# Services
USER_SERVICE_PORT=8082
PAYMENT_SERVICE_GRPC_PORT=50051

# Security
JWT_SECRET=JNYBQ+J0sg/hQ1ic3m0pXz7XwrBy0LFatvKqlwnHj3I=

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:19006,http://localhost:8082

# Redis
REDIS_URL=redis:6379

# Kafka
KAFKA_BROKER=kafka:9092
KAFKA_TOPIC=payment.completed

# SendGrid (Email)
SENDGRID_API_KEY=your-key-here
```

---

## 📝 Testing Credentials

You can use these test accounts:
```
Email: john@example.com
Password: Password123!

Email: jane@example.com
Password: Password456!
```

---

## 🚀 Next Steps

1. ✅ Backend running
2. ✅ Frontend wallet dashboard created
3. ✅ Form validation implemented
4. Next: Test the full payment flow
5. Optional: Set up SendGrid for email notifications

---

## 📞 Support

For issues or questions:
1. Check Docker logs: `docker-compose logs [service-name]`
2. Verify environment variables in `.env`
3. Ensure all ports are available
4. Check frontend console for API errors

Happy payments! 💳
