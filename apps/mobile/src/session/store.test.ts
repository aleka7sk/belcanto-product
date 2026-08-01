import type { IsoDateTime, SessionTokens } from "@/api/contracts";

import {
  createSessionStore,
  isAccessTokenUsable,
  isRefreshTokenUsable,
  type SecureKeyValueStore,
} from "./store";

const tokens: SessionTokens = {
  accessToken: "A".repeat(43),
  refreshToken: "R".repeat(43),
  accessExpiresAt: "2026-08-01T10:00:00Z" as IsoDateTime,
  refreshExpiresAt: "2026-09-01T10:00:00Z" as IsoDateTime,
};

function memoryStorage(initial: string | null = null) {
  let value = initial;
  const storage: SecureKeyValueStore = {
    getItemAsync: async () => value,
    setItemAsync: async (_key, next) => {
      value = next;
    },
    deleteItemAsync: async () => {
      value = null;
    },
  };
  return { storage, read: () => value };
}

describe("secure session store", () => {
  it("round-trips validated session tokens", async () => {
    const memory = memoryStorage();
    const store = createSessionStore(memory.storage);
    await store.save(tokens);
    await expect(store.load()).resolves.toEqual(tokens);
    expect(memory.read()).toContain('"version":1');
  });

  it("removes corrupt stored material", async () => {
    const memory = memoryStorage("not-json");
    const store = createSessionStore(memory.storage);
    await expect(store.load()).resolves.toBeNull();
    expect(memory.read()).toBeNull();
  });

  it("applies an expiry skew to access and refresh tokens", () => {
    const now = new Date("2026-08-01T09:59:40Z");
    expect(isAccessTokenUsable(tokens, now)).toBe(false);
    expect(isRefreshTokenUsable(tokens, now)).toBe(true);
  });

  it("orders corrupt-session repair before a concurrent valid save", async () => {
    let persisted: string | null = "corrupt";
    let deleteStarted: (() => void) | undefined;
    let releaseDelete: (() => void) | undefined;
    const started = new Promise<void>((resolve) => {
      deleteStarted = resolve;
    });
    const release = new Promise<void>((resolve) => {
      releaseDelete = resolve;
    });
    const storage: SecureKeyValueStore = {
      getItemAsync: async () => persisted,
      setItemAsync: async (_key, value) => {
        persisted = value;
      },
      deleteItemAsync: async () => {
        deleteStarted?.();
        await release;
        persisted = null;
      },
    };
    const store = createSessionStore(storage);
    const load = store.load();
    await started;
    const save = store.save(tokens);
    releaseDelete?.();
    await expect(load).resolves.toBeNull();
    await save;
    expect(persisted).toContain('"version":1');
    await expect(store.load()).resolves.toEqual(tokens);
  });
});
