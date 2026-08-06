import {
  decodeActivationPreview,
  decodeActivationProgress,
  decodeRecoveryCodes,
  decodeRevokeOtherSessionsResult,
  decodeSecurityEventsPage,
  decodeSessionDevices,
  decodeSignInOutcome,
  decodeTwofaEnrollment,
  decodeTwofaStatus,
  decodeVerifiedContact,
  decodeVerifiedContacts,
  decodeDataExport,
  decodeDataExports,
  decodeDeletionRequest,
  decodePolicyVersions,
  decodePrivacySettings,
  decodeProfileView,
  decodeCoreLessonSeries,
  decodeCoreLessonSeriesList,
  decodeRoom,
  decodeRooms,
  decodeSeriesGenerationResult,
  decodeEventCategories,
  decodeEventCategory,
  decodeEventOccurrence,
  decodeEventOccurrences,
  decodeEventSeries,
  decodeRescheduleRequest,
  decodeRescheduleRequests,
  decodeLessonJournal,
  decodeLessonJournals,
  decodeProgressEvidence,
  decodeMediaObject,
  decodeMediaAccess,
  decodeHomeworkAssignment,
  decodeHomeworkAssignments,
  decodeAttendanceRecords,
  decodeStudentSong,
  decodeStudentSongs,
  decodeStudentGoal,
  decodeStudentGoals,
  decodeAchievementDefinition,
  decodeAchievementDefinitions,
  decodeAchievementAward,
  decodeAchievementAwards,
  decodeBootstrapView,
  decodeDelegationResult,
  decodeFirstMinute,
  decodeInvitationResult,
  decodeLesson,
  decodeLessons,
  decodeReassignPrimaryTeachersResult,
  decodeReplaceLessonTeachersResult,
  decodeSessionTokens,
  decodeStaffMembers,
  decodeStudentDirectory,
  decodeStudentOnboardingItems,
  decodeStudentResult,
  decodeVoid,
  type ActivationPreview,
  type ActivationPreviewRequest,
  type BootstrapView,
  type CompleteActivationRequest,
  type CreateLessonRequest,
  type CreateStudentRequest,
  type Decoder,
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
  type ActivationCodeRequest,
  type ActivationContactRequest,
  type ActivationFinishRequest,
  type ActivationPasswordRequest,
  type ActivationProgressView,
  type ActivationTokenRequest,
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
  type StudentSong,
  type AddStudentSongRequest,
  type ChangeSongStageRequest,
  type StudentGoal,
  type CreateGoalRequest,
  type CompleteGoalRequest,
  type ReframeGoalRequest,
  type AchievementDefinition,
  type CreateAchievementDefinitionRequest,
  type AchievementAward,
  type AwardAchievementRequest,
  type RevokeAchievementRequest,
  type DisableTwofaRequest,
  type RequestPasswordResetRequest,
  type SignInOutcome,
  type StartContactChangeRequest,
  type TwofaCodeRequest,
  type TwofaEnrollment,
  type TwofaSignInRequest,
  type TwofaStatus,
  type VerifiedContact,
  type RevokeOtherSessionsResult,
  type RevokeSessionRequest,
  type CompletePasswordResetRequest,
  type SecurityEventsPage,
  type SecurityEventsQuery,
  type SessionDevice,
  type ReplaceLessonTeachersResult,
  type RevokeDelegationRequest,
  type SessionTokens,
  type SignInRequest,
  type StaffMember,
  type StaffRole,
  type StudentDirectoryItem,
  type StudentDirectoryQuery,
  type StudentOnboardingItem,
  type StudentResult,
} from "./contracts";
import { backendIdentifierIssue } from "@/validation/backend";

export type HttpMethod = "GET" | "POST" | "PUT" | "DELETE";
export type AuthMode = "public" | "required";
export type ErrorStatus = 400 | 401 | 403 | 404 | 409 | 422 | 429 | 500;

export interface RouteDescriptor<Response> {
  readonly method: HttpMethod;
  readonly path: string;
  readonly auth: AuthMode;
  readonly successStatus: 200 | 201 | 204;
  readonly errorStatuses: readonly ErrorStatus[];
  readonly decode: Decoder<Response>;
}

