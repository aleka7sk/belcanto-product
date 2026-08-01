---
Status: Draft
Version: 2.0.0
Last Updated: 2026-08-01

Document Id: AGGREGATE_CATALOG

Document Type:
  - Domain Contract
  - Aggregate Catalog
  - Consistency Boundary Specification
  - Invariant Ownership Specification
  - Lifecycle Specification
  - Transaction Boundary Specification

Owners:
  - Product Owner
  - Education Lead
  - Domain Architecture Lead
  - Technical Lead

Applies To:
  - Aggregate Roots
  - Aggregate Entities
  - Value Objects
  - Domain Invariants
  - Transaction Boundaries
  - Command Routing
  - Event Production
  - Optimistic Concurrency
  - Cross-Aggregate Coordination
  - Domain References

Related Directories:
  - ../commands/
  - ../events/
  - ../policies/
  - ../entities/
  - ../value-objects/
  - ../../product/
  - ../../school/

Related Documents:
  - ../commands/000-domain-command-catalog.md
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

# Aggregate Catalog

> **T7 · DRAFT.** Утверждённая редакция 1.0.0 возвращена в Draft для M-0003
> по PD-0030 и PD-0032 с явным разрешением Product Owner / Education Lead.
>
> Aggregate Catalog описывает границы согласованности Belcanto Product.
>
> Aggregate является группой доменных объектов, которые изменяются как единое согласованное целое.
>
> Каждый Aggregate имеет один Aggregate Root, который является единственной точкой входа для изменения состояния внутри этой границы.

---

# Purpose

Команды, события и политики уже определяют:

- что система может попросить выполнить;
- какие факты возникают после изменения;
- какие решения применяются в сложных ситуациях.

Но без явных Aggregate boundaries остается неясным:

- какой объект владеет конкретным инвариантом;
- что должно изменяться в одной транзакции;
- куда направляется Command;
- кто создает Domain Event;
- какие ссылки допустимы между объектами;
- какие состояния могут быть изменены атомарно;
- где требуется Process Manager;
- можно ли изменять Lesson и Homework одновременно;
- кому принадлежит Progress Evidence;
- является ли Concert Eligibility частью Concert;
- является ли Reminder частью Homework;
- можно ли хранить Notification Delivery внутри Notification Intent;
- какая версия проверяется через ExpectedAggregateVersion.

Этот документ фиксирует предварительный канонический набор Aggregate Roots и их ответственность.

---

# Core Principle

Aggregate является consistency boundary, а не способом сгруппировать похожие таблицы.

```text
Command
   |
   v
Aggregate Root
   |
   +--> validates aggregate invariants
   |
   +--> changes internal state
   |
   +--> emits Domain Events
```

Внешний объект не может изменять внутреннюю Entity напрямую.

Корректно:

```text
SubmitHomework
      |
      v
HomeworkAssignment Aggregate
      |
      v
HomeworkSubmitted
```

Некорректно:

```text
Student Service
      |
      v
UPDATE homework_submissions
```

без выполнения команды через владельца Aggregate.

---

## Aggregate Definition

Aggregate состоит из:

```text
Aggregate
├── Aggregate Root
├── Internal Entities
├── Value Objects
├── Invariants
├── Lifecycle
├── Aggregate Version
├── Pending Domain Events
└── References to Other Aggregates
```

---

## Aggregate Root

Aggregate Root:

- имеет глобально уникальный идентификатор;
- является единственной точкой изменения Aggregate;
- проверяет внутренние инварианты;
- контролирует lifecycle;
- увеличивает Aggregate Version;
- создает Domain Events;
- защищает внутренние Entities от прямой модификации;
- не предоставляет наружу mutable internal state.

---

## Aggregate Entity

Entity внутри Aggregate:

- имеет identity в пределах Aggregate;
- не должна загружаться и изменяться отдельно от Root;
- не имеет отдельной consistency boundary;
- может иметь собственный lifecycle;
- изменяется только методами Aggregate Root.

Примеры:

- Homework Submission внутри Homework Assignment;
- Goal Criterion State внутри Goal;
- Reminder Schedule Entry внутри Reminder Plan;
- Notification Delivery Attempt внутри Notification Delivery.

---

## Value Object

Value Object:

- не имеет самостоятельной identity;
- определяется значениями;
- должен быть immutable;
- валидируется при создании;
- может использоваться несколькими Aggregate как одинаковый тип;
- не должен содержать mutable references.

Примеры:

```text
DueDate
TimeWindow
Timezone
ReasonCode
EvidenceReference
SongVersionReference
DeliveryWindow
PrivacyLevel
```

---

## Aggregate Invariant

Invariant — условие, которое должно оставаться истинным после каждой успешной команды.

Примеры:

- Completed Lesson нельзя снова завершить без Reopen;
- Homework не может одновременно быть Completed и Expired;
- Goal Completion требует Completion Decision;
- один Achievement Award не должен выдаваться дважды по одному definition scope;
- Concert Participation не может иметь Slot без Approval;
- Notification Delivery не может быть Delivered до Sending;
- Reminder для Submitted Homework не должен оставаться активным после revalidation.

---

## Transaction Boundary

Одна транзакция должна гарантировать:

```text
Aggregate State Change
+
Aggregate Version Increment
+
Domain Event Outbox Records
```

Между несколькими Aggregate сильная транзакционная согласованность не предполагается по умолчанию.

---

## Aggregate Version

Каждый Aggregate Root имеет monotonic version.

```text
AggregateVersion: 1
AggregateVersion: 2
AggregateVersion: 3
```

Version увеличивается при каждом доменно значимом изменении состояния.

Не обязана увеличиваться при:

- query;
- projection rebuild;
- чтении;
- техническом обновлении, не являющемся доменным фактом;
- повторной идемпотентной команде без изменения.

---

## Aggregate References

Один Aggregate не должен содержать mutable object другого Aggregate.

Допустимо:

```text
RelatedGoalId
StudentId
LessonId
SongVersionId
ConcertId
EvidenceReference
```

Недопустимо:

```text
HomeworkAssignment
└── Full mutable Goal object
```

References разрешают:

- загрузить другой Aggregate отдельно;
- создать cross-aggregate Policy evaluation;
- построить Projection;
- сохранить traceability.

Они не дают права изменять referenced Aggregate.

---

## Cross-Aggregate Invariants

Сильные инварианты по возможности должны принадлежать одному Aggregate.

Если правило охватывает несколько Aggregate, применяются:

- Policy;
- Domain Service;
- Process Manager;
- Saga;
- read model;
- reservation pattern;
- eventual consistency;
- compensation.

Пример:

> Student не должен получить Concert Slot, пока Participation не Approved.

Эта проверка должна находиться внутри ConcertParticipation, если Slot принадлежит тому же Aggregate.

Но:

> Song Readiness должна быть актуальна на момент Eligibility evaluation.

Это cross-aggregate rule:

```text
SongReadiness
      |
      v
Eligibility Evaluation
      |
      v
ConcertParticipation
```

Она требует snapshot references и Policy Decision, но не общей транзакции.

---

# Aggregate Design Rules

## AG-001: Every aggregate has one root

---

## AG-002: All mutations pass through the root

---

## AG-003: Aggregate enforces only invariants it owns

Aggregate не должен притворяться владельцем чужого состояния.

---

## AG-004: Aggregate references other aggregates by identity

---

## AG-005: Aggregate should remain transactionally small

Внутри Aggregate не следует хранить неограниченные коллекции, если они должны загружаться для каждой команды.

---

## AG-006: Historical data may be split from active aggregate state

Например, тысячи Notification Attempts не должны бесконечно увеличивать активный Root.

---

## AG-007: Aggregate root creates domain events

Infrastructure не должна самостоятельно придумывать бизнес-событие.

---

## AG-008: Aggregate must validate lifecycle transitions

---

## AG-009: Invalid transition is rejected explicitly

---

## AG-010: Terminal state history is preserved

---

## AG-011: Reopen is a domain transition, not database flag reset

---

## AG-012: Aggregate version protects concurrent mutations

---

## AG-013: Cross-aggregate updates are not executed inside root methods

---

## AG-014: Policy decision reference is required where policy owns the decision

---

## AG-015: Aggregate must not call external services

---

## AG-016: Aggregate must remain deterministic

Для одинакового состояния и команды результат должен быть предсказуемым, кроме явно переданных values:

- current time;
- generated identifiers;
- policy decisions;
- external snapshots.

---

## AG-017: Current time must be passed explicitly

Aggregate не должен скрыто читать system clock.

---

## AG-018: Random identifiers should be provided or generated by application boundary

---

## AG-019: Aggregate must not depend on UI representation

---

## AG-020: Aggregate must not store unnecessary snapshots of other aggregates

---

## AG-021: Derived values should not be duplicated without need

---

## AG-022: Denormalized data requires ownership and refresh rules

---

## AG-023: Aggregate must not contain unbounded event history

---

## AG-024: Audit history and active state are different concerns

---

## AG-025: One database table does not automatically equal one aggregate

---

## AG-026: One aggregate may use multiple tables

---

## AG-027: Several aggregate types may technically share one table only if boundaries remain enforceable

---

## AG-028: Aggregate boundary is defined by invariants, not CRUD convenience

---

## AG-029: Aggregate methods should express domain language

Предпочтительно:

```text
homework.Submit(...)
goal.Complete(...)
participation.AssignSlot(...)
```

Нежелательно:

```text
aggregate.SetStatus(...)
aggregate.UpdateFields(...)
```

---

## AG-030: Generic status mutation is prohibited

Lifecycle status меняется только через конкретное доменное действие.

---

# Canonical Aggregate Roots

Предварительный набор Aggregate Roots:

```text
Person
SchoolMembership
StudentLearningProfile
Account
Invitation
TeacherAssignment
Lesson
HomeworkAssignment
ProgressRecord
Goal
AchievementDefinition
AchievementAward
StudentSong
SongReadiness
Concert
ConcertParticipation
HomeworkReminderPlan
NotificationIntent
NotificationDelivery
PeriodicReview
DomainIntegrityIssue
```

Некоторые из них могут быть объединены или разделены после implementation-level modeling, но изменение требует явного архитектурного решения.

---

# Aggregate Relationship Overview

```text
Person
    +--> SchoolMembership
    |        |
    |        +--> StudentLearningProfile
    |
    +--> Account
             |
             +--> Invitation
                      |
                      +--> StudentLearningProfile (reference)

StudentLearningProfile
    |
    +--> TeacherAssignment
    |
    +--> Lesson
    |
    +--> HomeworkAssignment
    |
    +--> ProgressRecord
    |
    +--> Goal
    |
    +--> AchievementAward
    |
    +--> StudentSong
    |       |
    |       +--> SongReadiness
    |
    +--> ConcertParticipation
    |
    +--> HomeworkReminderPlan
    |
    +--> NotificationIntent
```

Все связи представлены identifiers, а не вложенными mutable Aggregate.

---

# B.0 Identity and Access Boundaries

Нормативные определения терминов находятся в
`language/001-ubiquitous-language.md`. Настоящий раздел задаёт только структуры,
раздельные жизненные циклы и инварианты, выведенные из PD-0030.

## Person Aggregate

```text
Person
├── PersonId
├── DisplayName
├── ContactReferenceSet
├── CreatedAt
├── UpdatedAt
└── Version
```

`Person` не владеет учебным статусом, паролем или приглашением. Точные правила
слияния, архивирования и изменения контактных данных находятся вне B.0 и не
вводятся без отдельного источника.

## SchoolMembership Aggregate

```text
SchoolMembership
├── SchoolMembershipId
├── PersonId
├── SchoolId
├── MembershipStatusReference
├── RecordedAt
└── Version
```

Для создания `Student` требуется ссылка на существующий `SchoolMembership` в
том же tenant. Названия состояний и переходы membership не определены в B.0:
они требуют подтверждённых наблюдений реальной работы школы. Возможные
операционные storage flags реализации не являются утверждёнными доменными
состояниями и не могут использоваться как источник продуктового смысла.

## Account Aggregate

```text
Account
├── AccountId
├── PersonId
├── TenantId
├── AccessStatus
├── CredentialReference
├── RoleGrantReferences
├── ActivatedAt
├── SuspendedAt
├── CreatedAt
├── UpdatedAt
└── Version
```

Жизненный цикл B.0:

```text
PendingActivation --> Active --> Suspended
```

`PendingActivation` и `Active` принадлежат только `Account`. Изменение
`AccessStatus` не изменяет учебный статус `Student` и не удаляет его историю.
`Suspended` — производное техническое отображение продуктового понятия
заблокированного доступа, а не новое определение Language. Пароль создаётся
учеником только в ходе успешной `Activation`.

## Invitation Aggregate

```text
Invitation
├── InvitationId
├── TenantId
├── AccountId
├── StudentId
├── FirstBelcantoMinuteReference
├── TokenDigestReference
├── Status
├── IssuedAt
├── ExpiresAt
├── ConsumedAt
├── RevokedAt
├── SupersededByInvitationId
└── Version
```

Хранимые переходы B.0:

```text
Issued --> Consumed
   |----> Revoked
   `----> Superseded
```

Истечение является эффективным условием `Issued && ExpiresAt <= now`, а не
отдельным persisted status: такая запись остаётся `Issued`, но больше не
допускает `Activation` и при следующем выпуске переводится в `Superseded`.

Инварианты:

- только `Issued` и неистёкшее `Invitation` допускает `Activation`;
- одно `Invitation` потребляется не более одного раза;
- перевыпуск переводит прежнее неиспользованное `Invitation` в `Superseded`;
- выпуск запрещён без `FirstBelcantoMinuteReference`;
- отзыв или истечение не удаляет `Student`, `Account` или учебную историю;
- секрет приглашения не хранится в открытом виде в доменной записи;
- все ссылки принадлежат одному tenant.

---

# StudentLearningProfile Aggregate

## Aggregate Root

StudentLearningProfile

## Responsibility

Отвечает за состояние Student внутри learning product после enrollment boundary.

Не является:

- CRM lead;
- sales contact;
- billing account;
- marketing profile;
- full identity provider profile.

## Owned State

```text
StudentLearningProfile
├── StudentId
├── PersonId
├── SchoolMembershipId
├── SchoolId
├── EnrollmentReference
├── LearningStatus
├── Locale
├── Timezone
├── LearningPreferences
├── ActiveLearningPause
├── CreatedAt
├── UpdatedAt
└── Version
```

## Possible Statuses

- Active
- Learning Paused
- Inactive
- Archived

Это только учебные состояния. Состояние активации цифрового доступа находится
в `Account` и не входит в `StudentLearningProfile`.

## Owned Invariants

- Student belongs to one School tenant.
- Person, SchoolMembership and Student references belong to the same tenant.
- Active Learning Pause is unique.
- Pause cannot end before it starts.
- Archived Student cannot receive ordinary learning mutations.
- Timezone must be valid.
- Learning Pause does not erase learning history.
- Product profile cannot exist without enrollment reference unless created through migration.
- CRM lifecycle must not be stored as learning status.
- Account or Invitation lifecycle must not be stored as learning status.

## Commands

```text
CreateStudentProfile
UpdateStudentProfile
ChangeStudentTimezone
StartStudentLearningPause
EndStudentLearningPause
```

## Events

```text
StudentCreated
StudentProfileUpdated
StudentTimezoneChanged
StudentLearningPauseStarted
StudentLearningPauseEnded
```

## References

```text
EnrollmentReference
PersonId
SchoolMembershipId
CurrentTeacherAssignmentIds
```

Teacher assignments may be projected rather than stored as authoritative references.

---

# TeacherAssignment Aggregate

## Aggregate Root

TeacherAssignment

## Responsibility

Определяет формальную связь между Teacher и Student в конкретном educational scope.

Не следует хранить assignment только как поле teacher_id в Student, если требуется:

- история;
- несколько ролей Teacher;
- временные назначения;
- delegation;
- reassignment;
- scope by subject, song или lesson type.

## Owned State

```text
TeacherAssignment
├── TeacherAssignmentId
├── StudentId
├── TeacherId
├── AssignmentType
├── Scope
├── EffectiveFrom
├── EffectiveUntil
├── Status
├── DelegationReferences
├── AssignedBy
├── EndedBy
└── Version
```

## Assignment Types

```text
Primary Teacher
Substitute Teacher
Song Coach
Performance Coach
Group Teacher
Review Teacher
```

## Owned Invariants

- Assignment period is valid.
- End date cannot precede start date.
- Only one active Primary Teacher exists for the same scope unless explicitly supported.
- Ended assignment cannot authorize new commands.
- Delegation cannot exceed assignment scope.
- Assignment belongs to same tenant as Student and Teacher.
- Historical reassignment cannot be rewritten.

## Commands

```text
AssignTeacherToStudent
ReassignTeacher
EndTeacherAssignment
DelegateTeacherAssignment
RevokeTeacherDelegation
```

## Events

```text
TeacherAssignedToStudent
TeacherReassigned
TeacherAssignmentEnded
TeacherAssignmentDelegated
TeacherDelegationRevoked
```

---

# Lesson Aggregate

## Aggregate Root

Lesson

## Responsibility

Владеет lifecycle конкретного scheduled learning session.

Lesson отвечает за:

- schedule;
- participants;
- Teacher;
- format;
- start;
- completion;
- cancellation;
- authoritative Lesson status;
- completion evidence references;
- summary references.

Lesson не владеет:

- Student Progress;
- Goal Completion;
- Homework lifecycle;
- Notification delivery;
- Payment;
- CRM attendance sales logic.

## Owned State

```text
Lesson
├── LessonId
├── SchoolId
├── LessonType
├── StudentParticipantIds
├── TeacherId
├── ScheduledTime
├── ActualTime
├── Timezone
├── Format
├── LocationReference
├── Status
├── AttendanceRecords
├── CompletionRecord
├── SummaryReference
├── CancellationRecord
├── CreatedAt
├── UpdatedAt
└── Version
```

## Lesson Statuses

```text
Draft
Scheduled
In Progress
Completed
Cancelled
Archived
```

## Internal Entities

### AttendanceRecord

```text
AttendanceRecord
├── StudentId
├── AttendanceStatus
├── RecordedBy
├── RecordedAt
└── ReasonCategory
```

Possible statuses:

```text
Present
Late
Absent
Excused
Unknown
```

Attendance is not automatically Progress Evidence.

## Owned Invariants

- ScheduledEnd is after ScheduledStart.
- Lesson must have at least one Student.
- Lesson must have one responsible Teacher.
- Completed Lesson cannot be started again.
- Cancelled Lesson cannot be completed without explicit restoration flow.
- Completion requires Lesson Completion Policy conditions.
- ActualEnd cannot precede ActualStart.
- Attendance record is unique per Student.
- Nonparticipant Student cannot receive attendance record.
- Lesson completion history is immutable.
- Reschedule preserves prior schedule in event history.
- Lesson timezone must be explicit.

## Commands

```text
ScheduleLesson
RescheduleLesson
CancelLesson
StartLesson
CompleteLesson
RecordLessonSummary
RecordLessonAttendance
CorrectLessonAttendance
ArchiveLesson
```

## Events

```text
LessonScheduled
LessonRescheduled
LessonCancelled
LessonStarted
LessonCompleted
LessonCompletionRejected
LessonSummaryRecorded
LessonAttendanceRecorded
LessonAttendanceCorrected
LessonArchived
```

## Cross-Aggregate Reactions

LessonCompleted may lead to:

```text
EvaluateProgressUpdate
AssignHomework
RequestGoalReview
AddSongReadinessEvidence
```

Но Lesson Aggregate не выполняет эти изменения сам.

---

# HomeworkAssignment Aggregate

## Aggregate Root

HomeworkAssignment

## Responsibility

Владеет полным lifecycle конкретного Homework, назначенного конкретному Student или явно определенной группе.

Homework Assignment отвечает за:

- assignment;
- current version;
- instructions reference;
- deadline strategy;
- due date;
- submissions;
- review;
- correction requests;
- blockers;
- clarification;
- overdue state;
- grace period;
- expiration;
- replacement;
- cancellation;
- completion;
- reopening;
- archival.

## Owned State

```text
HomeworkAssignment
├── HomeworkAssignmentId
├── SchoolId
├── StudentId
├── AssignedByTeacherId
├── HomeworkDefinitionReference
├── CurrentHomeworkVersion
├── AssignmentType
├── Requiredness
├── InstructionsReference
├── MaterialReferences
├── DeadlineConfiguration
├── ReminderStrategyReference
├── ExpirationStrategy
├── RelatedLessonId
├── RelatedGoalIds
├── RelatedSongVersionIds
├── RelatedConcertId
├── Status
├── ActiveSubmission
├── SubmissionHistoryReferences
├── CurrentReview
├── CorrectionRequest
├── ActiveBlockers
├── ClarificationState
├── GracePeriod
├── ExpirationRecord
├── ReplacementReference
├── CancellationRecord
├── CompletionRecord
├── CreatedAt
├── UpdatedAt
└── Version
```

## Homework Statuses

```text
Draft
Assigned
In Progress
Submitted
Under Review
Clarification Required
Correction Requested
Overdue
Completed
Cancelled
Replaced
Expired
Archived
```

## Internal Entities

### HomeworkSubmission

```text
HomeworkSubmission
├── SubmissionId
├── SubmissionVersion
├── SubmittedBy
├── SubmittedAt
├── SubmissionMethod
├── AttachmentReferences
├── TextSubmissionReference
├── Status
└── WithdrawalRecord
```

### HomeworkReview

```text
HomeworkReview
├── ReviewId
├── SubmissionId
├── TeacherId
├── ReviewOutcome
├── FeedbackReference
├── EvidenceReferences
├── StartedAt
└── CompletedAt
```

### HomeworkBlocker

```text
HomeworkBlocker
├── BlockerId
├── BlockerCategory
├── ExplanationReference
├── ReportedBy
├── ReportedAt
├── Status
└── ResolutionRecord
```

### CorrectionRequest

```text
CorrectionRequest
├── CorrectionRequestId
├── SubmissionId
├── CorrectionReference
├── RequestedBy
├── RequestedAt
├── NewDueDate
├── DeadlineType
└── Status
```

## Owned Invariants

- Homework belongs to one Student unless group assignment is explicitly modeled separately.
- Homework Version increases when educational content or completion expectations change.
- Submitted Homework cannot be silently expired.
- Completed Homework cannot become Overdue.
- Replaced Homework cannot remain active.
- Cancelled and Expired are distinct.
- Only one active Submission exists unless multiple concurrent submissions are explicitly supported.
- Submission belongs to the same Student.
- Review refers to a valid Submission.
- Correction Request refers to reviewed Submission.
- Correction deadline is explicit when required.
- Due Date changes preserve history.
- Reopen preserves prior Expiration.
- Expiration requires policy Decision where configured.
- Completion requires valid Completion method.
- Reminder delivery status does not change Homework status.
- Notification open does not prove Homework start.
- Missing material Blocker may prevent ordinary overdue transition.
- Terminal states require explicit lifecycle commands.

## Commands

```text
AssignHomework
UpdateHomework
ChangeHomeworkDueDate
StartHomework
SubmitHomework
WithdrawHomeworkSubmission
StartHomeworkReview
ReviewHomework
RequestHomeworkCorrection
RequestHomeworkClarification
ReportHomeworkBlocker
ResolveHomeworkBlocker
MarkHomeworkOverdue
StartHomeworkGracePeriod
ExtendHomeworkDueDate
ExpireHomework
ReopenHomework
CancelHomework
ReplaceHomework
CompleteHomework
ArchiveHomework
```

## Events

```text
HomeworkAssigned
HomeworkUpdated
HomeworkDueDateChanged
HomeworkStarted
HomeworkSubmitted
HomeworkSubmissionWithdrawn
HomeworkReviewStarted
HomeworkReviewed
HomeworkCorrectionRequested
HomeworkClarificationRequested
HomeworkBlockerReported
HomeworkBlockerResolved
HomeworkMarkedOverdue
HomeworkGracePeriodStarted
HomeworkDueDateExtended
HomeworkExpired
HomeworkReopened
HomeworkCancelled
HomeworkReplaced
HomeworkCompleted
HomeworkArchived
```

## Boundary Decision

Reminder Plan не входит внутрь Homework Assignment.

Причины:

- Reminder имеет отдельную schedule lifecycle;
- Reminder entries могут быть многочисленными;
- Reminder должен пересчитываться независимо;
- Notification Delivery имеет собственную retry-модель;
- Homework не должен загружать reminder history при Submission;
- reminder processing может масштабироваться отдельно.

Связь:

```text
HomeworkAssignment
      |
      | HomeworkAssignmentId + HomeworkVersion
      v
