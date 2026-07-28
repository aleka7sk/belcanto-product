---
Status: Approved
Version: 1.0.0
Last Updated: 2026-07-27

Document Id: DOMAIN_COMMAND_CATALOG

Document Type:
  - Domain Contract
  - Command Catalog
  - Authorization Specification
  - Concurrency Specification
  - Traceability Specification

Owners:
  - Product Owner
  - Education Lead
  - Domain Architecture Lead
  - Technical Lead
  - Security Owner

Applies To:
  - Domain Commands
  - Command Handlers
  - Aggregate Mutations
  - Policy Reactions
  - Human Actions
  - Scheduled Processes
  - External Integrations
  - Authorization
  - Idempotency
  - Optimistic Concurrency

Related Directories:
  - ../events/
  - ../policies/
  - ../aggregates/
  - ../../product/
  - ../../school/

Related Documents:
  - ../events/000-domain-event-catalog.md
  - ../policies/000-domain-policy-overview.md
  - ../policies/lesson-completion-policy.md
  - ../policies/progress-update-policy.md
  - ../policies/goal-completion-policy.md
  - ../policies/achievement-award-policy.md
  - ../policies/song-readiness-policy.md
  - ../policies/concert-eligibility-policy.md
  - ../policies/homework-reminder-policy.md
  - ../policies/homework-expiration-policy.md
  - ../policies/notification-policy.md
  - ../policies/periodic-review-policy.md
---

# Domain Command Catalog

> Domain Command Catalog определяет единый канонический набор команд Belcanto Product.
>
> Command выражает намерение изменить доменное состояние или запросить доменное решение.
>
> Команда не гарантирует успех. Она может быть принята, отклонена, отложена или признана уже выполненной.

---

# Purpose

Доменные политики и Aggregate Belcanto используют множество действий:

- CompleteLesson;
- AssignHomework;
- SubmitHomework;
- UpdateProgress;
- CompleteGoal;
- AwardAchievement;
- EvaluateSongReadiness;
- EvaluateConcertEligibility;
- ScheduleHomeworkReminder;
- ExpireHomework;
- SendNotification;
- RequestPeriodicReview.

Если команды определяются локально внутри каждой Policy, возникают риски:

- одинаковые команды имеют разные payload;
- authorization реализуется по-разному;
- Actor может менять чужой Aggregate;
- отсутствует единая idempotency;
- команды выполняются без ExpectedVersion;
- UI напрямую создает события;
- Policy получает слишком широкие полномочия;
- технический retry создает повторное бизнес-действие;
- команда смешивается с событием;
- неясно, кто владеет handler;
- невозможно проследить Command → Decision → Event;
- массовые операции обходят per-aggregate validation.

Этот документ задает единые правила для всех Domain Commands.

---

# Core Principle

Command является просьбой выполнить конкретное действие.

```text
Actor / Policy / Scheduler / Integration
                |
                v
             Command
                |
                v
        Authorization Check
                |
                v
        Load Aggregate State
                |
                v
       Validate Preconditions
                |
                v
          Domain Decision
                |
        +-------+-------+
        |               |
        v               v
     Rejected        State Changed
                         |
                         v
                    Domain Event
```

Command не является фактом.

Корректно:

```text
CompleteLesson
SubmitHomework
ExtendHomeworkDueDate
EvaluateGoalCompletion
AwardAchievement
```

Соответствующие события:

```text
LessonCompleted
HomeworkSubmitted
HomeworkDueDateExtended
GoalCompletionEvaluated
AchievementAwarded
```

---

## Command Semantics

Command отвечает:

> Какое действие Actor или Policy просит выполнить?

Domain Event отвечает:

> Какой факт уже произошел?

Policy отвечает:

> Как принять решение при заданных условиях?

Query отвечает:

> Какие данные требуется прочитать без изменения доменного состояния?

---

## Command Naming

Команды именуются в повелительной форме.

Формат:

`<Verb><EntityOrBusinessAction>`

Примеры:

```text
ScheduleLesson
CompleteLesson
AssignHomework
SubmitHomework
EvaluateGoalCompletion
AwardAchievement
SendNotification
```

Не использовать:

```text
LessonCompleted
HomeworkSubmission
GoalCompletion
NotificationSent
```

Это факты или существительные, а не команды.

---

# Command Envelope

Каждая команда должна использовать единый envelope.

```text
DomainCommandEnvelope
├── CommandId
├── CommandType
├── CommandVersion
├── RequestedAt
├── EffectiveAt
├── TenantId
├── Target
├── Actor
├── AuthorizationContext
├── ExpectedAggregateVersion
├── IdempotencyKey
├── CorrelationId
├── CausationId
├── TraceId
├── PolicyReference
├── Payload
└── Metadata
```

---

# CommandId

Глобально уникальный идентификатор команды.

Требования:

- immutable;
- уникален;
- сохраняется при техническом retry;
- не генерируется заново на каждой попытке;
- используется для traceability;
- может применяться для idempotency.

---

# CommandType

Каноническое имя команды.

Пример:

`SubmitHomework`

---

# CommandVersion

Версия контракта команды.

Она не равна:

- AggregateVersion;
- PolicyVersion;
- EventVersion.

---

# RequestedAt

Время, когда команда была создана или запрошена.

---

# EffectiveAt

Время, когда изменение должно вступить в силу.

Для большинства интерактивных команд:

`EffectiveAt = processing time`

Для запланированных операций EffectiveAt может быть в будущем.

Примеры:

- Learning Pause начинается завтра;
- новый Teacher назначается со следующей недели;
- Homework Due Date изменяется с определенного момента.

Команда не должна использовать EffectiveAt для бесследного изменения прошлого.

---

# TenantId

Явная tenant boundary.

Command Handler должен проверить, что:

- Actor принадлежит разрешенному tenant;
- Target принадлежит тому же tenant;
- referenced entities доступны в tenant;
- cross-tenant mutation запрещена.

---

# Target

```text
CommandTarget
├── AggregateType
├── AggregateId
└── AggregateVersion
```

Для create-команд AggregateId может быть:

- заранее сгенерирован;
- создан handler;
- отсутствовать согласно утвержденному контракту.

Предпочтительно генерировать идентификатор до выполнения команды, если это упрощает idempotency.

---

# Actor

```text
CommandActor
├── ActorType
├── ActorId
├── Role
├── DelegatedBy
├── AuthenticationContextReference
└── ImpersonationContext
```

Допустимые ActorType:

- Student;
- Teacher;
- Administrator;
- Owner;
- Guardian;
- ConcertCoordinator;
- System;
- Policy;
- Scheduler;
- Integration;
- Migration.

---

# Authorization Context

```text
AuthorizationContext
├── AuthenticatedActorId
├── ActiveRole
├── PermissionReferences
├── Scope
├── DelegationReference
├── SessionReference
├── AuthenticationStrength
└── EvaluatedAt
```

Command payload не является источником прав.

Например, поле:

`teacherId: teacher_123`

не доказывает, что текущий Actor является этим Teacher.

---

# ExpectedAggregateVersion

Используется для optimistic concurrency.

Пример:

`ExpectedAggregateVersion: 7`

Если текущая версия Aggregate равна 8, команда не должна применяться как будто она была построена на актуальном состоянии.

Возможные результаты:

- Reject;
- Re-evaluate;
- Return current state;
- Apply only for explicitly mergeable command.

---

# IdempotencyKey

Защищает от повторного бизнес-действия.

Примеры:

```text
StudentId + HomeworkAssignmentId + SubmissionClientOperationId
LessonId + CompletionAttemptId
NotificationIntentId + RecipientId + Channel
ReviewId + AggregateId
```

IdempotencyKey не заменяет CommandId.

---

# CorrelationId

Объединяет команды и события одного бизнес-процесса.

