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
import { Button, Card } from '../components/UI';
import { colors } from '../styles/colors';
import { spacing, borderRadius } from '../styles/theme';

export default function LoginScreen() {
  const { navigate } = useNavigation();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState<{ email?: string; password?: string }>({});
  const [successMessage, setSuccessMessage] = useState('');

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
    if (!validateForm()) {
      return;
    }

    setLoading(true);
    setSuccessMessage('');
    try {
      console.log('🔄 Attempting login with email:', email);
      const response = await userAPI.login(email, password);

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
        contentContainerStyle={{ flexGrow: 1, justifyContent: 'center', padding: spacing.lg }}
        keyboardShouldPersistTaps="handled"
      >
        {/* Header Section */}
        <View style={{ alignItems: 'center', marginBottom: spacing['3xl'] }}>
          <View
            style={{
              width: 80,
              height: 80,
              borderRadius: borderRadius.full,
              backgroundColor: `${colors.primary}15`,
              justifyContent: 'center',
              alignItems: 'center',
              marginBottom: spacing.lg,
            }}
          >
            <Text style={{ fontSize: 40 }}>💳</Text>
          </View>
          <Text style={{ fontSize: 32, fontWeight: 'bold', color: colors.text, marginBottom: spacing.md }}>
            PaymentApp
          </Text>
          <Text style={{ fontSize: 16, color: colors.textSecondary, textAlign: 'center' }}>
            Secure payments, simplified
          </Text>
        </View>

        <Card padding={spacing['2xl']}>
          {/* Title */}
          <Text style={{ fontSize: 24, fontWeight: 'bold', color: colors.text, marginBottom: spacing.sm }}>
            Welcome Back
          </Text>
          <Text style={{ fontSize: 14, color: colors.textSecondary, marginBottom: spacing['2xl'] }}>
            Sign in to your account
          </Text>

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
            helperText="We'll never share your email"
          />

          {/* Password Input */}
          <Input
            label="Password"
            placeholder="Enter your password"
            value={password}
            onChangeText={(text) => {
              setPassword(text);
              if (text) setErrors({ ...errors, password: undefined });
            }}
            secureTextEntry
            error={errors.password}
          />

          {/* Forgot Password Link */}
          <TouchableOpacity
            onPress={() => navigate('forgot-password')}
            disabled={loading}
            style={{ marginBottom: spacing.lg }}
          >
            <Text style={{ color: colors.primary, fontSize: 14, fontWeight: '600' }}>
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
          <View style={{ flexDirection: 'row', alignItems: 'center', marginVertical: spacing.xl }}>
            <View style={{ flex: 1, height: 1, backgroundColor: colors.border }} />
            <Text style={{ marginHorizontal: spacing.md, color: colors.textTertiary }}>or</Text>
            <View style={{ flex: 1, height: 1, backgroundColor: colors.border }} />
          </View>

          {/* Sign Up Link */}
          <View style={{ flexDirection: 'row', justifyContent: 'center', gap: spacing.sm }}>
            <Text style={{ color: colors.textSecondary }}>Don't have an account?</Text>
            <TouchableOpacity onPress={() => navigate('signup')} disabled={loading}>
              <Text style={{ color: colors.primary, fontWeight: '600' }}>Sign up</Text>
            </TouchableOpacity>
          </View>
        </Card>

        {/* Footer */}
        <Text
          style={{
            fontSize: 12,
            color: colors.textTertiary,
            textAlign: 'center',
            marginTop: spacing['3xl'],
          }}
        >
          By signing in, you agree to our Terms of Service and Privacy Policy
        </Text>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}