HomeworkReminderPlan
```

---

# ProgressRecord Aggregate

## Aggregate Root

ProgressRecord

## Responsibility

Хранит authoritative learning Progress Student по определенному scope.

Scope должен быть явным.

Примеры:

```text
Overall Learning
Skill
Technique
Goal
Song
Performance Area
Curriculum Module
```

## Owned State

```text
ProgressRecord
├── ProgressId
├── StudentId
├── Scope
├── ProgressDimensions
├── CurrentLevelStates
├── EvidenceReferences
├── EvidenceValidityState
├── LastEvaluationReference
├── LastReviewedAt
├── CreatedAt
├── UpdatedAt
└── Version
```

## Internal Entities

### ProgressDimensionState

```text
ProgressDimensionState
├── DimensionId
├── CurrentState
├── Confidence
├── EvidenceReferences
├── UpdatedAt
└── EvaluationReference
```

## Owned Invariants

- Progress belongs to one Student and one explicit scope.
- Progress changes require accepted Evidence or authorized Teacher decision.
- Reminder, Notification Open и lateness are not Progress Evidence.
- Evidence must belong to Student.
- Invalidated Evidence cannot support new Progress decision.
- Progress cannot be overwritten by stale evaluation.
- Progress history is preserved through events.
- Multiple dimensions may change in one evaluation only if they belong to the same Progress scope.
- Cross-scope Progress changes require separate commands.
- Progress is not a generic numerical score unless a documented model defines it.

## Commands

```text
RecordProgressEvidence
InvalidateProgressEvidence
EvaluateProgressUpdate
UpdateProgress
RequestProgressReview
```

## Events

```text
ProgressEvidenceRecorded
ProgressEvidenceInvalidated
ProgressUpdateEvaluated
ProgressUpdated
ProgressUpdateRejected
ProgressReviewRequested
```

## Boundary Decision

Progress Evidence may be stored in a dedicated Evidence registry later, but ProgressRecord owns which evidence currently supports its state.

The source Aggregate still owns the original fact.

Пример:

```text
HomeworkReviewed
      |
      v
ProgressEvidenceRecorded
      |
      v
