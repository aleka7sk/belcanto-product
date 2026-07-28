---
Status: Approved
Version: 1.0.0
Last Updated: 2026-07-27

Document Id: ENTITY_CATALOG

Document Type:
  - Domain Contract
  - Entity Catalog
  - Aggregate Ownership Specification
  - Internal Lifecycle Specification
  - Entity Extraction Standard

Owners:
  - Product Owner
  - Education Lead
  - Domain Architecture Lead
  - Technical Lead

Applies To:
  - Aggregate Roots
  - Internal Aggregate Entities
  - Entity Identity
  - Entity Lifecycle
  - Internal Collections
  - Entity Versioning
  - Domain Events
  - Audit
  - Privacy
  - Aggregate Boundary Review

Related Directories:
  - ../architecture/
  - ../aggregates/
  - ../commands/
  - ../events/
  - ../policies/
  - ../value-objects/
  - ../services/
  - ../processes/

Related Documents:
  - ../architecture/000-domain-model-rules.md
  - ../aggregates/000-aggregate-catalog.md
  - ../commands/000-domain-command-catalog.md
  - ../events/000-domain-event-catalog.md
  - ../value-objects/000-value-object-catalog.md
---

# Entity Catalog

> Entity Catalog определяет канонические внутренние Entity доменной модели Belcanto Product.
>
> Entity имеет identity и lifecycle, но не является самостоятельной consistency boundary, пока принадлежит Aggregate.
>
> Внешний код не должен изменять Entity напрямую. Все изменения проходят через owning Aggregate Root.

---

# Purpose

Aggregate Catalog определил крупные consistency boundaries:

- Lesson;
- Homework Assignment;
- Progress Record;
- Goal;
- Song Readiness;
- Concert Participation;
- Reminder Plan;
- Notification Delivery.

Но внутри этих Aggregate существуют объекты, которым недостаточно быть Value Objects.

Например:

- Homework Submission имеет собственный identifier;
- Homework Review развивается во времени;
- Homework Blocker открывается и разрешается;
- Performance Slot назначается, изменяется и снимается;
- Delivery Attempt имеет уникальный номер и outcome;
- Attendance Record может быть исправлен;
- Scheduled Reminder имеет собственный lifecycle.

Эти объекты являются Entity, потому что:

- обладают identity;
- сохраняют непрерывность при изменении значений;
- имеют lifecycle;
- участвуют в инвариантах Aggregate;
- должны сохранять историю.

При этом они не являются Aggregate Roots, пока:

- не требуют независимой транзакции;
- не изменяются отдельно;
- не имеют самостоятельного repository;
- не являются целью внешней Command;
- не должны масштабироваться независимо;
- не владеют cross-aggregate consistency boundary.

---

# Entity Definition

Каноническая структура Entity:

```text
Entity
├── EntityId
├── EntityType
├── Owned State
├── Lifecycle Status
├── CreatedAt
├── UpdatedAt
├── EntityVersion — only when required
└── Historical References
```

Entity определяется identity, а не текущими значениями.

Две Submission с одинаковыми attachments и текстом остаются разными Submission, если имеют разные SubmissionId.

---

# Entity and Aggregate Root

Aggregate Root тоже является Entity, но обладает дополнительными обязанностями:

| Capability | Internal Entity | Aggregate Root |
| --- | --- | --- |
| Имеет identity | Да | Да |
| Имеет lifecycle | Может иметь | Да |
| Владеет Aggregate invariants | Нет | Да |
| Является consistency boundary | Нет | Да |
| Может загружаться отдельно | Нет | Да |
| Имеет repository | Нет | Да |
| Является Command target | Не напрямую | Да |
| Создает Domain Events наружу | Через Root | Да |
| Содержит другие Entity | Обычно нет | Может |

Internal Entity не публикует событие самостоятельно в infrastructure.

Она может вернуть Aggregate Root информацию о произошедшем изменении, но canonical Domain Event создает Root.

---

# Entity and Value Object

| Characteristic | Entity | Value Object |
| --- | --- | --- |
| Identity | Есть | Нет |
| Equality | По identity | По значению |
| Mutable lifecycle | Через доменные методы | Immutable |
| Может иметь history | Да | Обычно нет |
| Может быть replaced by equal value | Нет | Да |
| Repository | Только если становится Aggregate | Нет |

Пример:

```text
HomeworkSubmission
```

является Entity.

```text
SubmissionMethod
```

является Value Object.

---

# Core Entity Rules

## EN-001: Every Entity has an owner Aggregate

Для каждой Entity должен быть указан ровно один authoritative owning Aggregate.

## EN-002: Entity cannot exist independently from its owner unless explicitly promoted to Aggregate Root

Entity может физически храниться в отдельной таблице, но архитектурно остается частью Aggregate.

## EN-003: Entity mutations pass through Aggregate Root

Запрещено:

```text
submissionRepository.Withdraw(submissionId)
```

если HomeworkSubmission принадлежит HomeworkAssignment.

Корректно:

```text
homework.WithdrawSubmission(submissionId, actor, withdrawnAt)
```

## EN-004: Entity identity is stable

EntityId не меняется при:

- status transition;
- correction;
- review;
- reassignment;
- archive;
- restoration.

## EN-005: EntityId is never reused

Удаленный, archived или superseded EntityId не присваивается новой Entity.

## EN-006: Entity equality is identity-based

Две Entity с одинаковыми полями, но разными identifiers, не равны.

## EN-007: Entity lifecycle must be explicit

Entity со сложным поведением не должна описываться набором несвязанных booleans.

## EN-008: Generic setters are prohibited

Нежелательно:

```text
entity.SetStatus(...)
entity.SetData(...)
```

Предпочтительно:

```text
submission.Withdraw(...)
blocker.Resolve(...)
attempt.MarkFailed(...)
```

## EN-009: Entity validates local invariants

Entity может защищать правила собственного состояния, но Aggregate Root отвечает за Aggregate-wide invariants.

## EN-010: Aggregate Root validates relationships between its Entity

Например, Root проверяет, что Homework Review относится к Submission того же Homework Assignment.

## EN-011: Entity does not reference owning Aggregate as mutable object

Entity существует внутри Aggregate object graph и не должна формировать циклическую mutable ownership model.

## EN-012: Foreign aggregates are referenced by identity

## EN-013: Entity does not call repositories or external services

## EN-014: Entity does not read global time

Время передается в доменный метод.

## EN-015: Entity history is not silently overwritten

Исправление attendance или rescheduling reminder должно сохранять traceability.

## EN-016: Entity Version is introduced only when required

EntityVersion оправдана, если:

- Entity изменяется независимо внутри большого Aggregate;
- нужен precise concurrency check;
- Entity участвует в external reference;
- отдельные updates могут конфликтовать;
- history требует versioned snapshot reference.

Не следует добавлять EntityVersion автоматически.

## EN-017: Internal Entity must not expose mutable collections

## EN-018: Entity may become Aggregate Root only through explicit boundary decision

## EN-019: Entity extraction must preserve domain semantics

Нельзя выделить Entity в Aggregate только ради отдельной таблицы или endpoint.

## EN-020: Entity archival does not erase Aggregate history

---

# Entity Identity Scope

Возможные identity scopes:

## Globally Unique Identity

EntityId уникален во всей системе.

Предпочтительно для:

- SubmissionId;
- HomeworkReviewId;
- PerformanceSlotId;
- ReminderId;
- DeliveryAttemptId.

Преимущества:

- удобные references;
- безопасная трассировка;
- простая интеграция;
- меньше composite identifiers.

## Aggregate-Local Identity

EntityId уникален только внутри Aggregate.

Допустимо для:

- ProgressDimensionState, если DimensionId является definition key;
- ReadinessAreaState, если Area является canonical key;
- DeliveryAttempt sequence number.

Внешняя ссылка должна включать owning Aggregate.

---

# Canonical Entity Catalog

| Entity | Owning Aggregate |
| --- | --- |
| AttendanceRecord | Lesson |
| HomeworkSubmission | HomeworkAssignment |
| HomeworkReview | HomeworkAssignment |
| HomeworkBlocker | HomeworkAssignment |
| CorrectionRequest | HomeworkAssignment |
| ProgressDimensionState | ProgressRecord |
| ReadinessAreaState | SongReadiness |
| ConsentState | ConcertParticipation |
| EligibilityState | ConcertParticipation |
| ApprovalState | ConcertParticipation |
| PerformanceSlot | ConcertParticipation |
| ScheduledReminder | HomeworkReminderPlan |
| DeliveryAttempt | NotificationDelivery |

Дополнительные Entity могут появиться после уточнения domain model.

---

# AttendanceRecord Entity

## Owner Aggregate

Lesson

## Purpose

Представляет attendance state одного Student в одном Lesson.

Attendance Record отвечает за:

- принадлежность Student к Lesson;
- текущий attendance outcome;
- фиксацию времени;
- reason;
- correction history;
- actor provenance.

Attendance Record не отвечает за:

- Student Progress;
- payment;
- CRM attendance;
- Goal completion;
- Homework completion;
- disciplinary assessment.

## Identity

Рекомендуемый identity scope:

```text
LessonId + StudentId
```

или глобальный:

```text
AttendanceRecordId
```

Если в одном Lesson для одного Student возможна только одна attendance record, естественный identity:

```text
LessonId + StudentId
```

При этом corrections сохраняются как history внутри Record или события Aggregate.

