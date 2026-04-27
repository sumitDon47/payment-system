import { apiClient } from './axios';

/**
 * User Service API - All authentication and user-related endpoints
 */
export const userAPI = {
  /**
   * Register a new user
   * @param name - User's full name
   * @param email - User's email address
   * @param password - User's password (min 8 characters)
   * @returns { token: string, user: { id, name, email, balance, created_at, updated_at } }
   */
  register: async (name: string, email: string, password: string) => {
    const response = await apiClient.post('/register', {
      name,
      email,
      password,
    });
    return response.data;
  },

  /**
   * Login user and get JWT token
   * @param email - User's email address
   * @param password - User's password
   * @returns { token: string, user: { id, name, email, balance, created_at, updated_at } }
   */
  login: async (email: string, password: string) => {
    const response = await apiClient.post('/login', {
      email,
      password,
    });
    return response.data;
  },

  /**
   * Get authenticated user's profile
   * Requires JWT token in Authorization header
   * @returns { id, name, email, balance, created_at, updated_at }
   */
  getProfile: async () => {
    const response = await apiClient.get('/profile');
    return response.data;
  },

  /**
   * Get authenticated user's wallet balance
   * Requires JWT token in Authorization header
   * @returns { balance: number }
   */
  getWalletBalance: async () => {
    const response = await apiClient.get('/wallet');
    return response.data;
  },

  /**
   * Check API health status
   * @returns { status, service, redis }
   */
  healthCheck: async () => {
    const response = await apiClient.get('/health');
    return response.data;
  },

  /**
   * Request password reset token
   * @param email - User's email address
   * @returns { token: string (reset token) }
   */
  forgotPassword: async (email: string) => {
    const response = await apiClient.post('/forgot-password', {
      email,
    });
    return response.data;
  },

  /**
   * Reset password with reset token
   * @param token - Password reset token from forgotPassword
   * @param newPassword - New password (min 8 characters)
   * @returns { message: string }
   */
  resetPassword: async (token: string, newPassword: string) => {
    const response = await apiClient.post('/reset-password', {
      token,
      new_password: newPassword,
    });
    return response.data;
  },
};

/**
 * Payment Service API - gRPC payment operations (accessed via HTTP proxy if available)
 * Note: Currently gRPC-based, would need HTTP wrapper for direct frontend calls
 */
export const paymentAPI = {
  /**
   * Send payment from authenticated user to another user
   * Note: This would be called from backend for now
   * @param receiver_id - Recipient user ID
   * @param amount - Amount to transfer
   * @param currency - Currency code (e.g., 'USD', 'NPR')
   * @param note - Optional payment note
   */
  sendPayment: async (
    receiver_id: string,
    amount: number,
    currency: string = 'USD',
    note?: string
  ) => {
    try {
      // For now, this needs a backend proxy endpoint
      // In production, create a /transfer endpoint in user-service that calls payment-service
      const response = await apiClient.post('/transfer', {
        receiver_id,
        amount,
        currency,
        note,
      });
      return response.data;
    } catch (error) {
      throw error;
    }
  },

  /**
   * Get transaction details by ID
   * Note: This would need a backend proxy endpoint
   */
  getTransaction: async (transaction_id: string) => {
    try {
      const response = await apiClient.get(`/transactions/${transaction_id}`);
      return response.data;
    } catch (error) {
      throw error;
    }
  },

  /**
   * Get user's balance (more detailed than wallet endpoint)
   */
  getBalance: async (user_id?: string) => {
    try {
      const endpoint = user_id ? `/balance/${user_id}` : `/balance`;
      const response = await apiClient.get(endpoint);
      return response.data;
    } catch (error) {
      throw error;
    }
  },
};

export default { userAPI, paymentAPI };
