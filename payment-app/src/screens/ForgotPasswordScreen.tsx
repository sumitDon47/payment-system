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

export default function ForgotPasswordScreen() {
  const { navigate } = useNavigation();
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState<{ email?: string }>({});
  const [submitted, setSubmitted] = useState(false);

  const validateForm = (): boolean => {
    const newErrors: typeof errors = {};

    if (!email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      newErrors.email = 'Please enter a valid email';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleForgotPassword = async () => {
    if (!validateForm()) {
      return;
    }

    setLoading(true);
    try {
      console.log('🔄 Sending forgot password request for:', email);
      const response = await userAPI.forgotPassword(email);

      console.log('✅ Forgot password email sent:', response);
      setSubmitted(true);
      setEmail('');

      // Use window.alert for web compatibility
      window.alert('Check Your Email!\n\nWe\'ve sent a password reset link to your email. Please check your inbox and follow the link to reset your password.\n\n⏱️ The link will expire in 15 minutes.');
      
      // Navigate back to login after alert closes
      setTimeout(() => {
        navigate('login');
      }, 500);
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || error.message || 'Failed to process request';
      console.error('❌ Forgot password error:', error);
      window.alert('Error: ' + errorMessage);
    } finally {
      setLoading(false);
    }
  };

  if (submitted) {
    return (
      <KeyboardAvoidingView
        className="flex-1 bg-white justify-center items-center px-6"
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      >
        <View className="w-full max-w-sm items-center">
          <Text className="text-6xl mb-4 text-center">📧</Text>
          <Text className="text-3xl font-bold text-gray-900 mb-2 text-center">
            Check Your Email
          </Text>
          <Text className="text-gray-600 mb-8 text-center leading-6">
            We've sent a password reset link to{'\n'}
            <Text className="font-semibold">{email}</Text>
          </Text>

          <View className="w-full bg-blue-50 border border-blue-200 rounded-lg p-4 mb-8">
            <Text className="text-blue-900 text-sm leading-5 font-medium">
              ⏱️ The reset link expires in 15 minutes{'\n'}
              {'\n'}
              💡 Don't see the email? Check your spam folder
            </Text>
          </View>

          <TouchableOpacity
            className="w-full bg-blue-600 py-3 rounded-2xl mb-4"
            onPress={() => navigate('login')}
          >
            <Text className="text-white font-semibold text-center">Back to Login</Text>
          </TouchableOpacity>

          <TouchableOpacity
            className="py-3"
            onPress={() => {
              setSubmitted(false);
              setEmail('');
            }}
          >
            <Text className="text-blue-600 font-medium text-base">Try Another Email</Text>
          </TouchableOpacity>
        </View>
      </KeyboardAvoidingView>
    );
  }

  return (
    <KeyboardAvoidingView
      className="flex-1 bg-white justify-center items-center px-6"
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <View className="w-full max-w-sm">
        {/* Header */}
        <Text className="text-4xl font-bold text-gray-900 mb-2 text-center">
          Reset Password
        </Text>
        <Text className="text-gray-500 mb-8 text-center text-lg">
          Enter your email to receive a password reset link
        </Text>

        {/* Email Input */}
        <View className="mb-6">
          <Text className="text-gray-700 mb-2 font-medium ml-1">Email Address</Text>
          <TextInput
            className={`w-full bg-gray-50 border rounded-2xl px-4 py-4 text-gray-900 text-lg ${
              errors.email ? 'border-red-500' : 'border-gray-200'
            }`}
            placeholder="your@email.com"
            placeholderTextColor="#999"
            keyboardType="email-address"
            autoCapitalize="none"
            editable={!loading}
            value={email}
            onChangeText={(text) => {
              setEmail(text);
              if (errors.email) setErrors({});
            }}
          />
          {errors.email && (
            <Text className="text-red-500 text-sm mt-2 ml-1">{errors.email}</Text>
          )}
        </View>

        {/* Submit Button */}
        <TouchableOpacity
          className={`w-full py-4 rounded-2xl mb-4 flex-row justify-center items-center ${
            loading ? 'bg-gray-300' : 'bg-blue-600'
          }`}
          onPress={handleForgotPassword}
          disabled={loading}
        >
          {loading ? (
            <ActivityIndicator color="#fff" />
          ) : (
            <Text className="text-white font-semibold text-lg">Send Reset Link</Text>
          )}
        </TouchableOpacity>

        {/* Back to Login */}
        <TouchableOpacity
          className="py-3"
          onPress={() => {
            setEmail('');
            setErrors({});
            setSubmitted(false);
            navigate('login');
          }}
          disabled={loading}
        >
          <Text className="text-center text-blue-600 font-medium text-base">
            Back to Login
          </Text>
        </TouchableOpacity>

        {/* Info Box */}
        <View className="mt-8 bg-blue-50 border border-blue-200 rounded-lg p-4">
          <Text className="text-blue-900 text-sm leading-5">
            💡 We'll send you a password reset link via email. Check your inbox and follow the link to create a new password.
          </Text>
        </View>
      </View>
    </KeyboardAvoidingView>
  );
}
