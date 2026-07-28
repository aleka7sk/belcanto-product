---
Status: Accepted
Version: 1.0.0
Last Updated: 2026-07-27

Document Id: VALUE_OBJECT_CATALOG

Document Type:
  - Domain Contract
  - Value Object Catalog
  - Shared Type Specification
  - Validation Standard
  - Serialization Standard
  - Privacy and Traceability Standard

Owners:
  - Domain Architecture Lead
  - Technical Lead
  - Security Owner

Applies To:
  - Aggregate Roots
  - Aggregate Entities
  - Domain Commands
  - Domain Events
  - Policies
  - Decisions
  - Process Managers
  - Projections
  - Repositories
  - Integrations
  - Audit
  - AI-Assisted Operations

Related Directories:
  - ../architecture/
  - ../aggregates/
  - ../commands/
  - ../events/
  - ../policies/
  - ../entities/
  - ../services/
  - ../processes/

Related Documents:
  - ../architecture/000-domain-model-rules.md
  - ../aggregates/000-aggregate-catalog.md
  - ../commands/000-domain-command-catalog.md
  - ../events/000-domain-event-catalog.md
  - ../policies/000-domain-policy-overview.md
---

# Value Object Catalog

> Value Object Catalog определяет канонические общие типы доменной модели Belcanto Product.
>
> Value Object выражает доменное значение, не имеет самостоятельной identity, является immutable и считается равным другому объекту по значению.
>
> Этот каталог не должен превращаться в сборник всех полей системы. В него входят только общие типы, значение и правила которых должны быть одинаковыми во всех bounded contexts.

---

# Purpose

Commands, Events, Aggregates, Policies и Decisions используют повторяющиеся понятия:

- идентификатор;
- версия;
- Actor;
- Tenant;
- ссылка на Aggregate;
- Evidence;
- Policy Decision;
- Reason Code;
- момент времени;
- локальная дата;
- временное окно;
- Due Date;
- Grace Period;
- Timezone;
- Idempotency Key;
- Correlation;
- Privacy Level;
- Permission Scope;
- Validation Error.

Если эти понятия представлены обычными primitive values, возникают ошибки:

- StudentId случайно используется как TeacherId
- AggregateVersion смешивается с EventVersion
- Local Date трактуется как UTC timestamp
- Due Date теряет Timezone
- ActorId из payload принимается за authenticated Actor
- Reason Code хранится произвольным текстом
- Evidence не содержит provenance
- Privacy Level не ограничивает публикацию Event
- Idempotency Key повторно используется с другим payload

Канонические Value Objects должны:

- сохранять единый доменный смысл;
- ограничивать недопустимые состояния;
- предотвращать primitive obsession;
- обеспечивать type safety;
- фиксировать validation;
- сохранять traceability;
- обеспечивать единое serialization behavior;
- уменьшать дублирование контрактов.

---

# Core Value Object Rules

## VO-001: Value Object has no independent identity

Value Object не существует как отдельный бизнес-объект.

Например:

```text
Timezone("Asia/Almaty")
```

не имеет собственного lifecycle и repository.

## VO-002: Value Object is immutable

После создания его значение не меняется.

Изменение:

```text
oldDueDate -> newDueDate
```

означает создание нового DueDate.

## VO-003: Equality is based on semantic value

Два Value Object равны, если равны их нормализованные значения и semantics.

## VO-004: Invalid Value Object must not exist

Value Object должен валидироваться при создании.

Нельзя создать:

- пустой required identifier;
- несуществующий Timezone;
- TimeWindow с end раньше start;
- отрицательную Duration;
- неизвестный Privacy Level;
- ExpectedVersion со значением меньше нуля;
- Evidence Reference без source;
- Actor без ActorType.

## VO-005: Value Object may contain behavior

Value Object может отвечать на доменные вопросы:

```text
timeWindow.Contains(moment)
gracePeriod.HasEnded(at)
version.Matches(currentVersion)
privacyLevel.Allows(channel)
deliveryWindow.NextAllowedTime(at)
```

## VO-006: Behavior must be deterministic

Value Object не обращается к:

- clock;
- database;
- network;
- repository;
- environment;
- AI;
- provider.

## VO-007: Serialization is a contract boundary

Внутренняя реализация может отличаться, но serialized representation должна быть стабильной и версионируемой.

## VO-008: Domain types should not be aliases without semantics

Простой wrapper оправдан, если он предоставляет хотя бы одно из:

- validation;
- type safety;
- normalization;
- domain behavior;
- privacy classification;
- serialization rules;
- comparison rules.

## VO-009: Shared Value Object must retain the same meaning everywhere

AggregateVersion не может означать optimistic version в одном module и schema version в другом.

## VO-010: Domain-specific Value Objects may live outside this catalog

Например:

```text
VocalRange
RepertoirePurpose
SongKey
HomeworkRequiredness
ReadinessArea
ConcertPerformanceType
```

должны определяться в соответствующем bounded context, если не являются общими для всей системы.

---

# Value Object Categories

Канонические группы:

```text
Identity
Reference
Versioning
Time
Actor and Authorization
Policy and Decision
Evidence
Reason and Validation
Privacy and Security
Traceability
Idempotency
Lifecycle
Communication
Technical Provenance
```

---

# Identity Value Objects

## Identifier

Базовая семантическая модель typed identifier.

```text
Identifier
├── Value
└── Type
```

Требования:

- immutable;
- non-empty;
- canonical format;
- case-sensitive unless contract says otherwise;
- no embedded mutable business meaning;
- globally unique within declared scope;
- never reused.

Identifiers should be opaque.

Нежелательно:

```text
student_astana_active_2026_123
```

Предпочтительно:

```text
01K1STUDENTABC0123456789XYZ
```

Business meaning хранится отдельно.

## Identifier Format

Рекомендуемый формат:

- UUID;
- ULID;
- UUIDv7;
- другой time-sortable globally unique identifier.

Формат должен быть единым внутри product platform, если нет причины для исключения.

Identifier должен сериализоваться как string.

## TenantId

Определяет tenant boundary.

```text
TenantId
├── Value
└── Namespace
```

Правила:

- обязателен для tenant-owned domain artifacts;
- immutable;
- не определяется только через AggregateId;
- проверяется при каждом mutation;
- не изменяется после создания Aggregate;
- не берется из untrusted payload без сравнения с authorization context.

Пример:

```text
belcanto_astana
```

Public display name школы не должен использоваться как TenantId.

## AggregateId

Typed identifier Aggregate Root.

```text
AggregateId<TAggregate>
```

Примеры:

