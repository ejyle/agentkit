# EAS Build Reference

## Expo Application Services (EAS)

EAS Build provides cloud-based app compilation for iOS and Android without needing
local Xcode or Android Studio. EAS Submit handles app store delivery.

## Setup

```bash
npm install -g eas-cli
eas login
eas build:configure
```

This creates `eas.json` in your project root.

## eas.json Profiles

```json
{
  "cli": { "version": ">= 14.0.0" },
  "build": {
    "development": {
      "developmentClient": true,
      "distribution": "internal",
      "ios": { "simulator": true },
      "android": { "buildType": "apk" }
    },
    "preview": {
      "distribution": "internal",
      "android": { "buildType": "apk" },
      "ios": { "distribution": "internal" }
    },
    "production": {
      "android": { "buildType": "app-bundle" },
      "ios": { "distribution": "store" }
    }
  },
  "submit": {
    "production": {
      "ios": {
        "appleId": "dev@example.com",
        "ascAppId": "1234567890"
      },
      "android": {
        "serviceAccountKeyPath": "./service-account.json",
        "track": "internal"
      }
    }
  }
}
```

## Build Commands

```bash
# Build for development (with dev client)
eas build --profile development --platform ios
eas build --profile development --platform android

# Build for internal distribution (QR code install)
eas build --profile preview --platform all

# Build for app store
eas build --profile production --platform all

# Check build status
eas build:list

# Download build artifact
eas build:download --id <build-id>
```

## OTA Updates (EAS Update)

Push JavaScript updates without going through app store review:

```bash
npm install expo-updates
eas update:configure
```

```bash
# Publish update to a branch
eas update --branch preview --message "Fix login bug"

# Map branch to build profile
eas channel:edit preview --branch main
```

```typescript
// In your app — check for updates at launch
import * as Updates from "expo-updates";

async function checkForUpdate() {
  try {
    const update = await Updates.checkForUpdateAsync();
    if (update.isAvailable) {
      await Updates.fetchUpdateAsync();
      await Updates.reloadAsync();
    }
  } catch (err) {
    // Network or update error — continue with installed version
    console.warn("Update check failed:", err);
  }
}
```

## App Signing

EAS manages signing credentials in their secure key store. Run:

```bash
# Generate or upload credentials interactively
eas credentials
```

For CI/CD, set these environment variables:
- `EXPO_TOKEN` — EAS authentication token (`eas account:create` then `eas login --token`)

## Environment Variables in EAS

```bash
# Set a secret (not visible in build logs)
eas secret:create --scope project --name API_KEY --value "sk-..."

# List secrets
eas secret:list
```

Reference in `app.config.ts`:
```typescript
export default {
  expo: {
    extra: {
      apiKey: process.env.API_KEY,
    },
  },
};
```