## State

```text
AttendanceRecord
├── AttendanceRecordId
├── StudentId
├── Status
├── RecordedAt
├── RecordedBy
├── ReasonCategory
├── ExplanationReference
├── CorrectionHistoryReferences
├── LastCorrectedAt
├── LastCorrectedBy
└── EntityVersion — optional
```

## Attendance Statuses

```text
Unknown
Present
Late
Absent
Excused
PartiallyAttended
UnableToDetermine
```

UnableToDetermine отличается от Unknown:

- Unknown — attendance еще не зафиксирован;
- UnableToDetermine — была попытка определить attendance, но достоверного результата нет.

## Lifecycle

```text
Unknown
   |
   +--> Present
   +--> Late
   +--> Absent
   +--> Excused
   +--> PartiallyAttended
   +--> UnableToDetermine

Recorded State
   |
   +--> Corrected to another state
```

Attendance correction не удаляет предыдущий факт.

## Invariants

- Student является участником Lesson.
- Для Student существует не более одной authoritative AttendanceRecord в Lesson.
- RecordedBy является authorized Actor.
- Late MAY require arrival information.
- PartiallyAttended MAY require attended duration or explanation.
- Excused требует reason category.
- Correction требует reason.
- Correction actor и time обязательны.
- Attendance не является Progress Evidence автоматически.
- Notification или geolocation не считаются attendance без approved policy.
- AI не может authoritative определить attendance.
- Completed Lesson может иметь unresolved attendance только при явно разрешенном сценарии.
- Archived Lesson не принимает ordinary attendance correction.

## Root Methods

```text
lesson.RecordAttendance(...)
lesson.CorrectAttendance(...)
```

## Commands Affecting Entity

```text
RecordLessonAttendance
CorrectLessonAttendance
CompleteLesson
```

CompleteLesson может проверить completeness Attendance Records, но не обязан изменять их.

## Events Produced by Lesson Root

```text
LessonAttendanceRecorded
LessonAttendanceCorrected
```

Possible payload:

```text
AttendanceRecordId
StudentId
PreviousStatus
CurrentStatus
RecordedAt
RecordedBy
ReasonCode
```

## Privacy

Default:

```text
Confidential
```

Explanation may be:

```text
Sensitive
```

если содержит health, family или private circumstances.

## Audit Requirements

Сохраняются:

- original status;
- corrected status;
- correction reason;
- actor;
- time;
- command;
- Aggregate Version.

## AI Restrictions

AI может:

- предложить classification по staff note;
- проверить completeness;
- обнаружить contradiction.

AI не может:

- финально поставить Absent;
- определить Excused;
- изменить history;
- использовать camera/audio analysis без отдельного approved policy.

## Tests

- нельзя добавить Student, не участвующего в Lesson;
- duplicate attendance rejected;
- correction preserves previous status;
- Excused requires reason;
- unauthorized actor rejected;
- AI actor rejected;
- correction after archive rejected;
- attendance does not update Progress directly.

## Extraction Criteria

AttendanceRecord следует выделить в отдельный Aggregate, если:

- attendance редактируется независимо от Lesson;
- Lesson содержит сотни или тысячи participants;
- attendance обрабатывается массово и конкурентно;
- attendance имеет собственный review/appeal workflow;
- отдельная legal retention boundary;
- attendance нужен как external integration target.

Для обычных individual и small-group lessons Entity внутри Lesson предпочтительнее.

---

# HomeworkSubmission Entity

## Owner Aggregate

HomeworkAssignment

## Purpose

Представляет конкретную попытку Student отправить результат Homework.

Submission имеет identity, потому что:

- может быть withdrawn;
- может быть reviewed;
- может быть superseded;
- содержит attachments;
- участвует в evidence;
- должна сохраняться исторически.

## Identity

```text
SubmissionId
```

Рекомендуется globally unique identifier.

## State

```text
HomeworkSubmission
├── SubmissionId
├── SubmissionSequence
├── HomeworkVersion
├── SubmittedBy
├── SubmittedAt
├── SubmissionMethod
├── TextReference
├── AttachmentReferences
├── StudentCommentReference
├── Status
├── WithdrawalRecord
├── SupersedesSubmissionId
├── SupersededBySubmissionId
├── CreatedAt
└── EntityVersion — optional
```

## Submission Statuses

```text
Draft
Submitted
Withdrawn
Superseded
Invalidated
Archived
```

Draft включается только если server-side draft является частью domain model.

Client-local draft не является Domain Entity.

## Lifecycle

```text
Submitted
   |
   +--> Withdrawn
   +--> Superseded
   +--> Invalidated
   +--> Archived
```

Новая correction submission обычно создается как новая Entity, а не изменяет старую.

## Invariants

- Submission принадлежит тому же Homework Assignment.
- SubmittedBy соответствует Student или authorized representative flow.
- Homework Version фиксируется.
- Attachment принадлежит Student и доступен tenant.
- Submission не может быть изменена после Submitted, кроме разрешенных metadata corrections.
- Withdrawn Submission нельзя review как active.
- Superseded Submission сохраняется.
- Active Submission count соответствует Homework rules.
- Submission timestamp является trusted server time или validated imported time.
- Empty Submission запрещена, если Submission Method требует content.
- Attachment quarantine blocks submission acceptance.
- AI-generated work не скрывается, если образовательная policy требует disclosure.
- Submission не означает Homework Completion.
- Submission не означает Progress Update.

## Root Methods

```text
homework.Submit(...)
homework.WithdrawSubmission(...)
homework.SupersedeSubmission(...)
homework.InvalidateSubmission(...)
```

## Commands Affecting Entity

```text
SubmitHomework
WithdrawHomeworkSubmission
RequestHomeworkCorrection
ReopenHomework
CancelHomework
ReplaceHomework
```

## Events Produced by Homework Root

```text
HomeworkSubmitted
HomeworkSubmissionWithdrawn
HomeworkSubmissionSuperseded
HomeworkSubmissionInvalidated
```

## Privacy

Default:

```text
Sensitive
```

Attachments и private comments могут иметь более строгий classification.

## Audit Requirements

- SubmissionId;
- Homework Version;
- source Actor;
- submitted time;
- attachment checksums;
- withdrawal;
- supersession chain;
- review references.

## AI Restrictions

AI может:

- проверить формат;
- предложить transcript;
- обнаружить missing attachment;
- подготовить summary для Teacher.

AI не может:

- подтверждать авторство;
- считать Submission выполненным;
- создавать Progress Evidence без review;
- изменять Student submission;
- скрыто заменять original content.

## Tests

- empty required submission rejected;
- wrong Student rejected;
- stale Homework Version behavior defined;
- quarantined attachment rejected;
- withdrawal preserves content;
- withdrawn Submission cannot enter review;
- correction creates new Submission;
- supersession chain remains valid;
- duplicate Command is idempotent.

## Extraction Criteria

Выделить в Aggregate Root, если:

- Submission загружается и review независимо;
- attachment processing длительное;
- много concurrent workflows;
- отдельная moderation;
- отдельный access control;
- Submission может существовать после переноса между Homework;
- Aggregate становится слишком большим;
- Submission становится independent integration target.

До этого Submission остается Entity HomeworkAssignment.

---

# HomeworkReview Entity

## Owner Aggregate

HomeworkAssignment

## Purpose

Представляет authoritative Teacher Review конкретной Homework Submission.

Review отвечает за:

- начало проверки;
- reviewer;
- target Submission;
- outcome;
- feedback;
- evidence references;
- completion;
- correction or clarification result.

## Identity

```text
HomeworkReviewId
```

Globally unique.

## State

```text
HomeworkReview
├── HomeworkReviewId
├── SubmissionId
├── ReviewerTeacherId
├── Status
├── ReviewOutcome
├── FeedbackReference
├── EvidenceReferences
├── StartedAt
├── CompletedAt
├── CancelledAt
├── SupersedesReviewId
├── SupersededByReviewId
└── EntityVersion
```

## Review Statuses

```text
Pending
InProgress
Completed
Cancelled
Superseded
Archived
```

## Review Outcomes

```text
Accepted
AcceptedWithFeedback
CorrectionRequired
ClarificationRequired
InsufficientSubmission
UnableToAssess
RejectedForInvalidSubmission
```

Outcome должен отражать педагогический результат, а не технический error.

## Lifecycle

```text
Pending
   |
   +--> InProgress
           |
           +--> Completed
           +--> Cancelled

Completed
   |
   +--> Superseded by a new review
```

Completed Review не редактируется как mutable note.

Исправление создает correction/superseding review.

## Invariants

- Review относится к существующей Submission.
- Submission принадлежит тому же Homework.
- Withdrawn или invalidated Submission нельзя начать review как active.
- Reviewer имеет действующее Teacher Assignment или delegated authority.
- Один active Review на Submission по умолчанию.
- CompletedAt обязателен для Completed.
- ReviewOutcome обязателен для Completed.
- CorrectionRequired должен создать или поддержать CorrectionRequest.
- Evidence references должны быть подтверждены в допустимом scope.
- Completed Review immutable.
- Superseding Review сохраняет ссылку на прежний.
- AI не является ReviewerTeacherId.
- Feedback visibility определяется явно.
- Review Completion не обязана автоматически завершать Homework, если policy требует отдельной команды.

## Root Methods