```text
StudentLearningProfileId
TeacherAssignmentId
LessonId
HomeworkAssignmentId
ProgressId
GoalId
AchievementDefinitionId
AchievementAwardId
StudentSongId
SongReadinessId
ConcertId
ConcertParticipationId
HomeworkReminderPlanId
NotificationIntentId
NotificationDeliveryId
PeriodicReviewId
DomainIntegrityIssueId
```

Правила:

- типы нельзя смешивать;
- AggregateId не заменяет TenantId;
- serialization format может быть общим;
- domain type должен сохраняться в static type или explicit reference.

## EntityId

Identifier внутренней Entity.

Примеры:

```text
SubmissionId
HomeworkReviewId
BlockerId
CorrectionRequestId
AttendanceRecordId
DeliveryAttemptId
PerformanceSlotId
ConsentId
```

Правила:

- scope должен быть определен;
- EntityId может быть globally unique или unique inside Aggregate;
- внешний reference на internal Entity должен включать Aggregate reference;
- EntityId не дает права загружать Entity независимо от Aggregate.

## CommandId

Глобально уникальный идентификатор Command.

Правила:

- сохраняется при technical retry;
- не переиспользуется для другого intent;
- связывается с Command Result;
- используется в audit;
- может участвовать в causation.

## EventId

Глобально уникальный идентификатор Domain Event.

Правила:

- создается один раз;
- не меняется при повторной публикации;
- используется consumer Inbox;
- сохраняется при replay;
- не переиспользуется.

## DecisionId

Идентификатор Policy или Human Decision.

Правила:

- immutable;
- unique;
- references exact decision artifact;
- не используется для новой reevaluation;
- новая evaluation создает новый DecisionId.

## ReviewId

Идентификатор review artifact или review workflow.

Примеры:

```text
GoalReviewId
HomeworkReviewId
PeriodicReviewId
IntegrityReviewId
```

Необходимо использовать typed variants, чтобы не смешивать разные review semantics.

## EvaluationId

Идентификатор конкретного evaluation execution.

```text
EvaluationId
```

Evaluation и Decision — не всегда одно и то же.

```text
Evaluation
    |
    +--> may fail
    +--> may require human review
    +--> may produce Decision
```

## ProposalId

Идентификатор proposal, включая AI-assisted proposal.

Proposal не является Decision.

## OperationId

Идентификатор client или application operation.

Используется для связи пользовательского действия с одной или несколькими Commands.

Не заменяет CommandId.

---

# Reference Value Objects

## AggregateReference

Каноническая ссылка на Aggregate.

```text
AggregateReference
├── AggregateType
├── AggregateId
├── AggregateVersion
└── TenantId
```

Обязательные поля:

```text
AggregateType
AggregateId
```

AggregateVersion обязателен, когда reference используется как decision input или snapshot reference.

Правила:

- не содержит mutable Aggregate snapshot;
- не дает права изменить Aggregate;
- TenantId должен совпадать с текущим scope;
- versionless reference допустим для простой навигации;
- decision-critical reference должен быть versioned.

## EntityReference

Ссылка на внутреннюю Entity.

```text
EntityReference
├── AggregateReference
├── EntityType
├── EntityId
└── EntityVersion
```

Правила:

- всегда включает owning Aggregate;
- internal Entity не считается independent Aggregate;
- EntityVersion используется только если entity имеет отдельное version semantics;
- consumer не должен сохранять mutable копию Entity.

## SourceReference

Общая ссылка на источник доменного факта.

```text
SourceReference
├── SourceDomain
├── SourceType
├── SourceId
├── SourceVersion
├── SourceEventId
└── OccurredAt
```

Используется для:

- Evidence;
- Notification Intent;
- Decision;
- Audit;
- Integration provenance.

## EnrollmentReference

Ссылка на подтвержденный enrollment boundary.

```text
EnrollmentReference
├── SourceSystem
├── EnrollmentId
├── EnrollmentVersion
├── ConfirmedAt
└── TenantId
```

Правила:

- не содержит CRM lead state;
- подтверждает право создать Student Learning Profile;
- не переносит CRM lifecycle в learning domain;
- source system должен быть trusted integration.

## AttachmentReference

Ссылка на файл или attachment.

```text
AttachmentReference
├── AttachmentId
├── StorageReference
├── MediaType
├── Size
├── Checksum
├── OwnerReference
├── PrivacyLevel
├── CreatedAt
└── Status
```

Правила:

- Domain Event не содержит file bytes;
- storage URL не должен быть permanent public secret;
- attachment ownership проверяется;
- file type и size валидируются;
- checksum используется для integrity;
- deleted/quarantined attachment нельзя использовать как active Evidence;
- malware/security validation происходит до принятия attachment.

## MaterialReference

Ссылка на образовательный материал.

```text
MaterialReference
├── MaterialId
├── MaterialVersion
├── MaterialType
├── AccessScope
└── AvailabilityStatus
```

Material может быть:

- document;
- audio;
- video;
- backing track;
- lyrics;
- exercise;
- score;
- instruction.

## TemplateReference

Ссылка на immutable version шаблона.

```text
TemplateReference
├── TemplateId
├── TemplateVersion
├── Locale
├── Channel
└── Category
```

Используется для Notification rendering.

## ConfigurationReference

Ссылка на versioned configuration.

```text
ConfigurationReference
├── ConfigurationType
├── ConfigurationId
├── ConfigurationVersion
└── PublishedAt
```

Decision-critical configuration должна быть versioned.

## PolicyReference

Ссылка на Policy.

```text
PolicyReference
├── PolicyId
├── PolicyVersion
└── PolicyArtifactReference
```

Правила:

- PolicyId стабилен;
- PolicyVersion меняется при behavioral change;
- Decision хранит PolicyReference;
- deployment version не заменяет PolicyVersion.

## DecisionReference

```text
DecisionReference
├── DecisionId
├── DecisionType
├── DecisionSource
├── PolicyReference
├── DecidedAt
└── ValidUntil
```

PolicyReference может отсутствовать для Human Decision.

## NotificationReference

```text
NotificationReference
├── NotificationIntentId
├── NotificationDeliveryId
├── RecipientId
├── Channel
└── StatusAtReference
```

Не используется как Evidence образовательного прогресса.

---

# Version Value Objects

## AggregateVersion

Monotonic optimistic concurrency version.

```text
AggregateVersion
├── Value
└── Semantics: State Revision
```

Правила:

- integer >= 0;
- version 0 может означать new/non-persisted Aggregate;
- успешная mutation увеличивает version;
- rejected command не увеличивает version;
- не сравнивается с EventVersion;
- не сбрасывается после archive/reopen;
- не переиспользуется.

Behavior:

```text
Matches(expected)
Next()
IsNew()
IsBefore(other)
```

## ExpectedAggregateVersion

Ожидаемая версия target Aggregate.

