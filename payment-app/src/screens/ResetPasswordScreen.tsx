import React, { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { userAPI } from '../api/services';
import { useNavigation } from '../navigation/NavigationContext';

export default function ResetPasswordScreen() {
  const { navigate, resetToken: contextResetToken } = useNavigation();
  const [token, setToken] = useState(contextResetToken || '');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [errors, setErrors] = useState<{
    token?: string;
    password?: string;
    confirmPassword?: string;
  }>({});

  const validateForm = (): boolean => {
    const newErrors: typeof errors = {};

    if (!token.trim()) {
      newErrors.token = 'Reset token is required. Copy it from the email link.';
    }

    if (!password) {
      newErrors.password = 'Password is required';
    } else if (password.length < 8) {
      newErrors.password = 'Password must be at least 8 characters';
    } else if (!/(?=.*[A-Za-z])(?=.*\d)/.test(password)) {
      newErrors.password = 'Password must contain letters and numbers';
    }

    if (!confirmPassword) {
      newErrors.confirmPassword = 'Please confirm your password';
    } else if (password !== confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleResetPassword = async () => {
    if (!validateForm()) {
      return;
    }

    setLoading(true);
    try {
      console.log('🔄 Attempting password reset with token:', token.substring(0, 16) + '...');
      const response = await userAPI.resetPassword(token, password);

      console.log('✅ Password reset successful:', response);

      // Use window.alert for web compatibility
      setTimeout(() => {
        window.alert('Success! Your password has been reset successfully! You can now log in with your new password.');
        console.log('📱 Navigating to login screen...');
        // Clear form
        setPassword('');
        setConfirmPassword('');
        setToken('');
        setErrors({});
        // Navigate to login
        navigate('login');
      }, 100);
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || error.message || 'Failed to reset password';
      console.error('❌ Reset password error:', error);
      console.error('Response status:', error.response?.status);
      console.error('Response data:', error.response?.data);

      if (error.response?.status === 400) {
        setErrors({ 
          token: 'Invalid or expired reset token. Please request a new password reset.' 
        });
      }

      window.alert('Error: ' + errorMessage);
    } finally {
      setLoading(false);
    }
  };

  return (
    <KeyboardAvoidingView
      className="flex-1 bg-white justify-center items-center px-6"
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <View className="w-full max-w-sm">
        {/* Header */}
        <Text className="text-4xl font-bold text-gray-900 mb-2 text-center">
          Create New Password
        </Text>
        <Text className="text-gray-500 mb-8 text-center text-lg">
          Set a new password for your account
        </Text>

        {/* Token Input - only show if not from context */}
        {!contextResetToken && (
          <View className="mb-4">
            <Text className="text-gray-700 mb-2 font-medium ml-1">Reset Token</Text>
            <View className="bg-gray-50 border border-gray-200 rounded-2xl px-4 py-3">
              <TextInput
                className="text-gray-900 text-sm"
                placeholder="Paste the token from the email link here"
                placeholderTextColor="#999"
                editable={!loading}
                value={token}
                onChangeText={(text) => {
                  setToken(text.trim());
                  if (errors.token) setErrors({ ...errors, token: undefined });
                }}
              />
            </View>
            {errors.token && (
              <Text className="text-red-500 text-sm mt-2 ml-1">{errors.token}</Text>
            )}
            <Text className="text-gray-500 text-xs mt-2 ml-1">
              Copy the token from the "Click here to reset:" link in your email
            </Text>
          </View>
        )}

        {/* Token Error */}
        {errors.token && contextResetToken && (
          <View className="mb-6 bg-red-50 border border-red-200 rounded-lg p-4">
            <Text className="text-red-600 text-sm font-medium">{errors.token}</Text>
          </View>
        )}

        {/* New Password Input */}
        <View className="mb-4">
          <Text className="text-gray-700 mb-2 font-medium ml-1">New Password</Text>
          <View className="flex-row items-center bg-gray-50 border border-gray-200 rounded-2xl px-4 py-4">
            <TextInput
              className="flex-1 text-gray-900 text-lg"
              placeholder="Enter new password"
              placeholderTextColor="#999"
              secureTextEntry={!showPassword}
              editable={!loading}
              value={password}
              onChangeText={(text) => {
                setPassword(text);
                if (errors.password) setErrors({ ...errors, password: undefined });
              }}
            />
            <TouchableOpacity
              onPress={() => setShowPassword(!showPassword)}
              disabled={loading}
            >
              <Text className="text-blue-600 font-semibold ml-2">
                {showPassword ? 'Hide' : 'Show'}
              </Text>
            </TouchableOpacity>
          </View>
          {errors.password && (
            <Text className="text-red-500 text-sm mt-2 ml-1">{errors.password}</Text>
          )}
          <Text className="text-gray-500 text-xs mt-2 ml-1">
            • At least 8 characters{'\n'}• Must include letters and numbers
          </Text>
        </View>

        {/* Confirm Password Input */}
        <View className="mb-6">
          <Text className="text-gray-700 mb-2 font-medium ml-1">Confirm Password</Text>
          <View className="flex-row items-center bg-gray-50 border border-gray-200 rounded-2xl px-4 py-4">
            <TextInput
              className="flex-1 text-gray-900 text-lg"
              placeholder="Confirm password"
              placeholderTextColor="#999"
              secureTextEntry={!showConfirmPassword}
              editable={!loading}
              value={confirmPassword}
              onChangeText={(text) => {
                setConfirmPassword(text);
                if (errors.confirmPassword) setErrors({ ...errors, confirmPassword: undefined });
              }}
            />
            <TouchableOpacity
              onPress={() => setShowConfirmPassword(!showConfirmPassword)}
              disabled={loading}
            >
              <Text className="text-blue-600 font-semibold ml-2">
                {showConfirmPassword ? 'Hide' : 'Show'}
              </Text>
            </TouchableOpacity>
          </View>
          {errors.confirmPassword && (
            <Text className="text-red-500 text-sm mt-2 ml-1">{errors.confirmPassword}</Text>
          )}
        </View>

        {/* Submit Button */}
        <TouchableOpacity
          className={`w-full py-4 rounded-2xl mb-4 flex-row justify-center items-center ${
            loading ? 'bg-gray-300' : 'bg-blue-600'
          }`}
          onPress={handleResetPassword}
          disabled={loading}
        >
          {loading ? (
            <ActivityIndicator color="#fff" />
          ) : (
            <Text className="text-white font-semibold text-lg">Reset Password</Text>
          )}
        </TouchableOpacity>

        {/* Back to Login */}
        <TouchableOpacity
          className="py-3"
          onPress={() => navigate('login')}
          disabled={loading}
        >
          <Text className="text-center text-blue-600 font-medium text-base">
            Back to Login
          </Text>
        </TouchableOpacity>

        {/* Password Requirements */}
        <View className="mt-8 bg-green-50 border border-green-200 rounded-lg p-4">
          <Text className="text-green-900 text-xs font-semibold mb-2">Password Requirements:</Text>
          <Text className="text-green-800 text-xs leading-5">
            ✓ At least 8 characters long{'\n'}
            ✓ Contains both letters and numbers{'\n'}
            ✓ Passwords must match
          </Text>
        </View>
      </View>
    </KeyboardAvoidingView>
  );
}
