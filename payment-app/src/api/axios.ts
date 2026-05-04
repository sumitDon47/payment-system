import axios from 'axios';
import { Platform } from 'react-native';
import { StorageUtil } from './storage';

// Detect if running on web
const isWeb = typeof window !== 'undefined';

// For web, use localhost. For Android emulator, use 10.0.2.2. For iOS, use localhost.
const BASE_URL = isWeb
  ? 'http://localhost:8082'
  : Platform.OS === 'android'
  ? 'http://10.0.2.2:8082'
  : 'http://localhost:8082';

export const apiClient = axios.create({
  baseURL: BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 10000,
});

// Automatically attach JWT token to every request
// Works for both web and native apps
apiClient.interceptors.request.use(async (config) => {
  try {
    let token = null;

    if (isWeb) {
      // Web: use sessionStorage
      token = sessionStorage.getItem('jwt_token');
    } else {
      // Native: use secure storage
      token = await StorageUtil.getItem('jwt_token');
    }

    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  } catch (error) {
    console.error('Error reading token from storage:', error);
  }
  return config;
});

// Add response error interceptor for better error handling
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    // Handle 401 Unauthorized - token expired or invalid
    if (error.response?.status === 401) {
      console.warn('⚠️ Unauthorized - token may be expired');
      if (!isWeb) {
        StorageUtil.removeItem('jwt_token');
      }
    }
    return Promise.reject(error);
  }
);
