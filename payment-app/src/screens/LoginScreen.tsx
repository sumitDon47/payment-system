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
import { StorageUtil } from '../api/storage';
import { useNavigation } from '../navigation/NavigationContext';

export default function LoginScreen() {
  const { navigate } = useNavigation();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState<{ email?: string; password?: string }>({});

  const validateForm = (): boolean => {
    const newErrors: typeof errors = {};

    if (!email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      newErrors.email = 'Please enter a valid email';
    }

    if (!password) {
      newErrors.password = 'Password is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleLogin = async () => {
    if (!validateForm()) {
      return;
    }

    setLoading(true);
    try {
      // Call the login API
      const response = await userAPI.login(email, password);

      if (response.token) {
        // Save JWT Token and user info
        await StorageUtil.setItem('jwt_token', response.token);
        await StorageUtil.setItem('user_id', response.user.id);
        await StorageUtil.setItem('user_name', response.user.name);
        await StorageUtil.setItem('user_email', response.user.email);

        Alert.alert('Success', 'Logged in successfully! 🎉', [
          {
            text: 'OK',
            onPress: () => {
              // Clear form
              setEmail('');
              setPassword('');
              setErrors({});
            },
          },
        ]);
      }
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || error.message || 'Failed to login';
      console.error('Login error:', error);
      Alert.alert('Error', errorMessage);
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
          PaymentApp
        </Text>
        <Text className="text-gray-500 mb-8 text-center text-lg">
          Sign in to manage your money
        </Text>

        {/* Email Input */}
        <View className="mb-4">
          <Text className="text-gray-700 mb-2 font-medium ml-1">Email Address</Text>
          <TextInput
            className={`w-full bg-gray-50 border rounded-2xl px-4 py-4 text-gray-900 text-lg ${
              errors.email ? 'border-red-500' : 'border-gray-200'
            }`}
            placeholder="john@example.com"
            placeholderTextColor="#9ca3af"
            value={email}
            onChangeText={(text) => {
              setEmail(text);
              if (text.trim()) setErrors({ ...errors, email: undefined });
            }}
            keyboardType="email-address"
            autoCapitalize="none"
            editable={!loading}
          />
          {errors.email && (
            <Text className="text-red-500 text-sm mt-1 ml-1">{errors.email}</Text>
          )}
        </View>

        {/* Password Input */}
        <View className="mb-6">
          <Text className="text-gray-700 mb-2 font-medium ml-1">Password</Text>
          <TextInput
            className={`w-full bg-gray-50 border rounded-2xl px-4 py-4 text-gray-900 text-lg ${
              errors.password ? 'border-red-500' : 'border-gray-200'
            }`}
            placeholder="Enter your password"
            placeholderTextColor="#9ca3af"
            value={password}
            onChangeText={(text) => {
              setPassword(text);
              if (text) setErrors({ ...errors, password: undefined });
            }}
            secureTextEntry
            editable={!loading}
          />
          {errors.password && (
            <Text className="text-red-500 text-sm mt-1 ml-1">{errors.password}</Text>
          )}
          <TouchableOpacity className="mt-2 items-end" disabled={loading}>
            <Text className="text-blue-600 font-medium">Forgot password?</Text>
          </TouchableOpacity>
        </View>

        {/* Login Button */}
        <TouchableOpacity
          className={`w-full bg-blue-600 rounded-2xl py-4 items-center flex-row justify-center mb-4 ${
            loading ? 'opacity-70' : ''
          }`}
          onPress={handleLogin}
          disabled={loading}
        >
          {loading ? (
            <ActivityIndicator color="#ffffff" />
          ) : (
            <Text className="text-white font-bold text-lg">Log In</Text>
          )}
        </TouchableOpacity>

        {/* Sign Up Link */}
        <View className="flex-row justify-center items-center">
          <Text className="text-gray-600">Don't have an account? </Text>
          <TouchableOpacity onPress={() => navigate('signup')} disabled={loading}>
            <Text className="text-blue-600 font-semibold">Sign up</Text>
          </TouchableOpacity>
        </View>
      </View>
    </KeyboardAvoidingView>
  );
}