---

# CausationId

Указывает непосредственную причину команды.

Причиной может быть:

- EventId;
- previous CommandId;
- ReviewId;
- scheduled job occurrence;
- external request reference.

---

# PolicyReference

Для команд, инициированных Policy:

```text
PolicyCommandReference
├── PolicyId
├── PolicyVersion
├── DecisionId
├── ReasonCodes
└── EvaluationReference
```

Policy не должна отправлять команды без traceable Decision.

---

# Metadata

Допустимые данные:

- client operation id;
- source application;
- import batch;
- migration marker;
- replay marker;
- scheduled job occurrence;
- technical retry count;
- schema identifier.

Недопустимо:

- secret;
- access token;
- private note;
- произвольное изменение authorization;
- full Aggregate snapshot;
- неструктурированный sensitive content.

---

# Command Outcomes

Каждый Command Handler возвращает канонический результат.

```text
CommandResult
├── CommandId
├── Status
├── TargetAggregateType
├── TargetAggregateId
├── PreviousAggregateVersion
├── CurrentAggregateVersion
├── ProducedEventIds
├── DecisionReference
├── ReasonCodes
├── ValidationErrors
├── Retryable
├── ProcessedAt
└── CorrelationId
```

Допустимые Status:

- Accepted
- Completed
- Rejected
- Deferred
- No Change Required
- Already Processed
- Conflict
- Pending Human Review
- Partially Completed
- Failed

---

## Accepted

Команда принята в асинхронную обработку, но результат еще не завершен.

---

## Completed

Команда успешно обработана.

Completed не обязательно означает изменение Aggregate.

Например, идемпотентная повторная команда может вернуть:

`Already Processed`

---

## Rejected

Команда нарушает доменные, authorization или validation rules.

Повтор без изменения входных данных обычно не поможет.

---

## Deferred

Команда временно не может быть завершена.

Причины:

- временно недоступная dependency;
- ожидается другой доменный факт;
- требуется актуализация состояния;
- активная блокировка;
- scheduled effective time еще не наступило.

---

## No Change Required

Команда валидна, но текущее состояние уже соответствует желаемому результату.

---

## Already Processed

Команда или ее IdempotencyKey уже были успешно обработаны.

---

## Conflict

ExpectedAggregateVersion не совпала с текущей версией или возник другой concurrency conflict.

---

## Pending Human Review

Для завершения требуется решение уполномоченного Actor.

---

## Partially Completed

Разрешено только для явно определенных batch-команд.

Для единичного Aggregate mutation этот статус обычно запрещен.

---

## Failed

Техническая ошибка не позволила завершить обработку.

Следует отличать от Rejected.

---

# Command Categories

## Aggregate Mutation Commands

Изменяют состояние конкретного Aggregate.

Примеры:

```text
CompleteLesson
SubmitHomework
CompleteGoal
WithdrawConcertParticipation
```

---

## Evaluation Commands

Запрашивают применение Policy.

Примеры:

```text
EvaluateGoalCompletion
EvaluateSongReadiness
EvaluateConcertEligibility
EvaluateHomeworkExpiration
```

Оценка может не изменить основной Aggregate.

---

## Review Commands

Создают или завершают Human Review.

Примеры:

```text
RequestGoalReview
RecordGoalReviewDecision
RequestHomeworkExpirationReview
```

---

## Scheduling Commands

Создают будущую работу.

Примеры:

```text
ScheduleHomeworkReminder
ScheduleNotification
SchedulePeriodicReview
```

---

## Delivery Commands

Передают разрешенный Intent инфраструктурной доставке.

Примеры:

```text
SendNotification
RetryNotificationDelivery
```

---

## Administrative Commands

Исправляют операционные или целостностные проблемы.

Примеры:

```text
ResolveDomainIntegrityIssue
ArchiveHomework
CancelScheduledNotification
```

---

## Migration Commands

Используются только утвержденным migration process.

Они должны быть явно маркированы и не маскироваться под обычные пользовательские команды.

---

# General Command Rules

## DC-001: Command expresses intent, not fact

Command не должен называться в прошедшем времени.

---

## DC-002: Every mutation requires a command

Состояние Aggregate не изменяется:

- напрямую из UI;
- database script без migration contract;
- event consumer без команды;
- notification callback;
- analytics process;
- AI proposal.

---

## DC-003: Command must have one primary target

Команда должна иметь одного владельца и один основной Aggregate target.

Связанные изменения других Aggregate выполняются отдельными командами.

---

## DC-004: Command handler belongs to target owner

Handler команды должен находиться внутри bounded context или модуля, владеющего Target Aggregate.

---

## DC-005: Authorization precedes domain mutation

Authorization проверяется до изменения состояния.

---

## DC-006: Domain validation still applies after authorization

Право отправить команду не гарантирует, что действие допустимо в текущем состоянии.

---

## DC-007: Command must use current state

Handler загружает authoritative Aggregate state.

Payload не должен считаться текущим snapshot.

---

## DC-008: Expected version is required for user-driven updates

Для изменения существующего Aggregate интерактивная команда должна обычно содержать ExpectedAggregateVersion.

Исключения должны быть документированы.

---

## DC-009: Technical retry preserves command identity

Retry не создает новый CommandId.

---

## DC-010: Business retry may create a new command

После изменения намерения, payload или текущего состояния создается новая команда.

---

## DC-011: Duplicate command must not duplicate business effect

---

## DC-012: Successful command publishes domain events

Если состояние изменилось, должны быть созданы соответствующие Domain Events.

---

## DC-013: Rejected command must not publish success event

Можно создать отдельный audit record или rejection event, если он имеет доменную ценность.

---

## DC-014: No state change should not emit mutation event

Кроме явно определенных audit или evaluation events.

---

## DC-015: Commands cannot bypass policies

Например, AwardAchievement не должен обходить Achievement Award Policy.

---

## DC-016: Policy cannot grant itself unlimited authority

Policy Actor имеет только явно разрешенный command scope.

---

## DC-017: AI cannot be authoritative actor

AI может предложить Command Draft, но authoritative Actor должен быть:

- Human Actor;
- approved deterministic Policy;
- authorized System process.

---

## DC-018: Command payload must be minimal

Не следует передавать весь Aggregate snapshot.

---

## DC-019: Sensitive values should use references

Например:

- attachment reference;
- private note reference;
- assessment reference;
- consent reference.

---

## DC-020: Command must not trust client timestamps blindly

Client-provided time может быть:

- сохранено как reported time;
- проверено;
- ограничено;
- сопоставлено с server time.

---

## DC-021: Commands cannot rewrite immutable history

Для исправления создается новое решение, версия или compensating action.

---

## DC-022: Cancellation is a command, not deletion

---

## DC-023: Reopen must preserve prior terminal history

---

## DC-024: Batch commands require per-item validation

---

## DC-025: Partial completion must be explicit

---

## DC-026: Command result must be explainable

Rejected или Deferred result должен иметь Reason Codes.

---

## DC-027: Internal Reason Codes must not leak directly to end users

UI использует безопасное human-readable explanation.

---

## DC-028: Command authorization must be auditable

---

## DC-029: Cross-tenant command is rejected

---

## DC-030: Command version must be validated

---

# Authorization Model

Authorization должна учитывать:

```text
Actor
+
Role
+
Relationship to Target
+
Tenant
+
Command Type
+
Current Aggregate State
+
Delegation
+
Consent Scope
```

Пример:

Teacher может иметь право:

`ReviewHomework`

только если:

- он назначен Student;
- он участвует в соответствующем Lesson или учебном контексте;
- Assignment принадлежит школе;
- delegation не истекла;
- Homework находится в допустимом статусе.

---

# Actor Authority Guidance

