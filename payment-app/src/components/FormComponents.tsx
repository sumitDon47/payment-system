import React, { useState } from 'react';
import { View, Text, TextInput, TouchableOpacity, Animated } from 'react-native';
import { colors } from '../styles/colors';
import { borderRadius, spacing, shadows } from '../styles/theme';

interface InputProps {
  label?: string;
  placeholder: string;
  value: string;
  onChangeText: (text: string) => void;
  secureTextEntry?: boolean;
  keyboardType?: 'default' | 'email-address' | 'numeric' | 'phone-pad';
  error?: string;
  helperText?: string;
  onFocus?: () => void;
  onBlur?: () => void;
  icon?: string;
}

export const Input: React.FC<InputProps> = ({
  label,
  placeholder,
  value,
  onChangeText,
  secureTextEntry = false,
  keyboardType = 'default',
  error,
  helperText,
  onFocus,
  onBlur,
  icon,
}) => {
  const [showPassword, setShowPassword] = useState(!secureTextEntry);
  const [isFocused, setIsFocused] = useState(false);
  const animatedScale = React.useRef(new Animated.Value(1)).current;

  const handleFocus = () => {
    setIsFocused(true);
    Animated.spring(animatedScale, {
      toValue: 1.02,
      useNativeDriver: true,
    }).start();
    onFocus?.();
  };

  const handleBlur = () => {
    setIsFocused(false);
    Animated.spring(animatedScale, {
      toValue: 1,
      useNativeDriver: true,
    }).start();
    onBlur?.();
  };

  const borderColor = error ? colors.errorLight : isFocused ? colors.primaryLight : colors.border;
  const bgColor = isFocused ? colors.primaryBright : colors.surface;

  return (
    <View style={{ marginBottom: spacing.lg }}>
      {label && (
        <Text style={{ 
          fontSize: 14, 
          fontWeight: '700', 
          color: colors.text, 
          marginBottom: spacing.sm,
          letterSpacing: 0.3,
        }}>
          {label}
        </Text>
      )}
      <Animated.View
        style={{
          transform: [{ scale: animatedScale }],
        }}
      >
        <View
          style={{
            flexDirection: 'row',
            alignItems: 'center',
            borderWidth: 2.5,
            borderColor,
            borderRadius: borderRadius.lg,
            paddingHorizontal: spacing.md,
            backgroundColor: bgColor,
            ...shadows.md,
          }}
        >
          {icon && (
            <Text style={{ fontSize: 18, marginRight: spacing.sm }}>
              {icon}
            </Text>
          )}
          <TextInput
            placeholder={placeholder}
            value={value}
            onChangeText={onChangeText}
            secureTextEntry={secureTextEntry && !showPassword}
            keyboardType={keyboardType}
            onFocus={handleFocus}
            onBlur={handleBlur}
            style={{
              flex: 1,
              paddingVertical: spacing.md + 2,
              fontSize: 16,
              color: colors.text,
              fontWeight: '500',
            }}
            placeholderTextColor={colors.textTertiary}
          />
          {secureTextEntry && (
            <TouchableOpacity 
              onPress={() => setShowPassword(!showPassword)}
              style={{ padding: spacing.sm }}
            >
              <Text style={{ fontSize: 20 }}>
                {showPassword ? '👁️' : '🔒'}
              </Text>
            </TouchableOpacity>
          )}
        </View>
      </Animated.View>
      {error && (
        <View style={{ marginTop: spacing.sm }}>
          <Text style={{ fontSize: 12, color: colors.errorLight, fontWeight: '700', letterSpacing: 0.2 }}>
            ✕ {error}
          </Text>
        </View>
      )}
      {helperText && !error && (
        <Text style={{ fontSize: 12, color: colors.textTertiary, marginTop: spacing.sm, fontWeight: '500' }}>
          ℹ️ {helperText}
        </Text>
      )}
    </View>
  );
};

interface CheckboxProps {
  label: string;
  checked: boolean;
  onToggle: (checked: boolean) => void;
}