function route<Response>(
  method: HttpMethod,
  path: string,
  auth: AuthMode,
  successStatus: 200 | 201 | 204,
  errorStatuses: readonly ErrorStatus[],
  decode: Decoder<Response>,
): RouteDescriptor<Response> {
  return { method, path, auth, successStatus, errorStatuses, decode };
}

function pathPart(value: string, name: string): string {
  const normalized = value.trim();
  if (normalized.length === 0) {
    throw new TypeError(`${name} is required`);
  }
  if (backendIdentifierIssue(normalized) !== null) {
    throw new TypeError(`${name} must be a valid backend identifier`);
  }
  return encodeURIComponent(normalized);
}

function securityEventsPath(query: SecurityEventsQuery): string {
  const parameters: string[] = [];
  if (query.cursor !== undefined) {
    const cursor = query.cursor.trim();
    if (cursor.length === 0 || cursor.length > 64) {
      throw new TypeError("cursor must contain 1 to 64 characters");
    }
    parameters.push(`cursor=${encodeURIComponent(cursor)}`);
  }
  if (query.limit !== undefined) {
    if (!Number.isSafeInteger(query.limit) || query.limit < 1 || query.limit > 50) {
      throw new TypeError("limit must be an integer between 1 and 50");
    }
    parameters.push(`limit=${query.limit}`);
  }
  const suffix = parameters.length === 0 ? "" : `?${parameters.join("&")}`;
  return `/v1/me/security-events${suffix}`;
}

function lessonsPath(query: LessonListQuery): string {
  const parameters = [
    `from=${encodeURIComponent(query.from)}`,
    `to=${encodeURIComponent(query.to)}`,
  ];
  if (query.studentId !== undefined) {
    parameters.push(`studentId=${pathPart(query.studentId, "studentId")}`);
  }
  if (query.teacherAccountId !== undefined) {
    parameters.push(
      `teacherAccountId=${pathPart(query.teacherAccountId, "teacherAccountId")}`,
    );
  }
  return `/v1/lessons?${parameters.join("&")}`;
}

function studentDirectoryPath(query: StudentDirectoryQuery): string {
  return query.asOf === undefined
    ? "/v1/students"
    : `/v1/students?asOf=${encodeURIComponent(query.asOf)}`;
}

