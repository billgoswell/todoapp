/**
 * Secure storage wrapper
 *
 * Provides a consistent interface for storing and retrieving app data
 * Currently uses AsyncStorage (consider Keychain/Keystore for production)
 */

import AsyncStorage from '@react-native-async-storage/async-storage';

const STORAGE_PREFIX = '@commandlinetodo_';

/**
 * Store a string value
 */
export const setString = async (key: string, value: string): Promise<void> => {
  try {
    await AsyncStorage.setItem(`${STORAGE_PREFIX}${key}`, value);
  } catch (error) {
    console.error(`Error setting string for key ${key}:`, error);
    throw error;
  }
};

/**
 * Retrieve a string value
 */
export const getString = async (key: string): Promise<string | null> => {
  try {
    return await AsyncStorage.getItem(`${STORAGE_PREFIX}${key}`);
  } catch (error) {
    console.error(`Error getting string for key ${key}:`, error);
    return null;
  }
};

/**
 * Store a JSON object
 */
export const setJSON = async (key: string, value: any): Promise<void> => {
  try {
    const jsonString = JSON.stringify(value);
    await AsyncStorage.setItem(`${STORAGE_PREFIX}${key}`, jsonString);
  } catch (error) {
    console.error(`Error setting JSON for key ${key}:`, error);
    throw error;
  }
};

/**
 * Retrieve a JSON object
 */
export const getJSON = async (key: string): Promise<any | null> => {
  try {
    const jsonString = await AsyncStorage.getItem(`${STORAGE_PREFIX}${key}`);
    if (jsonString === null) return null;
    return JSON.parse(jsonString);
  } catch (error) {
    console.error(`Error getting JSON for key ${key}:`, error);
    return null;
  }
};

/**
 * Store a number value
 */
export const setNumber = async (key: string, value: number): Promise<void> => {
  try {
    await AsyncStorage.setItem(`${STORAGE_PREFIX}${key}`, value.toString());
  } catch (error) {
    console.error(`Error setting number for key ${key}:`, error);
    throw error;
  }
};

/**
 * Retrieve a number value
 */
export const getNumber = async (key: string): Promise<number | null> => {
  try {
    const value = await AsyncStorage.getItem(`${STORAGE_PREFIX}${key}`);
    if (value === null) return null;
    return parseInt(value, 10);
  } catch (error) {
    console.error(`Error getting number for key ${key}:`, error);
    return null;
  }
};

/**
 * Store a boolean value
 */
export const setBoolean = async (key: string, value: boolean): Promise<void> => {
  try {
    await AsyncStorage.setItem(`${STORAGE_PREFIX}${key}`, value ? 'true' : 'false');
  } catch (error) {
    console.error(`Error setting boolean for key ${key}:`, error);
    throw error;
  }
};

/**
 * Retrieve a boolean value
 */
export const getBoolean = async (key: string): Promise<boolean | null> => {
  try {
    const value = await AsyncStorage.getItem(`${STORAGE_PREFIX}${key}`);
    if (value === null) return null;
    return value === 'true';
  } catch (error) {
    console.error(`Error getting boolean for key ${key}:`, error);
    return null;
  }
};

/**
 * Remove a value from storage
 */
export const remove = async (key: string): Promise<void> => {
  try {
    await AsyncStorage.removeItem(`${STORAGE_PREFIX}${key}`);
  } catch (error) {
    console.error(`Error removing key ${key}:`, error);
    throw error;
  }
};

/**
 * Clear all app storage
 */
export const clear = async (): Promise<void> => {
  try {
    const allKeys = await AsyncStorage.getAllKeys();
    const appKeys = allKeys.filter(key => key.startsWith(STORAGE_PREFIX));
    await AsyncStorage.multiRemove(appKeys);
  } catch (error) {
    console.error('Error clearing storage:', error);
    throw error;
  }
};

/**
 * API Key management
 */
export const setApiKey = async (apiKey: string): Promise<void> => {
  await setString('api_key', apiKey);
};

export const getApiKey = async (): Promise<string | null> => {
  return getString('api_key');
};

export const removeApiKey = async (): Promise<void> => {
  await remove('api_key');
};

/**
 * Server URL management
 */
export const setServerUrl = async (url: string): Promise<void> => {
  await setString('server_url', url);
};

export const getServerUrl = async (): Promise<string | null> => {
  return getString('server_url');
};

/**
 * Last sync time management
 */
export const setLastSyncTime = async (timestamp: number): Promise<void> => {
  await setNumber('last_sync_time', timestamp);
};

export const getLastSyncTime = async (): Promise<number> => {
  const value = await getNumber('last_sync_time');
  return value || 0;
};
