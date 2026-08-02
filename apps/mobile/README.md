# Belcanto Mobile

Expo / React Native client for the closed-access B.0 flow. The app has no
public sign-up: a prepared Owner, Administrator, Teacher, or Student account is
activated through a one-time invitation and then signs in normally.

## Prerequisites

- Node.js `24.14.0` and pnpm `11.7.0`;
- Xcode for iOS or Android Studio for Android;
- the Belcanto API and PostgreSQL from [`../api/README.md`](../api/README.md).

Install the workspace dependencies from the repository root:

```sh
corepack enable
pnpm install --frozen-lockfile
```

## Local configuration

Create the local Expo environment file:

```sh
cp apps/mobile/.env.example apps/mobile/.env.local
```

For an iOS simulator, the default API origin is `http://localhost:8080`. An
Android emulator reaches the computer through its reserved host address:

```text
EXPO_PUBLIC_API_BASE_URL=http://10.0.2.2:8080
```

For a physical phone, replace the host with the computer's private LAN
address, for example:

```text
EXPO_PUBLIC_API_BASE_URL=http://192.168.1.20:8080
```

The phone and computer must be on the same trusted network, the API must listen
on a reachable interface, and the local firewall must allow the API port. Do
not use a private HTTP origin in production.

## Run

Welcome and sign-in can be inspected with the ordinary Expo development
server:

```sh
pnpm --filter @belcanto/mobile exec expo start --lan
```

The complete invitation flow uses the development-only `belcanto://` link.
Install a native development build before testing that link; Expo Go does not
own the app's custom scheme:

```sh
# macOS with Xcode
pnpm --filter @belcanto/mobile exec expo run:ios --device

# Android SDK installed
pnpm --filter @belcanto/mobile exec expo run:android --device
```

Each `expo run:*` command builds the native debug app, starts Metro, installs
the app, and opens it on the selected device. The API used for this local flow
must run with these development values in addition to its database and token
settings:

```text
APP_ENV=development
PUBLIC_ACTIVATION_BASE_URL=belcanto://activate
```

If the app was installed or opened directly from Xcode, start Metro explicitly
and open the generated QR link on the device. Opening the app before this link
is delivered produces React Native's `No script URL provided` screen:

```sh
pnpm --filter @belcanto/mobile exec expo start --dev-client --lan --clear
```

Regenerate an existing ignored native project after changing Expo config,
entitlements, or patched native dependencies. This is required once after
pulling a commit that changes those inputs because `expo run:ios` otherwise
reuses the existing `apps/mobile/ios/` directory:

```sh
pnpm install --frozen-lockfile
pnpm --filter @belcanto/mobile exec expo prebuild --clean --platform ios
pnpm --filter @belcanto/mobile exec expo run:ios --device
```

The development config intentionally omits Associated Domains so a local iOS
build can be signed by an Apple Personal Team. Production keeps the verified
HTTPS app-link entitlement.

Follow the controlled Owner and staff bootstrap instructions in
[`../api/README.md`](../api/README.md), then open the one-time link on the
installed native build.

## Checks

Run the client quality gate from the repository root:

```sh
pnpm mobile:check
pnpm --filter @belcanto/mobile doctor
```

Production builds require an HTTPS API origin, release bundle/package IDs, and
verified HTTPS activation links. The association-file deployment contract is
documented in [`deploy/app-links/README.md`](deploy/app-links/README.md).