```text
ExpectedAggregateVersion
├── Value
└── ConflictStrategy
```

Возможные стратегии:

```text
RejectOnMismatch
ReevaluateOnMismatch
ExplicitlyMergeable
```

Default:

```text
RejectOnMismatch
```

## EntityVersion

Версия внутренней Entity, если это требуется.

Не должна автоматически вводиться для каждой Entity.

## CommandVersion

Schema version Command contract.

```text
CommandVersion
```

Не отражает состояние Aggregate.

## EventVersion

Schema version Event contract.

```text
EventVersion
```

## PolicyVersion

Behavioral version Policy.

Изменяется, если одинаковые inputs могут привести к другому outcome.

## DefinitionVersion

Версия immutable business definition.

Примеры:

```text
AchievementDefinitionVersion
ConcertRequirementsVersion
SongVersion
HomeworkDefinitionVersion
TemplateVersion
```

## ProjectionVersion

Версия projection schema или projection handler.

Не является AggregateVersion.

---

# Time Value Objects

## Instant

Точный момент времени.

Каноническое представление:

```text
UTC timestamp with precision contract
```

Пример:

```text
2026-07-27T13:42:15Z
```

Правила:

- хранится как absolute instant;
- не содержит business timezone meaning сам по себе;
- serialization uses ISO 8601 / RFC 3339-compatible format;
- precision должна быть стабильной;
- comparison uses chronology.

## RecordedAt

Момент, когда факт записан системой.

## OccurredAt

Момент, когда доменный факт произошел.

```text
OccurredAt may be earlier than RecordedAt
```

Если точное время неизвестно, нельзя выдумывать его.

Следует использовать отдельную precision/provenance model.

## RequestedAt

Момент создания Command request.

## EffectiveAt

Момент вступления изменения в силу.

Правила:

- может быть future;
- не переписывает прошлую историю;
- semantics должны быть явными;
- delayed command должен пройти revalidation при выполнении.

## EvaluatedAt

Момент, относительно которого Policy выполняла evaluation.

Decision должна сохранять этот момент.

## LocalDate

Календарная дата без времени и timezone.

Пример:

```text
2026-07-27
```

Используется только когда business meaning действительно является датой.

## LocalTime

Локальное время без даты и timezone.

Пример:

```text
19:30
```

Не является exact instant.

## ZonedDateTime

Комбинация:

```text
Local Date
+
Local Time
+
Timezone
```

Она должна однозначно разрешаться в Instant или иметь explicit ambiguity handling.

## Timezone

Канонический IANA timezone identifier.

Пример:

```text
Asia/Almaty
```

Нельзя использовать только:

```text
UTC+5
```

как постоянную business timezone, потому что offset и timezone — разные понятия.

Правила:

- identifier должен существовать;
- normalization сохраняет canonical IANA name;
- timezone changes are explicit domain facts;
- исторические timestamps не пересчитываются скрыто.

## Duration

Продолжительность.

```text
Duration
├── Value
└── Unit-safe representation
```

Правила:

- неотрицательна, если domain type не допускает negative;
- не хранится как ambiguous integer без unit;
- calendar duration и exact duration различаются.

Пример:

```text
90 minutes
```

## TimeWindow

```text
TimeWindow
├── Start
├── End
├── Inclusivity
└── TimezoneContext
```

Инварианты:

- End > Start;
- inclusivity explicit where boundary matters;
- timezone semantics known.

Behavior:

```text
Contains(instant)
Overlaps(other)
Duration()
HasEnded(at)
HasStarted(at)
```

## DeliveryWindow

Разрешенное окно коммуникации.

```text
DeliveryWindow
├── EarliestAt
├── LatestAt
├── Timezone
├── QuietHoursPolicyReference
└── ExpirationBehavior
```

## QuietHours

```text
QuietHours
├── StartLocalTime
├── EndLocalTime
├── Timezone
├── ApplicableDays
├── Exceptions
└── PolicyReference
```

Инварианты:

- overnight interval поддерживается;
- recipient timezone явна;
- urgency exception требует authorization.

## DueDate

Каноническое значение deadline.

```text
DueDate
├── DeadlineType
├── LocalDate
├── LocalTime
├── Timezone
├── ExactInstant
└── Interpretation
```

Возможные DeadlineType:

```text
ExactInstant
EndOfLocalDay
BeforeNextLesson
RelativeToAssignment
RelativeToLesson
TeacherDefined
OpenEnded
```

OpenEnded означает отсутствие hard deadline, но не обязательно отсутствие review schedule.

## Deadline

Более общий Value Object.

```text
Deadline
├── Type
├── DueDate
├── Hardness
├── GracePeriod
├── SourceReference
└── Version
```

Hardness:

```text
Soft
Educational
Operational
Hard
```

Нельзя считать все deadline одинаковыми.

## GracePeriod

```text
GracePeriod
├── StartsAt
├── EndsAt
├── Reason
├── AllowsSubmission
├── AllowsReminder
└── PolicyReference
```

Инварианты:

- EndsAt > StartsAt;
- grace period не продлевается скрыто;
- effect должен быть explicit;
- не заменяет Due Date history.

## ExpirationWindow

```text
ExpirationWindow
├── EligibleFrom
├── MustReviewBy
├── ExpiresAt
├── PolicyReference
└── Conditions
```

Expiration eligibility и actual expiration — разные понятия.

## ValidityPeriod

```text
ValidityPeriod
├── ValidFrom
├── ValidUntil
└── InvalidatedAt
```

Используется для:

- Decision;
- Evidence;
- Consent;
- Eligibility;
- Readiness;
- Delegation;
- Assignment.

## Recurrence

```text
Recurrence
├── Frequency
├── Interval
├── Days
├── LocalTime
├── Timezone
├── StartsAt
├── EndsAt
└── MaximumOccurrences
```

Для сложной recurrence MAY использоваться standard rule representation, но domain validation остается обязательной.

---

# Actor Value Objects

## Actor

```text
Actor
├── ActorType
├── ActorId
├── ActiveRole
├── TenantId
├── DelegatedBy
├── AuthenticationContextReference
├── ImpersonationContext
└── SystemAuthorityReference
```

ActorType:

```text
Student
Teacher
Administrator
Owner
Guardian
ConcertCoordinator
System
Policy
Scheduler
Integration
Migration
```

Правила:

- Actor is not inferred solely from payload;
- ActorId required except approved anonymous system process;
- role does not automatically grant target relationship;
- Policy and Scheduler Actor have narrow scope;
- AI is not an ActorType with authoritative mutation rights.

## ActorId

Typed identity within ActorType.

Student ActorId должен быть StudentId.

Teacher ActorId должен быть TeacherId или trusted identity mapping.

## Role

