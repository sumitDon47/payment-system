import React, { useState } from 'react';
import { View, Text, TextInput, TouchableOpacity } from 'react-native';
import { colors } from '../styles/colors';
import { borderRadius, spacing } from '../styles/theme';

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
}) => {
  const [showPassword, setShowPassword] = useState(!secureTextEntry);
  const [isFocused, setIsFocused] = useState(false);

  return (
    <View style={{ marginBottom: spacing.lg }}>
      {label && (
        <Text style={{ fontSize: 14, fontWeight: '600', color: colors.text, marginBottom: spacing.sm }}>
          {label}
        </Text>
      )}
      <View
        style={{
          flexDirection: 'row',
          alignItems: 'center',
          borderWidth: 2,
          borderColor: error ? colors.error : isFocused ? colors.primary : colors.border,
          borderRadius: borderRadius.lg,
          paddingHorizontal: spacing.md,
          backgroundColor: isFocused ? `${colors.primary}08` : colors.surface,
          transition: 'all 0.2s',
        }}
      >
        <TextInput
          placeholder={placeholder}
          value={value}
          onChangeText={onChangeText}
          secureTextEntry={secureTextEntry && !showPassword}
          keyboardType={keyboardType}
          onFocus={() => {
            setIsFocused(true);
            onFocus?.();
          }}
          onBlur={() => {
            setIsFocused(false);
            onBlur?.();
          }}
          style={{
            flex: 1,
            paddingVertical: spacing.md,
            fontSize: 16,
            color: colors.text,
          }}
          placeholderTextColor={colors.textTertiary}
        />
        {secureTextEntry && (
          <TouchableOpacity onPress={() => setShowPassword(!showPassword)}>
            <Text style={{ fontSize: 20, marginLeft: spacing.sm }}>
              {showPassword ? '👁️' : '👁️‍🗨️'}
            </Text>
          </TouchableOpacity>
        )}
      </View>
      {error && (
        <Text style={{ fontSize: 12, color: colors.error, marginTop: spacing.xs }}>
          ✕ {error}
        </Text>
      )}
      {helperText && !error && (
        <Text style={{ fontSize: 12, color: colors.textTertiary, marginTop: spacing.xs }}>
          {helperText}
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

export const Checkbox: React.FC<CheckboxProps> = ({ label, checked, onToggle }) => (
  <TouchableOpacity
    onPress={() => onToggle(!checked)}
    style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.md, marginBottom: spacing.md }}
  >
    <View
      style={{
        width: 20,
        height: 20,
        borderWidth: 2,
        borderColor: checked ? colors.primary : colors.border,
        borderRadius: borderRadius.md,
        backgroundColor: checked ? colors.primary : 'transparent',
        justifyContent: 'center',
        alignItems: 'center',
      }}
    >
      {checked && <Text style={{ color: colors.textInverse, fontWeight: 'bold' }}>✓</Text>}
    </View>
    <Text style={{ fontSize: 14, color: colors.text }}>{label}</Text>
  </TouchableOpacity>
);

interface FormErrorProps {
  message: string;
}

export const FormError: React.FC<FormErrorProps> = ({ message }) => (
  <View
    style={{
      backgroundColor: `${colors.error}15`,
      borderLeftWidth: 4,
      borderLeftColor: colors.error,
      padding: spacing.md,
      borderRadius: borderRadius.md,
      marginBottom: spacing.lg,
    }}
  >
    <Text style={{ color: colors.error, fontWeight: '600', fontSize: 14 }}>
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
      backgroundColor: `${colors.success}15`,
      borderLeftWidth: 4,
      borderLeftColor: colors.success,
      padding: spacing.md,
      borderRadius: borderRadius.md,
      marginBottom: spacing.lg,
    }}
  >
    <Text style={{ color: colors.success, fontWeight: '600', fontSize: 14 }}>
      ✓ {message}
    </Text>
  </View>
);
