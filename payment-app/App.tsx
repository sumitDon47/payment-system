import { StatusBar } from 'expo-status-bar';
import { useColorScheme, LogBox } from 'react-native';
import { useEffect } from 'react';
import { StyleSheet } from 'react-native';
import { NavigationProvider, useNavigation } from './src/navigation/NavigationContext';
import LoginScreen from './src/screens/LoginScreen';
import SignUpScreen from './src/screens/SignUpScreen';
import OTPVerificationScreen from './src/screens/OTPVerificationScreen';
import WalletScreen from './src/screens/WalletScreen';
import ForgotPasswordScreen from './src/screens/ForgotPasswordScreen';
import ResetPasswordScreen from './src/screens/ResetPasswordScreen';
import ProfileScreen from './src/screens/ProfileScreen';
import TransactionHistoryScreen from './src/screens/TransactionHistoryScreen';
import './global.css';

// Suppress deprecated pointerEvents warning from NativeWind
LogBox.ignoreLogs(['props.pointerEvents is deprecated']);

// Initialize dark mode for NativeWind
try {
  StyleSheet.setFlag('darkMode', 'class');
} catch (e) {
  // Silently fail if not available on non-web platforms
}

function AppContent() {
  const { currentScreen, navigate, setResetToken, tempEmail, setTempEmail } = useNavigation();

  useEffect(() => {
    // Handle deep linking from email reset links
    if (typeof window !== 'undefined') {
      const params = new URLSearchParams(window.location.search);
      const token = params.get('token');
      
      if (token) {
        console.log('🔗 Deep link detected with token:', token.substring(0, 16) + '...');
        setResetToken(token);
        navigate('reset-password');
      }
    }
  }, [navigate, setResetToken]);

  return (
    <>
      {currentScreen === 'login' ? (
        <LoginScreen />
      ) : currentScreen === 'signup' ? (
        <SignUpScreen />
      ) : currentScreen === 'otp-verification' ? (
        <OTPVerificationScreen 
          email={tempEmail}
          onVerified={() => setTempEmail('')}
        />
      ) : currentScreen === 'forgot-password' ? (
        <ForgotPasswordScreen />
      ) : currentScreen === 'reset-password' ? (
        <ResetPasswordScreen />
      ) : currentScreen === 'profile' ? (
        <ProfileScreen />
      ) : currentScreen === 'transactions' ? (
        <TransactionHistoryScreen />
      ) : (
        <WalletScreen />
      )}
      <StatusBar style="auto" />
    </>
  );
}

export default function App() {
  return (
    <NavigationProvider>
      <AppContent />
    </NavigationProvider>
  );
}