```text
Role
├── RoleType
├── TenantId
├── Scope
├── EffectivePeriod
└── SourceReference
```

Role не является Permission.

## Permission

```text
Permission
├── Action
├── ResourceType
├── Scope
├── Conditions
└── Effect
```

Effect:

```text
Allow
Deny
```

Explicit deny имеет приоритет согласно Authorization Policy.

## PermissionScope

```text
PermissionScope
├── Tenant
├── AggregateTypes
├── AggregateIds
├── StudentIds
├── RelationshipTypes
├── TimeWindow
└── Restrictions
```

Scope должен быть минимальным.

## AuthorizationContext

```text
AuthorizationContext
├── AuthenticatedActor
├── ActiveRole
├── Permissions
├── TargetScope
├── Delegation
├── AuthenticationStrength
├── SessionReference
├── EvaluatedAt
└── AuthorizationDecisionReference
```

Правила:

- не хранится как источник истины дольше необходимого;
- stale authorization context не должен использоваться для future command без revalidation;
- sensitive commands могут требовать stronger authentication.

## Delegation

```text
Delegation
├── DelegationId
├── DelegatedBy
├── DelegatedTo
├── PermissionScope
├── EffectivePeriod
├── Status
├── CreatedAt
├── RevokedAt
└── ReasonReference
```

Инварианты:

- delegated authority <= original authority;
- expired/revoked delegation invalid;
- delegation is auditable;
- sensitive actions may prohibit delegation.

## ImpersonationContext

```text
ImpersonationContext
├── OriginalActor
├── ImpersonatedActor
├── Reason
├── ApprovedBy
├── StartedAt
├── ExpiresAt
└── RestrictedActions
```

Impersonation не меняет original audit identity.

## AuthenticationStrength

Каноническая классификация assurance.

```text
Anonymous
Session
Password
MultiFactor
StrongReauthentication
SystemCredential
TrustedIntegration
```

Не должна использоваться как полный security standard без отдельной specification.

---

# Decision Value Objects

## Decision

Общая структура immutable domain decision.

```text
Decision
├── DecisionId
├── DecisionType
├── DecisionSource
├── PolicyReference
├── Outcome
├── EvaluatedAt
├── InputReferences
├── InputVersions
├── EvidenceReferences
├── ReasonCodes
├── Conditions
├── BlockingConditions
├── ValidUntil
├── HumanReviewRequired
├── SupersedesDecisionId
└── Metadata
```

DecisionSource:

```text
DeterministicPolicy
AuthorizedHuman
ApprovedMigration
```

AI не является authoritative DecisionSource.

## DecisionOutcome

Общий outcome должен быть ограничен context-specific enum.

Общие meta-outcomes:

```text
Approved
Rejected
ConditionallyApproved
Deferred
NotApplicable
InsufficientEvidence
HumanReviewRequired
StaleInput
NoChangeRequired
```

Нельзя использовать один общий enum, если он разрушает domain semantics.

Например Song Readiness должна сохранять собственные статусы:

```text
Ready
ConditionallyReady
NotReady
ReviewRequired
```

## Condition

```text
Condition
├── ConditionCode
├── DescriptionReference
├── RequiredBy
├── Deadline
├── Status
├── EvidenceRequirement
└── Visibility
```

Condition допускает положительный outcome с ограничениями.

## BlockingCondition

```text
BlockingCondition
├── BlockingCode
├── DescriptionReference
├── SourceReference
├── Severity
├── ResolutionRequirement
├── ReviewRequired
└── Visibility
```

Blocking Condition препятствует положительному outcome.

## DecisionInputReference

```text
DecisionInputReference
├── ReferenceType
├── ReferenceId
├── Version
├── ObservedAt
├── Validity
└── PrivacyLevel
```

## DecisionValidity

```text
DecisionValidity
├── ValidFrom
├── ValidUntil
├── InvalidatingEvents
├── InvalidatingVersionChanges
└── ReevaluationRequired
```

## HumanReviewRequirement

```text
HumanReviewRequirement
├── ReviewType
├── RequiredRole
├── ReasonCodes
├── DueBy
├── EscalationRule
└── RestrictedActionsUntilDecision
```

---

# Evidence Value Objects

## EvidenceReference

```text
EvidenceReference
├── EvidenceId
├── EvidenceType
├── SourceReference
├── SubjectReference
├── Scope
├── OccurredAt
├── RecordedAt
├── RecordedBy
├── Validity
├── Confidence
├── PrivacyLevel
├── ConfirmationStatus
└── InvalidationReference
```

Обязательные данные:

- source;
- subject;
- type;
- occurred/recorded semantics;
- provenance.

## EvidenceType

Context-specific type.

Примеры:

```text
LessonCompletionEvidence
HomeworkReviewEvidence
TeacherAssessmentEvidence
PerformanceEvidence
AttendanceEvidence
GoalEvidence
SongReadinessEvidence
ConcertPerformanceEvidence
```

Notification Open не входит в образовательные Evidence Types.

## EvidenceScope

```text
EvidenceScope
├── StudentId
├── SkillIds
├── GoalIds
├── SongVersionIds
├── PerformanceType
├── CurriculumReference
└── ValidForPolicies
```

## EvidenceValidity

```text
EvidenceValidity
├── Status
├── ValidFrom
├── ValidUntil
├── InvalidatedAt
├── InvalidationReason
└── SupersedingEvidenceId
```

Statuses:

```text
Valid
Expired
Invalidated
Disputed
PendingConfirmation
Restricted
```

## EvidenceConfidence

```text
EvidenceConfidence
├── Level
├── Basis
└── AssessedBy
```

Possible levels:

```text
Confirmed
High
Medium
Low
Unknown
```

Confidence не заменяет confirmation.

## ConfirmationStatus

```text
Unconfirmed
SystemConfirmed
TeacherConfirmed
StudentConfirmed
MultiSourceConfirmed
Disputed
```

AI-generated inference обычно:

```text
Unconfirmed
```

## EvidenceInvalidationReference

```text
EvidenceInvalidationReference
├── InvalidationId
├── ReasonCode
├── InvalidatedBy
├── InvalidatedAt
├── SupersedingEvidenceId
└── DecisionReference
```

---

# Reason and Validation Value Objects

## ReasonCode

Стабильный machine-readable code.

```text
ReasonCode
├── Namespace
├── Code
├── Version
├── Severity
├── Visibility
└── DocumentationReference
```

Пример:

```text
HOMEWORK_BLOCKER_ACTIVE
COMMAND_VERSION_CONFLICT
SONG_READINESS_EVIDENCE_STALE
```

Правила:

- uppercase snake case;
- stable meaning;
- no dynamic values inside code;
- localized message хранится отдельно;
- internal code may be hidden from Student;
- removal requires deprecation.

