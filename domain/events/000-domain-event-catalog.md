---
Status: Draft
Version: 2.0.0
Last Updated: 2026-08-01

Document Id: DOMAIN_EVENT_CATALOG

Document Type:
  - Domain Contract
  - Event Catalog
  - Integration Boundary
  - Traceability Specification

Owners:
  - Product Owner
  - Education Lead
  - Domain Architecture Lead
  - Technical Lead

Applies To:
  - Domain Events
  - Policy Events
  - Aggregate Events
  - Cross-Domain Reactions
  - Event Consumers
  - Audit
  - Outbox
  - Event Replay

Related Directories:
  - ../policies/
  - ../commands/
  - ../aggregates/
  - ../../product/
  - ../../school/

Related Documents:
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
  - ../commands/000-domain-command-catalog.md
---

# Domain Event Catalog

> **T7 · DRAFT.** Утверждённая редакция 1.0.0 возвращена в Draft для M-0003
> по PD-0030 и PD-0032 с явным разрешением Product Owner / Education Lead.
>
> Domain Event Catalog описывает набор событий Belcanto Product.
>
> Событие фиксирует уже произошедший доменный факт.
>
> Оно не является запросом, намерением, рекомендацией, командой или техническим сообщением транспорта.

---

# Purpose

Доменные политики Belcanto используют большое количество событий:

- LessonCompleted;
- ProgressUpdated;
- GoalCompleted;
- AchievementAwarded;
- SongReadinessChanged;
- ConcertParticipationMarkedEligible;
- HomeworkSubmitted;
- HomeworkExpired;
- NotificationDelivered;
- PeriodicReviewRequested.

Если каждое событие определяется только внутри отдельной Policy, возникают риски:

- одинаковые события получают разные payload;
- одно имя используется для разных фактов;
- события содержат команды;
- технические данные смешиваются с доменными;
- отсутствует единое versioning;
- consumers зависят от случайных полей;
- невозможно безопасно применять replay;
- невозможно построить надежный Outbox;
- события раскрывают private data;
- сложно определить владельца события;
- Policy ссылаются на несуществующие контракты;
- невозможно проверить причинно-следственную цепочку.

Этот документ создает единый каталог и общие правила для всех Domain Events.

---

# Core Principle

Domain Event описывает факт, который уже произошел и не может быть отменен задним числом.

```text
Command
   |
   v
Aggregate Decision
   |
   v
State Transition
   |
   v
Domain Event
```

Корректно:

```text
HomeworkSubmitted
GoalCompleted
AchievementAwarded
LessonRescheduled
```

Некорректно как Domain Event:

```text
SubmitHomework
CompleteGoal
AwardAchievement
RescheduleLesson
```

Это команды.

---

## Event Semantics

Domain Event отвечает:

> Что уже произошло в доменной модели?

Command отвечает:

> Что система или Actor просит сделать?

Policy отвечает:

> Какое решение необходимо принять при определенных условиях?

Notification отвечает:

> Как донести разрешенный факт или запрос до Recipient?

---

## Event Naming

Все события именуются в прошедшем времени.

Формат:

`<Entity><FactInPastTense>`

Примеры:

```text
LessonCompleted
HomeworkAssigned
GoalCompleted
AchievementAwarded
ConcertParticipationWithdrawn
NotificationDelivered
```

Не использовать:

```text
LessonComplete
HomeworkAssign
GoalCompletion
AchievementAward
SendNotification
```

---

# Event Envelope

Каждое событие должно использовать единый envelope.

```text
DomainEventEnvelope
├── EventId
├── EventType
├── EventVersion
├── OccurredAt
├── RecordedAt
├── AggregateType
├── AggregateId
├── AggregateVersion
├── TenantId
├── Actor
├── CorrelationId
├── CausationId
├── TraceId
├── PolicyReference
├── CommandReference
├── Payload
└── Metadata
```

---

# EventId

Глобально уникальный идентификатор события.

Требования:

- immutable;
- уникален во всей системе;
- пригоден для deduplication;
- не переиспользуется;
- не зависит от transport offset.

---

# EventType

Каноническое имя события.

Пример:

`HomeworkSubmitted`

---

# EventVersion

Версия контракта конкретного Event Type.

Пример:

`1`

EventVersion не равен AggregateVersion.

---

# OccurredAt

Время, когда доменный факт произошел.

Пример:

- Student отправил Homework в 18:42;
- Teacher завершил Lesson в 20:00;
- Goal была подтверждена в 15:10.

---

# RecordedAt

Время фиксации события системой.

RecordedAt может отличаться от OccurredAt, например при:

- offline operation;
- delayed synchronization;
- import;
- recovery;
- external integration.

---

# AggregateType

Тип Aggregate, который является источником события.

Примеры:

- Lesson;
- HomeworkAssignment;
- Goal;
- AchievementAward;
- SongReadinessAssessment;
- ConcertParticipation;
- NotificationIntent.

---

# AggregateId

Идентификатор Aggregate source.

---

# AggregateVersion

Версия Aggregate после применения изменения, породившего событие.

Если Aggregate использует optimistic concurrency:

```text
Previous Version: 7
Event Applied
New Version: 8
```

Event содержит:

`AggregateVersion: 8`

---

# TenantId

Идентификатор School или другой tenant boundary.

Даже если MVP начинается с одной школы, контракт должен сохранять явную tenant boundary.

---

# Actor

Структура субъекта, вызвавшего изменение.

```text
Actor
├── ActorType
├── ActorId
├── Role
├── DelegatedBy
└── AuthenticationContextReference
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
- Integration;
- Migration.

---

# CorrelationId

Объединяет события и команды одного бизнес-процесса.

Пример:

```text
LessonCompleted
    |
    +--> ProgressUpdateRequested
    |
    +--> HomeworkAssigned
    |
    +--> GoalReviewRequested
