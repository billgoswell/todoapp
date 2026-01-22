# App Store Submission Guide

## Overview

This guide covers preparing CommandLineTodo for submission to Apple App Store and Google Play Store.

---

## Pre-Submission Checklist

### App Build & Configuration

- [ ] App version set to 1.0.0
- [ ] Build number set (e.g., 1)
- [ ] App icon (1024x1024) created and tested
- [ ] Splash screen implemented and tested
- [ ] All screens tested on target devices
- [ ] No console errors or warnings
- [ ] Animations smooth at 60fps
- [ ] All links working (API, auth, sync)
- [ ] Privacy policy URL accessible
- [ ] Terms of service ready (optional)

### Code Quality

- [ ] No hardcoded credentials
- [ ] No debug logging in production
- [ ] No test/debug screens visible
- [ ] Error handling in place
- [ ] Network timeouts handled
- [ ] Data validation complete
- [ ] TypeScript strict mode enabled
- [ ] No `any` types (except where necessary)

### Performance

- [ ] App launches in <2 seconds
- [ ] Database operations don't block UI
- [ ] 1000+ tasks load smoothly
- [ ] Animations are 60fps
- [ ] Memory usage <100MB
- [ ] No memory leaks on long usage
- [ ] Sync doesn't cause freezes

### Functionality Testing

- [ ] Create tasks works offline
- [ ] Edit tasks works offline
- [ ] Delete tasks works offline
- [ ] Mark complete works offline
- [ ] Sync works when online
- [ ] Conflict resolution works
- [ ] Network status shows correctly
- [ ] Offline indicator appears/disappears
- [ ] Due date picker works
- [ ] Search and filters work
- [ ] List management works
- [ ] Settings screen works
- [ ] Manual sync works
- [ ] Background sync works

---

## iOS App Store Submission

### Step 1: Create App Store Connect Account

