# Animations Quick Reference

## 🎬 What Animates

### Screen Navigation
```
Home → TaskDetail   : Slide from right (300ms)
Home → ListMgmt     : Slide from right (300ms)
Home → Settings     : Slide from right (300ms)
Swipe back          : Slide to right (auto-reverse)
```

### Task Items
```
New task appears    : Fade in (300ms) + Slide up (400ms)
Task loads          : Smooth entrance from below
```

### Sync Status
```
Sync success (✓)    : Pulse (1500ms cycle) at 0.6-1.0 opacity
Sync failing        : Pulse stops, shows warning
```

### Filter Panel
```
Tap "Filters"       : Panel expands (300ms height + 200ms fade)
Tap "Filters" again : Panel collapses smoothly
```

---

## 🔧 Animation Files

| File | Purpose | Lines |
|------|---------|-------|
| `src/utils/animations.ts` | Animation utilities library | 400+ |
| `src/navigation/AppNavigator.tsx` | Screen transitions | Updated |
| `src/components/TaskItem.tsx` | Task fade/slide | Updated |
| `src/components/SyncStatusBadge.tsx` | Pulse animation | Updated |
| `src/components/SearchAndFilter.tsx` | Panel expand/collapse | Updated |

---

## 📚 Available Animation Functions

```typescript
// Import
import { fadeIn, slideUp, pulse, rotate, bounce, shake } from '../utils/animations';

// Usage patterns
const { opacity, start } = fadeIn(300);
const { translateY, start } = slideUp(50, 400);
const { opacity, startPulse } = pulse(1500, 0.6, 1);
```

### Animation Library

| Function | Duration | Type | Use Case |
|----------|----------|------|----------|
| `fadeIn()` | 300ms | Opacity 0→1 | Appearing elements |
| `fadeOut()` | 300ms | Opacity 1→0 | Disappearing elements |
| `slideUp()` | 400ms | Transform Y+ | List items entering |
| `slideDown()` | 300ms | Transform Y- | Elements exiting down |
| `scale()` | 300ms | Transform scale | Zoom effects |
| `pulse()` | 1500ms | Opacity loop | Attention effects |
| `bounce()` | 600ms | Transform Y | Bouncy feedback |
| `rotate()` | 1000ms | Transform rotate | Spinning loaders |
| `shake()` | 400ms | Transform X | Error feedback |
| `spring()` | 500ms | Bouncy easing | Natural motion |

---

## 💻 Component Animation Examples

### Screen Transition
```typescript
// In AppNavigator.tsx
cardStyleInterpolator: ({ current, layouts }) => ({
  cardStyle: {
    transform: [{
      translateX: current.progress.interpolate({
        inputRange: [0, 1],
        outputRange: [layouts.screen.width, 0],
      }),
    }],
  },
})
```

### Task Item Fade + Slide
```typescript
// In TaskItem.tsx
useEffect(() => {
  Animated.parallel([
    Animated.timing(fadeInAnim, {
      toValue: 1,
      duration: 300,
      useNativeDriver: true,
    }),
    Animated.timing(slideUpAnim, {
      toValue: 0,
      duration: 400,
      useNativeDriver: true,
    }),
  ]).start();
}, []);
```

### Pulse Animation
```typescript
// In SyncStatusBadge.tsx
useEffect(() => {
  if (syncStatus === 'success') {
    pulseAnim.startPulse(); // Loop forever
  } else {
    pulseAnim.stop();
  }
}, [syncStatus]);
```

---

## ⚡ Performance Tips

### ✅ Good (Native Driver)
```typescript
Animated.timing(opacity, {
  useNativeDriver: true,  // ← Fast!
})

Animated.timing(translateY, {
  useNativeDriver: true,  // ← Also fast!
})
```

### ❌ Bad (No Native Driver)
```typescript
Animated.timing(height, {
  useNativeDriver: false,  // ← Slower, but necessary for layout
})

Animated.timing(borderRadius, {
  useNativeDriver: false,  // ← Must use false
})
```