```

Все могут иметь один CorrelationId.

---

# CausationId

Ссылается на непосредственную причину.

Причиной может быть:

- CommandId;
- EventId;
- ReviewId;
- external message id.

---

# TraceId

Используется для технического distributed tracing.

TraceId не заменяет CorrelationId.

---

# PolicyReference

Если событие является результатом Policy decision:

```text
PolicyReference
├── PolicyId
├── PolicyVersion
├── DecisionId
└── ReasonCodes
```

---

# CommandReference

Если событие возникло после Command:

```text
CommandReference
├── CommandId
├── CommandType
└── CommandVersion
```

---

# Metadata

Metadata не должна превращаться в бесконтрольный склад произвольных данных.

Допустимые данные:

- schema identifier;
- source service;
- migration marker;
- replay marker;
- import batch reference;
- technical partition key;
- data classification.

Недопустимо хранить в Metadata:

- secret;
- access token;
- полный private note;
- произвольный UI state;
- неструктурированный sensitive payload.

---

# Event Categories

## Aggregate State Events

Фиксируют изменение состояния Aggregate.

Примеры:

```text
HomeworkSubmitted
GoalCompleted
LessonCancelled
ConcertParticipationWithdrawn
```

---

## Decision Events

Фиксируют итог Policy или Review.

Примеры:

```text
SongReadinessEvaluated
ConcertEligibilityEvaluated
HomeworkExpirationEvaluated
GoalCompletionEvaluated
```

---

## Request Events

События вида ...Requested допустимы только если сам факт запроса является доменным фактом.

Примеры:

```text
GoalReviewRequested
HomeworkClarificationRequested
ConcertConsentRequested
```

Они не заменяют Command.

---

## Lifecycle Events

Примеры:

```text
HomeworkExpired
NotificationArchived
GoalCancelled
LearningPauseStarted
```

---

## Evidence Events

Фиксируют появление или изменение Evidence.

Примеры:

```text
TeacherAssessmentRecorded
HomeworkEvidenceAccepted
SongReadinessEvidenceAdded
ProgressEvidenceInvalidated
```

---

## Integration Events

Domain Event может быть преобразован в Integration Event.

Но эти сущности не идентичны.

```text
Domain Event
    |
    v
Integration Projection
    |
    v
External Contract
```

Domain Event может содержать внутренние identifiers, которые нельзя раскрывать наружу.

---

# Domain Event and Integration Event

## Domain Event

Используется внутри bounded context.

Может содержать:

- Aggregate references;
- Policy references;
- internal reason codes;
- domain-specific versioning.

---

## Integration Event

Используется между bounded contexts или внешними системами.

Должно иметь:

- отдельный контракт;
- минимальный payload;
- privacy review;
- compatibility guarantees;
- consumer ownership.

Запрещено автоматически публиковать все Domain Events наружу.

---

# Event Ownership

Каждое событие имеет одного владельца.

Владелец определяет:

- semantic meaning;
- payload;
- versioning;
- compatibility;
- allowed producers;
- deprecation;
- privacy classification.

Consumer не может самостоятельно расширять смысл события.

---

# Producer Rules

Событие может быть создано только:

- Aggregate;
- Domain Service, если факт не принадлежит одному Aggregate;
- Policy decision process;
- approved migration process.

UI не является producer Domain Event.

Transport adapter не является владельцем Domain Event.

Database trigger не должен создавать бизнес-событие без явно утвержденной архитектуры.

---

# Consumer Rules

Consumer должен:

- обрабатывать событие идемпотентно;
- проверять EventVersion;
- не предполагать наличие необязательных полей;
- не изменять source Aggregate напрямую;
- отправлять Command владельцу другого Aggregate;
- сохранять CausationId;
- уважать tenant boundary;
- не использовать событие как authorization proof.

---

# Event Immutability

После публикации событие не изменяется.

При ошибке создается новое событие.

Пример:

```text
LessonScheduled
        |
        v
