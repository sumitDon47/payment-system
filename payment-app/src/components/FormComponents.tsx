import React, { useState } from 'react';
import { View, Text, TextInput, TouchableOpacity, Animated, Dimensions } from 'react-native';
import { colors } from '../styles/colors';
import { borderRadius, spacing, shadows, scale, typography } from '../styles/theme';

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
  editable?: boolean;
  maxLength?: number;
  type?: 'text' | 'amount' | 'mpin';
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
  editable = true,
  maxLength,
  type = 'text',
}) => {
  const [showPassword, setShowPassword] = useState(!secureTextEntry);
  const [isFocused, setIsFocused] = useState(false);
  const animatedScale = React.useRef(new Animated.Value(1)).current;

  // Determine input type props based on field type
  const getInputConfig = () => {
    switch (type) {
      case 'amount':
        return {
          keyboardType: 'decimal-pad' as const,
          maxLength: 12,
          handleChange: (text: string) => {
            // Only allow numbers and one decimal point
            const filtered = text.replace(/[^0-9.]/g, '');
            const parts = filtered.split('.');
            const result = parts.length > 2 
              ? parts[0] + '.' + parts[1] 
              : filtered;
            onChangeText(result);
          },
        };
      case 'mpin':
        return {
          keyboardType: 'numeric' as const,
          maxLength: 4,
          handleChange: (text: string) => {
            const filtered = text.replace(/[^0-9]/g, '').slice(0, 4);
            onChangeText(filtered);
          },
        };
      default:
        return {
          keyboardType,
          maxLength,
          handleChange: onChangeText,
        };
    }
  };

  const config = getInputConfig();

  const handleFocus = () => {
    if (!editable) return;
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

  const borderColor = error ? colors.errorLight : isFocused ? colors.primary : colors.border;
  const bgColor = isFocused ? colors.surfaceCard : colors.surface;

  return (
    <View style={{ marginBottom: spacing.lg }}>
      {label && (
        <Text style={{ 
          fontSize: scale(14), 
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
            borderWidth: 2,
            borderColor,
            borderRadius: borderRadius.lg,
            paddingHorizontal: spacing.md,
            backgroundColor: bgColor,
            opacity: editable ? 1 : 0.6,
            ...shadows.md,
          }}
        >
          {icon && (
            <Text style={{ fontSize: scale(18), marginRight: spacing.sm }}>
              {icon}
            </Text>
          )}
          <TextInput
            placeholder={placeholder}
            value={value}
            onChangeText={config.handleChange}
            secureTextEntry={secureTextEntry && !showPassword}
            keyboardType={config.keyboardType}
            onFocus={handleFocus}
            onBlur={handleBlur}
            editable={editable}
            maxLength={config.maxLength}
            style={{
              flex: 1,
              paddingVertical: spacing.md,
              fontSize: scale(16),
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
              <Text style={{ fontSize: scale(20) }}>
                {showPassword ? '👁️' : '🔒'}
              </Text>
            </TouchableOpacity>
          )}
        </View>
      </Animated.View>
      {error && (
        <View style={{ marginTop: spacing.sm }}>
          <Text style={{ fontSize: scale(12), color: colors.errorLight, fontWeight: '600', letterSpacing: 0.2 }}>
            ✕ {error}
          </Text>
        </View>
      )}
      {helperText && !error && (
        <Text style={{ fontSize: scale(12), color: colors.textTertiary, marginTop: spacing.sm, fontWeight: '500' }}>
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
      Animated.timing(animatedScale, { toValue: 0.85, duration: 100, useNativeDriver: true }),
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
            width: scale(24),
            height: scale(24),
            borderWidth: 2,
            borderColor: checked ? colors.primary : colors.border,
            borderRadius: borderRadius.md,
            backgroundColor: checked ? colors.primary : 'transparent',
            justifyContent: 'center',
            alignItems: 'center',
            ...shadows.sm,
          }}
        >
          {checked && (
            <Text style={{ color: colors.textInverse, fontWeight: '800', fontSize: scale(14) }}>
              ✓
            </Text>
          )}
        </View>
      </Animated.View>
      <Text style={{ fontSize: scale(15), color: colors.text, fontWeight: '500', letterSpacing: 0.2 }}>
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
      borderLeftWidth: scale(5),
      borderLeftColor: colors.error,
      padding: spacing.md,
      borderRadius: borderRadius.lg,
      marginBottom: spacing.lg,
      ...shadows.md,
    }}
  >
    <Text style={{ color: colors.textInverse, fontWeight: '600', fontSize: scale(14), letterSpacing: 0.2 }}>
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
      borderLeftWidth: scale(5),
      borderLeftColor: colors.success,
      padding: spacing.md,
      borderRadius: borderRadius.lg,
      marginBottom: spacing.lg,
      ...shadows.md,
    }}
  >
    <Text style={{ color: colors.textInverse, fontWeight: '600', fontSize: scale(14), letterSpacing: 0.2 }}>
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
    <Text style={{ fontSize: scale(18), fontWeight: '800', color: colors.text, marginBottom: spacing.sm, letterSpacing: 0.3 }}>
      {title}
    </Text>
    {subtitle && (
      <Text style={{ fontSize: scale(14), color: colors.textSecondary, marginBottom: spacing.lg, fontWeight: '500' }}>
        {subtitle}
      </Text>
    )}
    {children}
  </View>
);

