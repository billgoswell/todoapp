import 'react-native-gesture-handler';
import React, { useEffect, useState } from 'react';
import { View, ActivityIndicator, Text } from 'react-native';
import { StatusBar } from 'expo-status-bar';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { NavigationContainer } from '@react-navigation/native';
import AppNavigator from './navigation/AppNavigator';
import AppProvider from './state/AppProvider';
import { db } from './database';
import { apiClient } from './api';

/**
 * Root component for the CommandLineTodo mobile app
 *
 * Handles:
 * - Database initialization
 * - Loading stored configuration
 * - API client initialization
 * - State management setup via AppProvider
 * - Navigation setup
 */
export default function App() {
  const [isLoading, setIsLoading] = useState(true);
  const [apiKey, setApiKey] = useState<string | null>(null);
  const [initError, setInitError] = useState<string | null>(null);

  useEffect(() => {
    bootstrapAsync();
  }, []);

  const bootstrapAsync = async () => {
    try {
      // Initialize database
      console.log('Initializing database...');
      await db.initialize();

      // Initialize API client with stored credentials
      console.log('Initializing API client...');
      await apiClient.initialize();

      // Load stored API key
      const storedApiKey = await AsyncStorage.getItem('api_key');
      setApiKey(storedApiKey);

      console.log('App bootstrapped successfully');
    } catch (e) {
      console.error('Failed to bootstrap app:', e);
      setInitError(e instanceof Error ? e.message : 'Initialization failed');
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return (
      <GestureHandlerRootView style={{ flex: 1 }}>
        <SafeAreaProvider>
          <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}>
            <ActivityIndicator size="large" />
          </View>
          <StatusBar style="auto" />
        </SafeAreaProvider>
      </GestureHandlerRootView>
    );
  }

  if (initError) {
    return (
      <GestureHandlerRootView style={{ flex: 1 }}>
        <SafeAreaProvider>
          <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center', padding: 20 }}>
            <View style={{ backgroundColor: '#ffebee', padding: 16, borderRadius: 8 }}>
              <Text style={{ color: '#c62828', fontSize: 16, fontWeight: 'bold' }}>
                Initialization Error
              </Text>
              <Text style={{ color: '#c62828', marginTop: 8 }}>
                {initError}
              </Text>
            </View>
          </View>
          <StatusBar style="auto" />
        </SafeAreaProvider>
      </GestureHandlerRootView>
    );
  }

  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <SafeAreaProvider>
        <AppProvider>
          <NavigationContainer>
            <AppNavigator isSignedIn={!!apiKey} />
          </NavigationContainer>
        </AppProvider>
        <StatusBar style="auto" />
      </SafeAreaProvider>
    </GestureHandlerRootView>
  );
}
