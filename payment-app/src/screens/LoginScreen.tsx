import React, { useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { userAPI } from '../api/services';
import { StorageUtil } from '../api/storage';
import { useNavigation } from '../navigation/NavigationContext';
import { Input, FormError, FormSuccess } from '../components/FormComponents';
import { Button, Card, Divider } from '../components/UI';
import { colors } from '../styles/colors';
import { spacing, borderRadius, scale } from '../styles/theme';

export default function LoginScreen() {
  const { navigate } = useNavigation();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [mpin, setMpin] = useState('');
  const [authMode, setAuthMode] = useState<'password' | 'mpin'>('password');
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState<{ email?: string; password?: string; mpin?: string }>({});
  const [successMessage, setSuccessMessage] = useState('');

  const validateForm = (): boolean => {
    const newErrors: typeof errors = {};

    if (!email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      newErrors.email = 'Please enter a valid email';
    }

    if (authMode === 'password') {
      if (!password) {
        newErrors.password = 'Password is required';
      } else if (password.length < 6) {
        newErrors.password = 'Password must be at least 6 characters';
      }
    } else {
      if (!/^\d{4}$/.test(mpin)) {
        newErrors.mpin = 'MPIN must be exactly 4 digits';
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleLogin = async () => {
    if (!validateForm()) {
      return;
    }

    setLoading(true);
    setSuccessMessage('');
    try {
      console.log('🔄 Attempting login with email:', email);
      const response = await userAPI.login(
        email,
        authMode === 'password' ? password : undefined,
        authMode === 'mpin' ? mpin : undefined
      );

      console.log('✅ Login successful:', response);

      if (response.token) {
        await StorageUtil.setItem('jwt_token', response.token);
        await StorageUtil.setItem('user_id', response.user.id);
        await StorageUtil.setItem('user_name', response.user.name);
        await StorageUtil.setItem('user_email', response.user.email);

        setSuccessMessage('✓ Logged in successfully! Redirecting...');
        
        setTimeout(() => {
          setEmail('');
          setPassword('');
          setMpin('');
          setErrors({});
          navigate('wallet');
        }, 1000);
      }
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || error.message || 'Failed to login';
      console.error('❌ Login error:', error);
      setErrors({ email: errorMessage });
    } finally {
      setLoading(false);
    }
  };

  return (
    <KeyboardAvoidingView
      style={{ flex: 1, backgroundColor: colors.background }}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView
        contentContainerStyle={{ flexGrow: 1, paddingHorizontal: spacing.lg }}
        keyboardShouldPersistTaps="handled"
        showsVerticalScrollIndicator={false}
      >
        {/* Header Section */}
        <View style={{ alignItems: 'center', marginTop: spacing['3xl'], marginBottom: spacing['3xl'] }}>
          <View
            style={{
              width: scale(80),
              height: scale(80),
              borderRadius: borderRadius.full,
              backgroundColor: colors.primary,
              justifyContent: 'center',
              alignItems: 'center',
              marginBottom: spacing.xl,
              shadowColor: colors.primary,
              shadowOffset: { width: 0, height: 4 },
              shadowOpacity: 0.25,
              shadowRadius: scale(12),
              elevation: 8,
            }}
          >
            <Text style={{ fontSize: scale(40) }}>💳</Text>
          </View>
          <Text style={{ fontSize: scale(32), fontWeight: '800', color: colors.text, marginBottom: spacing.md, letterSpacing: 0.5 }}>
            PaymentHub
          </Text>
          <Text style={{ fontSize: scale(15), color: colors.textSecondary, textAlign: 'center', fontWeight: '500' }}>
            Secure, fast & simple payments
          </Text>
        </View>

        <Card padding={spacing.xl} borderRadius={borderRadius.xl} shadow="lg">
          {/* Title */}
          <Text style={{ fontSize: scale(24), fontWeight: '800', color: colors.text, marginBottom: spacing.sm, letterSpacing: 0.3 }}>
            Welcome Back
          </Text>
          <Text style={{ fontSize: scale(14), color: colors.textSecondary, marginBottom: spacing.xl, fontWeight: '500' }}>
            Sign in to continue
          </Text>

          <View style={{ flexDirection: 'row', marginBottom: spacing.lg, backgroundColor: colors.surfaceDark, borderRadius: borderRadius.lg, padding: spacing.xs }}>
            <TouchableOpacity
              onPress={() => {
                setAuthMode('password');
                setErrors({ ...errors, mpin: undefined });
              }}
              style={{
                flex: 1,
                paddingVertical: spacing.sm,
                borderRadius: borderRadius.md,
                backgroundColor: authMode === 'password' ? colors.primary : 'transparent',
                alignItems: 'center',
              }}
            >
              <Text style={{ color: authMode === 'password' ? colors.textInverse : colors.textSecondary, fontWeight: '700', fontSize: scale(13) }}>
                Password
              </Text>
            </TouchableOpacity>
            <TouchableOpacity
              onPress={() => {
                setAuthMode('mpin');
                setErrors({ ...errors, password: undefined });
              }}
              style={{
                flex: 1,
                paddingVertical: spacing.sm,
                borderRadius: borderRadius.md,
                backgroundColor: authMode === 'mpin' ? colors.primary : 'transparent',
                alignItems: 'center',
              }}
            >
              <Text style={{ color: authMode === 'mpin' ? colors.textInverse : colors.textSecondary, fontWeight: '700', fontSize: scale(13) }}>
                MPIN
              </Text>
            </TouchableOpacity>
          </View>

          {/* Error Message */}
          {errors.email && !successMessage && (
            <FormError message={errors.email} />
          )}

          {/* Success Message */}
          {successMessage && (
            <FormSuccess message={successMessage} />
          )}

          {/* Email Input */}
          <Input
            label="Email Address"
            placeholder="you@example.com"
            value={email}
            onChangeText={(text) => {
              setEmail(text);
              if (text.trim()) setErrors({ ...errors, email: undefined });
            }}
            keyboardType="email-address"
            error={errors.email}
            icon="✉️"
            helperText="Enter your registered email"
          />

          {authMode === 'password' ? (
            <Input
              label="Password"
              placeholder="••••••••"
              value={password}
              onChangeText={(text) => {
                setPassword(text);
                if (text) setErrors({ ...errors, password: undefined });
              }}
              secureTextEntry
              error={errors.password}
              icon="🔐"
            />
          ) : (
            <Input
              label="MPIN"
              placeholder="0000"
              value={mpin}
              onChangeText={(text) => {
                const digits = text.replace(/\D/g, '').slice(0, 4);
                setMpin(digits);
                if (digits) setErrors({ ...errors, mpin: undefined });
              }}
              type="mpin"
              secureTextEntry
              error={errors.mpin}
              icon="🔢"
              helperText="Use your 4-digit security MPIN"
            />
          )}

          {/* Forgot Password Link */}
          <TouchableOpacity
            onPress={() => navigate('forgot-password')}
            disabled={loading}
            style={{ marginBottom: spacing.lg, alignItems: 'flex-end' }}
          >
            <Text style={{ color: colors.primary, fontSize: scale(14), fontWeight: '600', letterSpacing: 0.2 }}>
              Forgot password?
            </Text>
          </TouchableOpacity>

          {/* Login Button */}
          <Button
            title={loading ? 'Signing in...' : 'Sign In'}
            onPress={handleLogin}
            loading={loading}
            disabled={loading}
            fullWidth
            size="lg"
          />

          {/* Divider */}
          <Divider margin={spacing.xl} />

          {/* Sign Up Link */}
          <View style={{ flexDirection: 'row', justifyContent: 'center', gap: spacing.sm }}>
            <Text style={{ color: colors.textSecondary, fontSize: scale(14), fontWeight: '500' }}>
              Don't have an account?
            </Text>
            <TouchableOpacity onPress={() => navigate('signup')} disabled={loading}>
              <Text style={{ color: colors.primary, fontWeight: '700', fontSize: scale(14), letterSpacing: 0.2 }}>
                Sign up
              </Text>
            </TouchableOpacity>
          </View>
        </Card>

        {/* Footer */}
        <Text
          style={{
            fontSize: scale(11),
            color: colors.textTertiary,
            textAlign: 'center',
            marginTop: spacing['3xl'],
            marginBottom: spacing.xl,
            fontWeight: '500',
            letterSpacing: 0.2,
          }}
        >
          By signing in, you agree to our{'\n'}Terms of Service and Privacy Policy
        </Text>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}