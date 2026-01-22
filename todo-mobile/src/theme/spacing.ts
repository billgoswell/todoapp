/**
 * Spacing & Sizing
 *
 * Standardized spacing scale for consistent layouts
 */

export const spacing = {
  // Spacing scale (8pt base)
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  xl: 20,
  xxl: 24,
  xxxl: 32,
  huge: 48,

  // Padding/Margin shortcuts
  padding: {
    xs: 4,
    sm: 8,
    md: 12,
    lg: 16,
    xl: 20,
    xxl: 24,
  },
  margin: {
    xs: 4,
    sm: 8,
    md: 12,
    lg: 16,
    xl: 20,
    xxl: 24,
  },

  // Common layouts
  screenPadding: 16,
  screenMargin: 16,
  cardPadding: 12,
  buttonPadding: 12,
  inputPadding: 12,

  // Border radius
  borderRadius: {
    xs: 4,
    sm: 8,
    md: 12,
    lg: 16,
    full: 999,
  },

  // Icon sizes
  iconSmall: 16,
  iconMedium: 20,
  iconLarge: 24,
  iconXL: 32,

  // Component heights
  buttonHeight: 44,
  inputHeight: 44,
  cardHeight: 60,
  listItemHeight: 60,

  // Divider
  dividerHeight: 1,
  dividerBold: 2,

  // Shadows
  elevation: {
    none: {},
    small: {
      shadowColor: '#000',
      shadowOffset: { width: 0, height: 2 },
      shadowOpacity: 0.08,
      shadowRadius: 4,
      elevation: 2,
    },
    medium: {
      shadowColor: '#000',
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.12,
      shadowRadius: 8,
      elevation: 4,
    },
    large: {
      shadowColor: '#000',
      shadowOffset: { width: 0, height: 8 },
      shadowOpacity: 0.16,
      shadowRadius: 12,
      elevation: 8,
    },
  },
} as const;

export type Spacing = typeof spacing[keyof typeof spacing];