```text
homework.StartReview(...)
homework.CompleteReview(...)
homework.CancelReview(...)
homework.SupersedeReview(...)
```

## Commands Affecting Entity

```text
StartHomeworkReview
ReviewHomework
RequestHomeworkCorrection
RequestHomeworkClarification
CompleteHomework
```

## Events Produced by Homework Root

```text
HomeworkReviewStarted
HomeworkReviewed
HomeworkReviewCancelled
HomeworkReviewSuperseded
HomeworkCorrectionRequested
HomeworkClarificationRequested
```

## Privacy

```text
Sensitive
```

Internal Teacher notes may be:

```text
HighlyRestricted
```

Student-visible feedback и internal evaluation должны храниться отдельно.

## Audit Requirements

- reviewer identity;
- authority reference;
- target Submission;
- review start/completion;
- outcome;
- evidence;
- supersession;
- Student-visible feedback version.

## AI Restrictions

AI может:

- подготовить draft feedback;
- структурировать rubric;
- предложить potential evidence;
- проверить consistency.

AI не может:

- завершить Review;
- быть authoritative reviewer;
- определить correction без Teacher/Policy approval;
- публиковать feedback Student автоматически без approved flow.

## Tests

- review of withdrawn submission rejected;
- unauthorized reviewer rejected;
- duplicate active review rejected;
- completed review immutable;
- outcome required;
- correction outcome creates consistent CorrectionRequest;
- superseding preserves history;
- internal note not exposed to Student.

## Extraction Criteria

Выделить в отдельный Aggregate, если:

- Review имеет multi-reviewer workflow;
- moderation или appeal;
- независимый assignment queue;
- review длится долго;
- large rubric;
- concurrent rubric updates;
- отдельная authorization boundary;
- review интегрируется с external assessment platform.

---

# HomeworkBlocker Entity

## Owner Aggregate

HomeworkAssignment

## Purpose

Представляет конкретное препятствие, из-за которого Student не может начать, продолжить или завершить Homework обычным образом.

Примеры:

- отсутствует материал;
- непонятна инструкция;
- техническая проблема;
- health/personal issue;
- задача потеряла актуальность;
- необходима Teacher clarification;
- нет доступа к backing track;
- wrong Song Version.

## Identity

```text
BlockerId
```

## State

```text
HomeworkBlocker
├── BlockerId
├── BlockerCategory
├── Severity
├── ExplanationReference
├── ReportedBy
├── ReportedAt
├── Status
├── AffectsDeadline
├── AffectsReminders
├── RequiresTeacherAction
├── ResolutionRecord
├── SupersededByBlockerId
└── EntityVersion
```

## Blocker Statuses

```text
Open
Acknowledged
ResolutionInProgress
Resolved
Dismissed
Superseded
Archived
```

## Lifecycle

```text
Open
  |
  +--> Acknowledged
  |       |
  |       +--> ResolutionInProgress
  |       +--> Resolved
  |       +--> Dismissed
  |
  +--> Resolved
  +--> Dismissed
  +--> Superseded
```

## Invariants

- Blocker принадлежит текущему Homework.
- Category является canonical.
- Reporter authorized.
- Sensitive explanation не попадает в ordinary Event payload.
- Resolved Blocker содержит resolution.
- Dismissal требует reason.
- Closed Blocker не меняется обратно на Open; создается новый или Reopen action, если такой lifecycle утвержден.
- Active Blocker может влиять на overdue/expiration только через Policy.
- Любой Blocker не приостанавливает deadline автоматически.
- Personal/health category не требует раскрытия подробностей.
- AI не может dismiss Student-reported Blocker.
- Technical Blocker MAY be system-confirmed.
- Duplicate blockers MAY be merged, но history сохраняется.

## Resolution Record

```text
BlockerResolutionRecord
├── ResolutionType
├── ResolvedBy
├── ResolvedAt
├── ResolutionReference
├── FollowUpRequired
└── ReasonCodes
```

## Root Methods

```text
homework.ReportBlocker(...)
homework.AcknowledgeBlocker(...)
homework.StartBlockerResolution(...)
homework.ResolveBlocker(...)
homework.DismissBlocker(...)
homework.SupersedeBlocker(...)
```

## Commands Affecting Entity

```text
ReportHomeworkBlocker
ResolveHomeworkBlocker
RequestHomeworkClarification
ExtendHomeworkDueDate
StartHomeworkGracePeriod
EvaluateHomeworkExpiration
```

## Events Produced by Homework Root

```text
HomeworkBlockerReported
HomeworkBlockerAcknowledged
HomeworkBlockerResolutionStarted
HomeworkBlockerResolved
HomeworkBlockerDismissed
HomeworkBlockerSuperseded
```

## Privacy

Default:

```text
Sensitive
```

Health или personal explanation:

```text
HighlyRestricted
```

Events должны содержать category и reference, но не raw explanation.

## AI Restrictions

AI может:

- классифицировать draft category;
- предложить Teacher response;
- обнаружить duplicate;
- предложить resolution steps.

AI не может:

- dismiss blocker;
- требовать personal disclosure;
- считать Student dishonest;
- автоматически усилить Reminder;
- определить disciplinary consequence.

## Tests

- invalid category rejected;
- resolution requires active Blocker;
- dismissal requires reason;
- sensitive text excluded from Event;
- resolved blocker immutable;
- duplicate handling preserves references;
- active blocker does not directly change deadline;
- unauthorized actor rejected.

## Extraction Criteria

Выделить в Aggregate, если:

- Blocker становится support case;
- имеет assignment, SLA, escalation;
- несколько departments;
- attachments и comments;
- long-running workflow;
- independent notifications;
- external helpdesk integration.

---

# CorrectionRequest Entity

## Owner Aggregate

HomeworkAssignment

## Purpose

Представляет требование исправить конкретную Submission после Review.

Correction Request отличается от нового Homework:

- сохраняет связь с original Submission;
- сохраняет Review outcome;
- определяет correction scope;
- может иметь отдельный deadline;
- ведет к новой Submission того же Homework.

## Identity

```text
CorrectionRequestId
```

## State

```text
CorrectionRequest
├── CorrectionRequestId
├── ReviewId
├── TargetSubmissionId
├── Scope
├── InstructionsReference
├── Status
├── RequestedBy
├── RequestedAt
├── DueDate
├── DeadlineType
├── Requiredness
├── ReplacementSubmissionId
├── CompletionRecord
├── CancellationRecord
└── EntityVersion
```

## Statuses

```text
Open
InProgress
Submitted
Satisfied
Cancelled
Expired
Superseded
Archived
```

## Lifecycle

```text
Open
  |
  +--> InProgress
  +--> Submitted
  +--> Cancelled
  +--> Expired
  +--> Superseded

Submitted
  |
  +--> Satisfied
  +--> Open through new request
  +--> Superseded
```

Один Correction Request не должен бесконечно переиспользоваться для нескольких correction cycles.

## Invariants

- Request относится к Completed Review.
- Review outcome допускает correction.
- Target Submission принадлежит тому же Homework.
- Scope explicit.
- RequestedBy authorized Teacher.
- Due Date semantics explicit.
- Replacement Submission создается после Request.
- Satisfied требует accepted replacement Submission или approved completion method.
- Cancelled Request не принимает Submission.
- Expired Request не завершает Homework автоматически.
- Superseding сохраняет прежний Request.
- Correction не меняет original Submission.
- AI не может authoritative создать correction requirement.

## Root Methods

```text
homework.RequestCorrection(...)
homework.StartCorrection(...)
homework.AttachCorrectionSubmission(...)
homework.SatisfyCorrection(...)
homework.CancelCorrection(...)
homework.ExpireCorrection(...)
homework.SupersedeCorrection(...)
```

## Commands Affecting Entity

```text
RequestHomeworkCorrection
StartHomework
SubmitHomework
ReviewHomework
ChangeHomeworkDueDate
CancelHomework
ReplaceHomework
CompleteHomework
```

## Events Produced by Homework Root

```text
HomeworkCorrectionRequested
HomeworkCorrectionStarted
HomeworkCorrectionSubmitted
HomeworkCorrectionSatisfied
HomeworkCorrectionCancelled
HomeworkCorrectionExpired
HomeworkCorrectionSuperseded
```

## Privacy

```text
Sensitive
```

Student-facing instructions должны быть отделены от internal reviewer notes.

## AI Restrictions

AI может подготовить draft correction instructions.

AI не может:

- назначить correction самостоятельно;
- менять deadline;
- определить удовлетворенность correction;
- формулировать унизительное сообщение;
- расширять scope за пределы Review.

## Tests

- request without completed review rejected;
- invalid review outcome rejected;
- correction submission must supersede target appropriately;
- satisfied requires valid submission;
- cancelled request rejects submission;
- original Submission unchanged;
- deadline history preserved;
- AI actor rejected.

## Extraction Criteria

Обычно остается Entity.

Выделение оправдано при complex multi-step remediation program или independent workflow engine.

---

# ProgressDimensionState Entity

## Owner Aggregate

ProgressRecord

## Purpose

Представляет authoritative current state одной Progress Dimension внутри конкретного Progress scope.

Примеры Dimension:

- breath support;
- intonation;
- rhythm;
- diction;
- vocal range;
- stage confidence;
- musical memory;
- interpretation.

Dimension definition может принадлежать curriculum/configuration context.

## Identity

Рекомендуемый aggregate-local key:

