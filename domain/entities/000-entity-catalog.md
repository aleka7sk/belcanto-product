---
Status: Accepted
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