export const routes = {
  previewActivation: route<ActivationPreview>(
    "POST",
    "/v1/activations/preview",
    "public",
    200,
    [400, 422, 429, 500],
    decodeActivationPreview,
  ),
  completeActivation: route<void>(
    "POST",
    "/v1/activations/complete",
    "public",
    204,
    [400, 409, 422, 429, 500],
    decodeVoid,
  ),
  signIn: route<SignInOutcome>(
    "POST",
    "/v1/sessions",
    "public",
    200,
    [401, 422, 429, 500],
    decodeSignInOutcome,
  ),
  signInWithTwofa: route<SessionTokens>(
    "POST",
    "/v1/sessions/twofa",
    "public",
    200,
    [401, 422, 429, 500],
    decodeSessionTokens,
  ),
  refreshSession: route<SessionTokens>(
    "POST",
    "/v1/sessions/refresh",
    "public",
    200,
    [401, 422, 429, 500],
    decodeSessionTokens,
  ),
  signOut: route<void>(
    "DELETE",
    "/v1/sessions/current",
    "required",
    204,
    [401, 500],
    decodeVoid,
  ),
  activationProgress: route<ActivationProgressView>(
    "POST",
    "/v1/activations/progress",
    "public",
    200,
    [400, 422, 429, 500],
    decodeActivationProgress,
  ),
  setActivationPassword: route<void>(
    "POST",
    "/v1/activations/password",
    "public",
    204,
    [400, 422, 429, 500],
    decodeVoid,
  ),
  startActivationContact: route<void>(
    "POST",
    "/v1/activations/contact",
    "public",
    204,
    [400, 409, 422, 429, 500],
    decodeVoid,
  ),
  verifyActivationContact: route<void>(
    "POST",
    "/v1/activations/contact/verify",
    "public",
    204,
    [400, 409, 422, 429, 500],
    decodeVoid,
  ),
  startActivationTwofa: route<TwofaEnrollment>(
    "POST",
    "/v1/activations/twofa",
    "public",
    200,
    [400, 409, 422, 429, 500],
    decodeTwofaEnrollment,
  ),
  confirmActivationTwofa: route<string[]>(
    "POST",
    "/v1/activations/twofa/confirm",
    "public",
    200,
    [400, 409, 422, 429, 500],
    decodeRecoveryCodes,
  ),
  finishActivation: route<void>(
    "POST",
    "/v1/activations/finish",
    "public",
    204,
    [400, 409, 422, 429, 500],
    decodeVoid,
  ),
  listMyContacts: route<VerifiedContact[]>(
    "GET",
    "/v1/me/contacts",
    "required",
    200,
    [401, 422, 500],
    decodeVerifiedContacts,
  ),
  startContactChange: route<void>(
    "POST",
    "/v1/me/contacts/change",
    "required",
    204,
    [401, 409, 422, 429, 500],
    decodeVoid,
  ),
  confirmContactChange: route<VerifiedContact>(
    "POST",
    "/v1/me/contacts/confirm",
    "required",
    200,
    [401, 409, 422, 429, 500],
    decodeVerifiedContact,
  ),
  twofaStatus: route<TwofaStatus>(
    "GET",
    "/v1/me/twofa",
    "required",
    200,
    [401, 422, 500],
    decodeTwofaStatus,
  ),
  startTwofaEnrollment: route<TwofaEnrollment>(
    "POST",
    "/v1/me/twofa/enroll",
    "required",
    200,
    [401, 409, 422, 429, 500],
    decodeTwofaEnrollment,
  ),
  confirmTwofaEnrollment: route<string[]>(
    "POST",
    "/v1/me/twofa/confirm",
    "required",
    200,
    [401, 409, 422, 429, 500],
    decodeRecoveryCodes,
  ),
  disableTwofa: route<void>(
    "POST",
    "/v1/me/twofa/disable",
    "required",
    204,
    [401, 409, 422, 429, 500],
    decodeVoid,
  ),
  requestPasswordReset: route<void>(
    "POST",
    "/v1/password-resets",
    "public",
    204,
    [422, 429, 500],
    decodeVoid,
  ),
  completePasswordReset: route<void>(
    "POST",
    "/v1/password-resets/complete",
    "public",
    204,
    [401, 422, 429, 500],
    decodeVoid,
  ),
  listMySessions: route<SessionDevice[]>(
    "GET",
    "/v1/me/sessions",
    "required",
    200,
    [401, 422, 500],
    decodeSessionDevices,
  ),
  revokeOtherSessions: route<RevokeOtherSessionsResult>(
    "POST",
    "/v1/me/sessions/revoke-others",
    "required",
    200,
    [401, 422, 429, 500],
    decodeRevokeOtherSessionsResult,
  ),
  revokeMySession: (sessionId: string) =>
    route<void>(
      "POST",
      `/v1/me/sessions/${pathPart(sessionId, "sessionId")}/revoke`,
      "required",
      204,
      [401, 404, 409, 422, 429, 500],
      decodeVoid,
    ),
  listSecurityEvents: (query: SecurityEventsQuery = {}) =>
    route<SecurityEventsPage>(
      "GET",
      securityEventsPath(query),
      "required",
      200,
      [401, 422, 500],
      decodeSecurityEventsPage,
    ),
  myProfile: route<ProfileView>(
    "GET",
    "/v1/me/profile",
    "required",
    200,
    [401, 422, 500],
    decodeProfileView,
  ),
  updateMyProfile: route<ProfileView>(
    "PUT",
    "/v1/me/profile",
    "required",
    200,
    [401, 422, 500],
    decodeProfileView,
  ),
  listPolicies: route<PolicyVersion[]>(
    "GET",
    "/v1/policies",
    "required",
    200,
    [401, 422, 500],
    decodePolicyVersions,
  ),
  acceptPolicy: route<void>(
    "POST",
    "/v1/me/policy-acceptances",
    "required",
    204,
    [401, 404, 422, 500],
    decodeVoid,
  ),
  privacySettings: route<PrivacySettings>(
    "GET",
    "/v1/me/privacy",
    "required",
    200,
    [401, 422, 500],
    decodePrivacySettings,
  ),
  updatePrivacySettings: route<PrivacySettings>(
    "PUT",
    "/v1/me/privacy",
    "required",
    200,
    [401, 409, 422, 500],
    decodePrivacySettings,
  ),
  listDataExports: route<DataExportRequest[]>(
    "GET",
    "/v1/me/data-exports",
    "required",
    200,
    [401, 422, 500],
    decodeDataExports,
  ),
  createDataExport: route<DataExportRequest>(
    "POST",
    "/v1/me/data-exports",
    "required",
    201,
    [401, 409, 422, 429, 500],
    decodeDataExport,
  ),
  deletionRequest: route<DeletionRequest>(
    "GET",
    "/v1/me/deletion-request",
    "required",
    200,
    [401, 404, 422, 500],
    decodeDeletionRequest,
  ),
  createDeletionRequest: route<DeletionRequest>(
    "POST",
    "/v1/me/deletion-request",
    "required",
    201,
    [401, 409, 422, 429, 500],
    decodeDeletionRequest,
  ),
  cancelDeletionRequest: route<DeletionRequest>(
    "POST",
    "/v1/me/deletion-request/cancel",
    "required",
    200,
    [401, 409, 422, 500],
    decodeDeletionRequest,
  ),
  bootstrap: route<BootstrapView>(
    "GET",
    "/v1/me/bootstrap",
    "required",
    200,
    [401, 500],
    decodeBootstrapView,
  ),
  listStaff: (role: StaffRole) =>
    route<StaffMember[]>(
      "GET",
      `/v1/staff?role=${encodeURIComponent(role)}`,
      "required",
      200,
      [401, 403, 422, 500],
      decodeStaffMembers,
    ),
  listStudents: (query: StudentDirectoryQuery = {}) =>
    route<StudentDirectoryItem[]>(
      "GET",
      studentDirectoryPath(query),
      "required",
      200,
      [401, 403, 422, 500],
      decodeStudentDirectory,
    ),
  listLessons: (query: LessonListQuery) =>
    route<Lesson[]>(
      "GET",
      lessonsPath(query),
      "required",
      200,
      [401, 403, 422, 500],
      decodeLessons,
    ),
  getLesson: (lessonId: string) =>
    route<Lesson>(
      "GET",
      `/v1/lessons/${pathPart(lessonId, "lessonId")}`,
      "required",
      200,
      [401, 403, 404, 500],
      decodeLesson,
    ),
  listRooms: route<Room[]>(
    "GET",
    "/v1/rooms",
    "required",
    200,
    [401, 422, 500],
    decodeRooms,
  ),
  createRoom: route<Room>(
    "POST",
    "/v1/rooms",
    "required",
    201,
    [401, 403, 409, 422, 500],
    decodeRoom,
  ),
  listLessonSeries: route<CoreLessonSeries[]>(
    "GET",
    "/v1/lesson-series",
    "required",
    200,
    [401, 422, 500],
    decodeCoreLessonSeriesList,
  ),
  createLessonSeries: route<CoreLessonSeries>(
    "POST",
    "/v1/lesson-series",
    "required",
    201,
    [401, 403, 409, 422, 500],
    decodeCoreLessonSeries,
  ),
  getLessonSeries: (seriesId: string) =>
    route<CoreLessonSeries>(
      "GET",
      `/v1/lesson-series/${pathPart(seriesId, "seriesId")}`,
      "required",
      200,
      [401, 404, 422, 500],
      decodeCoreLessonSeries,
    ),
  generateSeriesOccurrences: (seriesId: string) =>
    route<SeriesGenerationResult>(
      "POST",
      `/v1/lesson-series/${pathPart(seriesId, "seriesId")}/occurrences`,
      "required",
      201,
      [401, 403, 404, 409, 422, 500],
      decodeSeriesGenerationResult,
    ),
  listEventCategories: route<EventCategory[]>(
    "GET",
    "/v1/event-categories",
    "required",
    200,
    [401, 422, 500],
    decodeEventCategories,
  ),
  createEventCategory: route<EventCategory>(
    "POST",
    "/v1/event-categories",
    "required",
    201,
    [401, 403, 409, 422, 500],
    decodeEventCategory,
  ),
  createEventSeries: route<EventSeries>(
    "POST",
    "/v1/event-series",
    "required",
    201,
    [401, 403, 409, 422, 500],
    decodeEventSeries,
  ),
  generateEventSeriesOccurrences: (seriesId: string) =>
    route<SeriesGenerationResult>(
      "POST",
      `/v1/event-series/${pathPart(seriesId, "seriesId")}/occurrences`,
      "required",
      201,
      [401, 403, 404, 409, 422, 500],
      decodeSeriesGenerationResult,
    ),
  listEvents: (window: EventListWindow) =>
    route<EventOccurrence[]>(
      "GET",
      `/v1/events?from=${encodeURIComponent(window.from)}&to=${encodeURIComponent(window.to)}`,
      "required",
      200,
      [401, 422, 500],
      decodeEventOccurrences,
    ),
  createEvent: route<EventOccurrence>(
    "POST",
    "/v1/events",
    "required",
    201,
    [401, 403, 409, 422, 500],
    decodeEventOccurrence,
  ),
  getEvent: (occurrenceId: string) =>
    route<EventOccurrence>(
      "GET",
      `/v1/events/${pathPart(occurrenceId, "occurrenceId")}`,
      "required",
      200,
      [401, 404, 422, 500],
      decodeEventOccurrence,
    ),
  rsvpToEvent: (occurrenceId: string) =>
    route<EventOccurrence>(
      "POST",
      `/v1/events/${pathPart(occurrenceId, "occurrenceId")}/rsvp`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeEventOccurrence,
    ),
  cancelEventRsvp: (occurrenceId: string) =>
    route<EventOccurrence>(
      "POST",
      `/v1/events/${pathPart(occurrenceId, "occurrenceId")}/rsvp/cancel`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeEventOccurrence,
    ),
  joinEventWaitlist: (occurrenceId: string) =>
    route<EventOccurrence>(
      "POST",
      `/v1/events/${pathPart(occurrenceId, "occurrenceId")}/waitlist`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeEventOccurrence,
    ),
  leaveEventWaitlist: (occurrenceId: string) =>
    route<EventOccurrence>(
      "POST",
      `/v1/events/${pathPart(occurrenceId, "occurrenceId")}/waitlist/leave`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeEventOccurrence,
    ),
  confirmSpotOffer: (offerId: string) =>
    route<EventOccurrence>(
      "POST",
      `/v1/event-offers/${pathPart(offerId, "offerId")}/confirm`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeEventOccurrence,
    ),
  declineSpotOffer: (offerId: string) =>
    route<EventOccurrence>(
      "POST",
      `/v1/event-offers/${pathPart(offerId, "offerId")}/decline`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeEventOccurrence,
    ),
  listRescheduleRequests: route<RescheduleRequest[]>(
    "GET",
    "/v1/reschedule-requests",
    "required",
    200,
    [401, 422, 500],
    decodeRescheduleRequests,
  ),
  createRescheduleRequest: route<RescheduleRequest>(
    "POST",
    "/v1/reschedule-requests",
    "required",
    201,
    [401, 403, 404, 409, 422, 500],
    decodeRescheduleRequest,
  ),
  decideRescheduleRequest: (requestId: string) =>
    route<RescheduleRequest>(
      "POST",
      `/v1/reschedule-requests/${pathPart(requestId, "requestId")}/decide`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeRescheduleRequest,
    ),
  withdrawRescheduleRequest: (requestId: string) =>
    route<RescheduleRequest>(
      "POST",
      `/v1/reschedule-requests/${pathPart(requestId, "requestId")}/withdraw`,
      "required",
      200,
      [401, 409, 422, 500],
      decodeRescheduleRequest,
    ),
  saveJournalDraft: route<LessonJournal>(
    "PUT",
    "/v1/journal-drafts",
    "required",
    200,
    [401, 403, 404, 422, 500],
    decodeLessonJournal,
  ),
  publishJournal: route<LessonJournal>(
    "POST",
    "/v1/journals/publish",
    "required",
    200,
    [401, 403, 404, 409, 422, 500],
    decodeLessonJournal,
  ),
  getJournal: (occurrenceId: string, studentId: string) =>
    route<LessonJournal>(
      "GET",
      `/v1/journals/${pathPart(occurrenceId, "occurrenceId")}/${pathPart(studentId, "studentId")}`,
      "required",
      200,
      [401, 404, 422, 500],
      decodeLessonJournal,
    ),
  listStudentJournals: (studentId: string) =>
    route<LessonJournal[]>(
      "GET",
      `/v1/students/${pathPart(studentId, "studentId")}/journals`,
      "required",
      200,
      [401, 422, 500],
      decodeLessonJournals,
    ),
  listProgressEvidence: (studentId: string) =>
    route<ProgressEvidence[]>(
      "GET",
      `/v1/students/${pathPart(studentId, "studentId")}/progress`,
      "required",
      200,
      [401, 403, 422, 500],
      decodeProgressEvidence,
    ),
  createMedia: route<MediaObject>(
    "POST",
    "/v1/media",
    "required",
    201,
    [401, 403, 422, 500],
    decodeMediaObject,
  ),
  appendMediaChunk: (mediaId: string) =>
    route<MediaObject>(
      "POST",
      `/v1/media/${pathPart(mediaId, "mediaId")}/chunks`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeMediaObject,
    ),
  getMedia: (mediaId: string) =>
    route<MediaObject>(
      "GET",
      `/v1/media/${pathPart(mediaId, "mediaId")}`,
      "required",
      200,
      [401, 404, 422, 500],
      decodeMediaObject,
    ),
  signMediaAccess: (mediaId: string) =>
    route<MediaAccess>(
      "POST",
      `/v1/media/${pathPart(mediaId, "mediaId")}/access`,
      "required",
      200,
      [401, 404, 409, 422, 500],
      decodeMediaAccess,
    ),
  createHomework: route<HomeworkAssignment>(
    "POST",
    "/v1/homework",
    "required",
    201,
    [401, 403, 404, 422, 500],
    decodeHomeworkAssignment,
  ),
  getHomework: (homeworkId: string) =>
    route<HomeworkAssignment>(
      "GET",
      `/v1/homework/${pathPart(homeworkId, "homeworkId")}`,
      "required",
      200,
      [401, 404, 422, 500],
      decodeHomeworkAssignment,
    ),
  assignHomework: (homeworkId: string) =>
    route<HomeworkAssignment>(
      "POST",
      `/v1/homework/${pathPart(homeworkId, "homeworkId")}/assign`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeHomeworkAssignment,
    ),
  startHomework: (homeworkId: string) =>
    route<HomeworkAssignment>(
      "POST",
      `/v1/homework/${pathPart(homeworkId, "homeworkId")}/start`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeHomeworkAssignment,
    ),
  cancelHomework: (homeworkId: string) =>
    route<HomeworkAssignment>(
      "POST",
      `/v1/homework/${pathPart(homeworkId, "homeworkId")}/cancel`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeHomeworkAssignment,
    ),
  markHomeworkTask: (homeworkId: string, taskId: string) =>
    route<HomeworkAssignment>(
      "PUT",
      `/v1/homework/${pathPart(homeworkId, "homeworkId")}/tasks/${pathPart(taskId, "taskId")}`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeHomeworkAssignment,
    ),
  submitHomework: (homeworkId: string) =>
    route<HomeworkAssignment>(
      "POST",
      `/v1/homework/${pathPart(homeworkId, "homeworkId")}/submissions`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeHomeworkAssignment,
    ),
  reviewHomework: (homeworkId: string) =>
    route<HomeworkAssignment>(
      "POST",
      `/v1/homework/${pathPart(homeworkId, "homeworkId")}/review`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeHomeworkAssignment,
    ),
  listStudentHomework: (studentId: string) =>
    route<HomeworkAssignment[]>(
      "GET",
      `/v1/students/${pathPart(studentId, "studentId")}/homework`,
      "required",
      200,
      [401, 403, 422, 500],
      decodeHomeworkAssignments,
    ),
  markAttendance: (lessonId: string, studentId: string) =>
    route<AttendanceRecord[]>(
      "PUT",
      `/v1/lessons/${pathPart(lessonId, "lessonId")}/attendance/${pathPart(studentId, "studentId")}`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeAttendanceRecords,
    ),
  listLessonAttendance: (lessonId: string) =>
    route<AttendanceRecord[]>(
      "GET",
      `/v1/lessons/${pathPart(lessonId, "lessonId")}/attendance`,
      "required",
      200,
      [401, 403, 404, 422, 500],
      decodeAttendanceRecords,
    ),
  addStudentSong: (studentId: string) =>
    route<StudentSong>(
      "POST",
      `/v1/students/${pathPart(studentId, "studentId")}/songs`,
      "required",
      201,
      [401, 403, 404, 422, 500],
      decodeStudentSong,
    ),
  listStudentSongs: (studentId: string) =>
    route<StudentSong[]>(
      "GET",
      `/v1/students/${pathPart(studentId, "studentId")}/songs`,
      "required",
      200,
      [401, 403, 422, 500],
      decodeStudentSongs,
    ),
  changeSongStage: (songId: string) =>
    route<StudentSong>(
      "POST",
      `/v1/songs/${pathPart(songId, "songId")}/stage`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeStudentSong,
    ),
  createGoal: (studentId: string) =>
    route<StudentGoal>(
      "POST",
      `/v1/students/${pathPart(studentId, "studentId")}/goals`,
      "required",
      201,
      [401, 403, 404, 422, 500],
      decodeStudentGoal,
    ),
  listStudentGoals: (studentId: string) =>
    route<StudentGoal[]>(
      "GET",
      `/v1/students/${pathPart(studentId, "studentId")}/goals`,
      "required",
      200,
      [401, 403, 422, 500],
      decodeStudentGoals,
    ),
  completeGoal: (goalId: string) =>
    route<StudentGoal>(
      "POST",
      `/v1/goals/${pathPart(goalId, "goalId")}/complete`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeStudentGoal,
    ),
  reframeGoal: (goalId: string) =>
    route<StudentGoal[]>(
      "POST",
      `/v1/goals/${pathPart(goalId, "goalId")}/reframe`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeStudentGoals,
    ),
  createAchievementDefinition: route<AchievementDefinition>(
    "POST",
    "/v1/achievement-definitions",
    "required",
    201,
    [401, 403, 409, 422, 500],
    decodeAchievementDefinition,
  ),
  listAchievementDefinitions: route<AchievementDefinition[]>(
    "GET",
    "/v1/achievement-definitions",
    "required",
    200,
    [401, 403, 500],
    decodeAchievementDefinitions,
  ),
  retireAchievementDefinition: (definitionId: string) =>
    route<AchievementDefinition>(
      "POST",
      `/v1/achievement-definitions/${pathPart(definitionId, "definitionId")}/retire`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeAchievementDefinition,
    ),
  awardAchievement: (studentId: string) =>
    route<AchievementAward>(
      "POST",
      `/v1/students/${pathPart(studentId, "studentId")}/achievements`,
      "required",
      201,
      [401, 403, 404, 409, 422, 500],
      decodeAchievementAward,
    ),
  listStudentAwards: (studentId: string) =>
    route<AchievementAward[]>(
      "GET",
      `/v1/students/${pathPart(studentId, "studentId")}/achievements`,
      "required",
      200,
      [401, 403, 422, 500],
      decodeAchievementAwards,
    ),
  revokeAchievement: (awardId: string) =>
    route<AchievementAward>(
      "POST",
      `/v1/achievements/${pathPart(awardId, "awardId")}/revoke`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeAchievementAward,
    ),
  createLesson: route<Lesson>(
    "POST",
    "/v1/lessons",
    "required",
    201,
    [401, 403, 409, 422, 500],
    decodeLesson,
  ),
  reassignPrimaryTeachers: route<ReassignPrimaryTeachersResult>(
    "POST",
    "/v1/students/primary-teacher-reassignments",
    "required",
    201,
    [401, 403, 404, 409, 422, 500],
    decodeReassignPrimaryTeachersResult,
  ),
  replaceLessonTeachers: route<ReplaceLessonTeachersResult>(
    "POST",
    "/v1/lessons/teacher-replacements",
    "required",
    200,
    [401, 403, 404, 409, 422, 500],
    decodeReplaceLessonTeachersResult,
  ),
  listStudentOnboarding: route<StudentOnboardingItem[]>(
    "GET",
    "/v1/student-onboarding",
    "required",
    200,
    [401, 403, 422, 500],
    decodeStudentOnboardingItems,
  ),
  grantDelegation: route<DelegationResult>(
    "POST",
    "/v1/access/delegations",
    "required",
    201,
    [401, 403, 409, 422, 429, 500],
    decodeDelegationResult,
  ),
  revokeDelegation: (delegationId: string) =>
    route<void>(
      "POST",
      `/v1/access/delegations/${pathPart(delegationId, "delegationId")}/revoke`,
      "required",
      204,
      [401, 403, 404, 409, 422, 429, 500],
      decodeVoid,
    ),
  createStudent: route<StudentResult>(
    "POST",
    "/v1/students",
    "required",
    201,
    [401, 403, 409, 422, 500],
    decodeStudentResult,
  ),
  publishFirstMinute: (studentId: string) =>
    route<FirstMinute>(
      "PUT",
      `/v1/students/${pathPart(studentId, "studentId")}/first-minute`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeFirstMinute,
    ),
  issueInvitation: (studentId: string) =>
    route<InvitationResult>(
      "POST",
      `/v1/students/${pathPart(studentId, "studentId")}/activation-invitations`,
      "required",
      201,
      [401, 403, 404, 409, 422, 500],
      decodeInvitationResult,
    ),
  reissueInvitation: (studentId: string) =>
    route<InvitationResult>(
      "POST",
      `/v1/students/${pathPart(studentId, "studentId")}/activation-invitations/reissue`,
      "required",
      201,
      [401, 403, 404, 409, 422, 500],
      decodeInvitationResult,
    ),
  revokeInvitation: (invitationId: string) =>
    route<void>(
      "POST",
      `/v1/activation-invitations/${pathPart(invitationId, "invitationId")}/revoke`,
      "required",
      204,
      [401, 403, 404, 409, 422, 500],
      decodeVoid,
    ),
} as const;

