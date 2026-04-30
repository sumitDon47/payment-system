import React, { useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
} from 'react-native';
import { StorageUtil } from '../api/services';
import { userAPI } from '../api/services';
import { useNavigation } from '../navigation/NavigationContext';
import { Card, Button, Divider, Badge } from '../components/UI';
import { colors } from '../styles/colors';
import { spacing, borderRadius, shadows } from '../styles/theme';

export default function ProfileScreen() {
  const { navigate } = useNavigation();
  const [user, setUser] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [isLogoutLoading, setIsLogoutLoading] = useState(false);

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
