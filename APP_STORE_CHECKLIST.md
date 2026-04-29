# App Store Submission Checklist

**Last Updated:** April 29, 2026  
**Status:** Pre-Submission Verification

---

## 1. ✅ Privacy Policy

**Status:** ⚠️ NOT CREATED  
**Requirement:** Both stores require a publicly accessible privacy policy

### What to include:
- [ ] Data collection practices (user info, payment data, location)
- [ ] Data storage methods (encrypted, secure servers)
- [ ] Third-party service usage (Firebase, SendGrid, gRPC)
- [ ] User rights (access, delete, export data)
- [ ] Data retention period
- [ ] Security measures
- [ ] Compliance (GDPR, local laws)

### Action Items:
1. Create `PRIVACY_POLICY.md` in repo
2. Host on public URL (GitHub Pages, your website)
3. Include link in app settings screen
4. Update: `payment-app/src/screens/SettingsScreen.tsx`

---

## 2. ✅ Terms of Service

**Status:** ⚠️ NOT CREATED  
**Requirement:** Both stores require Terms of Service

### What to include:
- [ ] User responsibilities
- [ ] Payment terms & conditions
- [ ] Transaction fees (if any)
- [ ] Acceptable use policy
- [ ] Limitation of liability
- [ ] Dispute resolution
- [ ] Account suspension conditions

### Action Items:
1. Create `TERMS_OF_SERVICE.md` in repo
2. Host on public URL
3. Include link in app settings
4. Get legal review if possible

---

## 3. 📱 App Screenshots

**Status:** ⚠️ NEEDS PREPARATION  
**Requirement:** 2-5 screenshots per platform (English minimum)

### Required Screenshots:
- [ ] Login screen
- [ ] Wallet/balance screen
- [ ] Transaction history
- [ ] Send money flow
- [ ] Success confirmation

### Specifications:
**Android:**
- Minimum: 2 screenshots
- Maximum: 8 screenshots
- Size: 1080x1920 px (minimum)
- Format: PNG or JPEG

**iOS:**
- Minimum: 2 screenshots
- Maximum: 10 screenshots
- iPhone: 1170x2532 px
- iPad: 2048x2732 px (if supporting tablets)

### Localization:
- [ ] English (required)
- [ ] Hindi (recommended for Nepal)
- [ ] Other languages

---

## 4. 📝 App Description & Keywords

**Status:** ⚠️ NEEDS CREATION  
**Requirement:** Store listing metadata

### Store Listing Details:
**Title:**
- [ ] Max 30 characters
- [ ] Clear, descriptive
- Example: "PaymentSystem - Send Money Fast"

**Short Description (Android):**
- [ ] Max 80 characters
- [ ] Hook users immediately

**Full Description:**
- [ ] 4000 characters max (Android)
- [ ] Explain features, benefits, security
- [ ] List supported currencies
- [ ] Mention encryption/security

**Keywords:**
- [ ] payment, money transfer, wallet, digital payment
- [ ] secure, fast, easy
- [ ] Nepal (if targeting)
- [ ] 5 keywords max

**Category:**
- [ ] Finance
- [ ] Banking

---

## 5. 🎂 Age Rating Questionnaire

**Status:** ⚠️ NEEDS COMPLETION  
**Requirement:** ESRB (Android) and IARC (iOS)

### Typical Answers for Payment App:
- [ ] Financial services: **YES**
- [ ] In-app purchases: **NO**
- [ ] Advertising: **NO** (if not implemented)
- [ ] User-generated content: **NO**
- [ ] Location data: **NO** (unless using geo-services)
- [ ] Personal data collection: **YES** (account info)

### Expected Rating: **3+** (All Ages)

---

## 6. 📲 Testing on Real Devices

**Status:** ⚠️ IN PROGRESS  
**Requirement:** Test on multiple devices before submission

### Android Testing Devices:
- [ ] Phone with Android 8.0+
- [ ] Phone with Android 12+
- [ ] Tablet (if supporting)
- [ ] Different screen sizes

### iOS Testing Devices:
- [ ] iPhone with iOS 14+
- [ ] iPhone with latest iOS
- [ ] iPad (if supporting)

### Testing Checklist:
- [ ] App launches without crashes
- [ ] All screens render correctly
- [ ] Text is readable (font sizes)
- [ ] Buttons are tappable
- [ ] No UI overlap on notch devices
- [ ] Dark mode works (if supported)
- [ ] Orientation changes work

---

## 7. 🐛 Fix Any Crashes/Bugs

**Status:** ⚠️ NEEDS VERIFICATION  
**Requirement:** App must be stable, crash-free

### Known Issues to Check:
- [ ] Password reset flow works end-to-end
- [ ] Dark mode doesn't break UI
- [ ] Network errors handled gracefully
- [ ] Login timeout handled
- [ ] Session expiry (24h) handled
- [ ] Rate limiting doesn't break UX

### Testing Tools:
```bash
# React Native crash detection
npx react-native doctor

# Build and test locally
npm run android
npm run ios

# Use Android Emulator/iOS Simulator
# Connect real device and test
```

---

## 8. 🔐 Ensure JWT Authentication Works

**Status:** ✅ IMPLEMENTED  
**Location:** `payment-app/src/api/axios.ts`, `user-service`

### Verification:
- [ ] Login generates JWT token
- [ ] Token stored securely (AsyncStorage with encryption)
- [ ] Token sent with every API request
- [ ] Token refresh on expiry (if implemented)
- [ ] Logout clears token
- [ ] Protected routes blocked without token
- [ ] Token expiry: 24 hours (check if still valid)

### Current Implementation:
```
✅ JWT generation in user-service/handler/user.go
✅ Token stored in AsyncStorage
✅ Axios interceptor adds token to headers
✅ Login/Register working
⚠️ Token refresh - CHECK if implemented
⚠️ Re-authentication on 401 - CHECK if implemented
```

