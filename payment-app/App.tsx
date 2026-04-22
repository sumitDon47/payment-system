import { StatusBar } from 'expo-status-bar';
import { NavigationProvider, useNavigation } from './src/navigation/NavigationContext';
import LoginScreen from './src/screens/LoginScreen';
import SignUpScreen from './src/screens/SignUpScreen';
import './global.css';

function AppContent() {
  const { currentScreen } = useNavigation();

  return (
    <>
      {currentScreen === 'login' ? <LoginScreen /> : <SignUpScreen />}
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
