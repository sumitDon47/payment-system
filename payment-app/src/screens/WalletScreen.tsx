import React, { useState, useEffect, useCallback } from 'react';
import {
  View,
  Text,
  ScrollView,
  TextInput,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
  RefreshControl,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { userAPI, paymentAPI } from '../api/services';
import { StorageUtil } from '../api/storage';
import { useNavigation } from '../navigation/NavigationContext';

interface Transaction {
  id: string;
  sender_id: string;
  receiver_id: string;
  amount: number;
  status: string;
  created_at: string;
  note?: string;
}

export default function WalletScreen() {
  const { navigate } = useNavigation();
  const [balance, setBalance] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [userName, setUserName] = useState('');
  const [userId, setUserId] = useState('');
  
  // Transfer form state
  const [showTransferForm, setShowTransferForm] = useState(false);
  const [transferLoading, setTransferLoading] = useState(false);
  const [receiverEmail, setReceiverEmail] = useState('');
  const [receiverId, setReceiverId] = useState('');
  const [transferAmount, setTransferAmount] = useState('');
  const [transferNote, setTransferNote] = useState('');
  const [transferMpin, setTransferMpin] = useState('');
  const [transferCurrency, setTransferCurrency] = useState('NPR');
  const [errors, setErrors] = useState<{
    receiverEmail?: string;
    transferAmount?: string;
    transferMpin?: string;
  }>({});

  // Fetch user data and balance
  const fetchWalletData = useCallback(async () => {
    try {
      console.log('💳 Fetching wallet data...');
      setLoading(true);
      // Get user name from storage
      const name = await StorageUtil.getItem('user_name');
      const id = await StorageUtil.getItem('user_id');
      console.log('👤 User:', { name, id });
      setUserName(name || 'User');
      setUserId(id || '');

      // Fetch wallet balance
      console.log('📊 Fetching balance...');
      const walletResponse = await userAPI.getWalletBalance();
      console.log('✅ Balance fetched:', walletResponse);
      setBalance(walletResponse.data?.balance || 0);
    } catch (error: any) {
      const errorMsg = error.response?.data?.error || error.message || 'Failed to fetch wallet data';
      console.error('❌ Wallet fetch error:', error);
      console.error('Error response:', error.response);
      window.alert('Error: ' + errorMsg);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchWalletData();
  }, [fetchWalletData]);

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await fetchWalletData();
    setRefreshing(false);
  }, [fetchWalletData]);

  // Lookup receiver by email
  const handleLookupReceiver = async (email: string) => {
    if (!email.trim() || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      setReceiverId('');
      return;
    }

    try {
      const response = await paymentAPI.lookupUserByEmail(email);
      console.log('✅ User found:', response);
      const userId = response.data?.id;
      if (userId) {
        setReceiverId(userId);
        setErrors({ ...errors, receiverEmail: undefined });
      }
    } catch (error: any) {
      console.error('❌ Lookup failed:', error);
      setReceiverId('');
      const errorMsg = error.response?.data?.error || 'User not found';
      setErrors({ ...errors, receiverEmail: errorMsg });
    }
  };

  // Validate transfer form
  const validateTransfer = (): boolean => {
    const newErrors: typeof errors = {};

    if (!receiverEmail.trim()) {
      newErrors.receiverEmail = 'Receiver email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(receiverEmail)) {
      newErrors.receiverEmail = 'Please enter a valid email';
    } else if (!receiverId) {
      newErrors.receiverEmail = 'Please verify receiver email first';
    }

    const amount = parseFloat(transferAmount);
    if (!transferAmount.trim()) {
      newErrors.transferAmount = 'Amount is required';
    } else if (isNaN(amount) || amount <= 0) {
      newErrors.transferAmount = 'Amount must be greater than 0';
    } else if (amount > (balance || 0)) {
      newErrors.transferAmount = `Insufficient balance. Available: ${balance?.toFixed(2)}`;
    }

    if (!/^\d{4}$/.test(transferMpin)) {
      newErrors.transferMpin = 'Enter your 4-digit MPIN to confirm transfer';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleTransfer = async () => {
    if (!validateTransfer()) {
      return;
    }

    setTransferLoading(true);
    try {
      const response = await paymentAPI.sendPayment(
        receiverId,
        parseFloat(transferAmount),
        transferCurrency,
        transferNote,
        transferMpin
      );

      window.alert(`Success! Transfer of ${transferAmount} ${transferCurrency} completed!`);
      // Reset form and refresh balance
      setShowTransferForm(false);
      setReceiverEmail('');
      setReceiverId('');
      setTransferAmount('');
      setTransferNote('');
      setTransferMpin('');
      setErrors({});
      fetchWalletData();
    } catch (error: any) {
      const errorMsg = error.response?.data?.error || error.message || 'Transfer failed';
      console.error('Transfer error:', error);
      window.alert('Error: ' + errorMsg);
    } finally {
      setTransferLoading(false);
    }
  };

  const handleLogout = async () => {
    if (window.confirm('Are you sure you want to logout?')) {
      await StorageUtil.removeItem('jwt_token');
      await StorageUtil.removeItem('user_id');
      await StorageUtil.removeItem('user_name');
      await StorageUtil.removeItem('user_email');
      navigate('login');
    }
  };

  if (loading) {
    return (
      <View className="flex-1 justify-center items-center bg-gray-50">
        <ActivityIndicator size="large" color="#2563eb" />
      </View>
    );
  }

  return (
    <KeyboardAvoidingView
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      className="flex-1 bg-gray-50"
    >
      <ScrollView
        className="flex-1"
        refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} />}
      >
        {/* Header */}
        <View className="bg-gradient-to-br from-blue-600 to-blue-800 pt-10 pb-8 px-6">
          <Text className="text-white text-lg font-semibold">Welcome back,</Text>
          <Text className="text-white text-2xl font-bold mt-1">{userName}</Text>
        </View>

        {/* Balance Card */}
        <View className="mx-6 mt-8 bg-white rounded-2xl shadow-lg p-8 border border-gray-100">
          <Text className="text-gray-600 text-sm font-medium">Total Balance</Text>
          <View className="flex-row items-baseline mt-2">
            <Text className="text-4xl font-bold text-blue-600">
              {balance?.toFixed(2)}
            </Text>
            <Text className="text-gray-600 text-lg ml-2">{transferCurrency}</Text>
          </View>
          <Text className="text-gray-500 text-xs mt-4">Wallet ID: {userId?.slice(0, 8)}...</Text>
        </View>

        {/* Action Buttons */}
        <View className="flex-row gap-4 mx-6 mt-8">
          <TouchableOpacity
            onPress={() => setShowTransferForm(!showTransferForm)}
            className="flex-1 bg-blue-600 rounded-xl py-4 items-center shadow-md"
          >
            <Text className="text-white font-bold text-lg">Send Money</Text>
          </TouchableOpacity>
          <TouchableOpacity
            onPress={onRefresh}
            className="flex-1 bg-gray-200 rounded-xl py-4 items-center shadow-md"
          >
            <Text className="text-gray-800 font-bold text-lg">Refresh</Text>
          </TouchableOpacity>
        </View>

        {/* Transfer Form */}
        {showTransferForm && (
          <View className="mx-6 mt-8 bg-white rounded-2xl shadow-lg p-6 border border-gray-100 mb-8">
            <View className="flex-row justify-between items-center mb-6">
              <Text className="text-lg font-bold text-gray-800">Send Money</Text>
              <TouchableOpacity onPress={() => setShowTransferForm(false)}>
                <Text className="text-gray-500 text-2xl">×</Text>
              </TouchableOpacity>
            </View>

            {/* Receiver Email Input */}
            <View className="mb-5">
              <Text className="text-gray-700 font-semibold text-sm mb-2">Receiver Email</Text>
              <TextInput
                className={`border-2 rounded-lg px-4 py-3 text-gray-800 ${
                  errors.receiverEmail ? 'border-red-500' : 'border-gray-300'
                }`}
                placeholder="receiver@example.com"
                placeholderTextColor="#9ca3af"
                value={receiverEmail}
                onChangeText={(text) => {
                  setReceiverEmail(text);
                  if (errors.receiverEmail) {
                    setErrors({ ...errors, receiverEmail: undefined });
                  }
                }}
                onBlur={() => {
                  if (receiverEmail.trim()) {
                    handleLookupReceiver(receiverEmail);
                  }
                }}
                keyboardType="email-address"
                editable={!transferLoading}
              />
              {errors.receiverEmail && (
                <Text className="text-red-500 text-xs mt-1">{errors.receiverEmail}</Text>
              )}
            </View>

            {/* Amount Input */}
            <View className="mb-5">
              <Text className="text-gray-700 font-semibold text-sm mb-2">Amount</Text>
              <View className="flex-row items-center border-2 border-gray-300 rounded-lg overflow-hidden">
                <TextInput
                  className="flex-1 px-4 py-3 text-gray-800"
                  placeholder="0.00"
                  placeholderTextColor="#9ca3af"
                  value={transferAmount}
                  onChangeText={(text) => {
                    setTransferAmount(text);
                    if (errors.transferAmount) {
                      setErrors({ ...errors, transferAmount: undefined });
                    }
                  }}
                  keyboardType="decimal-pad"
                  editable={!transferLoading}
                />
                <View className="bg-gray-100 px-4 py-3">
                  <Text className="text-gray-700 font-semibold">{transferCurrency}</Text>
                </View>
              </View>
              {errors.transferAmount && (
                <Text className="text-red-500 text-xs mt-1">{errors.transferAmount}</Text>
              )}
              <Text className="text-gray-500 text-xs mt-2">
                Available: {balance?.toFixed(2)} {transferCurrency}
              </Text>
            </View>

            {/* Note Input */}
            <View className="mb-6">
              <Text className="text-gray-700 font-semibold text-sm mb-2">Note (Optional)</Text>
              <TextInput
                className="border-2 border-gray-300 rounded-lg px-4 py-3 text-gray-800 h-20"
                placeholder="Add a note for the receiver"
                placeholderTextColor="#9ca3af"
                value={transferNote}
                onChangeText={setTransferNote}
                multiline
                editable={!transferLoading}
              />
            </View>

            {/* MPIN Input */}
            <View className="mb-6">
              <Text className="text-gray-700 font-semibold text-sm mb-2">Confirm with MPIN</Text>
              <TextInput
                className={`border-2 rounded-lg px-4 py-3 text-gray-800 ${
                  errors.transferMpin ? 'border-red-500' : 'border-gray-300'
                }`}
                placeholder="Enter 4-digit MPIN"
                placeholderTextColor="#9ca3af"
                value={transferMpin}
                onChangeText={(text) => {
                  const digits = text.replace(/\D/g, '').slice(0, 4);
                  setTransferMpin(digits);
                  if (errors.transferMpin) {
                    setErrors({ ...errors, transferMpin: undefined });
                  }
                }}
                keyboardType="numeric"
                secureTextEntry
                editable={!transferLoading}
              />
              {errors.transferMpin && (
                <Text className="text-red-500 text-xs mt-1">{errors.transferMpin}</Text>
              )}
            </View>

            {/* Transfer Button */}
            <TouchableOpacity
              onPress={handleTransfer}
              disabled={transferLoading}
              className={`rounded-lg py-4 items-center ${
                transferLoading ? 'bg-blue-400' : 'bg-blue-600'
              }`}
            >
              {transferLoading ? (
                <ActivityIndicator color="white" />
              ) : (
                <Text className="text-white font-bold text-lg">Confirm Transfer</Text>
              )}
            </TouchableOpacity>
          </View>
        )}

        {/* Info Sections */}
        <View className="mx-6 mt-8 mb-8">
          {/* Features */}
          <View className="bg-white rounded-2xl shadow-lg p-6 border border-gray-100 mb-6">
            <Text className="text-lg font-bold text-gray-800 mb-4">Available Features</Text>
            <View className="space-y-3">
              <View className="flex-row items-start">
                <View className="w-6 h-6 rounded-full bg-green-100 justify-center items-center mr-3 mt-0.5">
                  <Text className="text-green-600 font-bold">✓</Text>
                </View>
                <View className="flex-1">
                  <Text className="text-gray-800 font-semibold">Instant Transfers</Text>
                  <Text className="text-gray-600 text-xs mt-1">Send money to other users instantly</Text>
                </View>
              </View>
              <View className="flex-row items-start">
                <View className="w-6 h-6 rounded-full bg-green-100 justify-center items-center mr-3 mt-0.5">
                  <Text className="text-green-600 font-bold">✓</Text>
                </View>
                <View className="flex-1">
                  <Text className="text-gray-800 font-semibold">Secure Transactions</Text>
                  <Text className="text-gray-600 text-xs mt-1">End-to-end encrypted and verified</Text>
                </View>
              </View>
              <View className="flex-row items-start">
                <View className="w-6 h-6 rounded-full bg-green-100 justify-center items-center mr-3 mt-0.5">
                  <Text className="text-green-600 font-bold">✓</Text>
                </View>
                <View className="flex-1">
                  <Text className="text-gray-800 font-semibold">Multi-Currency</Text>
                  <Text className="text-gray-600 text-xs mt-1">Support for multiple currency types</Text>
                </View>
              </View>
            </View>
          </View>

          {/* Help Section */}
          <View className="bg-blue-50 rounded-2xl border border-blue-200 p-6 mb-6">
            <Text className="text-blue-900 font-bold text-sm mb-2">💡 Tip</Text>
            <Text className="text-blue-800 text-xs">
              You can only transfer funds if you have sufficient balance. Make sure the receiver's email is correct before confirming the transfer.
            </Text>
          </View>
        </View>

        {/* Logout Button */}
        <TouchableOpacity
          onPress={handleLogout}
          className="mx-6 mb-8 bg-red-600 rounded-lg py-4 items-center"
        >
          <Text className="text-white font-bold">Logout</Text>
        </TouchableOpacity>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}
