import {
  ContractDecodeError,
  decodeApiErrorEnvelope,
  decodeInvitationResult,
  type ActivationPreview,
  type ActivationPreviewRequest,
  type BootstrapView,
  type CompleteActivationRequest,
  type CreateLessonRequest,
  type CreateStudentRequest,
  type DelegationResult,
  type FirstMinute,
  type GrantDelegationRequest,
  type InvitationResult,
  type Lesson,
  type LessonListQuery,
  type PublishFirstMinuteRequest,
  type ReassignPrimaryTeachersRequest,
  type ReassignPrimaryTeachersResult,
  type RefreshSessionRequest,
  type ReplaceLessonTeachersRequest,
  type ReplaceLessonTeachersResult,
  type RevokeDelegationRequest,
  type ActivationCodeRequest,
  type ActivationContactRequest,
  type ActivationFinishRequest,
  type ActivationPasswordRequest,
  type ActivationProgressView,
  type ActivationTokenRequest,
  type CompletePasswordResetRequest,
  type ConfirmContactChangeRequest,
  type CurrentPasswordRequest,
  type AcceptPolicyRequest,
  type DataExportRequest,
  type DeletionRequest,
  type PolicyVersion,
  type PrivacySettings,
  type ProfileView,
  type UpdateProfileRequest,
  type CoreLessonSeries,
  type CreateLessonSeriesRequest,
  type CreateRoomRequest,
  type GenerateOccurrencesRequest,
  type Room,
  type SeriesGenerationResult,
  type CreateEventCategoryRequest,
  type CreateEventRequest,
  type CreateEventSeriesRequest,
  type EventCategory,
  type EventListWindow,
  type EventOccurrence,
  type EventSeries,
  type CreateRescheduleRequestRequest,
  type DecideRescheduleRequestRequest,
  type RescheduleRequest,
  type JournalDraftRequest,
  type LessonJournal,
  type ProgressEvidence,
  type PublishJournalRequest,
  type MediaObject,
  type MediaAccess,
  type CreateMediaRequest,
  type HomeworkAssignment,
  type CreateHomeworkRequest,
  type CancelHomeworkRequest,
  type MarkHomeworkTaskRequest,
  type SubmitHomeworkRequest,
  type ReviewHomeworkRequest,
  type AttendanceRecord,
  type MarkAttendanceRequest,
  type DisableTwofaRequest,
  type SignInOutcome,
  type StartContactChangeRequest,
  type TwofaCodeRequest,
  type TwofaEnrollment,
  type TwofaSignInRequest,
  type TwofaStatus,
  type VerifiedContact,
  type RequestPasswordResetRequest,
  type RevokeOtherSessionsResult,
  type RevokeSessionRequest,
  type SecurityEventsPage,
  type SecurityEventsQuery,
  type SessionDevice,
  type SessionTokens,
  type SignInRequest,
  type StaffMember,
  type StaffRole,
  type StudentDirectoryItem,
  type StudentDirectoryQuery,
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
  binaryBody?: Uint8Array | undefined;
  extraHeaders?: Record<string, string> | undefined;
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

  signIn(body: SignInRequest, signal?: AbortSignal): Promise<SignInOutcome> {
    return this.request(routes.signIn, { body, signal });
  }

  signInWithTwofa(
    body: TwofaSignInRequest,
    signal?: AbortSignal,
  ): Promise<SessionTokens> {
    return this.request(routes.signInWithTwofa, { body, signal });
  }

  activationProgress(
    body: ActivationTokenRequest,
    signal?: AbortSignal,
  ): Promise<ActivationProgressView> {
    return this.request(routes.activationProgress, { body, signal });
  }

  setActivationPassword(
    body: ActivationPasswordRequest,
    signal?: AbortSignal,
  ): Promise<void> {
    return this.request(routes.setActivationPassword, { body, signal });
  }

  startActivationContact(
    body: ActivationContactRequest,
    signal?: AbortSignal,
  ): Promise<void> {
    return this.request(routes.startActivationContact, { body, signal });
  }

  verifyActivationContact(
    body: ActivationCodeRequest,
    signal?: AbortSignal,
  ): Promise<void> {
    return this.request(routes.verifyActivationContact, { body, signal });
  }

  startActivationTwofa(
    body: ActivationTokenRequest,
    signal?: AbortSignal,
  ): Promise<TwofaEnrollment> {
    return this.request(routes.startActivationTwofa, { body, signal });
  }

  confirmActivationTwofa(
    body: ActivationCodeRequest,
    signal?: AbortSignal,
  ): Promise<string[]> {
    return this.request(routes.confirmActivationTwofa, { body, signal });
  }

  finishActivation(
    body: ActivationFinishRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<void> {
    return this.request(routes.finishActivation, { body, idempotencyKey, signal });
  }

  listMyContacts(
    accessToken: string,
    signal?: AbortSignal,
  ): Promise<VerifiedContact[]> {
    return this.request(routes.listMyContacts, { accessToken, signal });
  }

  startContactChange(
    accessToken: string,
    body: StartContactChangeRequest,
    signal?: AbortSignal,
  ): Promise<void> {
    return this.request(routes.startContactChange, { accessToken, body, signal });
  }

  confirmContactChange(
    accessToken: string,
    body: ConfirmContactChangeRequest,
    signal?: AbortSignal,
  ): Promise<VerifiedContact> {
    return this.request(routes.confirmContactChange, { accessToken, body, signal });
  }

  twofaStatus(accessToken: string, signal?: AbortSignal): Promise<TwofaStatus> {
    return this.request(routes.twofaStatus, { accessToken, signal });
  }

  startTwofaEnrollment(
    accessToken: string,
    body: CurrentPasswordRequest,
    signal?: AbortSignal,
  ): Promise<TwofaEnrollment> {
    return this.request(routes.startTwofaEnrollment, { accessToken, body, signal });
  }

  confirmTwofaEnrollment(
    accessToken: string,
    body: TwofaCodeRequest,
    signal?: AbortSignal,
  ): Promise<string[]> {
    return this.request(routes.confirmTwofaEnrollment, { accessToken, body, signal });
  }

  disableTwofa(
    accessToken: string,
    body: DisableTwofaRequest,
    signal?: AbortSignal,
  ): Promise<void> {
    return this.request(routes.disableTwofa, { accessToken, body, signal });
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

  requestPasswordReset(
    body: RequestPasswordResetRequest,
    signal?: AbortSignal,
  ): Promise<void> {
    return this.request(routes.requestPasswordReset, { body, signal });
  }

  completePasswordReset(
    body: CompletePasswordResetRequest,
    signal?: AbortSignal,
  ): Promise<void> {
    return this.request(routes.completePasswordReset, { body, signal });
  }

  listMySessions(
    accessToken: string,
    signal?: AbortSignal,
  ): Promise<SessionDevice[]> {
    return this.request(routes.listMySessions, { accessToken, signal });
  }

  revokeOtherSessions(
    accessToken: string,
    body: RevokeSessionRequest,
    signal?: AbortSignal,
  ): Promise<RevokeOtherSessionsResult> {
    return this.request(routes.revokeOtherSessions, { accessToken, body, signal });
  }

  revokeMySession(
    accessToken: string,
    sessionId: string,
    body: RevokeSessionRequest,
    signal?: AbortSignal,
  ): Promise<void> {
    return this.request(routes.revokeMySession(sessionId), {
      accessToken,
      body,
      signal,
    });
  }

  listSecurityEvents(
    accessToken: string,
    query: SecurityEventsQuery = {},
    signal?: AbortSignal,
  ): Promise<SecurityEventsPage> {
    return this.request(routes.listSecurityEvents(query), {
      accessToken,
      signal,
    });
  }

  myProfile(accessToken: string, signal?: AbortSignal): Promise<ProfileView> {
    return this.request(routes.myProfile, { accessToken, signal });
  }

  updateMyProfile(
    accessToken: string,
    body: UpdateProfileRequest,
    signal?: AbortSignal,
  ): Promise<ProfileView> {
    return this.request(routes.updateMyProfile, { accessToken, body, signal });
  }

  listPolicies(
    accessToken: string,
    signal?: AbortSignal,
  ): Promise<PolicyVersion[]> {
    return this.request(routes.listPolicies, { accessToken, signal });
  }

  acceptPolicy(
    accessToken: string,
    body: AcceptPolicyRequest,
    signal?: AbortSignal,
  ): Promise<void> {
    return this.request(routes.acceptPolicy, { accessToken, body, signal });
  }

  privacySettings(
    accessToken: string,
    signal?: AbortSignal,
  ): Promise<PrivacySettings> {
    return this.request(routes.privacySettings, { accessToken, signal });
  }

  updatePrivacySettings(
    accessToken: string,
    body: PrivacySettings,
    signal?: AbortSignal,
  ): Promise<PrivacySettings> {
    return this.request(routes.updatePrivacySettings, {
      accessToken,
      body,
      signal,
    });
  }

  listDataExports(
    accessToken: string,
    signal?: AbortSignal,
  ): Promise<DataExportRequest[]> {
    return this.request(routes.listDataExports, { accessToken, signal });
  }

  createDataExport(
    accessToken: string,
    body: CurrentPasswordRequest,
    signal?: AbortSignal,
  ): Promise<DataExportRequest> {
    return this.request(routes.createDataExport, { accessToken, body, signal });
  }

  deletionRequest(
    accessToken: string,
    signal?: AbortSignal,
  ): Promise<DeletionRequest> {
    return this.request(routes.deletionRequest, { accessToken, signal });
  }

  createDeletionRequest(
    accessToken: string,
    body: CurrentPasswordRequest,
    signal?: AbortSignal,
  ): Promise<DeletionRequest> {
    return this.request(routes.createDeletionRequest, {
      accessToken,
      body,
      signal,
    });
  }

  cancelDeletionRequest(
    accessToken: string,
    signal?: AbortSignal,
  ): Promise<DeletionRequest> {
    return this.request(routes.cancelDeletionRequest, { accessToken, signal });
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

  listStudents(
    accessToken: string,
    query: StudentDirectoryQuery = {},
    signal?: AbortSignal,
  ): Promise<StudentDirectoryItem[]> {
    return this.request(routes.listStudents(query), { accessToken, signal });
  }

  listLessons(
    accessToken: string,
    query: LessonListQuery,
    signal?: AbortSignal,
  ): Promise<Lesson[]> {
    return this.request(routes.listLessons(query), { accessToken, signal });
  }

  getLesson(
    accessToken: string,
    lessonId: string,
    signal?: AbortSignal,
  ): Promise<Lesson> {
    return this.request(routes.getLesson(lessonId), { accessToken, signal });
  }

  listRooms(accessToken: string, signal?: AbortSignal): Promise<Room[]> {
    return this.request(routes.listRooms, { accessToken, signal });
  }

  createRoom(
    accessToken: string,
    body: CreateRoomRequest,
    signal?: AbortSignal,
  ): Promise<Room> {
    return this.request(routes.createRoom, { accessToken, body, signal });
  }

  listLessonSeries(
    accessToken: string,
    signal?: AbortSignal,
  ): Promise<CoreLessonSeries[]> {
    return this.request(routes.listLessonSeries, { accessToken, signal });
  }

  createLessonSeries(
    accessToken: string,
    body: CreateLessonSeriesRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<CoreLessonSeries> {
    return this.request(routes.createLessonSeries, {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  getLessonSeries(
    accessToken: string,
    seriesId: string,
    signal?: AbortSignal,
  ): Promise<CoreLessonSeries> {
    return this.request(routes.getLessonSeries(seriesId), { accessToken, signal });
  }

  generateSeriesOccurrences(
    accessToken: string,
    seriesId: string,
    body: GenerateOccurrencesRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<SeriesGenerationResult> {
    return this.request(routes.generateSeriesOccurrences(seriesId), {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  listEventCategories(
    accessToken: string,
    signal?: AbortSignal,
  ): Promise<EventCategory[]> {
    return this.request(routes.listEventCategories, { accessToken, signal });
  }

  createEventCategory(
    accessToken: string,
    body: CreateEventCategoryRequest,
    signal?: AbortSignal,
  ): Promise<EventCategory> {
    return this.request(routes.createEventCategory, { accessToken, body, signal });
  }

  createEventSeries(
    accessToken: string,
    body: CreateEventSeriesRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<EventSeries> {
    return this.request(routes.createEventSeries, {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  generateEventSeriesOccurrences(
    accessToken: string,
    seriesId: string,
    body: GenerateOccurrencesRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<SeriesGenerationResult> {
    return this.request(routes.generateEventSeriesOccurrences(seriesId), {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  listEvents(
    accessToken: string,
    window: EventListWindow,
    signal?: AbortSignal,
  ): Promise<EventOccurrence[]> {
    return this.request(routes.listEvents(window), { accessToken, signal });
  }

  createEvent(
    accessToken: string,
    body: CreateEventRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<EventOccurrence> {
    return this.request(routes.createEvent, {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  getEvent(
    accessToken: string,
    occurrenceId: string,
    signal?: AbortSignal,
  ): Promise<EventOccurrence> {
    return this.request(routes.getEvent(occurrenceId), { accessToken, signal });
  }

  rsvpToEvent(
    accessToken: string,
    occurrenceId: string,
    signal?: AbortSignal,
  ): Promise<EventOccurrence> {
    return this.request(routes.rsvpToEvent(occurrenceId), { accessToken, signal });
  }

  cancelEventRsvp(
    accessToken: string,
    occurrenceId: string,
    signal?: AbortSignal,
  ): Promise<EventOccurrence> {
    return this.request(routes.cancelEventRsvp(occurrenceId), {
      accessToken,
      signal,
    });
  }

  joinEventWaitlist(
    accessToken: string,
    occurrenceId: string,
    signal?: AbortSignal,
  ): Promise<EventOccurrence> {
    return this.request(routes.joinEventWaitlist(occurrenceId), {
      accessToken,
      signal,
    });
  }

  leaveEventWaitlist(
    accessToken: string,
    occurrenceId: string,
    signal?: AbortSignal,
  ): Promise<EventOccurrence> {
    return this.request(routes.leaveEventWaitlist(occurrenceId), {
      accessToken,
      signal,
    });
  }

  confirmSpotOffer(
    accessToken: string,
    offerId: string,
    signal?: AbortSignal,
  ): Promise<EventOccurrence> {
    return this.request(routes.confirmSpotOffer(offerId), { accessToken, signal });
  }

  declineSpotOffer(
    accessToken: string,
    offerId: string,
    signal?: AbortSignal,
  ): Promise<EventOccurrence> {
    return this.request(routes.declineSpotOffer(offerId), { accessToken, signal });
  }

  listRescheduleRequests(
    accessToken: string,
    signal?: AbortSignal,
  ): Promise<RescheduleRequest[]> {
    return this.request(routes.listRescheduleRequests, { accessToken, signal });
  }

  createRescheduleRequest(
    accessToken: string,
    body: CreateRescheduleRequestRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<RescheduleRequest> {
    return this.request(routes.createRescheduleRequest, {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  decideRescheduleRequest(
    accessToken: string,
    requestId: string,
    body: DecideRescheduleRequestRequest,
    signal?: AbortSignal,
  ): Promise<RescheduleRequest> {
    return this.request(routes.decideRescheduleRequest(requestId), {
      accessToken,
      body,
      signal,
    });
  }

  withdrawRescheduleRequest(
    accessToken: string,
    requestId: string,
    signal?: AbortSignal,
  ): Promise<RescheduleRequest> {
    return this.request(routes.withdrawRescheduleRequest(requestId), {
      accessToken,
      signal,
    });
  }

  saveJournalDraft(
    accessToken: string,
    body: JournalDraftRequest,
    signal?: AbortSignal,
  ): Promise<LessonJournal> {
    return this.request(routes.saveJournalDraft, { accessToken, body, signal });
  }

  publishJournal(
    accessToken: string,
    body: PublishJournalRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<LessonJournal> {
    return this.request(routes.publishJournal, {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  getJournal(
    accessToken: string,
    occurrenceId: string,
    studentId: string,
    signal?: AbortSignal,
  ): Promise<LessonJournal> {
    return this.request(routes.getJournal(occurrenceId, studentId), {
      accessToken,
      signal,
    });
  }

  listStudentJournals(
    accessToken: string,
    studentId: string,
    signal?: AbortSignal,
  ): Promise<LessonJournal[]> {
    return this.request(routes.listStudentJournals(studentId), {
      accessToken,
      signal,
    });
  }

  listProgressEvidence(
    accessToken: string,
    studentId: string,
    signal?: AbortSignal,
  ): Promise<ProgressEvidence[]> {
    return this.request(routes.listProgressEvidence(studentId), {
      accessToken,
      signal,
    });
  }

  createMedia(
    accessToken: string,
    body: CreateMediaRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<MediaObject> {
    return this.request(routes.createMedia, {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  uploadMediaChunk(
    accessToken: string,
    mediaId: string,
    offset: number,
    chunk: Uint8Array,
    signal?: AbortSignal,
  ): Promise<MediaObject> {
    return this.request(routes.appendMediaChunk(mediaId), {
      accessToken,
      binaryBody: chunk,
      extraHeaders: { "Upload-Offset": String(offset) },
      signal,
    });
  }

  getMedia(
    accessToken: string,
    mediaId: string,
    signal?: AbortSignal,
  ): Promise<MediaObject> {
    return this.request(routes.getMedia(mediaId), { accessToken, signal });
  }

  signMediaAccess(
    accessToken: string,
    mediaId: string,
    signal?: AbortSignal,
  ): Promise<MediaAccess> {
    return this.request(routes.signMediaAccess(mediaId), { accessToken, signal });
  }

  createHomework(
    accessToken: string,
    body: CreateHomeworkRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<HomeworkAssignment> {
    return this.request(routes.createHomework, {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  getHomework(
    accessToken: string,
    homeworkId: string,
    signal?: AbortSignal,
  ): Promise<HomeworkAssignment> {
    return this.request(routes.getHomework(homeworkId), { accessToken, signal });
  }

  assignHomework(
    accessToken: string,
    homeworkId: string,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<HomeworkAssignment> {
    return this.request(routes.assignHomework(homeworkId), {
      accessToken,
      idempotencyKey,
      signal,
    });
  }

  startHomework(
    accessToken: string,
    homeworkId: string,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<HomeworkAssignment> {
    return this.request(routes.startHomework(homeworkId), {
      accessToken,
      idempotencyKey,
      signal,
    });
  }

  cancelHomework(
    accessToken: string,
    homeworkId: string,
    body: CancelHomeworkRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<HomeworkAssignment> {
    return this.request(routes.cancelHomework(homeworkId), {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  markHomeworkTask(
    accessToken: string,
    homeworkId: string,
    taskId: string,
    body: MarkHomeworkTaskRequest,
    signal?: AbortSignal,
  ): Promise<HomeworkAssignment> {
    return this.request(routes.markHomeworkTask(homeworkId, taskId), {
      accessToken,
      body,
      signal,
    });
  }

  submitHomework(
    accessToken: string,
    homeworkId: string,
    body: SubmitHomeworkRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<HomeworkAssignment> {
    return this.request(routes.submitHomework(homeworkId), {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  reviewHomework(
    accessToken: string,
    homeworkId: string,
    body: ReviewHomeworkRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<HomeworkAssignment> {
    return this.request(routes.reviewHomework(homeworkId), {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  listStudentHomework(
    accessToken: string,
    studentId: string,
    signal?: AbortSignal,
  ): Promise<HomeworkAssignment[]> {
    return this.request(routes.listStudentHomework(studentId), {
      accessToken,
      signal,
    });
  }

  markAttendance(
    accessToken: string,
    lessonId: string,
    studentId: string,
    body: MarkAttendanceRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<AttendanceRecord[]> {
    return this.request(routes.markAttendance(lessonId, studentId), {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  listLessonAttendance(
    accessToken: string,
    lessonId: string,
    signal?: AbortSignal,
  ): Promise<AttendanceRecord[]> {
    return this.request(routes.listLessonAttendance(lessonId), {
      accessToken,
      signal,
    });
  }

  createLesson(
    accessToken: string,
    body: CreateLessonRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<Lesson> {
    return this.request(routes.createLesson, {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  reassignPrimaryTeachers(
    accessToken: string,
    body: ReassignPrimaryTeachersRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<ReassignPrimaryTeachersResult> {
    return this.request(routes.reassignPrimaryTeachers, {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
  }

  replaceLessonTeachers(
    accessToken: string,
    body: ReplaceLessonTeachersRequest,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<ReplaceLessonTeachersResult> {
    return this.request(routes.replaceLessonTeachers, {
      accessToken,
      body,
      idempotencyKey,
      signal,
    });
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
    if (options.binaryBody !== undefined) {
      headers["Content-Type"] = "application/octet-stream";
    }
    if (options.accessToken) {
      headers.Authorization = `Bearer ${options.accessToken}`;
    }
    if (options.idempotencyKey) {
      headers["Idempotency-Key"] = options.idempotencyKey;
    }
    Object.assign(headers, options.extraHeaders ?? {});

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
      if (options.binaryBody !== undefined) {
        // React Native's fetch accepts typed arrays; lib.dom's BodyInit
        // predates Uint8Array<ArrayBufferLike> and needs the assertion.
        init.body = options.binaryBody as unknown as BodyInit;
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