## ReasonDetails

Структурированные параметры Reason Code.

```text
ReasonDetails
├── ReasonCode
├── Parameters
├── RelatedReferences
└── Visibility
```

Нельзя помещать sensitive free text без privacy classification.

## ValidationError

```text
ValidationError
├── ErrorCode
├── FieldPath
├── Parameters
├── Severity
├── Visibility
└── Source
```

Source:

```text
Schema
Domain
Authorization
Security
Integration
```

## ValidationResult

```text
ValidationResult
├── IsValid
├── Errors
├── Warnings
└── EvaluatedAt
```

IsValid должен быть derived from Errors, а не независимым противоречивым полем.

## FailureCategory

```text
Validation
Authorization
Conflict
DependencyUnavailable
Timeout
RateLimited
ProviderFailure
Serialization
Storage
Security
Unknown
```

Domain Rejection и Technical Failure должны оставаться раздельными.

## Retryability

```text
Retryability
├── Retryable
├── RetryAfter
├── MaximumAttempts
├── StopAt
└── ReasonCode
```

## Severity

Канонический уровень серьезности.

```text
Informational
Low
Medium
High
Critical
```

Context-specific use must be documented.

---

# Privacy and Security Value Objects

## PrivacyLevel

Каноническая классификация данных.

```text
Public
Internal
Confidential
Sensitive
HighlyRestricted
```

Примерная семантика:

Public

Может быть опубликовано без ограничения после approval.

Internal

Доступно сотрудникам в разрешенном organizational scope.

Confidential

Персональные или операционные данные с ограниченным доступом.

Sensitive

Образовательные оценки, private feedback, Student progress, consent-related data.

HighlyRestricted

Security incident, integrity issue, privileged investigation, особо чувствительные заметки.

### PrivacyLevel Behavior

```text
AllowsChannel(channel)
AllowsAnalyticsExport()
RequiresEncryption()
RequiresRestrictedAudit()
CanAppearInEventPayload()
```

Exact behavior определяется Security and Privacy Policy.

## DataClassification

Более детальное описание.

```text
DataClassification
├── PrivacyLevel
├── ContainsPersonalData
├── ContainsSensitiveLearningData
├── ContainsContactData
├── ContainsSecurityData
├── RetentionClass
├── AllowedPurposes
└── AllowedRecipients
```

## Visibility

```text
StudentVisible
TeacherVisible
AdministratorVisible
OwnerVisible
GuardianVisible
InternalOnly
RestrictedReviewOnly
```

Visibility не заменяет authorization.

## ConsentScope

```text
ConsentScope
├── Subject
├── Purpose
├── DataScope
├── ActionScope
├── EffectivePeriod
├── GrantedBy
└── WithdrawalBehavior
```

Consent должен быть specific.

Нельзя считать один checkbox универсальным согласием на все действия.

## RetentionClass

```text
OperationalShortTerm
OperationalLongTerm
EducationalHistory
Audit
Security
LegalHold
Temporary
```

Точные сроки определяются отдельной Retention Policy.

## SensitiveReference

Общий wrapper для ссылки на защищенный content.

```text
SensitiveReference
├── ReferenceId
├── ContentType
├── PrivacyLevel
├── AccessPolicyReference
├── IntegrityChecksum
└── RetentionClass
```

---

# Traceability Value Objects

## CorrelationId

Объединяет артефакты одного business flow.

```text
CorrelationId
```

Правила:

- сохраняется между Commands, Events, Decisions и Process Manager;
- один correlation может включать несколько Aggregate;
- не используется как security credential;
- не обязан быть AggregateId.

## CausationId

Идентификатор непосредственной причины.

```text
CausationId
├── CauseType
├── CauseId
└── CauseVersion
```

CauseType:

```text
Command
Event
Decision
Review
ScheduleOccurrence
ExternalRequest
Migration
```

## TraceId

Технический distributed tracing identifier.

Не заменяет CorrelationId.

Trace может быть короткоживущим, Correlation — долгоживущим business identifier.

## RootCauseReference

```text
RootCauseReference
├── RootType
├── RootId
├── StartedAt
└── InitiatingActor
```

Используется в длинных process chains.

## CommandReference

```text
CommandReference
├── CommandId
├── CommandType
├── CommandVersion
├── ActorReference
└── RequestedAt
```

## EventReference

```text
EventReference
├── EventId
├── EventType
├── EventVersion
├── AggregateReference
└── OccurredAt
```

## AuditReference

```text
AuditReference
├── AuditRecordId
├── AuditCategory
├── RecordedAt
└── PrivacyLevel
```

---

# Idempotency Value Objects

## IdempotencyKey

```text
IdempotencyKey
├── Scope
├── Value
├── PayloadFingerprint
├── ValidityPeriod
└── Source
```

Scope examples:

```text
PerCommandType
PerAggregate
PerActor
PerRecipientChannel
PerExternalOperation
PerScheduledOccurrence
```

Правила:

- normalized;
- non-empty;
- reused only for semantically identical operation;
- different payload with same key produces Conflict;
- sensitive raw values should be hashed or tokenized;
- retention period is explicit.

## PayloadFingerprint

Stable hash of canonical command payload.

```text
PayloadFingerprint
├── Algorithm
├── CanonicalizationVersion
└── Digest
```

Нельзя fingerprint arbitrary unstable JSON serialization без canonicalization rules.

## DeduplicationKey

Используется для предотвращения duplicate intents или events.

```text
DeduplicationKey
├── DomainScope
├── Value
├── TimeWindow
└── Version
```

Отличается от IdempotencyKey:

```text
IdempotencyKey
= same requested operation

DeduplicationKey
= semantically duplicate business artifacts
```

## OccurrenceId

Идентификатор конкретного scheduler occurrence.

```text
OccurrenceId
├── ScheduleId
├── ScheduledFor
├── Sequence
└── Value
```

Повторная доставка одного occurrence сохраняет тот же OccurrenceId.

---

# Lifecycle Value Objects

## LifecycleStatus

Общий concept, но не общий универсальный enum.

Каждый Aggregate должен иметь собственный ограниченный набор статусов.

Запрещено создавать:

```text
status: string
```

без state model.

## LifecycleTransition

```text
LifecycleTransition
├── From
├── To
├── Action
├── OccurredAt
├── Actor
├── DecisionReference
├── ReasonCodes
└── PreviousTransitionReference
```

## TerminalStateRecord

```text
TerminalStateRecord
├── TerminalStatus
├── EnteredAt
├── CommandReference
├── ActorReference
├── DecisionReference
├── ReasonCodes
└── SupersededByReopenReference
```

## ReopenRecord

