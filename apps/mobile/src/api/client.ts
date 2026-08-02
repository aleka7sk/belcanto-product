import {
  ContractDecodeError,
  decodeApiErrorEnvelope,
  decodeInvitationResult,
  type ActivationPreview,
  type ActivationPreviewRequest,
  type BootstrapView,
  type CompleteActivationRequest,
  type CreateStudentRequest,
  type DelegationResult,
  type FirstMinute,
  type GrantDelegationRequest,
  type InvitationResult,
  type PublishFirstMinuteRequest,
  type RefreshSessionRequest,
  type RevokeDelegationRequest,
  type SessionTokens,
  type SignInRequest,
  type StaffMember,
  type StaffRole,
  type StudentOnboardingItem,
  type StudentResult,
} from "./contracts";
import { routes, type RouteDescriptor } from "./routes";
import type { ActivationLinkPolicy } from "@/activation/links";
import {
  isLoopbackHost,
  isPrivateDevelopmentHost,
} from "@/runtime/developmentOrigin";
import { isValidIdempotencyKey } from "@/validation/backend";

type Fetch = typeof globalThis.fetch;

const ERROR_STATUS_BY_CODE = {
  INVALID_INPUT: 422,
  UNAUTHENTICATED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  CONFLICT: 409,
  INVALID_STATE: 409,
  ACTIVATION_INVALID: 400,
  RATE_LIMITED: 429,
  UNAVAILABLE: 503,
  INTERNAL: 500,
} as const;

export interface ApiClientOptions {
  baseUrl: string;
  fetch?: Fetch;
  timeoutMs?: number;
  activationLinkPolicy?: ActivationLinkPolicy;
  allowInsecureDevelopmentOrigin?: boolean;
}

export interface RequestOptions {
  accessToken?: string | undefined;
  body?: unknown;
  idempotencyKey?: string | undefined;
  signal?: AbortSignal | undefined;
}

export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    message: string,
    readonly requestId?: string,
    override readonly cause?: unknown,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export class ApiTransportError extends Error {
  constructor(message: string, override readonly cause?: unknown) {
    super(message);
    this.name = "ApiTransportError";
  }
}

function normalizedBaseUrl(
  value: string,
  allowInsecureDevelopmentOrigin: boolean,
): string {
  const trimmed = value.trim().replace(/\/+$/, "");
  let parsed: URL;
  try {
    parsed = new URL(trimmed);
  } catch (error) {
    throw new TypeError("API base URL must be an absolute URL", { cause: error });
  }
  const developmentHttp =
    parsed.protocol === "http:" &&
    (isLoopbackHost(parsed.hostname) ||
      (allowInsecureDevelopmentOrigin &&
        isPrivateDevelopmentHost(parsed.hostname)));
  if (parsed.protocol !== "https:" && !developmentHttp) {
    throw new TypeError("API base URL must use HTTPS outside localhost");
  }
  if (
    parsed.username !== "" ||
    parsed.password !== "" ||
    parsed.pathname !== "/" ||
    parsed.search !== "" ||
    parsed.hash !== ""
  ) {
    throw new TypeError("API base URL must be an origin without credentials or path");
  }
  return trimmed;
}

async function parseJson(response: Response): Promise<unknown> {
  const text = await response.text();
  if (text.length === 0) {
    return undefined;
  }
  try {
    return JSON.parse(text) as unknown;
  } catch (error) {
    throw new ContractDecodeError("JSON", error instanceof Error ? error.message : "$");
  }
}

export class ApiClient {
  private readonly baseUrl: string;
  private readonly fetch: Fetch;
  private readonly timeoutMs: number;
  private readonly activationLinkPolicy: ActivationLinkPolicy;

  constructor(options: ApiClientOptions) {
    this.baseUrl = normalizedBaseUrl(
      options.baseUrl,
      options.allowInsecureDevelopmentOrigin === true,
    );
    this.fetch = options.fetch ?? globalThis.fetch;
    this.timeoutMs = options.timeoutMs ?? 15_000;
    this.activationLinkPolicy = {
      allowedHttpsOrigins: [
        ...(options.activationLinkPolicy?.allowedHttpsOrigins ?? []),
      ],
      allowCustomScheme: options.activationLinkPolicy?.allowCustomScheme === true,
    };
  }

  previewActivation(
    body: ActivationPreviewRequest,
    signal?: AbortSignal,
  ): Promise<ActivationPreview> {
    return this.request(routes.previewActivation, { body, signal });
  }

