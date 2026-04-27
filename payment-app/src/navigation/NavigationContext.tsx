import React, { createContext, useContext, useState } from 'react';

type Screen = 'login' | 'signup' | 'wallet' | 'forgot-password' | 'reset-password';

interface NavigationContextType {
  currentScreen: Screen;
  navigate: (screen: Screen) => void;
  resetToken?: string;
  setResetToken: (token: string) => void;
}

const NavigationContext = createContext<NavigationContextType | undefined>(undefined);

export function NavigationProvider({ children }: { children: React.ReactNode }) {
  const [currentScreen, setCurrentScreen] = useState<Screen>('login');
  const [resetToken, setResetToken] = useState<string>('');

  return (
    <NavigationContext.Provider value={{ currentScreen, navigate: setCurrentScreen, resetToken, setResetToken }}>
      {children}
    </NavigationContext.Provider>
  );
}

export function useNavigation() {
  const context = useContext(NavigationContext);
  if (!context) {
    throw new Error('useNavigation must be used within NavigationProvider');
  }
  return context;
}