| Actor | Typical Allowed Scope |
| --- | --- |
| Student | собственные Submission, consent, blocker, extension request |
| Teacher | assigned Students, Lessons, Homework, Goals, evaluations |
| Administrator | operational scheduling and technical corrections |
| Owner | configuration, definitions, reporting, governance |
| Guardian | explicitly delegated consent and limited communication |
| Concert Coordinator | concert program, slots, operational approval |
| Policy | narrow deterministic reaction commands |
| Scheduler | due evaluation and periodic review triggers |
| Integration | explicitly mapped external operations |
| Migration | approved migration-only commands |

---

# Delegation

Delegated action должна содержать:

- original authority;
- delegated Actor;
- command scope;
- target scope;
- effective period;
- revocation status;
- audit reference.

Delegation не должна расширять исходные полномочия.

---

# Impersonation

Staff impersonation, если вообще поддерживается, требует:

- explicit reason;
- elevated authorization;
- visible audit;
- limited duration;
- original Actor identity;
- prohibition for sensitive actions where required.

System не должен записывать impersonated Actor как исходного Student.

---

# Concurrency

Рекомендуемый процесс:

```text
Receive Command
      |
      v
Load Aggregate version 7
      |
      v
Validate ExpectedVersion = 7
      |
      v
Apply Decision
      |
      v
Persist version 8 + Events
```

При mismatch:

```text
ExpectedVersion: 7
CurrentVersion: 8
```

Результат: Conflict

или повторная Policy evaluation, если команда является системной и допускает re-evaluation.

---

# Mergeable Commands

Некоторые команды могут быть mergeable.

Пример:

`ReportHomeworkBlocker`

может быть обработана независимо от некоторых несвязанных изменений Homework.

Но mergeable behavior должен быть явно определен.

Нельзя считать все команды mergeable по умолчанию.

---

# Command Idempotency

## Natural Idempotency

Состояние уже соответствует цели.

Пример:

`CancelHomework`

для уже Cancelled Homework может вернуть:

`No Change Required`

---

## Key-Based Idempotency

Используется IdempotencyKey.

Пример:

`SubmitHomework`

с одним client operation id не создает две Submission.

---

## Command-Based Idempotency

Повтор одного CommandId возвращает сохраненный результат.

---

# Command Processing Guarantees

Предпочтительная модель:

```text
At-Least-Once Command Delivery
+
Idempotent Command Handling
+
Atomic State and Event Persistence
```

Exactly-once transport не предполагается.

---

# Command Persistence

Для долгих или асинхронных команд может храниться:

```text
CommandExecution
├── CommandId
├── CommandType
├── Status
├── Target
├── AttemptCount
├── RequestedAt
├── StartedAt
├── CompletedAt
├── LastFailure
├── ProducedEventIds
└── ResultReference
```

Не все интерактивные команды обязаны становиться отдельным Aggregate.

---

# Canonical Command Groups

## Student Commands

### CreateStudentProfile

Создает product-profile Student после подтвержденного enrollment boundary.

Payload:

```text
StudentId
EnrollmentReference
InitialLocale
InitialTimezone
```

Allowed Actors:

- authorized Integration;
- Administrator;
- Migration.

Не является CRM lead creation.

---

### UpdateStudentProfile

Изменяет только разрешенные product-profile fields.

Payload:

```text
StudentId
ExpectedStudentVersion
ChangedFields
```

Sensitive поля требуют отдельного authorization.

---

### ChangeStudentTimezone

Payload:

```text
StudentId
PreviousTimezone
NewTimezone
EffectiveAt
```

Consumers после события:

- Reminder Policy;
- Notification Policy;
- Lesson presentation.

---

### StartStudentLearningPause

Payload:

```text
StudentId
LearningPauseId
StartedAt
ExpectedEndAt
Scope
StudentVisibleReasonCategory
```

Allowed Actors:

- Student, если self-service разрешен;
- Teacher;
- Administrator;
- authorized Policy.

---

### EndStudentLearningPause

Payload:

```text
StudentId
LearningPauseId
EndedAt
EndReason
```

---

## Teacher Assignment Commands

### AssignTeacherToStudent

Payload:

```text
StudentId
TeacherId
AssignmentType
EffectiveFrom
```

---

### ReassignTeacher

Payload:

```text
StudentId
PreviousTeacherId
NewTeacherId
EffectiveAt
ReasonCategory
```

Private HR reason не входит в payload общего доменного контракта.

---

## Lesson Commands

### ScheduleLesson

Payload:

```text
LessonId
StudentIds
TeacherId
LessonType
ScheduledStart
ScheduledEnd
Timezone
Format
LocationReference
```

Validation:

- Teacher доступен;
- Student scope валиден;
- end позже start;
- timezone известна;
- location или online format корректны;
- конфликт расписания обработан согласно Scheduling Policy.

---

### RescheduleLesson

Payload:

```text
LessonId
ExpectedLessonVersion
PreviousScheduledStart
NewScheduledStart
NewScheduledEnd
Timezone
ReasonCategory
```

Команда не должна молча изменять связанные Homework deadlines.

Она создает LessonRescheduled, после чего соответствующие Policy переоценивают зависимости.

---

### CancelLesson

Payload:

```text
LessonId
ExpectedLessonVersion
CancellationCategory
ReplacementLessonId
StudentVisibleExplanation
```

---

### StartLesson

Payload:

```text
LessonId
ExpectedLessonVersion
StartedAt
```

---

### CompleteLesson

Payload:

```text
LessonId
ExpectedLessonVersion
CompletedAt
CompletionMethod
StudentAttendanceReferences
LessonSummaryReference
EvidenceReferences
```

Handler должен применять Lesson Completion Policy.

---

### RecordLessonSummary

Payload:

```text
LessonId
SummaryId
SummaryReference
Visibility
RecordedAt
```

Полный текст не должен публиковаться в event payload.

---

## Homework Commands

### AssignHomework

Payload:

```text
HomeworkAssignmentId
StudentId
TeacherId
HomeworkDefinitionId
HomeworkVersion
AssignedAt
DueDate
DeadlineType
Requiredness
ReminderStrategy
ExpirationStrategy
RelatedLessonId
RelatedGoalIds
RelatedSongVersionIds
RelatedConcertId
MaterialReferences
InstructionReference
```

Validation:

- Teacher authorized;
- Student активен;
- references доступны;
- срок корректен;
- strategy согласована;
- Assignment не дублирует replacement context;
- required materials существуют.

---

### UpdateHomework

Payload:

```text
HomeworkAssignmentId
ExpectedHomeworkVersion
ChangedFields
ChangeReason
```

Изменения, требующие отдельной команды, не должны скрываться внутри generic update.

Например:

- ChangeHomeworkDueDate;
- CancelHomework;
- ReplaceHomework.

---

### ChangeHomeworkDueDate

Payload:

```text
HomeworkAssignmentId
ExpectedHomeworkVersion
PreviousDueDate
NewDueDate
Timezone
ChangeReason
```

---

### StartHomework

Фиксирует явное начало работы.

Payload:

```text
HomeworkAssignmentId
StudentId
ExpectedHomeworkVersion
StartedAt
StartMethod
```

Открытие карточки не должно автоматически вызывать эту команду.

---

### SubmitHomework

Payload:

```text
HomeworkAssignmentId
ExpectedHomeworkVersion
SubmissionId
SubmissionVersion
StudentId
SubmittedAt
SubmissionMethod
AttachmentReferences
TextSubmissionReference
ClientOperationId
```

Validation:

- Student владеет Assignment;
- Homework допускает Submission;
- версия актуальна;
- terminal status обработан;
- late submission policy применена;
- attachment references безопасны;
- повторная операция идемпотентна.

---

### WithdrawHomeworkSubmission

Payload:

```text
HomeworkAssignmentId
SubmissionId
ExpectedSubmissionVersion
WithdrawalReason
```

