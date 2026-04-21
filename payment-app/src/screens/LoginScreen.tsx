import React, { useState } from 'react';
import { 
  View, 
  Text, 
  TextInput, 
  TouchableOpacity, 
  ActivityIndicator, 
  Alert, 
  KeyboardAvoidingView, 
  Platform 
} from 'react-native';
import * as SecureStore from 'expo-secure-store';
import { apiClient } from '../api/axios';

export default function LoginScreen() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);

  const handleLogin = async () => {
    if (!email || !password) {
      Alert.alert('Error', 'Please enter email and password');
      return;
    }

    setLoading(true);
    try {
      // Makes a POST request to your User Service running on port 8080
      const response = await apiClient.post('/login', { email, password });
      
      // Save the JWT Token securely on the device
      if (response.data.token) {
        await SecureStore.setItemAsync('jwt_token', response.data.token);
        Alert.alert('Success', 'Logged in successfully!');
        
        // Here you would typically navigate to the main dashboard/wallet screen:
        // navigation.replace('Home');
      }
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || 'Failed to login';
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
            className="w-full bg-gray-50 border border-gray-200 rounded-2xl px-4 py-4 text-gray-900 text-lg"
            placeholder="john@example.com"
            placeholderTextColor="#9ca3af"
            value={email}
            onChangeText={setEmail}
            keyboardType="email-address"
            autoCapitalize="none"
          />
        </View>

        {/* Password Input */}
        <View className="mb-8">
          <Text className="text-gray-700 mb-2 font-medium ml-1">Password</Text>
          <TextInput
            className="w-full bg-gray-50 border border-gray-200 rounded-2xl px-4 py-4 text-gray-900 text-lg"
            placeholder="Enter your password"
            placeholderTextColor="#9ca3af"
            value={password}
            onChangeText={setPassword}
            secureTextEntry
          />
          <TouchableOpacity className="mt-2 items-end">
            <Text className="text-blue-600 font-medium">Forgot password?</Text>
          </TouchableOpacity>
        </View>

        {/* Login Button */}
        <TouchableOpacity
          className={`w-full bg-blue-600 rounded-2xl py-4 items-center flex-row justify-center ${loading ? 'opacity-70' : ''}`}
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
        <TouchableOpacity className="mt-6 py-2 items-center flex-row justify-center">
          <Text className="text-gray-600">Don't have an account? </Text>
          <Text className="text-blue-600 font-bold">Sign up</Text>
        </TouchableOpacity>
      </View>
    </KeyboardAvoidingView>
  );
}