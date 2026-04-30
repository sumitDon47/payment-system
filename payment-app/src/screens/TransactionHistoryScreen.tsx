import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  ScrollView,
  ActivityIndicator,
  TouchableOpacity,
} from 'react-native';
import { Card, Badge, EmptyState, Divider } from '../components/UI';
import { colors } from '../styles/colors';
import { spacing, borderRadius, shadows } from '../styles/theme';

interface Transaction {
  id: string;
  type: 'send' | 'receive';
  amount: number;
  recipient: string;
  date: string;
  status: 'completed' | 'pending' | 'failed';
  description?: string;
}

export default function TransactionHistoryScreen() {
  const [transactions, setTransactions] = useState<Transaction[]>([
    {
      id: '1',
      type: 'send',
      amount: 50,
      recipient: 'John Smith',
      date: '2024-04-28',
      status: 'completed',
      description: 'Dinner',
    },
    {
      id: '2',
      type: 'receive',
      amount: 100,
      recipient: 'Sarah Johnson',
      date: '2024-04-27',
      status: 'completed',
      description: 'Loan repayment',
    },
    {
      id: '3',
      type: 'send',
      amount: 25.50,
      recipient: 'Pizza Store',
      date: '2024-04-26',
      status: 'completed',
      description: 'Pizza order',
    },
  ]);
  const [loading, setLoading] = useState(false);
  const [filter, setFilter] = useState<'all' | 'send' | 'receive'>('all');

  const filteredTransactions = transactions.filter((t) => filter === 'all' || t.type === filter);

  const getTypeIcon = (type: 'send' | 'receive') => (type === 'send' ? '📤' : '📥');
  const getTypeLabel = (type: 'send' | 'receive') => (type === 'send' ? 'Sent' : 'Received');
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return colors.success;
      case 'pending':
        return colors.warning;
      case 'failed':
        return colors.error;
      default:
        return colors.textSecondary;
    }
  };

  return (
    <ScrollView style={{ flex: 1, backgroundColor: colors.background }} contentContainerStyle={{ padding: spacing.lg }}>
      {/* Header */}
      <View style={{ marginBottom: spacing.xl }}>
        <Text style={{ fontSize: 24, fontWeight: 'bold', color: colors.text, marginBottom: spacing.md }}>
          Transaction History
        </Text>
        <Text style={{ fontSize: 14, color: colors.textSecondary }}>
          All your transactions in one place
        </Text>
      </View>

      {/* Filter Tabs */}
      <View style={{ flexDirection: 'row', gap: spacing.md, marginBottom: spacing.xl }}>
        {(['all', 'send', 'receive'] as const).map((f) => (
          <TouchableOpacity
            key={f}
            onPress={() => setFilter(f)}
            style={{
              flex: 1,
              paddingVertical: spacing.md,
              paddingHorizontal: spacing.md,
              borderRadius: borderRadius.lg,
              backgroundColor: filter === f ? colors.primary : colors.surface,
              borderWidth: filter === f ? 0 : 1,
              borderColor: colors.border,
            }}
          >
            <Text
              style={{
                textAlign: 'center',
                fontWeight: '600',
                color: filter === f ? colors.textInverse : colors.text,
                fontSize: 12,
              }}
            >
              {f.charAt(0).toUpperCase() + f.slice(1)}
            </Text>
          </TouchableOpacity>
        ))}
      </View>

      {/* Transactions List */}
      {loading ? (
        <View style={{ justifyContent: 'center', alignItems: 'center', paddingVertical: spacing['3xl'] }}>
          <ActivityIndicator size="large" color={colors.primary} />
        </View>
      ) : filteredTransactions.length === 0 ? (
        <EmptyState
          icon="📭"
          title="No transactions"
          subtitle={`You haven't made any ${filter !== 'all' ? filter : ''} transactions yet`}
        />
      ) : (
        <View style={{ gap: spacing.md }}>
          {filteredTransactions.map((transaction, index) => (
            <Card key={transaction.id} padding={spacing.lg} borderRadius={borderRadius.lg}>
              <TouchableOpacity activeOpacity={0.7}>
                <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }}>
                  {/* Left Side - Icon & Details */}
                  <View style={{ flex: 1, flexDirection: 'row', gap: spacing.md, alignItems: 'center' }}>
                    <View
                      style={{
                        width: 48,
                        height: 48,
                        borderRadius: borderRadius.lg,
                        backgroundColor:
                          transaction.type === 'send'
                            ? `${colors.error}15`
                            : `${colors.success}15`,
                        justifyContent: 'center',
                        alignItems: 'center',
                      }}
                    >
                      <Text style={{ fontSize: 24 }}>{getTypeIcon(transaction.type)}</Text>
                    </View>

                    <View style={{ flex: 1 }}>
                      <View style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.sm, marginBottom: spacing.xs }}>
                        <Text style={{ fontSize: 14, fontWeight: '600', color: colors.text, flex: 1 }}>
                          {getTypeLabel(transaction.type)} to {transaction.recipient}
                        </Text>
                        <Badge
                          label={transaction.status.charAt(0).toUpperCase() + transaction.status.slice(1)}
                          variant={
                            transaction.status === 'completed'
                              ? 'success'
                              : transaction.status === 'pending'
                              ? 'warning'
                              : 'error'
                          }
                        />
                      </View>
                      <Text style={{ fontSize: 12, color: colors.textSecondary }}>
                        {transaction.date}
                        {transaction.description && ` • ${transaction.description}`}
                      </Text>
                    </View>
                  </View>

                  {/* Right Side - Amount */}
                  <View style={{ alignItems: 'flex-end' }}>
                    <Text
                      style={{
                        fontSize: 16,
                        fontWeight: 'bold',
                        color: transaction.type === 'send' ? colors.error : colors.success,
                      }}
                    >
                      {transaction.type === 'send' ? '-' : '+'}${transaction.amount.toFixed(2)}
                    </Text>
                  </View>
                </View>
              </TouchableOpacity>
            </Card>
          ))}
        </View>
      )}

      {/* Summary Section */}
      {filteredTransactions.length > 0 && (
        <>
          <Divider margin={spacing.xl} />

          <Card padding={spacing.lg}>
            <Text style={{ fontSize: 14, fontWeight: '600', color: colors.textSecondary, marginBottom: spacing.md }}>
              Summary
            </Text>
            <View style={{ gap: spacing.md }}>
              <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
                <Text style={{ color: colors.textSecondary }}>Total Sent</Text>
                <Text style={{ fontSize: 16, fontWeight: 'bold', color: colors.error }}>
                  $
                  {filteredTransactions
                    .filter((t) => t.type === 'send')
                    .reduce((sum, t) => sum + t.amount, 0)
                    .toFixed(2)}
                </Text>
              </View>
              <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
                <Text style={{ color: colors.textSecondary }}>Total Received</Text>
                <Text style={{ fontSize: 16, fontWeight: 'bold', color: colors.success }}>
                  $
                  {filteredTransactions
                    .filter((t) => t.type === 'receive')
                    .reduce((sum, t) => sum + t.amount, 0)
                    .toFixed(2)}
                </Text>
              </View>
            </View>
          </Card>
        </>
      )}

      <View style={{ height: spacing.xl }} />
    </ScrollView>
  );
}
