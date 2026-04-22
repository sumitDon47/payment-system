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
  ScrollView,
} from 'react-native';
import { userAPI } from '../api/services';
import { StorageUtil } from '../api/storage';
import { useNavigation } from '../navigation/NavigationContext';

export default function SignUpScreen() {
  const { navigate } = useNavigation();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState<{
    name?: string;
    email?: string;
    password?: string;
    confirmPassword?: string;
  }>({});

  // Validate inputs
  const validateForm = (): boolean => {
    const newErrors: typeof errors = {};

    if (!name.trim()) {
      newErrors.name = 'Name is required';
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
    }

    if (password !== confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSignUp = async () => {
    if (!validateForm()) {
      return;
    }

    setLoading(true);
    try {
      // Call backend register endpoint
      const response = await userAPI.register(name, email, password);

      if (response.token) {
        // Save JWT token and user info
        await StorageUtil.setItem('jwt_token', response.token);
        await StorageUtil.setItem('user_id', response.user.id);
        await StorageUtil.setItem('user_name', response.user.name);
        await StorageUtil.setItem('user_email', response.user.email);

        Alert.alert('Success', 'Account created successfully! 🎉', [
          {
            text: 'OK',
            onPress: () => {
              // Clear form
              setName('');
              setEmail('');
              setPassword('');
              setConfirmPassword('');
              setErrors({});
              // Navigate to Login after success
              navigate('login');
            },
          },
        ]);
      }
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || error.message || 'Failed to sign up';
      console.error('SignUp error:', error);

      // Handle specific error cases
      if (error.response?.status === 409) {
        setErrors({ email: 'Email already exists. Please try another.' });
        Alert.alert('Error', 'This email is already registered. Try logging in instead.');
      } else {
        Alert.alert('Error', errorMessage);
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <KeyboardAvoidingView
      className="flex-1 bg-white"
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView
        contentContainerStyle={{ flexGrow: 1, justifyContent: 'center' }}
        showsVerticalScrollIndicator={false}
        keyboardShouldPersistTaps="handled"
      >
        <View className="px-6 py-8">
          {/* Header */}
          <Text className="text-4xl font-bold text-gray-900 mb-2 text-center">
            PaymentApp
          </Text>
          <Text className="text-gray-500 mb-8 text-center text-lg">
            Create your account
          </Text>

          {/* Name Input */}
          <View className="mb-4">
            <Text className="text-gray-700 mb-2 font-medium ml-1">Full Name</Text>
            <TextInput
              className={`w-full bg-gray-50 border rounded-2xl px-4 py-4 text-gray-900 text-lg ${
                errors.name ? 'border-red-500' : 'border-gray-200'
              }`}
              placeholder="John Doe"
              placeholderTextColor="#9ca3af"
              value={name}
              onChangeText={(text) => {
                setName(text);
                if (text.trim()) setErrors({ ...errors, name: undefined });
              }}
              autoCapitalize="words"
              editable={!loading}
            />
            {errors.name && (
              <Text className="text-red-500 text-sm mt-1 ml-1">{errors.name}</Text>
            )}
          </View>

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
          <View className="mb-4">
            <Text className="text-gray-700 mb-2 font-medium ml-1">Password</Text>
            <TextInput
              className={`w-full bg-gray-50 border rounded-2xl px-4 py-4 text-gray-900 text-lg ${
                errors.password ? 'border-red-500' : 'border-gray-200'
              }`}
              placeholder="At least 8 characters"
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
          </View>

          {/* Confirm Password Input */}
          <View className="mb-6">
            <Text className="text-gray-700 mb-2 font-medium ml-1">Confirm Password</Text>
            <TextInput
              className={`w-full bg-gray-50 border rounded-2xl px-4 py-4 text-gray-900 text-lg ${
                errors.confirmPassword ? 'border-red-500' : 'border-gray-200'
              }`}
              placeholder="Re-enter your password"
              placeholderTextColor="#9ca3af"
              value={confirmPassword}
              onChangeText={(text) => {
                setConfirmPassword(text);
                if (text) setErrors({ ...errors, confirmPassword: undefined });
              }}
              secureTextEntry
              editable={!loading}
            />
            {errors.confirmPassword && (
              <Text className="text-red-500 text-sm mt-1 ml-1">
                {errors.confirmPassword}
              </Text>
            )}
          </View>

          {/* Sign Up Button */}
          <TouchableOpacity
            className={`w-full bg-blue-600 rounded-2xl py-4 items-center flex-row justify-center mb-4 ${
              loading ? 'opacity-70' : ''
            }`}
            onPress={handleSignUp}
            disabled={loading}
          >
            {loading ? (
              <ActivityIndicator color="#ffffff" />
            ) : (
              <Text className="text-white font-bold text-lg">Create Account</Text>
            )}
          </TouchableOpacity>

          {/* Login Link */}
          <View className="flex-row justify-center items-center">
            <Text className="text-gray-600">Already have an account? </Text>
            <TouchableOpacity onPress={() => navigate('login')} disabled={loading}>
              <Text className="text-blue-600 font-semibold">Log in</Text>
            </TouchableOpacity>
          </View>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}