export type RouteRequestBodies = {
  previewActivation: ActivationPreviewRequest;
  completeActivation: CompleteActivationRequest;
  signIn: SignInRequest;
  refreshSession: RefreshSessionRequest;
  requestPasswordReset: RequestPasswordResetRequest;
  signInWithTwofa: TwofaSignInRequest;
  activationProgress: ActivationTokenRequest;
  setActivationPassword: ActivationPasswordRequest;
  startActivationContact: ActivationContactRequest;
  verifyActivationContact: ActivationCodeRequest;
  startActivationTwofa: ActivationTokenRequest;
  confirmActivationTwofa: ActivationCodeRequest;
  finishActivation: ActivationFinishRequest;
  startContactChange: StartContactChangeRequest;
  confirmContactChange: ConfirmContactChangeRequest;
  startTwofaEnrollment: CurrentPasswordRequest;
  confirmTwofaEnrollment: TwofaCodeRequest;
  disableTwofa: DisableTwofaRequest;
  updateMyProfile: UpdateProfileRequest;
  acceptPolicy: AcceptPolicyRequest;
  updatePrivacySettings: PrivacySettings;
  createDataExport: CurrentPasswordRequest;
  createDeletionRequest: CurrentPasswordRequest;
  completePasswordReset: CompletePasswordResetRequest;
  revokeOtherSessions: RevokeSessionRequest;
  revokeMySession: RevokeSessionRequest;
  grantDelegation: GrantDelegationRequest;
  revokeDelegation: RevokeDelegationRequest;
  createStudent: CreateStudentRequest;
  publishFirstMinute: PublishFirstMinuteRequest;
  createLesson: CreateLessonRequest;
  createRoom: CreateRoomRequest;
  createEventCategory: CreateEventCategoryRequest;
  createEventSeries: CreateEventSeriesRequest;
  generateEventSeriesOccurrences: GenerateOccurrencesRequest;
  createEvent: CreateEventRequest;
  createRescheduleRequest: CreateRescheduleRequestRequest;
  saveJournalDraft: JournalDraftRequest;
  publishJournal: PublishJournalRequest;
  createMedia: CreateMediaRequest;
  createHomework: CreateHomeworkRequest;
  cancelHomework: CancelHomeworkRequest;
  markHomeworkTask: MarkHomeworkTaskRequest;
  submitHomework: SubmitHomeworkRequest;
  reviewHomework: ReviewHomeworkRequest;
  markAttendance: MarkAttendanceRequest;
  addStudentSong: AddStudentSongRequest;
  changeSongStage: ChangeSongStageRequest;
  createGoal: CreateGoalRequest;
  completeGoal: CompleteGoalRequest;
  reframeGoal: ReframeGoalRequest;
  createAchievementDefinition: CreateAchievementDefinitionRequest;
  awardAchievement: AwardAchievementRequest;
  revokeAchievement: RevokeAchievementRequest;
  decideRescheduleRequest: DecideRescheduleRequestRequest;
  createLessonSeries: CreateLessonSeriesRequest;
  generateSeriesOccurrences: GenerateOccurrencesRequest;
  reassignPrimaryTeachers: ReassignPrimaryTeachersRequest;
  replaceLessonTeachers: ReplaceLessonTeachersRequest;
};
