# 🚀 Quick Setup Guide - Frontend API Integration

## What Was Created

I've created a complete API integration setup for your payment system frontend:

### 📁 New Files

1. **[src/api/services.ts](src/api/services.ts)** - Centralized API service with all endpoints
2. **[src/screens/SignUpScreen.tsx](src/screens/SignUpScreen.tsx)** - Complete sign-up screen with validation
3. **[API_INTEGRATION.md](API_INTEGRATION.md)** - Comprehensive API documentation
4. **This file** - Quick setup guide

### ✅ Updated Files

1. **[src/screens/LoginScreen.tsx](src/screens/LoginScreen.tsx)** - Updated to use API service

---

## 🔧 API Services Available

### `userAPI` - User authentication and profile

```typescript
import { userAPI } from '../api/services';

// Register
const { token, user } = await userAPI.register(name, email, password);

// Login
const { token, user } = await userAPI.login(email, password);

// Get profile
const profile = await userAPI.getProfile();

// Get wallet balance
const { balance } = await userAPI.getWalletBalance();

// Health check
const health = await userAPI.healthCheck();
```

### `paymentAPI` - Payment operations (to be implemented)

```typescript
import { paymentAPI } from '../api/services';

// Send payment (requires backend proxy endpoint first)
const { transaction_id } = await paymentAPI.sendPayment(
  receiver_id,
  amount,
  currency,
  note
);
```

---

## 📱 Using the Screens

### LoginScreen

```typescript
<LoginScreen
  onLoginSuccess={() => navigation.replace('Home')}
  onNavigateToSignUp={() => navigation.navigate('SignUp')}
/>
```

**Features:**
- ✅ Email validation
- ✅ Password validation
- ✅ Error handling
- ✅ Loading state
- ✅ Link to sign up

### SignUpScreen

```typescript
<SignUpScreen
  onSignUpSuccess={() => navigation.replace('Home')}
  onNavigateToLogin={() => navigation.replace('Login')}
/>
```

**Features:**
- ✅ Name, email, password validation
- ✅ Password confirmation
- ✅ Password strength requirements (8+ chars)
- ✅ Email format validation
- ✅ Error messages
- ✅ Link to login

---

## 🌐 Running the Frontend

### 1. Install Dependencies

```bash
cd payment-app
npm install
```

### 2. Start Expo Web

```bash
npm run web
```

This will open your app at **http://localhost:3000** (or next available port)

### 3. Sign Up / Login

1. Click "Sign up" to create an account
2. Or click "Log In" if you already have an account
3. JWT token will be saved automatically
4. You'll be logged in

---

## 🔗 Connecting Navigation

Update your `App.tsx` to include both screens:

