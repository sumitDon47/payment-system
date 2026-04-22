# 🔧 Troubleshooting & Setup Checklist

## ❌ Why Sign Up Wasn't Working (Root Cause Analysis)

### The Problem
You were unable to sign up because:

1. **Missing SignUp Screen** - Only LoginScreen existed
2. **No Sign Up Navigation** - No way to navigate to registration
3. **No API Service** - Raw axios calls scattered in components
4. **No Form Validation** - Input validation was missing

### The Solution
✅ Created complete signup integration:
- SignUpScreen with full validation
- Centralized API service (services.ts)
- Updated LoginScreen with proper error handling
- Comprehensive documentation

---

## ✅ Complete Setup Checklist

### Step 1: Verify Backend is Running

```bash
# In a terminal, check Docker containers
docker ps
```

You should see:
- postgres:15-alpine ✅
- redis:7-alpine ✅
- confluentinc/cp-kafka:7.4.0 ✅
- user-service (custom) ✅
- payment-service (custom) ✅
- notification-service (custom) ✅

**If not running:**
```bash
cd c:\Users\Raja\Desktop\payment-system
docker-compose up
```

Wait for all services to be healthy (check logs for "service healthy")

### Step 2: Test Backend API

```bash
# Check if user service is responding
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "service": "user-service",
  "redis": "ok"
}
```

### Step 3: Frontend Dependencies

```bash
cd c:\Users\Raja\Desktop\payment-system\payment-app
npm install
```

### Step 4: Start Frontend

```bash
npm run web
```

Visit http://localhost:3000

### Step 5: Test the Full Flow

#### Sign Up
1. Click "Sign up" link on login page
2. Fill in:
   - Name: John Doe
   - Email: test@example.com
   - Password: TestPass123!
   - Confirm: TestPass123!
3. Click "Create Account"
4. You should see success message
5. Token should be saved (check browser Storage)

#### Login
1. Use same credentials:
   - Email: test@example.com
   - Password: TestPass123!
2. Click "Log In"
3. You should see success message

#### Verify Token Storage
Open browser DevTools → Application → Session Storage or LocalStorage
Look for `jwt_token` key

---

## 🔴 Common Issues & Solutions

### Issue 1: "Cannot connect to 10.0.2.2:8080" (Android)

**Cause:** BASE_URL is incorrect for your setup

**Solution:** Update [src/api/axios.ts](src/api/axios.ts)

```typescript
// Change from:
const BASE_URL = Platform.OS === 'android' 
  ? 'http://10.0.2.2:8080'
  : 'http://localhost:8080';

// To (for web/local testing):
const BASE_URL = 'http://localhost:8080';
```

### Issue 2: "Failed to sign up" with no specific error

**Cause:** Backend not running or not responding

**Solution:**
1. Check docker: `docker ps`
2. Check logs: `docker logs user-service`
3. Verify port: `curl http://localhost:8080/health`
4. Restart: `docker-compose restart user-service`

### Issue 3: "Email already exists" error

**Cause:** This email was already registered

**Solution:** 
- Use a different email, OR
- Reset database:
  ```bash
  docker-compose down -v  # Remove volumes
  docker-compose up       # Recreate fresh
  ```

### Issue 4: "401 Unauthorized" after signup

**Cause:** Token not being sent in subsequent requests

**Solution:** Token is automatically handled by axios interceptor. Check:
1. Token was saved: `await SecureStore.setItemAsync('jwt_token', response.token)`
2. SecureStore is available in simulator/emulator

### Issue 5: CORS errors

**Cause:** API not configured for your domain

**Solution:** CORS is enabled in backend for:
- http://localhost:3000
- http://localhost:5173

If using different port, update in [user-service/middleware/cors.go](../../user-service/middleware/cors.go):

```go
CORS_ALLOWED_ORIGINS: "http://localhost:3000,http://localhost:5173,http://localhost:YOUR_PORT"
```

### Issue 6: "Rate limit exceeded" error

**Cause:** Too many requests in short time

**Solution:**
- Wait 1 minute before retrying
- Reduce request frequency in code

Rate limits:
- Sign up/login: 5 requests/minute
- Profile/wallet: 100 requests/minute

---