```text
ReopenRecord
├── PreviousTerminalState
├── ReopenedAt
├── ReopenedBy
├── ReasonCodes
├── NewVersion
├── DecisionReference
└── DependentReviewReferences
```

## CancellationRecord

```text
CancellationRecord
├── Category
├── CancelledAt
├── CancelledBy
├── ReasonCodes
├── StudentVisibleExplanationReference
└── ReplacementReference
```

## ExpirationRecord

```text
ExpirationRecord
├── ExpirationType
├── EligibleAt
├── ExpiredAt
├── PolicyDecisionReference
├── ReasonCodes
└── ReopenAllowed
```

## ArchiveRecord

```text
ArchiveRecord
├── ArchivedAt
├── ArchivedBy
├── ReasonCode
├── RetentionClass
└── PreviousStatus
```

---

# Communication Value Objects

## Recipient

```text
Recipient
├── RecipientType
├── RecipientId
├── TenantId
├── Locale
├── Timezone
└── ContactDestinationReferences
```

RecipientType:

```text
Student
Teacher
Administrator
Owner
Guardian
ConcertCoordinator
```

Notification Intent должен обычно иметь одного Recipient.

## CommunicationChannel

```text
InApp
Push
Email
Sms
Messenger
StaffInbox
```

Поддерживаемые каналы определяются configuration.

## NotificationCategory

Примеры:

```text
Educational
Homework
Lesson
Goal
Achievement
Concert
Operational
Security
System
```

Marketing должен оставаться отдельным context и consent model.

## NotificationPriority

```text
Low
Normal
High
Critical
```

Critical требует explicit authorization и не должен использоваться для искусственной срочности.

## Urgency

```text
NonUrgent
TimeSensitive
Urgent
Emergency
```

Educational Reminder обычно не должен быть Emergency.

## RequiredAction

```text
RequiredAction
├── ActionType
├── TargetReference
├── DueBy
├── DeepLinkReference
└── CompletionEventType
```

Notification click не равен completion action.

## ChannelPreference

```text
ChannelPreference
├── Channel
├── Enabled
├── Priority
├── CategoryScope
├── TimeWindow
├── ConsentReference
└── UpdatedAt
```

## Locale

Канонический locale identifier.

Пример:

```text
ru-KZ
```

Locale не заменяет Timezone.

## MessageParameterReference

Ссылка на rendering parameters.

```text
MessageParameterReference
├── ParameterSetId
├── SchemaVersion
├── PrivacyLevel
└── SourceReferences
```

---

# Technical Provenance Value Objects

## Provenance

```text
Provenance
├── SourceType
├── SourceSystem
├── SourceReference
├── ImportedAt
├── RecordedAt
├── ActorReference
├── MigrationReference
└── Confidence
```

## SourceSystem

```text
SourceSystem
├── SystemId
├── SystemType
├── Environment
├── TrustLevel
└── IntegrationVersion
```

## MigrationReference

```text
MigrationReference
├── MigrationId
├── BatchId
├── SourceSystem
├── SourceRecordId
├── ImportedAt
├── MappingVersion
└── ValidationResultReference
```

## ExternalReference

```text
ExternalReference
├── SystemId
├── ExternalType
├── ExternalId
├── ExternalVersion
└── LastVerifiedAt
```

ExternalReference не является Domain AggregateId.

## ProcessingMetadata

```text
ProcessingMetadata
├── SourceApplication
├── ClientVersion
├── DeviceReference
├── RetryCount
├── ReplayMarker
├── ImportMarker
├── SchemaIdentifier
└── ProcessingNode
```

Processing Metadata не должна влиять на доменное решение, если это явно не предусмотрено.

---

# AI Provenance Value Objects

## AIProposalReference

```text
AIProposalReference
├── ProposalId
├── ProposalType
├── Model
├── ModelVersion
├── InstructionReference
├── SourceReferences
├── GeneratedAt
├── Confidence
├── ValidationStatus
├── ApprovedBy
└── ApprovedAt
```

## AIConfidence

```text
AIConfidence
├── Score
├── Scale
├── CalibrationReference
└── Meaning
```

Score без documented scale запрещен.

## AIValidationStatus

```text
NotValidated
SchemaValidated
DomainValidated
HumanReviewed
Approved
Rejected
Expired
```

Даже Approved AI proposal не становится историческим фактом без выполнения authoritative Command.

---

# Typed Identifier Catalog

Канонические identifier types:

```text
TenantId

StudentId
TeacherId
AdministratorId
OwnerId
GuardianId

StudentLearningProfileId
TeacherAssignmentId
LessonId
HomeworkAssignmentId
SubmissionId
HomeworkReviewId
BlockerId
CorrectionRequestId

ProgressId
ProgressEvidenceId

GoalId
GoalReviewId

AchievementDefinitionId
AchievementAwardId

SongId
SongVersionId
StudentSongId
SongReadinessId
SongReadinessEvidenceId
SongReadinessEvaluationId

ConcertId
ConcertRequirementsId
ConcertParticipationId
ConsentRequestId
ConsentId
PerformanceSlotId
ConcertEligibilityEvaluationId

ReminderPlanId
ReminderId

NotificationIntentId
NotificationDeliveryId
NotificationBundleId
DeliveryAttemptId

PeriodicReviewId
PeriodicReviewCycleId
IntegrityIssueId
IntegrityReviewId

CommandId
EventId
DecisionId
EvaluationId
ProposalId
OperationId
OccurrenceId
AuditRecordId

AttachmentId
MaterialId
TemplateId
ConfigurationId
DelegationId
MigrationId
BatchId
```

Каждый typed identifier должен иметь:

- одинаковый primitive storage format;
- отдельный domain type;
- canonical parser;
- canonical serializer;
- empty validation;
- equality;
- safe log representation.

---

# Canonical Serialization Rules

## JSON Property Naming

Рекомендуемый внешний формат:

```text
camelCase
```

Пример:

```json
{
  "aggregateType": "HomeworkAssignment",
  "aggregateId": "01K1HOMEWORK123",
  "aggregateVersion": 7
}
```

## Null Semantics

null, omitted field и empty value не должны использоваться взаимозаменяемо без contract.

Пример:

```text
validUntil omitted
```

может означать бессрочное значение.

```text
validUntil: null
```

может быть запрещено, если semantics неоднозначна.

## Enum Serialization

Enums сериализуются стабильными canonical strings.

Пример:

```text
ConditionallyReady
```

Не следует сериализовать случайный numeric ordinal.

## Date and Time Serialization

```text
Instant      -> RFC 3339 UTC
LocalDate    -> YYYY-MM-DD
LocalTime    -> HH:mm[:ss]
Timezone     -> IANA identifier
Duration     -> explicit standard format or structured unit
```

## Identifier Serialization

