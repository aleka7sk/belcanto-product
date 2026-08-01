export interface SingleFlight<Value> {
  run(operation: () => Promise<Value>): Promise<Value>;
  isRunning(): boolean;
}

export function createSingleFlight<Value>(): SingleFlight<Value> {
  let active: Promise<Value> | null = null;
  return {
    run(operation) {
      if (active !== null) return active;
      const current = Promise.resolve().then(operation);
      active = current;
      const clear = () => {
        if (active === current) active = null;
      };
      void current.then(clear, clear);
      return current;
    },
    isRunning() {
      return active !== null;
    },
  };
}

export interface KeyedSingleFlight<Key, Value> {
  run(key: Key, operation: () => Promise<Value>): Promise<Value>;
  isRunning(): boolean;
}

export function createKeyedSingleFlight<Key, Value>(): KeyedSingleFlight<
  Key,
  Value
> {
  let active: { key: Key; promise: Promise<Value> } | null = null;
  const run = (key: Key, operation: () => Promise<Value>): Promise<Value> => {
    if (active !== null) {
      if (Object.is(active.key, key)) return active.promise;
      return active.promise.then(
        () => run(key, operation),
        () => run(key, operation),
      );
    }
    const promise = Promise.resolve().then(operation);
    const flight = { key, promise };
    active = flight;
    const clear = () => {
      if (active === flight) active = null;
    };
    void promise.then(clear, clear);
    return promise;
  };
  return {
    run,
    isRunning: () => active !== null,
  };
}

export interface OperationEpoch {
  current(): number;
  begin(): number;
  isCurrent(epoch: number): boolean;
}

export function createOperationEpoch(): OperationEpoch {
  let current = 0;
  return {
    current: () => current,
    begin() {
      current += 1;
      return current;
    },
    isCurrent: (epoch) => epoch === current,
  };
}

export interface SerialExecutor {
  run<Value>(operation: () => Promise<Value>): Promise<Value>;
}

export function createSerialExecutor(): SerialExecutor {
  let tail: Promise<void> = Promise.resolve();
  return {
    run(operation) {
      const result = tail.then(operation, operation);
      tail = result.then(
        () => undefined,
        () => undefined,
      );
      return result;
    },
  };
}