LessonScheduleCorrected
```

Нельзя переписать старое LessonScheduled.

---

# Event Ordering

Гарантированный глобальный порядок всех событий не требуется.

Нужен порядок внутри одного Aggregate stream:

`AggregateId + AggregateVersion`

Consumer должен быть готов к:

- duplicate;
- delayed delivery;
- retry;
- out-of-order delivery между разными Aggregate;
- temporary gaps.

---

# Event Delivery Semantics

Предпочтительная модель:

```text
At-Least-Once Delivery
+
Idempotent Consumers
```

Система не должна предполагать exactly-once transport.

---

# Outbox Requirement

Domain Event и изменение Aggregate должны фиксироваться атомарно.

```text
Database Transaction
├── Aggregate State Change
└── Outbox Event Record
```

После commit отдельный publisher передает событие в broker или event dispatcher.

---

# Inbox Requirement

Consumer, выполняющий значимые side effects, должен использовать Inbox или эквивалентный механизм deduplication.

```text
Inbox Key
├── ConsumerId
└── EventId
```

---

# Replay

Replay допустим для:

- rebuilding projection;
- recovery;
- migration;
- audit verification;
- analytics reconstruction.

Replay не должен автоматически повторять внешние side effects:

- Notification delivery;
- email;
- SMS;
- payment;
- external API call;
- human escalation.

Consumer должен различать:

- Live Processing;
- Replay Processing.

---

# Event Privacy Classification

Каждое событие получает уровень:

- Public Internal;
- Internal;
- Confidential;
- Sensitive;
- Highly Restricted.

## Public Internal

Допустимо широкое внутреннее распространение.

Пример:

`AchievementDefinitionPublished`

---

## Internal

Обычные доменные данные без sensitive content.

---

## Confidential

Данные конкретного Student или Teacher.

---

## Sensitive

Assessment, private learning context, Guardian scope или иные чувствительные сведения.

---

## Highly Restricted

Редкие события с максимально ограниченным доступом.

---

# Payload Minimization

Событие должно содержать данные, необходимые для понимания факта.

Не следует без необходимости включать:

- полный Aggregate snapshot;
- весь профиль Student;
- contact details;
- полный текст Homework;
- private Teacher notes;
- rendered Notification body;
- attachment content;
- медицинские сведения.

Предпочтительно использовать references.

---

# Canonical Event Groups

> **B.0 transport boundary.** Текущие минимальные payload в `apps/api` outbox
> являются неканонической транспортной проекцией. Они не заявляют соответствие
> `DomainEventEnvelope` или `Required Payload` настоящего Draft до Technical P7;
> до этих ворот ни проекция, ни новые B.0-контракты не считаются каноническими.

## B.0 Delegated Access Events

### StudentOnboardingDelegationGranted

Фиксирует выдачу Owner действующего `DelegatedSuperadminAccess` для
`Administrator`. Событие не создаёт новую роль.

Required Payload:

```text
DelegationId
OwnerAccountId
AdministratorAccountId
PermissionSetReference
GrantedAt
EffectivePeriod
```

`PermissionSetReference = StudentOnboardingManager.v1` является производным
техническим отображением текущего среза, а не человеческим утверждением класса B.

### StudentOnboardingDelegationRevoked

Required Payload:

```text
DelegationId
OwnerAccountId
AdministratorAccountId
RevokedAt
ReasonReference
```

---

## Student Events

### StudentCreated

Фиксирует создание учебной идентичности `Student` школой. Событие не означает
активацию `Account`.

Owner: Student

Required Payload:

```text
StudentId
PersonId
SchoolMembershipId
SchoolId
EnrollmentReference
CreatedAt
```

Не содержит:

- полного CRM profile;
- платежных данных;
- маркетинговой истории.

---

### StudentProfileUpdated

Фиксирует значимое изменение product profile.

Не должен создаваться для каждого технического обновления поля.

---

### StudentTimezoneChanged

Required Payload:

```text
StudentId
PreviousTimezone
NewTimezone
EffectiveAt
```

Consumers:

- Homework Reminder Policy;
- Notification Policy;
- Lesson presentation;
- Periodic Review.

---

### StudentLearningPauseStarted

Required Payload:

```text
StudentId
LearningPauseId
StartedAt
ExpectedEndAt
Scope
```

---

### StudentLearningPauseEnded

Required Payload:

```text
StudentId
LearningPauseId
EndedAt
EndReason
```

---

## Account and Invitation Events

### FirstBelcantoMinutePublished

Фиксирует наличие первого учебного ориентира, требуемого до выпуска Invitation.
Содержательная форма ориентира остаётся в педагогическом владельце и не
определяется настоящим каталогом.

Required Payload:

```text
StudentId
OrientationReference
PublishedBy
PublishedAt
```

### StudentActivationInvitationIssued

Required Payload:

```text
InvitationId
StudentId
AccountId
IssuedByAccountId
AuthorizationReference
IssuedAt
ExpiresAt
```

`AuthorizationReference` указывает только на полномочие Owner: delegation grant
не авторизует управление Invitation. Секрет Invitation и пароль в событие не
включаются.

### StudentActivationInvitationReissued

Required Payload:

```text
PreviousInvitationId
InvitationId
StudentId
AccountId
ReissuedByAccountId
AuthorizationReference
ReissuedAt
ExpiresAt
```

Факт означает, что прежнее неиспользованное Invitation стало `Superseded`.

### StudentActivationInvitationRevoked

Required Payload:

```text
InvitationId
StudentId
RevokedByAccountId
AuthorizationReference
RevokedAt
```

### AccountActivated

Required Payload:

```text
AccountId
PersonId
StudentId
InvitationId
ActivatedAt
```

Событие подтверждает атомарное потребление Invitation и переход Account в
`Active`. Оно не создаёт и не изменяет учебный статус Student и не содержит
пароль или credential hash.

---

## Teacher Events

### TeacherAssignedToStudent

Required Payload:

```text
StudentId
TeacherId
AssignmentType
EffectiveFrom
```

---

### TeacherReassigned

Required Payload:

```text
StudentId
PreviousTeacherId
NewTeacherId
EffectiveAt
ReasonCategory
```

Не содержит private HR reason.

---

## Lesson Events

### LessonScheduled

Required Payload:

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

---

### LessonRescheduled

Required Payload:

```text
LessonId
PreviousScheduledStart
PreviousScheduledEnd
NewScheduledStart
NewScheduledEnd
Timezone
ChangedBy
ReasonCategory
```

Consumers:

- Homework Expiration Policy;
- Homework Reminder Policy;
- Notification Policy;
- Concert preparation projections.

---

### LessonCancelled

Required Payload:

```text
LessonId
CancelledAt
CancelledBy
CancellationCategory
ReplacementLessonId
```

---

### LessonStarted

Required Payload:

```text
LessonId
StartedAt
StartedBy
```

---

### LessonCompleted

Required Payload:

```text
LessonId
StudentIds
TeacherId
CompletedAt
CompletionMethod
LessonVersion
```

Не должен содержать полный lesson summary.

Consumers:

- Lesson Completion Policy;
- Progress Update Policy;
- Homework assignment flow;
- Goal Review;
- Periodic Review scheduling.

---

### LessonCompletionRejected

Фиксирует, что попытка Completion была отклонена.

Используется только если сам факт отклонения имеет доменную ценность.

Required Payload:

```text
LessonId
AttemptId
RejectedAt
ReasonCodes
```

---

### LessonSummaryRecorded

Required Payload:

```text
LessonId
SummaryId
RecordedBy
RecordedAt
Visibility
```

Полный текст хранится отдельно.

---

## Homework Events

### HomeworkAssigned

Required Payload:

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
RelatedLessonId
RelatedGoalIds
RelatedSongVersionIds
RelatedConcertId
```

---

### HomeworkUpdated

Required Payload:

```text
HomeworkAssignmentId
PreviousHomeworkVersion
NewHomeworkVersion
ChangedFields
UpdatedBy
UpdatedAt
```

ChangedFields содержит имена измененных доменных полей, а не полный diff sensitive content.

---

### HomeworkDueDateChanged

Required Payload:

```text
HomeworkAssignmentId
PreviousDueDate
NewDueDate
Timezone
ChangeReason
ChangedBy
ChangedAt
```

Consumers:

- Homework Reminder Policy;
- Homework Expiration Policy;
- Notification Policy.

---

### HomeworkStarted

Фиксируется только после явного доменного действия.

Открытие карточки Homework не является HomeworkStarted.

Required Payload:

```text
HomeworkAssignmentId
StudentId
StartedAt
StartMethod
```

---

### HomeworkSubmitted

Required Payload:

```text
HomeworkAssignmentId
HomeworkVersion
SubmissionId
SubmissionVersion
StudentId
SubmittedAt
SubmissionMethod
SubmittedAfterDueDate
```

Не содержит attachment body.

Consumers:

- Homework Reminder Policy;
- Homework Expiration Policy;
- Progress Update Policy;
- Teacher Review flow.

---

### HomeworkSubmissionWithdrawn

Используется, если Student имеет право отозвать Submission.

---

### HomeworkReviewStarted

Required Payload:

```text
HomeworkAssignmentId
SubmissionId
TeacherId
StartedAt
```

---

### HomeworkReviewed

Required Payload:

```text
HomeworkAssignmentId
SubmissionId
ReviewId
TeacherId
ReviewOutcome
ReviewedAt
```

Допустимые ReviewOutcome:

```text
Accepted
AcceptedWithFeedback
CorrectionRequested
NotAssessable
Cancelled
```

---

### HomeworkCorrectionRequested

Required Payload:

```text
HomeworkAssignmentId
SubmissionId
ReviewId
CorrectionRequestId
RequestedBy
RequestedAt
NewDueDate
```

Полный feedback хранится отдельно.

---

### HomeworkClarificationRequested

Required Payload:

```text
HomeworkAssignmentId
StudentId
ClarificationRequestId
RequestedAt
ClarificationCategory
```

---

### HomeworkBlockerReported

Required Payload:

```text
HomeworkAssignmentId
StudentId
BlockerId
BlockerCategory
ReportedAt
```

Не требует сохранения sensitive explanation в event payload.

---

### HomeworkBlockerResolved

Required Payload:

```text
HomeworkAssignmentId
BlockerId
ResolvedBy
ResolvedAt
ResolutionCategory
```

---

### HomeworkMarkedOverdue

Required Payload:

```text
HomeworkAssignmentId
HomeworkVersion
StudentId
DueDate
GracePeriodEnd
DeadlineType
MarkedOverdueAt
ReasonCodes
```

---

### HomeworkGracePeriodStarted

Required Payload:

```text
HomeworkAssignmentId
GracePeriodStart
GracePeriodEnd
StartedAt
```

---

### HomeworkDueDateExtended

Required Payload:

```text
HomeworkAssignmentId
PreviousDueDate
NewDueDate
Timezone
ExtensionReason
RequestedBy
ApprovedBy
EffectiveAt
```

---

### HomeworkExpired

Required Payload:

```text
HomeworkAssignmentId
HomeworkVersion
StudentId
PreviousStatus
ExpiredAt
ExpirationCategory
DecisionId
ReasonCodes
```

---

### HomeworkReopened

Required Payload:

```text
HomeworkAssignmentId
PreviousExpiredAt
NewHomeworkVersion
NewDueDate
ReopenReason
ReopenedBy
ReopenedAt
```

---

### HomeworkCancelled

Required Payload:

```text
HomeworkAssignmentId
CancelledBy
CancelledAt
CancellationCategory
```

---

### HomeworkReplaced

Required Payload:

```text
HomeworkAssignmentId
ReplacementHomeworkAssignmentId
ReplacedBy
ReplacedAt
ReplacementReason
```

---

### HomeworkCompleted

Фиксирует итоговое завершение Homework.

Required Payload:

```text
HomeworkAssignmentId
SubmissionId
ReviewId
CompletedAt
CompletionMethod
```

---

### HomeworkArchived

Required Payload:

```text
HomeworkAssignmentId
ArchivedAt
ArchiveReason
```

---

## Progress Events

### ProgressEvidenceRecorded

Required Payload:

```text
ProgressEvidenceId
StudentId
EvidenceType
SourceEntityType
SourceEntityId
RecordedBy
RecordedAt
ValidityScope
```

---

### ProgressEvidenceInvalidated

Required Payload:

```text
ProgressEvidenceId
InvalidatedBy
InvalidatedAt
InvalidationReason
SupersedingEvidenceId
```

---

### ProgressUpdateEvaluated

Required Payload:

```text
ProgressEvaluationId
StudentId
EvaluationScope
PreviousProgressVersion
ProposedProgressVersion
Outcome
EvaluatedAt
PolicyVersion
```

---

### ProgressUpdated

Required Payload:

```text
ProgressId
StudentId
PreviousProgressVersion
NewProgressVersion
ChangedDimensions
EvidenceReferences
UpdatedAt
```

Не содержит полный Progress snapshot.

---

### ProgressUpdateRejected

Required Payload:

```text
ProgressEvaluationId
StudentId
RejectedAt
ReasonCodes
```

---

### ProgressReviewRequested

Required Payload:

```text
ProgressReviewId
StudentId
Scope
RequestedBy
RequestedAt
ReasonCodes
```

---

## Goal Events

### GoalCreated

Required Payload:

```text
GoalId
StudentId
GoalType
CriterionReference
CreatedBy
CreatedAt
TargetReviewAt
```

---

### GoalActivated

Required Payload:

```text
GoalId
ActivatedBy
ActivatedAt
```

---

### GoalUpdated

Required Payload:

```text
GoalId
PreviousGoalVersion
NewGoalVersion
ChangedFields
UpdatedBy
UpdatedAt
```

---

### GoalProgressUpdated

Required Payload:

```text
GoalId
PreviousProgressState
NewProgressState
EvidenceReferences
UpdatedAt
```

---

### GoalReviewRequested

Required Payload:

```text
GoalId
GoalReviewId
RequestedBy
RequestedAt
ReasonCodes
```

---

### GoalCompletionEvaluated

Required Payload:

```text
GoalId
GoalVersion
EvaluationId
Outcome
EvidenceReferences
BlockingConditions
EvaluatedAt
PolicyVersion
```

---

### GoalCompleted

Required Payload:

```text
GoalId
StudentId
CompletedAt
CompletedBy
CompletionDecisionId
EvidenceReferences
```

---

### GoalReopened

Required Payload:

```text
GoalId
PreviousCompletedAt
ReopenedBy
ReopenedAt
ReopenReason
NewGoalVersion
```

---

### GoalCancelled

Required Payload:

```text
GoalId
CancelledBy
CancelledAt
CancellationCategory
```

---

### GoalArchived

Required Payload:

```text
GoalId
ArchivedAt
ArchiveReason
```

---

## Achievement Events

### AchievementDefinitionPublished

Required Payload:

```text
AchievementDefinitionId
DefinitionVersion
PublishedBy
PublishedAt
```

---

### AchievementEligibilityEvaluated

Required Payload:

```text
AchievementDefinitionId
StudentId
EvaluationId
Outcome
EvidenceReferences
EvaluatedAt
PolicyVersion
```

---

### AchievementAwarded

Required Payload:

```text
AchievementAwardId
AchievementDefinitionId
AchievementDefinitionVersion
StudentId
AwardedAt
AwardedBy
AwardDecisionId
EvidenceReferences
```

---

### AchievementAwardRejected

Required Payload:

```text
AchievementDefinitionId
StudentId
EvaluationId
RejectedAt
ReasonCodes
```

---

### AchievementRevoked

Required Payload:

```text
AchievementAwardId
StudentId
RevokedBy
RevokedAt
RevocationReason
ReplacementAwardId
```

Revocation не удаляет AchievementAwarded.

---

### AchievementRestored

Required Payload:

```text
AchievementAwardId
RestoredBy
RestoredAt
RestoreReason
```

---

## Song Events

### SongAddedToStudentRepertoire

Required Payload:

```text
StudentSongId
StudentId
SongId
SongVersionId
AddedBy
AddedAt
Purpose
```

---

### SongVersionCreated

Required Payload:

```text
SongVersionId
SongId
VersionNumber
Key
ArrangementReference
CreatedBy
CreatedAt
```

---

### StudentSongVersionChanged

Required Payload:

```text
StudentSongId
PreviousSongVersionId
NewSongVersionId
ChangedBy
ChangedAt
ChangeReason
```

---

### SongReadinessEvidenceAdded

Required Payload:

```text
SongReadinessEvidenceId
StudentId
SongVersionId
ReadinessArea
EvidenceType
SourceReference
RecordedBy
RecordedAt
```

---

### SongReadinessEvidenceInvalidated

Required Payload:

```text
SongReadinessEvidenceId
InvalidatedBy
InvalidatedAt
InvalidationReason
```