ProgressRecord
```

ProgressRecord does not own HomeworkReview.

---

# Goal Aggregate

## Aggregate Root

Goal

## Responsibility

Владеет lifecycle конкретной образовательной Goal Student.

Goal отвечает за:

- criterion;
- target scope;
- activation;
- progress state;
- evidence references;
- review state;
- completion decision;
- reopening;
- cancellation;
- archival.

## Owned State

```text
Goal
├── GoalId
├── StudentId
├── GoalType
├── CriterionReference
├── DescriptionReference
├── RelatedSkillIds
├── RelatedSongVersionIds
├── RelatedConcertId
├── Status
├── ProgressState
├── EvidenceReferences
├── BlockingConditions
├── ReviewSchedule
├── CurrentReview
├── CompletionRecord
├── ReopenHistoryReferences
├── CancellationRecord
├── CreatedAt
├── UpdatedAt
└── Version
```

## Goal Statuses

```text
Draft
Active
Review Required
Completed
Cancelled
Archived
```

## Owned Invariants

- Goal has one Student.
- Goal criterion is explicit.
- Goal cannot be Completed without Completion Decision.
- Goal cannot be completed twice without Reopen.
- Reopen preserves previous Completion.
- Cancelled Goal cannot receive ordinary progress updates.
- Evidence belongs to Goal scope or is explicitly mapped.
- Stale Evidence cannot complete Goal where freshness matters.
- Goal Completion does not automatically complete related Homework.
- Expired Homework does not automatically cancel Goal.
- Goal cannot be rewritten into a materially different objective without versioning or replacement.
- Archived Goal is immutable except retention operations.

## Commands

```text
CreateGoal
ActivateGoal
UpdateGoal
UpdateGoalProgress
RequestGoalReview
EvaluateGoalCompletion
CompleteGoal
ReopenGoal
CancelGoal
ArchiveGoal
```

## Events

```text
GoalCreated
GoalActivated
GoalUpdated
GoalProgressUpdated
GoalReviewRequested
GoalCompletionEvaluated
GoalCompleted
GoalReopened
GoalCancelled
GoalArchived
```

---

# AchievementDefinition Aggregate

## Aggregate Root

AchievementDefinition

## Responsibility

Определяет версионируемое правило Achievement.

Definition и Award разделены.

Причина:

- изменение definition не должно переписывать уже выданные Awards;
- один Definition применяется к множеству Student;
- Award имеет собственный lifecycle;
- публикация Definition и присуждение Award имеют разные permissions.

## Owned State

```text
AchievementDefinition
├── AchievementDefinitionId
├── Name
├── DescriptionReference
├── DefinitionVersion
├── CriterionReference
├── EvidenceRequirements
├── Visibility
├── RepeatabilityRule
├── Status
├── PublishedAt
├── RetiredAt
└── Version
```

## Statuses

```text
Draft
Published
Retired
Archived
```

## Owned Invariants

- Published Definition is immutable within its DefinitionVersion.
- Material change creates new DefinitionVersion.
- Retired Definition cannot create new Award unless explicitly allowed.
- Repeatability rule is explicit.
- Criterion must be evaluable.
- Visibility must be defined.
- Definition cannot contain hidden punitive behavior.

## Commands

```text
CreateAchievementDefinition
UpdateAchievementDefinition
PublishAchievementDefinition
RetireAchievementDefinition
ArchiveAchievementDefinition
```

## Events

```text
AchievementDefinitionCreated
AchievementDefinitionUpdated
AchievementDefinitionPublished
AchievementDefinitionRetired
AchievementDefinitionArchived
```

---

# AchievementAward Aggregate

## Aggregate Root

AchievementAward

## Responsibility

Владеет конкретным Award, выданным конкретному Student.

## Owned State

```text
AchievementAward
├── AchievementAwardId
├── AchievementDefinitionId
├── AchievementDefinitionVersion
├── StudentId
├── Status
├── AwardDecisionId
├── EvidenceReferences
├── AwardedAt
├── AwardedBy
├── RevocationRecord
├── RestoreHistoryReferences
└── Version
```

## Statuses

```text
Awarded
Revoked
Restored
Archived
```

## Owned Invariants

- Award references immutable DefinitionVersion.
- Award requires Award Decision.
- Duplicate Award follows repeatability rule.
- Revocation preserves original Award.
- Restoration preserves Revocation.
- AI cannot be authoritative Award actor.
- Notification delivery does not create Award.
- Achievement does not imply current ability forever unless definition says so.
- Archived Award remains historically traceable.

## Commands

```text
AwardAchievement
RevokeAchievement
RestoreAchievement
ArchiveAchievementAward
```

## Events

```text
AchievementAwarded
AchievementRevoked
AchievementRestored
AchievementAwardArchived
```

---

# StudentSong Aggregate

## Aggregate Root

StudentSong

## Responsibility

Определяет связь Student с конкретной Song и активной Song Version.

Song catalog itself may belong to a separate content context.

StudentSong отвечает за:

- repertoire membership;
- current Song Version;
- purpose;
- active/inactive state;
- version change history;
- removal from active repertoire.

## Owned State

```text
StudentSong
├── StudentSongId
├── StudentId
├── SongId
├── CurrentSongVersionId
├── Purpose
├── Status
├── AddedBy
├── AddedAt
├── VersionHistoryReferences
├── RemovedAt
└── Version
```

## Statuses

```text
Active
Paused
Removed
Archived
```

## Owned Invariants

- One StudentSong represents one Student + Song relationship.
- Current Song Version must belong to same Song.
- Removed StudentSong cannot receive ordinary readiness updates.
- Version change preserves history.
- Song Version change triggers reevaluation but does not directly change Readiness.
- Repertoire membership does not imply Concert Eligibility.
- Song removal does not erase prior performance history.

## Commands

```text
AddSongToStudentRepertoire
ChangeStudentSongVersion
PauseStudentSong
ResumeStudentSong
RemoveSongFromStudentRepertoire
ArchiveStudentSong
```

## Events

```text
SongAddedToStudentRepertoire
StudentSongVersionChanged
StudentSongPaused
StudentSongResumed
SongRemovedFromStudentRepertoire
StudentSongArchived
```

---

# SongReadiness Aggregate

## Aggregate Root

SongReadiness

## Identity Scope

Рекомендуемый key:

```text
StudentId
+
SongVersionId
+
PerformanceType
```

Каждая комбинация имеет отдельный Aggregate.

## Responsibility

Владеет текущей readiness assessment для конкретного Student, Song Version и Performance Type.

## Owned State

```text
SongReadiness
├── SongReadinessId
├── StudentId
├── SongVersionId
├── PerformanceType
├── Status
├── AreaStates
├── EvidenceReferences
├── BlockingConditions
├── Conditions
├── LastEvaluation
├── ReviewState
├── EvaluatedAt
├── ValidUntil
└── Version
```

## Readiness Statuses

```text
Not Evaluated
Not Ready
Conditionally Ready
Ready
Review Required
Stale
```

## Internal Entity

### ReadinessAreaState

```text
ReadinessAreaState
├── Area
├── Outcome
├── EvidenceReferences
├── BlockingConditions
├── AssessedAt
└── Confidence
```

Possible areas:

```text
Vocal
Technical
Memory
Interpretation
Performance
Safety
Rehearsal
Material Preparedness
```

## Owned Invariants

- Readiness is scoped to one Song Version and Performance Type.
- Changing Song Version does not transfer Readiness automatically.
- Readiness requires policy Evaluation.
- Evidence belongs to Student and Song Version.
- Stale Evidence cannot support current Ready status when freshness is required.
- Ready does not mean Concert Approved.
- Concert outcome does not rewrite historical Readiness.
- AI cannot finalize Readiness independently.
- Current Status corresponds to latest accepted Evaluation.
- Previous evaluations remain auditable.
- Readiness may expire or become Stale without becoming Not Ready automatically.

## Commands

```text
AddSongReadinessEvidence
InvalidateSongReadinessEvidence
RequestSongReadinessEvaluation
EvaluateSongReadiness
ChangeSongReadiness
RequestSongReadinessReview
MarkSongReadinessStale
```

## Events

```text
SongReadinessEvidenceAdded
SongReadinessEvidenceInvalidated
SongReadinessEvaluationRequested
SongReadinessEvaluated
SongReadinessChanged
SongReadinessReviewRequired
SongReadinessMarkedStale
```

---

# Concert Aggregate

## Aggregate Root

Concert

## Responsibility

Владеет самим event:

- identity;
- schedule;
- venue;
- lifecycle;
- requirements version;
- program publication;
- general configuration.

Concert не должен содержать все Participation как mutable children, если число участников может быть большим и каждый Participation имеет сложный lifecycle.

## Owned State

```text
Concert
├── ConcertId
├── SchoolId
├── Title
├── ConcertType
├── Schedule
├── Timezone
├── VenueReference
├── Status
├── CurrentRequirementsVersion
├── RequirementsReference
├── CurrentProgramVersion
├── ProgramReference
├── CapacityConfiguration
├── CreatedAt
├── UpdatedAt
├── CancelledAt
├── CompletedAt
└── Version
```

## Statuses

```text
Draft
Published
Active Preparation
Completed
Cancelled
Archived
```

## Owned Invariants

- ScheduledEnd is after ScheduledStart.
- Requirements Version is immutable after publication.
- Material requirement change creates new version.
- Published Program has a Program Version.
- Cancelled Concert cannot be completed.
- Completed Concert cannot accept new ordinary Participation.
- Timezone is explicit.
- Venue changes preserve history.
- Concert does not decide Student Eligibility.
- Concert does not own Student Song Readiness.
- Program publication must reference valid approved Participations through application validation.

## Commands

```text
CreateConcert
UpdateConcert
PublishConcertRequirements
PublishConcert
CancelConcert
CompleteConcert
PublishConcertProgram
ArchiveConcert
```

## Events

```text
ConcertCreated
ConcertUpdated
ConcertRequirementsPublished
ConcertPublished
ConcertCancelled
ConcertCompleted
ConcertProgramPublished
ConcertArchived
```

---

# ConcertParticipation Aggregate

## Aggregate Root

ConcertParticipation

## Identity Scope

Одна Participation представляет:

```text
Student
+
Concert
+
Performance Type
+
one participation proposal
```

Song Versions may be one or several depending on performance model.

## Responsibility

Владеет lifecycle участия Student в Concert:

- proposal;
- consent;
- eligibility;
- approval;
- conditions;
- program placement;
- slot assignment;
- withdrawal;
- performance completion.

## Owned State

```text
ConcertParticipation
├── ConcertParticipationId
├── ConcertId
├── StudentId
├── PerformanceType
├── SongVersionIds
├── Status
├── ConsentState
├── EligibilityState
├── ApprovalState
├── ProgramPlacementState
├── PerformanceSlot
├── RehearsalRequirements
├── Conditions
├── BlockingConditions
├── WithdrawalRecord
├── PerformanceCompletionRecord
├── CreatedAt
├── UpdatedAt
└── Version
```

## Participation Statuses

```text
Proposed
Consent Required
Under Eligibility Review
Not Eligible
Conditionally Eligible
Eligible
Awaiting Approval
Approved
Program Placed
Slot Assigned
Withdrawn
Performance Completed
Cancelled
Archived
```

Не все состояния обязаны быть реализованы одним enum. Возможно использование orthogonal state components:

```text
ConsentState
EligibilityState
ApprovalState
ProgramState
PerformanceState
```

Это предпочтительнее, если один линейный status создает комбинационный взрыв.

## Internal Value Objects

### EligibilityState

```text
EligibilityState
├── Status
├── DecisionId
├── RequirementsVersion
├── EvaluatedAt
├── Conditions
├── BlockingConditions
└── ValidUntil
```

### ConsentState

```text
ConsentState
├── Status
├── ConsentRequestId
├── ConsentId
├── Scope
├── GrantedAt
├── WithdrawnAt
└── ExpiresAt
```

### ApprovalState

```text
ApprovalState
├── Status
├── ApprovedBy
├── ApprovedAt
└── ApprovalScope
```

### PerformanceSlot

```text
PerformanceSlot
├── PerformanceSlotId
├── ScheduledStart
├── StageReference
├── AssignedBy
└── AssignedAt
```

## Owned Invariants

- Participation belongs to one Student and one Concert.
- Eligibility is scoped to Song Versions and Performance Type.
- Eligibility Decision references Concert Requirements Version.
- Approval is distinct from Eligibility.
- Program Placement is distinct from Approval.
- Slot cannot be assigned before required Approval.
- Withdrawn Participation cannot receive a new Slot without Reopen flow.
- Consent withdrawal invalidates dependent participation actions according to scope.
- Not Eligible Participation cannot be Approved without new Eligibility Decision.
- Conditional Eligibility requires explicit Conditions.
- Conditions have status and optional deadline.
- Song Version change invalidates or reevaluates Eligibility.
- Concert Requirements change may mark Eligibility stale.
- Performance Completion requires valid assigned or otherwise authorized performance context.
- Performance Completion does not automatically award Achievement.
- AI cannot finalize Eligibility or Approval.
- Student-visible explanation does not expose private evaluation notes.

## Commands

```text
ProposeConcertParticipation
RequestConcertConsent
GrantConcertConsent
WithdrawConcertConsent
RequestConcertEligibilityEvaluation
EvaluateConcertEligibility
MarkConcertParticipationEligible
MarkConcertParticipationConditionallyEligible
MarkConcertParticipationNotEligible
ApproveConcertParticipation
WithdrawConcertParticipation
AssignConcertPerformanceSlot
ChangeConcertPerformanceSlot
RemoveConcertPerformanceSlot
CompleteConcertPerformance
ReopenConcertParticipation
CancelConcertParticipation
ArchiveConcertParticipation
```

## Events

```text
ConcertParticipationProposed
ConcertConsentRequested
ConcertConsentGranted
ConcertConsentWithdrawn
ConcertEligibilityEvaluationRequested
ConcertEligibilityEvaluated
ConcertParticipationMarkedEligible
ConcertParticipationMarkedConditionallyEligible
ConcertParticipationMarkedNotEligible
ConcertParticipationApproved
ConcertParticipationWithdrawn
ConcertPerformanceSlotAssigned
ConcertPerformanceSlotChanged
ConcertPerformanceSlotRemoved
ConcertPerformanceCompleted
ConcertParticipationReopened
ConcertParticipationCancelled
ConcertParticipationArchived
```

---

# HomeworkReminderPlan Aggregate

## Aggregate Root

HomeworkReminderPlan

## Responsibility

Владеет reminder strategy и scheduled reminder lifecycle для одного Homework Assignment version.

Рекомендуемый scope:

```text
HomeworkAssignmentId
+
HomeworkVersion
+
StudentId
```

## Owned State

```text
HomeworkReminderPlan
├── ReminderPlanId
├── HomeworkAssignmentId
├── HomeworkVersion
├── StudentId
├── Strategy
├── Status
├── Timezone
├── DueDateSnapshot
├── QuietHoursSnapshotReference
├── MaximumReminderCount
├── ScheduledReminders
├── DeliveredReminderCount
├── SuppressionState
├── RecalculationHistoryReferences
├── CreatedAt
├── UpdatedAt
└── Version
```

## Plan Statuses

```text
Draft
Active
Suspended
Completed
Suppressed
Cancelled
Expired
Archived
```

## Internal Entity

### ScheduledReminder

```text
ScheduledReminder
├── ReminderId
├── ReminderType
├── ScheduledFor
├── Status
├── NotificationIntentId
├── DeliveryReference
├── SuppressionReason
├── AttemptReference
└── Version
```

## Reminder Statuses

```text
Scheduled
Due
Delivery Requested
Delivered
Failed
Rescheduled
Suppressed
Cancelled
Expired
```

## Owned Invariants

- Plan belongs to one Homework Assignment Version.
- Reminder count does not exceed configured maximum.
- Duplicate reminder type and time window are prevented.
- Completed, Submitted, Cancelled, Replaced or Expired Homework suppresses ordinary reminders after revalidation.
- Reminder does not evaluate Homework quality.
- Reminder does not change Progress.
- Delivery and Reading are distinct.
- Quiet Hours affect scheduling, not educational status.
- Learning Pause suppression is category-specific.
- Plan recalculation preserves previous scheduled history.
- Old Homework Version plan cannot schedule new reminders after replacement.
- Reminder due event does not authorize direct delivery without revalidation.
- AI cannot increase pressure or maximum count.

## Commands

```text
CreateHomeworkReminderPlan
ScheduleHomeworkReminder
RescheduleHomeworkReminder
EvaluateHomeworkReminder
RequestHomeworkReminderDelivery
SuppressHomeworkReminder
CancelHomeworkReminder
ExpireHomeworkReminder
RecalculateHomeworkReminderPlan
CompleteHomeworkReminderPlan
ArchiveHomeworkReminderPlan
```

## Events

```text
HomeworkReminderPlanCreated
HomeworkReminderScheduled
HomeworkReminderRescheduled
HomeworkReminderDue
HomeworkReminderSuppressed
HomeworkReminderCancelled
HomeworkReminderExpired
HomeworkReminderDelivered
HomeworkReminderDeliveryFailed
HomeworkReminderPlanCompleted
HomeworkReminderPlanArchived
```

---

# NotificationIntent Aggregate

## Aggregate Root

NotificationIntent

## Responsibility

Владеет доменным намерением коммуникации.

Отвечает за:

- source reference;
- recipient;
- category;
- purpose;
- priority;
- privacy;
- delivery window;
- requested channels;
- approval;
- suppression;
- cancellation;
- expiration;
- bundling membership.

Не отвечает за технические attempts конкретного канала.

## Owned State

```text
NotificationIntent
├── NotificationIntentId
├── IntentType
├── SourceDomain
├── SourceEntityReference
├── Recipient
├── Category
├── Priority
├── Urgency
├── RequiredAction
├── RequestedChannels
├── DeliveryWindow
├── ExpiresAt
├── TemplateReference
├── RenderingParametersReference
├── PrivacyLevel
├── DeduplicationKey
├── Status
├── ApprovalRecord
├── SuppressionRecord
├── CancellationRecord
├── BundleReference
├── CreatedAt
├── UpdatedAt
└── Version
```

## Intent Statuses

```text
Draft
Pending Review
Approved
Scheduled
Bundled
Suppressed
Cancelled
Expired
Completed
Archived
```

## Owned Invariants

- Intent has one Recipient.
- Source Entity reference is required.
- Deduplication Key is stable.
- Intent cannot be Approved after Expiration.
- Suppressed Intent cannot create external Delivery unless unsuppressed through an explicit flow.
- Cancelled Intent cannot be sent.
- Privacy Level limits allowed channels.
- Critical Priority requires authorization.
- Intent source version is immutable after approval; source change requires new evaluation or new Intent.
- Marketing and Educational categories remain separate.
- Recipient cannot be changed silently.
- Intent does not mark domain action completed.
- Delivery success does not modify source domain state.

## Commands

```text
CreateNotificationIntent
ApproveNotificationIntent
RejectNotificationIntent
SuppressNotification
CancelNotification
BundleNotifications
ExpireNotification
CompleteNotificationIntent
ArchiveNotification
RequestNotificationReview
```

## Events

```text
NotificationIntentCreated
NotificationIntentApproved
NotificationIntentRejected
NotificationBundled
NotificationSuppressed
NotificationCancelled
NotificationExpired
NotificationActionCompleted
NotificationArchived
NotificationReviewRequested
```

---

# NotificationDelivery Aggregate

## Aggregate Root

NotificationDelivery

## Responsibility

Владеет технически значимым lifecycle одной channel delivery попытки или delivery process.

Рекомендуемый scope:

```text
NotificationIntentId
+
RecipientId
+
Channel
+
Delivery sequence
```

## Owned State

```text
NotificationDelivery
├── NotificationDeliveryId
├── NotificationIntentId
├── RecipientId
├── Channel
├── DestinationReference
├── Status
├── ScheduledFor
├── ExpiresAt
├── RenderedContentReference
├── TemplateVersion
├── IdempotencyKey
├── Attempts
├── CurrentAttemptNumber
├── ProviderReference
├── DeliveredAt
├── OpenedAt
├── ActionCompletedAt
├── FailureState
├── FallbackState
├── CreatedAt
├── UpdatedAt
└── Version
```

## Delivery Statuses

```text
Draft
Scheduled
Queued
Rendering
Ready
Sending
Delivered
Opened
Action Completed
Delivery Failed
Retry Scheduled
Cancelled
Suppressed
Expired
Archived
```

## Internal Entity

### DeliveryAttempt

```text
DeliveryAttempt
├── AttemptNumber
├── RequestedAt
├── StartedAt
├── CompletedAt
├── ProviderReference
├── Outcome
├── FailureCategory
├── FailureCode
├── Retryable
└── IdempotencyReference
```

Active aggregate may retain only bounded recent attempts, while full attempt history can move to audit storage.

## Owned Invariants

- Delivery belongs to one Intent, Recipient and Channel.
- Sending requires valid rendered content.
- Delivered cannot transition back to Failed due to duplicate callback.
- Opened requires Delivered or equivalent provider semantics.
- Action Completed does not execute source domain mutation itself.
- Retry attempt number is monotonic.
- Retry cannot outlive ExpiresAt.
- Permanent failure stops Retry.
- Fallback must be separately authorized.
- Duplicate provider callback is idempotent.
- Destination reference cannot be changed after Sending without new delivery sequence.
- Delivery status cannot authorize an educational decision.
- Notification Open is not Progress Evidence.
- Provider callback must be trusted.
- Archived Delivery remains auditable.

## Commands

```text
ScheduleNotification
RescheduleNotification
RenderNotification
SendNotification
RetryNotificationDelivery
StopNotificationRetry
SwitchNotificationChannel
MarkNotificationDelivered
MarkNotificationDeliveryFailed
MarkNotificationOpened
CompleteNotificationAction
CancelNotificationDelivery
ExpireNotificationDelivery
ArchiveNotificationDelivery
```

## Events

```text
NotificationScheduled
NotificationRescheduled
NotificationRenderingRequested
NotificationRendered
NotificationSendingRequested
NotificationDelivered
NotificationOpened
NotificationActionCompleted
NotificationDeliveryFailed
NotificationRetryScheduled
NotificationRetryStopped
NotificationChannelSwitched
NotificationDeliveryCancelled
NotificationDeliveryExpired
NotificationDeliveryArchived
```

## Boundary Decision

NotificationIntent and NotificationDelivery are separate Aggregate Roots.

Причины:

- один Intent может иметь несколько deliveries;
- multi-channel delivery;
- retries and attempts are operationally heavy;
- recipient intent lifecycle differs from provider delivery lifecycle;
- privacy and authorization decision should not be locked by provider callback;
- high-volume callbacks should not contend on Intent Aggregate;
- delivery can be archived separately.

---

# PeriodicReview Aggregate

## Aggregate Root

PeriodicReview

## Responsibility

Владеет одной review request или review execution для одного Aggregate target.

Review Cycle discovery may be a separate operational process rather than one huge Aggregate.

## Owned State

```text
PeriodicReview
├── ReviewId
├── ReviewCategory
├── TargetAggregateType
├── TargetAggregateId
├── TargetAggregateVersionAtRequest
├── TriggerReference
├── ReasonCodes
├── Status
├── RequestedAt
├── StartedAt
├── CompletedAt
├── Outcome
├── RequestedCommandIds
├── FailureState
├── RetryState
└── Version
```

## Statuses

```text
Requested
In Progress
Completed
Failed
Retry Scheduled
Cancelled
Archived
```

## Owned Invariants

- Review targets one Aggregate.
- Review does not mutate target directly.
- Review uses current target state at evaluation.
- Duplicate active Review for same category and target is prevented.
- Completed Review is immutable.
- Technical Retry preserves Review identity.
- Requested Commands are traceable.
- Review cannot create infinite reaction loops.
- Review outcome does not become target domain truth without target Command.
- Terminal target may be skipped with explicit outcome.

## Commands

```text
RequestPeriodicReview
StartPeriodicReview
CompletePeriodicReview
FailPeriodicReview
RetryPeriodicReview
CancelPeriodicReview
ArchivePeriodicReview
```

## Events

```text
PeriodicReviewRequested
PeriodicReviewStarted
PeriodicReviewCompleted
PeriodicReviewFailed
PeriodicReviewRetryScheduled
PeriodicReviewCancelled
PeriodicReviewArchived
```

## PeriodicReviewCycle

Review Cycle should not be one Aggregate containing all discovered items.

Recommended implementation concept:

```text
PeriodicReviewCycle
├── CycleId
├── Category
├── Scope
├── Cursor
├── Counters
├── StartedAt
├── CompletedAt
└── operational status
```

Cycle may be modeled as:

- lightweight Aggregate;
- job execution record;
- workflow instance.

It must not own target domain decisions.

---

# DomainIntegrityIssue Aggregate

## Aggregate Root

DomainIntegrityIssue

## Responsibility

Tracks a detected violation or suspected inconsistency requiring diagnosis and resolution.

## Owned State

```text
DomainIntegrityIssue
├── IntegrityIssueId
├── IssueType
├── Severity
├── AffectedAggregateReferences
├── EvidenceReferences
├── Status
├── DetectedAt
├── DetectedBy
├── ReviewReference
├── ResolutionRecord
├── CreatedAt
├── UpdatedAt
└── Version
```

## Statuses

```text
Detected
Under Review
Resolution Planned
Resolved
Accepted Risk
False Positive
Archived
```

## Owned Invariants

- Issue references at least one affected object.
- Resolution does not directly rewrite affected Aggregate.
- Repair uses approved domain commands or migration.
- False Positive requires reason.
- Resolved issue references resolution evidence.
- Security-related issue access is restricted.
- Issue does not expose sensitive payload unnecessarily.
- Accepted Risk requires authorized Actor.
- Historical issue cannot be deleted as if it never existed.

## Commands

```text
CreateDomainIntegrityIssue
RequestDomainIntegrityReview
StartDomainIntegrityReview
PlanDomainIntegrityResolution
ResolveDomainIntegrityIssue
MarkDomainIntegrityIssueFalsePositive
AcceptDomainIntegrityRisk
ArchiveDomainIntegrityIssue
```

## Events

```text
DomainIntegrityIssueDetected
DomainIntegrityReviewRequested
DomainIntegrityReviewStarted
DomainIntegrityResolutionPlanned
DomainIntegrityIssueResolved
DomainIntegrityIssueMarkedFalsePositive
DomainIntegrityRiskAccepted
DomainIntegrityIssueArchived
```

---

# Evidence Ownership

Evidence является сложной cross-domain концепцией.

Не следует создавать один гигантский Evidence Aggregate, который владеет всеми фактами.

Рекомендуемая модель:

```text
Source Aggregate
    |
    +--> creates source fact

