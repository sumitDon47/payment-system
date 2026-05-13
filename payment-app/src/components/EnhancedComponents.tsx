// Enhanced UI Components with Animations & Better Styling
// Add these components to src/components/UI.tsx

import React from 'react';
import { View, Text, TouchableOpacity, Animated, Dimensions, ActivityIndicator } from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import { colors } from '../styles/colors';
import { borderRadius, spacing, shadows, scale, typography } from '../styles/theme';

const { width } = Dimensions.get('window');

// ============================================================================
// 1. ENHANCED GRADIENT CONTAINER
// ============================================================================

interface GradientContainerProps {
  children: React.ReactNode;
  variant?: 'primary' | 'secondary' | 'neutral';
  style?: any;
}

export const GradientContainer: React.FC<GradientContainerProps> = ({
  children,
  variant = 'neutral',
  style,
}) => {
  const getGradient = () => {
    switch (variant) {
      case 'primary':
        return ['#f0f4f8', '#e6f0ff'];
      case 'secondary':
        return ['#f0f4f8', '#e6f9f7'];
      default:
        return ['#f9fafb', '#f0f4f8'];
    }
  };

  return (
    <LinearGradient
      colors={getGradient()}
      start={{ x: 0, y: 0 }}
      end={{ x: 1, y: 1 }}
      style={[{ flex: 1 }, style]}
    >
      {children}
    </LinearGradient>
  );
};

// ============================================================================
// 2. ENHANCED CARD WITH ANIMATIONS
// ============================================================================

interface CardProps {
  children: React.ReactNode;
  variant?: 'default' | 'elevated' | 'outlined';
  onPress?: () => void;
  style?: any;
  gradient?: boolean;
  animated?: boolean;
}

export const Card: React.FC<CardProps> = ({
  children,
  variant = 'default',
  onPress,
  style,
  gradient = false,
  animated = true,
}) => {
  const scaleAnim = React.useRef(new Animated.Value(1)).current;

  const handlePressIn = () => {
    if (!animated || !onPress) return;
    Animated.spring(scaleAnim, {
      toValue: 0.98,
      useNativeDriver: true,
    }).start();
  };

  const handlePressOut = () => {
    if (!animated || !onPress) return;
    Animated.spring(scaleAnim, {
      toValue: 1,
      useNativeDriver: true,
    }).start();
  };

  const getStyles = () => {
    switch (variant) {
      case 'elevated':
        return {
          backgroundColor: colors.surface,
          ...shadows.lg,
          borderRadius: borderRadius.xl,
          padding: spacing.lg,
        };
      case 'outlined':
        return {
          backgroundColor: colors.surface,
          borderWidth: 1,
          borderColor: colors.border,
          borderRadius: borderRadius.lg,
          padding: spacing.lg,
        };
      default:
        return {
          backgroundColor: colors.surface,
          ...shadows.md,
          borderRadius: borderRadius.lg,
          padding: spacing.lg,
        };
    }
  };

  const content = (
    <View style={getStyles()}>
      {gradient && (
        <LinearGradient
          colors={['rgba(0, 102, 204, 0.05)', 'rgba(0, 204, 187, 0.05)']}
          style={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            borderRadius: borderRadius.lg,
          }}
        />
      )}
      <View style={{ zIndex: 1 }}>{children}</View>
    </View>
  );

  if (onPress) {
    return (
      <TouchableOpacity
        onPress={onPress}
        onPressIn={handlePressIn}
        onPressOut={handlePressOut}
        activeOpacity={0.7}
      >
        <Animated.View style={[{ transform: [{ scale: scaleAnim }] }, style]}>
          {content}
        </Animated.View>
      </TouchableOpacity>
    );
  }

  return <Animated.View style={[style]}>{content}</Animated.View>;
};

// ============================================================================
// 3. ANIMATED LOADING SPINNER
// ============================================================================

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  color?: string;
}

export const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
  size = 'md',
  color = colors.primary,
}) => {
  const spinAnim = React.useRef(new Animated.Value(0)).current;

  React.useEffect(() => {
    Animated.loop(
      Animated.timing(spinAnim, {
        toValue: 1,
        duration: 1200,
        useNativeDriver: true,
      })
    ).start();
  }, []);

  const spin = spinAnim.interpolate({
    inputRange: [0, 1],
    outputRange: ['0deg', '360deg'],
  });

  const sizes = { sm: 24, md: 40, lg: 60 };
  const borderWidth = size === 'sm' ? 2 : 3;

  return (
    <Animated.View
      style={{
        width: sizes[size],
        height: sizes[size],
        borderRadius: sizes[size] / 2,
        borderWidth,
        borderColor: `${color}20`,
        borderTopColor: color,
        borderRightColor: color,
        transform: [{ rotate: spin }],
      }}
    />
  );
};