---

### SongReadinessEvaluationRequested

Required Payload:

```text
SongReadinessEvaluationId
StudentId
SongVersionId
PerformanceType
RequestedBy
RequestedAt
ReasonCodes
```

---

### SongReadinessEvaluated

Required Payload:

```text
SongReadinessEvaluationId
StudentId
SongVersionId
PerformanceType
Outcome
AreaResults
EvidenceReferences
BlockingConditions
EvaluatedAt
PolicyVersion
```

---

### SongReadinessChanged

Required Payload:

```text
StudentId
SongVersionId
PerformanceType
PreviousReadinessStatus
NewReadinessStatus
EvaluationId
ChangedAt
```

---

### SongReadinessReviewRequired

Required Payload:

```text
StudentId
SongVersionId
PerformanceType
ReviewId
ReasonCodes
RequestedAt
```

---

## Concert Events

### ConcertCreated

Required Payload:

```text
ConcertId
ConcertVersion
Title
ScheduledStart
ScheduledEnd
Timezone
VenueReference
CreatedBy
CreatedAt
```

---

### ConcertUpdated

Required Payload:

```text
ConcertId
PreviousConcertVersion
NewConcertVersion
ChangedFields
UpdatedBy
UpdatedAt
```

---

### ConcertCancelled

Required Payload:

```text
ConcertId
CancelledBy
CancelledAt
CancellationCategory
```

---

### ConcertCompleted

Required Payload:

```text
ConcertId
CompletedAt
CompletedBy
```

---

### ConcertRequirementsPublished

Required Payload:

```text
ConcertId
ConcertRequirementsVersion
PublishedBy
PublishedAt
```

---

### ConcertParticipationProposed

Required Payload:

```text
ConcertParticipationId
ConcertId
StudentId
PerformanceType
SongVersionIds
ProposedBy
ProposedAt
```

---

### ConcertConsentRequested

Required Payload:

```text
ConcertParticipationId
StudentId
ConsentRequestId
RequestedAt
ExpiresAt
```

---

### ConcertConsentGranted

Required Payload:

```text
ConcertParticipationId
ConsentId
GrantedBy
GrantedAt
ConsentScope
```

---

### ConcertConsentWithdrawn

Required Payload:

```text
ConcertParticipationId
ConsentId
WithdrawnBy
WithdrawnAt
```

---

### ConcertEligibilityEvaluationRequested

Required Payload:

```text
ConcertEligibilityEvaluationId
ConcertParticipationId
ConcertId
StudentId
SongVersionIds
PerformanceType
RequestedBy
RequestedAt
ReasonCodes
```

---

### ConcertEligibilityEvaluated

Required Payload:

```text
ConcertEligibilityEvaluationId
ConcertParticipationId
ConcertRequirementsVersion
Outcome
DimensionResults
EvidenceReferences
Conditions
BlockingConditions
EvaluatedAt
PolicyVersion
```

---

### ConcertParticipationMarkedEligible

Required Payload:

```text
ConcertParticipationId
ConcertId
StudentId
PerformanceType
SongVersionIds
EligibilityDecisionId
EligibilityStatus
Conditions
MarkedEligibleAt
```

---

### ConcertParticipationMarkedConditionallyEligible

Required Payload:

```text
ConcertParticipationId
EligibilityDecisionId
Conditions
ConditionDeadline
MarkedAt
```

---

### ConcertParticipationMarkedNotEligible

Required Payload:

```text
ConcertParticipationId
EligibilityDecisionId
BlockingConditions
MarkedAt
```

---

### ConcertParticipationApproved

Approval отделено от Eligibility.

Required Payload:

```text
ConcertParticipationId
ApprovedBy
ApprovedAt
ApprovalScope
```

---

### ConcertParticipationWithdrawn

Required Payload:

```text
ConcertParticipationId
WithdrawnBy
WithdrawnAt
WithdrawalCategory
```

---

### ConcertPerformanceSlotAssigned

Required Payload:

```text
ConcertParticipationId
PerformanceSlotId
ScheduledStart
StageReference
AssignedBy
AssignedAt
```

---

### ConcertPerformanceSlotChanged

Required Payload:

```text
ConcertParticipationId
PerformanceSlotId
PreviousScheduledStart
NewScheduledStart
ChangedBy
ChangedAt
```

---

### ConcertProgramPublished

Required Payload:

```text
ConcertId
ProgramVersion
PublishedBy
PublishedAt
```

---

### ConcertPerformanceCompleted

Required Payload:

```text
ConcertParticipationId
PerformanceSlotId
CompletedAt
CompletionMethod
```

---

## Reminder Events

### HomeworkReminderPlanCreated

Required Payload:

```text
ReminderPlanId
HomeworkAssignmentId
HomeworkVersion
StudentId
Strategy
MaximumReminderCount
Timezone
CreatedAt
PolicyVersion
```

---

### HomeworkReminderScheduled

Required Payload:

```text
ReminderId
ReminderPlanId
HomeworkAssignmentId
HomeworkVersion
StudentId
ReminderType
ScheduledFor
Timezone
ChannelPreference
TemplateId
```

---

### HomeworkReminderRescheduled

Required Payload:

```text
ReminderId
PreviousScheduledFor
NewScheduledFor
RescheduleReason
RescheduledAt
```

---

### HomeworkReminderDue

Required Payload:

```text
ReminderId
ReminderPlanId
DueAt
```

Этот Event не разрешает отправку автоматически.

---

### HomeworkReminderSuppressed

Required Payload:

```text
ReminderId
HomeworkAssignmentId
StudentId
SuppressionReason
SuppressedAt
```

---

### HomeworkReminderCancelled

Required Payload:

```text
ReminderId
CancelledAt
CancellationReason
```

---

### HomeworkReminderExpired

Required Payload:

```text
ReminderId
ExpiredAt
ExpirationReason
```

---

### HomeworkReminderDelivered

Required Payload:

```text
ReminderId
NotificationDeliveryId
DeliveredAt
Channel
```

---

### HomeworkReminderDeliveryFailed

Required Payload:

```text
ReminderId
NotificationDeliveryId
FailedAt
FailureCategory
```

---

## Notification Events

### NotificationIntentCreated

