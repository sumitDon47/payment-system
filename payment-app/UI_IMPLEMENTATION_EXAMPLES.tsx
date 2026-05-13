// BEFORE & AFTER: Implementing Enhanced UI
// Copy these patterns to your existing screens

// ============================================================================
// EXAMPLE 1: Enhanced Login Screen
// ============================================================================

/*
FILE: src/screens/LoginScreen.tsx
*/

import React, { useState, useRef } from 'react';
import {
  View,
  ScrollView,
  KeyboardAvoidingView,
  Platform,
  Animated,
} from 'react-native';
import { userAPI } from '../api/services';
import { StorageUtil } from '../api/storage';
import { useNavigation } from '../navigation/NavigationContext';
import { Input, FormError, FormSuccess } from '../components/FormComponents';
import {
  GradientContainer,
  Card,
  EnhancedButton,
  Divider,
  AnimatedMessage,
  LoadingSpinner,
} from '../components/EnhancedComponents';
import { colors } from '../styles/colors';
import { spacing, borderRadius, scale } from '../styles/theme';

export default function LoginScreen() {
  const { navigate } = useNavigation();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [authMode, setAuthMode] = useState<'password' | 'mpin'>('password');
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState<{ email?: string; password?: string }>({});
  const [successMessage, setSuccessMessage] = useState('');

  // Animation references
  const fadeAnim = useRef(new Animated.Value(0)).current;
  const slideAnim = useRef(new Animated.Value(100)).current;

  React.useEffect(() => {
    // Animate on screen load
    Animated.parallel([
      Animated.timing(fadeAnim, {
        toValue: 1,
        duration: 500,
        useNativeDriver: true,
      }),
      Animated.spring(slideAnim, {
        toValue: 0,
        useNativeDriver: true,
      }),
    ]).start();
  }, []);

  const validateForm = (): boolean => {
    const newErrors: typeof errors = {};

    if (!email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      newErrors.email = 'Please enter a valid email';
    }

    if (!password) {
      newErrors.password = 'Password is required';
    } else if (password.length < 6) {
      newErrors.password = 'Password must be at least 6 characters';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleLogin = async () => {
    if (!validateForm()) return;

    setLoading(true);
    try {
      const response = await userAPI.login(email, password);

      if (response.token) {
        await StorageUtil.setItem('jwt_token', response.token);
        await StorageUtil.setItem('user_id', response.user.id);

        setSuccessMessage('✓ Welcome back! Redirecting...');

        setTimeout(() => {
          navigate('Dashboard');
        }, 1500);
      }
    } catch (error: any) {
      setErrors({ email: error.message || 'Login failed' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <GradientContainer variant="primary">
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        style={{ flex: 1 }}
      >
        <ScrollView
          contentContainerStyle={{ flexGrow: 1, justifyContent: 'center' }}
          showsVerticalScrollIndicator={false}
        >
          <Animated.View
            style={{
              opacity: fadeAnim,
              transform: [{ translateY: slideAnim }],
              paddingHorizontal: spacing.lg,
              paddingVertical: spacing.xl,
            }}
          >
            {/* Success Message */}
            <AnimatedMessage
              message={successMessage}
              visible={!!successMessage}
              type="success"
            />

            {/* Header */}
            <View style={{ marginBottom: spacing.xl }}>
              <Text
                style={{
                  fontSize: scale(32),
                  fontWeight: '800',
                  color: colors.primary,
                  marginBottom: spacing.sm,
                  letterSpacing: -0.5,
                }}
              >
                Welcome Back
              </Text>
              <Text
                style={{
                  fontSize: scale(16),
                  color: colors.textSecondary,
                  lineHeight: 24,
                }}
              >
                Sign in to access your account
              </Text>
            </View>

            {/* Login Form Card */}
            <Card variant="elevated" gradient>
              {/* Email Input */}
              <Input
                label="Email Address"
                placeholder="your.email@example.com"
                value={email}
                onChangeText={setEmail}
                keyboardType="email-address"
                icon="✉️"
                error={errors.email}
              />

              {/* Password Input */}
              <Input
                label="Password"
                placeholder="••••••••"
                value={password}
                onChangeText={setPassword}
                secureTextEntry
                icon="🔐"
                error={errors.password}
              />

              {/* Forgot Password Link */}
              <TouchableOpacity style={{ marginBottom: spacing.lg }}>
                <Text
                  style={{
                    color: colors.primary,
                    fontWeight: '600',
                    fontSize: scale(14),
                    textAlign: 'right',
                  }}
                >
                  Forgot Password?
                </Text>
              </TouchableOpacity>

              {/* Login Button */}
              <EnhancedButton
                title={loading ? 'Signing in...' : 'Sign In'}
                onPress={handleLogin}
                variant="primary"
                size="lg"
                loading={loading}
                disabled={loading}
                fullWidth
                showGlow
              />

              {/* Divider */}
              <Divider spacing={spacing.lg} />

              {/* Footer */}
              <View style={{ flexDirection: 'row', justifyContent: 'center' }}>
                <Text style={{ color: colors.textSecondary }}>
                  Don't have an account?{' '}
                </Text>
                <TouchableOpacity onPress={() => navigate('Register')}>
                  <Text
                    style={{
                      color: colors.primary,
                      fontWeight: '700',
                      textDecorationLine: 'underline',
                    }}
                  >
                    Sign up
                  </Text>
                </TouchableOpacity>
              </View>
            </Card>
          </Animated.View>
        </ScrollView>
      </KeyboardAvoidingView>
    </GradientContainer>
  );
}

// ============================================================================
// EXAMPLE 2: Enhanced OTP Verification Screen
// ============================================================================

/*
FILE: src/screens/OTPVerificationScreen.tsx
*/

import React, { useState, useEffect } from 'react';
import { View, Text, TouchableOpacity, Animated } from 'react-native';
import {
  GradientContainer,
  Card,
  EnhancedButton,
  LoadingSpinner,
  Badge,
} from '../components/EnhancedComponents';
import { Input } from '../components/FormComponents';
import { colors } from '../styles/colors';
import { spacing, scale, borderRadius } from '../styles/theme';

export default function OTPVerificationScreen({ route }: any) {
  const [otp, setOtp] = useState('');
  const [loading, setLoading] = useState(false);
  const [timeLeft, setTimeLeft] = useState(120);
  const scaleAnim = React.useRef(new Animated.Value(0)).current;

  useEffect(() => {
    // Animate card entrance
    Animated.spring(scaleAnim, {
      toValue: 1,
      useNativeDriver: true,
    }).start();

    // Countdown timer
    const timer = setInterval(() => {
      setTimeLeft((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);

    return () => clearInterval(timer);
  }, []);

  const handleVerify = async () => {
    if (otp.length !== 6) {
      alert('Please enter 6-digit OTP');
      return;
    }
    setLoading(true);
    // Verify OTP logic
    setLoading(false);
  };

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const isExpired = timeLeft === 0;
  const isExpiringSoon = timeLeft < 30;

  return (
    <GradientContainer variant="secondary">
      <View
        style={{
          flex: 1,
          justifyContent: 'center',
          paddingHorizontal: spacing.lg,
        }}
      >
        <Animated.View
          style={{
            transform: [{ scale: scaleAnim }],
          }}
        >
          <Card variant="elevated">
            {/* Header */}
            <Text
              style={{
                fontSize: scale(28),
                fontWeight: '800',
                color: colors.primary,
                marginBottom: spacing.md,
                textAlign: 'center',
              }}
            >
              Verify OTP
            </Text>

            <Text
              style={{
                fontSize: scale(14),
                color: colors.textSecondary,
                textAlign: 'center',
                marginBottom: spacing.lg,
                lineHeight: 20,
              }}
            >
              Enter the 6-digit code sent to your email
            </Text>

            {/* OTP Input */}
            <Input
              label="OTP Code"
              placeholder="000000"
              value={otp}
              onChangeText={(text) => setOtp(text.slice(0, 6))}
              type="mpin"
              icon="🔑"
              keyboardType="numeric"
            />

            {/* Timer Badge */}
            <View
              style={{
                alignItems: 'center',
                marginBottom: spacing.lg,
              }}
            >
              <Badge
                label={`Expires in ${formatTime(timeLeft)}`}
                variant={
                  isExpired
                    ? 'error'
                    : isExpiringSoon
                      ? 'warning'
                      : 'success'
                }
                size="md"
              />
            </View>

            {/* Verify Button */}
            <EnhancedButton
              title={loading ? 'Verifying...' : 'Verify'}
              onPress={handleVerify}
              variant="primary"
              size="lg"
              loading={loading}
              disabled={loading || isExpired}
              fullWidth
              showGlow
            />

            {/* Resend Section */}
            <View
              style={{
                marginTop: spacing.lg,
                flexDirection: 'row',
                justifyContent: 'center',
                alignItems: 'center',
              }}
            >
              <Text style={{ color: colors.textSecondary }}>
                Didn't receive code?{' '}
              </Text>
              <TouchableOpacity
                disabled={timeLeft > 0}
                onPress={() => setTimeLeft(120)}
              >
                <Text
                  style={{
                    color: timeLeft > 0 ? colors.borderDark : colors.primary,
                    fontWeight: '700',
                    textDecorationLine: 'underline',
                  }}
                >
                  Resend
                </Text>
              </TouchableOpacity>
            </View>
          </Card>
        </Animated.View>
      </View>
    </GradientContainer>
  );
}

// ============================================================================
// EXAMPLE 3: Enhanced Dashboard with Cards
// ============================================================================

/*
FILE: src/screens/DashboardScreen.tsx (excerpt)
*/

const DashboardCard = ({ title, amount, icon, onPress }: any) => {
  const scaleAnim = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    Animated.spring(scaleAnim, {
      toValue: 1,
      useNativeDriver: true,
    }).start();
  }, []);

  return (
    <Animated.View style={{ transform: [{ scale: scaleAnim }] }}>
      <Card onPress={onPress} variant="elevated" gradient>
        <View style={{ flexDirection: 'row', justifyContent: 'space-between' }}>
          <View style={{ flex: 1 }}>
            <Text style={{ fontSize: scale(14), color: colors.textSecondary }}>
              {title}
            </Text>
            <Text
              style={{
                fontSize: scale(24),
                fontWeight: '800',
                color: colors.primary,
                marginTop: spacing.sm,
              }}
            >
              {amount}
            </Text>
          </View>
          <Text style={{ fontSize: scale(32) }}>{icon}</Text>
        </View>
      </Card>
    </Animated.View>
  );
};

// Usage in Dashboard:
<View style={{ gap: spacing.md }}>
  <DashboardCard
    title="Total Balance"
    amount="$5,430.50"
    icon="💰"
    onPress={() => navigate('ViewBalance')}
  />
  <DashboardCard
    title="This Month"
    amount="$1,250.00"
    icon="📊"
    onPress={() => navigate('ViewTransactions')}
  />
  <DashboardCard
    title="Recent Payments"
    amount="5 transactions"
    icon="💳"
    onPress={() => navigate('PaymentHistory')}
  />
</View>

// ============================================================================
// KEY IMPLEMENTATION TIPS
// ============================================================================

/*
1. COLORS
   - Import colors from src/styles/colors.ts
   - Use color constants consistently
   - Example: colors.primary, colors.success, colors.error

2. SPACING
   - Import spacing from src/styles/theme.ts
   - Use spacing.xs, spacing.sm, spacing.md, spacing.lg, spacing.xl
   - Maintains consistent whitespace

3. ANIMATIONS
   - Use Animated from 'react-native'
   - Combine with GestureHandler for smooth interactions
   - Keep animations < 500ms for responsiveness

4. COMPONENTS
   - Use EnhancedButton instead of Button for better UX
   - Use Card for content grouping
   - Use AnimatedMessage for feedback

5. RESPONSIVE DESIGN
   - Test on multiple screen sizes
   - Use scale() for text sizing
   - Use Dimensions.get('window') for layout calculations

6. ACCESSIBILITY
   - Always provide labels for inputs
   - Use high contrast colors (WCAG AA minimum)
   - Support screen readers with proper text

7. PERFORMANCE
   - Use React.memo for expensive components
   - Avoid re-renders with useMemo
   - Profile with React Native Profiler
*/