```text
ProgressDimensionId
```

или:

```text
DimensionDefinitionId
```

внутри ProgressRecord.

## State

```text
ProgressDimensionState
├── DimensionId
├── DefinitionVersion
├── CurrentLevel
├── PreviousLevel
├── State
├── Confidence
├── EvidenceReferences
├── LastDecisionReference
├── LastUpdatedAt
├── LastReviewedAt
├── ValidUntil
└── EntityVersion
```

## State Types

Не следует вводить один универсальный numerical score.

Возможные модели:

```text
NotAssessed
Emerging
Developing
Stable
Advanced
ReviewRequired
Stale
```

или curriculum-specific level.

## Lifecycle

```text
NotAssessed
    |
    +--> Assessed State
             |
             +--> Updated State
             +--> ReviewRequired
             +--> Stale
```

Progress может улучшаться, оставаться стабильным или снижаться, если domain policy это допускает.

Снижение не должно скрыто трактоваться как наказание.

## Invariants

- Dimension входит в scope ProgressRecord.
- Definition Version фиксируется.
- Current state поддержан Evidence или authorized Decision.
- Invalidated Evidence не является активной опорой.
- Stale Decision не применяется.
- Одно evaluation не перезаписывает более новое состояние.
- Confidence и confirmation различаются.
- Previous state сохраняется через Event history.
- Dimension State не изменяется Reminder interaction.
- AI не может authoritative изменить state.
- Cross-dimension calculation принадлежит Policy, а не Entity.
- Unknown state не заменяется default level.
- Архивная Dimension сохраняет history.

## Root Methods

```text
progress.RecordEvidence(...)
progress.UpdateDimension(...)
progress.MarkDimensionForReview(...)
progress.MarkDimensionStale(...)
```

## Commands Affecting Entity

```text
RecordProgressEvidence
InvalidateProgressEvidence
EvaluateProgressUpdate
UpdateProgress
RequestProgressReview
```

## Events Produced by Progress Root

```text
ProgressEvidenceRecorded
ProgressDimensionUpdated
ProgressDimensionReviewRequired
ProgressDimensionMarkedStale
```

ProgressUpdated MAY aggregate several dimension changes in one Event, если payload остается bounded.

## Privacy

```text
Sensitive
```

## AI Restrictions

AI может:

- предложить candidate dimension;
- подготовить summary;
- определить missing evidence;
- классифицировать unconfirmed evidence.

AI не может:

- изменить CurrentLevel;
- создать confirmed Evidence;
- сравнивать Student с другими без explicit product policy;
- использовать engagement metrics как ability evidence.

## Tests

- unknown Dimension rejected;
- stale Decision rejected;
- invalidated Evidence rejected;
- newer state not overwritten;
- missing Evidence behavior defined;
- AI source cannot finalize;
- multiple dimensions change atomically only within same ProgressRecord;
- wrong Student Evidence rejected.

## Extraction Criteria

Выделить отдельный Aggregate, если:

- каждая Dimension обновляется независимо с высокой конкуренцией;
- Student имеет сотни Dimensions;
- разные teachers владеют разными scopes;
- Dimension имеет independent review;
- ProgressRecord становится слишком большим;
- external systems адресуют Dimension напрямую.

Более вероятный вариант — один ProgressRecord на ограниченный scope, а не один огромный Aggregate на весь Student.

---

# ReadinessAreaState Entity

## Owner Aggregate

SongReadiness

## Purpose

Представляет assessment одной readiness area для конкретного:

```text
Student
+
Song Version
+
Performance Type
```

Примеры areas:

- Vocal;
- Technical;
- Memory;
- Interpretation;
- Stage Performance;
- Safety;
- Rehearsal;
- Material Preparedness.

## Identity

Aggregate-local canonical key:

```text
ReadinessArea
```

или:

```text
ReadinessAreaStateId
```

если area может иметь несколько параллельных assessments.

## State

```text
ReadinessAreaState
├── Area
├── Outcome
├── Confidence
├── EvidenceReferences
├── Conditions
├── BlockingConditions
├── LastEvaluationReference
├── AssessedAt
├── ValidUntil
├── ReviewRequired
└── EntityVersion
```

## Area Outcomes

```text
NotEvaluated
NotReady
ConditionallyReady
Ready
ReviewRequired
Stale
NotApplicable
```

## Lifecycle

```text
NotEvaluated
    |
    +--> NotReady
    +--> ConditionallyReady
    +--> Ready
    +--> ReviewRequired
    +--> NotApplicable

Evaluated State
    |
    +--> reevaluated state
    +--> Stale
```

## Invariants

- Area входит в применимую readiness model.
- Assessment относится к текущей Song Version.
- Evidence scope совпадает с Area.
- Positive outcome requires sufficient Evidence.
- Conditional outcome содержит Conditions.
- NotReady MAY contain Blocking Conditions.
- Ready может иметь ValidUntil.
- Stale outcome требует reevaluation.
- Не все areas обязаны быть Ready, если Policy допускает weighted decision.
- Overall Song Readiness вычисляется Policy.
- Area State не утверждает Concert Eligibility.
- AI не может authoritative определить outcome.
- Safety area MAY require special human authority.
- Previous evaluation не переписывается.

## Root Methods

```text
songReadiness.ApplyAreaEvaluation(...)
songReadiness.MarkAreaStale(...)
songReadiness.RequireAreaReview(...)
```

## Commands Affecting Entity

```text
AddSongReadinessEvidence
InvalidateSongReadinessEvidence
EvaluateSongReadiness
ChangeSongReadiness
RequestSongReadinessReview
MarkSongReadinessStale
```

## Events Produced by SongReadiness Root

```text
SongReadinessAreaEvaluated
SongReadinessAreaChanged
SongReadinessAreaReviewRequired
SongReadinessAreaMarkedStale
SongReadinessChanged
```

## Privacy

```text
Sensitive
```

Internal performance notes may be restricted.

## AI Restrictions

AI может анализировать approved recordings как advisory input, если consent и privacy позволяют.

AI не может:

- финализировать readiness;
- делать health diagnosis;
- маркировать safety outcome без authorized human;
- сравнивать Student публично;
- менять overall readiness напрямую.

## Tests

- wrong Song Version rejected;
- insufficient evidence cannot produce Ready;
- conditional outcome requires Conditions;
- stale evidence rejected;
- AI decision source rejected;
- area not applicable handled explicitly;
- overall readiness computed outside Entity;
- previous assessment retained.

## Extraction Criteria

Обычно остается Entity.

Отдельный Aggregate возможен, если каждая area имеет independent reviewers, workflow и concurrency.

---

# ConsentState Entity

## Owner Aggregate

ConcertParticipation

## Purpose

Представляет текущее consent state участия Student в Concert.

В MVP может быть Entity/structured state внутри Participation.

Если consent становится reusable legal artifact, его следует выделить в Aggregate.

## Identity

Для простого scope:

```text
ConcertParticipationId + ConsentScope
```

Если сохраняются отдельные consent artifacts:

```text
ConsentId
```

## State

```text
ConsentState
├── ConsentRequestId
├── ConsentId
├── Scope
├── Status
├── RequestedFrom
├── RequestedAt
├── GrantedBy
├── GrantedAt
├── ValidFrom
├── ExpiresAt
├── WithdrawnAt
├── WithdrawalReason
├── ConsentVersion
└── EvidenceReference
```

## Statuses

```text
NotRequired
NotRequested
Requested
Granted
Declined
Withdrawn
Expired
Invalidated
```

## Lifecycle

```text
NotRequested
    |
    +--> Requested
             |
             +--> Granted
             +--> Declined
             +--> Expired

Granted
    |
    +--> Withdrawn
    +--> Expired
    +--> Invalidated
```

## Invariants

- Consent scope explicit.
- Grantor has authority.
- Guardian authority verified where required.
- Consent is versioned.
- Consent applies to specific Participation or declared scope.
- Withdrawal effective behavior explicit.
- Expired Consent cannot authorize action.
- Material scope change requires new Consent.
- Song Version or media/publication scope changes may require reevaluation.
- Consent cannot be bundled invisibly with unrelated purposes.
- AI cannot grant, decline or withdraw Consent.
- Absence of decline is not Consent.
- Notification delivery is not Consent.
- Consent history immutable.

## Root Methods

```text
participation.RequestConsent(...)
participation.GrantConsent(...)
participation.DeclineConsent(...)
participation.WithdrawConsent(...)
participation.ExpireConsent(...)
participation.InvalidateConsent(...)
```

## Commands Affecting Entity

```text
RequestConcertConsent
GrantConcertConsent
WithdrawConcertConsent
ProposeConcertParticipation
ApproveConcertParticipation
WithdrawConcertParticipation
```

## Events Produced by Participation Root

```text
ConcertConsentRequested
ConcertConsentGranted
ConcertConsentDeclined
ConcertConsentWithdrawn
ConcertConsentExpired
ConcertConsentInvalidated
```

## Privacy

```text
Sensitive
```

## AI Restrictions

Полный запрет authoritative consent actions.

AI MAY explain consent text, but explanation must not replace official scope.

## Tests

- unauthorized grantor rejected;
- missing scope rejected;
- expired consent rejected;
- withdrawal preserves grant history;
- scope change requires reevaluation;
- AI actor rejected;
- delivery/open does not grant consent;
- declined consent blocks dependent action according to policy.

## Extraction Criteria