Разрешено только при отдельном правиле.

---

### StartHomeworkReview

Payload:

```text
HomeworkAssignmentId
SubmissionId
TeacherId
StartedAt
```

---

### ReviewHomework

Payload:

```text
HomeworkAssignmentId
ExpectedHomeworkVersion
SubmissionId
ExpectedSubmissionVersion
ReviewId
ReviewOutcome
FeedbackReference
EvidenceReferences
ReviewedAt
```

---

### RequestHomeworkCorrection

Payload:

```text
HomeworkAssignmentId
SubmissionId
ReviewId
CorrectionRequestId
CorrectionReference
NewDueDate
DeadlineType
RequestedAt
```

---

### RequestHomeworkClarification

Payload:

```text
HomeworkAssignmentId
StudentId
ClarificationRequestId
ClarificationCategory
QuestionReference
RequestedAt
```

---

### ReportHomeworkBlocker

Payload:

```text
HomeworkAssignmentId
StudentId
BlockerId
BlockerCategory
ExplanationReference
ReportedAt
```

Sensitive explanation хранится отдельно.

---

### ResolveHomeworkBlocker

Payload:

```text
HomeworkAssignmentId
BlockerId
ResolutionCategory
ResolutionReference
ResolvedAt
```

---

### MarkHomeworkOverdue

Обычно вызывается Homework Expiration Policy.

Payload:

```text
HomeworkAssignmentId
ExpectedHomeworkVersion
DueDate
GracePeriodEnd
DeadlineType
DecisionId
ReasonCodes
```

---

### StartHomeworkGracePeriod

Payload:

```text
HomeworkAssignmentId
ExpectedHomeworkVersion
GracePeriodStart
GracePeriodEnd
DecisionId
```

---

### ExtendHomeworkDueDate

Payload:

```text
HomeworkAssignmentId
ExpectedHomeworkVersion
PreviousDueDate
NewDueDate
Timezone
ExtensionReason
RequestedBy
ApprovedBy
DecisionId
```

---

### EvaluateHomeworkExpiration

Payload:

```text
HomeworkAssignmentId
HomeworkVersion
EvaluationId
EvaluationTime
TriggerReference
```

Эта команда запрашивает Policy evaluation, а не гарантирует Expiration.

---

### RequestHomeworkExpirationReview

Payload:

```text
HomeworkAssignmentId
HomeworkVersion
ReviewId
ReasonCodes
RequestedAt
```

---

### ExpireHomework

Payload:

```text
HomeworkAssignmentId
ExpectedHomeworkVersion
ExpirationCategory
DecisionId
ReasonCodes
EffectiveAt
```

Allowed Actors:

- authorized Teacher;
- deterministic Policy при разрешенном auto-expiration;
- Administrator только в техническом scope.

---

### ReopenHomework

Payload:

```text
HomeworkAssignmentId
ExpectedHomeworkVersion
PreviousExpirationDecisionId
NewHomeworkVersion
NewDueDate
ReopenReason
ReopenedAt
```

---

### CancelHomework

Payload:

```text
HomeworkAssignmentId
ExpectedHomeworkVersion
CancellationCategory
StudentVisibleExplanation
CancelledAt
```

---

### ReplaceHomework

Payload:

```text
HomeworkAssignmentId
ExpectedHomeworkVersion
ReplacementHomeworkAssignmentId
ReplacementReason
ReplacedAt
```

---

### CompleteHomework

Payload:

```text
HomeworkAssignmentId
ExpectedHomeworkVersion
SubmissionId
ReviewId
CompletionMethod
CompletedAt
```

Completion должна соответствовать отдельным completion criteria.

---

### ArchiveHomework

Payload:

```text
HomeworkAssignmentId
ExpectedHomeworkVersion
ArchiveReason
ArchivedAt
```

Archive не заменяет Expiration или Cancellation.

---

## Progress Commands

### RecordProgressEvidence

Payload:

```text
ProgressEvidenceId
StudentId
EvidenceType
SourceEntityType
SourceEntityId
ValidityScope
RecordedAt
```

Allowed Producers:

- Teacher;
- approved Policy;
- authorized Assessment process;
- trusted Integration.

Notification interaction не является допустимым Evidence.

---

### InvalidateProgressEvidence

Payload:

```text
ProgressEvidenceId
InvalidationReason
SupersedingEvidenceId
InvalidatedAt
```

---

### EvaluateProgressUpdate

Payload:

```text
ProgressEvaluationId
StudentId
CurrentProgressVersion
EvaluationScope
EvidenceReferences
TriggerReference
```

---

### UpdateProgress

Payload:

```text
ProgressId
StudentId
ExpectedProgressVersion
ProgressEvaluationId
ChangedDimensions
EvidenceReferences
DecisionId
```

Нельзя передавать arbitrary final score без Evidence и Policy reference.

---

### RequestProgressReview

Payload:

```text
ProgressReviewId
StudentId
Scope
ReasonCodes
RequestedAt
```

---

## Goal Commands

### CreateGoal

Payload:

```text
GoalId
StudentId
GoalType
CriterionReference
CreatedAt
TargetReviewAt
RelatedSongVersionIds
RelatedSkillIds
RelatedConcertId
```

---

### ActivateGoal

Payload:

```text
GoalId
ExpectedGoalVersion
ActivatedAt
```

---

### UpdateGoal

Payload:

```text
GoalId
ExpectedGoalVersion
ChangedFields
ChangeReason
```

Критическое изменение Criterion может потребовать новой Goal version или replacement.

---

### UpdateGoalProgress

Payload:

```text
GoalId
ExpectedGoalVersion
PreviousProgressState
NewProgressState
EvidenceReferences
```

---

### RequestGoalReview

Payload:

```text
GoalId
GoalVersion
GoalReviewId
ReasonCodes
RequestedAt
```

---

### EvaluateGoalCompletion

Payload:

```text
GoalId
GoalVersion
EvaluationId
EvidenceReferences
TriggerReference
EvaluatedAt
```

---

### CompleteGoal

Payload:

```text
GoalId
ExpectedGoalVersion
CompletionDecisionId
EvidenceReferences
CompletedAt
```

Allowed only after Goal Completion Policy decision.

---

### ReopenGoal

Payload:

```text
GoalId
ExpectedGoalVersion
PreviousCompletionDecisionId
ReopenReason
NewGoalVersion
ReopenedAt
```

---

### CancelGoal

Payload:

```text
GoalId
ExpectedGoalVersion
CancellationCategory
CancelledAt
```

---

### ArchiveGoal

Payload:

```text
GoalId
ExpectedGoalVersion
ArchiveReason
ArchivedAt
```

---

## Achievement Commands

### PublishAchievementDefinition

Payload:

```text
AchievementDefinitionId
DefinitionVersion
CriterionReference
Visibility
PublishedAt
```

---

### EvaluateAchievementEligibility

Payload:

```text
AchievementDefinitionId
AchievementDefinitionVersion
StudentId
EvaluationId
EvidenceReferences
TriggerReference
```

---

### AwardAchievement

Payload:

```text
AchievementAwardId
AchievementDefinitionId
AchievementDefinitionVersion
StudentId
AwardDecisionId
EvidenceReferences
AwardedAt
```

Allowed only after Achievement Award Policy.

---

### RejectAchievementAward

Используется только если rejection является значимым доменным решением.

Payload:

```text
AchievementDefinitionId
StudentId
EvaluationId
ReasonCodes
RejectedAt
```

---

### RevokeAchievement

Payload:

```text
AchievementAwardId
ExpectedAwardVersion
RevocationReason
EvidenceReferences
ReplacementAwardId
RevokedAt
```

Revocation не удаляет Award history.

---

### RestoreAchievement

Payload:

