import axios from 'axios';
import { Platform } from 'react-native';

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
});

// Automatically attach JWT token to every request
apiClient.interceptors.request.use(async (config) => {
  try {
    let token = null;

    if (isWeb) {
      // Web: use sessionStorage
      token = sessionStorage.getItem('jwt_token');
    } else {
      // Native: will be handled by storage utility
      // For now, just skip - LoginScreen will use StorageUtil
    }

    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  } catch (error) {
    console.error('Error reading token:', error);
  }
  return config;
});