export const Checkbox: React.FC<CheckboxProps> = ({ label, checked, onToggle }) => {
  const [pressed, setPressed] = useState(false);
  const animatedScale = React.useRef(new Animated.Value(1)).current;

  const handlePress = () => {
    Animated.sequence([
      Animated.timing(animatedScale, { toValue: 0.9, duration: 100, useNativeDriver: true }),
      Animated.timing(animatedScale, { toValue: 1, duration: 100, useNativeDriver: true }),
    ]).start();
    onToggle(!checked);
  };

  return (
    <TouchableOpacity
      onPress={handlePress}
      activeOpacity={0.7}
      style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.md, marginBottom: spacing.md }}
    >
      <Animated.View
        style={{
          transform: [{ scale: animatedScale }],
        }}
      >
        <View
          style={{
            width: 24,
            height: 24,
            borderWidth: 2.5,
            borderColor: checked ? colors.primaryLight : colors.border,
            borderRadius: borderRadius.md,
            backgroundColor: checked ? colors.primaryLight : 'transparent',
            justifyContent: 'center',
            alignItems: 'center',
            ...shadows.sm,
          }}
        >
          {checked && (
            <Text style={{ color: colors.textInverse, fontWeight: '800', fontSize: 14 }}>
              ✓
            </Text>
          )}
        </View>
      </Animated.View>
      <Text style={{ fontSize: 15, color: colors.text, fontWeight: '500', letterSpacing: 0.2 }}>
        {label}
      </Text>
    </TouchableOpacity>
  );
};

interface FormErrorProps {
  message: string;
}

export const FormError: React.FC<FormErrorProps> = ({ message }) => (
  <View
    style={{
      backgroundColor: colors.errorLight,
      borderLeftWidth: 5,
      borderLeftColor: colors.error,
      padding: spacing.md + 2,
      borderRadius: borderRadius.lg,
      marginBottom: spacing.lg,
      ...shadows.md,
    }}
  >
    <Text style={{ color: colors.textInverse, fontWeight: '700', fontSize: 14, letterSpacing: 0.3 }}>
      ⚠️ {message}
    </Text>
  </View>
);

interface FormSuccessProps {
  message: string;
}

export const FormSuccess: React.FC<FormSuccessProps> = ({ message }) => (
  <View
    style={{
      backgroundColor: colors.successLight,
      borderLeftWidth: 5,
      borderLeftColor: colors.success,
      padding: spacing.md + 2,
      borderRadius: borderRadius.lg,
      marginBottom: spacing.lg,
      ...shadows.md,
    }}
  >
    <Text style={{ color: colors.textInverse, fontWeight: '700', fontSize: 14, letterSpacing: 0.3 }}>
      ✓ {message}
    </Text>
  </View>
);

interface FormSectionProps {
  title: string;
  subtitle?: string;
  children: React.ReactNode;
}

export const FormSection: React.FC<FormSectionProps> = ({ title, subtitle, children }) => (
  <View style={{ marginBottom: spacing['2xl'] }}>
    <Text style={{ fontSize: 18, fontWeight: '800', color: colors.text, marginBottom: spacing.sm, letterSpacing: 0.5 }}>
      {title}
    </Text>
    {subtitle && (
      <Text style={{ fontSize: 14, color: colors.textSecondary, marginBottom: spacing.lg, fontWeight: '500' }}>
        {subtitle}
      </Text>
    )}
    {children}
  </View>
);

interface TextAreaProps {
  placeholder: string;
  value: string;
  onChangeText: (text: string) => void;
  label?: string;
  rows?: number;
  error?: string;
}

export const TextArea: React.FC<TextAreaProps> = ({ 
  placeholder, 
  value, 
  onChangeText, 
  label, 
  rows = 4,
  error,
}) => {
  const [isFocused, setIsFocused] = useState(false);

  return (
    <View style={{ marginBottom: spacing.lg }}>
      {label && (
        <Text style={{ fontSize: 14, fontWeight: '700', color: colors.text, marginBottom: spacing.sm, letterSpacing: 0.3 }}>
          {label}
        </Text>
      )}
      <TextInput
        placeholder={placeholder}
        value={value}
        onChangeText={onChangeText}
        multiline
        numberOfLines={rows}
        onFocus={() => setIsFocused(true)}
        onBlur={() => setIsFocused(false)}
        style={{
          borderWidth: 2.5,
          borderColor: error ? colors.errorLight : isFocused ? colors.primaryLight : colors.border,
          borderRadius: borderRadius.lg,
          paddingHorizontal: spacing.md,
          paddingVertical: spacing.md,
          fontSize: 16,
          color: colors.text,
          backgroundColor: isFocused ? colors.primaryBright : colors.surface,
          fontWeight: '500',
          textAlignVertical: 'top',
          ...shadows.md,
        }}
        placeholderTextColor={colors.textTertiary}
      />
      {error && (
        <Text style={{ fontSize: 12, color: colors.errorLight, marginTop: spacing.sm, fontWeight: '700' }}>
          ✕ {error}
        </Text>
      )}
    </View>
  );
};