```text
AchievementAwardId
ExpectedAwardVersion
RestoreReason
RestoredAt
```

---

## Song Commands

### AddSongToStudentRepertoire

Payload:

```text
StudentSongId
StudentId
SongId
SongVersionId
Purpose
AddedAt
```

---

### CreateSongVersion

Payload:

```text
SongVersionId
SongId
VersionNumber
Key
ArrangementReference
BackingTrackReference
LyricsReference
CreatedAt
```

---

### ChangeStudentSongVersion

Payload:

```text
StudentSongId
ExpectedStudentSongVersion
PreviousSongVersionId
NewSongVersionId
ChangeReason
ChangedAt
```

После изменения связанные Homework и Readiness должны пройти review.

---

### AddSongReadinessEvidence

Payload:

```text
SongReadinessEvidenceId
StudentId
SongVersionId
ReadinessArea
EvidenceType
SourceReference
RecordedAt
```

---

### InvalidateSongReadinessEvidence

Payload:

```text
SongReadinessEvidenceId
InvalidationReason
InvalidatedAt
```

---

### RequestSongReadinessEvaluation

Payload:

```text
SongReadinessEvaluationId
StudentId
SongVersionId
PerformanceType
ReasonCodes
RequestedAt
```

---

### EvaluateSongReadiness

Payload:

```text
SongReadinessEvaluationId
StudentId
SongVersionId
PerformanceType
EvidenceReferences
EvaluationContext
```

---

### ChangeSongReadiness

Payload:

```text
StudentId
SongVersionId
PerformanceType
ExpectedReadinessVersion
PreviousReadinessStatus
NewReadinessStatus
EvaluationId
DecisionId
```

Не может выполняться AI самостоятельно.

---

### RequestSongReadinessReview

Payload:

```text
StudentId
SongVersionId
PerformanceType
ReviewId
ReasonCodes
RequestedAt
```

---

## Concert Commands

### CreateConcert

Payload:

```text
ConcertId
Title
ScheduledStart
ScheduledEnd
Timezone
VenueReference
ConcertType
CreatedAt
```

---

### UpdateConcert

Payload:

```text
ConcertId
ExpectedConcertVersion
ChangedFields
ChangeReason
```

Material changes должны создавать новые notifications и reevaluations.

---

### CancelConcert

Payload:

```text
ConcertId
ExpectedConcertVersion
CancellationCategory
CancelledAt
StudentVisibleExplanation
```

---

### CompleteConcert

Payload:

```text
ConcertId
ExpectedConcertVersion
CompletedAt
CompletionMethod
```

---

### PublishConcertRequirements

Payload:

```text
ConcertId
ExpectedConcertVersion
ConcertRequirementsVersion
RequirementsReference
PublishedAt
```

---

### ProposeConcertParticipation

Payload:

```text
ConcertParticipationId
ConcertId
StudentId
PerformanceType
SongVersionIds
ProposedAt
```

Proposal не означает Eligibility, Approval или Program placement.

---

### RequestConcertConsent

Payload:

```text
ConcertParticipationId
ConsentRequestId
ConsentScope
RequestedAt
ExpiresAt
```

---

### GrantConcertConsent

Payload:

```text
ConcertParticipationId
ConsentId
GrantedBy
ConsentScope
GrantedAt
```

---

### WithdrawConcertConsent

Payload:

```text
ConcertParticipationId
ConsentId
WithdrawnAt
```

---

### RequestConcertEligibilityEvaluation

Payload:

```text
ConcertEligibilityEvaluationId
ConcertParticipationId
ConcertId
StudentId
SongVersionIds
PerformanceType
ReasonCodes
RequestedAt
```

---

### EvaluateConcertEligibility

Payload:

```text
ConcertEligibilityEvaluationId
ConcertParticipationId
ConcertRequirementsVersion
SongReadinessReferences
EvidenceReferences
EvaluationContext
```

---

### MarkConcertParticipationEligible

Payload:

```text
ConcertParticipationId
ExpectedParticipationVersion
EligibilityDecisionId
Conditions
MarkedAt
```

---

### MarkConcertParticipationConditionallyEligible

Payload:

```text
ConcertParticipationId
ExpectedParticipationVersion
EligibilityDecisionId
Conditions
ConditionDeadline
MarkedAt
```

---

### MarkConcertParticipationNotEligible

Payload:

```text
ConcertParticipationId
ExpectedParticipationVersion
EligibilityDecisionId
BlockingConditions
MarkedAt
```

Student-facing explanation должна быть нейтральной и ограниченной.

---

### ApproveConcertParticipation

Payload:

```text
ConcertParticipationId
ExpectedParticipationVersion
ApprovalScope
ApprovedAt
```

Approval отделено от Eligibility.

---

### WithdrawConcertParticipation

Payload:

```text
ConcertParticipationId
ExpectedParticipationVersion
WithdrawalCategory
WithdrawnAt
```

---

### AssignConcertPerformanceSlot

Payload:

```text
ConcertParticipationId
ExpectedParticipationVersion
PerformanceSlotId
ScheduledStart
StageReference
AssignedAt
```

---

### ChangeConcertPerformanceSlot

Payload:

```text
ConcertParticipationId
PerformanceSlotId
PreviousScheduledStart
NewScheduledStart
ChangeReason
ChangedAt
```

---

### PublishConcertProgram

Payload:

```text
ConcertId
ExpectedConcertVersion
ProgramVersion
ProgramReference
PublishedAt
```

---

### CompleteConcertPerformance

Payload:

```text
ConcertParticipationId
PerformanceSlotId
CompletedAt
CompletionMethod
```

Performance completion не должна автоматически означать высокий Progress или Achievement.

---

## Reminder Commands

### CreateHomeworkReminderPlan

Payload:

```text
ReminderPlanId
HomeworkAssignmentId
HomeworkVersion
StudentId
Strategy
MaximumReminderCount
Timezone
CreatedAt
```

---

### ScheduleHomeworkReminder

Payload:

```text
ReminderId
ReminderPlanId
HomeworkAssignmentId
HomeworkVersion
StudentId
ReminderType
ScheduledFor
Timezone
TemplateId
RequestedChannels
```

---

### RescheduleHomeworkReminder

Payload:

```text
ReminderId
ExpectedReminderVersion
PreviousScheduledFor
NewScheduledFor
RescheduleReason
```

---

### EvaluateHomeworkReminder

Payload:

```text
ReminderId
ReminderPlanId
HomeworkAssignmentId
HomeworkVersion
EvaluationTime
TriggerReference
```

---

### RequestHomeworkReminderDelivery

Payload:

```text
ReminderId
ReminderPlanId
HomeworkAssignmentId
HomeworkVersion
NotificationIntentId
RequestedAt
```

Эта команда должна выполняться только после send-time revalidation.

---

### SuppressHomeworkReminder

Payload:

```text
ReminderId
ExpectedReminderVersion
SuppressionReason
SuppressedAt
```

---

### CancelHomeworkReminder

Payload:

```text
ReminderId
ExpectedReminderVersion
CancellationReason
CancelledAt
```

---

### ExpireHomeworkReminder

Payload:

```text
ReminderId
ExpectedReminderVersion
ExpirationReason
ExpiredAt
```

---

### RecalculateHomeworkReminderPlan

Payload:

```text
ReminderPlanId
HomeworkAssignmentId
CurrentHomeworkVersion
RecalculationReason
RequestedAt
```

---

## Notification Commands

### CreateNotificationIntent

Payload:

```text
NotificationIntentId
IntentType
SourceDomain
SourceEntityType
SourceEntityId
SourceEntityVersion
RecipientType
RecipientId
Category
Priority
Urgency
RequiredAction
RequestedChannels
DeliveryWindow
ExpiresAt
TemplateReference
RenderingParametersReference
PrivacyLevel
DeduplicationKey
```

