import { Dimensions, Platform } from 'react-native';

const windowWidth = Dimensions.get('window').width;
const windowHeight = Dimensions.get('window').height;

const clamp = (value: number, min: number, max: number) => Math.min(max, Math.max(min, value));

const widthScaleFactor =
  Platform.OS === 'web'
    ? clamp(windowWidth / 1280, 0.9, 1.08)
    : clamp(windowWidth / 375, 0.9, 1.15);

const heightScaleFactor =
  Platform.OS === 'web'
    ? clamp(windowHeight / 900, 0.95, 1.08)
    : clamp(windowHeight / 812, 0.9, 1.12);

// Responsive scaling function
export const scale = (size: number) => size * widthScaleFactor;
export const verticalScale = (size: number) => size * heightScaleFactor;

export const isSmallScreen = windowWidth < 375;
export const isMediumScreen = windowWidth >= 375 && windowWidth < 768;
export const isLargeScreen = windowWidth >= 768;

export const spacing = {
  xs: scale(4),
  sm: scale(8),
  md: scale(12),
  lg: scale(16),
  xl: scale(20),
  '2xl': scale(24),
  '3xl': scale(32),
  '4xl': scale(40),
  '5xl': scale(48),
};

export const typography = {
  sizes: {
    xs: scale(12),
    sm: scale(14),
    base: scale(16),
    lg: scale(18),
    xl: scale(20),
    '2xl': scale(24),
    '3xl': scale(30),
    '4xl': scale(36),
    '5xl': scale(48),
  },
  weights: {
    light: '300',
    normal: '400',
    medium: '500',
    semibold: '600',
    bold: '700',
    extrabold: '800',
  },
  lineHeight: {
    tight: 1.2,
    normal: 1.5,
    relaxed: 1.625,
    loose: 2,
  },
};

export const borderRadius = {
  none: 0,
  sm: scale(4),
  md: scale(8),
  lg: scale(12),
  xl: scale(16),
  '2xl': scale(20),
  full: 9999,
};

export const shadows = {
  sm: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.06,
    shadowRadius: scale(2),
    elevation: 2,
  },
  md: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.1,
    shadowRadius: scale(8),
    elevation: 4,
  },
  lg: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.12,
    shadowRadius: scale(16),
    elevation: 8,
  },
  xl: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 12 },
    shadowOpacity: 0.15,
    shadowRadius: scale(24),
    elevation: 12,
  },
  glow: {
    shadowColor: '#0066cc',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.25,
    shadowRadius: scale(12),
    elevation: 8,
  },
};

export const animations = {
  duration: {
    fast: 200,
    normal: 300,
    slow: 500,
  },
  timing: {
    easeIn: 'ease-in',
    easeOut: 'ease-out',
    easeInOut: 'ease-in-out',
  },
};
