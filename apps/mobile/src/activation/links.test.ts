import {
  createActivationLinkObserver,
  parseCanonicalActivationLink,
  parseActivationLink,
  type LinkingAdapter,
} from "./links";

const token = "A".repeat(43);

describe("activation links", () => {
  it("reads the server fragment form", () => {
    expect(
      parseActivationLink(`https://app.example/activate#token=${token}`, {
        allowedHttpsOrigins: ["https://app.example"],
      }),
    ).toEqual({ ok: true, token });
  });

  it("reads a custom-scheme query form", () => {
    expect(
      parseActivationLink(`belcanto://activate?token=${token}`, {
        allowCustomScheme: true,
      }),
    ).toEqual({ ok: true, token });
  });

  it("fails closed for duplicate tokens and unrelated routes", () => {
    expect(
      parseActivationLink(
        `https://app.example/activate?token=${token}#token=${token}`,
        { allowedHttpsOrigins: ["https://app.example"] },
      ),
    ).toEqual({ ok: false, reason: "multiple_tokens" });
    expect(
      parseActivationLink(`https://app.example/sign-in#token=${token}`, {
        allowedHttpsOrigins: ["https://app.example"],
      }),
    ).toEqual({ ok: false, reason: "wrong_route" });
  });

  it("rejects dangerous protocols and HTTPS origins outside policy", () => {
    expect(parseActivationLink(`javascript://activate?token=${token}`)).toEqual({
      ok: false,
      reason: "unsupported_protocol",
    });
    expect(
      parseActivationLink(`https://app.example/activate#token=${token}`),
    ).toEqual({ ok: false, reason: "untrusted_origin" });
    expect(
      parseActivationLink(`https://evil.example/activate#token=${token}`, {
        allowedHttpsOrigins: ["https://app.example"],
      }),
    ).toEqual({ ok: false, reason: "untrusted_origin" });
    expect(
      parseActivationLink(`https://app.example/activate#token=${token}`, {
        allowedHttpsOrigins: ["https://app.example"],
      }),
    ).toEqual({ ok: true, token });
  });

  it("keeps the custom scheme development-only", () => {
    expect(parseActivationLink(`belcanto://activate#token=${token}`)).toEqual({
      ok: false,
      reason: "custom_scheme_disabled",
    });
    expect(
      parseActivationLink(`belcanto://activate#token=${token}`, {
        allowCustomScheme: true,
      }),
    ).toEqual({ ok: true, token });
  });

  it("requires the exact raw canonical server-response form", () => {
    const policy = {
      allowedHttpsOrigins: ["https://app.example"],
      allowCustomScheme: true,
    };
    expect(
      parseCanonicalActivationLink(
        `https://app.example/activate#token=${token}`,
        policy,
      ),
    ).toEqual({ ok: true, token });
    expect(
      parseCanonicalActivationLink(`belcanto://activate#token=${token}`, policy),
    ).toEqual({ ok: true, token });
    for (const value of [
      `belcanto://activate/#token=${token}`,
      ` https://app.example/activate#token=${token}`,
      `https://app.example/activate#token=${token} `,
      `HTTPS://app.example/activate#token=${token}`,
      `https://app.example/activate?token=${token}`,
      `https://app.example/activate#token=%${token.slice(1)}`,
    ]) {
      expect(parseCanonicalActivationLink(value, policy)).toEqual({
        ok: false,
        reason: "invalid_url",
      });
    }
  });

  it("requires exact activation targets", () => {
    const policy = { allowedHttpsOrigins: ["https://app.example"] };
    expect(
      parseActivationLink(`belcanto://evil/activate?token=${token}`, {
        allowCustomScheme: true,
      }),
    ).toEqual({ ok: false, reason: "wrong_route" });
    expect(
      parseActivationLink(
        `https://app.example/nested/activate#token=${token}`,
        policy,
      ),
    ).toEqual({ ok: false, reason: "wrong_route" });
    expect(
      parseActivationLink(
        `https://user@app.example/activate#token=${token}`,
        policy,
      ),
    ).toEqual({ ok: false, reason: "wrong_route" });
    expect(
      parseActivationLink(`https://app.example:8443/activate#token=${token}`, policy),
    ).toEqual({ ok: false, reason: "untrusted_origin" });
    expect(
      parseActivationLink(`https://app.example:8443/activate#token=${token}`, {
        allowedHttpsOrigins: ["https://app.example:8443"],
      }),
    ).toEqual({ ok: true, token });
  });

  it("observes initial and subsequent links through an injected adapter", async () => {
    let listener: ((event: { url: string }) => void) | undefined;
    const adapter: LinkingAdapter = {
      getInitialURL: async () => `https://app.example/activate#token=${token}`,
      addEventListener: (_event, next) => {
        listener = next;
        return { remove: jest.fn() };
      },
    };
    const observer = createActivationLinkObserver(adapter, {
      allowedHttpsOrigins: ["https://app.example"],
    });
    await expect(observer.readInitial()).resolves.toEqual({ ok: true, token });
    const received: unknown[] = [];
    observer.subscribe((result) => received.push(result));
    listener?.({ url: "https://app.example/activate#token=bad" });
    expect(received).toEqual([{ ok: false, reason: "invalid_token" }]);
  });
});