Required Payload:

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
PrivacyLevel
TemplateReference
ExpiresAt
DeduplicationKey
```

---

### NotificationIntentApproved

Required Payload:

```text
NotificationIntentId
ApprovedBy
ApprovedAt
DecisionId
```

---

### NotificationIntentRejected

Required Payload:

```text
NotificationIntentId
RejectedBy
RejectedAt
ReasonCodes
```

---

### NotificationScheduled

Required Payload:

```text
NotificationDeliveryId
NotificationIntentId
RecipientId
Channel
ScheduledFor
ExpiresAt
```

---

### NotificationRescheduled

Required Payload:

```text
NotificationDeliveryId
PreviousScheduledFor
NewScheduledFor
RescheduleReason
```

---

### NotificationBundled

Required Payload:

```text
NotificationBundleId
RecipientId
IncludedIntentIds
BundleType
ScheduledFor
Channel
```

---

### NotificationSendingRequested

Фиксирует, что разрешенное сообщение передано delivery process.

Required Payload:

```text
NotificationDeliveryId
NotificationIntentId
Channel
IdempotencyKey
RequestedAt
```

---

### NotificationDelivered

Required Payload:

```text
NotificationDeliveryId
NotificationIntentId
RecipientId
Channel
ProviderReference
DeliveredAt
AttemptCount
```

---

### NotificationOpened

Required Payload:

```text
NotificationDeliveryId
RecipientId
OpenedAt
OpenSource
```

---

### NotificationActionCompleted

Required Payload:

```text
NotificationDeliveryId
NotificationIntentId
ActionType
SourceEntityId
CompletedAt
```

---

### NotificationDeliveryFailed

Required Payload:

```text
NotificationDeliveryId
NotificationIntentId
Channel
FailureCategory
FailureCode
FailedAt
AttemptCount
Retryable
```

---

### NotificationRetryScheduled

Required Payload:

```text
NotificationDeliveryId
AttemptNumber
ScheduledFor
RetryUntil
FailureCategory
```

---

### NotificationRetryStopped

Required Payload:

```text
NotificationDeliveryId
StoppedAt
StopReason
FinalAttemptCount
```

---

### NotificationChannelSwitched

Required Payload:

```text
NotificationDeliveryId
PreviousChannel
NewChannel
SwitchReason
SwitchedAt
```

---

### NotificationSuppressed

Required Payload:

```text
NotificationIntentId
RecipientId
Category
SuppressionReason
SuppressedAt
```

---

### NotificationCancelled

Required Payload:

```text
NotificationIntentId
NotificationDeliveryId
CancelledAt
CancellationReason
```

---

### NotificationExpired

Required Payload:

```text
NotificationIntentId
NotificationDeliveryId
ExpiredAt
ExpirationReason
```

---

### NotificationArchived

Required Payload:

```text
NotificationIntentId
ArchivedAt
ArchiveReason
```

---

## Periodic Review Events

### PeriodicReviewCycleStarted

Required Payload:

```text
ReviewCycleId
ReviewCategory
StartedAt
RequestedBy
Scope
```

---

### PeriodicReviewItemDiscovered

Фиксирует, что Aggregate соответствует критериям проверки.

Required Payload:

```text
ReviewCycleId
ReviewItemId
AggregateType
AggregateId
AggregateVersion
DiscoveryReason
DiscoveredAt
```

Не является итоговым доменным решением.

---

### PeriodicReviewRequested

Required Payload:

```text
ReviewId
ReviewCategory
AggregateType
AggregateId
AggregateVersion
RequestedAt
ReasonCodes
```

---

### PeriodicReviewCompleted

Required Payload:

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

### PeriodicReviewFailed

Required Payload:

```text
ReviewId
ReviewCategory
AggregateType
AggregateId
FailureCategory
Retryable
FailedAt
```

---

### PeriodicReviewCycleCompleted

Required Payload:

```text
ReviewCycleId
ReviewCategory
StartedAt
CompletedAt
DiscoveredCount
RequestedReviewCount
SkippedCount
FailedCount
```

---

## Integrity Events

### DomainIntegrityIssueDetected

Required Payload:

```text
IntegrityIssueId
IssueType
AggregateType
AggregateId
DetectedAt
Severity
EvidenceReferences
```

---

### DomainIntegrityReviewRequested

Required Payload:

```text
IntegrityReviewId
IntegrityIssueId
RequestedAt
```

---

### DomainIntegrityIssueResolved

Required Payload:

```text
IntegrityIssueId
ResolvedBy
ResolvedAt
ResolutionCategory
```

---

# Canonical Reason Codes

Reason Codes принадлежат Policy, а не Event transport.

Event может ссылаться на них, но не должен менять их смысл.

Примеры:

```text
HOMEWORK_ALREADY_COMPLETED
HOMEWORK_REPLACED
STUDENT_LEARNING_PAUSE_ACTIVE
GOAL_CRITERIA_SATISFIED
ACHIEVEMENT_EVIDENCE_INSUFFICIENT
SONG_READINESS_EVIDENCE_STALE
CONCERT_REQUIREMENT_NOT_SATISFIED
NOTIFICATION_INTENT_NO_LONGER_CURRENT
PERIODIC_REVIEW_ALREADY_PROCESSED
```

---

# Event Versioning

## Backward-Compatible Change

Обычно допускается в той же EventVersion:

- добавление optional field;
- уточнение documentation;
- новый optional enum value при tolerant consumers;
- расширение Metadata.

---

## Breaking Change

Требует новой EventVersion:

- удаление field;
- изменение meaning;
- изменение type;
- превращение optional field в required;
- изменение time semantics;
- изменение identifier semantics;
- изменение privacy classification;
- изменение aggregate ownership.

---

# Event Type Versioning Example

```text
HomeworkSubmitted v1
HomeworkSubmitted v2
```

Consumer должен явно объявить поддерживаемые версии.

---

# Deprecation

Event нельзя просто удалить после появления consumers.

Процесс:

```text
Active
  |
  v
Deprecated
  |
  v
Dual Publish or Migration
  |
  v
Consumer Migration
  |
  v
Publication Stopped
  |
  v
