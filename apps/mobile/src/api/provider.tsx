import {
  createContext,
  useContext,
  type PropsWithChildren,
} from "react";

import { ApiClient } from "./client";

const ApiClientContext = createContext<ApiClient | null>(null);

export interface ApiClientProviderProps extends PropsWithChildren {
  client: ApiClient;
}

export function ApiClientProvider({
  children,
  client,
}: ApiClientProviderProps) {
  return (
    <ApiClientContext.Provider value={client}>
      {children}
    </ApiClientContext.Provider>
  );
}

export function useApiClient(): ApiClient {
  const client = useContext(ApiClientContext);
  if (client === null) {
    throw new Error("useApiClient must be used within ApiClientProvider");
  }
  return client;
}
