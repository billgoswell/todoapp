/**
 * Splash Screen
 *
 * Displays while app initializes database and loads initial data.
 * Shows app branding and loading indicator.
 */

import React, { useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  Animated,
  ActivityIndicator,
  SafeAreaView,
} from 'react-native';
import { colors, spacing, typography } from '../theme';

interface SplashScreenProps {
  isVisible: boolean;
  onComplete?: () => void;
}

export const SplashScreen: React.FC<SplashScreenProps> = ({
  isVisible,
  onComplete,
}) => {
  const fadeAnim = React.useRef(new Animated.Value(0)).current;
  const scaleAnim = React.useRef(new Animated.Value(0.8)).current;
  const [hasFadedOut, setHasFadedOut] = React.useState(false);

  // Animation on mount
  useEffect(() => {
    if (isVisible) {
      setHasFadedOut(false);
      Animated.parallel([
        Animated.timing(fadeAnim, {
          toValue: 1,
          duration: 400,
          useNativeDriver: true,
        }),
        Animated.timing(scaleAnim, {
          toValue: 1,
          duration: 600,
          useNativeDriver: true,
        }),
      ]).start();
    } else {
      // Fade out when app loads
      Animated.timing(fadeAnim, {
        toValue: 0,
        duration: 300,
        useNativeDriver: true,
      }).start(() => {
        setHasFadedOut(true);
        onComplete?.();
      });
    }
  }, [isVisible]);

  if (!isVisible && hasFadedOut) {
    return null;
  }

  return (
    <Animated.View
      style={[
        styles.container,
        {
          opacity: fadeAnim,
        },
      ]}
      pointerEvents={isVisible ? 'auto' : 'none'}
    >
      <SafeAreaView style={styles.safeArea}>
        <View style={styles.content}>
          {/* App Icon */}
          <Animated.View
            style={[
              styles.iconContainer,
              {
                transform: [{ scale: scaleAnim }],
              },
            ]}
          >
            <View style={styles.icon}>
              <Text style={styles.iconEmoji}>✓</Text>
            </View>
          </Animated.View>

          {/* App Name */}
          <Text style={styles.appName}>CommandLineTodo</Text>
          <Text style={styles.tagline}>Your Tasks, Everywhere</Text>

          {/* Loading Indicator */}
          <View style={styles.loaderContainer}>
            <ActivityIndicator
              size="large"
              color={colors.primary}
              style={styles.loader}
            />
            <Text style={styles.loadingText}>Initializing...</Text>
          </View>

          {/* Footer */}
          <Text style={styles.version}>v1.0.0</Text>
        </View>
      </SafeAreaView>
    </Animated.View>
  );
};

const styles = StyleSheet.create({
  container: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: colors.background,
    zIndex: 1000,
    justifyContent: 'center',
    alignItems: 'center',
  },
  safeArea: {
    flex: 1,
    width: '100%',
  },
  content: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
  },
  iconContainer: {
    marginBottom: spacing.xl,
  },
  icon: {
    width: 120,
    height: 120,
    borderRadius: 30,
    backgroundColor: colors.primaryLight,
    justifyContent: 'center',
    alignItems: 'center',
    shadowColor: colors.primary,
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.2,
    shadowRadius: 16,
    elevation: 8,
  },
  iconEmoji: {
    fontSize: 60,
    fontWeight: '700',
  },
  appName: {
    ...typography.styles.h1,
    color: colors.text,
    marginBottom: spacing.sm,
    textAlign: 'center',
  },
  tagline: {
    ...typography.styles.body,
    color: colors.textSecondary,
    marginBottom: spacing.xl,
    textAlign: 'center',
    fontStyle: 'italic',
  },
  loaderContainer: {
    alignItems: 'center',
    marginTop: spacing.xl,
  },
  loader: {
    marginBottom: spacing.md,
  },
  loadingText: {
    ...typography.styles.bodySmall,
    color: colors.textSecondary,
  },
  version: {
    position: 'absolute',
    bottom: spacing.xl,
    ...typography.styles.caption,
    color: colors.textTertiary,
  },
});

export default SplashScreen;