Consent следует выделить в отдельный Aggregate, если:

- применяется к нескольким Participation;
- имеет legal document;
- digital signature;
- multiple scopes;
- guardian relationships;
- independent withdrawal;
- separate retention;
- audit/legal boundary;
- reuse across concerts/media publication.

---

# EligibilityState Entity

## Owner Aggregate

ConcertParticipation

## Purpose

Представляет authoritative current eligibility decision для Participation.

Eligibility отвечает на вопрос:

> Соответствует ли Participation текущим требованиям Concert при данных evidence и conditions?

Она не отвечает:

- утверждено ли участие администрацией;
- включено ли оно в программу;
- назначен ли slot.

## Identity

Обычно один current state внутри Participation.

Каждая evaluation имеет отдельный:

```text
ConcertEligibilityEvaluationId
DecisionId
```

## State

```text
EligibilityState
├── Status
├── DecisionReference
├── RequirementsVersion
├── SongVersionIds
├── InputReferences
├── Conditions
├── BlockingConditions
├── EvaluatedAt
├── ValidUntil
├── StaleReasonCodes
└── EntityVersion
```

## Statuses

```text
NotEvaluated
EvaluationRequested
UnderReview
NotEligible
ConditionallyEligible
Eligible
ReviewRequired
Stale
```

## Lifecycle

```text
NotEvaluated
    |
    +--> EvaluationRequested
             |
             +--> UnderReview
             +--> Eligible
             +--> ConditionallyEligible
             +--> NotEligible
             +--> ReviewRequired

Evaluated State
    |
    +--> Stale
    +--> reevaluated state
```

## Invariants

- Requirements Version explicit.
- Song Versions explicit.
- Decision references exact inputs.
- Eligible requires valid Decision.
- Conditionally Eligible requires Conditions.
- Not Eligible MAY require Blocking Conditions or Reason Codes.
- Stale eligibility cannot support new Approval.
- Material Concert Requirements change invalidates state.
- Song Version change invalidates state.
- Consent change MAY invalidate dependent eligibility.
- Readiness expiration MAY invalidate state.
- Eligibility does not imply Approval.
- Human override preserves original Decision.
- AI cannot be authoritative DecisionSource.
- Current state corresponds to latest accepted Decision.
- Prior Decisions remain auditable.

## Root Methods

```text
participation.RequestEligibilityEvaluation(...)
participation.ApplyEligibilityDecision(...)
participation.MarkEligibilityStale(...)
participation.RequireEligibilityReview(...)
```

## Commands Affecting Entity

```text
RequestConcertEligibilityEvaluation
EvaluateConcertEligibility
MarkConcertParticipationEligible
MarkConcertParticipationConditionallyEligible
MarkConcertParticipationNotEligible
ApproveConcertParticipation
ChangeStudentSongVersion
PublishConcertRequirements
```

Последние две обычно вызывают reaction/process, а не напрямую target Participation.

## Events Produced by Participation Root

```text
ConcertEligibilityEvaluationRequested
ConcertEligibilityEvaluated
ConcertParticipationMarkedEligible
ConcertParticipationMarkedConditionallyEligible
ConcertParticipationMarkedNotEligible
ConcertEligibilityMarkedStale
ConcertEligibilityReviewRequired
```

## Privacy

```text
Sensitive
```

Student-visible explanation должна быть отделена от private evaluator notes.

## AI Restrictions

AI может подготовить input summary или advisory assessment.

AI не может:

- final eligibility decision;
- override blocker;
- create fake evidence;
- approve participation;
- silently alter requirements.

## Tests

- missing Requirements Version rejected;
- stale Song Readiness rejected;
- conditional outcome requires Conditions;
- NotEligible cannot approve;
- stale eligibility blocks Approval;
- human override preserves original;
- AI source rejected;
- requirements change invalidates state.

## Extraction Criteria

Обычно state остается частью Participation, потому что Approval и Slot invariants зависят от Eligibility.

Выделение в Aggregate оправдано, если Evaluation имеет сложный independent workflow и multiple reviewers, но Participation все равно хранит accepted snapshot.

---

# ApprovalState Entity

## Owner Aggregate

ConcertParticipation

## Purpose

Представляет административное или организационное утверждение Participation после Eligibility.

Approval отличается от Eligibility:

- Eligibility — соответствие требованиям;
- Approval — authorized организационное решение;
- Program Placement — включение в published program;
- Slot Assignment — назначение времени/места.

## Identity

Обычно single state внутри Participation.

Каждый approval action MAY иметь:

```text
ApprovalId
```

## State

```text
ApprovalState
├── Status
├── ApprovalId
├── ApprovedBy
├── ApprovedAt
├── ApprovalScope
├── EligibilityDecisionReference
├── Conditions
├── RevokedAt
├── RevokedBy
├── RevocationReasonCodes
└── EntityVersion
```

## Statuses

```text
NotRequested
Pending
Approved
ConditionallyApproved
Rejected
Revoked
Expired
```

## Lifecycle

```text
NotRequested
    |
    +--> Pending
             |
             +--> Approved
             +--> ConditionallyApproved
             +--> Rejected

Approved
    |
    +--> Revoked
    +--> Expired
```

## Invariants

- Approver authorized.
- Approval references current valid Eligibility Decision.
- NotEligible Participation cannot be Approved.
- Conditional Eligibility conditions remain visible.
- Conditional Approval requires Conditions.
- Approval scope explicit.
- Revocation preserves original Approval.
- Expired or stale Eligibility may invalidate Approval according to Policy.
- Approval does not assign Slot.
- Approval does not publish Program.
- AI cannot approve or revoke.
- Teacher recommendation is not administrative Approval unless role explicitly grants authority.
- Duplicate same decision is idempotent.

## Root Methods

```text
participation.RequestApproval(...)
participation.Approve(...)
participation.ConditionallyApprove(...)
participation.RejectApproval(...)
participation.RevokeApproval(...)
participation.ExpireApproval(...)
```

## Commands Affecting Entity

```text
ApproveConcertParticipation
RejectConcertParticipationApproval
RevokeConcertParticipationApproval
WithdrawConcertParticipation
AssignConcertPerformanceSlot
```

## Events Produced by Participation Root

```text
ConcertParticipationApprovalRequested
ConcertParticipationApproved
ConcertParticipationConditionallyApproved
ConcertParticipationApprovalRejected
ConcertParticipationApprovalRevoked
ConcertParticipationApprovalExpired
```

## Privacy

```text
Confidential
```

Reason details may be Sensitive.

## AI Restrictions

AI may prepare operational recommendation.

AI cannot approve, reject or revoke.

## Tests

- invalid eligibility blocks Approval;
- unauthorized approver rejected;
- conditional approval requires Conditions;
- revocation preserves original;
- slot requires active approval;
- AI actor rejected;
- stale eligibility behavior enforced;
- duplicate command idempotent.

## Extraction Criteria

ApprovalState обычно остается внутри Participation, поскольку Slot invariant должен быть strongly consistent с Approval.

---

# PerformanceSlot Entity

## Owner Aggregate

ConcertParticipation

## Purpose

Представляет назначенное место и время выступления Participation в Concert program.

В текущей модели Slot принадлежит Participation, но глобальная уникальность concert schedule может потребовать отдельного ConcertProgram или reservation Aggregate.

## Identity

```text
PerformanceSlotId
```

## State

```text
PerformanceSlot
├── PerformanceSlotId
├── ScheduledStart
├── ExpectedDuration
├── StageReference
├── Sequence
├── Status
├── AssignedBy
├── AssignedAt
├── ChangedAt
├── PreviousSlotReference
├── RemovalRecord
└── EntityVersion
```

## Statuses

```text
Reserved
Assigned
Confirmed
Changed
Removed
Completed
Cancelled
```

Применяемые статусы зависят от Concert Program model.

## Lifecycle

```text
Reserved
   |
   +--> Assigned
           |
           +--> Confirmed
           +--> Changed
           +--> Removed
           +--> Cancelled

Confirmed
   |
   +--> Changed
   +--> Removed
   +--> Completed
```

Изменение Slot может создавать новую slot revision или новую Entity — решение должно быть единообразным.

Рекомендуется сохранять тот же PerformanceSlotId и history changes, пока это то же program placement.

## Invariants

- Participation Approved.
- Participation not withdrawn.
- Concert active.
- ScheduledStart находится в Concert TimeWindow.
- ExpectedDuration положительна.
- StageReference принадлежит Concert.
- Slot change preserves previous values.
- Removal requires reason.
- Completed Slot immutable.
- Один Participation имеет не более одного active Slot, если Policy не допускает multiple performances.
- Global overlap нельзя надежно проверить только внутри Participation.
- Capacity/overlap требует reservation или ConcertProgram coordination.
- AI не может назначать Slot самостоятельно.
- Timezone semantics explicit.

## Root Methods

```text
participation.AssignSlot(...)
participation.ChangeSlot(...)
participation.ConfirmSlot(...)
participation.RemoveSlot(...)
participation.CompletePerformance(...)
```

## Commands Affecting Entity

```text
AssignConcertPerformanceSlot
ChangeConcertPerformanceSlot
RemoveConcertPerformanceSlot
CompleteConcertPerformance
WithdrawConcertParticipation
CancelConcert
```

## Events Produced by Participation Root

