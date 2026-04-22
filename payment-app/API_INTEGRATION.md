# 📚 Payment System - API Integration Documentation

## Overview

This document provides a complete guide to integrating the Payment System APIs with the frontend application. It covers all available endpoints, request/response formats, authentication, and common error handling.

---

## 🔧 API Configuration

### Base URL

The API client is configured in [src/api/axios.ts](src/api/axios.ts):

```typescript
const BASE_URL = Platform.OS === 'android' 
  ? 'http://10.0.2.2:8080'      // Android Emulator
  : 'http://localhost:8080';     // iOS/Web
```

**Note:** Update this if running on a different machine or port.

### Authentication

JWT Token is automatically attached to all requests via axios interceptor:

```typescript
// Token is read from secure storage and added to headers
Authorization: Bearer <jwt_token>
```

---

## 👤 User Service API (HTTP/REST)

All endpoints are hosted on **`http://localhost:8080`**

### 1. Register (Sign Up)

**Endpoint:** `POST /register`

**Purpose:** Create a new user account

**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

**Validation Rules:**
- `name`: Required, non-empty string
- `email`: Required, valid email format
- `password`: Required, minimum 8 characters

**Response (201 Created):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Doe",
    "email": "john@example.com",
    "balance": 0,
    "created_at": "2025-04-22T10:30:00Z",
    "updated_at": "2025-04-22T10:30:00Z"
  }
}
```

**Error Response (409 Conflict):**
```json
{
  "error": "email already exists"
}
```

**Frontend Implementation:**
```typescript
import { userAPI } from '../api/services';

try {
  const response = await userAPI.register('John Doe', 'john@example.com', 'SecurePass123!');
  
  // Save token
  await SecureStore.setItemAsync('jwt_token', response.token);
  await SecureStore.setItemAsync('user_id', response.user.id);
} catch (error) {
  console.error('Signup failed:', error.response?.data?.error);
}
```

---

### 2. Login

**Endpoint:** `POST /login`

**Purpose:** Authenticate user and receive JWT token

**Request Body:**
```json
{
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Doe",
    "email": "john@example.com",
    "balance": 500.50,
    "created_at": "2025-04-22T10:30:00Z",
    "updated_at": "2025-04-22T10:30:00Z"
  }
}
```

**Error Response (401 Unauthorized):**
```json
{
  "error": "invalid email or password"
}
```

**Frontend Implementation:**
```typescript
try {
  const response = await userAPI.login('john@example.com', 'SecurePass123!');
  
  // Save token and user info
  await SecureStore.setItemAsync('jwt_token', response.token);
  await SecureStore.setItemAsync('user_id', response.user.id);
  await SecureStore.setItemAsync('user_name', response.user.name);
  
  // Navigate to dashboard
  navigation.replace('Home');
} catch (error) {
  Alert.alert('Error', error.response?.data?.error);
}
```

---

### 3. Get User Profile

**Endpoint:** `GET /profile`

**Authentication:** ✅ Required (Bearer Token)

**Purpose:** Retrieve authenticated user's profile information

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "John Doe",
  "email": "john@example.com",
  "balance": 500.50,
  "created_at": "2025-04-22T10:30:00Z",
  "updated_at": "2025-04-22T10:30:00Z"
}
```

**Error Response (401 Unauthorized):**
```json
{
  "error": "unauthorized"
}
```

**Frontend Implementation:**
```typescript
try {
  const profile = await userAPI.getProfile();
  console.log('User:', profile.name, 'Balance:', profile.balance);
} catch (error) {
  console.error('Failed to fetch profile:', error);
}
```

---

### 4. Get Wallet Balance

**Endpoint:** `GET /wallet`

**Authentication:** ✅ Required (Bearer Token)

**Purpose:** Get user's account balance (cached from Redis)

**Response (200 OK):**
```json
{
  "balance": 500.50
}
```

**Frontend Implementation:**
```typescript
try {
  const wallet = await userAPI.getWalletBalance();
  console.log('Current balance:', wallet.balance);
} catch (error) {
  console.error('Failed to fetch balance:', error);
}
```

---

### 5. Health Check

**Endpoint:** `GET /health`

**Authentication:** ❌ Not Required

**Purpose:** Check API and Redis status

**Response (200 OK):**
```json
{
  "status": "ok",
  "service": "user-service",
  "redis": "ok"
}
```

**Frontend Implementation:**
```typescript
try {
  const health = await userAPI.healthCheck();
  console.log('API Status:', health.status);
  console.log('Redis Status:', health.redis);
} catch (error) {
  console.error('API is down');
}
```

---

## 💳 Payment Service API (gRPC)

**Note:** Payment Service uses gRPC for service-to-service communication. Frontend calls should go through a proxy endpoint or dedicated backend endpoint.

### Current Status

The payment service endpoints are **not directly callable from frontend** because they use gRPC protocol. To enable frontend calls, you need to:

1. **Option A (Recommended):** Create HTTP proxy endpoints in user-service
2. **Option B:** Implement gRPC-Web gateway
3. **Option C:** Call from backend (Node.js/Go server) and expose via REST

### Proposed Frontend Endpoints (To Be Implemented)

Once implemented in the user-service, these would be available:

#### Send Payment (Transfer)

**Endpoint:** `POST /transfer` _(to be added)_

**Request Body:**
```json
{
  "receiver_id": "recipient-user-id",
  "amount": 100.50,
  "currency": "USD",
  "note": "Payment for services"
}
```