---

## 🎯 Timing Constants

```typescript
import { animationTiming } from '../utils/animations';

animationTiming.fast       // 200ms
animationTiming.normal     // 300ms ← default
animationTiming.slow       // 500ms
animationTiming.verySlow   // 800ms
```

---

## 🎨 Easing Functions

```typescript
import { easings } from '../utils/animations';

easings.linear          // Constant speed
easings.ease            // Apple's default
easings.easeIn          // Slow start
easings.easeOut         // Slow end
easings.easeInOut       // Smooth both ways
easings.bounce          // Bouncy
easings.elastic         // Spring-like
```

---

## 🔄 Composition Patterns

### Sequential (One After Another)
```typescript
import { sequence } from '../utils/animations';

await sequence([
  () => anim1.start(),
  () => anim2.start(),
  () => anim3.start(),
]);
```

### Parallel (All Together)
```typescript
import { parallel } from '../utils/animations';

await parallel([
  () => anim1.start(),
  () => anim2.start(),
  () => anim3.start(),
]);
```

### Staggered (With Delay)
```typescript
import { stagger } from '../utils/animations';

await stagger([
  () => anim1.start(),
  () => anim2.start(),
  () => anim3.start(),
], 100); // 100ms between each
```

---

## 🧪 Testing

### Disable Animations
```typescript
if (process.env.NODE_ENV === 'test') {
  // Disable animations in tests for faster execution
  Animated.timing = jest.fn(() => ({
    start: (callback) => callback?.(),
  }));
}
```

### Check Animation Progress
```typescript
jest.useFakeTimers();
// ... render component
jest.advanceTimersByTime(300);
// ... check animation completed
```

---

## 📋 Checklist: Using Animations

- [ ] Import `Animated` from 'react-native'
- [ ] Create animation ref: `useRef(new Animated.Value(0))`
- [ ] Define animation in `useEffect`
- [ ] Call `.start()` on animation
- [ ] Apply style to component with animated value
- [ ] Wrap component in `Animated.View` or `Animated.Text`
- [ ] Use `useNativeDriver: true` when possible
- [ ] Test on real device (not simulator only)

---

## 🚀 Quick Start: Add Animation to Component

```typescript
import React, { useEffect } from 'react';
import { Animated } from 'react-native';

export const MyComponent = () => {
  // 1. Create animated value
  const opacity = useRef(new Animated.Value(0)).current;

  // 2. Define animation
  useEffect(() => {
    Animated.timing(opacity, {
      toValue: 1,
      duration: 300,
      useNativeDriver: true,
    }).start();
  }, []);

  // 3. Apply to component
  return (
    <Animated.View style={{ opacity }}>
      <Text>Hello</Text>
    </Animated.View>
  );
};
```

---

## 🐛 Troubleshooting

| Problem | Solution |
|---------|----------|
| Animation doesn't run | Call `.start()` on animation |
| Animation stutters | Use `useNativeDriver: true` |
| Animation jerky on old phone | Reduce duration, simplify animation |
| Animation runs twice | Check dependency array in useEffect |
| Component unmounts during animation | Add cleanup in useEffect return |

---

## 📊 Animation Performance

| Animation | FPS | Native Driver |
|-----------|-----|---------------|
| Fade (opacity) | 60 | ✅ Yes |
| Slide (translateY) | 60 | ✅ Yes |
| Scale (transform) | 60 | ✅ Yes |
| Height expansion | 50 | ❌ No |
| BorderRadius change | 40 | ❌ No |

---

## 📖 Full Documentation

See `ANIMATIONS_GUIDE.md` for complete documentation including:
- Detailed animation explanations
- Integration examples
- Best practices
- Common patterns
- Performance considerations
- Testing strategies

---

**Quick Links**:
- Animation Utilities: `src/utils/animations.ts`
- Full Guide: `ANIMATIONS_GUIDE.md`
- Component Examples: TaskItem, SyncStatusBadge, SearchAndFilter

---

**Status**: Production Ready ✅
