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
import { Input, Checkbox, FormError, FormSuccess } from '../components/FormComponents';
import { Button, Card, Divider } from '../components/UI';
import { colors } from '../styles/colors';
import { spacing, borderRadius } from '../styles/theme';

export default function SignUpScreen() {
  const { navigate, setTempEmail } = useNavigation();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [mpin, setMpin] = useState('');
  const [confirmMpin, setConfirmMpin] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [agreeToTerms, setAgreeToTerms] = useState(false);
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState<{
    name?: string;
    email?: string;
    password?: string;
    confirmPassword?: string;
    mpin?: string;
    confirmMpin?: string;
    terms?: string;
  }>({});
  const [successMessage, setSuccessMessage] = useState('');

  const validateForm = (): boolean => {
    const newErrors: typeof errors = {};

    if (!name.trim()) {
      newErrors.name = 'Full name is required';
    }

    if (!email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      newErrors.email = 'Please enter a valid email';
    }

    if (!password) {
      newErrors.password = 'Password is required';
    } else if (password.length < 8) {
      newErrors.password = 'Password must be at least 8 characters';
    } else if (!/(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/.test(password)) {
      newErrors.password = 'Password must contain uppercase, lowercase, and numbers';
    }

    if (password !== confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
    }

    if (!/^\d{4}$/.test(mpin)) {
      newErrors.mpin = 'MPIN must be exactly 4 digits';
    }

    if (mpin !== confirmMpin) {
      newErrors.confirmMpin = 'MPIN does not match';
    }

    if (!agreeToTerms) {
      newErrors.terms = 'You must agree to terms and conditions';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSignUp = async () => {
    if (!validateForm()) {
      return;
    }

    setLoading(true);
    setSuccessMessage('');
    try {
      console.log('🔄 Sending OTP to:', email);
      const response = await userAPI.registerWithOTP(name, email, password, mpin);

      setSuccessMessage('✓ Verification code sent to your email!');
      
      setTimeout(() => {
        setName('');
        setPassword('');
        setMpin('');
        setConfirmMpin('');
        setConfirmPassword('');
        setAgreeToTerms(false);
        setErrors({});
        setTempEmail(email);
        navigate('otp-verification');
      }, 1500);
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || error.message || 'Failed to sign up';
      console.error('SignUp error:', error);

      if (error.response?.status === 409) {
        setErrors({ email: 'This email is already registered. Please log in instead.' });
      } else {
        setErrors({ email: errorMessage });
      }
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
        contentContainerStyle={{ flexGrow: 1, padding: spacing.lg }}
        keyboardShouldPersistTaps="handled"
        showsVerticalScrollIndicator={false}
      >
        {/* Header */}
        <View style={{ alignItems: 'center', marginBottom: spacing['3xl'] }}>
          <View
            style={{
              width: 70,
              height: 70,
              borderRadius: borderRadius.full,
              backgroundColor: `${colors.secondary}15`,
              justifyContent: 'center',
              alignItems: 'center',
              marginBottom: spacing.lg,
            }}
          >
            <Text style={{ fontSize: 36 }}>✨</Text>
          </View>
          <Text style={{ fontSize: 32, fontWeight: 'bold', color: colors.text, marginBottom: spacing.md }}>
            Join PaymentApp
          </Text>
          <Text style={{ fontSize: 14, color: colors.textSecondary, textAlign: 'center' }}>
            Create an account to get started
          </Text>
        </View>

        <Card padding={spacing['2xl']}>
          {/* Error Message */}
          {errors.email && !successMessage && (
            <FormError message={errors.email} />
          )}

          {/* Success Message */}
          {successMessage && (
            <FormSuccess message={successMessage} />
          )}

          {/* Name Input */}
          <Input
            label="Full Name"
            placeholder="John Doe"
            value={name}
            onChangeText={(text) => {
              setName(text);
              if (text.trim()) setErrors({ ...errors, name: undefined });
            }}
            error={errors.name}
          />

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
            helperText="We'll send a verification email"
          />

          {/* Password Input */}
          <Input
            label="Password"
            placeholder="At least 8 characters"
            value={password}
            onChangeText={(text) => {
              setPassword(text);
              if (text) setErrors({ ...errors, password: undefined });
            }}
            secureTextEntry
            error={errors.password}
            helperText="Must contain uppercase, lowercase, and numbers"
          />

          {/* Confirm Password Input */}
          <Input
            label="Confirm Password"
            placeholder="Re-enter your password"
            value={confirmPassword}
            onChangeText={(text) => {
              setConfirmPassword(text);
              if (text) setErrors({ ...errors, confirmPassword: undefined });
            }}
            secureTextEntry
            error={errors.confirmPassword}
          />

          <Input
            label="Create MPIN"
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
            helperText="4-digit security code for fast login and payments"
          />

          <Input
            label="Confirm MPIN"
            placeholder="0000"
            value={confirmMpin}
            onChangeText={(text) => {
              const digits = text.replace(/\D/g, '').slice(0, 4);
              setConfirmMpin(digits);
              if (digits) setErrors({ ...errors, confirmMpin: undefined });
            }}
            type="mpin"
            secureTextEntry
            error={errors.confirmMpin}
          />

          {/* Terms Checkbox */}
          <View style={{ marginBottom: spacing.xl }}>
            <Checkbox
              label="I agree to Terms of Service and Privacy Policy"
              checked={agreeToTerms}
              onToggle={setAgreeToTerms}
            />
            {errors.terms && (
              <Text style={{ color: colors.error, fontSize: 12, marginLeft: spacing.md }}>
                {errors.terms}
              </Text>
            )}
          </View>

          {/* Sign Up Button */}
          <Button
            title={loading ? 'Creating account...' : 'Create Account'}
            onPress={handleSignUp}
            loading={loading}
            disabled={loading}
            fullWidth
            size="lg"
          />

          <Divider margin={spacing.xl} />

          {/* Login Link */}
          <View style={{ flexDirection: 'row', justifyContent: 'center', gap: spacing.sm }}>
            <Text style={{ color: colors.textSecondary }}>Already have an account?</Text>
            <TouchableOpacity onPress={() => navigate('login')} disabled={loading}>
              <Text style={{ color: colors.primary, fontWeight: '600' }}>Sign in</Text>
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
          By creating an account, you agree to our policies
        </Text>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}
