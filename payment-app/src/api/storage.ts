/**
 * Storage utility for cross-platform token management
 * Uses sessionStorage on web, expo-secure-store on native
 */

const isWeb = typeof window !== 'undefined';

// Import SecureStore only on native
let SecureStore: any = null;
if (!isWeb) {
  try {
    // This will only be imported on native platforms
    SecureStore = require('expo-secure-store');
  } catch (e) {
    console.warn('SecureStore not available');
  }
}

export const StorageUtil = {
  /**
   * Set a value in storage
   */
  setItem: async (key: string, value: string) => {
    try {
      if (isWeb) {
        // Web: use sessionStorage
        sessionStorage.setItem(key, value);
      } else if (SecureStore) {
        // Native: use expo-secure-store
        await SecureStore.setItemAsync(key, value);
      }
    } catch (error) {
      console.error(`Error saving ${key}:`, error);
    }
  },

  /**
   * Get a value from storage
   */
  getItem: async (key: string): Promise<string | null> => {
    try {
      if (isWeb) {
        // Web: use sessionStorage
        return sessionStorage.getItem(key);
      } else if (SecureStore) {
        // Native: use expo-secure-store
        return await SecureStore.getItemAsync(key);
      }
    } catch (error) {
      console.error(`Error reading ${key}:`, error);
    }
    return null;
  },

  /**
   * Remove a value from storage
   */
  removeItem: async (key: string) => {
    try {
      if (isWeb) {
        // Web: use sessionStorage
        sessionStorage.removeItem(key);
      } else if (SecureStore) {
        // Native: use expo-secure-store
        await SecureStore.deleteItemAsync(key);
      }
    } catch (error) {
      console.error(`Error removing ${key}:`, error);
    }
  },

  /**
   * Clear all storage
   */
  clear: async () => {
    try {
      if (isWeb) {
        sessionStorage.clear();
      } else if (SecureStore) {
        // Manually remove known keys
        const keysToRemove = ['jwt_token', 'user_id', 'user_name', 'user_email'];
        for (const key of keysToRemove) {
          await SecureStore.deleteItemAsync(key);
        }
      }
    } catch (error) {
      console.error('Error clearing storage:', error);
    }
  },
};