Evidence Reference
    |
    +--> points to source fact

Evaluation Aggregate
    |
    +--> decides whether reference is usable
```

Пример:

```text
HomeworkReviewed
      |
      v
EvidenceReference
      |
      v
Goal Completion Evaluation
```

---

# Evidence Reference

Канонический Value Object:

```text
EvidenceReference
├── EvidenceType
├── SourceDomain
├── SourceEntityType
├── SourceEntityId
├── SourceEntityVersion
├── SourceEventId
├── StudentId
├── Scope
├── OccurredAt
├── ValidityReference
└── PrivacyClassification
```

---

# Evidence Rules

- Evidence source remains owned by source Aggregate.
- Consumer stores a reference, not mutable source object.
- Evidence can be invalidated without deleting source fact.
- Evidence validity is contextual.
- One source fact may support several evaluations.
- Evidence acceptance does not change source Aggregate.
- Notification interaction cannot become Evidence unless a separate explicit domain meaning exists.
- AI-generated inference is not equivalent to confirmed Evidence.
- Evidence freshness is evaluated by receiving Policy.
- Invalidated Evidence remains auditable.

---

# Aggregate Command Routing Matrix

| Command Group | Target Aggregate |
| --- | --- |
| Student profile commands | StudentLearningProfile |
| Teacher assignment commands | TeacherAssignment |
| Lesson commands | Lesson |
| Homework commands | HomeworkAssignment |
| Progress commands | ProgressRecord |
| Goal commands | Goal |
| Achievement definition commands | AchievementDefinition |
| Achievement lifecycle commands | AchievementAward |
| Student repertoire commands | StudentSong |
| Song readiness commands | SongReadiness |
| Concert lifecycle commands | Concert |
| Concert participation commands | ConcertParticipation |
| Homework reminder commands | HomeworkReminderPlan |
| Notification intent commands | NotificationIntent |
| Delivery commands | NotificationDelivery |
| Periodic review commands | PeriodicReview |
| Integrity commands | DomainIntegrityIssue |

---

# Aggregate Event Production Matrix

| Aggregate | Main Events |
| --- | --- |
| StudentLearningProfile | StudentCreated, StudentTimezoneChanged, StudentLearningPauseStarted |
| TeacherAssignment | TeacherAssignedToStudent, TeacherReassigned |
| Lesson | LessonScheduled, LessonCompleted, LessonCancelled |
| HomeworkAssignment | HomeworkAssigned, HomeworkSubmitted, HomeworkExpired, HomeworkCompleted |
| ProgressRecord | ProgressEvidenceRecorded, ProgressUpdated |
| Goal | GoalCreated, GoalCompleted, GoalReopened |
| AchievementDefinition | AchievementDefinitionPublished |
| AchievementAward | AchievementAwarded, AchievementRevoked |
| StudentSong | SongAddedToStudentRepertoire, StudentSongVersionChanged |
| SongReadiness | SongReadinessEvaluated, SongReadinessChanged |
| Concert | ConcertCreated, ConcertRequirementsPublished, ConcertCompleted |
| ConcertParticipation | ConcertEligibilityEvaluated, ConcertParticipationApproved, ConcertPerformanceCompleted |
| HomeworkReminderPlan | HomeworkReminderScheduled, HomeworkReminderSuppressed |
| NotificationIntent | NotificationIntentCreated, NotificationSuppressed |
| NotificationDelivery | NotificationDelivered, NotificationDeliveryFailed |
| PeriodicReview | PeriodicReviewRequested, PeriodicReviewCompleted |
| DomainIntegrityIssue | DomainIntegrityIssueDetected, DomainIntegrityIssueResolved |

---

# Multi-Aggregate Process Examples

## Lesson Completion Flow

```text
CompleteLesson
      |
      v