---

## 9. 💳 Test Payment Flows Thoroughly

**Status:** ⚠️ NEEDS TESTING  
**Requirement:** All payment scenarios must work

### Test Scenarios:
- [ ] **Happy Path:** Successful transaction
  - Register sender & receiver
  - Sender sends money to receiver
  - Both balances update correctly
  - Transaction appears in history

- [ ] **Error Handling:**
  - [ ] Insufficient funds → Show error
  - [ ] Invalid receiver → Show error
  - [ ] Network error → Show retry
  - [ ] Transaction timeout → Handle gracefully

- [ ] **Edge Cases:**
  - [ ] Zero amount → Reject
  - [ ] Negative amount → Reject
  - [ ] Very large amount (>1M) → Reject (if limited)
  - [ ] Same sender/receiver → Reject
  - [ ] Non-existent user → Reject

- [ ] **Race Conditions:**
  - [ ] Double-spend prevention (SERIALIZABLE isolation)
  - [ ] Concurrent transactions from same user
  - [ ] Outbox pattern delivering correctly

### Manual Testing Checklist:
```
1. Create Test Users:
   - Sender: sender@test.com
   - Receiver: receiver@test.com
   
2. Verify Initial Balance:
   - Set via database or API
   
3. Send Transactions:
   - Send 100 NPR from sender to receiver
   - Verify sender balance: -100
   - Verify receiver balance: +100
   
4. Test All Payment Endpoints:
   - /wallet (balance)
   - /transfer (send money)
   - Check transaction history
   
5. Test Error Cases:
   - Try with 0 balance
   - Try with invalid user
   - Try with network offline
```

---

## 10. 🔒 Data Encryption for Sensitive Data

**Status:** ⚠️ NEEDS VERIFICATION  
**Requirement:** Passwords, tokens, payment data must be encrypted

### Current Implementation Check:

**Backend (Go):**
- [ ] ✅ Passwords hashed with bcrypt (user-service/middleware/auth.go)
- [ ] ✅ Database connections encrypted (sslmode=disable → should be sslmode=require in production)
- [ ] ⚠️ Kafka messages - CHECK if encrypted
- [ ] ⚠️ gRPC channels - CHECK if using TLS

**Frontend (React Native):**
- [ ] ⚠️ Token storage - CHECK encryption method
- [ ] ✅ HTTPS for all API calls (axios)
- [ ] ⚠️ AsyncStorage - NOT encrypted by default

### Required Actions:

1. **Frontend Token Encryption:**
```bash
npm install react-native-keychain
# or
npm install @react-native-async-storage/async-storage
npm install react-native-encrypted-storage
```

2. **Update Storage:**
```typescript
// Before: Plain AsyncStorage (NOT SECURE)
await AsyncStorage.setItem('token', response.token);

// After: Encrypted Storage
import EncryptedStorage from 'react-native-encrypted-storage';
await EncryptedStorage.setItem('authToken', response.token);
```

3. **Backend HTTPS:**
- [ ] Use TLS/SSL for gRPC connections
- [ ] Certificate validation
- [ ] Production: sslmode=require (not disable)

4. **Data in Transit:**
- [ ] All API calls over HTTPS ✅
- [ ] gRPC with TLS ⚠️ (verify)
- [ ] Kafka TLS/SASL ⚠️ (verify)

---

## Submission Priority Order

### Phase 1: CRITICAL (Do First)
1. Fix crashes/bugs
2. Test payment flows
3. Encrypt JWT tokens on device
4. Write Privacy Policy
5. Create app screenshots

### Phase 2: IMPORTANT (Before Submission)
6. Write Terms of Service
7. Create store listing (description, keywords)
8. Complete age rating questionnaire
9. Test on real devices
10. Enable HTTPS/TLS for production

### Phase 3: NICE-TO-HAVE (After Launch)
11. Localization (multiple languages)
12. Analytics/Crash reporting
13. User support/feedback system
14. A/B testing setup

---

## Store-Specific Requirements

### Android (Google Play)
- [ ] API level 21+ (Android 5.0)
- [ ] 1024x500 feature graphic
- [ ] 512x512 icon
- [ ] Content rating questionnaire
- [ ] Privacy policy URL
- [ ] Contact email

### iOS (App Store)
- [ ] iOS 12.0+
- [ ] 1024x1024 app icon
- [ ] 6+ screenshots per device type
- [ ] Privacy policy URL
- [ ] Support URL
- [ ] App Preview (video optional)
- [ ] IDFA declaration (if using analytics)

---

## Quick Action List

```
TODAY:
  [ ] Review this checklist completely
  [ ] Test payment flow manually
  [ ] List any crashes found
  [ ] Check token encryption

THIS WEEK:
  [ ] Create Privacy Policy (use template)
  [ ] Create Terms of Service
  [ ] Encrypt JWT tokens
  [ ] Take app screenshots
  [ ] Fix all bugs

NEXT WEEK:
  [ ] Test on real devices
  [ ] Complete age rating form
  [ ] Create store listings
  [ ] Final QA testing
  [ ] Submit to Android Play Store

LATER:
  [ ] Submit to iOS App Store
  [ ] Monitor reviews
  [ ] Plan updates
```

---

## Resources

- [Google Play Console Guide](https://support.google.com/googleplay/android-developer/answer/9859152)
- [Apple App Store Guide](https://developer.apple.com/app-store/review/guidelines/)
- [React Native Production Build](https://reactnative.dev/docs/signed-apk-android)
- [OWASP Mobile Security](https://owasp.org/www-project-mobile-top-10/)

---

**Next Step:** Review each section and mark what's complete, what needs work, and what's missing.
