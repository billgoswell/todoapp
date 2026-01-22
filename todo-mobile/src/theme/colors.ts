/**
 * Color Theme
 *
 * Centralized color definitions for consistent UI styling
 */

export const colors = {
  // Primary colors
  primary: '#007AFF',
  primaryLight: '#E8F4FF',
  primaryDark: '#0051D5',

  // Secondary colors
  secondary: '#5856D6',
  accent: '#FF9500',

  // Status colors
  success: '#34C759',
  warning: '#FF9500',
  error: '#FF3B30',
  info: '#00C7FF',

  // Priority colors
  priorityHigh: '#FF3B30',     // Red
  priorityMedium: '#FF9500',   // Orange
  priorityLow: '#FFCC00',      // Yellow
  priorityNone: '#8E8E93',     // Gray

  // Grayscale
  black: '#000000',
  white: '#FFFFFF',
  gray900: '#1A1A1A',
  gray800: '#2C2C2C',
  gray700: '#3A3A3C',
  gray600: '#545458',
  gray500: '#8E8E93',
  gray400: '#A9A9AF',
  gray300: '#D1D1D6',
  gray200: '#E5E5EA',
  gray100: '#F2F2F7',
  gray50: '#FAFAFA',

  // Backgrounds
  background: '#FFFFFF',
  backgroundSecondary: '#F2F2F7',
  backgroundTertiary: '#E5E5EA',

  // Surfaces
  surface: '#FFFFFF',
  surfaceSecondary: '#F2F2F7',
  surfaceElevated: '#FFFFFF',

  // Text
  text: '#000000',
  textSecondary: '#3A3A3C',
  textTertiary: '#8E8E93',
  textInverse: '#FFFFFF',

  // Borders
  border: '#E5E5EA',
  borderLight: '#F2F2F7',
  borderDark: '#C7C7CC',

  // Semantic colors
  disabled: '#C7C7CC',
  placeholder: '#A9A9AF',

  // Status backgrounds
  successLight: '#E8F5E9',
  warningLight: '#FFF3E0',
  errorLight: '#FFEBEE',
  infoLight: '#E0F7FF',
} as const;

export type Color = typeof colors[keyof typeof colors];