## 🧪 Manual API Testing (With cURL)

### 1. Test API Health

```bash
curl http://localhost:8080/health
```

### 2. Register a User

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "TestPass123!"
  }'
```

Response:
```json
{
  "token": "eyJhbGc...",
  "user": {
    "id": "uuid",
    "name": "Test User",
    "email": "test@example.com",
    "balance": 0,
    "created_at": "2025-04-22T...",
    "updated_at": "2025-04-22T..."
  }
}
```

### 3. Login

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123!"
  }'
```

### 4. Get Profile (Authenticated)

```bash
# Replace TOKEN with actual token from login response
curl http://localhost:8080/profile \
  -H "Authorization: Bearer TOKEN"
```

### 5. Get Wallet Balance

```bash
curl http://localhost:8080/wallet \
  -H "Authorization: Bearer TOKEN"
```

---

## 📊 Database Setup Verification

### Check if PostgreSQL has data

```bash
# Connect to postgres
docker exec -it payment-system-postgres-1 psql -U postgres -d payment_db

# List tables
\dt

# Check users table
SELECT * FROM users;

# Exit
\q
```

### Check Redis

```bash
docker exec -it payment-system-redis-1 redis-cli

# Ping redis
PING

# List keys
KEYS *

# Exit
EXIT
```

---

## 🚀 Frontend API Service Methods

### How to Use

```typescript
import { userAPI, paymentAPI } from '../api/services';

// Example in a component
useEffect(() => {
  const fetchUserProfile = async () => {
    try {
      const profile = await userAPI.getProfile();
      setUser(profile);
    } catch (error) {
      if (error.response?.status === 401) {
        // Token expired
        navigation.replace('Login');
      }
    }
  };
  
  fetchUserProfile();
}, []);
```

### Available Methods

**User API:**
```typescript
userAPI.register(name, email, password)
userAPI.login(email, password)
userAPI.getProfile()
userAPI.getWalletBalance()
userAPI.healthCheck()
```

**Payment API (requires backend proxy):**
```typescript
paymentAPI.sendPayment(receiver_id, amount, currency, note)
paymentAPI.getTransaction(transaction_id)
paymentAPI.getBalance(user_id)
```

---

## 📝 File Structure

```
payment-app/
├── src/
│   ├── api/
│   │   ├── axios.ts          ✅ Axios config with BASE_URL
│   │   └── services.ts       ✅ NEW - Centralized API service
│   ├── screens/
│   │   ├── LoginScreen.tsx   ✅ UPDATED - Uses services.ts
│   │   └── SignUpScreen.tsx  ✅ NEW - Complete signup UI
│   └── ...
├── API_INTEGRATION.md        ✅ NEW - Full API documentation
├── QUICK_START.md           ✅ NEW - Quick setup guide
└── ...
```

---

## ✨ What You Can Do Now

✅ **Sign Up** - Create new accounts
✅ **Login** - Authenticate existing users
✅ **Get Profile** - Fetch user information
✅ **Get Balance** - Check wallet balance
✅ **Form Validation** - Proper input validation
✅ **Error Handling** - User-friendly error messages
✅ **Token Management** - Automatic JWT handling

---

## 🎯 Next Steps

1. **Test everything works** - Follow checklist above
2. **Create Dashboard screen** - Show user info after login
3. **Add logout** - Clear token and navigate to login
4. **Implement transfers** - Connect payment service (requires backend proxy)
5. **Add transaction history** - Display past payments
6. **Improve UI** - Add loading skeletons, animations
7. **Add notifications** - Show real-time updates from Kafka

---

## 📞 Quick Debug Commands

```bash
# Check if services are running
docker ps

# View logs
docker logs user-service
docker logs postgres
docker logs redis

# Test API
curl http://localhost:8080/health

# Restart services
docker-compose restart

# Full reset
docker-compose down -v && docker-compose up
```

---

## 📖 Documentation Files

1. **[API_INTEGRATION.md](API_INTEGRATION.md)** - Complete endpoint reference
2. **[QUICK_START.md](QUICK_START.md)** - Setup and usage guide
3. **[This file](TROUBLESHOOTING.md)** - Issues and solutions

---

**Everything is now connected! Your payment app should be fully functional.** 🎉