// ============================================================================
// 4. ENHANCED BUTTON WITH GLOW & BETTER ANIMATIONS
// ============================================================================

interface EnhancedButtonProps {
  onPress: () => void;
  title: string;
  variant?: 'primary' | 'secondary' | 'accent' | 'outline' | 'danger' | 'success';
  size?: 'sm' | 'md' | 'lg';
  loading?: boolean;
  disabled?: boolean;
  fullWidth?: boolean;
  icon?: React.ReactNode;
  showGlow?: boolean;
}

export const EnhancedButton: React.FC<EnhancedButtonProps> = ({
  onPress,
  title,
  variant = 'primary',
  size = 'md',
  loading = false,
  disabled = false,
  fullWidth = false,
  icon,
  showGlow = true,
}) => {
  const scaleAnim = React.useRef(new Animated.Value(1)).current;
  const glowAnim = React.useRef(new Animated.Value(0)).current;

  const getBackgroundColor = () => {
    if (disabled) return colors.borderDark;
    switch (variant) {
      case 'primary':
        return colors.primary;
      case 'secondary':
        return colors.secondary;
      case 'accent':
        return colors.accent;
      case 'danger':
        return colors.error;
      case 'success':
        return colors.success;
      case 'outline':
        return 'transparent';
      default:
        return colors.primary;
    }
  };

  const getGlowColor = () => {
    switch (variant) {
      case 'primary':
        return colors.primary;
      case 'secondary':
        return colors.secondary;
      case 'accent':
        return colors.accent;
      default:
        return colors.primary;
    }
  };

  const handlePressIn = () => {
    Animated.parallel([
      Animated.spring(scaleAnim, {
        toValue: 0.95,
        useNativeDriver: true,
      }),
      Animated.timing(glowAnim, {
        toValue: 1,
        duration: 200,
        useNativeDriver: false,
      }),
    ]).start();
  };

  const handlePressOut = () => {
    Animated.parallel([
      Animated.spring(scaleAnim, {
        toValue: 1,
        useNativeDriver: true,
      }),
      Animated.timing(glowAnim, {
        toValue: 0,
        duration: 200,
        useNativeDriver: false,
      }),
    ]).start();
  };

  const getPadding = () => {
    switch (size) {
      case 'sm':
        return { horizontal: spacing.md, vertical: spacing.xs };
      case 'md':
        return { horizontal: spacing.lg, vertical: spacing.sm };
      case 'lg':
        return { horizontal: spacing.xl, vertical: spacing.md };
      default:
        return { horizontal: spacing.lg, vertical: spacing.sm };
    }
  };

  const padding = getPadding();
  const glowOpacity = glowAnim.interpolate({
    inputRange: [0, 1],
    outputRange: [0, 0.3],
  });

  return (
    <Animated.View style={{ transform: [{ scale: scaleAnim }] }}>
      <TouchableOpacity
        onPress={onPress}
        onPressIn={handlePressIn}
        onPressOut={handlePressOut}
        disabled={disabled || loading}
        activeOpacity={0.8}
      >
        <View
          style={{
            backgroundColor: getBackgroundColor(),
            paddingHorizontal: padding.horizontal,
            paddingVertical: padding.vertical,
            borderRadius: borderRadius.lg,
            borderWidth: variant === 'outline' ? 2 : 0,
            borderColor: variant === 'outline' ? colors.primary : undefined,
            width: fullWidth ? '100%' : 'auto',
            opacity: disabled ? 0.5 : 1,
            flexDirection: 'row',
            alignItems: 'center',
            justifyContent: 'center',
            gap: spacing.sm,
            ...shadows.md,
          }}
        >
          {showGlow && (
            <Animated.View
              style={{
                position: 'absolute',
                top: -4,
                left: -4,
                right: -4,
                bottom: -4,
                backgroundColor: getGlowColor(),
                borderRadius: borderRadius.lg,
                opacity: glowOpacity,
                zIndex: -1,
              }}
            />
          )}
          {loading ? (
            <LoadingSpinner size="sm" color={colors.textInverse} />
          ) : (
            <>
              {icon && icon}
              <Text
                style={{
                  fontSize: size === 'sm' ? scale(13) : size === 'lg' ? scale(18) : scale(16),
                  fontWeight: '700',
                  color: variant === 'outline' ? colors.primary : colors.textInverse,
                  letterSpacing: 0.5,
                }}
              >
                {title}
              </Text>
            </>
          )}
        </View>
      </TouchableOpacity>
    </Animated.View>
  );
};