  completeActivation(
    body: CompleteActivationRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<void> {
    return this.request(routes.completeActivation, { body, idempotencyKey, signal });
  }

  signIn(body: SignInRequest, signal?: AbortSignal): Promise<SessionTokens> {
    return this.request(routes.signIn, { body, signal });
  }

  refreshSession(
    body: RefreshSessionRequest,
    signal?: AbortSignal,
  ): Promise<SessionTokens> {
    return this.request(routes.refreshSession, { body, signal });
  }

  signOut(accessToken: string, signal?: AbortSignal): Promise<void> {
    return this.request(routes.signOut, { accessToken, signal });
  }

  bootstrap(accessToken: string, signal?: AbortSignal): Promise<BootstrapView> {
    return this.request(routes.bootstrap, { accessToken, signal });
  }

  listStaff(
    accessToken: string,
    role: StaffRole,
    signal?: AbortSignal,
  ): Promise<StaffMember[]> {
    return this.request(routes.listStaff(role), { accessToken, signal });
  }

  listStudentOnboarding(
    accessToken: string,
    signal?: AbortSignal,
  ): Promise<StudentOnboardingItem[]> {
    return this.request(routes.listStudentOnboarding, { accessToken, signal });
  }

  grantDelegation(
    accessToken: string,
    body: GrantDelegationRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<DelegationResult> {
    return this.request(routes.grantDelegation, {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  revokeDelegation(
    accessToken: string,
    delegationId: string,
    body: RevokeDelegationRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<void> {
    return this.request(routes.revokeDelegation(delegationId), {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  createStudent(
    accessToken: string,
    body: CreateStudentRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<StudentResult> {
    return this.request(routes.createStudent, {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  publishFirstMinute(
    accessToken: string,
    studentId: string,
    body: PublishFirstMinuteRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<FirstMinute> {
    return this.request(routes.publishFirstMinute(studentId), {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  issueInvitation(
    accessToken: string,
    studentId: string,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<InvitationResult> {
    return this.request(
      routes.issueInvitation(studentId),
      { accessToken, idempotencyKey, signal },
      (value) => decodeInvitationResult(value, this.activationLinkPolicy),
    );
  }

  reissueInvitation(
    accessToken: string,
    studentId: string,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<InvitationResult> {
    return this.request(
      routes.reissueInvitation(studentId),
      { accessToken, idempotencyKey, signal },
      (value) => decodeInvitationResult(value, this.activationLinkPolicy),
    );
  }

  revokeInvitation(
    accessToken: string,
    invitationId: string,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<void> {
    return this.request(routes.revokeInvitation(invitationId), {
      accessToken,
      idempotencyKey,
      signal,
    });
  }

  private async request<ResponseBody>(
    route: RouteDescriptor<ResponseBody>,
    options: RequestOptions,
    decode: (value: unknown) => ResponseBody = route.decode,
  ): Promise<ResponseBody> {
    if (options.signal?.aborted) {
      throw new ApiTransportError("API request was aborted", options.signal.reason);
    }
    if (route.auth === "required" && !options.accessToken) {
      throw new TypeError(`Access token is required for ${route.method} ${route.path}`);
    }
    if (
      options.idempotencyKey !== undefined &&
      !isValidIdempotencyKey(options.idempotencyKey)
    ) {
      throw new TypeError(
        `Idempotency-Key is invalid for ${route.method} ${route.path}`,
      );
    }
    const headers: Record<string, string> = { Accept: "application/json" };
    if (options.body !== undefined) {
      headers["Content-Type"] = "application/json";
    }
    if (options.accessToken) {
      headers.Authorization = `Bearer ${options.accessToken}`;
    }
    if (options.idempotencyKey) {
      headers["Idempotency-Key"] = options.idempotencyKey;
    }

    const controller = new AbortController();
    const abort = () => controller.abort(options.signal?.reason);
    options.signal?.addEventListener("abort", abort, { once: true });
    const timeout = setTimeout(() => controller.abort("timeout"), this.timeoutMs);

    let response: Response;
    try {
      const init: RequestInit = {
        method: route.method,
        headers,
        credentials: "omit",
        signal: controller.signal,
      };
      if (options.body !== undefined) {
        init.body = JSON.stringify(options.body);
      }
      response = await this.fetch(`${this.baseUrl}${route.path}`, init);
    } catch (error) {
      throw new ApiTransportError("API request failed", error);
    } finally {
      clearTimeout(timeout);
      options.signal?.removeEventListener("abort", abort);
    }

    let payload: unknown;
    try {
      payload = await parseJson(response);
    } catch (error) {
      throw new ApiError(response.status, "UNEXPECTED_RESPONSE", "API returned invalid JSON", undefined, error);
    }

    if (!response.ok) {
      try {
        const envelope = decodeApiErrorEnvelope(payload);
        if (ERROR_STATUS_BY_CODE[envelope.error.code] !== response.status) {
          throw new ApiError(
            response.status,
            "UNEXPECTED_RESPONSE",
            "API error status does not match its error code",
            envelope.error.requestId,
          );
        }
        if (
          !route.errorStatuses.some(
            (expectedStatus) => expectedStatus === response.status,
          )
        ) {
          throw new ApiError(
            response.status,
            "UNEXPECTED_RESPONSE",
            "API returned an error that is not declared for this route",
            envelope.error.requestId,
          );
        }
        throw new ApiError(
          response.status,
          envelope.error.code,
          envelope.error.message,
          envelope.error.requestId,
        );
      } catch (error) {
        if (error instanceof ApiError) {
          throw error;
        }
        throw new ApiError(
          response.status,
          "UNEXPECTED_RESPONSE",
          "API returned an invalid error response",
          undefined,
          error,
        );
      }
    }

    if (response.status !== route.successStatus) {
      throw new ApiError(
        response.status,
        "UNEXPECTED_RESPONSE",
        `API returned ${response.status}; expected ${route.successStatus}`,
      );
    }

    try {
      return decode(payload);
    } catch (error) {
      throw new ApiError(
        response.status,
        "UNEXPECTED_RESPONSE",
        "API returned a response that does not match its contract",
        undefined,
        error,
      );
    }
  }
}