Historical Contract Retained
```

---

# Event Schema Registry

Для каждого Event Type должен существовать schema contract.

Рекомендуемая структура:

```text
domain/events/
├── 000-domain-event-catalog.md
├── shared/
│   ├── domain-event-envelope.md
│   ├── actor.md
│   └── policy-reference.md
├── lesson/
│   ├── lesson-completed.v1.md
│   ├── lesson-rescheduled.v1.md
│   └── lesson-cancelled.v1.md
├── homework/
├── progress/
├── goals/
├── achievements/
├── songs/
├── concerts/
├── notifications/
└── reviews/
```

На текущем specification stage этот каталог является каноническим индексом.

Отдельные event files создаются при переходе к implementation contracts.

---

# Event Validation Rules

Каждое событие должно проходить validation:

- EventId существует.
- EventType известен.
- EventVersion поддерживается.
- OccurredAt существует.
- AggregateType существует.
- AggregateId существует.
- AggregateVersion допустима.
- TenantId существует.
- CorrelationId существует.
- CausationId существует, кроме root event.
- Payload соответствует schema.
- Privacy classification указана.
- Required fields заполнены.
- Producer авторизован.
- Event не дублирует уже зафиксированную версию Aggregate.

---

# Root Events

Некоторые события могут не иметь CausationId.

Примеры:

- первоначальный import;
- ручная инициатива Actor;
- scheduled review root;
- system startup recovery.

В этом случае Metadata должна содержать:

`RootCauseType`

---

# Event Publication Rules

Event публикуется только после успешного commit Aggregate.

Нельзя:

- публиковать до commit;
- публиковать событие для отклоненной команды;
- повторно создавать EventId при retry;
- использовать broker delivery как доказательство commit;
- silently drop publish failure.

---

# Event Failure Handling

## Aggregate committed, event not yet published

Outbox сохраняет событие и повторяет publication.

---

## Event published twice

Consumer deduplicates по EventId.

---

## Consumer failed

Event остается доступным для retry.

---

## Consumer permanently rejects schema

Создается operational incident.

---

## Invalid Event Produced

Событие помещается в quarantine.

Не следует публиковать его как будто оно корректно.

---

# Event Audit

Для каждого события должны быть доступны:

- producer;
- schema version;
- aggregate reference;
- actor;
- command;
- policy decision;
- correlation chain;
- publication attempts;
- consumer processing status, если технически необходимо;
- privacy classification;
- retention category.

---

# Event Retention

Retention зависит от категории.

## Long-Term Audit

Обычно:

- GoalCompleted;
- AchievementAwarded;
- AchievementRevoked;
- ConcertParticipationApproved;
- HomeworkCompleted;
- ProgressUpdated;
- TeacherAssessmentRecorded.

---

## Operational Retention

Обычно:

- NotificationRetryScheduled;
- PeriodicReviewCycleStarted;
- delivery attempt events.

---

## Privacy-Limited Retention

События с sensitive content должны минимизироваться и храниться согласно отдельной retention policy.

---

# Event Access

Не каждый Staff Actor должен видеть raw event stream.

Рекомендуемые уровни:

- Domain Administrator;
- Technical Operator;
- Privacy Reviewer;
- Audit Reviewer;
- Product Analytics projection.

UI должен использовать projections, а не raw events.

---

# Event Analytics

Analytics должна потреблять специально подготовленные projections.

Нельзя без отдельной модели использовать raw events для:

- рейтинга Student;
- оценки мотивации;
- наказаний;
- оценки Teacher по одному показателю;
- поведенческого профилирования;
- скрытого измерения эмоционального состояния.

---

# AI and Events

AI может:

- классифицировать Event для диагностики;
- искать semantic duplicates;
- предложить schema improvements;
- обнаруживать пропущенные causation links;
- предлагать migration;
- анализировать event chains;
- создавать human-readable explanations.

AI не может:

- создавать исторические Domain Events без подтвержденного факта;
- менять Event payload после публикации;
- придумывать Actor;
- придумывать Evidence;
- изменять OccurredAt;
- скрывать invalid event;
- заменять Policy decision;
- публиковать Event без authorization.

---

# Security

Необходимо защищать:

- forged Event;
- cross-tenant Event;
- EventId collision;
- AggregateVersion tampering;
- Actor spoofing;
- replay attack;
- unauthorized publication;
- schema downgrade;
- sensitive payload leakage;
- malicious Metadata;
- forged provider callback event;
- event stream enumeration;
- deletion of audit events.

---

# Idempotency

Producer idempotency:

`CommandId + AggregateId + ExpectedVersion`

Consumer idempotency:

`ConsumerId + EventId`

Business deduplication может дополнительно использовать:

`DeduplicationKey`

---

# Concurrency

При conflict:

`Expected AggregateVersion != Current AggregateVersion`

Command должен быть отклонен или переоценен.

Нельзя публиковать Event для stale state transition.

---

# Event Choreography

Допустимый процесс:

```text
LessonCompleted
        |
        +--> ProgressUpdateRequested
        |
        +--> GoalReviewRequested
        |
        +--> HomeworkAssigned
```

Каждый consumer:

- принимает Event;
- оценивает собственный контекст;
- отправляет Command;
- не модифицирует чужой Aggregate напрямую.

---

# Event Cycle Prevention

Необходимо предотвращать циклы.

Пример запрещенного цикла:

```text
ProgressUpdated
    ↓
GoalReviewRequested
    ↓
GoalProgressUpdated
    ↓
ProgressUpdateRequested
    ↓