---

### ApproveNotificationIntent

Payload:

```text
NotificationIntentId
ExpectedIntentVersion
DecisionId
ApprovedAt
```

---

### RejectNotificationIntent

Payload:

```text
NotificationIntentId
ExpectedIntentVersion
ReasonCodes
RejectedAt
```

---

### ScheduleNotification

Payload:

```text
NotificationDeliveryId
NotificationIntentId
RecipientId
Channel
ScheduledFor
ExpiresAt
IdempotencyKey
```

---

### RescheduleNotification

Payload:

```text
NotificationDeliveryId
ExpectedDeliveryVersion
PreviousScheduledFor
NewScheduledFor
RescheduleReason
```

---

### CancelNotification

Payload:

```text
NotificationIntentId
NotificationDeliveryId
CancellationReason
CancelledAt
```

---

### SuppressNotification

Payload:

```text
NotificationIntentId
RecipientId
SuppressionReason
SuppressedAt
```

---

### BundleNotifications

Payload:

```text
NotificationBundleId
RecipientId
IncludedIntentIds
BundleType
ScheduledFor
Channel
TemplateId
```

---

### RenderNotification

Payload:

```text
NotificationDeliveryId
NotificationIntentId
TemplateId
TemplateVersion
Locale
Channel
RenderingParametersReference
```

---

### SendNotification

Payload:

```text
NotificationDeliveryId
NotificationIntentId
RecipientId
Channel
RenderedContentReference
DestinationReference
IdempotencyKey
ExpiresAt
```

Delivery adapter не должен заново принимать доменное решение.

---

### RetryNotificationDelivery

Payload:

```text
NotificationDeliveryId
ExpectedDeliveryVersion
AttemptNumber
ScheduledFor
RetryUntil
FailureCategory
```

---

### StopNotificationRetry

Payload:

```text
NotificationDeliveryId
ExpectedDeliveryVersion
StopReason
StoppedAt
```

---

### SwitchNotificationChannel

Payload:

```text
NotificationDeliveryId
ExpectedDeliveryVersion
PreviousChannel
NewChannel
SwitchReason
```

---

### MarkNotificationDelivered

Обычно вызывается trusted provider adapter.

Payload:

```text
NotificationDeliveryId
ProviderReference
DeliveredAt
ProviderCallbackReference
```

Callback должен быть аутентифицирован и идемпотентен.

---

### MarkNotificationDeliveryFailed

Payload:

```text
NotificationDeliveryId
FailureCategory
FailureCode
FailedAt
Retryable
ProviderCallbackReference
```

---

### MarkNotificationOpened

Payload:

```text
NotificationDeliveryId
RecipientId
OpenedAt
OpenSource
```

Не является Progress Evidence.

---

### CompleteNotificationAction

Payload:

```text
NotificationDeliveryId
NotificationIntentId
ActionType
SourceEntityId
CompletedAt
```

Domain action все равно должна быть выполнена отдельной domain command.

---

### ExpireNotification

Payload:

```text
NotificationIntentId
NotificationDeliveryId
ExpirationReason
ExpiredAt
```

---

### ArchiveNotification

Payload:

```text
NotificationIntentId
ArchiveReason
ArchivedAt
```

---

### RequestNotificationReview

Payload:

```text
NotificationIntentId
ReviewId
ReasonCodes
RequestedAt
```

---

## Periodic Review Commands

### StartPeriodicReviewCycle

Payload:

```text
ReviewCycleId
ReviewCategory
Scope
StartedAt
ScheduledOccurrenceId
```

---

### DiscoverPeriodicReviewItems

Payload:

```text
ReviewCycleId
ReviewCategory
Cursor
BatchSize
SnapshotBoundary
```

Discovery не должна изменять Aggregate.

---

### RequestPeriodicReview

Payload:

```text
ReviewId
ReviewCategory
AggregateType
AggregateId
AggregateVersion
ReasonCodes
RequestedAt
```

---

### CompletePeriodicReview

Payload:

```text
ReviewId
ReviewCategory
AggregateType
AggregateId
Outcome
RequestedCommandIds
CompletedAt
```

---

### FailPeriodicReview

Payload:

```text
ReviewId
FailureCategory
Retryable
FailedAt
```

---

### CompletePeriodicReviewCycle

Payload:

```text
ReviewCycleId
DiscoveredCount
RequestedReviewCount
SkippedCount
FailedCount
CompletedAt
```

---

## Integrity Commands

### RequestDomainIntegrityReview

Payload:

```text
IntegrityReviewId
IntegrityIssueId
AggregateType
AggregateId
EvidenceReferences
RequestedAt
```

---

### ResolveDomainIntegrityIssue

Payload:

```text
IntegrityIssueId
ResolutionCategory
ResolutionReference
ResolvedAt
```

Это не generic bypass для изменения Aggregate.

Исправление должно выполняться разрешенными domain commands или утвержденной migration.

---

# Command Validation Order

Рекомендуемый порядок:

1. Parse Command
2. Validate Command Type and Version
3. Validate Envelope
4. Authenticate Actor
5. Validate Tenant
6. Authorize Command Scope
7. Check Idempotency
8. Load Aggregate
9. Validate Expected Version
10. Validate Domain Preconditions
11. Apply Policy or Aggregate Decision
12. Persist State and Events Atomically
13. Store Command Result
14. Publish Outbox

Порядок может оптимизироваться технически, но security и domain guarantees должны сохраняться.

---

# Command Rejection Categories

- Validation Rejection
- Authorization Rejection
- Tenant Rejection
- Not Found
- State Rejection
- Policy Rejection
- Version Conflict
- Duplicate
- Expired Request
- Unsupported Command Version
- Security Rejection

---

# Canonical Reason Codes

Общие Reason Codes:

```text
COMMAND_TYPE_UNKNOWN
COMMAND_VERSION_UNSUPPORTED
COMMAND_PAYLOAD_INVALID
COMMAND_ACTOR_REQUIRED
COMMAND_NOT_AUTHORIZED
COMMAND_SCOPE_NOT_ALLOWED
COMMAND_TENANT_MISMATCH
COMMAND_TARGET_NOT_FOUND
COMMAND_EXPECTED_VERSION_REQUIRED
COMMAND_VERSION_CONFLICT
COMMAND_ALREADY_PROCESSED
COMMAND_IDEMPOTENCY_KEY_REQUIRED
COMMAND_EFFECTIVE_TIME_INVALID
COMMAND_SOURCE_REFERENCE_REQUIRED
COMMAND_POLICY_DECISION_REQUIRED
COMMAND_HUMAN_REVIEW_REQUIRED
COMMAND_TARGET_STATE_INVALID
COMMAND_CROSS_TENANT_REFERENCE
COMMAND_SENSITIVE_DATA_NOT_ALLOWED
```

Domain-specific Reason Codes остаются в соответствующих Policy.

---

# Human Review

Human Review требуется, если:

- Policy не может принять deterministic decision;
- AI предложил значимое изменение;
- sensitive Student outcome неоднозначен;
- требуется Reopen terminal Aggregate;
- меняется hard deadline после его завершения;
- Achievement отзывается;
- Concert Eligibility спорна;
- Guardian scope неоднозначен;
- bulk command затрагивает много Student;
- требуется bypass Quiet Hours;
- нарушена data integrity;
- migration изменяет исторические данные;
- authorization основана на delegation с неопределенным scope.

---

# Human Review Commands

Рекомендуемый общий паттерн:

```text
Request<Domain>Review
Record<Domain>ReviewDecision
```

Review Decision должна содержать:

```text
ReviewId
ReviewerId
Outcome
ReasonCodes
DecisionReference
DecidedAt
ExpectedTargetVersion
```