Identifier сериализуется как opaque string.

## Version Serialization

Version сериализуется как non-negative integer.

## Reason Code Serialization

```text
UPPER_SNAKE_CASE
```

## Backward Compatibility

Добавление optional field может быть backward compatible.

Изменение semantics existing field — несовместимо даже при сохранении имени.

## Normalization Rules

Normalization выполняется при создании Value Object.

Примеры:

- trimming запрещенных пробелов;
- canonical timezone;
- canonical locale;
- uppercase Reason Code;
- normalized identifier representation;
- canonical channel name.

Normalization не должна:

- скрыто изменять business meaning;
- исправлять invalid input без trace;
- менять user-authored content;
- превращать unknown в default.

---

# Equality Rules

## Identifier Equality

По exact canonical value и type.

```text
StudentId("123") != TeacherId("123")
```

## Reference Equality

По type, id и tenant.

Versioned reference с другой version не равна exact snapshot reference.

## Time Equality

Instant сравнивается по absolute moment.

ZonedDateTime может иметь:

- instant equality;
- representation equality.

Контракт должен определить нужный тип сравнения.

## Decision Equality

Decision сравнивается по DecisionId, а не по совпадению outcome.

Decision остается artifact с identity-like reference, хотя сама структура может быть immutable.

---

# Validation Rules

Каждый Value Object должен документировать:

```text
Required fields
Allowed formats
Allowed values
Normalization
Boundary values
Cross-field invariants
Serialization
Privacy
Equality
Failure codes
```

---

# Canonical Validation Reason Codes

```text
VALUE_REQUIRED
VALUE_FORMAT_INVALID
VALUE_OUT_OF_RANGE
VALUE_NOT_SUPPORTED
VALUE_CONFLICT
IDENTIFIER_REQUIRED
IDENTIFIER_FORMAT_INVALID
IDENTIFIER_TYPE_MISMATCH
REFERENCE_INVALID
REFERENCE_TENANT_MISMATCH
VERSION_INVALID
VERSION_CONFLICT
TIMEZONE_INVALID
TIME_WINDOW_INVALID
TIME_IN_PAST_NOT_ALLOWED
TIME_IN_FUTURE_NOT_ALLOWED
DEADLINE_INVALID
DURATION_INVALID
PRIVACY_LEVEL_INVALID
ACTOR_INVALID
AUTHORIZATION_CONTEXT_INVALID
EVIDENCE_REFERENCE_INVALID
EVIDENCE_PROVENANCE_REQUIRED
DECISION_REFERENCE_INVALID
REASON_CODE_INVALID
IDEMPOTENCY_KEY_INVALID
IDEMPOTENCY_PAYLOAD_CONFLICT
```

Domain-specific validation codes remain local.

---

# Persistence Rules

Value Object MAY persist as:

- one scalar column;
- several columns;
- composite type;
- JSON structure.

Persistence choice must not destroy semantics.

Например, DueDate cannot be reduced to one timestamp if its meaning includes:

- timezone;
- deadline type;
- local-day semantics.

## Database Constraints

Database constraints SHOULD reinforce domain validation for:

- non-null identifier;
- positive version;
- unique idempotency scope;
- valid time ordering where feasible;
- tenant reference;
- unique typed identity;
- immutable audit references.

Database constraint does not replace domain validation.

---

# API Boundary Rules

API may use transport DTO.

Transport DTO maps into Value Object through explicit validation.

Нельзя передавать raw API DTO directly into Aggregate.

```text
Request DTO
    |
    v
Validation and Mapping
    |
    v
Domain Value Objects
    |
    v
Command
```

---

# Integration Mapping Rules

External values must be mapped.

Пример:

```text
external status "done"
```

не должен автоматически стать:

```text
HomeworkStatus.Completed
```

Нужен explicit integration mapping and validation.

---

# Privacy Rules

Value Object должен сохранять privacy semantics.

Примеры:

- SensitiveReference нельзя безопасно логировать полностью;
- Actor не должен включать unnecessary profile;
- AttachmentReference не должен публиковать permanent URL;
- EvidenceReference должен ограничивать private source;
- ReasonDetails не должны содержать raw sensitive explanation.

---

# Logging Rules

Каждый Value Object должен определять safe representation.

Пример:

```text
StudentId: visible in authorized internal logs
Contact destination: masked
Attachment storage reference: hidden
Idempotency raw value: hashed
SensitiveReference: identifier only
Authorization token: never logged
```

---

# AI Rules

AI может:

- предложить значения Value Object;
- выявить missing fields;
- классифицировать draft Reason Code;
- предложить Evidence type;
- предложить TimeWindow;
- обнаружить format error.

AI не может:

- создавать authenticated Actor;
- подтверждать Consent;
- назначать authoritative DecisionId;
- подтверждать Evidence provenance;
- менять Privacy Level на менее строгий;
- генерировать historical OccurredAt как факт;
- создавать fake PolicyReference;
- обходить validation.

---

# Testing Requirements

## Construction Tests

- valid value создается;
- invalid value отклоняется;
- required field проверяется;
- boundary values проверяются;
- normalization deterministic;
- serialization round-trip сохраняет semantics.

## Equality Tests

- одинаковые значения равны;
- разные значения не равны;
- разные typed identifiers не равны;
- versioned references сравниваются корректно;
- timezone representation normalized.

## Immutability Tests

- public mutation unavailable;
- collections defensively copied;
- serialized input cannot mutate existing object;
- nested Value Objects immutable.

## Time Tests

- Instant UTC normalization;
- LocalDate preserves date;
- TimeWindow rejects invalid order;
- overnight Quiet Hours supported;
- timezone conversion deterministic;
- DST ambiguity handled;
- Due Date boundary exact;
- Grace Period boundaries explicit.

## Reference Tests

- AggregateReference type required;
- tenant mismatch rejected;
- internal Entity Reference includes Aggregate;
- version-required context rejects versionless reference;
- deleted or unavailable attachment rejected where required.

## Authorization Tests

- Actor type matches ActorId;
- delegation scope limited;
- expired delegation rejected;
- impersonation retains original Actor;
- Policy Actor cannot escalate permissions.

## Evidence Tests

- source provenance required;
- subject required;
- invalidated Evidence not accepted as valid;
- AI inference remains unconfirmed;
- Notification Open cannot become learning Evidence;
- cross-student Evidence rejected.

## Decision Tests

- PolicyReference required for deterministic policy;
- input versions preserved;
- expired Decision rejected;
- override preserves original Decision;
- AI cannot be DecisionSource.

## Idempotency Tests

- same key and same fingerprint accepted;
- same key and different fingerprint causes Conflict;
- scope included in comparison;
- expired retention behavior explicit;
- raw sensitive key not exposed in logs.