// ============================================================================
// 5. ANIMATED SUCCESS/ERROR MESSAGE
// ============================================================================

interface MessageProps {
  message: string;
  visible?: boolean;
  type?: 'success' | 'error' | 'warning' | 'info';
}

export const AnimatedMessage: React.FC<MessageProps> = ({
  message,
  visible = true,
  type = 'success',
}) => {
  const slideAnim = React.useRef(new Animated.Value(-100)).current;
  const fadeAnim = React.useRef(new Animated.Value(0)).current;

  React.useEffect(() => {
    if (visible) {
      Animated.parallel([
        Animated.spring(slideAnim, {
          toValue: 0,
          useNativeDriver: true,
        }),
        Animated.timing(fadeAnim, {
          toValue: 1,
          duration: 300,
          useNativeDriver: true,
        }),
      ]).start();
    } else {
      Animated.parallel([
        Animated.spring(slideAnim, {
          toValue: -100,
          useNativeDriver: true,
        }),
        Animated.timing(fadeAnim, {
          toValue: 0,
          duration: 200,
          useNativeDriver: true,
        }),
      ]).start();
    }
  }, [visible]);

  const getBackgroundColor = () => {
    switch (type) {
      case 'success':
        return colors.successLight;
      case 'error':
        return colors.errorLight;
      case 'warning':
        return colors.warningLight;
      default:
        return colors.primaryLight;
    }
  };

  const getTextColor = () => {
    switch (type) {
      case 'success':
        return colors.success;
      case 'error':
        return colors.error;
      case 'warning':
        return colors.warning;
      default:
        return colors.primary;
    }
  };

  return (
    <Animated.View
      style={{
        transform: [{ translateY: slideAnim }],
        opacity: fadeAnim,
        margin: spacing.md,
        padding: spacing.md,
        backgroundColor: getBackgroundColor(),
        borderRadius: borderRadius.lg,
        borderLeftWidth: 4,
        borderLeftColor: getTextColor(),
      }}
    >
      <Text style={{ color: getTextColor(), fontWeight: '600', fontSize: scale(14) }}>
        {message}
      </Text>
    </Animated.View>
  );
};

// ============================================================================
// 6. DIVIDER COMPONENT
// ============================================================================

interface DividerProps {
  variant?: 'horizontal' | 'vertical';
  spacing?: number;
  color?: string;
}

export const Divider: React.FC<DividerProps> = ({
  variant = 'horizontal',
  spacing: sp = spacing.md,
  color = colors.border,
}) => {
  return (
    <View
      style={
        variant === 'horizontal'
          ? {
              height: 1,
              backgroundColor: color,
              marginVertical: sp,
            }
          : {
              width: 1,
              backgroundColor: color,
              marginHorizontal: sp,
            }
      }
    />
  );
};

// ============================================================================
// 7. BADGE COMPONENT
// ============================================================================

interface BadgeProps {
  label: string;
  variant?: 'primary' | 'secondary' | 'success' | 'error' | 'warning';
  size?: 'sm' | 'md';
}

export const Badge: React.FC<BadgeProps> = ({
  label,
  variant = 'primary',
  size = 'sm',
}) => {
  const getStyles = () => {
    const baseStyles = {
      paddingHorizontal: size === 'sm' ? spacing.sm : spacing.md,
      paddingVertical: size === 'sm' ? 4 : 6,
      borderRadius: borderRadius.full,
      alignItems: 'center',
      justifyContent: 'center',
    };

    switch (variant) {
      case 'success':
        return { ...baseStyles, backgroundColor: colors.successLight };
      case 'error':
        return { ...baseStyles, backgroundColor: colors.errorLight };
      case 'warning':
        return { ...baseStyles, backgroundColor: colors.warningLight };
      case 'secondary':
        return { ...baseStyles, backgroundColor: colors.secondaryLight };
      default:
        return { ...baseStyles, backgroundColor: colors.primaryLight };
    }
  };

  const getTextColor = () => {
    switch (variant) {
      case 'success':
        return colors.success;
      case 'error':
        return colors.error;
      case 'warning':
        return colors.warning;
      case 'secondary':
        return colors.secondary;
      default:
        return colors.primary;
    }
  };

  return (
    <View style={getStyles()}>
      <Text
        style={{
          fontSize: size === 'sm' ? scale(12) : scale(14),
          fontWeight: '600',
          color: getTextColor(),
        }}
      >
        {label}
      </Text>
    </View>
  );
};
