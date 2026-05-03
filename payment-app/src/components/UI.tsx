import React from 'react';
import { View, Text, TouchableOpacity, ActivityIndicator, Animated, Dimensions } from 'react-native';
import { colors } from '../styles/colors';
import { borderRadius, spacing, shadows, scale, typography } from '../styles/theme';

interface ButtonProps {
  onPress: () => void;
  title: string;
  variant?: 'primary' | 'secondary' | 'accent' | 'outline' | 'danger' | 'success';
  size?: 'sm' | 'md' | 'lg';
  loading?: boolean;
  disabled?: boolean;
  fullWidth?: boolean;
  icon?: React.ReactNode;
}

export const Button: React.FC<ButtonProps> = ({
  onPress,
  title,
  variant = 'primary',
  size = 'md',
  loading = false,
  disabled = false,
  fullWidth = false,
  icon,
}) => {
  const [pressed, setPressed] = React.useState(false);

  const getBackgroundColor = () => {
    if (disabled) return colors.borderDark;
    switch (variant) {
      case 'primary':
        return pressed ? colors.primaryDark : colors.primary;
      case 'secondary':
        return pressed ? colors.secondaryDark : colors.secondary;
      case 'accent':
        return pressed ? colors.accentDark : colors.accent;
      case 'danger':
        return pressed ? colors.error : colors.error;
      case 'success':
        return pressed ? colors.success : colors.successLight;
      case 'outline':
        return 'transparent';
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
        return { horizontal: spacing.md, vertical: spacing.xs };
      case 'md':
        return { horizontal: spacing.lg, vertical: spacing.sm };
      case 'lg':
        return { horizontal: spacing.xl, vertical: spacing.md };
      default:
        return { horizontal: spacing.lg, vertical: spacing.sm };
    }
  };

  const getFontSize = () => {
    switch (size) {
      case 'sm':
        return scale(13);
      case 'md':
        return scale(16);
      case 'lg':
        return scale(18);
      default:
        return scale(16);
    }
  };

  const padding = getPadding();

  return (
    <TouchableOpacity
      onPress={onPress}
      onPressIn={() => setPressed(true)}
      onPressOut={() => setPressed(false)}
      disabled={disabled || loading}
      activeOpacity={0.7}
      style={{
        backgroundColor: getBackgroundColor(),
        paddingHorizontal: padding.horizontal,
        paddingVertical: padding.vertical,
        borderRadius: borderRadius.lg,
        borderWidth: variant === 'outline' ? 2 : 0,
        borderColor: variant === 'outline' ? colors.primary : undefined,
        width: fullWidth ? '100%' : 'auto',
        opacity: disabled ? 0.5 : 1,
        transform: [{ scale: pressed && !disabled ? 0.95 : 1 }],
        ...shadows.md,
      }}
    >
      <View style={{ flexDirection: 'row', justifyContent: 'center', alignItems: 'center', gap: spacing.sm }}>
        {loading && <ActivityIndicator color={getTextColor()} size="small" />}
        {icon && !loading && <View>{icon}</View>}
        <Text
          style={{
            color: getTextColor(),
            fontSize: getFontSize(),
            fontWeight: '600',
            textAlign: 'center',
            letterSpacing: 0.3,
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
  shadow?: 'sm' | 'md' | 'lg' | 'xl';
  onPress?: () => void;
  highlight?: boolean;
  borderColor?: string;
  gradient?: boolean;
}

export const Card: React.FC<CardProps> = ({ 
  children, 
  padding = spacing.lg, 
  borderRadius: radius = borderRadius.xl,
  shadow = 'md',
  onPress,
  highlight = false,
  borderColor,
  gradient = false,
}) => {
  const [pressed, setPressed] = React.useState(false);

  const shadowStyle = shadows[shadow as keyof typeof shadows] || shadows.md;

  const CardContent = (
    <View
      style={{
        backgroundColor: gradient ? colors.surfaceDark : colors.surfaceCard,
        padding,
        borderRadius: radius,
        ...shadowStyle,
        borderWidth: borderColor || highlight ? 2 : 0,
        borderColor: borderColor || (highlight ? colors.primary : 'transparent'),
        transform: [{ scale: pressed && onPress ? 0.98 : 1 }],
      }}
    >
      {children}
    </View>
  );

  if (onPress) {
    return (
      <TouchableOpacity
        onPress={onPress}
        onPressIn={() => setPressed(true)}
        onPressOut={() => setPressed(false)}
        activeOpacity={0.8}
      >
        {CardContent}
      </TouchableOpacity>
    );
  }

  return CardContent;
};

interface BadgeProps {
  label: string;
  variant?: 'success' | 'error' | 'warning' | 'info' | 'primary' | 'secondary';
  size?: 'sm' | 'md' | 'lg';
}

export const Badge: React.FC<BadgeProps> = ({ label, variant = 'info', size = 'md' }) => {
  const getColors = () => {
    switch (variant) {
      case 'success':
        return { bg: colors.successLight, text: colors.textInverse };
      case 'error':
        return { bg: colors.errorLight, text: colors.textInverse };
      case 'warning':
        return { bg: colors.warningLight, text: colors.text };
      case 'info':
        return { bg: colors.infoLight, text: colors.textInverse };
      case 'primary':
        return { bg: colors.primaryLight, text: colors.textInverse };
      case 'secondary':
        return { bg: colors.secondaryLight, text: colors.textInverse };
      default:
        return { bg: colors.infoLight, text: colors.textInverse };
    }
  };

  const getPadding = () => {
    switch (size) {
      case 'sm':
        return { h: spacing.xs, v: spacing.xs / 2 };
      case 'md':
        return { h: spacing.md, v: spacing.xs };
      case 'lg':
        return { h: spacing.lg, v: spacing.sm };
      default:
        return { h: spacing.md, v: spacing.xs };
    }
  };

  const { bg, text } = getColors();
  const pad = getPadding();

  return (
    <View
      style={{
        backgroundColor: bg,
        paddingHorizontal: pad.h,
        paddingVertical: pad.v,
        borderRadius: borderRadius.full,
        alignSelf: 'flex-start',
        ...shadows.sm,
      }}
    >
      <Text style={{ color: text, fontSize: scale(12), fontWeight: '600', letterSpacing: 0.2 }}>
        {label}
      </Text>
    </View>
  );
};

interface DividerProps {
  margin?: number;
  color?: string;
}

export const Divider: React.FC<DividerProps> = ({ margin = spacing.lg, color = colors.border }) => (
  <View
    style={{
      height: 1.5,
      backgroundColor: color,
      marginVertical: margin,
      borderRadius: 1,
    }}
  />
);

interface EmptyStateProps {
  icon: string;
  title: string;
  subtitle?: string;
  action?: React.ReactNode;
}

export const EmptyState: React.FC<EmptyStateProps> = ({ icon, title, subtitle, action }) => (
  <View style={{ alignItems: 'center', justifyContent: 'center', paddingVertical: spacing['3xl'], paddingHorizontal: spacing.lg }}>
    <Text style={{ fontSize: scale(48), marginBottom: spacing.lg, opacity: 0.7 }}>{icon}</Text>
    <Text style={{ fontSize: scale(20), fontWeight: '700', color: colors.text, marginBottom: spacing.md, textAlign: 'center' }}>
      {title}
    </Text>
    {subtitle && (
      <Text style={{ fontSize: scale(14), color: colors.textSecondary, textAlign: 'center', marginBottom: spacing.lg }}>
        {subtitle}
      </Text>
    )}
    {action && <View style={{ marginTop: spacing.lg }}>{action}</View>}
  </View>
);

interface SkeletonProps {
  width?: number | `${number}%` | 'auto';
  height?: number;
  borderRadius?: number;
  marginVertical?: number;
}

export const Skeleton: React.FC<SkeletonProps> = ({
  width = '100%',
  height = 20,
  borderRadius: radius = borderRadius.md,
  marginVertical = spacing.md,
}) => {
  const animatedValue = React.useRef(new Animated.Value(0.4)).current;

  React.useEffect(() => {
    const animation = Animated.loop(
      Animated.sequence([
        Animated.timing(animatedValue, { toValue: 0.8, duration: 1000, useNativeDriver: false }),
        Animated.timing(animatedValue, { toValue: 0.4, duration: 1000, useNativeDriver: false }),
      ])
    );
    animation.start();
    return () => animation.stop();
  }, [animatedValue]);

  return (
    <Animated.View
      style={{
        width,
        height,
        backgroundColor: colors.border,
        borderRadius: radius,
        marginVertical,
        opacity: animatedValue,
      }}
    />
  );
};

interface SectionHeaderProps {
  title: string;
  subtitle?: string;
  action?: React.ReactNode;
}

export const SectionHeader: React.FC<SectionHeaderProps> = ({ title, subtitle, action }) => (
  <View style={{ marginVertical: spacing.lg }}>
    <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
      <View style={{ flex: 1 }}>
        <Text style={{ fontSize: scale(22), fontWeight: '800', color: colors.text, letterSpacing: 0.3 }}>
          {title}
        </Text>
        {subtitle && (
          <Text style={{ fontSize: scale(13), color: colors.textSecondary, marginTop: spacing.xs, fontWeight: '500' }}>
            {subtitle}
          </Text>
        )}
      </View>
      {action && <View>{action}</View>}
    </View>
  </View>
);

interface StatBoxProps {
  label: string;
  value: string;
  icon?: string;
  color?: string;
}

export const StatBox: React.FC<StatBoxProps> = ({ label, value, icon, color = colors.primaryLight }) => (
  <Card shadow="md" padding={spacing.lg} borderRadius={borderRadius.xl}>
    <View style={{ alignItems: 'center', gap: spacing.md }}>
      {icon && <Text style={{ fontSize: scale(32) }}>{icon}</Text>}
      <Text style={{ fontSize: scale(13), color: colors.textSecondary, fontWeight: '600', letterSpacing: 0.2 }}>
        {label}
      </Text>
      <Text style={{ fontSize: scale(28), fontWeight: '800', color, letterSpacing: 0.1 }}>
        {value}
      </Text>
    </View>
  </Card>
);

interface ProgressBarProps {
  progress: number; // 0-100
  height?: number;
  color?: string;
  backgroundColor?: string;
}

export const ProgressBar: React.FC<ProgressBarProps> = ({
  progress,
  height = scale(6),
  color = colors.primary,
  backgroundColor = colors.border,
}) => (
  <View
    style={{
      height,
      backgroundColor,
      borderRadius: borderRadius.full,
      overflow: 'hidden',
      width: '100%',
    }}
  >
    <Animated.View
      style={{
        height: '100%',
        width: `${Math.min(progress, 100)}%`,
        backgroundColor: color,
        borderRadius: borderRadius.full,
      }}
    />
  </View>
);

interface LoaderProps {
  size?: 'sm' | 'md' | 'lg';
  color?: string;
}

export const Loader: React.FC<LoaderProps> = ({ size = 'md', color = colors.primary }) => {
  const getSize = () => {
    switch (size) {
      case 'sm':
        return 'small';
      case 'md':
        return 'large';
      case 'lg':
        return 'large';
      default:
        return 'large';
    }
  };

  return <ActivityIndicator size={getSize()} color={color} />;
};


