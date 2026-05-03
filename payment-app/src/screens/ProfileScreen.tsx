import React, { useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
} from 'react-native';
import { StorageUtil } from '../api/storage';
import { userAPI } from '../api/services';
import { useNavigation } from '../navigation/NavigationContext';
import { Card, Button, Divider, Badge } from '../components/UI';
import { Input, FormError, FormSuccess } from '../components/FormComponents';
import { colors } from '../styles/colors';
import { spacing, borderRadius, shadows } from '../styles/theme';

export default function ProfileScreen() {
  const { navigate } = useNavigation();
  const [user, setUser] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [isLogoutLoading, setIsLogoutLoading] = useState(false);
  const [mpin, setMpin] = useState('');
  const [confirmMpin, setConfirmMpin] = useState('');
  const [mpinLoading, setMpinLoading] = useState(false);
  const [mpinError, setMpinError] = useState('');
  const [mpinSuccess, setMpinSuccess] = useState('');

  React.useEffect(() => {
    loadUserData();
  }, []);

  const loadUserData = async () => {
    try {
      setLoading(true);
      const name = await StorageUtil.getItem('user_name');
      const email = await StorageUtil.getItem('user_email');
      const userId = await StorageUtil.getItem('user_id');
      
      setUser({
        name: name || 'User',
        email: email || 'loading...',
        id: userId,
        verified: true,
        joinedDate: '2024-01-15',
      });
    } catch (error) {
      console.error('Error loading user data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = async () => {
    Alert.alert(
      'Logout',
      'Are you sure you want to logout?',
      [
        { text: 'Cancel', onPress: () => {}, style: 'cancel' },
        {
          text: 'Logout',
          onPress: async () => {
            setIsLogoutLoading(true);
            try {
              await StorageUtil.removeItem('jwt_token');
              await StorageUtil.removeItem('user_id');
              await StorageUtil.removeItem('user_name');
              await StorageUtil.removeItem('user_email');
              navigate('login');
            } catch (error) {
              Alert.alert('Error', 'Failed to logout');
            } finally {
              setIsLogoutLoading(false);
            }
          },
          style: 'destructive',
        },
      ]
    );
  };

  const handleSetMpin = async () => {
    setMpinError('');
    setMpinSuccess('');

    if (!/^\d{4}$/.test(mpin)) {
      setMpinError('MPIN must be exactly 4 digits');
      return;
    }

    if (mpin !== confirmMpin) {
      setMpinError('MPIN values do not match');
      return;
    }

    try {
      setMpinLoading(true);
      await userAPI.setMPIN(mpin);
      setMpin('');
      setConfirmMpin('');
      setMpinSuccess('MPIN updated successfully. Use it for login and transfers.');
    } catch (error: any) {
      const message = error?.response?.data?.error || error?.message || 'Failed to update MPIN';
      setMpinError(message);
    } finally {
      setMpinLoading(false);
    }
  };

  if (loading) {
    return (
      <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: colors.background }}>
        <ActivityIndicator size="large" color={colors.primary} />
      </View>
    );
  }

  return (
    <ScrollView style={{ flex: 1, backgroundColor: colors.background }} contentContainerStyle={{ padding: spacing.lg }}>
      {/* Profile Header */}
      <Card padding={spacing['2xl']}>
        <View style={{ alignItems: 'center', marginBottom: spacing.xl }}>
          <View
            style={{
              width: 100,
              height: 100,
              borderRadius: borderRadius.full,
              backgroundColor: `${colors.primary}20`,
              justifyContent: 'center',
              alignItems: 'center',
              marginBottom: spacing.lg,
            }}
          >
            <Text style={{ fontSize: 48 }}>👤</Text>
          </View>
          <Text style={{ fontSize: 24, fontWeight: 'bold', color: colors.text, marginBottom: spacing.xs }}>
            {user?.name}
          </Text>
          <Text style={{ fontSize: 14, color: colors.textSecondary }}>{user?.email}</Text>
          <View style={{ flexDirection: 'row', gap: spacing.md, marginTop: spacing.md }}>
            <Badge label="Verified" variant="success" />
          </View>
        </View>
      </Card>

      <Divider margin={spacing.xl} />

      {/* Account Information */}
      <Text style={{ fontSize: 18, fontWeight: 'bold', color: colors.text, marginBottom: spacing.md }}>
        Account Information
      </Text>

      <Card padding={spacing.lg} borderRadius={borderRadius.lg}>
        <View style={{ gap: spacing.lg }}>
          <View>
            <Text style={{ fontSize: 12, color: colors.textTertiary, fontWeight: '600', marginBottom: spacing.xs }}>
              USER ID
            </Text>
            <Text style={{ fontSize: 14, color: colors.text, fontFamily: 'monospace' }}>
              {user?.id}
            </Text>
          </View>

          <View>
            <Text style={{ fontSize: 12, color: colors.textTertiary, fontWeight: '600', marginBottom: spacing.xs }}>
              JOINED
            </Text>
            <Text style={{ fontSize: 14, color: colors.text }}>
              January 15, 2024
            </Text>
          </View>

          <View>
            <Text style={{ fontSize: 12, color: colors.textTertiary, fontWeight: '600', marginBottom: spacing.xs }}>
              STATUS
            </Text>
            <Badge label="Active" variant="success" />
          </View>
        </View>
      </Card>

      <Divider margin={spacing.xl} />

      <Text style={{ fontSize: 18, fontWeight: 'bold', color: colors.text, marginBottom: spacing.md }}>
        Security MPIN
      </Text>

      <Card padding={spacing.lg} borderRadius={borderRadius.lg}>
        <Text style={{ fontSize: 13, color: colors.textSecondary, marginBottom: spacing.lg }}>
          Set a 4-digit MPIN for quick login and mandatory transfer confirmation.
        </Text>

        {mpinError ? <FormError message={mpinError} /> : null}
        {mpinSuccess ? <FormSuccess message={mpinSuccess} /> : null}

        <Input
          label="New MPIN"
          placeholder="Enter 4-digit MPIN"
          value={mpin}
          onChangeText={(text) => setMpin(text.replace(/\D/g, '').slice(0, 4))}
          keyboardType="numeric"
          secureTextEntry
          helperText="Example: 1234"
        />

        <Input
          label="Confirm MPIN"
          placeholder="Re-enter MPIN"
          value={confirmMpin}
          onChangeText={(text) => setConfirmMpin(text.replace(/\D/g, '').slice(0, 4))}
          keyboardType="numeric"
          secureTextEntry
        />

        <Button
          title={mpinLoading ? 'Updating MPIN...' : 'Set / Change MPIN'}
          onPress={handleSetMpin}
          loading={mpinLoading}
          disabled={mpinLoading}
          fullWidth
        />
      </Card>

      <Divider margin={spacing.xl} />

      {/* Quick Actions */}
      <Text style={{ fontSize: 18, fontWeight: 'bold', color: colors.text, marginBottom: spacing.md }}>
        Quick Actions
      </Text>

      <View style={{ gap: spacing.md }}>
        <TouchableOpacity
          style={{
            ...styles.actionButton,
            backgroundColor: `${colors.info}15`,
            borderLeftColor: colors.info,
          }}
        >
          <Text style={{ fontSize: 24 }}>🔐</Text>
          <View style={{ flex: 1 }}>
            <Text style={{ fontSize: 16, fontWeight: '600', color: colors.text }}>Change Password</Text>
            <Text style={{ fontSize: 12, color: colors.textSecondary }}>Update your security</Text>
          </View>
          <Text style={{ fontSize: 20 }}>›</Text>
        </TouchableOpacity>

        <TouchableOpacity
          style={{
            ...styles.actionButton,
            backgroundColor: `${colors.warning}15`,
            borderLeftColor: colors.warning,
          }}
        >
          <Text style={{ fontSize: 24 }}>🔔</Text>
          <View style={{ flex: 1 }}>
            <Text style={{ fontSize: 16, fontWeight: '600', color: colors.text }}>Notifications</Text>
            <Text style={{ fontSize: 12, color: colors.textSecondary }}>Manage preferences</Text>
          </View>
          <Text style={{ fontSize: 20 }}>›</Text>
        </TouchableOpacity>

        <TouchableOpacity
          style={{
            ...styles.actionButton,
            backgroundColor: `${colors.secondary}15`,
            borderLeftColor: colors.secondary,
          }}
        >
          <Text style={{ fontSize: 24 }}>📄</Text>
          <View style={{ flex: 1 }}>
            <Text style={{ fontSize: 16, fontWeight: '600', color: colors.text }}>Terms & Privacy</Text>
            <Text style={{ fontSize: 12, color: colors.textSecondary }}>View our policies</Text>
          </View>
          <Text style={{ fontSize: 20 }}>›</Text>
        </TouchableOpacity>
      </View>

      <Divider margin={spacing.xl} />

      {/* Logout Button */}
      <Button
        title={isLogoutLoading ? 'Signing out...' : 'Logout'}
        onPress={handleLogout}
        variant="danger"
        fullWidth
        loading={isLogoutLoading}
        disabled={isLogoutLoading}
        size="lg"
      />

      <View style={{ height: spacing.xl }} />
    </ScrollView>
  );
}

const styles = {
  actionButton: {
    flexDirection: 'row' as const,
    alignItems: 'center' as const,
    padding: spacing.lg,
    borderRadius: borderRadius.lg,
    borderLeftWidth: 4,
    gap: spacing.md,
  },
};