```text
ConcertPerformanceSlotAssigned
ConcertPerformanceSlotChanged
ConcertPerformanceSlotConfirmed
ConcertPerformanceSlotRemoved
ConcertPerformanceCompleted
```

## Privacy

```text
Internal
```

До публикации программы — Confidential/Internal.

После публикации selected fields могут быть Public.

## AI Restrictions

AI может предложить оптимизированный schedule.

AI не может authoritative назначить Slot без approved command.

## Tests

- slot before Approval rejected;
- slot outside Concert window rejected;
- invalid Stage rejected;
- active slot uniqueness enforced locally;
- change preserves history;
- withdrawn Participation rejects assignment;
- completed slot immutable;
- AI actor rejected.

## Extraction Criteria

Performance Slot следует вынести в ConcertProgram Aggregate или reservation model, если необходимо strongly enforce:

- global non-overlap;
- shared stage capacity;
- sequence uniqueness;
- multiple stages;
- program publication atomicity;
- mass schedule edits;
- drag-and-drop concurrent planning.

Вероятно, после MVP потребуется:

```text
ConcertProgram Aggregate
├── ProgramVersion
├── ProgramEntries
└── SlotReservations
```

Тогда Participation будет хранить ProgramEntryReference, а не владеть Slot.

---

# ScheduledReminder Entity

## Owner Aggregate

HomeworkReminderPlan

## Purpose

Представляет один конкретный Reminder occurrence, запланированный в рамках Reminder Plan.

## Identity

```text
ReminderId
```

Globally unique.

## State

```text
ScheduledReminder
├── ReminderId
├── ReminderType
├── Sequence
├── ScheduledFor
├── Status
├── DueAt
├── NotificationIntentId
├── DeliveryReference
├── SuppressionRecord
├── RescheduleHistoryReferences
├── AttemptReference
├── ExpirationAt
├── CreatedAt
└── EntityVersion
```

## Reminder Statuses

```text
Scheduled
Due
EvaluationRequired
DeliveryRequested
Delivered
DeliveryFailed
Rescheduled
Suppressed
Cancelled
Expired
```

## Lifecycle

```text
Scheduled
   |
   +--> Due
   +--> Rescheduled
   +--> Suppressed
   +--> Cancelled
   +--> Expired

Due
   |
   +--> EvaluationRequired
             |
             +--> DeliveryRequested
             +--> Suppressed
             +--> Rescheduled
             +--> Cancelled
             +--> Expired

DeliveryRequested
   |
   +--> Delivered
   +--> DeliveryFailed
```

Reminder Plan не должен считать provider delivery своей authoritative responsibility, но может хранить delivery outcome reference.

## Invariants

- Reminder принадлежит Plan.
- ScheduledFor соответствует Plan Timezone и strategy.
- Duplicate type/sequence/window rejected.
- Sequence monotonic.
- Maximum reminder count enforced by Root.
- Delivery request требует revalidation.
- Submitted, Completed, Cancelled, Replaced или Expired Homework может подавить Reminder.
- Reminder не меняет Homework.
- Delivered не означает Read.
- Failed delivery не означает Student fault.
- Reschedule preserves previous time.
- Suppression requires Reason Code.
- Expired Reminder cannot be delivered.
- Old Homework Version Reminder cannot be reactivated.
- AI cannot увеличить количество или urgency.

## Root Methods

```text
plan.ScheduleReminder(...)
plan.MarkReminderDue(...)
plan.RequestReminderDelivery(...)
plan.RescheduleReminder(...)
plan.SuppressReminder(...)
plan.CancelReminder(...)
plan.ExpireReminder(...)
plan.RecordDeliveryOutcome(...)
```

## Commands Affecting Entity

```text
ScheduleHomeworkReminder
RescheduleHomeworkReminder
EvaluateHomeworkReminder
RequestHomeworkReminderDelivery
SuppressHomeworkReminder
CancelHomeworkReminder
ExpireHomeworkReminder
RecalculateHomeworkReminderPlan
```

## Events Produced by Reminder Plan Root

```text
HomeworkReminderScheduled
HomeworkReminderDue
HomeworkReminderDeliveryRequested
HomeworkReminderRescheduled
HomeworkReminderSuppressed
HomeworkReminderCancelled
HomeworkReminderExpired
HomeworkReminderDelivered
HomeworkReminderDeliveryFailed
```

## Privacy

```text
Confidential
```

Reminder content принадлежит Notification Intent, не Reminder Entity.

## AI Restrictions

AI может предложить schedule в пределах approved strategy.

AI не может:

- увеличить maximum count;
- обходить Quiet Hours;
- применять Critical priority;
- писать манипулятивный текст;
- отменять suppression;
- считать lack of response основанием для давления.

## Tests

- duplicate Reminder rejected;
- maximum count enforced;
- stale Homework Version suppressed;
- due Reminder revalidates state;
- reschedule preserves history;
- expired reminder cannot deliver;
- delivered does not change Homework;
- quiet hours observed;
- AI escalation rejected.

## Extraction Criteria

Выделить Reminder в отдельный Aggregate, если:

- Plan содержит большое количество reminders;
- каждое reminder независимо конкурентно обрабатывается;
- provider callbacks адресуют Reminder;
- отдельные retries и locks;
- Plan становится contention hotspot;
- reminders создаются динамически без bounded collection.

Для обычного Homework количество reminders должно быть малым и bounded.

---

# DeliveryAttempt Entity

## Owner Aggregate

NotificationDelivery

## Purpose

Представляет одну техническую попытку отправить Notification через provider.

Attempt является Entity, потому что:

- имеет sequence;
- имеет provider reference;
- имеет отдельный outcome;
- участвует в retry lifecycle;
- должен сохраняться для audit и diagnostics.

## Identity

Возможные варианты:

```text
DeliveryAttemptId
```

или aggregate-local:

```text
AttemptNumber
```

Рекомендуется:

```text
DeliveryAttemptId
+
AttemptNumber
```

## State

```text
DeliveryAttempt
├── DeliveryAttemptId
├── AttemptNumber
├── Status
├── RequestedAt
├── StartedAt
├── CompletedAt
├── ProviderReference
├── ProviderRequestId
├── IdempotencyReference
├── Outcome
├── FailureCategory
├── FailureCode
├── Retryability
├── ProviderResponseReference
└── EntityVersion — optional
```

## Attempt Statuses

```text
Created
Queued
Started
Succeeded
Failed
TimedOut
Cancelled
UnknownOutcome
```

## Lifecycle

```text
Created
   |
   +--> Queued
           |
           +--> Started
                   |
                   +--> Succeeded
                   +--> Failed
                   +--> TimedOut
                   +--> UnknownOutcome

Created / Queued
   |
   +--> Cancelled
```

Terminal attempt не изменяется.

Provider callback MAY enrich reference, но не должен менять Succeeded обратно в Failed.

## Invariants

- Attempt принадлежит одной Notification Delivery.
- AttemptNumber monotonic.
- Только один active send attempt, если provider semantics не допускает concurrent attempts.
- Attempt uses stable idempotency reference.
- StartedAt >= RequestedAt.
- CompletedAt >= StartedAt.
- Terminal status immutable.
- Retryability derived from failure classification.
- Permanent failure не retryable.
- Unknown outcome требует provider reconciliation перед duplicate send where possible.
- Provider secret/token не хранится.
- Provider payload reference follows privacy rules.
- Duplicate callback idempotent.
- Attempt success не означает Notification Open.
- Attempt не изменяет source domain.
- AI не участвует в send authority.

## Root Methods

```text
delivery.CreateAttempt(...)
delivery.MarkAttemptStarted(...)
delivery.MarkAttemptSucceeded(...)
delivery.MarkAttemptFailed(...)
delivery.MarkAttemptTimedOut(...)
delivery.MarkAttemptUnknown(...)
delivery.CancelAttempt(...)
```

## Commands Affecting Entity

```text
SendNotification
RetryNotificationDelivery
MarkNotificationDelivered
MarkNotificationDeliveryFailed
StopNotificationRetry
CancelNotificationDelivery
```

Provider callbacks обычно маппятся в trusted integration commands.

## Events Produced by Delivery Root

Внешний Domain Event не обязан публиковаться на каждое внутреннее изменение.

Возможные события:

```text
NotificationSendingRequested
NotificationDeliveryAttemptStarted
NotificationDelivered
NotificationDeliveryFailed
NotificationRetryScheduled
NotificationRetryStopped
```

AttemptStarted может быть operational event, а не Domain Event.

## Privacy

```text
Confidential
```

Provider payload и destination могут быть Sensitive.

## AI Restrictions

AI не нужен для lifecycle Attempt.

AI не может:

- определять success;
- инициировать retry вне policy;
- менять provider response;
- видеть unmasked destination без authorization.

## Tests

- attempt number monotonic;
- duplicate callback idempotent;
- terminal status immutable;
- permanent failure stops retry;
- timeout classification correct;
- unknown outcome does not blindly duplicate send;
- provider secret never persisted;
- success does not mark opened;
- expired Delivery rejects new Attempt.

## Extraction Criteria

DeliveryAttempt может быть вынесен из active Aggregate storage в отдельный append-only operational store, если attempts многочисленны.

При этом NotificationDelivery сохраняет authoritative summary:

```text
CurrentAttemptNumber
CurrentStatus
LastFailure
DeliveredAt
```

Отдельным Aggregate Root Attempt обычно становиться не должен, так как не владеет самостоятельным бизнес-инвариантом.