interface RadioProps {
  label: string;
  selected: boolean;
  onSelect: () => void;
}

export const Radio: React.FC<RadioProps> = ({ label, selected, onSelect }) => (
  <TouchableOpacity
    onPress={onSelect}
    activeOpacity={0.7}
    style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.md, marginBottom: spacing.md }}
  >
    <View
      style={{
        width: scale(22),
        height: scale(22),
        borderWidth: 2,
        borderColor: selected ? colors.primary : colors.border,
        borderRadius: borderRadius.full,
        justifyContent: 'center',
        alignItems: 'center',
        backgroundColor: selected ? colors.primary : 'transparent',
        ...shadows.sm,
      }}
    >
      {selected && (
        <View
          style={{
            width: scale(8),
            height: scale(8),
            borderRadius: borderRadius.full,
            backgroundColor: colors.textInverse,
          }}
        />
      )}
    </View>
    <Text style={{ fontSize: scale(15), color: colors.text, fontWeight: '500', letterSpacing: 0.1 }}>
      {label}
    </Text>
  </TouchableOpacity>
);

interface SelectOption {
  label: string;
  value: string;
}

interface SelectProps {
  label?: string;
  options: SelectOption[];
  selectedValue: string;
  onValueChange: (value: string) => void;
  placeholder?: string;
  error?: string;
}

export const Select: React.FC<SelectProps> = ({
  label,
  options,
  selectedValue,
  onValueChange,
  placeholder = 'Select an option',
  error,
}) => {
  const [isOpen, setIsOpen] = useState(false);

  const selectedLabel = options.find(opt => opt.value === selectedValue)?.label || placeholder;

  return (
    <View style={{ marginBottom: spacing.lg }}>
      {label && (
        <Text style={{ fontSize: scale(14), fontWeight: '700', color: colors.text, marginBottom: spacing.sm, letterSpacing: 0.2 }}>
          {label}
        </Text>
      )}
      <TouchableOpacity
        onPress={() => setIsOpen(!isOpen)}
        style={{
          borderWidth: 2,
          borderColor: error ? colors.errorLight : isOpen ? colors.primary : colors.border,
          borderRadius: borderRadius.lg,
          paddingHorizontal: spacing.md,
          paddingVertical: spacing.md,
          backgroundColor: colors.surfaceCard,
          flexDirection: 'row',
          justifyContent: 'space-between',
          alignItems: 'center',
          ...shadows.md,
        }}
      >
        <Text style={{ fontSize: scale(16), color: colors.text, fontWeight: '500' }}>
          {selectedLabel}
        </Text>
        <Text style={{ fontSize: scale(18) }}>
          {isOpen ? '▲' : '▼'}
        </Text>
      </TouchableOpacity>
      {isOpen && (
        <View
          style={{
            backgroundColor: colors.surfaceCard,
            borderWidth: 2,
            borderColor: colors.primary,
            borderTopWidth: 0,
            borderBottomLeftRadius: borderRadius.lg,
            borderBottomRightRadius: borderRadius.lg,
            marginTop: -2,
            ...shadows.lg,
          }}
        >
          {options.map((option) => (
            <TouchableOpacity
              key={option.value}
              onPress={() => {
                onValueChange(option.value);
                setIsOpen(false);
              }}
              style={{
                paddingHorizontal: spacing.md,
                paddingVertical: spacing.md,
                borderBottomWidth: option === options[options.length - 1] ? 0 : 1,
                borderBottomColor: colors.border,
                backgroundColor: selectedValue === option.value ? colors.surfaceDark : colors.surfaceCard,
              }}
            >
              <Text style={{ fontSize: scale(15), color: colors.text, fontWeight: '500' }}>
                {option.label}
              </Text>
            </TouchableOpacity>
          ))}
        </View>
      )}
      {error && (
        <Text style={{ fontSize: scale(12), color: colors.errorLight, fontWeight: '600', marginTop: spacing.sm, letterSpacing: 0.2 }}>
          ✕ {error}
        </Text>
      )}
    </View>
  );
};

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

