/**
 * Theme
 *
 * Centralized design tokens for consistent UI
 */

export { colors, type Color } from './colors';
export { spacing, type Spacing } from './spacing';
export { typography, type Typography } from './typography';

// Common shortcuts
export const theme = {
  colors: require('./colors').colors,
  spacing: require('./spacing').spacing,
  typography: require('./typography').typography,
};