---

# Entity Lifecycle Modeling

Для Entity со сложными переходами требуется state transition table.

Пример:

| Entity | Current State | Action | New State |
| --- | --- | --- | --- |
| HomeworkSubmission | Submitted | Withdraw | Withdrawn |
| HomeworkReview | Pending | Start | InProgress |
| HomeworkReview | InProgress | Complete | Completed |
| HomeworkBlocker | Open | Resolve | Resolved |
| CorrectionRequest | Open | Attach Submission | Submitted |
| ConsentState | Requested | Grant | Granted |
| EligibilityState | NotEvaluated | Apply Decision | Eligible / Conditional / NotEligible |
| ApprovalState | Pending | Approve | Approved |
| PerformanceSlot | Assigned | Remove | Removed |
| ScheduledReminder | Scheduled | Mark Due | Due |
| DeliveryAttempt | Started | Provider Success | Succeeded |

Неописанный переход запрещен.

---

# Entity Collections

Aggregate может содержать Entity collection только если она:

- bounded;
- требуется для invariant;
- разумно загружается;
- не создает постоянный contention;
- не является историческим журналом без ограничений.

## Bounded Collection Rules

### EC-001

Maximum size или natural bound должны быть известны.

Примеры:

- Attendance Records ограничены participants Lesson;
- active Homework Blockers ограничены нормальным use case;
- Scheduled Reminders ограничены strategy;
- Delivery Attempts ограничены retry policy.

### EC-002

Historical records могут быть вынесены в archive/audit store.

### EC-003

Root должен хранить минимальный authoritative summary после externalization.

### EC-004

Pagination не решает consistency problem внутри Aggregate автоматически.

Если Root должен загрузить все Entity для invariant, paginated collection может означать неверную boundary.

---

# Entity Versioning

EntityVersion вводится, когда Entity reference должен фиксировать точное состояние.

Пример:

```text
EntityReference
├── HomeworkAssignmentId
├── SubmissionId
└── SubmissionVersion
```

Но если все изменения защищены AggregateVersion и Entity не изменяется независимо, отдельная версия может быть лишней.

## Versioning Decision Table

| Situation | EntityVersion |
| --- | --- |
| Entity immutable after creation | Обычно не нужна |
| Изменяется только вместе с небольшим Aggregate | Необязательно |
| На Entity ссылаются Decisions | Рекомендуется |
| Entity обновляется конкурентно | Нужна |
| Entity вынесена в отдельное storage document | Рекомендуется |
| External integration references revisions | Нужна |
| Identity + immutable supersession model | Может не понадобиться |

---

# Entity Event Production

Internal Entity не должна напрямую публиковать Event в broker.

Канонический flow:

```text
Aggregate Root method
       |
       v
Entity method
       |
       v
Entity transition result
       |
       v
Aggregate Root updates aggregate state
       |
       v
Aggregate Root records Domain Event
```

## Internal Transition Result

Допустимый internal contract:

```text
EntityTransitionResult
├── Changed
├── PreviousState
├── CurrentState
├── ReasonCodes
├── EntityReference
└── DomainDetails
```

Он не является публичным Domain Event.

## Entity References in Events

Event SHOULD содержать EntityId, если факт относится к конкретной Entity.

Пример:

```text
HomeworkSubmissionWithdrawn
├── HomeworkAssignmentId
├── SubmissionId
├── HomeworkVersion
├── WithdrawnAt
└── ReasonCode
```

Не следует публиковать весь Entity snapshot.

---

# Entity Repository Rules

## ER-001

Public repository создается для Aggregate Root, а не внутренней Entity.

## ER-002

Infrastructure MAY использовать отдельные DAO для таблиц, но Domain/Application layer не должны трактовать их как Entity repositories.

## ER-003

DAO не должен позволять business mutation в обход Root.

## ER-004

Entity rows сохраняются в транзакции owning Aggregate.

## ER-005

Удаление child row не должно разрушать domain history.

---

# Entity Serialization Rules

Entity serialization является internal persistence concern, если Entity не входит в external contract.

При публикации наружу используется DTO или Event payload.

Не следует отдавать internal Entity object напрямую mobile client.

---

# Entity Privacy Rules

- Entity inherits минимум privacy owning Aggregate.
- Field MAY иметь более строгую classification.
- Internal notes отделяются от Student-visible content.
- Event payload минимизируется.
- EntityId не должен предоставлять unauthorized lookup.
- Attachment и destination references маскируются.
- Health/personal blocker details не публикуются.
- Provider response не попадает в ordinary domain logs.

---

# Entity Audit Rules

Для meaningful Entity transition сохраняются:

```text
Owning Aggregate Type
Owning Aggregate Id
Aggregate Version
Entity Type
Entity Id
Previous State
Current State
Command Id
Actor
OccurredAt
Reason Codes
Decision Reference
Correlation Id
Causation Id
```

Audit не обязан хранить полный before/after snapshot, если это нарушает privacy.

---

# Entity AI Rules

AI может:

- предлагать draft fields;
- классифицировать unconfirmed data;
- помогать Teacher;
- выявлять inconsistency;
- готовить human-readable summary;
- предложить next action.

AI не может:

- создавать authoritative Entity transition;
- подменять Actor;
- изменять identity;
- переписывать history;
- выдавать Consent;
- завершать Review;
- подтверждать Attendance;
- назначать Approval;
- считать Delivery успешной;
- ослаблять Privacy Level.

Любое AI-assisted изменение выполняется через обычную Command и authorized Actor/Policy.

---

# Entity Extraction Rules

Internal Entity должна быть рассмотрена как candidate Aggregate Root, если выполняется несколько условий:

## Independent Lifecycle

Entity имеет lifecycle, который развивается независимо от owner.

## Independent Concurrency

Разные actors регулярно изменяют Entity параллельно, не затрагивая остальной Aggregate.

## Independent Transaction

Для Entity требуется собственная atomic boundary.

## Independent Authorization

Доступ к Entity значительно отличается от access к owner Aggregate.

## Independent Scale

Entity collection становится большой или unbounded.

## Independent Availability

Entity должна быть доступна и изменяема, даже когда owner Aggregate недоступен или archived.

## External Target

External systems и commands адресуют Entity напрямую.

## Independent Invariants

Entity владеет правилами, не требующими полной загрузки owner.

## Operational Contention

Изменения Entity создают hot Aggregate.

## Independent Retention

Entity имеет отдельный retention/legal boundary.

## Extraction Is Not Justified By

Нельзя выделять Aggregate только потому что:

- Entity хранится в отдельной таблице;
- нужен отдельный API endpoint;
- UI имеет отдельный экран;
- объект имеет много полей;
- ORM удобнее;
- разработчик хочет отдельный service;
- объект используется в join;
- нужен отдельный DTO;
- нужна pagination;
- код кажется большим.

## Extraction Procedure

При выделении Entity в Aggregate Root требуется:

1. Определить новую consistency boundary.
2. Назначить AggregateId.
3. Определить repository.
4. Перенести owned invariants.
5. Заменить object ownership на references.
6. Определить commands.
7. Определить events.
8. Ввести eventual consistency.
9. Определить coordination process.
10. Пересмотреть authorization.
11. Пересмотреть privacy.
12. Пересмотреть transaction guarantees.
13. Подготовить migration.
14. Обновить Aggregate Catalog.
15. Создать Architecture Decision.

---

# Candidate Future Aggregate Roots

Следующие Entity могут стать Aggregate Roots позже:

## HomeworkSubmission

При complex independent review workflow.

## HomeworkReview

При multi-reviewer assessment и appeal.

## Consent

При reusable legal consent model.

## ConcertProgram

Для global slot consistency.

## ScheduledReminder

При high-volume independent reminder processing.

Не следует выделять заранее без реальной необходимости.

---

# Entity Command Routing

Внешняя Command всегда target owning Aggregate.

Пример:

```text
WithdrawHomeworkSubmission
├── Target:
│   ├── AggregateType: HomeworkAssignment
│   └── AggregateId: HomeworkAssignmentId
└── Payload:
    └── SubmissionId
```

Не:

```text
TargetAggregateType: HomeworkSubmission
```

пока Submission не является Root.

---

# Entity Not Found Behavior

Если target Entity отсутствует внутри существующего Aggregate:

```text
CommandResult.Status: Rejected
ReasonCode: AGGREGATE_ENTITY_NOT_FOUND
```

Для idempotent terminal commands MAY возвращаться:

```text
Already Processed
```

только если система может доказать prior processing.

---

# Entity Conflict Behavior

Entity-level conflict обычно обнаруживается через:

```text
ExpectedAggregateVersion
```

При использовании EntityVersion может дополнительно применяться:

```text
ExpectedEntityVersion
```

Но нельзя позволять EntityVersion обойти stale AggregateVersion, если операция влияет на Aggregate invariants.

---

# Entity Reason Codes

Общие reason codes:

