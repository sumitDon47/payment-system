import React from 'react';
import { View, Text, TouchableOpacity, ActivityIndicator } from 'react-native';
import { colors } from '../styles/colors';
import { borderRadius, spacing } from '../styles/theme';

interface ButtonProps {
  onPress: () => void;
  title: string;
  variant?: 'primary' | 'secondary' | 'outline' | 'danger';
  size?: 'sm' | 'md' | 'lg';
  loading?: boolean;
  disabled?: boolean;
  fullWidth?: boolean;
}

export const Button: React.FC<ButtonProps> = ({
  onPress,
  title,
  variant = 'primary',
  size = 'md',
  loading = false,
  disabled = false,
  fullWidth = false,
}) => {
  const getBackgroundColor = () => {
    if (disabled) return colors.border;
    switch (variant) {
      case 'primary':
        return colors.primary;
      case 'secondary':
        return colors.secondary;
      case 'outline':
        return 'transparent';
      case 'danger':
        return colors.error;
      default:
        return colors.primary;
    }
  };

  const getTextColor = () => {
    if (variant === 'outline') return colors.primary;
    return colors.textInverse;
  };

  const getPadding = () => {
    switch (size) {
      case 'sm':
        return spacing.sm;
      case 'md':
        return spacing.md;
      case 'lg':
        return spacing.lg;
      default:
        return spacing.md;
    }
  };

  return (
    <TouchableOpacity
      onPress={onPress}
      disabled={disabled || loading}
      style={{
        backgroundColor: getBackgroundColor(),
        paddingHorizontal: getPadding() + 4,
        paddingVertical: getPadding(),
        borderRadius: borderRadius.lg,
        borderWidth: variant === 'outline' ? 2 : 0,
        borderColor: variant === 'outline' ? colors.primary : undefined,
        width: fullWidth ? '100%' : 'auto',
        opacity: disabled ? 0.5 : 1,
      }}
    >
      <View style={{ flexDirection: 'row', justifyContent: 'center', alignItems: 'center', gap: spacing.sm }}>
        {loading && <ActivityIndicator color={getTextColor()} size="small" />}
        <Text
          style={{
            color: getTextColor(),
            fontSize: 16,
            fontWeight: '600',
            textAlign: 'center',
          }}
        >
          {title}
        </Text>
      </View>
    </TouchableOpacity>
  );
};

interface CardProps {
  children: React.ReactNode;
  padding?: number;
  borderRadius?: number;
}

export const Card: React.FC<CardProps> = ({ children, padding = spacing.lg, borderRadius: radius = borderRadius.xl }) => (
  <View
    style={{
      backgroundColor: colors.surface,
      padding,
      borderRadius: radius,
      borderColor: colors.border,
      borderWidth: 1,
    }}
  >
    {children}
  </View>
);

interface BadgeProps {
  label: string;
  variant?: 'success' | 'error' | 'warning' | 'info';
}

export const Badge: React.FC<BadgeProps> = ({ label, variant = 'info' }) => {
  const getColors = () => {
    switch (variant) {
      case 'success':
        return { bg: `${colors.success}20`, text: colors.success };
      case 'error':
        return { bg: `${colors.error}20`, text: colors.error };
      case 'warning':
        return { bg: `${colors.warning}20`, text: colors.warning };
      case 'info':
        return { bg: `${colors.info}20`, text: colors.info };
    }
  };

  const { bg, text } = getColors();

  return (
    <View
      style={{
        backgroundColor: bg,
        paddingHorizontal: spacing.sm,
        paddingVertical: spacing.xs,
        borderRadius: borderRadius.full,
        alignSelf: 'flex-start',
      }}
    >
      <Text style={{ color: text, fontSize: 12, fontWeight: '600' }}>{label}</Text>
    </View>
  );
};

interface DividerProps {
  margin?: number;
}

export const Divider: React.FC<DividerProps> = ({ margin = spacing.lg }) => (
  <View
    style={{
      height: 1,
      backgroundColor: colors.border,
      marginVertical: margin,
    }}
  />
);

interface EmptyStateProps {
  icon: string;
  title: string;
  subtitle?: string;
}

export const EmptyState: React.FC<EmptyStateProps> = ({ icon, title, subtitle }) => (
  <View style={{ alignItems: 'center', justifyContent: 'center', paddingVertical: spacing['3xl'] }}>
    <Text style={{ fontSize: 40, marginBottom: spacing.md }}>{icon}</Text>
    <Text style={{ fontSize: 18, fontWeight: '600', color: colors.text, marginBottom: spacing.sm }}>
      {title}
    </Text>
    {subtitle && (
      <Text style={{ fontSize: 14, color: colors.textSecondary, textAlign: 'center' }}>
        {subtitle}
      </Text>
    )}
  </View>
);

interface SkeletonProps {
  width?: number | string;
  height?: number;
  borderRadius?: number;
  marginVertical?: number;
}

export const Skeleton: React.FC<SkeletonProps> = ({
  width = '100%',
  height = 20,
  borderRadius: radius = borderRadius.md,
  marginVertical = spacing.md,
}) => (
  <View
    style={{
      width,
      height,
      backgroundColor: colors.border,
      borderRadius: radius,
      marginVertical,
    }}
  />
);