Lesson Aggregate
      |
      v
LessonCompleted
      |
      +--> EvaluateProgressUpdate
      |         |
      |         v
      |    ProgressRecord
      |
      +--> RequestGoalReview
      |         |
      |         v
      |       Goal
      |
      +--> AssignHomework
                |
                v
        HomeworkAssignment
```

No single transaction spans all Aggregate.

---

## Homework Submission Flow

```text
SubmitHomework
      |
      v
HomeworkAssignment
      |
      v
HomeworkSubmitted
      |
      +--> Cancel or suppress Reminder
      |         |
      |         v
      |  HomeworkReminderPlan
      |
      +--> Create Teacher Notification Intent
                |
                v
        NotificationIntent
```

---

## Goal Completion Flow

```text
EvaluateGoalCompletion
      |
      v
Goal Completion Policy
      |
      v
CompleteGoal
      |
      v
Goal Aggregate
      |
      v
GoalCompleted
      |
      +--> EvaluateAchievementEligibility
      |
      +--> CreateNotificationIntent
```

---

## Concert Eligibility Flow

```text
RequestConcertEligibilityEvaluation
      |
      v
Load:
- Concert Requirements
- ConcertParticipation
- SongReadiness
- Consent
- relevant Evidence
      |
      v
Concert Eligibility Policy
      |
      v
