# React Native Skills (via vercel-labs/agent-skills)

---
name: react-native-skills
description: >
  Use when building React Native mobile apps — covers navigation, native APIs, styling
  with StyleSheet and NativeWind, platform differences, Expo workflow, performance
  patterns, and testing for iOS and Android.
license: MIT
source: https://github.com/vercel-labs/agent-skills
---

## When to Use

Activate this skill when the task involves:

- Scaffolding or extending a React Native (bare or Expo) application
- Implementing navigation with React Navigation (stack, tab, drawer)
- Using native device APIs (camera, location, push notifications, biometrics)
- Styling components with StyleSheet or NativeWind (Tailwind for RN)
- Diagnosing platform-specific behavior differences between iOS and Android
- Optimizing FlatList rendering for large data sets
- Writing and running tests with Jest and React Native Testing Library

## Project Setup

### Expo (Recommended for New Projects)

```bash
npx create-expo-app my-app --template blank-typescript
cd my-app
npx expo start
```

### Bare React Native

```bash
npx react-native@latest init MyApp --template react-native-template-typescript
```

## Navigation with React Navigation

```bash
npm install @react-navigation/native @react-navigation/native-stack
npm install react-native-screens react-native-safe-area-context
```

### Stack + Tab Setup

```tsx
import { NavigationContainer } from "@react-navigation/native";
import { createNativeStackNavigator } from "@react-navigation/native-stack";
import { createBottomTabNavigator } from "@react-navigation/bottom-tabs";

type RootStackParams = {
  Main: undefined;
  Detail: { id: string };
};

const Stack = createNativeStackNavigator<RootStackParams>();
const Tab = createBottomTabNavigator();

function MainTabs() {
  return (
    <Tab.Navigator>
      <Tab.Screen name="Home" component={HomeScreen} />
      <Tab.Screen name="Profile" component={ProfileScreen} />
    </Tab.Navigator>
  );
}

export default function App() {
  return (
    <NavigationContainer>
      <Stack.Navigator>
        <Stack.Screen name="Main" component={MainTabs} options={{ headerShown: false }} />
        <Stack.Screen name="Detail" component={DetailScreen} />
      </Stack.Navigator>
    </NavigationContainer>
  );
}
```

### Navigate and Pass Params

```tsx
// Navigate
navigation.navigate("Detail", { id: "123" });
navigation.push("Detail", { id: "456" }); // push new instance even if already in stack
navigation.goBack();

// Receive params
function DetailScreen({ route }: { route: RouteProp<RootStackParams, "Detail"> }) {
  const { id } = route.params;
}
```

## Styling

### StyleSheet API

```tsx
import { StyleSheet, View, Text } from "react-native";

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#fff",
    alignItems: "center",
    justifyContent: "center",
  },
  title: {
    fontSize: 24,
    fontWeight: "700",
    color: "#111827",
  },
});

export function Screen() {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>Hello</Text>
    </View>
  );
}
```

### Platform-Specific Styling

```tsx
import { Platform, StyleSheet } from "react-native";

const styles = StyleSheet.create({
  header: {
    paddingTop: Platform.OS === "ios" ? 44 : 24,
    backgroundColor: Platform.select({ ios: "#fff", android: "#f5f5f5" }),
  },
});
```

## Native APIs (Expo)

### Camera

```bash
npx expo install expo-camera
```

```tsx
import { CameraView, useCameraPermissions } from "expo-camera";

function CameraScreen() {
  const [permission, requestPermission] = useCameraPermissions();

  if (!permission?.granted) {
    return <Button title="Grant Camera Access" onPress={requestPermission} />;
  }

  return <CameraView style={{ flex: 1 }} facing="back" />;
}
```

### Location

```bash
npx expo install expo-location
```

```tsx
import * as Location from "expo-location";

async function getCurrentLocation() {
  const { status } = await Location.requestForegroundPermissionsAsync();
  if (status !== "granted") return null;

  const location = await Location.getCurrentPositionAsync({
    accuracy: Location.Accuracy.Balanced,
  });
  return { lat: location.coords.latitude, lng: location.coords.longitude };
}
```

## Performance: FlatList

```tsx
import { FlatList, ListRenderItem } from "react-native";

const renderItem: ListRenderItem<Item> = ({ item }) => <ItemCard item={item} />;
const keyExtractor = (item: Item) => item.id;

function ItemList({ data }: { data: Item[] }) {
  return (
    <FlatList
      data={data}
      renderItem={renderItem}
      keyExtractor={keyExtractor}
      initialNumToRender={10}
      maxToRenderPerBatch={10}
      windowSize={5}
      removeClippedSubviews={true}
      getItemLayout={(_, index) => ({ length: 72, offset: 72 * index, index })}
    />
  );
}
```

Use `getItemLayout` whenever items have a fixed height — it removes the measurement pass.

## Testing

```bash
npm install --save-dev @testing-library/react-native jest-expo
```

```tsx
import { render, fireEvent } from "@testing-library/react-native";
import { Counter } from "../Counter";

test("increments count on press", () => {
  const { getByText } = render(<Counter />);
  fireEvent.press(getByText("Increment"));
  expect(getByText("Count: 1")).toBeTruthy();
});
```

## Reference Files

| Task | Reference File |
|------|---------------|
| Expo EAS Build, OTA updates, app signing | `references/eas-build.md` |
| Gestures, animations with Reanimated | `references/animations.md` |

## Common Gotchas

- **No CSS box model** — RN uses Yoga (Flexbox); `display: grid` and many CSS properties don't exist
- **Text must be in `<Text>`** — any string outside a `<Text>` component throws a runtime error
- **`flex: 1` not always enough** — the parent also needs `flex: 1` or an explicit height for children to fill space
- **Keyboard avoiding** — wrap forms in `<KeyboardAvoidingView behavior={Platform.OS === "ios" ? "padding" : "height"}>` to prevent keyboard overlap
- **Metro cache** — clear cache with `npx expo start --clear` or `react-native start --reset-cache` when imports seem stale
- **SafeAreaView vs SafeAreaProvider** — use `react-native-safe-area-context`'s `SafeAreaView`, not RN's built-in (which doesn't handle Android correctly)
