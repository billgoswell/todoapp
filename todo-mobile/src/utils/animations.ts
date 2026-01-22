/**
 * Animation Utilities
 *
 * Reusable animation configurations and helpers for smooth transitions.
 * Uses React Native Animated API for performance.
 */

import { Animated, Easing } from 'react-native';

/**
 * Screen transition animation configurations
 */
export const screenTransitionConfig = {
  // Fade transition (default for modals)
  fade: {
    cardStyleInterpolator: ({ current }: any) => ({
      cardStyle: {
        opacity: current.progress,
      },
    }),
  },

  // Slide from bottom (for modals)
  slideFromBottom: {
    cardStyleInterpolator: ({ current, layouts }: any) => ({
      cardStyle: {
        transform: [
          {
            translateY: current.progress.interpolate({
              inputRange: [0, 1],
              outputRange: [layouts.screen.height, 0],
            }),
          },
        ],
      },
    }),
  },

  // Slide from right (for navigation forward)
  slideFromRight: {
    cardStyleInterpolator: ({ current, layouts }: any) => ({
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
  },

  // Slide to left (for navigation back)
  slideToLeft: {
    cardStyleInterpolator: ({ current, layouts }: any) => ({
      cardStyle: {
        transform: [
          {
            translateX: current.progress.interpolate({
              inputRange: [0, 1],
              outputRange: [0, -layouts.screen.width],
            }),
          },
        ],
      },
    }),
  },
};

/**
 * Fade in animation
 * @param duration - Animation duration in milliseconds (default: 300)
 */
export const fadeIn = (duration: number = 300) => {
  const opacity = new Animated.Value(0);

  return {
    opacity,
    start: () =>
      new Promise<void>((resolve) => {
        Animated.timing(opacity, {
          toValue: 1,
          duration,
          easing: Easing.inOut(Easing.ease),
          useNativeDriver: true,
        }).start(resolve);
      }),
  };
};

/**
 * Fade out animation
 * @param duration - Animation duration in milliseconds (default: 300)
 */
export const fadeOut = (duration: number = 300) => {
  const opacity = new Animated.Value(1);

  return {
    opacity,
    start: () =>
      new Promise<void>((resolve) => {
        Animated.timing(opacity, {
          toValue: 0,
          duration,
          easing: Easing.inOut(Easing.ease),
          useNativeDriver: true,
        }).start(resolve);
      }),
  };
};

/**
 * Slide up animation (from bottom to original position)
 * @param distance - Distance to slide up in pixels (default: 50)
 * @param duration - Animation duration in milliseconds (default: 400)
 */
export const slideUp = (distance: number = 50, duration: number = 400) => {
  const translateY = new Animated.Value(distance);

  return {
    translateY,
    start: () =>
      new Promise<void>((resolve) => {
        Animated.timing(translateY, {
          toValue: 0,
          duration,
          easing: Easing.out(Easing.cubic),
          useNativeDriver: true,
        }).start(resolve);
      }),
  };
};

/**
 * Slide down animation (from original position to bottom)
 * @param distance - Distance to slide down in pixels (default: 50)
 * @param duration - Animation duration in milliseconds (default: 300)
 */
export const slideDown = (distance: number = 50, duration: number = 300) => {
  const translateY = new Animated.Value(0);

  return {
    translateY,
    start: () =>
      new Promise<void>((resolve) => {
        Animated.timing(translateY, {
          toValue: distance,
          duration,
          easing: Easing.in(Easing.cubic),
          useNativeDriver: true,
        }).start(resolve);
      }),
  };
};

/**
 * Scale animation (zoom in or out)
 * @param fromScale - Starting scale (default: 0.8)
 * @param toScale - Ending scale (default: 1)
 * @param duration - Animation duration in milliseconds (default: 300)
 */
export const scale = (
  fromScale: number = 0.8,
  toScale: number = 1,
  duration: number = 300
) => {
  const scale = new Animated.Value(fromScale);

  return {
    scale,
    start: () =>
      new Promise<void>((resolve) => {
        Animated.timing(scale, {
          toValue: toScale,
          duration,
          easing: Easing.out(Easing.cubic),
          useNativeDriver: true,
        }).start(resolve);
      }),
  };
};

/**
 * Pulse animation (opacity pulse)
 * @param duration - Full pulse cycle duration in milliseconds (default: 1500)
 * @param fromOpacity - Starting opacity (default: 0.5)
 * @param toOpacity - Peak opacity (default: 1)
 */
export const pulse = (
  duration: number = 1500,
  fromOpacity: number = 0.5,
  toOpacity: number = 1
) => {
  const opacity = new Animated.Value(toOpacity);

  const startPulse = () => {
    Animated.loop(
      Animated.sequence([
        Animated.timing(opacity, {
          toValue: fromOpacity,
          duration: duration / 2,
          easing: Easing.inOut(Easing.ease),
          useNativeDriver: true,
        }),
        Animated.timing(opacity, {
          toValue: toOpacity,
          duration: duration / 2,
          easing: Easing.inOut(Easing.ease),
          useNativeDriver: true,
        }),
      ])
    ).start();
  };

  return {
    opacity,
    startPulse,
    stop: () => {
      opacity.setValue(toOpacity);
    },
  };
};

/**
 * Bounce animation
 * @param height - Bounce height in pixels (default: 20)
 * @param duration - Animation duration in milliseconds (default: 600)
 */
export const bounce = (height: number = 20, duration: number = 600) => {
  const translateY = new Animated.Value(0);

  return {
    translateY,
    start: () =>
      new Promise<void>((resolve) => {
        Animated.sequence([
          Animated.timing(translateY, {
            toValue: -height,
            duration: duration / 2,
            easing: Easing.out(Easing.cubic),
            useNativeDriver: true,
          }),
          Animated.timing(translateY, {
            toValue: 0,
            duration: duration / 2,
            easing: Easing.in(Easing.cubic),
            useNativeDriver: true,
          }),
        ]).start(resolve);
      }),
  };
};

/**
 * Height animation (collapse/expand)
 * @param fromHeight - Starting height in pixels
 * @param toHeight - Ending height in pixels
 * @param duration - Animation duration in milliseconds (default: 300)
 */
export const expandHeight = (
  fromHeight: number,
  toHeight: number,
  duration: number = 300
) => {
  const height = new Animated.Value(fromHeight);

  return {
    height,
    start: () =>
      new Promise<void>((resolve) => {
        Animated.timing(height, {
          toValue: toHeight,
          duration,
          easing: Easing.inOut(Easing.ease),
          useNativeDriver: false, // Can't use native driver for width/height
        }).start(resolve);
      }),
  };
};

/**
 * Rotate animation
 * @param fromDegrees - Starting rotation in degrees (default: 0)
 * @param toDegrees - Ending rotation in degrees (default: 360)
 * @param duration - Animation duration in milliseconds (default: 1000)
 */
export const rotate = (
  fromDegrees: number = 0,
  toDegrees: number = 360,
  duration: number = 1000
) => {
  const rotation = new Animated.Value(fromDegrees);

  return {
    rotation,
    rotate: rotation.interpolate({
      inputRange: [0, 1],
      outputRange: ['0deg', '360deg'],
    }),
    start: () =>
      new Promise<void>((resolve) => {
        Animated.timing(rotation, {
          toValue: toDegrees,
          duration,
          easing: Easing.linear,
          useNativeDriver: true,
        }).start(resolve);
      }),
  };
};

/**
 * Stagger animation (animate items with delay)
 * Creates a sequence of animations with specified delay between each
 * @param animations - Array of animation promises
 * @param delayBetween - Delay between each animation in milliseconds (default: 100)
 */
export const stagger = async (
  animations: (() => Promise<void>)[],
  delayBetween: number = 100
) => {
  for (let i = 0; i < animations.length; i++) {
    if (i > 0) {
      await new Promise((resolve) => setTimeout(resolve, delayBetween));
    }
    await animations[i]();
  }
};

/**
 * Parallel animation (run multiple animations at the same time)
 * @param animations - Array of animation promises
 */
export const parallel = async (animations: (() => Promise<void>)[]) => {
  await Promise.all(animations.map((anim) => anim()));
};

/**
 * Sequence animation (run animations one after another)
 * @param animations - Array of animation promises
 */
export const sequence = async (animations: (() => Promise<void>)[]) => {
  for (const anim of animations) {
    await anim();
  }
};

/**
 * Spring animation (bouncy animation)
 * @param duration - Animation duration in milliseconds (default: 500)
 * @param tension - Spring tension (default: 40)
 * @param friction - Spring friction (default: 7)
 */
export const spring = (
  duration: number = 500,
  tension: number = 40,
  friction: number = 7
) => {
  const value = new Animated.Value(0);

  return {
    value,
    start: () =>
      new Promise<void>((resolve) => {
        Animated.spring(value, {
          toValue: 1,
          tension,
          friction,
          useNativeDriver: true,
        }).start(resolve);
      }),
  };
};

/**
 * Shake animation (side-to-side movement)
 * @param distance - Shake distance in pixels (default: 10)
 * @param duration - Animation duration in milliseconds (default: 400)
 */
export const shake = (distance: number = 10, duration: number = 400) => {
  const translateX = new Animated.Value(0);

  return {
    translateX,
    start: () =>
      new Promise<void>((resolve) => {
        Animated.sequence([
          Animated.timing(translateX, {
            toValue: distance,
            duration: duration / 4,
            useNativeDriver: true,
          }),
          Animated.timing(translateX, {
            toValue: -distance,
            duration: duration / 4,
            useNativeDriver: true,
          }),
          Animated.timing(translateX, {
            toValue: distance,
            duration: duration / 4,
            useNativeDriver: true,
          }),
          Animated.timing(translateX, {
            toValue: 0,
            duration: duration / 4,
            useNativeDriver: true,
          }),
        ]).start(resolve);
      }),
  };
};

/**
 * Timing configuration for consistent animations
 */
export const animationTiming = {
  fast: 200,
  normal: 300,
  slow: 500,
  verySlow: 800,
};

/**
 * Easing functions
 */
export const easings = {
  linear: Easing.linear,
  ease: Easing.ease,
  easeIn: Easing.in(Easing.cubic),
  easeOut: Easing.out(Easing.cubic),
  easeInOut: Easing.inOut(Easing.cubic),
  bounce: Easing.bounce,
  elastic: Easing.elastic,
};
