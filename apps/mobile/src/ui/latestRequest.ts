export interface LatestRequestToken {
  signal: AbortSignal;
  isCurrent(): boolean;
}

export interface LatestRequestGuard {
  begin(): LatestRequestToken;
  cancel(): void;
}

/** Cancels the prior request and rejects late results from superseded requests. */
export function createLatestRequestGuard(): LatestRequestGuard {
  let generation = 0;
  let controller: AbortController | null = null;
  return {
    begin() {
      controller?.abort("superseded");
      controller = new AbortController();
      generation += 1;
      const requestGeneration = generation;
      const requestController = controller;
      return {
        signal: requestController.signal,
        isCurrent: () =>
          generation === requestGeneration &&
          controller === requestController &&
          !requestController.signal.aborted,
      };
    },
    cancel() {
      controller?.abort("cancelled");
      controller = null;
      generation += 1;
    },
  };
}
