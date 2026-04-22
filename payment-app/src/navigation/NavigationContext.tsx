import React, { createContext, useContext, useState } from 'react';

type Screen = 'login' | 'signup';

interface NavigationContextType {
  currentScreen: Screen;
  navigate: (screen: Screen) => void;
}

const NavigationContext = createContext<NavigationContextType | undefined>(undefined);

export function NavigationProvider({ children }: { children: React.ReactNode }) {
  const [currentScreen, setCurrentScreen] = useState<Screen>('login');

  return (
    <NavigationContext.Provider value={{ currentScreen, navigate: setCurrentScreen }}>
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