```text
ENTITY_NOT_FOUND
ENTITY_ALREADY_EXISTS
ENTITY_DUPLICATE
ENTITY_STATUS_INVALID
ENTITY_TRANSITION_NOT_ALLOWED
ENTITY_TERMINAL_STATE
ENTITY_VERSION_CONFLICT
ENTITY_OWNER_MISMATCH
ENTITY_TENANT_MISMATCH
ENTITY_REFERENCE_INVALID
ENTITY_REFERENCE_STALE
ENTITY_REQUIRED_RELATION_MISSING
ENTITY_ALREADY_SUPERSEDED
ENTITY_ALREADY_WITHDRAWN
ENTITY_ALREADY_RESOLVED
ENTITY_ALREADY_ARCHIVED
ENTITY_REOPEN_NOT_ALLOWED
ENTITY_ACTOR_NOT_AUTHORIZED
ENTITY_DECISION_REQUIRED
ENTITY_EVIDENCE_REQUIRED
ENTITY_HISTORY_CONFLICT
```

Domain-specific codes предпочтительнее generic code, когда известна точная причина.

---

# Entity Test Standard

Каждая Entity specification должна иметь следующие категории тестов.

## Construction Tests

- valid Entity creation;
- required fields;
- identity;
- default state;
- invalid references;
- tenant mismatch.

## Lifecycle Tests

- every allowed transition;
- every forbidden transition;
- terminal behavior;
- supersession;
- cancellation;
- archive.

## Identity Tests

- identity stability;
- no reuse;
- equality by identity;
- local/global scope.

## Ownership Tests

- wrong owner rejected;
- Entity cannot mutate independently;
- repository boundary preserved.

## Concurrency Tests

- stale Aggregate Version;
- stale Entity Version, если используется;
- competing transitions;
- idempotent retry.

## History Tests

- correction preserves original;
- supersession chain;
- terminal record retained;
- audit references valid.

## Privacy Tests

- restricted fields not in Event;
- Student-visible/internal separation;
- safe logging;
- wrong actor access rejected.

## AI Tests

- AI cannot perform authoritative transition;
- AI proposal remains advisory;
- provenance retained;
- no fabricated evidence.

---

# Cross-Entity Invariants

Cross-Entity rules внутри одного Aggregate принадлежат Root.

Примеры HomeworkAssignment:

- Review references existing Submission.
- CorrectionRequest references Completed Review.
- Replacement Submission corresponds to active CorrectionRequest.
- One active Review per Submission.
- One active CorrectionRequest per review cycle.
- Completed Homework has no unresolved required CorrectionRequest.

Примеры ConcertParticipation:

- Approval references current Eligibility.
- Slot requires Approval.
- Consent scope covers Participation.
- Withdrawal affects active Slot.
- Requirements change invalidates Eligibility.

Internal Entity не должна самостоятельно искать siblings через repository.

Root передает необходимые ссылки или выполняет проверку.

---

# Aggregate-Specific Entity Map

## Lesson

```text
Lesson
└── AttendanceRecord[]
```

Potential future:

```text
Lesson
├── AttendanceRecord[]
├── LessonParticipant[]
└── LessonCorrection[]
```

## HomeworkAssignment

```text
HomeworkAssignment
├── HomeworkSubmission[]
├── HomeworkReview[]
├── HomeworkBlocker[]
└── CorrectionRequest[]
```

Collections должны быть bounded или historical records externalized.

## ProgressRecord

```text
ProgressRecord
└── ProgressDimensionState[]
```

## SongReadiness

```text
SongReadiness
└── ReadinessAreaState[]
```

## ConcertParticipation

```text
ConcertParticipation
├── ConsentState
├── EligibilityState
├── ApprovalState
└── PerformanceSlot
```

Часть из них может быть modeled as structured Entity или specialized state object. Поскольку у них есть history и identity-bearing decisions, простых enum недостаточно.

## HomeworkReminderPlan

```text
HomeworkReminderPlan
└── ScheduledReminder[]
```

## NotificationDelivery

```text
NotificationDelivery
└── DeliveryAttempt[]
```

---

# Entity Modeling Decisions

## AttendanceRecord

Решение:

```text
Entity inside Lesson
```

Причина:

attendance strongly tied to Lesson participant invariant.

## HomeworkSubmission

Решение:

```text
Entity inside HomeworkAssignment
```

с возможным future extraction.

## HomeworkReview

Решение:

```text
Entity inside HomeworkAssignment
```

до появления independent review workflow.

## HomeworkBlocker

Решение:

```text
Entity inside HomeworkAssignment
```

Переходит в support case Aggregate только при расширении scope.

## CorrectionRequest

Решение:

```text
Entity inside HomeworkAssignment
```

## ProgressDimensionState

Решение:

```text
Entity inside scoped ProgressRecord
```

Размер ProgressRecord должен ограничиваться scope.

## ReadinessAreaState

Решение:

```text
Entity inside SongReadiness
```

## ConsentState

Решение:

```text
Entity inside ConcertParticipation for MVP
```

Требует review при legal expansion.

## EligibilityState

Решение:

```text
Entity inside ConcertParticipation
```

для strong consistency с Approval.

## ApprovalState

Решение:

```text
Entity inside ConcertParticipation
```

для strong consistency со Slot assignment.

## PerformanceSlot

Решение:

```text
Entity inside ConcertParticipation for MVP
```

Но global scheduling may require ConcertProgram Aggregate.

## ScheduledReminder

Решение:

```text
Entity inside HomeworkReminderPlan
```

пока collection bounded.

## DeliveryAttempt

Решение:

```text
Entity inside NotificationDelivery
```

с возможным archival в operational store.

---

# Non-Goals

Этот документ не определяет:

- database tables;
- foreign keys;
- ORM;
- Go structs;
- API endpoints;
- JSON schemas;
- UI forms;
- exact storage format;
- microservice ownership;
- deployment topology;
- complete Song catalog;
- curriculum definitions;
- CRM entities;
- billing entities;
- payment entities;
- employee HR entities;
- marketing entities;
- legal wording of Consent;
- provider-specific delivery schema.

---

# Open Questions

Необходимо определить:

- нужен ли AttendanceRecordId или достаточно LessonId + StudentId;
- может ли Attendance иметь appeal;
- может ли один Student участвовать в Lesson частично несколько раз;
- нужен ли LessonParticipant Entity отдельно от Attendance;
- сколько historical Submission хранить внутри active Homework Aggregate;
- можно ли иметь несколько active Submission;
- нужен ли server-side Submission Draft;
- как обрабатывать corrupted attachments после Submission;
- нужно ли выделять Submission в Aggregate для media processing;
- может ли Review иметь co-reviewers;
- нужен ли review rubric;
- является ли rubric response Entity;
- как исправлять ошибочный Completed Review;
- нужен ли Review Appeal;
- могут ли Blockers быть private только для Teacher;
- какие blocker categories влияют на deadline;
- может ли Blocker быть подтвержден системой;
- нужен ли support integration;
- сколько correction cycles разрешено;
- когда correction превращается в replacement Homework;
- нужен ли отдельный Correction Deadline type;
- является ли ProgressDimension definition частью curriculum context;
- сколько Dimensions может содержать один ProgressRecord;
- один ProgressRecord на Student или на scope;
- может ли Dimension иметь несколько parallel assessments;
- нужен ли teacher-specific Progress view;
- нужен ли Progress Assessment Entity;
- являются ли Readiness Areas configurable;
- может ли Safety area быть обязательной всегда;
- как хранить history Area evaluations;
- нужен ли отдельный SongReadinessEvaluation Aggregate;
- когда Consent становится отдельным Aggregate;
- как моделировать Guardian Consent;
- нужен ли signed consent artifact;
- может ли Consent покрывать несколько Concerts;
- как обрабатывать consent withdrawal после публикации программы;
- должна ли Eligibility иметь собственный Evaluation Entity;
- сколько Eligibility Decisions хранить внутри Participation;
- можно ли Approval сохранить после Eligibility Stale;
- кто имеет право Approval;
- нужен ли Approval quorum;
- где обеспечить global Slot uniqueness;
- нужен ли ConcertProgram Aggregate в MVP;
- можно ли Participation иметь несколько Slots;
- как моделировать duet/group performance;
- кто владеет Slot для ensemble;
- сколько ScheduledReminder может быть в Plan;
- нужно ли выделять каждый Reminder в Aggregate;
- как синхронизировать Reminder Delivery outcome;
- должен ли Reminder знать NotificationDeliveryId;
- сколько DeliveryAttempt хранить в Aggregate;
- когда архивировать attempts;
- нужен ли separate provider callback inbox;
- какие EntityVersion действительно нужны;
- должны ли все EntityId быть globally unique;
- как сериализовать internal Entity references;
- как защищать Entity rows от direct DAO mutation;
- как автоматически проверять bounded collections;
- нужен ли static architecture test для Entity repositories;
- какой шаблон использовать для отдельных Entity specifications;
- какие Entity войдут в MVP;
- какие Entity можно упростить до Value Object в первой версии;
- какие extraction decisions надо принять до реализации;
- какие Entity должны поддерживать offline mobile commands;
- как разрешать conflict по internal Entity;
- как выполнять migration historical Entity;
- как сохранять unknown lifecycle history;
- как анонимизировать sensitive Entity;
- как хранить Student-visible и internal notes отдельно;
- какие Event payload должны включать EntityVersion;
- когда Entity transition заслуживает отдельного Event;
- когда достаточно Aggregate-level Event;
- как отображать Entity history в Teacher UI.

---

# History

| Version | Description |
| --- | --- |
| 1.0.0 | Определены канонические внутренние Entity Belcanto Product, их Aggregate ownership, identity, lifecycle, invariants, commands, events, privacy, AI restrictions, testing и criteria выделения в отдельный Aggregate Root. |