```typescript
import React, { useState, useEffect } from 'react';
import * as SecureStore from 'expo-secure-store';
import LoginScreen from './src/screens/LoginScreen';
import SignUpScreen from './src/screens/SignUpScreen';
import HomeScreen from './src/screens/HomeScreen'; // (create this)

export default function App() {
  const [state, dispatch] = React.useReducer(
    (prevState, action) => {
      switch (action.type) {
        case 'RESTORE_TOKEN':
          return {
            ...prevState,
            userToken: action.token,
            isLoading: false,
          };
        case 'SIGN_IN':
          return {
            ...prevState,
            isSignout: false,
            userToken: action.token,
          };
        case 'SIGN_UP':
          return {
            ...prevState,
            isSignout: false,
            userToken: action.token,
          };
        case 'SIGN_OUT':
          return {
            ...prevState,
            isSignout: true,
            userToken: null,
          };
      }
    },
    {
      isLoading: true,
      isSignout: false,
      userToken: null,
    }
  );

  useEffect(() => {
    const bootstrapAsync = async () => {
      let userToken;
      try {
        userToken = await SecureStore.getItemAsync('jwt_token');
      } catch (e) {
        // Restoring token failed
      }

      dispatch({ type: 'RESTORE_TOKEN', token: userToken });
    };

    bootstrapAsync();
  }, []);

  const authContext = React.useMemo(
    () => ({
      signIn: async (email, password) => {
        const response = await userAPI.login(email, password);
        await SecureStore.setItemAsync('jwt_token', response.token);
        dispatch({ type: 'SIGN_IN', token: response.token });
      },
      signUp: async (name, email, password) => {
        const response = await userAPI.register(name, email, password);
        await SecureStore.setItemAsync('jwt_token', response.token);
        dispatch({ type: 'SIGN_UP', token: response.token });
      },
      signOut: async () => {
        await SecureStore.deleteItemAsync('jwt_token');
        dispatch({ type: 'SIGN_OUT' });
      },
    }),
    []
  );

  if (state.isLoading) {
    return <SplashScreen />; // Your loading screen
  }

  return (
    <NavigationContainer>
      {state.userToken == null ? (
        <Stack.Navigator
          screenOptions={{
            headerShown: false,
          }}
        >
          <Stack.Screen
            name="SignIn"
            options={{
              animationEnabled: false,
            }}
          >
            {(props) => (
              <LoginScreen
                {...props}
                onNavigateToSignUp={() => props.navigation.navigate('SignUp')}
              />
            )}
          </Stack.Screen>
          <Stack.Screen
            name="SignUp"
            options={{
              animationEnabled: false,
            }}
          >
            {(props) => (
              <SignUpScreen
                {...props}
                onNavigateToLogin={() => props.navigation.navigate('SignIn')}
              />
            )}
          </Stack.Screen>
        </Stack.Navigator>
      ) : (
        <Stack.Navigator>
          <Stack.Screen name="Home" component={HomeScreen} />
          {/* Other authenticated screens */}
        </Stack.Navigator>
      )}
    </NavigationContainer>
  );
}
```

---

## 📋 Checklist

- [ ] Backend is running (`docker-compose up`)
  - [ ] PostgreSQL (port 5432)
  - [ ] Redis (port 6379)
  - [ ] User Service (port 8080)
  - [ ] Payment Service (port 9090)
  
- [ ] Frontend dependencies installed (`npm install`)

- [ ] Expo web is running (`npm run web`)

- [ ] Test sign up with valid data:
  ```
  Name: John Doe
  Email: john@example.com
  Password: SecurePass123!
  ```

- [ ] Test login with same credentials

- [ ] Token saved in secure storage

---

## 🔍 Debugging

### Check if API is responding

```bash
curl http://localhost:8080/health
```

Should return:
```json
{
  "status": "ok",
  "service": "user-service",
  "redis": "ok"
}
```

### Check browser console for errors

1. Open browser DevTools (F12)
2. Go to Console tab
3. Look for error messages

### Check backend logs

```bash
docker logs user-service
docker logs postgres
docker logs redis
```

### Test registration API directly

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "TestPass123!"
  }'
```

---

## 📚 Next Steps

1. **Create Home/Dashboard screen** - Show user profile and wallet
2. **Add logout functionality** - Clear token and navigate to login
3. **Implement payment/transfer** - Connect to payment service
4. **Add transaction history** - Fetch and display past payments
5. **Add error handling** - Better error messages and recovery
6. **Add refresh logic** - Auto-refresh balance, handle token expiration

---

## 💡 Best Practices

1. **Always save token after login/signup**
   ```typescript
   await SecureStore.setItemAsync('jwt_token', response.token);
   ```

2. **Check token before authenticated requests**
   ```typescript
   const token = await SecureStore.getItemAsync('jwt_token');
   if (!token) navigation.replace('Login');
   ```

3. **Handle 401 errors (token expired)**
   ```typescript
   if (error.response?.status === 401) {
     await SecureStore.deleteItemAsync('jwt_token');
     navigation.replace('Login');
   }
   ```

4. **Validate input before sending**
   ```typescript
   const validateForm = () => {
     // Check name, email, password, etc.
     return isValid;
   };
   ```

5. **Use API service methods consistently**
   ```typescript
   // ✅ Good
   const { token } = await userAPI.login(email, password);
   
   // ❌ Avoid
   const { token } = await apiClient.post('/login', { email, password });
   ```

---

## 📞 Need Help?

- **Check API_INTEGRATION.md** for detailed endpoint documentation
- **Check console errors** for specific error messages
- **Check backend logs** with `docker logs`
- **Test with curl** to isolate frontend vs backend issues