MarkConcertParticipationEligible
      |
      v
ConcertParticipation Aggregate
```

Concert and SongReadiness are not changed in the same transaction.

---

## Notification Delivery Flow

```text
CreateNotificationIntent
      |
      v
NotificationIntent
      |
      v
NotificationIntentApproved
      |
      v
ScheduleNotification
      |
      v
NotificationDelivery
      |
      v
NotificationDelivered
```

---

# Process Manager Candidates

Процессы, вероятно требующие отдельного Process Manager:

## Lesson Completion Process

Coordinates:

- Lesson;
- Progress;
- Homework;
- Goal;
- Song Readiness.

## Homework Lifecycle Process

Coordinates:

- Homework Assignment;
- Reminder Plan;
- Notification Intent;
- Teacher Review.

## Goal Completion Process

Coordinates:

- Goal Evaluation;
- Goal;
- Achievement Eligibility;
- Notification.

## Concert Preparation Process

Coordinates:

- Concert;
- Participation;
- Consent;
- Song Readiness;
- Rehearsal;
- Slot;
- Notification.

## Notification Delivery Process

Coordinates:

- Intent;
- Delivery;
- provider;
- fallback;
- retries.

## Periodic Review Process

Coordinates:

- discovery;
- individual review;
- domain evaluation commands;
- retry;
- audit.

---

# Process Manager Rules

- Process Manager owns workflow state, not domain state.
- It sends commands to Aggregate.
- It reacts to events.
- It does not mutate Aggregate storage directly.
- It preserves CorrelationId and CausationId.
- It is idempotent.
- It handles timeout and compensation.
- It does not reimplement Aggregate invariants.
- It stores only workflow-required references.
- It detects terminal process completion.
- It prevents event-command cycles.

---

# Aggregate Loading Rules

Command Handler should load:

- Target Aggregate;
- required immutable configuration;
- Policy Decision;
- minimal external snapshots required for validation.

Aggregate should not lazily fetch other Aggregate during mutation.

Application Service may prepare:

- Command
- Current Aggregate
- Policy Decision
- External Snapshot References
- Current Time

Then invoke deterministic domain behavior.

---

# Aggregate Persistence

Persistence model may use:

- relational state storage;
- JSON document;
- event sourcing;
- hybrid state + events.

Aggregate contract does not depend on storage format.

However, persistence must preserve:

- Aggregate Id;
- Aggregate Version;
- authoritative state;
- terminal history references;
- event atomicity;
- tenant boundary.

---

# Event Sourcing Guidance

Full event sourcing is not required for every Aggregate.

Possible candidates:

- Goal;
- Homework Assignment;
- Concert Participation;
- Achievement Award.

But state storage + Outbox may be simpler for MVP.

Decision should be based on:

- reconstruction value;
- history complexity;
- correction needs;
- query patterns;
- team maturity;
- operational cost;
- audit requirements.

Event catalog does not imply full event sourcing.

---

# Snapshot Rules

If event sourcing is used:

- Snapshot is an optimization.
- Event stream remains authoritative.
- Snapshot includes Aggregate Version.
- Invalid snapshot can be rebuilt.
- Snapshot does not remove event history.
- Snapshot privacy follows Aggregate classification.

---

# Aggregate Deletion

Domain Aggregate should not normally be physically deleted as a lifecycle action.

Use:

- Cancel
- Expire
- Archive
- Anonymize
- Retain audit reference

Physical deletion belongs to:

- retention policy;
- privacy request handling;
- legal deletion workflow;
- technical cleanup.

Deletion must not be disguised as domain cancellation.

---

# Aggregate Archival

Archive means:

- excluded from active operational queries;
- ordinary commands rejected;
- history remains accessible according to permissions;
- audit remains intact;
- retention workflow may later act.

Archive does not mean:

- delete;
- cancel;
- expire;
- complete;
- revoke.

---

# Aggregate Reopen

Reopen requires:

- terminal state that supports reopening;
- authorized Actor;
- current relevance;
- explicit reason;
- new Aggregate Version;
- preserved terminal record;
- reevaluation of dependent Aggregate.

Possible reopenable Aggregate:

- HomeworkAssignment;
- Goal;
- ConcertParticipation;
- AchievementAward through Restore semantics.

Lesson reopening should be exceptional and may instead use Correction records.

---

# Read Models and Projections

Aggregate should not be designed around every UI screen.

Read Models may combine:

- Student;
- Teacher Assignment;
- Lesson;
- Homework;
- Goal;
- Song Readiness;
- Concert Participation;
- Notifications.

Example dashboard projection:

```text
StudentLearningDashboard
├── UpcomingLessons
├── ActiveHomework
├── Goals
├── SongReadiness
├── ConcertPreparation
└── Notifications
```

This Projection is not an Aggregate and cannot be mutated directly.

---

# Query Rules

Queries:

- do not change Aggregate;
- may use denormalized projections;
- may join multiple Aggregate;
- must enforce authorization;
- may be eventually consistent;
- should expose freshness where relevant;
- must not be used as mutation bypass.

---

# Aggregate Privacy

Each Aggregate has a privacy classification.

Suggested guidance:

| Aggregate | Default Classification |
| --- | --- |
| StudentLearningProfile | Confidential |
| TeacherAssignment | Confidential |
| Lesson | Confidential |
| HomeworkAssignment | Sensitive |
| ProgressRecord | Sensitive |
| Goal | Sensitive |
| AchievementDefinition | Internal |
| AchievementAward | Confidential |
| StudentSong | Confidential |
| SongReadiness | Sensitive |
| Concert | Internal |
| ConcertParticipation | Sensitive |
| HomeworkReminderPlan | Confidential |
| NotificationIntent | Sensitive |
| NotificationDelivery | Confidential |
| PeriodicReview | Internal / Confidential |
| DomainIntegrityIssue | Highly Restricted |

Actual classification may vary by payload.

---

# Aggregate Security

Every Aggregate mutation must protect against:

- forged Actor;
- tenant mismatch;
- stale version;
- cross-student reference;
- unauthorized Teacher;
- policy decision substitution;
- hidden generic status mutation;
- replay attack;
- duplicate command;
- private data leakage;
- attachment reference substitution;
- AI authority escalation;
- migration misuse;
- archival bypass;
- raw database mutation.

---

# Aggregate Audit

For each mutation preserve:

- Aggregate Type;
- Aggregate Id;
- previous Version;
- new Version;
- Command Id;
- Actor;
- Policy Decision;
- Reason Codes;
- produced Events;
- occurred time;
- recorded time;
- tenant;
- correlation;
- causation;
- authorization outcome;
- privacy classification.

Audit should not require exposing full mutable Aggregate snapshots to all operators.

---

# Aggregate Failure Modes

## Aggregate not found

- Result: Rejected
- Reason Code: AGGREGATE_NOT_FOUND

---

## Version conflict

- Result: Conflict
- Reason Code: AGGREGATE_VERSION_CONFLICT

---

## Invalid lifecycle transition

- Result: Rejected
- Reason Code: AGGREGATE_TRANSITION_NOT_ALLOWED

---

## Missing Policy Decision

- Result: Rejected
- Reason Code: AGGREGATE_POLICY_DECISION_REQUIRED

---

## Cross-tenant reference

- Result: Rejected
- Reason Code: AGGREGATE_TENANT_MISMATCH

---

## Invariant violation

- Result: Rejected
- Reason Code: AGGREGATE_INVARIANT_VIOLATION

---

## Duplicate command

- Result: Already Processed

---

## Referenced aggregate stale

- Result: Deferred (или новая Policy evaluation)
- Reason Code: AGGREGATE_REFERENCE_STATE_STALE

---

## Invalid policy snapshot

- Result: Rejected
- Reason Code: AGGREGATE_POLICY_SNAPSHOT_INVALID

---

## Internal entity not found

- Result: Rejected
- Reason Code: AGGREGATE_ENTITY_NOT_FOUND

---

# Aggregate Reason Codes

Общие Reason Codes:

```text
AGGREGATE_NOT_FOUND
AGGREGATE_VERSION_CONFLICT
AGGREGATE_TRANSITION_NOT_ALLOWED
AGGREGATE_INVARIANT_VIOLATION
AGGREGATE_POLICY_DECISION_REQUIRED
AGGREGATE_POLICY_SNAPSHOT_INVALID
AGGREGATE_REFERENCE_STATE_STALE
AGGREGATE_ENTITY_NOT_FOUND
AGGREGATE_TENANT_MISMATCH
AGGREGATE_ALREADY_ARCHIVED
AGGREGATE_TERMINAL_STATE
AGGREGATE_REOPEN_NOT_ALLOWED
AGGREGATE_REOPEN_REASON_REQUIRED
AGGREGATE_DUPLICATE_ENTITY
AGGREGATE_REFERENCE_INVALID
AGGREGATE_COMMAND_ALREADY_PROCESSED
```

Domain-specific Reason Codes remain in corresponding policy.

---

# Test Requirements

## Root Boundary Tests

- only Root mutates internal Entities;
- repository returns Aggregate Root;
- internal Entity cannot be saved independently;
- external Aggregate stored only as reference;
- generic status update is unavailable;
- mutation emits expected Events.

---

## Invariant Tests

- invalid transition is rejected;
- valid transition preserves all invariants;
- terminal state history is retained;
- Reopen creates new Version;
- stale Policy Decision is rejected;
- cross-tenant reference is rejected;
- duplicate internal Entity is rejected.

---

## Version Tests

- successful mutation increments Version;
- rejected mutation does not increment Version;
- No Change does not increment Version unless explicit audit mutation;
- ExpectedVersion mismatch returns Conflict;
- concurrent commands cannot overwrite each other;
- technical retry preserves command result.

---

## Transaction Tests

- Aggregate state and Outbox commit atomically;
- rollback does not persist Event;
- multiple Events from one mutation share resulting Aggregate Version rules;
- repository cannot expose partially persisted Aggregate;
- retry after commit returns prior result;
- event publication failure does not roll back committed domain state.

---

## Cross-Aggregate Tests

- Lesson does not mutate Progress directly;
- Homework does not cancel Reminder records directly;
- Goal does not Award Achievement directly;
- Concert does not update Song Readiness;
- Notification Delivery does not change Homework;
- Periodic Review does not mutate target Aggregate;
- cross-aggregate flow preserves CausationId.

---

## Homework Tests

- Submitted Homework cannot be expired by stale command;
- Completed Homework cannot become Overdue;
- Replaced Homework has successor;
- Reopen preserves Expiration history;
- Review references valid Submission;
- Correction refers to valid Review;
- Reminder delivery does not change Homework state.

---

## Goal Tests

- Completion requires Decision;
- stale Evidence cannot complete Goal when freshness required;
- duplicate Completion is idempotent;
- Reopen preserves prior Completion;
- Cancelled Goal rejects completion;
- Homework Expiration does not directly cancel Goal.

---

## Song Readiness Tests

- identity includes Song Version;
- version change does not transfer Ready;
- AI cannot finalize evaluation;
- stale readiness can be marked Stale;
- Ready does not approve Concert Participation;
- prior evaluations remain auditable.

---

## Concert Participation Tests

- Eligibility and Approval remain separate;
- Slot requires Approval;
- Withdrawal cancels active Slot according to rules;
- Song Version change invalidates Eligibility;
- Requirements Version is recorded;
- Not Eligible cannot become Approved without reevaluation;
- Performance completion does not create Achievement directly.

---

## Reminder Tests

- one plan is scoped to one Homework Version;
- Submitted Homework suppresses due Reminder after revalidation;
- maximum count is enforced;
- duplicate Reminder is prevented;
- Delivery does not equal Reading;
- old plan cannot schedule after Homework replacement;
- recalculation preserves history.

---

## Notification Tests

- Intent and Delivery are separate;
- cancelled Intent cannot create Delivery;
- expired Delivery cannot Retry;
- duplicate callback is idempotent;
- Delivered cannot regress to Failed;
- Opened does not perform source domain action;
- privacy level limits channel.

---

## Periodic Review Tests

- duplicate active Review is prevented;
- Review does not mutate target;
- stale target version triggers reevaluation;
- retry preserves Review Id;
- Completed Review is immutable;
- requested Commands are recorded;
- cycle protection works.

---

## Privacy Tests

- Aggregate does not embed unnecessary foreign snapshots;
- private notes are references;
- read models enforce authorization;
- event payload is minimized;
- archived sensitive Aggregate remains access-controlled;
- wrong Student reference is rejected;
- raw Aggregate is not exposed to analytics by default.

---

## AI Tests

- AI cannot mutate Aggregate directly;
- AI proposal requires authoritative Command;
- AI cannot supply fake Policy Decision;
- AI cannot bypass versioning;
- AI cannot alter immutable history;
- AI-generated Evidence remains unconfirmed until accepted;
- approved Actor remains distinct from AI.

---

# Non-Goals

Aggregate Catalog не определяет:

- database tables;
- repository implementation;
- ORM;
- serialization;
- HTTP API;
- service boundaries;
- deployment topology;
- microservice count;
- programming language;
- exact event store;
- broker topics;
- UI state;
- reporting schema;
- analytics warehouse;
- full identity model;
- CRM Aggregate;
- billing;
- payments;
- payroll;
- marketing campaigns;
- legal retention periods;
- file storage architecture.

---

# Open Questions

Необходимо определить:

- будет ли StudentLearningProfile отдельным Aggregate или projection enrollment data;
- нужен ли отдельный LearningPause Aggregate;
- может ли Student иметь несколько одновременных pause scopes;
- нужен ли отдельный TeacherAssignment Aggregate в MVP;
- поддерживаются ли substitute Teachers;
- является ли Attendance частью Lesson;
- нужен ли отдельный Attendance Aggregate для групповых Lessons;
- поддерживаются ли Group Lessons;
- как моделировать Group Homework;
- одно Homework Assignment на группу или отдельный Aggregate на Student;
- может ли Homework иметь несколько активных Submissions;
- нужно ли выделять Submission в отдельный Aggregate;
- насколько большим может стать Submission history;
- нужно ли выделять Homework Review;
- нужен ли отдельный Evidence Registry;
- является ли ProgressRecord одним Aggregate на Student или отдельным на Scope;
- какие Progress scopes допустимы;
- нужен ли общий curriculum model;
- является ли Goal Criterion Value Object или отдельной Definition;
- нужна ли Goal Definition;
- может ли Goal иметь несколько Students;
- как моделировать group Goal;
- может ли Goal Completion быть частичной;
- нужен ли отдельный Goal Review Aggregate;
- где хранится Achievement Eligibility Evaluation;
- является ли Evaluation отдельным Aggregate;
- может ли Award повторяться;
- как моделировать Achievement series;
- где находится Song Catalog;
- является ли Song Version отдельным Aggregate;
- как хранить arrangements;
- может ли StudentSong иметь несколько активных Song Versions;
- нужен ли отдельный Performance Preparation Aggregate;
- является ли Song Readiness одной записью или history Aggregate;
- нужно ли хранить Evaluation как отдельный Aggregate;
- как определяется Readiness ValidUntil;
- может ли один Student иметь несколько Participation в одном Concert;
- поддерживаются ли duet и group performances;
- как моделировать Participation с несколькими Students;
- нужен ли отдельный Ensemble Aggregate;
- принадлежит ли Slot Participation или Program;
- нужен ли ConcertProgram отдельный Aggregate;
- как обеспечивать уникальность Slot;
- где проверяется capacity;
- нужен ли reservation mechanism;
- является ли Rehearsal отдельным Lesson или Concert entity;
- нужно ли выделять Consent в отдельный Aggregate;
- может ли Guardian grant Consent;
- как хранить Consent Version;
- один Reminder Plan на Homework или на Recipient;
- как моделировать multiple channels;
- нужно ли отделить scheduled Reminder от Plan;
- сколько Reminder можно хранить внутри Plan;
- нужно ли выделить каждый Reminder в Aggregate;
- должен ли Notification Intent иметь несколько Recipients;
- лучше ли создавать отдельный Intent на Recipient;
- может ли Delivery менять Channel или нужен новый Delivery Aggregate;
- где хранить provider attempts;
- сколько attempts допускается внутри Root;
- нужен ли отдельный DeliveryAttempt store;
- является ли PeriodicReview Aggregate или workflow record;
- нужен ли ReviewCycle Aggregate;
- как хранить Review discovery cursor;
- когда создавать DomainIntegrityIssue;
- кто может Accept Risk;
- какие Aggregate будут event-sourced;
- какие будут state + Outbox;
- нужны ли snapshots;
- какой transaction isolation использовать;
- как реализовать Aggregate repositories;
- как проверять uniqueness across Aggregate;
- нужен ли reservation Aggregate;
- как выполнять bulk mutations;
- нужен ли Unit of Work;
- разрешены ли транзакции между Aggregate в одном bounded context;
- когда eventual consistency недостаточна;
- какие compensation commands нужны;
- нужен ли Saga framework;
- как хранить process manager state;
- как предотвращать process loops;
- какой maximum aggregate size;
- как архивировать internal collections;
- как переносить history в audit storage;
- как анонимизировать Aggregate;
- какие fields являются PII;
- как реализовать tenant isolation;
- нужно ли encryption per Aggregate;
- как ограничить repository access;
- может ли analytics читать primary Aggregate store;
- какие projections нужны Student;
- какие projections нужны Teacher;
- какие projections нужны Owner;
- как показывать projection freshness;
- как восстанавливать projections;
- нужно ли CQRS разделение;
- можно ли mobile client работать offline;
- как применять offline commands;
- какие Aggregate допускают mergeable updates;
- как разрешать version conflict в UI;
- как импортировать legacy lessons and homework;
- создаются ли Aggregate для historical data;
- как маркировать migrated Aggregate;
- как сохранить provenance;
- как провести CRM enrollment boundary;
- какие CRM references допускаются;
- как предотвратить CRM status leakage в learning Aggregate.

---

# Таблица вывода редакции M-0003/B.0

| Изменённая часть | Класс | Источник | Версия источника |
|------------------|-------|----------|------------------|
| T7 и граница редакции | B | PD-0030, PD-0032; разрешение Product Owner / Education Lead 2026-08-01 | — |
| Identity и access boundaries | A | PD-0030, пп. 1–6 | — |
| Invitation после первого ориентира | A | PD-0030, п. 11 | — |
| Student без access-state | A | PD-0030, п. 3 | — |
| Persisted account/invitation statuses | A | производное отображение реализации B.0, не решение класса B | — |

---

# History

| Version | Description |
| --- | --- |
| 2.0.0 | T7 · DRAFT · M-0003/B.0: Person, SchoolMembership, Account и Invitation отделены от StudentLearningProfile; цифровой access-state удалён из Student. Требуется P7 Education Lead + Technical Lead. |
| 1.0.0 | Определен канонический набор Aggregate Roots, consistency boundaries, invariants, ownership, command routing, event production и cross-aggregate process rules Belcanto Product. |
