import { StatusBar } from 'expo-status-bar';
import LoginScreen from './src/screens/LoginScreen';
import './global.css';

export default function App() {
  return (
    <>
      <LoginScreen />
      <StatusBar style="auto" />
    </>
  );
}