## Privacy Tests

- channel restrictions applied;
- sensitive log representation masked;
- public serialization excludes protected fields;
- restricted references cannot be exposed in student response;
- privacy level cannot be silently lowered.

## Compatibility Tests

- old serialized form still readable while supported;
- new optional field does not break old consumer;
- enum rename handled through versioning;
- semantics change requires new contract version.

---

# Anti-Patterns

## Primitive Identifier Everywhere

```text
studentId string
teacherId string
lessonId string
```

без typed validation.

## Generic Reference

```text
type: string
id: string
```

без tenant, version и allowed type constraints.

## Timestamp for Everything

Один timestamp используется одновременно как:

- local due date;
- occurred time;
- recorded time;
- scheduled local time.

## Boolean Consent

```text
consent = true
```

без scope, grantor, time и withdrawal semantics.

## String Reason

```text
reason: "not ready"
```

без Reason Code и structured details.

## Unversioned Decision Input

Decision опирается на GoalId, но не сохраняет GoalVersion.

## Evidence as Free Text

Progress обновляется на основании arbitrary note без provenance.

## Shared Universal Status

Один enum используется для Lesson, Homework, Goal и Notification.

## Mutable Value Object

DueDate изменяется через setter и historical state теряется.

## Timezone as Offset

```text
UTC+5
```

используется вместо Asia/Almaty.

## Privacy as Comment

Чувствительность данных описана только в Markdown, но не отражена в contracts.

## AI as Provenance

```text
RecordedBy: AI
ConfirmationStatus: Confirmed
```

без Human или Policy validation.

---

# Definition of Ready for a Value Object

Value Object готов к реализации, если определены:

- Name.
- Domain meaning.
- Fields.
- Required fields.
- Invariants.
- Normalization.
- Equality.
- Serialization.
- Validation errors.
- Privacy classification.
- Safe logging.
- Versioning behavior.
- Usage contexts.
- Non-goals.
- Tests.

---

# Canonical Compliance Checklist

- [ ] Value Object действительно выражает значение, а не Entity.
- [ ] Нет самостоятельного lifecycle.
- [ ] Object immutable.
- [ ] Equality определена.
- [ ] Invalid state невозможно создать.
- [ ] Validation выполняется при construction.
- [ ] Serialization канонична.
- [ ] Null semantics определена.
- [ ] Privacy определена.
- [ ] Safe logging определен.
- [ ] Type не смешивается с близкими primitive types.
- [ ] Timezone semantics явна.
- [ ] Version semantics явна.
- [ ] Tenant boundary сохранена.
- [ ] External values проходят mapping.
- [ ] AI не получает authority через этот тип.
- [ ] Есть construction и boundary tests.

---

# Non-Goals

Этот документ не определяет:

- concrete Go types;
- package names;
- database schema;
- ORM mappings;
- protobuf;
- OpenAPI;
- GraphQL scalars;
- UI input components;
- localization messages;
- cryptographic algorithms;
- identity provider;
- exact retention periods;
- all domain-specific enums;
- financial types;
- payment types;
- CRM types;
- marketing types;
- payroll types.

---

# Open Questions

Необходимо определить:

- какой identifier format использовать;
- использовать ли UUIDv7 или ULID;
- будут ли ids иметь textual prefixes;
- нужен ли единый parser package;
- какой timestamp precision хранить;
- допускается ли leap-second normalization;
- как представлять ambiguous local time;
- какой стандарт Duration использовать в contracts;
- нужен ли отдельный SchoolDate;
- нужен ли AcademicPeriod;
- нужен ли LessonOccurrence;
- как моделировать recurring Lesson schedule;
- является ли Due Date exact instant или composite type в MVP;
- какие deadline types поддерживать;
- как моделировать deadline before next lesson;
- какой timezone является default tenant timezone;
- может ли Student timezone отличаться от school timezone;
- как хранить timezone history;
- нужен ли VersionedReference generic type;
- какие references обязаны содержать TenantId;
- нужно ли хранить AggregateType строкой или enum;
- нужны ли globally unique EntityId;
- может ли SubmissionId быть unique only inside Homework;
- как долго хранить IdempotencyKey;
- какой canonical payload hashing использовать;
- как canonicalize JSON;
- нужен ли separate DeduplicationPolicy;
- является ли Decision отдельным persisted artifact для всех Policies;
- где хранить Decision inputs;
- нужно ли хранить full Decision или reference;
- какой Decision validity model использовать;
- как моделировать Human Decision;
- нужна ли цифровая подпись важных Decisions;
- как version Reason Codes;
- можно ли deprecated Reason Code продолжать читать;
- какие Reason Codes видит Teacher;
- какие Reason Codes видит Student;
- нужен ли единый error catalog;
- нужен ли Evidence Registry;
- кто создает EvidenceId;
- как подтверждать Teacher Evidence;
- как спорить с Evidence;
- как хранить Evidence confidence;
- как моделировать multi-source Evidence;
- нужен ли Consent Aggregate;
- какие Consent scopes нужны для Concert;
- какие Privacy Levels юридически необходимы;
- нужен ли field-level privacy;
- нужна ли encryption classification;
- как передавать SensitiveReference между services;
- как маскировать identifiers в logs;
- можно ли Actor содержать multiple active roles;
- как выбирать ActiveRole;
- какой authentication strength нужен для sensitive actions;
- как моделировать Guardian authority;
- как моделировать temporary substitute Teacher;
- нужен ли Delegation Aggregate;
- нужен ли PermissionSnapshot в Command;
- как валидировать stale AuthorizationContext;
- какой format CorrelationId использовать;
- может ли один business process иметь несколько CorrelationId;
- нужен ли RootCauseReference во всех events;
- как связывать external trace и business correlation;
- где хранить MigrationReference;
- как импортировать unknown historical time;
- как обозначать approximate OccurredAt;
- нужен ли TemporalPrecision Value Object;
- как хранить imported timezone;
- какие AI confidence scales разрешены;
- как version AI model provenance;
- нужно ли сохранять prompt hash;
- какие AI proposals можно хранить;
- как удалять sensitive prompt content;
- какие Value Objects должны быть shared kernel;
- какие должны копироваться между bounded contexts;
- как предотвращать чрезмерный shared kernel;
- как автоматически проверять serialization compatibility;
- нужен ли schema registry для Value Objects;
- как отражать Value Object contracts в code generation;
- какие Value Objects должны быть реализованы первыми для MVP.

---

# History

| Version | Description |
| --- | --- |
| 1.0.0 | Определен канонический каталог общих Value Objects Belcanto Product: identity, references, versions, time, authorization, decisions, evidence, validation, privacy, traceability, idempotency, lifecycle, communication и provenance. |