Review не изменяет Aggregate напрямую без domain command.

---

# Batch Commands

Batch Command должна содержать:

```text
BatchCommandId
ItemCommands
RequestedBy
RequestedAt
ExecutionMode
FailureMode
```

Допустимые ExecutionMode:

- ValidateOnly
- Execute
- DryRun

Допустимые FailureMode:

- StopOnFirstFailure
- ContinuePerItem
- AtomicAllOrNothing

AtomicAllOrNothing между множеством Aggregate обычно не рекомендуется без сильной причины.

---

# Batch Safety

Перед batch execution:

- показать audience или target set;
- зафиксировать query snapshot;
- проверить permissions;
- выполнить dry run;
- оценить volume;
- проверить sensitive actions;
- ограничить concurrency;
- поддерживать cancellation;
- сохранить per-item result;
- не обходить per-item ExpectedVersion.

---

# Scheduled Commands

Scheduler не принимает педагогические решения.

Он может отправлять:

```text
EvaluateHomeworkExpiration
EvaluateHomeworkReminder
RequestPeriodicReview
EvaluateNotificationDelivery
```

Scheduler не должен напрямую отправлять:

```text
ExpireHomework
CompleteGoal
AwardAchievement
```

если только это не является явно утвержденным deterministic policy outcome, уже зафиксированным Decision.

---

# Event-Triggered Commands

Event consumer должен:

- проверить EventVersion;
- обработать Event идемпотентно;
- загрузить собственный контекст;
- принять собственное решение;
- создать Command;
- сохранить CausationId;
- не изменять Aggregate напрямую.

Пример:

```text
HomeworkSubmitted
        |
        v
Homework Reminder Consumer
        |
        v
CancelHomeworkReminder
```

---

# Command and Event Traceability

Каждая успешная цепочка должна быть прослеживаема:

```text
CommandId
    |
    v
DecisionId
    |
    v
AggregateVersion
    |
    v
EventId
```

Для реактивного процесса:

```text
EventId
    |
    v
Reaction CommandId
    |
    v
New DecisionId
    |
    v
New EventId
```

---

# Command Privacy

Command может содержать более чувствительные данные, чем Event, но payload все равно должен минимизироваться.

Не следует передавать:

- medical narrative;
- full private Teacher note;
- full attachment content;
- credentials;
- provider secrets;
- raw contact data без необходимости;
- unrelated Student profile.

Sensitive content передается через защищенную reference model.

---

# Command Security

Необходимо защищать:

- forged Actor;
- altered TenantId;
- replay attack;
- duplicate submission;
- stale version mutation;
- unauthorized Teacher access;
- Guardian scope escalation;
- Policy overreach;
- scheduler privilege escalation;
- mass command abuse;
- command payload injection;
- attachment reference substitution;
- provider callback forgery;
- migration misuse;
- command log leakage.

---

# Command Audit

Для каждой команды сохраняются:

- CommandId;
- CommandType;
- CommandVersion;
- RequestedAt;
- EffectiveAt;
- TenantId;
- Actor;
- authorization result;
- Target;
- ExpectedAggregateVersion;
- IdempotencyKey;
- CorrelationId;
- CausationId;
- PolicyReference;
- status;
- Reason Codes;
- previous version;
- resulting version;
- produced EventIds;
- retry count;
- processing timestamps;
- source application;
- sensitive-data classification.

Rejected commands также должны иметь audit, особенно для:

- authorization failure;
- tenant mismatch;
- identity mismatch;
- suspicious replay;
- bulk action denial;
- provider callback failure.

---

# Command Retention

Retention зависит от типа.

Дольше хранятся:

- Completion;
- Award;
- Revocation;
- Reopen;
- Consent;
- Eligibility;
- Progress;
- historical correction;
- migration commands.

Короче могут храниться:

- technical retry;
- periodic discovery commands;
- temporary scheduling operations.

Retention должна быть определена отдельной policy.

---

# AI and Commands

AI может:

- подготовить Command Draft;
- предложить payload;
- найти missing fields;
- классифицировать Reason Codes;
- предложить Review;
- обнаружить stale ExpectedVersion;
- объяснить rejection;
- предложить batch dry run.

AI не может:

- подделывать Actor;
- выбирать чужого Student;
- создавать consent;
- подтверждать Achievement;
- менять Progress;
- Expire Homework;
- Complete Goal;
- одобрять Concert Participation;
- обходить Policy;
- повышать Priority;
- выполнять bulk mutation без Human authorization.

AI proposal должна содержать:

```text
ProposalId
ProposedCommandType
ProposedPayloadReference
SourceReferences
Model
ModelVersion
Confidence
GeneratedAt
ValidationStatus
ApprovedBy
```

---

# Failure Handling

## Invalid payload

- Result: Rejected
- Reason Code: COMMAND_PAYLOAD_INVALID

---

## Unknown command version

- Result: Rejected
- Reason Code: COMMAND_VERSION_UNSUPPORTED

---

## Actor not authenticated

- Result: Rejected
- Reason Code: COMMAND_ACTOR_REQUIRED

---

## Actor not authorized

- Result: Rejected
- Reason Code: COMMAND_NOT_AUTHORIZED

Security Audit обязателен для чувствительных операций.

---

## Tenant mismatch

- Result: Rejected
- Reason Code: COMMAND_TENANT_MISMATCH

---

## Target not found

- Result: Rejected
- Reason Code: COMMAND_TARGET_NOT_FOUND

Ответ end user не должен раскрывать существование недоступного чужого Aggregate.

---

## Expected version mismatch

- Result: Conflict
- Reason Code: COMMAND_VERSION_CONFLICT

---

## Duplicate CommandId

- Result: Already Processed

Возвращается сохраненный результат.

---

## Duplicate IdempotencyKey

- Result: Already Processed (или Conflict, если payload отличается)

---

## Policy decision missing

- Result: Rejected
- Reason Code: COMMAND_POLICY_DECISION_REQUIRED

---

## Human review required

- Result: Pending Human Review
- Reason Code: COMMAND_HUMAN_REVIEW_REQUIRED

---

## Temporary technical failure

- Result: Failed
- Retryable: true

CommandId сохраняется.

---

## Permanent technical failure

- Result: Failed
- Retryable: false

Создается operational incident при необходимости.

---

# Example Command

```json
{
  "commandId": "01K1CMDABC0123456789XYZ123",
  "commandType": "SubmitHomework",
  "commandVersion": 1,
  "requestedAt": "2026-07-27T13:42:15Z",
  "effectiveAt": "2026-07-27T13:42:15Z",
  "tenantId": "belcanto_astana",
  "target": {
    "aggregateType": "HomeworkAssignment",
    "aggregateId": "homework_assignment_123"
  },
  "actor": {
    "actorType": "Student",
    "actorId": "student_456",
    "role": "Student"
  },
  "authorizationContext": {
    "authenticatedActorId": "student_456",
    "activeRole": "Student",
    "scope": "OwnHomework"
  },
  "expectedAggregateVersion": 7,
  "idempotencyKey": "student_456:homework_assignment_123:client_op_987",
  "correlationId": "correlation_789",
  "causationId": "client_request_321",
  "traceId": "trace_654",
  "policyReference": null,
  "payload": {
    "homeworkAssignmentId": "homework_assignment_123",
    "submissionId": "submission_987",
    "submissionVersion": 1,
    "studentId": "student_456",
    "submittedAt": "2026-07-27T13:42:15Z",
    "submissionMethod": "InApp",
    "attachmentReferences": [
      "attachment_111"
    ],
    "clientOperationId": "client_op_987"
  },
  "metadata": {
    "sourceApplication": "student-mobile",
    "privacyClassification": "Confidential"
  }
}
```

---

# Example Successful Result

