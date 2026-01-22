/**
 * App Navigator
 *
 * Navigation setup for authenticated app screens.
 * Uses React Navigation Stack for screen management.
 */

import React from 'react';
import { createStackNavigator, StackScreenProps, TransitionPresets } from '@react-navigation/stack';
import { colors } from '../theme';
import { screenTransitionConfig } from '../utils/animations';

// Screens
import { HomeScreen } from '../screens/HomeScreen';
import { TaskDetailScreen } from '../screens/TaskDetailScreen';
import { ListManagementScreen } from '../screens/ListManagementScreen';
import { SettingsScreen } from '../screens/SettingsScreen';

export type RootStackParamList = {
  Home: undefined;
  TaskDetail: { taskId?: number; listId: number };
  ListManagement: undefined;
  Settings: undefined;
};

const Stack = createStackNavigator<RootStackParamList>();

interface AppNavigatorProps {
  isSignedIn: boolean;
}

export const AppNavigator: React.FC<AppNavigatorProps> = ({ isSignedIn }) => {
  return (
    <Stack.Navigator
      screenOptions={{
        headerShown: false,
        cardStyle: { backgroundColor: colors.background },
        animationEnabled: true,
        // Smooth card style with fade-in animation
        cardStyleInterpolator: ({ current }) => ({
          cardStyle: {
            opacity: current.progress.interpolate({
              inputRange: [0, 0.5, 1],
              outputRange: [0, 0.5, 1],
            }),
          },
        }),
        // Smooth gesture-based transitions
        gestureEnabled: true,
        gestureResponseDistance: 50,
      }}
    >
      <Stack.Screen
        name="Home"
        component={HomeScreen}
        options={{
          animationTypeForReplace: isSignedIn ? 'pop' : 'fade',
          cardStyleInterpolator: ({ current }) => ({
            cardStyle: {
              opacity: current.progress,
            },
          }),
        }}
      />
      <Stack.Screen
        name="TaskDetail"
        component={TaskDetailScreen}
        options={{
          // Slide from right transition
          cardStyleInterpolator: ({ current, layouts }) => ({
            cardStyle: {
              transform: [
                {
                  translateX: current.progress.interpolate({
                    inputRange: [0, 1],
                    outputRange: [layouts.screen.width, 0],
                  }),
                },
              ],
            },
          }),
          // Reverse animation for pop
          gestureResponseDistance: 50,
        }}
      />
      <Stack.Screen
        name="ListManagement"
        component={ListManagementScreen}
        options={{
          // Slide from right transition
          cardStyleInterpolator: ({ current, layouts }) => ({
            cardStyle: {
              transform: [
                {
                  translateX: current.progress.interpolate({
                    inputRange: [0, 1],
                    outputRange: [layouts.screen.width, 0],
                  }),
                },
              ],
            },
          }),
        }}
      />
      <Stack.Screen
        name="Settings"
        component={SettingsScreen}
        options={{
          // Slide from right transition
          cardStyleInterpolator: ({ current, layouts }) => ({
            cardStyle: {
              transform: [
                {
                  translateX: current.progress.interpolate({
                    inputRange: [0, 1],
                    outputRange: [layouts.screen.width, 0],
                  }),
                },
              ],
            },
          }),
        }}
      />
    </Stack.Navigator>
  );
};

export default AppNavigator;