ProgressUpdated
```

Защиты:

- Causation chain inspection;
- policy guards;
- deduplication;
- version checks;
- maximum reaction depth;
- explicit ownership;
- state-change requirement.

Событие не должно публиковаться, если фактического изменения состояния не произошло, кроме явно определенных audit events.

---

# Event Processing Result

Consumer может завершить обработку как:

- Processed;
- Ignored;
- Deferred;
- Retryable Failure;
- Permanent Failure;
- Quarantined.

Ignored должен иметь причину.

---

# Event Documentation Requirements

Для каждого отдельного event contract необходимо определить:

- Event Type;
- Event Version;
- Owner;
- Producer;
- Meaning;
- Trigger;
- Required Payload;
- Optional Payload;
- Invariants;
- Privacy Classification;
- Consumers;
- Idempotency;
- Ordering;
- Compatibility;
- Example;
- Tests.

---

# Example Event

```json
{
  "eventId": "01K1ABCDEF0123456789XYZ123",
  "eventType": "HomeworkSubmitted",
  "eventVersion": 1,
  "occurredAt": "2026-07-27T13:42:15Z",
  "recordedAt": "2026-07-27T13:42:16Z",
  "aggregateType": "HomeworkAssignment",
  "aggregateId": "homework_assignment_123",
  "aggregateVersion": 8,
  "tenantId": "belcanto_astana",
  "actor": {
    "actorType": "Student",
    "actorId": "student_456",
    "role": "Student"
  },
  "correlationId": "correlation_789",
  "causationId": "command_submit_homework_321",
  "traceId": "trace_654",
  "policyReference": null,
  "commandReference": {
    "commandId": "command_submit_homework_321",
    "commandType": "SubmitHomework",
    "commandVersion": 1
  },
  "payload": {
    "homeworkAssignmentId": "homework_assignment_123",
    "homeworkVersion": 4,
    "submissionId": "submission_987",
    "submissionVersion": 1,
    "studentId": "student_456",
    "submittedAt": "2026-07-27T13:42:15Z",
    "submissionMethod": "InApp",
    "submittedAfterDueDate": false
  },
  "metadata": {
    "privacyClassification": "Confidential",
    "sourceService": "learning"
  }
}
```

---

# Test Requirements

## Envelope Tests

- EventId обязателен;
- EventType обязателен;
- EventVersion обязателен;
- Aggregate reference обязателен;
- TenantId обязателен;
- OccurredAt не позже недопустимого будущего времени;
- RecordedAt не раньше невозможного системного времени;
- CorrelationId сохраняется;
- CausationId сохраняется;
- Actor валиден.

---

## Schema Tests

- payload соответствует EventVersion;
- неизвестное required field невозможно;
- optional field может отсутствовать;
- enum validation работает;
- identifier types не смешиваются;
- timestamp имеет timezone;
- private field не попадает в запрещенное событие.

---

## Producer Tests

- только владелец может создать Event;
- UI не публикует Domain Event напрямую;
- rejected Command не создает success Event;
- stale Aggregate Version не создает Event;
- duplicate Command не создает новый бизнес-факт;
- migration producer отмечен явно.

---

## Consumer Tests

- duplicate Event обрабатывается один раз;
- unknown EventVersion не повреждает state;
- out-of-order Event обрабатывается безопасно;
- cross-tenant Event отклоняется;
- consumer сохраняет CausationId;
- consumer не изменяет чужой Aggregate напрямую.

---

## Outbox Tests

- Aggregate и Outbox сохраняются атомарно;
- rollback не оставляет Event;
- publisher retry не меняет EventId;
- duplicate publication допустима;
- unpublished records восстанавливаются после restart;
- Event order внутри Aggregate сохраняется.

---

## Inbox Tests

- duplicate delivery определяется;
- failed consumer может retry;
- completed processing не повторяет side effect;
- replay не отправляет external Notification;
- Inbox isolation учитывает ConsumerId;
- retention не удаляет dedup data слишком рано.

---

## Versioning Tests

- backward-compatible optional field поддерживается;
- breaking change требует новой версии;
- consumer объявляет supported versions;
- deprecated Event остается документированным;
- dual-publish не создает два бизнес-факта;
- migration сохраняет traceability.

---

## Privacy Tests

- event payload минимален;
- Teacher private notes отсутствуют;
- rendered Notification body не публикуется без необходимости;
- Guardian data ограничены;
- cross-student leakage невозможен;
- analytics используют projections;
- raw Event access ограничен.

---

## Causation Tests

- Command → Event связь сохраняется;
- Event → Command → Event цепочка прослеживается;
- root event имеет RootCauseType;
- cycle detection работает;
- новый CorrelationId не создается посередине процесса без причины;
- TraceId не подменяет CorrelationId.

---

## Domain Semantics Tests

- Event именуется в прошедшем времени;
- Event описывает факт;
- Event не содержит просьбу выполнить действие;
- Delivery Event не создает educational truth;
- Reminder Event не становится Progress Evidence;
- Expiration Event не означает negative Assessment;
- Eligibility Event не означает Participation approval.

---

# Non-Goals

Domain Event Catalog не определяет:

- command handlers;
- API endpoints;
- message broker;
- конкретную базу данных;
- topic names;
- partition count;
- serialization library;
- programming language;
- retry intervals;
- deployment topology;
- event sourcing для всех Aggregate;
- UI notification text;
- analytics dashboards;
- external public API;
- legal retention periods;
- event schema implementation format.

---

# Open Questions

Необходимо определить:

- какие Aggregate используют full event sourcing;
- какие используют state storage + Outbox;
- какой формат schema выбрать;
- JSON Schema или Protobuf;
- нужен ли Schema Registry;
- какой identifier format использовать;
- ULID или UUID;
- какой message broker использовать;
- нужны ли отдельные topics по bounded context;
- нужна ли tenant partitioning;
- какой ordering гарантируется;
- какой максимальный размер события;
- где хранить large payload references;
- как хранить attachment references;
- какие события разрешено экспортировать;
- какие Integration Events нужны CRM;
- какие Integration Events нужны mobile client;
- нужен ли public webhook API;
- как подписывать события;
- нужен ли encryption payload;
- какие события относятся к Sensitive;
- сколько хранить Outbox;
- сколько хранить Inbox;
- как выполнять replay;
- как запрещать side effects при replay;
- нужен ли Event Store;
- как строить projections;
- как восстанавливать projection;
- как обрабатывать poison events;
- нужен ли quarantine topic;
- кто получает alert о schema failure;
- как проводить event migration;
- поддерживается ли dual publish;
- как документировать deprecated events;
- кто утверждает breaking changes;
- нужна ли отдельная Event Compatibility Policy;
- нужна ли отдельная Event Privacy Policy;
- нужна ли отдельная Integration Event Catalog;
- как сопоставлять Domain Events с Notifications;
- может ли mobile client получать Domain Events напрямую;
- как ограничить Staff access к raw stream;
- какие события должны быть видны Student в activity history;
- какие события должны быть видны Teacher;
- как отображать corrected historical facts;
- как учитывать offline mobile events;
- можно ли принимать client OccurredAt;
- как валидировать client time;
- как импортировать исторические Lessons;
- как маркировать imported events;
- создаются ли события для legacy data;
- как предотвращать fabricated history;
- как поддерживать GDPR-style deletion без уничтожения обязательного Audit;
- как анонимизировать event history;
- какие projections нужны для Owner analytics;
- как не превратить event stream в surveillance system.

---

# Таблица вывода редакции M-0003/B.0

| Изменённая часть | Класс | Источник | Версия источника |
|------------------|-------|----------|------------------|
| T7 и граница редакции | B | PD-0030, PD-0032; разрешение Product Owner / Education Lead 2026-08-01 | — |
| StudentCreated identity references | A | PD-0030, пп. 1–3 | — |
| Delegation events | A | PD-0030, пп. 7–10 | — |
| Invitation и Account activation events | A | PD-0030, пп. 4–6, 9, 11, 13 | — |
| `StudentOnboardingManager.v1` | A | производное техническое отображение B.0 scope, не решение класса B | — |
| B.0 outbox transport boundary | A | текущая реализация `apps/api` + состояние Draft; соответствие отложено до Technical P7 | — |

---

# History

| Version | Description |
| --- | --- |
| 2.0.0 | T7 · DRAFT · M-0003/B.0: добавлены события делегирования, Invitation и Account activation; обязательный payload StudentCreated расширен identity references, поэтому редакция имеет MAJOR-версию. Текущий минимальный outbox отмечен как неканоническая transport-проекция до Technical P7. Технический package literal отмечен как производный. Требуется P7 Education Lead + Technical Lead. |
| 1.0.0 | Создан единый канонический каталог Domain Events, envelope, ownership, versioning, Outbox, Inbox, replay, privacy и event contracts для основных доменов Belcanto. |