1. Go to [App Store Connect](https://appstoreconnect.apple.com)
2. Sign in with Apple Developer account
3. Select **My Apps**
4. Click **+** → **New App**
5. Choose platform: **iOS**
6. Fill in details:
   - **Name**: CommandLineTodo
   - **Primary Language**: English
   - **Bundle ID**: com.commandlinetodo.mobile
   - **SKU**: commandlinetodo (unique)
   - **User Access**: Limited Access (recommended)

### Step 2: Complete App Information

In App Store Connect:

**App Information**:
- **Name**: CommandLineTodo
- **Subtitle**: (Optional)
- **Category**: Productivity
- **Content Rights**: Owned by you
- **Apple ID**: (Generate if needed)

**Rating**:
- Select appropriate ratings
- Indicate content in app (if any)

**App Privacy**:
- Link to Privacy Policy
- Indicate data collection:
  - ✓ User Data (API key)
  - ✓ Sync & Backup (task data)
  - No tracking or advertising

### Step 3: Create Release

**Pricing & Availability**:
- Price: Free
- Availability: Select countries (suggest: Worldwide)
- Release date: Now or scheduled

**Version Information**:
- Version: 1.0.0
- Build: Upload via Xcode or Transporter

### Step 4: Prepare Screenshots

Required sizes: **1242x2208** (iPhone) and **2048x2732** (iPad)

Create 2-5 screenshots showing:
1. **Home Screen**: Task list with sync status
2. **Create Task**: Using DueDatePicker
3. **Search & Filter**: Demonstrating filtering
4. **Offline Capability**: Offline indicator
5. **Settings**: API key configuration

**Screenshot Requirements**:
- Clean, professional appearance
- No placeholders or test data
- Optional: Overlay text explaining features
- Consistent branding across all shots

### Step 5: Write App Description

```
CommandLineTodo - Your Tasks, Everywhere

Stay productive with CommandLineTodo, the intuitive todo app that works offline and syncs seamlessly to the cloud.

FEATURES:
• Create, edit, and complete tasks offline
• Automatic cloud sync when connected
• Search and filter tasks by status or priority
• Set due dates with beautiful date picker
• Real-time sync status indicator
• Support for multiple task lists
• Beautiful, intuitive interface
• Fast and reliable task management

PRIVACY:
CommandLineTodo respects your privacy. Your task data is stored locally on your device. Sync requires API key connection. We do not track, collect, or sell user data.

OFFLINE FIRST:
Work without internet. All changes sync automatically when you reconnect.

VERSION 1.0.0:
Initial release with complete offline support and cloud sync.
```

### Step 6: Review Guidelines & Compliance

Review Apple's [App Review Guidelines](https://developer.apple.com/app-store/review/guidelines/):

Your app must:
- ✅ Be a complete, functional app
- ✅ Have accurate description
- ✅ Work as described
- ✅ Not crash or freeze
- ✅ Have privacy policy
- ✅ Not use private APIs
- ✅ Not jailbreak detection bypass

### Step 7: Build for Release

In Xcode:

```bash
# Clean build
xcodebuild clean -workspace ios/CommandLineTodo.xcworkspace \
  -scheme CommandLineTodo \
  -configuration Release

# Build for archiving
xcodebuild archive -workspace ios/CommandLineTodo.xcworkspace \
  -scheme CommandLineTodo \
  -configuration Release \
  -archivePath ./CommandLineTodo.xcarchive
```

Or use Xcode GUI:
1. Product → Destination → Generic iOS Device
2. Product → Build
3. Product → Archive
4. Organizer → Validate App
5. Upload to App Store

### Step 8: Submit for Review

In App Store Connect:
1. Select version
2. Review all sections (filled in)
3. Check: "Ready to Submit for Review"
4. Click **Submit for Review**
5. Answer any questions about:
   - Encryption use
   - Privacy details
   - Export compliance
6. Submit

**Expected wait**: 24-48 hours for initial review

---

## Google Play Store Submission

### Step 1: Create Google Play Developer Account

1. Go to [Google Play Console](https://play.google.com/console)
2. Sign in with Google account
3. Complete developer registration:
   - Accept Developer Agreement
   - Pay $25 one-time fee
   - Complete store listing

### Step 2: Create App

1. Click **Create App**
2. Fill in:
   - **App Name**: CommandLineTodo
   - **Language**: English
   - **App or Game**: App
   - **Category**: Productivity
   - **Default Language**: English

### Step 3: Complete Store Listing

**App Details**:
- Title: CommandLineTodo
- Short description: Your Tasks, Everywhere
- Full description: (see above)
- Developer name: Your Name
- Developer email: your@email.com

**App Icon**:
- Size: 512x512 PNG
- No rounded corners (store adds them)

**Feature Graphic**:
- Size: 1024x500 PNG
- Showcase your app's best feature

**Screenshots**:
- Minimum 2, maximum 8
- Size: 1080x1920 (phone) or 2560x1440 (tablet)
- Same as iOS screenshots (adapted)

**Category & Rating**:
- Category: Productivity
- Content Rating: Fill questionnaire
  - No sexual content
  - No violence
  - No hate speech
  - Safe for all ages

### Step 4: Prepare Release

**Release Strategy**:
1. Start with **Internal Testing**
2. Expand to **Closed Testing** (invite beta testers)
3. Move to **Staged Rollout** (10% users first)
4. Full release when stable

### Step 5: Configure Release

**App Signing**:
- Google Play handles signing for you
- Create signing key if first time
- Save backup key safely

**Release Details**:
- Version: 1.0.0
- Release notes:
  ```
  Initial release of CommandLineTodo!

  • Create and manage tasks offline
  • Automatic cloud synchronization
  • Search and filter capabilities
  • Beautiful date picker
  • Real-time sync status

  Enjoy productive task management!
  ```

### Step 6: Build for Release

```bash
# Create release keystore
keytool -genkey -v -keystore ~/commandlinetodo-release-key.keystore \
  -keyalg RSA -keysize 2048 -validity 10000 \
  -alias commandlinetodo

# Build release APK
cd android
./gradlew bundleRelease
```

Or use Android Studio:
1. Build → Generate Signed Bundle/APK
2. Select Bundle (AAB format - preferred)
3. Create signing key or use existing
4. Select Release config
5. Generate

### Step 7: Review Content Rating Questionnaire

Google Play requires content rating:
1. Click **Content Rating**
2. Answer questionnaire:
   - Violence: No
   - Hate Speech: No
   - Sexual Content: No
   - Substance Use: No
   - Gambling: No
   - Other Restrictions: No
3. Save ratings

### Step 8: Set Pricing

1. Click **Pricing & Distribution**
2. Select: **Free**
3. Select countries (all recommended)
4. Accept terms
5. Save

### Step 9: Submit for Review

1. Go to **Release** section
2. Click **Create new release**
3. Upload AAB/APK file
4. Add release notes
5. Review app details
6. Click **Review** to validate
7. Click **Submit for review**

**Expected wait**: 2-4 hours for initial review (much faster than iOS!)

---

## TestFlight Beta Testing (iOS)

### Before Public Release

1. In App Store Connect → **TestFlight**
2. Click **Create a New Group**
3. Name: "Beta Testers"
4. Add testers:
   - Enter email addresses
   - Up to 10,000 external testers
   - Click "Add External Tester"
5. Send invitation link to testers
6. Testers download TestFlight app
7. Accept invitation
8. Test your app

### Feedback Collection

In TestFlight tester feedback:
- Collect crash reports
- Monitor performance
- Gather feature requests
- Fix bugs before public release

---

## Google Play Beta Testing (Android)

### Before Public Release

1. In Google Play Console → **Closed Testing**
2. Click **Create new track**
3. Name: "Beta"
4. Click **Create release**
5. Upload AAB
6. Add release notes
7. Under Manage testers:
   - Add Google Group or email addresses
   - Up to 2,000 testers
8. Click **Save**
9. Click **Review** and submit

### Testing Duration

- Run beta for minimum 1-2 weeks
- Collect feedback
- Fix reported bugs
- Monitor crash reports in Google Play Console

---

## Pre-Launch Checks

### 24 Hours Before Submission

- [ ] Run final build test on real device
- [ ] Check app doesn't crash
- [ ] Verify all buttons work
- [ ] Test offline functionality
- [ ] Test online sync
- [ ] Check UI text for typos
- [ ] Verify app icons display correctly
- [ ] Confirm version number
- [ ] Review privacy policy one more time

### During Beta Phase

- [ ] Monitor crash reports
- [ ] Fix reported bugs
- [ ] Collect user feedback
- [ ] Optimize based on crashes
- [ ] Run for 1-2 weeks before public release

---

## Common Issues & Solutions

### Issue: App Rejected for Policy Violation

**Solution**:
- Review rejection reason carefully
- Most common: Missing privacy policy
- Add privacy policy link
- Address specific concerns in resubmission
- Resubmit

### Issue: Crash Reports in Beta

**Solution**:
- Use stack traces to identify issue
- Fix bug in code
- Create new beta build
- Resubmit to beta testers
- Wait for confirmation
- Then submit to public

### Issue: Slow App Startup

**Solution**:
- Profile app with Xcode/Android Studio
- Optimize database initialization
- Lazy load features
- Use background threads for heavy work
- Target: <2 second startup

### Issue: High Memory Usage

**Solution**:
- Profile with memory tools
- Fix memory leaks
- Reduce task list loading (paginate)
- Release unused resources
- Target: <100MB

---

## Monitoring After Release

### iOS App Store

In App Store Connect:
- **Analytics**: View downloads, crashes, ratings
- **Customer Reviews**: Read user feedback
- **Crashes**: Monitor and fix
- **Engagement**: Track user metrics

### Google Play

In Google Play Console:
- **Vitals**: Crash reports, ANRs, performance
- **Analytics**: User acquisition, engagement
- **Ratings & Reviews**: User feedback
- **Statistics**: Downloads by region, device

---

## Update Strategy

### Bug Fixes
- Fix immediately
- Submit new version (1.0.1)
- Push to beta first
- Then public release

### New Features
- Plan next version (1.1.0)
- Add feature requests from beta testers
- Test thoroughly in beta
- Submit after public feedback period

### Maintenance
- Monitor iOS/Android OS updates
- Test on new devices
- Update dependencies regularly
- Keep privacy policy current

---

## Checklist Summary

### Before First Submission
- [ ] App builds without errors
- [ ] No crashes on real devices
- [ ] All features work
- [ ] App icons present
- [ ] Splash screen configured
- [ ] Privacy policy written
- [ ] Screenshots created
- [ ] App description written
- [ ] Version set to 1.0.0
- [ ] Tested offline functionality

### iOS Specific
- [ ] Xcode project configured
- [ ] App signing setup
- [ ] App capabilities defined
- [ ] Screenshots in 1242x2208 size
- [ ] 1024x1024 icon created
- [ ] Privacy policy URL valid

### Android Specific
- [ ] Gradle build configured
- [ ] Keystore created
- [ ] Signing key backed up
- [ ] Screenshots in 1080x1920 size
- [ ] 512x512 icon created
- [ ] Release notes written

---

## Timeline

**Week 1**: Final testing, beta launch
**Week 2**: Beta feedback, bug fixes
**Week 3**: Public release preparation
**Week 4**: Submit to stores

**From Submission**:
- iOS: 24-48 hours to approval
- Android: 2-4 hours to approval
- After approval: Immediate availability in stores

---

## Conclusion

CommandLineTodo is now ready for production release!

✅ **Complete Functionality**: All features implemented
✅ **Professional UI**: Beautiful, polished interface
✅ **Performance**: Optimized and tested
✅ **Privacy**: User data protected
✅ **Documentation**: Clear and accessible

**Next Step**: Follow this guide to submit your app to the App Store and Google Play!

---

**Status**: Ready for Store Submission ✅
**Last Updated**: Phase 5 - Complete
