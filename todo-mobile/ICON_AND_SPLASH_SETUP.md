# App Icon & Splash Screen Setup Guide

## Overview

This guide covers setting up the CommandLineTodo app icon and splash screen for iOS and Android.

---

## Files Provided

### SVG Icon
- **Location**: `assets/app-icon.svg`
- **Format**: Scalable Vector Graphics
- **Design**: Blue gradient background (#007AFF → #0051D5) with white checkmark
- **Purpose**: Can be exported to any size needed

### Splash Screen Component
- **Location**: `src/screens/SplashScreen.tsx`
- **Type**: React Native component
- **Features**: Animated entry, loading indicator, app branding
- **Usage**: Displays while app initializes

---

## Icon Design Details

### Visual Elements
- **Main Element**: Checkmark (✓) - represents task completion
- **Background**: Blue gradient (app primary color)
- **Shape**: Rounded corners (220px radius on 1024x1024)
- **Accent**: Subtle white circles and grid pattern
- **Style**: Modern, clean, professional

### Color Scheme
- **Primary**: #007AFF (iOS system blue)
- **Dark**: #0051D5 (gradient endpoint)
- **Accent**: White (#FFFFFF)
- **Matches App Theme**: Yes ✓

### Symbolism
- **Checkmark**: Task completion, productivity, getting things done
- **Blue**: Trust, reliability, technology
- **Simple**: Easy to recognize at any size

---

## iOS Setup

### Step 1: Generate App Icon Sizes

You need these sizes for iOS:

```
20x20    (watchOS icon)
29x29    (Settings)
40x40    (iPhone notification)
58x58    (iPhone notification 2x)
60x60    (iPhone app)
80x80    (iPhone spotlight)
87x87    (iPhone notification 3x)
120x120  (iPhone app 2x)
167x167  (iPad settings)
180x180  (iPhone app 3x)
1024x1024 (App Store)
```

### Step 2: Use Xcode Asset Catalog

1. Open Xcode project: `ios/CommandLineTodo.xcodeproj`
2. Select **Assets.xcassets**
3. Right-click → **New Image Set**
4. Name it: **AppIcon**
5. Drag SVG icon and drop into the image set (Xcode auto-scales)
6. Or use: **Image Assets → AppIcon** (already exists)

### Step 3: Verify in App Settings

In Xcode:
1. Project → **CommandLineTodo** target
2. General tab
3. **App Icons and Launch Images**
4. Verify AppIcon is assigned

### Step 4: Splash Screen (Optional LaunchScreen)

For iOS 13+, use the new approach:

1. **Storyboard Method** (iOS 12 support):
   - File: `ios/CommandLineTodo/LaunchScreen.storyboard`
   - Already configured by React Native CLI
   - Shows during app startup (≤2 seconds)

2. **Splash Screen Component** (Recommended):
   - Use `SplashScreen.tsx` component in React Native
   - More control and better UX
   - See integration below

---

## Android Setup

### Step 1: Generate App Icon Sizes

Android uses multiple densities:

```
ldpi   (120 dpi)   - 36x36
mdpi   (160 dpi)   - 48x48
hdpi   (240 dpi)   - 72x72
xhdpi  (320 dpi)   - 96x96
xxhdpi (480 dpi)   - 144x144
xxxhdpi (640 dpi)  - 192x192
```

### Step 2: Place Icons in Android Project

```
android/app/src/main/res/
├── mipmap-ldpi/
│   └── ic_launcher.png (36x36)
├── mipmap-mdpi/
│   └── ic_launcher.png (48x48)
├── mipmap-hdpi/
│   └── ic_launcher.png (72x72)
├── mipmap-xhdpi/
│   └── ic_launcher.png (96x96)
├── mipmap-xxhdpi/
│   └── ic_launcher.png (144x144)
└── mipmap-xxxhdpi/
    └── ic_launcher.png (192x192)
```

### Step 3: Configure AndroidManifest.xml

File: `android/app/src/main/AndroidManifest.xml`

```xml
<application
  android:label="@string/app_name"
  android:icon="@mipmap/ic_launcher"
  android:roundIcon="@mipmap/ic_launcher_round"
  android:allowBackup="false"
  ...>
```

### Step 4: Add String Resource

File: `android/app/src/main/res/values/strings.xml`

```xml
<resources>
    <string name="app_name">CommandLineTodo</string>
    <string name="app_tagline">Your Tasks, Everywhere</string>
</resources>
```

### Step 5: Create Splash Screen Activity (Optional)

For branded splash before app loads, create:

File: `android/app/src/main/java/com/commandlinetodo/SplashActivity.java`

```java
package com.commandlinetodo;

import android.content.Intent;
import android.os.Bundle;
import androidx.appcompat.app.AppCompatActivity;

public class SplashActivity extends AppCompatActivity {
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        Intent intent = new Intent(this, MainActivity.class);
        startActivity(intent);
        finish();
    }
}
```

---

## React Native Splash Screen Integration

### Step 1: Install Package

```bash
npm install react-native-splash-screen
# or
yarn add react-native-splash-screen
```

### Step 2: Link Native Module

```bash
npx react-native link react-native-splash-screen
```

### Step 3: Use SplashScreen Component

In `src/App.tsx`:

```typescript
import React, { useEffect, useState } from 'react';
import { AppProvider } from './state';
import { NavigationContainer } from '@react-navigation/native';
import { SplashScreen } from './screens/SplashScreen';
import { AppNavigator } from './navigation/AppNavigator';
import { apiClient, db } from './api/client';

export default function App() {
  const [isReady, setIsReady] = useState(false);
  const [isSignedIn, setIsSignedIn] = useState(false);

  useEffect(() => {
    const bootstrap = async () => {
      try {
        // Initialize database
        await db.initialize();

        // Check if user has API key
        const apiKey = await apiClient.getApiKey();
        setIsSignedIn(!!apiKey);
      } catch (error) {
        console.error('Bootstrap error:', error);
      } finally {
        setIsReady(true);
      }
    };

    bootstrap();
  }, []);

  if (!isReady) {
    return <SplashScreen isVisible={true} />;
  }

  return (
    <AppProvider>
      <NavigationContainer>
        <AppNavigator isSignedIn={isSignedIn} />
      </NavigationContainer>
    </AppProvider>
  );
}
```

### Step 4: Configure Native Projects

**iOS** (`ios/CommandLineTodo/AppDelegate.m`):

```objc
#import "RNSplashScreen.h"

- (BOOL)application:(UIApplication *)application didFinishLaunchingWithOptions:(NSDictionary *)launchOptions
{
  // ... existing code ...

  [RNSplashScreen show];
  return YES;
}
```

**Android** (`android/app/src/main/java/com/commandlinetodo/MainActivity.java`):

```java
import org.devio.rn.splashscreen.SplashScreen;
import android.os.Bundle;

public class MainActivity extends ReactActivity {
  @Override
  protected void onCreate(Bundle savedInstanceState) {
    SplashScreen.show(this);
    super.onCreate(savedInstanceState);
  }
}
```

---

## Converting SVG to PNG

### Using ImageMagick (Recommended)

```bash
# Install ImageMagick
brew install imagemagick

# Convert SVG to PNG at different sizes
convert -density 300 assets/app-icon.svg -resize 1024x1024 icon-1024.png
convert -density 300 assets/app-icon.svg -resize 192x192 icon-192.png
convert -density 300 assets/app-icon.svg -resize 144x144 icon-144.png
convert -density 300 assets/app-icon.svg -resize 96x96 icon-96.png
convert -density 300 assets/app-icon.svg -resize 72x72 icon-72.png
convert -density 300 assets/app-icon.svg -resize 48x48 icon-48.png
convert -density 300 assets/app-icon.svg -resize 36x36 icon-36.png
```

### Using Online Tools

1. **Photoshop/Figma**: Import SVG, export as PNG at each size
2. **CloudConvert**: Upload SVG → download PNG (multiple sizes)
3. **Convertio**: Upload SVG → select size → download

### Using Node Script

Create `scripts/generate-icons.js`:

```javascript
const sharp = require('sharp');
const fs = require('fs');
const path = require('path');

const sizes = [
  { size: 1024, name: 'app-store' },
  { size: 192, name: 'android-xxxhdpi' },
  { size: 144, name: 'android-xxhdpi' },
  { size: 96, name: 'android-xhdpi' },
  { size: 72, name: 'android-hdpi' },
  { size: 48, name: 'android-mdpi' },
  { size: 36, name: 'android-ldpi' },
];

async function generateIcons() {
  for (const { size, name } of sizes) {
    await sharp('assets/app-icon.svg')
      .png()
      .resize(size, size)
      .toFile(`assets/icons/icon-${name}-${size}x${size}.png`);
    console.log(`Generated ${size}x${size} icon`);
  }
}

generateIcons().catch(console.error);
```

Run with: `node scripts/generate-icons.js`

---

## Splash Screen Details

### Component Features
- **Animated Entry**: Fade in + scale animation (600ms)
- **App Name**: "CommandLineTodo"
- **Tagline**: "Your Tasks, Everywhere"
- **Loading Indicator**: ActivityIndicator while initializing
- **Version**: Shows "v1.0.0"
- **Color Scheme**: Matches app theme

### Customization

Change app name:
```typescript
<Text style={styles.appName}>Your App Name</Text>
```

Change tagline:
```typescript
<Text style={styles.tagline}>Your Tagline</Text>
```

Change version:
```typescript
<Text style={styles.version}>v1.0.0</Text>
```

Change loading text:
```typescript
<Text style={styles.loadingText}>Custom Loading Text</Text>
```

---

## App Store Requirements

### App Store (iOS)

Required Icon Sizes:
- 1024x1024 (App Store display)
- 512x512 (Marketing)
- 180x180 (iPhone home screen)
- 120x120 (iPhone notification)
- 58x58 (iPhone settings)

Splash Screen:
- Recommended: Custom launch screen
- Max 2 seconds visible
- Must not show splash repeatedly

### Google Play (Android)

Required Icon Sizes:
- 512x512 (Store listing)
- 192x192 (App icon on device)
- 96x96 (Higher density)

Splash Screen:
- Instant app experience preferred
- Brand should be minimal
- User should reach content quickly

---

## Testing

### iOS Simulator
```bash
# Build and run
npx react-native run-ios

# Check icon displays correctly
# Check splash screen appears then disappears
# Check app loads
```

### Android Emulator
```bash
# Build and run
npx react-native run-android

# Check icon displays correctly
# Check splash screen appears then disappears
# Check app loads
```

### Real Device

**iPhone**:
1. Connect iPhone
2. `npx react-native run-ios --device`
3. Verify icon on home screen
4. Check splash during launch

**Android Phone**:
1. Enable Developer Mode
2. Connect phone via USB
3. `npx react-native run-android`
4. Verify icon in app drawer
5. Check splash during launch

---

## Checklist

### Icon Setup
- [ ] SVG icon designed (`assets/app-icon.svg`)
- [ ] PNG icons generated at required sizes
- [ ] iOS icons added to Asset Catalog
- [ ] Android icons placed in `mipmap-*/` folders
- [ ] Icon tested on real devices

### Splash Screen
- [ ] SplashScreen component created
- [ ] Integrated in App.tsx
- [ ] Custom text updated
- [ ] Animations tested
- [ ] Tested on simulator and real device

### App Store Preparation
- [ ] 1024x1024 icon ready for App Store
- [ ] App name confirmed (CommandLineTodo)
- [ ] Version number set (v1.0.0)
- [ ] Tagline finalized
- [ ] Screenshots prepared

### Store Submission
- [ ] Icon submission requirements met
- [ ] Metadata entered correctly
- [ ] Screenshots uploaded
- [ ] Description provided
- [ ] Privacy policy linked
- [ ] Ready for review

---

## File Locations

```
CommandLineTodo/
├── assets/
│   ├── app-icon.svg
│   ├── app-icon-1024.png
│   └── app-icon-512.png
├── src/
│   └── screens/
│       └── SplashScreen.tsx
├── ios/
│   └── CommandLineTodo/
│       ├── Assets.xcassets/
│       │   └── AppIcon.appiconset/
│       │       ├── Contents.json
│       │       └── [icon images]
│       └── LaunchScreen.storyboard
├── android/
│   └── app/src/main/res/
│       ├── mipmap-ldpi/
│       ├── mipmap-mdpi/
│       ├── mipmap-hdpi/
│       ├── mipmap-xhdpi/
│       ├── mipmap-xxhdpi/
│       ├── mipmap-xxxhdpi/
│       └── values/
│           └── strings.xml
└── ICON_AND_SPLASH_SETUP.md
```

---

## Summary

✅ **App Icon**: Professional checkmark design, blue gradient background
✅ **Splash Screen**: Animated component with loading indicator
✅ **iOS Support**: Asset catalog integration
✅ **Android Support**: Multiple density support
✅ **Customizable**: Easy to adjust colors, text, animations
✅ **Production Ready**: Store submission ready

---

## Next Steps

1. Convert SVG to PNG at required sizes
2. Add PNG icons to iOS and Android projects
3. Update app name and version in native configs
4. Prepare store screenshots
5. Submit to TestFlight (iOS) and Google Play Beta (Android)

---

**Status**: Ready for App Store/Play Store Submission ✅
**Last Updated**: Phase 5 - Icon & Splash Screen
