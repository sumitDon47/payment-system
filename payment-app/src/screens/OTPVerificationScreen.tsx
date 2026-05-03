import React, { useState, useEffect } from 'react';
import { View, Text, ScrollView, TouchableOpacity, ActivityIndicator, TextInput } from 'react-native';
import { Card, Button, Divider } from '../components/UI';
import { FormError, FormSuccess, Input } from '../components/FormComponents';
import { colors } from '../styles/colors';
import { spacing, borderRadius } from '../styles/theme';
import { userAPI } from '../api/services';
import { StorageUtil } from '../api/storage';
import { useNavigation } from '../navigation/NavigationContext';

export default function OTPVerificationScreen() {
  const { navigate, tempEmail, setTempEmail } = useNavigation();
  const [email, setEmail] = useState(tempEmail || '');
  const [otpCode, setOtpCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [resendLoading, setResendLoading] = useState(false);
  const [resendDisabled, setResendDisabled] = useState(false);
  const [resendTimer, setResendTimer] = useState(0);

  // Initialize email from context on mount
  useEffect(() => {
    if (tempEmail && !email) {
      setEmail(tempEmail);
    }
  }, [tempEmail]);

  // Resend timer countdown
  useEffect(() => {
    let timer: ReturnType<typeof setTimeout> | undefined;
    if (resendTimer > 0) {
      timer = setTimeout(() => setResendTimer(resendTimer - 1), 1000);
    } else if (resendTimer === 0 && resendDisabled) {
      setResendDisabled(false);
    }
    return () => {
      if (timer) {
        clearTimeout(timer);
      }
    };
  }, [resendTimer, resendDisabled]);

  const handleVerifyOTP = async () => {
    if (!email.trim()) {
      setError('Email is required');
      return;
    }

    if (!otpCode.trim()) {
      setError('Verification code is required');
      return;
    }

    if (otpCode.length !== 6 || !/^\d+$/.test(otpCode)) {
      setError('Verification code must be 6 digits');
      return;
    }

    setError('');
    setSuccess('');
    setLoading(true);

    try {
      console.log('🔐 Verifying OTP...');
      const response = await userAPI.verifyOTP(email, otpCode);

      // Save token and user info
      await StorageUtil.setItem('jwt_token', response.token);
      await StorageUtil.setItem('user_id', response.user.id);
      await StorageUtil.setItem('user_name', response.user.name);
      await StorageUtil.setItem('user_email', response.user.email);

      setSuccess('✅ Account verified successfully! Redirecting...');
      console.log('✅ OTP verified, user created:', response.user.email);

      // Redirect to wallet after short delay
      setTimeout(() => {
        setTempEmail('');
        navigate('wallet');
      }, 1500);
    } catch (err: any) {
      const errorMsg = err.response?.data?.error || err.message || 'OTP verification failed';
      console.error('❌ OTP verification error:', err);
      setError(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  const handleResendOTP = async () => {
    if (!email.trim()) {
      setError('Email is required');
      return;
    }

    setError('');
    setSuccess('');
    setResendLoading(true);

    try {
      console.log('📧 Resending OTP...');
      await userAPI.resendOTP(email);
      setSuccess('✅ Verification code resent to your email');
      setResendDisabled(true);
      setResendTimer(60);
      setOtpCode(''); // Clear the input
      console.log('✅ OTP resent to', email);
    } catch (err: any) {
      const errorMsg = err.response?.data?.error || err.message || 'Failed to resend OTP';
      console.error('❌ Resend OTP error:', err);
      setError(errorMsg);
    } finally {
      setResendLoading(false);
    }
  };

  return (
    <ScrollView
      style={{ flex: 1, backgroundColor: colors.background }}
      contentContainerStyle={{ paddingVertical: spacing.lg }}
    >
      {/* Header */}
      <View style={{ paddingHorizontal: spacing.lg, marginBottom: spacing.xl }}>
        <Text style={{ fontSize: 28, fontWeight: 'bold', color: colors.text, marginBottom: spacing.md }}>
          🔐 Verify Your Email
        </Text>
        <Text style={{ fontSize: 14, color: colors.textSecondary, lineHeight: 20 }}>
          We've sent a 6-digit verification code to{'\n'}
          <Text style={{ fontWeight: '600', color: colors.text }}>{email}</Text>
        </Text>
      </View>

      {/* Main Card */}
      <Card padding={spacing.xl} borderRadius={borderRadius.xl}>
        {/* Error Message */}
        {error && <FormError message={error} />}

        {/* Success Message */}
        {success && <FormSuccess message={success} />}

        {/* OTP Input */}
        <View style={{ marginBottom: spacing.lg }}>
          <Text style={{ fontSize: 12, fontWeight: '600', color: colors.textSecondary, marginBottom: spacing.md }}>
            Enter Verification Code
          </Text>
          <TextInput
            style={{
              borderWidth: 2,
              borderColor: otpCode ? colors.primary : colors.border,
              borderRadius: borderRadius.lg,
              padding: spacing.md,
              fontSize: 24,
              fontWeight: 'bold',
              letterSpacing: 8,
              textAlign: 'center',
              color: colors.text,
              backgroundColor: colors.surface,
            }}
            placeholder="000000"
            placeholderTextColor={colors.textSecondary}
            maxLength={6}
            keyboardType="number-pad"
            value={otpCode}
            onChangeText={(text) => {
              setOtpCode(text.replace(/[^0-9]/g, ''));
              if (error) setError('');
            }}
            editable={!loading}
          />
          <Text style={{ fontSize: 12, color: colors.textSecondary, marginTop: spacing.sm }}>
            6-digit code from your email
          </Text>
        </View>

        {/* Email Confirmation */}
        <View
          style={{
            backgroundColor: `${colors.info}10`,
            borderLeftWidth: 3,
            borderLeftColor: colors.info,
            padding: spacing.md,
            borderRadius: borderRadius.md,
            marginBottom: spacing.lg,
          }}
        >
          <Text style={{ fontSize: 12, color: colors.text, lineHeight: 18 }}>
            💡 <Text style={{ fontWeight: '600' }}>Tip:</Text> Check your email's spam folder if you don't see the code within a few minutes.
          </Text>
        </View>

        {/* Verify Button */}
        <Button
          title={loading ? '⏳ Verifying...' : '✓ Verify Account'}
          onPress={handleVerifyOTP}
          variant="primary"
          size="lg"
          loading={loading}
          disabled={loading || !otpCode || otpCode.length !== 6}
          fullWidth
        />

        {/* Divider */}
        <Divider margin={spacing.lg} />

        {/* Resend Section */}
        <View style={{ alignItems: 'center', marginBottom: spacing.lg }}>
          <Text style={{ fontSize: 12, color: colors.textSecondary, marginBottom: spacing.md }}>
            Didn't receive the code?
          </Text>
          <TouchableOpacity
            onPress={handleResendOTP}
            disabled={resendDisabled || resendLoading}
            style={{
              paddingVertical: spacing.md,
              paddingHorizontal: spacing.lg,
              borderRadius: borderRadius.lg,
              borderWidth: 1,
              borderColor: resendDisabled ? colors.border : colors.primary,
              backgroundColor: resendDisabled ? `${colors.primary}10` : 'transparent',
            }}
          >
            {resendLoading ? (
              <ActivityIndicator color={colors.primary} />
            ) : (
              <Text
                style={{
                  color: resendDisabled ? colors.textSecondary : colors.primary,
                  fontWeight: '600',
                  fontSize: 13,
                  textAlign: 'center',
                }}
              >
                {resendDisabled ? `Resend in ${resendTimer}s` : '🔄 Resend Code'}
              </Text>
            )}
          </TouchableOpacity>
        </View>

        {/* Help Text */}
        <View
          style={{
            backgroundColor: colors.surface,
            padding: spacing.md,
            borderRadius: borderRadius.md,
            borderWidth: 1,
            borderColor: colors.border,
          }}
        >
          <Text style={{ fontSize: 11, color: colors.textSecondary, lineHeight: 16 }}>
            <Text style={{ fontWeight: '600', color: colors.text }}>🔒 Security Note:</Text> Never share your verification code with anyone. PaymentApp staff will never ask for your OTP.
          </Text>
        </View>
      </Card>

      {/* Go Back Link */}
      <View style={{ alignItems: 'center', marginTop: spacing.xl }}>
        <TouchableOpacity onPress={() => navigate('signup')}>
          <Text style={{ fontSize: 13, color: colors.primary, fontWeight: '600' }}>
            ← Go Back to Signup
          </Text>
        </TouchableOpacity>
      </View>

      <View style={{ height: spacing.xl }} />
    </ScrollView>
  );
}