```json
{
  "commandId": "01K1CMDABC0123456789XYZ123",
  "status": "Completed",
  "targetAggregateType": "HomeworkAssignment",
  "targetAggregateId": "homework_assignment_123",
  "previousAggregateVersion": 7,
  "currentAggregateVersion": 8,
  "producedEventIds": [
    "01K1EVTABC0123456789XYZ999"
  ],
  "decisionReference": null,
  "reasonCodes": [],
  "validationErrors": [],
  "retryable": false,
  "processedAt": "2026-07-27T13:42:16Z",
  "correlationId": "correlation_789"
}
```

---

# Example Conflict Result

```json
{
  "commandId": "01K1CMDABC0123456789XYZ123",
  "status": "Conflict",
  "targetAggregateType": "HomeworkAssignment",
  "targetAggregateId": "homework_assignment_123",
  "previousAggregateVersion": 8,
  "currentAggregateVersion": 8,
  "producedEventIds": [],
  "reasonCodes": [
    "COMMAND_VERSION_CONFLICT"
  ],
  "validationErrors": [],
  "retryable": false,
  "processedAt": "2026-07-27T13:42:16Z",
  "correlationId": "correlation_789"
}
```

---

# Test Requirements

## Envelope Tests

- CommandId обязателен;
- CommandType обязателен;
- CommandVersion обязателен;
- TenantId обязателен;
- Actor обязателен;
- CorrelationId сохраняется;
- Target валиден;
- ExpectedVersion проверяется;
- IdempotencyKey проверяется, когда требуется;
- sensitive Metadata отклоняется.

---

## Naming Tests

- команда именуется в повелительной форме;
- событие не используется как команда;
- generic Update не скрывает critical lifecycle action;
- Command Type соответствует владельцу Aggregate.

---

## Authorization Tests

- Student изменяет только собственные данные;
- Teacher действует только в разрешенном scope;
- Administrator не принимает педагогическое решение без полномочий;
- Guardian scope проверяется;
- Policy имеет ограниченный command scope;
- Scheduler не выполняет pedagogical mutation напрямую;
- Integration не получает лишние полномочия;
- cross-tenant command отклоняется.

---

## Idempotency Tests

- duplicate CommandId не повторяет effect;
- duplicate IdempotencyKey не создает duplicate;
- одинаковый key с другим payload создает Conflict;
- retry сохраняет CommandId;
- повторный provider callback безопасен;
- repeated batch item обрабатывается один раз.

---

## Concurrency Tests

- правильный ExpectedVersion принимается;
- stale version отклоняется;
- concurrent Submission и Expiration не перезаписывают друг друга;
- concurrent Goal update вызывает Conflict;
- mergeable command обрабатывается только по явному правилу;
- stale Policy decision не применяется.

---

## Domain Tests

- CompleteLesson применяет Lesson Completion Policy;
- SubmitHomework учитывает current lifecycle;
- CompleteGoal требует completion decision;
- AwardAchievement требует award decision;
- ChangeSongReadiness требует evaluation;
- Concert approval не подменяет eligibility;
- Notification delivery не меняет educational state.

---

## Event Production Tests

- успешная mutation создает Event;
- rejected command не создает success Event;
- state и Outbox сохраняются атомарно;
- Event сохраняет CommandReference;
- CorrelationId переносится;
- AggregateVersion увеличивается корректно;
- No Change не создает mutation event.

---

## Failure Tests

- validation failure возвращает Rejected;
- technical failure возвращает Failed;
- temporary failure отмечен Retryable;
- authorization failure аудитируется;
- tenant mismatch не раскрывает чужие данные;
- missing Policy decision отклоняется;
- unsupported version не обрабатывается частично.

---

## Batch Tests

- dry run не меняет состояние;
- per-item authorization применяется;
- per-item ExpectedVersion применяется;
- partial result документирован;
- cancellation останавливает новые items;
- completed items не откатываются без compensating command;
- sensitive batch требует Review.

---

## Privacy Tests

- Command payload минимален;
- full private note не передается;
- attachment используется по reference;
- command log ограничен;
- Student не видит internal Reason Codes;
- cross-student references отклоняются;
- AI proposal не содержит лишние sensitive данные.

---

## AI Tests

- AI может создать Draft;
- AI не является authoritative Actor;
- AI не может Award Achievement;
- AI не может Complete Goal;
- AI не может Expire Homework;
- AI не может создавать consent;
- AI proposal требует validation;
- approved Actor сохраняется отдельно от AI.

---

## Traceability Tests

- Command → Decision доступна;
- Command → Event доступна;
- Event-triggered Command сохраняет CausationId;
- Policy Command содержит DecisionId;
- root Command имеет RootCause metadata;
- retry не ломает trace;
- batch сохраняет item-level trace.

---

# Non-Goals

Domain Command Catalog не определяет:

- HTTP endpoints;
- GraphQL mutations;
- transport protocol;
- конкретный message broker;
- programming language;
- application service layout;
- UI forms;
- database schema;
- exact retry intervals;
- authentication provider;
- RBAC implementation;
- workflow engine;
- full scheduling algorithm;
- API error format;
- external public API;
- CRM commands до enrollment;
- payment commands;
- payroll;
- marketing campaign commands;
- legal retention periods.

---

# Open Questions

Необходимо определить:

- какие команды синхронные;
- какие асинхронные;
- какие команды требуют ExpectedVersion;
- какой identifier format использовать;
- как хранить Command Result;
- сколько хранить idempotency records;
- нужен ли Command Bus;
- нужен ли workflow engine;
- где проходит граница Application Service и Domain Service;
- какие Policy выполняются внутри transaction;
- какие Policy асинхронны;
- нужен ли Saga или Process Manager;
- как выполнять multi-aggregate workflows;
- какие команды разрешены Policy Actor;
- какие команды разрешены Scheduler;
- как ограничить Integration Actor;
- как реализовать delegation;
- нужна ли impersonation;
- кто может Reopen Homework;
- кто может Reopen Goal;
- кто может Revoke Achievement;
- кто может менять Concert Eligibility;
- кто может назначать Performance Slot;
- какие команды доступны Guardian;
- какие команды доступны Student self-service;
- какие команды требуют Human Review;
- какие команды разрешают delayed EffectiveAt;
- как обрабатывать offline mobile commands;
- как валидировать client RequestedAt;
- можно ли предварительно генерировать AggregateId;
- нужен ли per-device sequence;
- как обрабатывать duplicate mobile submission;
- как обрабатывать stale offline update;
- какие команды mergeable;
- как оформлять batch commands;
- нужен ли dry-run API;
- как отменять queued command;
- какие команды допускают cancellation;
- как хранить sensitive payload;
- нужно ли шифровать command store;
- нужно ли хранить rejected payload;
- как анонимизировать command history;
- как ограничить Staff access;
- нужна ли отдельная Authorization Policy;
- нужна ли отдельная Idempotency Policy;
- нужна ли отдельная Concurrency Policy;
- нужна ли отдельная Human Review specification;
- нужна ли отдельная Batch Operation Policy;
- как сопоставлять API mutation с Command;
- может ли один API request создавать несколько Commands;
- как возвращать Pending Human Review;
- как уведомлять о завершении async command;
- как обрабатывать poison command;
- нужен ли quarantine;
- как мигрировать CommandVersion;
- нужен ли dual handling нескольких версий;
- как проверять backward compatibility;
- какие команды публикуются как external API;
- можно ли внешней системе вызывать domain command напрямую;
- какие Integration Commands нужны CRM;
- как провести enrollment boundary;
- как предотвратить проникновение CRM semantics в learning domain.

---

# History

| Version | Description |
| --- | --- |
| 1.0.0 | Создан единый каталог Domain Commands, envelope, authorization, idempotency, concurrency, command outcomes и канонические команды основных доменов Belcanto. |
