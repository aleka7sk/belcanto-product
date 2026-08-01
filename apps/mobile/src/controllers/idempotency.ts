import * as Crypto from "expo-crypto";

import { requiredIdempotencyKey } from "@/validation/backend";

export interface IdempotencyKeyFactory {
  create(): string;
}

export const expoIdempotencyKeyFactory: IdempotencyKeyFactory = {
  create: () => Crypto.randomUUID(),
};

export interface IntentIdempotency {
  key(): string;
  complete(): void;
  abandon(): void;
}

export function createIntentIdempotency(
  factory: IdempotencyKeyFactory = expoIdempotencyKeyFactory,
): IntentIdempotency {
  let current: string | null = null;
  return {
    key() {
      current ??= requiredIdempotencyKey(factory.create());
      return current;
    },
    complete() {
      current = null;
    },
    abandon() {
      current = null;
    },
  };
}

export interface IdempotentSubmission<Input, Command> {
  prepare(input: Input): Command;
  succeeded(): void;
  abandoned(): void;
}

export function createIdempotentSubmission<Input, Command>(
  prepare: (input: Input, idempotencyKey: string) => Command,
  factory: IdempotencyKeyFactory = expoIdempotencyKeyFactory,
): IdempotentSubmission<Input, Command> {
  const intent = createIntentIdempotency(factory);
  return {
    prepare(input) {
      return prepare(input, intent.key());
    },
    succeeded() {
      intent.complete();
    },
    abandoned() {
      intent.abandon();
    },
  };
}
