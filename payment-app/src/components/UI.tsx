import React from 'react';
import { View, Text, TouchableOpacity, ActivityIndicator, Animated } from 'react-native';
import { colors } from '../styles/colors';
import { borderRadius, spacing, shadows } from '../styles/theme';

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

  const getGradientColors = () => {
    if (disabled) return [colors.border, colors.border];
    switch (variant) {
      case 'primary':
        return [colors.primaryLight, colors.primary];
      case 'secondary':
        return [colors.secondaryLight, colors.secondaryBright];
      case 'accent':
        return [colors.accentLight, colors.accent];
      case 'danger':
        return [colors.errorLight, colors.error];
      case 'success':
        return [colors.successLight, colors.success];
      default:
        return [colors.primaryLight, colors.primary];
    }
  };

  const getTextColor = () => {
    if (variant === 'outline') return colors.primaryLight;
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

  const gradientColors = getGradientColors();
  const backgroundColor = pressed ? gradientColors[1] : gradientColors[0];

  return (
    <TouchableOpacity
      onPress={onPress}
      onPressIn={() => setPressed(true)}
      onPressOut={() => setPressed(false)}
      disabled={disabled || loading}
      activeOpacity={0.8}
      style={{
        backgroundColor,
        paddingHorizontal: getPadding() + 8,
        paddingVertical: getPadding() + 2,
        borderRadius: borderRadius.lg,
        borderWidth: variant === 'outline' ? 2.5 : 0,
        borderColor: variant === 'outline' ? colors.primaryLight : undefined,
        width: fullWidth ? '100%' : 'auto',
        opacity: disabled ? 0.5 : 1,
        transform: [{ scale: pressed && !disabled ? 0.96 : 1 }],
        ...shadows.lg,
      }}
    >
      <View style={{ flexDirection: 'row', justifyContent: 'center', alignItems: 'center', gap: spacing.sm }}>
        {loading && <ActivityIndicator color={getTextColor()} size="small" />}
        {icon && !loading && <View>{icon}</View>}
        <Text
          style={{
            color: getTextColor(),
            fontSize: 16,
            fontWeight: '700',
            textAlign: 'center',
            letterSpacing: 0.4,
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
}

export const Card: React.FC<CardProps> = ({ 
  children, 
  padding = spacing.lg, 
  borderRadius: radius = borderRadius.xl,
  shadow = 'md',
  onPress,
  highlight = false,
  borderColor,
}) => {
  const [pressed, setPressed] = React.useState(false);

  const shadowStyle = shadows[shadow as keyof typeof shadows] || shadows.md;

  const CardContent = (
    <View
      style={{
        backgroundColor: colors.surfaceCard,
        padding,
        borderRadius: radius,
        ...shadowStyle,
        borderWidth: borderColor || highlight ? 2 : 0,
        borderColor: borderColor || (highlight ? colors.accentLight : 'transparent'),
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
        activeOpacity={0.85}
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
}

export const Badge: React.FC<BadgeProps> = ({ label, variant = 'info' }) => {
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

  const { bg, text } = getColors();

  return (
    <View
      style={{
        backgroundColor: bg,
        paddingHorizontal: spacing.md,
        paddingVertical: spacing.xs + 2,
        borderRadius: borderRadius.full,
        alignSelf: 'flex-start',
        ...shadows.sm,
      }}
    >
      <Text style={{ color: text, fontSize: 12, fontWeight: '700', letterSpacing: 0.3 }}>
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
      height: 2,
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
}

export const EmptyState: React.FC<EmptyStateProps> = ({ icon, title, subtitle }) => (
  <View style={{ alignItems: 'center', justifyContent: 'center', paddingVertical: spacing['3xl'] }}>
    <Text style={{ fontSize: 48, marginBottom: spacing.lg, opacity: 0.8 }}>{icon}</Text>
    <Text style={{ fontSize: 20, fontWeight: '700', color: colors.text, marginBottom: spacing.md, textAlign: 'center' }}>
      {title}
    </Text>
    {subtitle && (
      <Text style={{ fontSize: 14, color: colors.textSecondary, textAlign: 'center', paddingHorizontal: spacing.lg }}>
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
}) => {
  const animatedValue = React.useRef(new Animated.Value(0.5)).current;

  React.useEffect(() => {
    const animation = Animated.loop(
      Animated.sequence([
        Animated.timing(animatedValue, { toValue: 1, duration: 800, useNativeDriver: false }),
        Animated.timing(animatedValue, { toValue: 0.5, duration: 800, useNativeDriver: false }),
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
  action?: React.ReactNode;
}

export const SectionHeader: React.FC<SectionHeaderProps> = ({ title, action }) => (
  <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginVertical: spacing.lg }}>
    <Text style={{ fontSize: 20, fontWeight: '800', color: colors.text, letterSpacing: 0.5 }}>
      {title}
    </Text>
    {action}
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
      {icon && <Text style={{ fontSize: 32 }}>{icon}</Text>}
      <Text style={{ fontSize: 14, color: colors.textSecondary, fontWeight: '600', letterSpacing: 0.2 }}>
        {label}
      </Text>
      <Text style={{ fontSize: 28, fontWeight: '800', color, letterSpacing: 0.2 }}>
        {value}
      </Text>
    </View>
  </Card>
);