**Response:**
```json
{
  "transaction_id": "txn_123456",
  "sender_id": "...",
  "receiver_id": "...",
  "amount": 100.50,
  "status": "completed",
  "timestamp": "2025-04-22T10:35:00Z"
}
```

---

## 🔐 Authentication & Security

### JWT Token Structure

The JWT token contains:
- **Header:** Algorithm (HS256)
- **Payload:** User ID, Email, Issued At (iat), Expiration (exp)
- **Signature:** HMAC signed with JWT_SECRET

### Token Storage

Tokens are stored securely using `expo-secure-store`:

```typescript
// Save token
await SecureStore.setItemAsync('jwt_token', token);

// Retrieve token (automatic in axios interceptor)
const token = await SecureStore.getItemAsync('jwt_token');

// Clear on logout
await SecureStore.deleteItemAsync('jwt_token');
```

### Token Expiration

- Default expiration: 24 hours
- Check token status before making requests:

```typescript
const isTokenValid = async () => {
  const token = await SecureStore.getItemAsync('jwt_token');
  return !!token;
};
```

---

## 📊 Rate Limiting

| Endpoint | Limit | Burst | Window |
|----------|-------|-------|--------|
| `/register`, `/login` | 5 req/min | 1 | Per IP |
| `/profile`, `/wallet` | 100 req/min | 5 | Per IP |
| `/transfer` | 100 req/min | 5 | Per IP |

**Error Response (429 Too Many Requests):**
```json
{
  "error": "rate limit exceeded"
}
```

**Handling in Frontend:**
```typescript
try {
  await userAPI.login(email, password);
} catch (error) {
  if (error.response?.status === 429) {
    Alert.alert('Rate Limit', 'Too many attempts. Please try again later.');
  }
}
```

---

## 🚨 Common Error Codes

| Code | Message | Meaning | Solution |
|------|---------|---------|----------|
| 400 | Bad Request | Invalid request body | Check request format |
| 401 | Unauthorized | Missing/invalid token | Login again |
| 409 | Conflict | Email already exists | Use different email |
| 429 | Rate Limited | Too many requests | Wait before retrying |
| 500 | Server Error | Internal server error | Retry or contact support |

---

## 🔗 API Service Methods Reference

All API calls are organized in [src/api/services.ts](src/api/services.ts)

### User API Methods

```typescript
// Register new user
await userAPI.register(name, email, password);

// Login
await userAPI.login(email, password);

// Get profile
await userAPI.getProfile();

// Get wallet balance
await userAPI.getWalletBalance();

// Health check
await userAPI.healthCheck();
```

---

## 🔄 Complete Auth Flow Example

```typescript
// 1. SIGN UP
const signUp = async () => {
  try {
    const { token, user } = await userAPI.register(
      'John Doe',
      'john@example.com',
      'SecurePass123!'
    );
    
    // Save token
    await SecureStore.setItemAsync('jwt_token', token);
    
    // Automatically logged in, navigate to home
    navigation.replace('Home');
  } catch (error) {
    Alert.alert('Error', error.response?.data?.error);
  }
};

// 2. LOGIN
const logIn = async () => {
  try {
    const { token } = await userAPI.login(email, password);
    await SecureStore.setItemAsync('jwt_token', token);
    navigation.replace('Home');
  } catch (error) {
    Alert.alert('Error', error.response?.data?.error);
  }
};

// 3. FETCH PROFILE (Authenticated)
const fetchProfile = async () => {
  try {
    const profile = await userAPI.getProfile();
    console.log('User:', profile.name);
  } catch (error) {
    if (error.response?.status === 401) {
      // Token expired, redirect to login
      navigation.replace('Login');
    }
  }
};

// 4. LOGOUT
const logOut = async () => {
  await SecureStore.deleteItemAsync('jwt_token');
  navigation.replace('Login');
};
```

---

## 🛠️ Troubleshooting

### API Connection Issues

**Problem:** `Failed to connect to 10.0.2.2:8080`

**Solutions:**
1. Verify Docker containers are running: `docker ps`
2. Check if user-service port is 8080: `docker logs user-service`
3. If using physical device, update BASE_URL to your machine IP

### Authentication Errors

**Problem:** `401 Unauthorized`

**Solutions:**
1. Ensure JWT token is saved: `await SecureStore.getItemAsync('jwt_token')`
2. Check token hasn't expired (24 hours)
3. Login again to get fresh token

### CORS Errors

**Problem:** `No 'Access-Control-Allow-Origin' header`

**Solution:** CORS is enabled in user-service for `http://localhost:3000` and `http://localhost:5173`

Update in [user-service/middleware/cors.go](../../user-service/middleware/cors.go) if needed

---

## 📱 Environment Configuration

### Web (Expo Web)

```typescript
// src/api/axios.ts
const BASE_URL = 'http://localhost:8080';
```

### Android Emulator

```typescript
const BASE_URL = 'http://10.0.2.2:8080';
```

### iOS Simulator

```typescript
const BASE_URL = 'http://localhost:8080';
```

### Physical Device

```typescript
const BASE_URL = 'http://<YOUR_MACHINE_IP>:8080';
```

---

## 📞 Support

For issues or questions:
1. Check this documentation
2. Review error logs in terminal
3. Check backend logs: `docker logs user-service`
4. Open issue in GitHub repository
