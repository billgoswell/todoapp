# CommandLineTodo Mobile App

A React Native mobile application for the CommandLineTodo system with full offline functionality and real-time synchronization across devices.

## Features

- ✅ **Full Offline Support** - Complete todo management works without internet
- ✅ **Real-Time Sync** - Automatic synchronization when online
- ✅ **Cross-Platform** - Works on iOS and Android with single codebase
- ✅ **Conflict Resolution** - Automatic last-write-wins conflict handling
- ✅ **Multiple Lists** - Organize tasks into multiple todo lists
- ✅ **Task Management** - Create, edit, complete, delete tasks
- ✅ **Priority & Dates** - Set task priority and due dates
- ✅ **Background Sync** - Sync even when app is in background

## Tech Stack

- **Framework**: React Native 0.73.0
- **State Management**: React Context API + Hooks
- **Local Database**: SQLite
- **HTTP Client**: Axios
- **Navigation**: React Navigation
- **Sync Protocol**: REST API with timestamp-based pull/push

## Project Structure

```
src/
├── api/                    # HTTP client and API calls
├── database/              # SQLite database and repositories
├── sync/                  # Sync engine and conflict resolution
├── state/                 # Global state management
├── screens/               # App screens (HomeScreen, TaskDetail, etc)
├── components/            # Reusable UI components
├── navigation/            # Navigation configuration
├── utils/                 # Utility functions
└── App.tsx                # Root component
```

## Getting Started

### Prerequisites

- Node.js 18+ and npm/yarn
- iOS development: Xcode 14+, CocoaPods
- Android development: Android Studio, JDK 11+
- React Native CLI

### Installation

1. Clone the repository:
```bash
cd /home/bill/personal/workspace/commandlinetodo-mobile
```

2. Install dependencies:
```bash
npm install
```

3. Install iOS pods (macOS only):
```bash
cd ios && pod install && cd ..
```

4. Set up environment:
```bash
cp .env.example .env
```

## Development

### Start Development Server
```bash
npm start
```

### Run on iOS (macOS only)
```bash
npm run ios
```

### Run on Android
```bash
npm run android
```

### Type Checking
```bash
npm run type-check
```

### Linting
```bash
npm run lint
```

### Run Tests
```bash
npm test
```

## Development Phases

See [PLAN.md](./PLAN.md) for detailed implementation roadmap:

- **Phase 1 (Week 1-2)**: Foundation - local database & basic UI
- **Phase 2 (Week 2-3)**: Full offline functionality - CRUD operations
- **Phase 3 (Week 3-4)**: API integration - connect to server
- **Phase 4 (Week 4-5)**: Sync engine - bidirectional synchronization
- **Phase 5 (Week 5-6)**: Polish - background sync & UX improvements
- **Phase 6 (Week 6-7)**: Testing & launch - production release

## Configuration

### Server Configuration

Update the server URL and API key in Settings screen:

1. Open the Settings screen in the app
2. Enter your server URL (e.g., `https://api.example.com/api/v1`)
3. Enter your API key
4. Tap "Test Connection" to verify

Or set environment variables:
```bash
SERVER_URL=https://api.example.com/api/v1
API_KEY=your-api-key-here
```

## API Integration

The app communicates with the CommandLineTodo server using these endpoints:

### Sync Protocol
- `POST /api/v1/sync/pull` - Get changes since timestamp
- `POST /api/v1/sync/push` - Push local changes

### Tasks
- `GET /api/v1/tasks` - Get all tasks
- `POST /api/v1/tasks` - Create task
- `PUT /api/v1/tasks/:id` - Update task
- `DELETE /api/v1/tasks/:id` - Delete task

### Lists
- `GET /api/v1/lists` - Get all lists
- `POST /api/v1/lists` - Create list
- `PUT /api/v1/lists/:id` - Update list
- `DELETE /api/v1/lists/:id` - Delete list

See [PLAN.md](./PLAN.md) for detailed API documentation.

## Database Schema

SQLite local database includes:
- `todo_lists` - User's todo lists
- `tasks` - Tasks in each list
- `sync_metadata` - Sync state tracking
- `change_log` - Track pending changes for push sync

See [PLAN.md](./PLAN.md) for complete schema details.

## Architecture

```
┌─────────────────────────────────────┐
│      React Native UI Layer          │
│  (Screens & Components)             │
└────────────────┬────────────────────┘
                 │
┌────────────────▼────────────────────┐
│   State Management (Context + Hooks)│
└────────────────┬────────────────────┘
                 │
┌────────────────▼────────────────────┐
│   Business Logic                    │
│   (Sync Engine, Repositories)       │
└────────────────┬────────────────────┘
                 │
    ┌────────────┴────────────┐
    │                         │
┌───▼────────────┐   ┌────────▼──────┐
│  SQLite DB     │   │  HTTP Client   │
│  (Offline)     │   │  (Server Sync) │
└────────────────┘   └────────────────┘
```

## Testing

### Unit Tests
```bash
npm test -- --testPathPattern="unit"
```

### Integration Tests
```bash
npm test -- --testPathPattern="integration"
```

### Test Coverage
```bash
npm test -- --coverage
```

## Troubleshooting

### Android Issues

**Error: `local.properties` not found**
```bash
cd android
echo "sdk.dir=$(which android-sdk)" > local.properties
cd ..
```

**Gradle issues**
```bash
cd android
./gradlew clean
cd ..
npm run android
```

### iOS Issues

**Pod install fails**
```bash
cd ios
rm Podfile.lock
pod install
cd ..
```

**Build fails with "module not found"**
```bash
npm install
cd ios && pod install && cd ..
npm run ios
```

### Sync Issues

**API key not working**
- Verify API key in Settings
- Check server is running
- Ensure server URL is correct

**Changes not syncing**
- Check device has internet connection
- Verify server is reachable
- Check app logs for errors

## Performance Tips

- Use React.memo for expensive components
- Implement lazy loading for large task lists
- Optimize database queries with proper indexes
- Profile with React DevTools Profiler

## Security

- API keys are stored in AsyncStorage (should use Keychain/Keystore for production)
- All API calls use HTTPS
- Input validation on all user data
- Sensitive data is not logged

## Contributing

1. Create a feature branch: `git checkout -b feature/your-feature`
2. Commit changes: `git commit -am 'Add feature'`
3. Push to branch: `git push origin feature/your-feature`
4. Submit a pull request

## License

This project is part of the CommandLineTodo system.

## Support

For issues and questions:
1. Check [PLAN.md](./PLAN.md) for detailed documentation
2. Review troubleshooting section above
3. Check server logs for API errors
4. Open an issue with error details

## Related Projects

- **CLI App**: Command-line todo application
- **Server**: Go REST API backend (PostgreSQL)
- **Documentation**: Complete sync protocol and API specs

## Roadmap

### MVP Features (Weeks 1-5)
- ✅ Local offline todo management
- ✅ SQLite database
- ✅ Task and list CRUD
- ✅ Server synchronization
- ✅ Conflict resolution

### Post-MVP Features (Future)
- Search and filters
- Dark mode
- Subtasks and notes
- Attachments
- Shared lists (collaboration)
- Push notifications
- Task recurring
- Widgets
- Analytics
