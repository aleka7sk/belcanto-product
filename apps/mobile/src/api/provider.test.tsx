import type { ReactElement } from "react";

import { ApiClient } from "./client";
import { ApiClientProvider, useApiClient } from "./provider";

const { renderToStaticMarkup } = jest.requireActual<{
  renderToStaticMarkup(element: ReactElement): string;
}>("react-dom/server");

function ApiClientProbe({ expected }: { expected: ApiClient }) {
  expect(useApiClient()).toBe(expected);
  return null;
}

describe("ApiClientProvider", () => {
  const client = new ApiClient({ baseUrl: "http://localhost:8080" });

  it("provides the exact ApiClient instance", () => {
    renderToStaticMarkup(
      <ApiClientProvider client={client}>
        <ApiClientProbe expected={client} />
      </ApiClientProvider>,
    );
  });

  it("fails closed outside the provider", () => {
    expect(() =>
      renderToStaticMarkup(<ApiClientProbe expected={client} />),
    ).toThrow("useApiClient must be used within ApiClientProvider");
  });
});
