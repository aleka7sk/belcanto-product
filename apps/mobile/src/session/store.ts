import * as SecureStore from "expo-secure-store";
import { Platform } from "react-native";

import { decodeSessionTokens, type SessionTokens } from "@/api/contracts";

const SESSION_KEY = "belcanto.session.v1";
const SESSION_VERSION = 1;

export interface SecureKeyValueStore {
  getItemAsync(key: string): Promise<string | null>;
  setItemAsync(key: string, value: string): Promise<void>;
  deleteItemAsync(key: string): Promise<void>;
}

export function createMemorySessionKeyValueStore(): SecureKeyValueStore {
  const values = new Map<string, string>();
  return {
    getItemAsync: async (key) => values.get(key) ?? null,
    setItemAsync: async (key, value) => {
      values.set(key, value);
    },
    deleteItemAsync: async (key) => {
      values.delete(key);
    },
  };
}

export const expoSecureKeyValueStore: SecureKeyValueStore = {
  getItemAsync: (key) => SecureStore.getItemAsync(key),
  setItemAsync: (key, value) =>
    SecureStore.setItemAsync(key, value, {
      keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
    }),
  deleteItemAsync: (key) => SecureStore.deleteItemAsync(key),
};

/**
 * Web sessions intentionally survive only for the lifetime of this JavaScript
 * context. Persisting bearer and refresh tokens in browser storage would turn
 * an XSS defect into a durable credential leak.
 */
export const memoryOnlyWebKeyValueStore = createMemorySessionKeyValueStore();

export function selectSessionKeyValueStore(
  platform: string,
  nativeStore: SecureKeyValueStore = expoSecureKeyValueStore,
  webStore: SecureKeyValueStore = memoryOnlyWebKeyValueStore,
): SecureKeyValueStore {
  return platform === "web" ? webStore : nativeStore;
}

export const defaultSessionKeyValueStore: SecureKeyValueStore =
  selectSessionKeyValueStore(Platform.OS);

export interface SessionStore {
  load(): Promise<SessionTokens | null>;
  save(tokens: SessionTokens): Promise<void>;
  clear(): Promise<void>;
}

function decodeStoredSession(value: string): SessionTokens {
  const parsed = JSON.parse(value) as unknown;
  if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) {
    throw new TypeError("Stored session must be an object");
  }
  const record = parsed as Record<string, unknown>;
  if (record.version !== SESSION_VERSION) {
    throw new TypeError("Stored session version is unsupported");
  }
  return decodeSessionTokens(record.tokens);
}

export function createSessionStore(
  storage: SecureKeyValueStore = defaultSessionKeyValueStore,
): SessionStore {
  let tail: Promise<void> = Promise.resolve();
  const serialize = <Value>(operation: () => Promise<Value>): Promise<Value> => {
    const result = tail.then(operation, operation);
    tail = result.then(
      () => undefined,
      () => undefined,
    );
    return result;
  };
  return {
    load: () =>
      serialize(async () => {
        const stored = await storage.getItemAsync(SESSION_KEY);
        if (stored === null) {
          return null;
        }
        try {
          return decodeStoredSession(stored);
        } catch {
          await storage.deleteItemAsync(SESSION_KEY);
          return null;
        }
      }),
    save: (tokens) =>
      serialize(async () => {
        const validated = decodeSessionTokens(tokens);
        await storage.setItemAsync(
          SESSION_KEY,
          JSON.stringify({ version: SESSION_VERSION, tokens: validated }),
        );
      }),
    clear() {
      return serialize(() => storage.deleteItemAsync(SESSION_KEY));
    },
  };
}

function expiresAfter(value: string, now: Date, skewMs: number): boolean {
  return Date.parse(value) > now.getTime() + skewMs;
}

export function isAccessTokenUsable(
  tokens: SessionTokens,
  now = new Date(),
  skewMs = 30_000,
): boolean {
  return expiresAfter(tokens.accessExpiresAt, now, skewMs);
}

export function isRefreshTokenUsable(
  tokens: SessionTokens,
  now = new Date(),
  skewMs = 30_000,
): boolean {
  return expiresAfter(tokens.refreshExpiresAt, now, skewMs);
}
