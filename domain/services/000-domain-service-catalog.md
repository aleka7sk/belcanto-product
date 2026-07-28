---
Status: Approved
Version: 1.0.0
Last Updated: 2026-07-28

Document Id: DOMAIN_SERVICE_CATALOG

Document Type:
  - Domain Contract
  - Domain Service Catalog
  - Cross-Aggregate Evaluation Specification
  - Deterministic Calculation Standard
  - Policy Execution Boundary

Owners:
  - Product Owner
  - Education Lead
  - Domain Architecture Lead
  - Technical Lead

Applies To:
  - Domain Services
  - Policies
  - Policy Decisions
  - Aggregate Roots
  - Application Services
  - Process Managers
  - Evidence Evaluation
  - Cross-Aggregate Calculations
  - AI-Assisted Proposals

Related Directories:
  - ../architecture/
  - ../aggregates/
  - ../commands/
  - ../events/
  - ../entities/
  - ../policies/
  - ../value-objects/
  - ../processes/

Related Documents:
  - ../architecture/000-domain-model-rules.md
  - ../aggregates/000-aggregate-catalog.md
  - ../commands/000-domain-command-catalog.md
  - ../events/000-domain-event-catalog.md
  - ../entities/000-entity-catalog.md
  - ../value-objects/000-value-object-catalog.md
  - ../policies/000-domain-policy-overview.md
---

# Domain Service Catalog

> Domain Service Catalog определяет чистые доменные операции Belcanto Product, которые выражают важное бизнес-поведение, но не принадлежат естественным образом одному Aggregate Root, Entity или Value Object.
>
> Domain Service не является application orchestration service, repository, integration adapter, command handler, workflow engine или инфраструктурным helper.
>
> Канонический Domain Service получает полностью подготовленные immutable inputs и возвращает deterministic result, calculation или Policy Decision.

---

# Purpose

В Belcanto существуют решения, для которых недостаточно состояния одного Aggregate.

Например:

- завершение Lesson может зависеть от attendance, duration, lesson summary и completion policy;
- обновление Progress может зависеть от нескольких Evidence sources;
- завершение Goal может зависеть от Goal state, Progress dimensions и подтвержденных evidence;
- присуждение Achievement зависит от versioned definition и фактов нескольких доменных областей;
- Song Readiness зависит от нескольких readiness areas;
- Concert Eligibility зависит от consent, song readiness, concert requirements и организационных ограничений;
- Homework Expiration зависит от deadline, grace period, blockers, submissions и policy configuration;
- Reminder Evaluation зависит от Homework state, due date, prior reminders, quiet hours и suppression rules;
- Notification Decision зависит от recipient preferences, priority, privacy и delivery window;
- Evidence Validity зависит от source state, age, invalidations и consumer policy.

Если поместить такую логику внутрь одного Aggregate:

- Aggregate начнет владеть чужими фактами;
- появятся скрытые cross-aggregate dependencies;
- Aggregate будет обращаться в repository;
- consistency boundary станет ложной;
- правила будет трудно version;
- replay станет недетерминированным;
- тестирование потребует инфраструктуры.

Если поместить эту логику в Application Service:

- доменные решения растворятся в orchestration;
- появятся произвольные `if`;
- Policy Decisions перестанут быть first-class artifacts;
- станет сложно объяснить, почему было принято решение;
- один и тот же rule будет реализован в нескольких handlers.

Domain Service нужен для сохранения доменного смысла и чистой вычислительной границы.

---

# Domain Service Definition

Канонический Domain Service:

```text
Domain Service
├── Domain-specific responsibility
├── Immutable input
├── Deterministic evaluation
├── Versioned rules
├── Structured output
├── Reason Codes
├── Evidence references
└── No infrastructure dependencies
```

Канонический flow:

```text
Application Service / Process Manager
             |
             +--> loads authoritative state
             +--> builds immutable snapshots
             +--> resolves configuration versions
             +--> supplies evaluation time
             |
             v
        Domain Service
             |
             +--> evaluates domain rules
             +--> produces result or Decision
             |
             v
        Aggregate Command
             |
             v
        Aggregate Root
```

## Domain Service vs Policy

Policy и Domain Service близки, но не идентичны.

### Policy

Policy отвечает на нормативный вопрос:

- Разрешено ли?
- Требуется ли?
- Подходит ли?
- Какой outcome должен быть применен?

Policy возвращает versioned Decision.

Пример:

`Concert Eligibility Policy`

возвращает:

- Eligible
- ConditionallyEligible
- NotEligible
- HumanReviewRequired

### Domain Service

Domain Service выполняет доменное вычисление, которое:

- может подготовить Policy Input;
- может применить Policy;
- может объединить несколько calculations;
- может вернуть calculation result;
- может вернуть Decision, если является реализацией Policy evaluation;
- не обязательно принимает нормативное решение.

Пример:

`EvidenceValidityEvaluator`

может вернуть:

- Valid Evidence Set
- Invalid Evidence Set
- Stale Evidence Set
- Missing Evidence Requirements

После этого GoalCompletionPolicy принимает Decision.

## Domain Service vs Application Service

| Responsibility | Domain Service | Application Service |
| --- | --- | --- |
| Domain calculation | Да | Оркестрирует |
| Load repositories | Нет | Да |
| Start transaction | Нет | Да |
| Authorization | Не основная задача | Да |
| Idempotency | Нет | Да |
| Call infrastructure | Нет | Может |
| Produce Decision | Может | Сохраняет/применяет |
| Send Command | Нет | Может |
| Persist Aggregate | Нет | Да |
| Retry coordination | Нет | Да |

Неправильно:

```text
GoalCompletionService
├── load goal repository
├── load progress repository
├── call AI
├── update database
├── send notification
└── publish event
```

Это Application Service или Process Manager, но не Domain Service.

## Domain Service vs Process Manager

Domain Service:

```text
input -> deterministic result
```

Process Manager:

```text
event -> stored workflow state -> command -> waiting -> next event
```

Domain Service не хранит состояние ожидания и не управляет долгоживущим процессом.

## Domain Service vs Aggregate

Behavior принадлежит Aggregate, если:

- изменяет state одного Aggregate;
- защищает его внутренний invariant;
- использует только owned state;
- естественно выражается методом Root.

Behavior принадлежит Domain Service, если:

- требуется несколько independent domain snapshots;
- ни один Aggregate не является естественным владельцем;
- вычисление должно быть reusable;
- правило имеет отдельную version;
- результат нужен нескольким workflows;
- решение не должно мутировать state напрямую.

## Core Domain Service Rules

### DS-001: Domain Service must have a domain-specific name

Допустимо:

- GoalCompletionEvaluator
- ConcertEligibilityEvaluator
- EvidenceValidityEvaluator

Запрещено:

- DomainService
- BusinessService
- CommonService
- UtilityService
- Manager
- Helper
- Processor

если название не раскрывает доменный смысл.

### DS-002: Domain Service must be stateless

Domain Service не сохраняет mutable state между вызовами.

Versioned configuration передается как immutable input или constructor dependency, если она immutable и явно идентифицирована.

### DS-003: Domain Service must be deterministic

Одинаковые:

- inputs;
- service version;
- policy version;
- configuration version;
- evaluation time

должны давать одинаковый результат.

### DS-004: Domain Service must not read system clock

Неправильно:

```text
now := time.Now()
```

Корректно:

```text
Evaluate(input, evaluatedAt)
```

### DS-005: Domain Service must not load repositories

Все необходимые данные загружаются до вызова.

### DS-006: Domain Service must not persist state

### DS-007: Domain Service must not publish Events

Domain Event создает Aggregate Root после применения результата.

### DS-008: Domain Service must not send Commands

Commands отправляет Application Service или Process Manager.

### DS-009: Domain Service must not call external APIs

Включая:

- CRM;
- notification provider;
- file storage;
- AI provider;
- identity provider;
- calendar;
- payment system.

### DS-010: Domain Service must not authorize infrastructure access

Он может проверить domain relationship, если она передана как fact, но authentication и permission enforcement принадлежат Application Layer.

### DS-011: Domain Service input must be explicit

Нельзя скрыто использовать:

- global configuration;
- current tenant;
- current user;
- latest policy;
- default timezone;
- environment variable;
- mutable cache.

### DS-012: Domain Service output must be structured

Нежелательно:

- true
- false
- "ok"
- "not ready"

Предпочтительно:

- Decision
- EvaluationResult
- CalculationResult
- ValidationResult
- ReasonCodes
- Conditions
- BlockingConditions
- InputReferences

### DS-013: Domain Service must preserve input provenance

Результат должен ссылаться на exact input identities и versions.

### DS-014: Domain Service must have an explicit version

Behavioral change требует новой version.

### DS-015: A Domain Service must not become a rule dumping ground

Если service содержит несвязанные методы, его нужно разделить по responsibilities.

### DS-016: Aggregate invariants must not be moved into a Domain Service merely for reuse

Aggregate остается владельцем своих внутренних invariants.

### DS-017: Domain Service does not bypass Aggregate transition rules

Даже положительный Decision не меняет state без Aggregate Command.

### DS-018: Domain Service result may become stale

Result должен определять:

- source versions;
- evaluated time;
- validity;
- invalidation triggers;
- reevaluation requirements.

### DS-019: Human Review must be an explicit outcome

Service не должен угадывать, когда evidence недостаточно.

### DS-020: Missing data must not be silently replaced by defaults

### DS-021: AI output is never trusted as authoritative input by default

AI proposal должен быть:

- validated;
- classified;
- linked to provenance;
- confirmed by Human или approved deterministic policy.

### DS-022: Domain Service should operate on domain snapshots, not persistence DTOs

### DS-023: Domain Service must not expose mutable input references

### DS-024: Domain Service errors and negative decisions are distinct

Negative domain outcome:

- NotEligible
- NotReady
- InsufficientEvidence

не является technical error.

### DS-025: Unexpected internal inability to evaluate must not fabricate a Decision

# Canonical Domain Services

- LessonCompletionEvaluator
- ProgressUpdateEvaluator
- GoalCompletionEvaluator
- AchievementEligibilityEvaluator
- SongReadinessEvaluator
- ConcertEligibilityEvaluator
- HomeworkExpirationEvaluator
- HomeworkReminderEvaluator
- NotificationDecisionEvaluator
- EvidenceValidityEvaluator

Дополнительные candidate services:

- HomeworkCompletionEvaluator
- PeriodicReviewEvaluator
- NotificationBundlingEvaluator
- ConcertProgramConstraintEvaluator
- GoalEvidenceEvaluator
- AchievementRevocationEvaluator

Они не вводятся автоматически, пока responsibility не подтверждена.

# Shared Evaluation Contract

Все canonical evaluators должны по возможности использовать совместимую структуру.

```text
EvaluationRequest
├── EvaluationId
├── EvaluationType
├── TenantId
├── SubjectReference
├── RequestedBy
├── EvaluatedAt
├── ServiceReference
├── PolicyReference
├── ConfigurationReferences
├── InputReferences
├── InputSnapshots
├── EvidenceReferences
├── CorrelationId
└── CausationId
```

Результат:

```text
EvaluationResult
├── EvaluationId
├── EvaluationType
├── ServiceReference
├── PolicyReference
├── Outcome
├── EvaluatedAt
├── InputReferences
├── InputVersions
├── EvidenceReferences
├── AcceptedEvidence
├── RejectedEvidence
├── ReasonCodes
├── Conditions
├── BlockingConditions
├── ValidUntil
├── ReevaluationTriggers
├── HumanReviewRequirement
├── Warnings
└── CalculationDetailsReference
```

## ServiceReference

```text
ServiceReference
├── ServiceId
├── ServiceVersion
└── ImplementationContractVersion
```

ImplementationContractVersion не обязательно совпадает с deployment version.

# Evaluation Outcome Categories

Общие meta-outcomes:

- Evaluated
- InsufficientInput
- InsufficientEvidence
- StaleInput
- HumanReviewRequired
- NotApplicable
- UnableToEvaluate

Domain-specific outcome хранится отдельно.

Пример:

```text
EvaluationStatus: Evaluated
DomainOutcome: ConditionallyEligible
```

# LessonCompletionEvaluator

## Purpose

Определяет, достаточно ли фактов для завершения Lesson и какой completion outcome должен быть предложен Aggregate.

Evaluator не завершает Lesson самостоятельно.

## Why This Is a Domain Service

Completion может зависеть от:

- Lesson lifecycle;
- scheduled and actual time;
- attendance records;
- lesson type;
- required summary;
- Teacher authority snapshot;
- cancellation facts;
- completion method;
- special completion policy;
- imported/offline evidence.

Некоторые данные находятся внутри Lesson, но evaluation может использовать versioned Policy и external authorization/evidence snapshots.

Если вся логика проста и полностью принадлежит Lesson, Aggregate MAY выполнять ее самостоятельно.

Domain Service оправдан, когда completion rules становятся configurable и cross-context.

## Inputs

```text
LessonCompletionInput
├── LessonSnapshot
├── AttendanceSnapshot
├── CompletionPolicyReference
├── LessonTypeConfiguration
├── TeacherAssignmentReference
├── CompletionEvidenceReferences
├── RequestedCompletionMethod
├── EvaluatedAt
└── ActorContextSnapshot
```

## Outputs

```text
LessonCompletionDecision
├── DecisionId
├── Outcome
├── CompletionMethod
├── EffectiveCompletionTime
├── ReasonCodes
├── Conditions
├── BlockingConditions
├── MissingRequirements
├── EvidenceReferences
├── PolicyReference
├── InputVersions
├── EvaluatedAt
├── ValidUntil
└── HumanReviewRequired
```

## Outcomes

- CompletionAllowed
- CompletionAllowedWithWarnings
- CompletionDeferred
- CompletionRejected
- HumanReviewRequired
- AlreadyCompleted
- NotApplicable

## Core Rules

- Lesson должен существовать в completion-eligible lifecycle state.
- Cancelled Lesson нельзя Complete обычным способом.
- Already Completed Lesson возвращает idempotent outcome.
- Required attendance должна быть зафиксирована или явно waived.
- Completion time не может быть раньше допустимого момента без special method.
- Required summary должна существовать, если Lesson type ее требует.
- Actor relationship должна соответствовать Lesson.
- Imported completion требует provenance.
- Offline completion требует trusted synchronization metadata.
- Missing attendance не заменяется Present.
- AI-generated summary не подтверждает проведение Lesson.
- Completion Decision не обновляет Progress автоматически.
- Completion Decision может содержать follow-up condition для Process Manager.

## Reason Codes

- LESSON_NOT_COMPLETION_ELIGIBLE
- LESSON_ALREADY_COMPLETED
- LESSON_CANCELLED
- LESSON_ATTENDANCE_REQUIRED
- LESSON_SUMMARY_REQUIRED
- LESSON_COMPLETION_TOO_EARLY
- LESSON_COMPLETION_METHOD_NOT_ALLOWED
- LESSON_TEACHER_ASSIGNMENT_INVALID
- LESSON_COMPLETION_EVIDENCE_MISSING
- LESSON_COMPLETION_INPUT_STALE
- LESSON_COMPLETION_REVIEW_REQUIRED

## Aggregate Application

```text
decision = evaluator.Evaluate(input)

lesson.Complete(
    decisionReference,
    completionMethod,
    effectiveCompletionTime,
)
```

Lesson должен повторно проверить:

- Decision target;
- Decision validity;
- input Lesson Version;
- allowed lifecycle transition.

## AI Boundaries

AI может:

- подготовить draft summary;
- выявить missing fields;
- предложить completion method;
- проверить contradiction.

AI не может подтвердить проведение Lesson.

## Tests

- scheduled Lesson with complete attendance;
- missing required attendance;
- cancelled Lesson;
- already completed idempotency;
- stale Lesson Version;
- imported completion;
- unauthorized Teacher relationship;
- AI summary without human confirmation;
- exact time boundary;
- group Lesson partial attendance.

# ProgressUpdateEvaluator

## Purpose

Определяет, должно ли подтвержденное Evidence изменить один или несколько Progress Dimension State.

## Why This Is a Domain Service

Progress update может зависеть от:

- current Progress state;
- Evidence from Lesson;
- Homework Review;
- performance;
- Teacher assessment;
- curriculum definitions;
- evidence freshness;
- confidence;
- multiple evidence sources;
- conflicting evaluations;
- progress update policy.

Ни один Evidence source не должен владеть Progress Aggregate.

## Inputs

```text
ProgressUpdateInput
├── ProgressRecordSnapshot
├── DimensionDefinitions
├── CandidateEvidenceReferences
├── EvidenceValidityResults
├── ExistingEvidenceReferences
├── ProgressUpdatePolicyReference
├── CurriculumReference
├── TeacherAuthoritySnapshot
├── EvaluatedAt
└── RequestedScope
```

## Outputs

```text
ProgressUpdateDecision
├── DecisionId
├── Outcome
├── DimensionUpdates
├── AcceptedEvidence
├── RejectedEvidence
├── ConflictingEvidence
├── ReasonCodes
├── ReviewRequirements
├── PolicyReference
├── InputVersions
├── EvaluatedAt
└── ValidUntil
```

Dimension update:

```text
ProgressDimensionUpdate
├── DimensionId
├── PreviousState
├── ProposedState
├── Confidence
├── SupportingEvidence
├── OpposingEvidence
├── ReasonCodes
└── ReviewRequired
```

## Outcomes

- UpdateApproved
- PartialUpdateApproved
- NoChangeRequired
- InsufficientEvidence
- ConflictingEvidence
- HumanReviewRequired
- StaleInput

## Core Rules

- Evidence относится к тому же Student.
- Evidence scope совпадает с Progress scope.
- Invalidated Evidence не используется.
- Evidence validity учитывается на evaluated time.
- Один Event не обязан менять Progress.
- Notification interaction не используется как Evidence.
- Количество Evidence не заменяет качество.
- Новое Evidence не должно автоматически ухудшать state без approved policy.
- Unknown не заменяется baseline level.
- Dimension Definition Version фиксируется.
- More recent weak Evidence не обязательно заменяет older confirmed Evidence.
- Conflicting authoritative Evidence требует review или explicit conflict policy.
- AI inference остается unconfirmed.
- Update Decision должен содержать exact previous Progress Version.
- Application Aggregate повторно проверяет version.

## Reason Codes

- PROGRESS_EVIDENCE_MISSING
- PROGRESS_EVIDENCE_INVALID
- PROGRESS_EVIDENCE_STALE
- PROGRESS_EVIDENCE_SCOPE_MISMATCH
- PROGRESS_EVIDENCE_STUDENT_MISMATCH
- PROGRESS_EVIDENCE_CONFLICT
- PROGRESS_DIMENSION_UNKNOWN
- PROGRESS_DIMENSION_DEFINITION_STALE
- PROGRESS_UPDATE_NOT_SUPPORTED
- PROGRESS_UPDATE_NO_CHANGE
- PROGRESS_UPDATE_REVIEW_REQUIRED
- PROGRESS_INPUT_STALE

## AI Boundaries

AI может:

- классифицировать candidate dimension;
- предложить update;
- summarise evidence;
- обнаружить contradiction.

AI не может быть authoritative source обновления.

## Tests

- one confirmed Evidence;
- multiple supporting Evidence;
- invalidated Evidence;
- Student mismatch;
- scope mismatch;
- conflicting Teacher assessments;
- stale Definition Version;
- AI-only input;
- no-change result;
- partial dimension update;
- optimistic version mismatch.

# GoalCompletionEvaluator

## Purpose

Определяет, выполнены ли критерии Goal и может ли Goal перейти в Completed.

## Why This Is a Domain Service

Goal может ссылаться на:

- Progress dimensions;
- Homework completion;
- Lesson evidence;
- performance evidence;
- Song Readiness;
- explicit Teacher assessment;
- deadline;
- completion criteria version.

Goal Aggregate не должен загружать все источники самостоятельно.

## Inputs

```text
GoalCompletionInput
├── GoalSnapshot
├── GoalCriteriaVersion
├── ProgressSnapshots
├── EvidenceValidityResults
├── CompletionEvidenceReferences
├── PreviousGoalDecisions
├── GoalCompletionPolicyReference
├── EvaluatedAt
└── ActorAuthoritySnapshot
```

## Outputs

```text
GoalCompletionDecision
├── DecisionId
├── Outcome
├── CompletedCriteria
├── UnmetCriteria
├── PartiallyMetCriteria
├── SupportingEvidence
├── RejectedEvidence
├── ReasonCodes
├── Conditions
├── BlockingConditions
├── PolicyReference
├── InputVersions
├── EvaluatedAt
├── ValidUntil
└── HumanReviewRequirement
```

## Outcomes

- CompletionApproved
- CompletionConditionallyApproved
- NotCompleted
- InsufficientEvidence
- HumanReviewRequired
- AlreadyCompleted
- StaleInput

## Core Rules

- Goal должен быть Active или в другом explicitly completion-eligible state.
- Goal criteria version фиксируется.
- Каждый required criterion оценивается отдельно.
- Completion не определяется процентом, если criteria не допускают scoring.
- Missing Evidence не считается отрицательным Evidence.
- Expired Evidence не используется без override.
- Completed criterion должен иметь supporting Evidence или approved human Decision.
- Goal deadline не завершает и не отменяет Goal автоматически.
- Goal Completion не означает Achievement Award.
- Reopened Goal требует новой evaluation.
- Already Completed Goal обрабатывается идемпотентно.
- AI proposal не является completion evidence.
- Human override сохраняет исходный evaluation.

## Reason Codes

- GOAL_NOT_ACTIVE
- GOAL_ALREADY_COMPLETED
- GOAL_CRITERIA_VERSION_MISSING
- GOAL_CRITERIA_UNMET
- GOAL_CRITERIA_PARTIALLY_MET
- GOAL_EVIDENCE_MISSING
- GOAL_EVIDENCE_INVALID
- GOAL_EVIDENCE_STALE
- GOAL_EVIDENCE_CONFLICT
- GOAL_COMPLETION_REVIEW_REQUIRED
- GOAL_COMPLETION_INPUT_STALE
- GOAL_COMPLETION_POLICY_NOT_APPLICABLE

## AI Boundaries

AI может:

- сопоставлять Evidence и criteria как proposal;
- формировать summary;
- подсвечивать unmet criteria.

AI не может Complete Goal.

## Tests

- all required criteria met;
- optional criteria missing;
- required criterion missing;
- stale evidence;
- reopened Goal;
- already completed;
- deadline passed but criteria met;
- deadline passed and criteria not met;
- AI-only proposal;
- human override preservation.

# AchievementEligibilityEvaluator

## Purpose

Определяет, соответствует ли Student versioned Achievement Definition и может ли быть создан Achievement Award.

## Why This Is a Domain Service

Achievement criteria могут включать:

- Goal completion;
- Progress;
- Lesson participation;
- Homework;
- Song repertoire;
- Concert performance;
- historical awards;
- minimum time periods;
- uniqueness rules.

Achievement Definition и Student facts принадлежат разным Aggregates.

## Inputs

```text
AchievementEligibilityInput
├── AchievementDefinitionSnapshot
├── AchievementDefinitionVersion
├── StudentReference
├── QualificationEvidence
├── ExistingAwardReferences
├── EvidenceValidityResults
├── AchievementAwardPolicyReference
├── EvaluatedAt
└── ActorAuthoritySnapshot
```

## Outputs

```text
AchievementEligibilityDecision
├── DecisionId
├── Outcome
├── QualifiedCriteria
├── UnmetCriteria
├── SupportingEvidence
├── DuplicateAwardAssessment
├── ReasonCodes
├── Conditions
├── BlockingConditions
├── PolicyReference
├── InputVersions
├── EvaluatedAt
└── ValidUntil
```

## Outcomes

- EligibleForAward
- ConditionallyEligible
- NotEligible
- AlreadyAwarded
- HumanReviewRequired
- DefinitionUnavailable
- StaleInput

## Core Rules

- Definition должна быть Published.
- Exact Definition Version фиксируется.
- Retired Definition может использоваться только согласно retirement rules.
- Award uniqueness scope explicit.
- Existing active Award предотвращает duplicate, если repeatable=false.
- Revoked Award behavior определяется Definition.
- Evidence должно относиться к Student.
- Historical imported facts требуют provenance.
- Achievement Award не создается непосредственно evaluator.
- AI не может выдавать Achievement.
- Criteria change требует новой Definition Version.
- Уже выданный Award не пересчитывается скрыто по новой Definition.

## Reason Codes

- ACHIEVEMENT_DEFINITION_NOT_PUBLISHED
- ACHIEVEMENT_DEFINITION_RETIRED
- ACHIEVEMENT_DEFINITION_VERSION_MISMATCH
- ACHIEVEMENT_CRITERIA_UNMET
- ACHIEVEMENT_EVIDENCE_MISSING
- ACHIEVEMENT_EVIDENCE_INVALID
- ACHIEVEMENT_ALREADY_AWARDED
- ACHIEVEMENT_REPEAT_NOT_ALLOWED
- ACHIEVEMENT_REVIEW_REQUIRED
- ACHIEVEMENT_INPUT_STALE

## AI Boundaries

AI может находить candidate Achievement.

AI не может быть award authority.

## Tests

- published definition;
- draft definition;
- retired definition;
- all criteria met;
- duplicate non-repeatable award;
- repeatable achievement;
- revoked prior award;
- imported evidence;
- AI candidate;
- definition version change.

# SongReadinessEvaluator

## Purpose

Вычисляет overall Song Readiness на основании Readiness Area States, Evidence и Policy.

## Why This Is a Domain Service

Readiness Aggregate владеет текущими area states, но overall result может зависеть от:

- performance type;
- required areas;
- weighted or mandatory areas;
- safety rules;
- Song Version;
- recency;
- conditions;
- evidence quality;
- teacher review.

Evaluation policy может быть independently versioned.

## Inputs

```text
SongReadinessEvaluationInput
├── SongReadinessSnapshot
├── SongVersionReference
├── PerformanceType
├── ReadinessAreaStates
├── EvidenceValidityResults
├── ReadinessModelVersion
├── SongReadinessPolicyReference
├── EvaluatedAt
└── EvaluatorAuthoritySnapshot
```

## Outputs

```text
SongReadinessDecision
├── DecisionId
├── Outcome
├── AreaOutcomes
├── RequiredAreas
├── MissingAreas
├── Conditions
├── BlockingConditions
├── AcceptedEvidence
├── RejectedEvidence
├── ReasonCodes
├── PolicyReference
├── InputVersions
├── EvaluatedAt
├── ValidUntil
└── HumanReviewRequirement
```

## Outcomes

- Ready
- ConditionallyReady
- NotReady
- ReviewRequired
- InsufficientEvidence
- StaleInput
- NotApplicable

## Core Rules

- Song Version exact.
- Performance Type exact.
- Required Areas определяются Readiness Model.
- Mandatory Safety Area нельзя компенсировать weighted score.
- Ready требует прохождения всех mandatory blockers.
- ConditionallyReady требует explicit Conditions.
- NotReady содержит причины, но не унизительные оценки.
- Stale Area не считается Ready.
- NotApplicable Area не считается missing.
- Performance Type change требует reevaluation.
- Song Version change invalidates previous result.
- Readiness не означает Concert Eligibility.
- AI analysis recording остается advisory.
- Human review может быть mandatory для selected areas.
- ValidUntil определяется evidence freshness и performance date.

## Reason Codes

- SONG_VERSION_MISMATCH
- SONG_READINESS_MODEL_VERSION_MISMATCH
- SONG_READINESS_AREA_MISSING
- SONG_READINESS_AREA_STALE
- SONG_READINESS_EVIDENCE_MISSING
- SONG_READINESS_EVIDENCE_INVALID
- SONG_READINESS_SAFETY_BLOCKER
- SONG_READINESS_CONDITIONS_REQUIRED
- SONG_READINESS_REVIEW_REQUIRED
- SONG_READINESS_INPUT_STALE

## AI Boundaries

AI может анализировать pitch/rhythm только как proposal при наличии consent и validated model.

AI не может финализировать readiness.

## Tests

- all mandatory areas Ready;
- optional area NotReady;
- mandatory area NotReady;
- conditional area;
- Safety blocker;
- stale area;
- changed Song Version;
- changed Performance Type;
- AI-only assessment;
- human review requirement.

# ConcertEligibilityEvaluator

## Purpose

Определяет eligibility конкретного Concert Participation относительно versioned Concert Requirements.

## Why This Is a Domain Service

Eligibility зависит от нескольких Aggregates:

- Concert;
- Concert Requirements;
- Concert Participation;
- Consent;
- Student Song;
- Song Readiness;
- possibly attendance or rehearsal state;
- restrictions;
- age/guardian requirements;
- performance type.

Ни один Aggregate не должен владеть всеми facts.

## Inputs

```text
ConcertEligibilityInput
├── ConcertSnapshot
├── ConcertRequirementsSnapshot
├── ParticipationSnapshot
├── ConsentSnapshot
├── StudentSongReferences
├── SongReadinessDecisions
├── AdditionalEligibilityEvidence
├── EvidenceValidityResults
├── ConcertEligibilityPolicyReference
├── EvaluatedAt
└── EvaluatorAuthoritySnapshot
```

## Outputs

```text
ConcertEligibilityDecision
├── DecisionId
├── Outcome
├── RequirementsVersion
├── SatisfiedRequirements
├── UnsatisfiedRequirements
├── Conditions
├── BlockingConditions
├── AcceptedEvidence
├── RejectedEvidence
├── ReasonCodes
├── PolicyReference
├── InputVersions
├── EvaluatedAt
├── ValidUntil
├── InvalidatingEvents
└── HumanReviewRequirement
```

## Outcomes

- Eligible
- ConditionallyEligible
- NotEligible
- ReviewRequired
- InsufficientEvidence
- NotApplicable
- StaleInput

## Core Rules

- Concert принимает Participation в текущем lifecycle state.
- Requirements Version exact.
- Participation относится к Concert.
- Consent проверяется отдельно и в нужном scope.
- Song Version matches readiness Decision.
- Stale Readiness не поддерживает positive eligibility.
- Conditionally Ready MAY поддерживать conditional eligibility, если requirements допускают.
- NotReady блокирует соответствующую performance.
- Required rehearsal evidence может быть обязательным.
- Eligibility не означает Approval.
- Capacity не должна подменять Eligibility, если это организационный Approval concern.
- Slot availability не является Eligibility.
- Requirements change invalidates existing Decision.
- Consent withdrawal invalidates dependent Decision.
- AI не является authoritative evaluator.
- Negative outcome должен быть explainable и non-shaming.

## Reason Codes

- CONCERT_NOT_ACCEPTING_PARTICIPATION
- CONCERT_REQUIREMENTS_VERSION_MISMATCH
- CONCERT_PARTICIPATION_MISMATCH
- CONCERT_CONSENT_REQUIRED
- CONCERT_CONSENT_INVALID
- CONCERT_SONG_REQUIRED
- CONCERT_SONG_VERSION_MISMATCH
- CONCERT_SONG_NOT_READY
- CONCERT_SONG_READINESS_STALE
- CONCERT_REHEARSAL_REQUIRED
- CONCERT_REQUIREMENT_UNMET
- CONCERT_ELIGIBILITY_REVIEW_REQUIRED
- CONCERT_ELIGIBILITY_INPUT_STALE

## AI Boundaries

AI может подготовить checklist, но не принять eligibility Decision.

## Tests

- ready song and valid consent;
- conditional readiness;
- invalid consent;
- withdrawn consent;
- stale readiness;
- wrong Song Version;
- changed requirements;
- concert closed;
- no slot capacity but otherwise eligible;
- AI assessment;
- human override.

# HomeworkExpirationEvaluator

## Purpose

Определяет, может ли Homework Assignment перейти в Expired, должно ли получить Grace Period, быть отложено на review или остаться активным.

## Why This Is a Domain Service

Expiration зависит от:

- Homework lifecycle;
- Due Date semantics;
- Grace Period;
- Submission state;
- active Correction Requests;
- Blockers;
- Teacher decision;
- replacement;
- requiredness;
- reminder state;
- expiration policy.

Aggregate владеет transition, но evaluator формирует versioned Decision.

## Inputs

```text
HomeworkExpirationInput
├── HomeworkSnapshot
├── SubmissionSnapshots
├── ReviewSnapshots
├── ActiveBlockers
├── CorrectionRequestSnapshots
├── DueDate
├── GracePeriod
├── HomeworkExpirationPolicyReference
├── EvaluatedAt
├── StudentContextSnapshot
└── TeacherAuthoritySnapshot
```

## Outputs

```text
HomeworkExpirationDecision
├── DecisionId
├── Outcome
├── EffectiveExpirationAt
├── GracePeriodProposal
├── ReviewRequirement
├── ActiveBlockerAssessment
├── ReasonCodes
├── Conditions
├── BlockingConditions
├── PolicyReference
├── InputVersions
├── EvaluatedAt
└── ValidUntil
```

## Outcomes

- Expire
- DoNotExpire
- StartGracePeriod
- ExtendGracePeriod
- DeferForTeacherReview
- AlreadyTerminal
- NotApplicable
- StaleInput

## Core Rules

- Deadline crossing does not automatically mean Expire.
- Completed, Cancelled, Replaced или already Expired Homework не expire повторно.
- Valid submitted work может блокировать expiration pending review.
- Active required Correction Request может иметь собственный deadline.
- Active Blocker влияет только согласно Policy.
- Health/personal blocker не требует раскрытия деталей.
- Grace Period history сохраняется.
- Expiration не удаляет Submission.
- Expiration не означает Student failure.
- Reopen остается отдельным action.
- Reminder Delivery не влияет на eligibility for expiration.
- Evaluated time explicit.
- Timezone and deadline type exact.
- AI не может Expire Homework.
- Scheduled evaluator обладает trigger authority, не decision authority.

## Reason Codes

- HOMEWORK_NOT_EXPIRATION_ELIGIBLE
- HOMEWORK_ALREADY_TERMINAL
- HOMEWORK_DEADLINE_NOT_REACHED
- HOMEWORK_VALID_SUBMISSION_EXISTS
- HOMEWORK_REVIEW_PENDING
- HOMEWORK_CORRECTION_ACTIVE
- HOMEWORK_BLOCKER_ACTIVE
- HOMEWORK_GRACE_PERIOD_ACTIVE
- HOMEWORK_GRACE_PERIOD_RECOMMENDED
- HOMEWORK_EXPIRATION_REVIEW_REQUIRED
- HOMEWORK_EXPIRATION_INPUT_STALE

## AI Boundaries

AI может предложить review, но не expire.

## Tests

- deadline not reached;
- deadline exact boundary;
- deadline passed no submission;
- valid submission pending review;
- active blocker;
- blocker not affecting deadline;
- active correction;
- grace period;
- already completed;
- already expired;
- timezone boundary;
- AI actor.

# HomeworkReminderEvaluator

## Purpose

Определяет, следует ли создать, перенести, подавить или отменить Homework Reminder occurrence.

## Why This Is a Domain Service

Reminder Decision зависит от:

- current Homework state;
- Due Date;
- reminder strategy;
- existing reminders;
- submission status;
- blockers;
- quiet hours;
- recipient timezone;
- preferences;
- notification policy;
- maximum reminder frequency;
- educational pressure constraints.

Reminder Plan владеет reminders, но не должен загружать Homework Aggregate.

## Inputs

```text
HomeworkReminderInput
├── HomeworkSnapshotReference
├── HomeworkLifecycleSnapshot
├── DueDate
├── SubmissionSummary
├── BlockerSummary
├── ReminderPlanSnapshot
├── PriorReminderSummaries
├── RecipientPreferences
├── QuietHours
├── HomeworkReminderPolicyReference
├── NotificationPolicyReference
├── EvaluatedAt
└── CandidateOccurrence
```

## Outputs

```text
HomeworkReminderDecision
├── DecisionId
├── Outcome
├── ScheduledFor
├── ReminderType
├── Priority
├── NotificationCategory
├── SuppressionReasonCodes
├── Conditions
├── PolicyReferences
├── InputVersions
├── EvaluatedAt
├── ValidUntil
└── RevalidationRequiredAtSendTime
```

## Outcomes

- Schedule
- Reschedule
- Suppress
- Cancel
- ExpireOccurrence
- NoReminderRequired
- HumanReviewRequired
- StaleInput

## Core Rules

- Completed, Cancelled, Replaced или Expired Homework не получает ordinary reminder.
- Submitted Homework не получает reminder выполнить то, что уже submitted.
- Reminder не должен нарушать Quiet Hours.
- Maximum count и minimum interval enforced.
- Active blocker может suppress reminder.
- Lack of response не оправдывает escalation.
- Reminder не должен shame Student.
- Priority обычно Low или Normal.
- Critical запрещен для ordinary Homework.
- Recipient preference учитывается согласно requiredness.
- Reminder send требует revalidation.
- Stale Homework Version suppresses delivery.
- Reminder outcome не изменяет Homework.
- AI не увеличивает frequency.
- Repeated failure delivery не создает дополнительные reminders автоматически.

## Reason Codes

- HOMEWORK_REMINDER_NOT_REQUIRED
- HOMEWORK_REMINDER_HOMEWORK_TERMINAL
- HOMEWORK_REMINDER_ALREADY_SUBMITTED
- HOMEWORK_REMINDER_BLOCKER_ACTIVE
- HOMEWORK_REMINDER_QUIET_HOURS
- HOMEWORK_REMINDER_MAXIMUM_REACHED
- HOMEWORK_REMINDER_INTERVAL_TOO_SHORT
- HOMEWORK_REMINDER_DUPLICATE
- HOMEWORK_REMINDER_RECIPIENT_DISABLED
- HOMEWORK_REMINDER_STALE_HOMEWORK
- HOMEWORK_REMINDER_REVIEW_REQUIRED

## AI Boundaries

AI может предложить supportive wording отдельно от Decision.

AI не может менять frequency, urgency или suppression policy.

## Tests

- active Homework near deadline;
- completed Homework;
- submitted Homework;
- active blocker;
- quiet hours;
- maximum reminders;
- duplicate occurrence;
- timezone;
- stale version;
- disabled channel;
- AI escalation attempt;
- send-time revalidation.

# NotificationDecisionEvaluator

## Purpose

Определяет, разрешено ли создать или доставить Notification Intent конкретному Recipient, через какие channels и в какое время.

## Why This Is a Domain Service

Notification Decision зависит от:

- domain source;
- category;
- priority;
- recipient;
- consent/preferences;
- quiet hours;
- privacy;
- channel capability;
- deduplication;
- bundling;
- expiration;
- suppression rules.

NotificationIntent Aggregate хранит outcome, но не должен загружать preferences и provider capabilities самостоятельно.

## Inputs

```text
NotificationDecisionInput
├── SourceDomainReference
├── ProposedNotification
├── RecipientSnapshot
├── RecipientPreferences
├── ConsentReferences
├── QuietHours
├── ChannelCapabilities
├── ExistingIntentSummaries
├── PrivacyClassification
├── NotificationPolicyReference
├── EvaluatedAt
└── DeliveryPurpose
```

## Outputs

```text
NotificationDecision
├── DecisionId
├── Outcome
├── AllowedChannels
├── ProhibitedChannels
├── DeliveryWindow
├── Priority
├── BundleKey
├── DeduplicationKey
├── TemplateRequirements
├── ContentRestrictions
├── ReasonCodes
├── Conditions
├── PolicyReference
├── InputVersions
├── EvaluatedAt
├── ValidUntil
└── SendTimeRevalidationRequired
```

## Outcomes

- Allow
- AllowWithRestrictions
- Defer
- Bundle
- Suppress
- Reject
- HumanReviewRequired
- StaleInput

## Core Rules

- Source domain fact должен существовать.
- Notification не создает domain truth.
- Recipient relationship valid.
- Channel allowed for Privacy Level.
- Sensitive data not exposed through unsafe channel.
- Quiet Hours respected.
- Critical priority requires explicit category and authority.
- Duplicate intent suppressed or bundled.
- Expired content not delivered.
- Consent/preferences applied by category.
- Security notifications may override selected preferences according to policy.
- Educational notifications must avoid pressure.
- Delivery Window explicit.
- Send-time revalidation required for stale-prone facts.
- AI-generated content проходит content validation.
- Notification Open не является educational Evidence.

## Reason Codes

- NOTIFICATION_SOURCE_INVALID
- NOTIFICATION_RECIPIENT_INVALID
- NOTIFICATION_CHANNEL_NOT_ALLOWED
- NOTIFICATION_PRIVACY_RESTRICTION
- NOTIFICATION_QUIET_HOURS
- NOTIFICATION_DUPLICATE
- NOTIFICATION_BUNDLE_REQUIRED
- NOTIFICATION_PREFERENCE_DISABLED
- NOTIFICATION_CONSENT_REQUIRED
- NOTIFICATION_EXPIRED
- NOTIFICATION_PRIORITY_NOT_ALLOWED
- NOTIFICATION_CONTENT_RESTRICTED
- NOTIFICATION_REVIEW_REQUIRED
- NOTIFICATION_INPUT_STALE

## AI Boundaries

AI может подготовить текст только после Decision о допустимом communication intent.

AI не решает:

- recipient;
- channel permission;
- urgency;
- consent;
- privacy;
- delivery time.

## Tests

- normal educational notification;
- quiet hours;
- sensitive content over SMS;
- duplicate intent;
- bundling;
- disabled preference;
- required security message;
- expired intent;
- critical priority misuse;
- stale source;
- AI-generated unsafe text.

# EvidenceValidityEvaluator

## Purpose

Определяет, может ли Evidence Reference использоваться в конкретной Policy Evaluation.

## Why This Is a Domain Service

Evidence validity не является универсальным свойством только Evidence.

Один факт может быть:

- valid для Progress;
- stale для Goal;
- insufficient для Achievement;
- prohibited для Concert Eligibility;
- restricted для privacy reasons.

Validity зависит от consumer context.

## Inputs

```text
EvidenceValidityInput
├── EvidenceReferences
├── EvidenceSourceSnapshots
├── ConsumerPolicyReference
├── RequiredEvidenceTypes
├── SubjectReference
├── Scope
├── MaximumAge
├── ConfirmationRequirements
├── PrivacyContext
├── EvaluatedAt
└── InvalidatingEventReferences
```

## Outputs

```text
EvidenceValidityResult
├── EvaluationId
├── AcceptedEvidence
├── RejectedEvidence
├── StaleEvidence
├── RestrictedEvidence
├── DisputedEvidence
├── MissingEvidenceRequirements
├── ReasonCodes
├── PolicyReference
├── EvaluatedAt
└── InputVersions
```

Для каждого Evidence:

```text
EvidenceAssessment
├── EvidenceReference
├── Outcome
├── ValidForPurpose
├── ReasonCodes
├── EffectiveValidity
├── Confidence
└── ConfirmationStatus
```

## Outcomes

- Valid
- Invalid
- Stale
- Restricted
- Disputed
- Insufficient
- NotApplicable
- Unknown

## Core Rules

- Evidence source exists or explainably unavailable.
- Source version matches reference.
- Subject matches expected Student.
- Scope compatible.
- Evidence type accepted by consumer Policy.
- Confirmation level sufficient.
- Evidence not invalidated.
- Evidence not expired for purpose.
- Privacy allows use.
- Disputed Evidence behavior explicit.
- AI inference remains unconfirmed unless validated.
- Notification interactions are excluded from learning Evidence by default.
- Missing Evidence not converted to negative.
- Imported Evidence includes provenance and temporal precision.
- Same Evidence MAY have different result for different purposes.
- Evaluator does not alter Evidence state.

## Reason Codes

- EVIDENCE_SOURCE_MISSING
- EVIDENCE_SOURCE_VERSION_MISMATCH
- EVIDENCE_SUBJECT_MISMATCH
- EVIDENCE_SCOPE_MISMATCH
- EVIDENCE_TYPE_NOT_ACCEPTED
- EVIDENCE_CONFIRMATION_INSUFFICIENT
- EVIDENCE_INVALIDATED
- EVIDENCE_EXPIRED
- EVIDENCE_STALE
- EVIDENCE_DISPUTED
- EVIDENCE_PRIVACY_RESTRICTED
- EVIDENCE_PROVENANCE_MISSING
- EVIDENCE_AI_UNCONFIRMED
- EVIDENCE_NOT_APPLICABLE

## AI Boundaries

AI может классифицировать candidate Evidence, но:

```text
AI classification != Evidence confirmation
```

## Tests

- valid confirmed evidence;
- wrong Student;
- wrong scope;
- expired;
- invalidated;
- disputed;
- missing provenance;
- AI proposal;
- imported approximate time;
- valid for one Policy but invalid for another;
- privacy restriction.

# Optional Domain Services

Следующие services не считаются canonical до отдельного Architecture Decision.

## HomeworkCompletionEvaluator

Может быть выделен отдельно, если completion logic существенно отличается от Homework Review и Aggregate становится перегружен.

Возможные inputs:

- latest Submission;
- completed Review;
- active Correction Request;
- blockers;
- completion policy.

До этого completion может оставаться внутри Homework Assignment с использованием approved Review Decision.

## PeriodicReviewEvaluator

Может определять, какие domain areas требуют Periodic Review.

Не должен:

- загружать Aggregates;
- создавать review tasks напрямую;
- считать отсутствие activity негативным результатом;
- генерировать educational judgment без evidence.

## NotificationBundlingEvaluator

Может использоваться для объединения нескольких intents.

Bundling не должно:

- смешивать разные privacy levels;
- скрывать critical item;
- нарушать expiry;
- создавать misleading summary;
- объединять сообщения разных recipients.

## ConcertProgramConstraintEvaluator

Может проверять:

- overlaps;
- stage capacity;
- duration;
- transition time;
- equipment constraints.

Если global schedule consistency становится authoritative, вероятно потребуется ConcertProgram Aggregate, а evaluator будет использоваться внутри его command flow.

## AchievementRevocationEvaluator

Revocation требует особенно строгой модели.

Она не должна быть автоматическим обратным действием при изменении текущего Progress.

Achievement history обычно сохраняется.

Revocation допустима только при:

- ошибочном Award;
- invalidated supporting facts;
- fraud/integrity process;
- policy-defined administrative correction.

Требуется отдельная Policy.

# Domain Snapshot Rules

Domain Service не должен получать mutable Aggregate object, если это позволяет случайно его изменить.

Предпочтительно использовать immutable snapshot.

```text
GoalSnapshot
├── GoalId
├── AggregateVersion
├── StudentId
├── Status
├── CriteriaVersion
├── CurrentProgress
├── EvidenceReferences
└── RelevantHistory
```

Snapshot должен содержать только данные, нужные evaluation.

# Snapshot Requirements

Каждый snapshot определяет:

- source Aggregate;
- source Aggregate Version;
- created/observed time;
- included fields;
- privacy level;
- snapshot schema version;
- freshness;
- authoritative source.

Snapshot не заменяет Aggregate для mutation.

# Policy Configuration Rules

Configuration, влияющая на Decision, должна быть:

- immutable в рамках version;
- published;
- referenced by ID/version;
- доступна для audit;
- included in result;
- not silently replaced by latest version.

# Service Versioning

## Behavioral Change

Service Version изменяется, если при одинаковых inputs может измениться output.

## Schema Change

Input/Output Contract Version меняется при изменении serialized contract.

## Refactoring

Внутренний refactoring без behavioral change не требует Domain Service Version change.

## Policy Change

Если behavior определяется Policy, должна измениться Policy Version.

Service Version MAY остаться прежней, если executor semantics не изменились.

# Compatibility Rules

- Historical Decision должен оставаться explainable.
- Старый Decision не пересчитывается автоматически.
- New Service Version не изменяет past Events.
- Re-evaluation создает новый EvaluationId и DecisionId.
- Superseding Decision references previous Decision.
- Consumers должны знать, какие outcome versions поддерживаются.

# Error Model

Domain Service возвращает один из:

- EvaluationResult
- DomainRejection
- UnableToEvaluate
- TechnicalFailure

## Domain Rejection

Корректный отрицательный результат.

Пример:

```text
Outcome: NotEligible
```

## Unable To Evaluate

Недостаточно данных или unsupported input.

## Technical Failure

Нарушение execution contract:

- malformed snapshot;
- unsupported schema;
- internal inconsistency;
- arithmetic failure;
- required configuration unavailable before invocation.

Repository/network failure не должен возникать внутри service, потому что external calls запрещены.

## Canonical Service Errors

- DOMAIN_SERVICE_INPUT_INVALID
- DOMAIN_SERVICE_INPUT_INCOMPLETE
- DOMAIN_SERVICE_INPUT_STALE
- DOMAIN_SERVICE_VERSION_UNSUPPORTED
- DOMAIN_SERVICE_POLICY_VERSION_UNSUPPORTED
- DOMAIN_SERVICE_CONFIGURATION_MISSING
- DOMAIN_SERVICE_CONFIGURATION_INVALID
- DOMAIN_SERVICE_SNAPSHOT_INCONSISTENT
- DOMAIN_SERVICE_EVIDENCE_INVALID
- DOMAIN_SERVICE_UNABLE_TO_EVALUATE
- DOMAIN_SERVICE_INTERNAL_INVARIANT_VIOLATION

# Decision Persistence

Domain Service не сохраняет Decision.

Application Service отвечает за:

- assigning persistence reference;
- storing Decision artifact;
- applying Decision to Aggregate;
- handling concurrency;
- atomicity where required;
- audit;
- event publication.

Possible flow:

1. Load inputs.
2. Evaluate.
3. Persist Decision artifact.
4. Load/revalidate target Aggregate.
5. Apply Command with DecisionReference.
6. Persist Aggregate and Outbox.

Если Decision и Aggregate mutation должны быть atomic, Application Layer определяет transaction model.

# Concurrency Rules

Domain Service не управляет optimistic concurrency, но result содержит input versions.

Aggregate при применении проверяет:

- target version;
- critical foreign input versions where necessary;
- Decision validity;
- invalidating facts.

При conflict:

- Reject

or

- Re-evaluate

согласно command contract.

# Evaluation Caching

Evaluation Result MAY кешироваться только если cache key включает:

- Service Version
- Policy Version
- Configuration Versions
- Input References
- Input Versions
- Evaluated Time Bucket — if allowed
- Purpose
- Tenant

Caching не должен скрывать staleness.

Critical Decisions лучше сохранять как explicit artifacts, а не только cache entry.

# Audit Rules

Для каждого authoritative evaluation сохраняются:

- EvaluationId
- Service Id
- Service Version
- Policy Reference
- Configuration References
- Input References
- Input Versions
- Evidence References
- Outcome
- Reason Codes
- Conditions
- Blocking Conditions
- EvaluatedAt
- RequestedBy
- CorrelationId
- CausationId
- AI Involvement

Full private input snapshot не обязан храниться в audit, если это нарушает data minimization.

# Privacy Rules

Domain Service:

- получает только необходимые fields;
- не логирует raw private content;
- возвращает references вместо full sensitive values;
- сохраняет output privacy classification;
- не понижает Privacy Level;
- не переносит private Teacher notes в Student-visible reasons;
- разделяет internal и public explanation.

# Explainability Rules

Каждый Decision должен позволять ответить:

- Какое правило применялось?
- Какая версия правила?
- Какие inputs использовались?
- Какие evidence были приняты?
- Какие evidence были отклонены?
- Какие criteria выполнены?
- Какие criteria не выполнены?
- Почему требуется Human Review?
- Когда Decision перестанет быть valid?
- Что вызовет reevaluation?

Объяснение не должно строиться только из свободного AI-generated текста.

# AI Rules

### AI-DS-001

AI не является Domain Service authority.

### AI-DS-002

AI MAY быть upstream advisory processor.

```text
Recording
   |
   v
AI Analysis Proposal
   |
   v
Validation / Human Review
   |
   v
Confirmed Evidence
   |
   v
Domain Service
```

### AI-DS-003

AI proposal сохраняет provenance.

### AI-DS-004

Domain Service не вызывает AI provider.

### AI-DS-005

AI confidence не заменяет Policy criteria.

### AI-DS-006

AI result с отсутствующим source не используется.

### AI-DS-007

Model update не должен скрыто менять authoritative behavior.

### AI-DS-008

AI-generated explanation маркируется как generated content и не заменяет structured Reason Codes.

### AI-DS-009

AI не может ослаблять:

- privacy;
- consent;
- evidence requirements;
- human review;
- safety blockers.

### AI-DS-010

Human correction AI proposal сохраняется для quality audit, если privacy позволяет.

# Testing Standard

Каждый Domain Service должен иметь следующие тестовые категории.

## Determinism Tests

- identical input produces identical output;
- input ordering does not affect result where order is semantically irrelevant;
- system clock is not used;
- locale does not affect machine outcome;
- random values are absent.

## Version Tests

- supported Policy Version;
- unsupported Policy Version;
- configuration version mismatch;
- service behavior version recorded;
- historical fixture remains reproducible.

## Input Tests

- required input missing;
- stale snapshot;
- tenant mismatch;
- subject mismatch;
- invalid reference;
- duplicate evidence;
- conflicting evidence.

## Boundary Tests

- exact deadline;
- exact validity expiration;
- minimum evidence threshold;
- maximum reminder count;
- status transition boundary;
- empty optional set;
- all optional conditions present.

## Negative Outcome Tests

- correct structured rejection;
- negative result is not technical error;
- Reason Codes stable;
- Human Review outcome distinct.

## Staleness Tests

- source Aggregate Version changed;
- Policy Version changed;
- Evidence invalidated;
- Song Version changed;
- Requirements Version changed;
- Consent withdrawn;
- Decision expired.

## Privacy Tests

- restricted input not copied into public output;
- Student-visible reasons separated;
- sensitive references preserved;
- no raw attachment URL;
- no contact destination leak.

## AI Tests

- AI proposal alone insufficient;
- unconfirmed AI Evidence rejected;
- AI cannot become DecisionSource;
- AI confidence not used as confirmed Evidence;
- human-approved AI-derived Evidence accepted according to policy.

## Property Tests

Where applicable:

- valid TimeWindow always preserves ordering;
- increasing evidence quantity does not automatically force positive outcome;
- removed blocker does not produce hidden negative state;
- reevaluation preserves historical Decision identity;
- same input permutations produce same accepted evidence set.

## Golden Decision Fixtures

Для critical Policies следует хранить canonical fixtures:

- Input Snapshot
- Policy Version
- Expected Outcome
- Expected Reason Codes
- Expected Conditions
- Expected Validity

Они используются для regression при изменении implementation.

# Service Review Checklist

- [ ] Service имеет одно доменное responsibility.
- [ ] Название выражает доменный смысл.
- [ ] Поведение не принадлежит одному Aggregate.
- [ ] Service stateless.
- [ ] Service deterministic.
- [ ] Все inputs явны.
- [ ] Evaluation time передается.
- [ ] Нет repositories.
- [ ] Нет network calls.
- [ ] Нет persistence.
- [ ] Нет Event publication.
- [ ] Нет Command dispatch.
- [ ] Output structured.
- [ ] Reason Codes определены.
- [ ] Policy Version определена.
- [ ] Service Version определена.
- [ ] Input versions сохраняются.
- [ ] Evidence provenance сохраняется.
- [ ] Stale behavior определен.
- [ ] Human Review outcome определен.
- [ ] Privacy определена.
- [ ] AI не имеет authority.
- [ ] Tests покрывают boundary и negative outcomes.

# Domain Service Implementation Readiness

Service готов к реализации, если определены:

- Service name.
- Responsibility.
- Reason it does not belong to one Aggregate.
- Inputs.
- Input snapshot schemas.
- Outputs.
- Domain outcomes.
- Reason Codes.
- Policy relationship.
- Service Version.
- Configuration Versions.
- Determinism rules.
- Staleness rules.
- Validity.
- Human Review behavior.
- Privacy.
- Audit.
- AI restrictions.
- Tests.
- Non-goals.

# Domain Service Non-Goals

Domain Service не должен:

- обслуживать HTTP;
- парсить request;
- проверять access token;
- открывать transaction;
- загружать repository;
- сохранять Aggregate;
- публиковать broker message;
- отправлять notification;
- выполнять retry;
- управлять scheduler;
- хранить workflow state;
- вызывать AI;
- быть generic helper;
- содержать unrelated methods;
- возвращать database entity;
- управлять DTO mapping за пределами domain snapshots.

# Service Placement Guidance

Логическая структура:

```text
domain/
  services/
    lesson-completion/
    progress-update/
    goal-completion/
    achievement-eligibility/
    song-readiness/
    concert-eligibility/
    homework-expiration/
    homework-reminder/
    notification-decision/
    evidence-validity/
```

Конкретная package structure определяется implementation architecture отдельно.

# Dependency Direction

```text
Value Objects
      ^
      |
Policies / Domain Services
      ^
      |
Aggregates
```

Возможна обратная композиция на уровне вызова:

```text
Application Service
   |
   +--> Domain Service
   |
   +--> Aggregate
```

Но Domain Service не зависит от repository или Application Service.

Domain Service MAY depend on:

- Value Object contracts;
- immutable Policy definitions;
- immutable domain snapshots;
- pure calculation components.

# Canonical Service Matrix

| Domain Service | Primary Input Domains | Primary Output |
| --- | --- | --- |
| LessonCompletionEvaluator | Lesson, Attendance, Teacher Assignment | Lesson Completion Decision |
| ProgressUpdateEvaluator | Progress, Evidence, Curriculum | Progress Update Decision |
| GoalCompletionEvaluator | Goal, Progress, Evidence | Goal Completion Decision |
| AchievementEligibilityEvaluator | Achievement Definition, Awards, Evidence | Achievement Eligibility Decision |
| SongReadinessEvaluator | Song Readiness, Areas, Evidence | Song Readiness Decision |
| ConcertEligibilityEvaluator | Concert, Participation, Consent, Readiness | Concert Eligibility Decision |
| HomeworkExpirationEvaluator | Homework, Submission, Blocker, Correction | Homework Expiration Decision |
| HomeworkReminderEvaluator | Homework, Reminder Plan, Preferences | Reminder Decision |
| NotificationDecisionEvaluator | Notification Intent, Recipient, Privacy | Notification Decision |
| EvidenceValidityEvaluator | Evidence, Consumer Policy, Source State | Evidence Validity Result |

# Command Relationship Matrix

| Service Decision | Command Applying Decision |
| --- | --- |
| Lesson Completion Decision | CompleteLesson |
| Progress Update Decision | UpdateProgress |
| Goal Completion Decision | CompleteGoal |
| Achievement Eligibility Decision | AwardAchievement |
| Song Readiness Decision | ChangeSongReadiness |
| Concert Eligibility Decision | MarkConcertParticipationEligible / Conditional / NotEligible |
| Homework Expiration Decision | ExpireHomework / StartHomeworkGracePeriod |
| Homework Reminder Decision | Schedule / Reschedule / Suppress Homework Reminder |
| Notification Decision | Create / Approve / Suppress Notification Intent |
| Evidence Validity Result | Used by another evaluator; usually no direct mutation |

# Event Relationship Matrix

| Applied Decision | Expected Event |
| --- | --- |
| Lesson Completion | LessonCompleted |
| Progress Update | ProgressUpdated |
| Goal Completion | GoalCompleted |
| Achievement Award | AchievementAwarded |
| Song Readiness Change | SongReadinessChanged |
| Concert Eligibility | ConcertEligibilityEvaluated |
| Homework Expiration | HomeworkExpired |
| Reminder Schedule | HomeworkReminderScheduled |
| Notification Approval | NotificationIntentApproved |
| Evidence Invalidation | Event belongs to source Aggregate, not evaluator |

# Failure Modes

## Hidden Repository Access

Service behavior зависит от live database и становится нерепродуцируемым.

## Latest Configuration Lookup

Historical Decision невозможно восстановить.

## Boolean Decision

Невозможно понять причины и conditions.

## Missing Input Versions

Stale Decision применяется к изменившемуся Aggregate.

## AI Authority Leakage

AI proposal становится final Decision.

## Policy in Application Handler

Одинаковый rule дублируется в нескольких workflows.

## Aggregate Bypass

Service напрямую сохраняет state.

## Excessive Shared Service

Один evaluator содержит все образовательные rules.

## Snapshot Overload

Service получает полный graph всех Aggregates вместо минимальных inputs.

## Negative Outcome as Exception

Обычное NotEligible логируется как system failure.

## Event Publication from Service

Event публикуется до Aggregate transition и transaction commit.

# Non-Goals

Этот документ не определяет:

- конкретный язык реализации;
- Go interfaces;
- package names;
- dependency injection framework;
- database schema;
- message broker;
- service deployment;
- microservice boundaries;
- API endpoints;
- exact policy algorithms;
- scoring formulas;
- curriculum levels;
- concert capacity model;
- notification provider;
- AI model;
- CRM behavior;
- billing;
- payments;
- payroll;
- marketing communication.

# Open Questions

Необходимо определить:

- какие evaluators действительно нужны в MVP;
- должен ли LessonCompletionEvaluator быть отдельным сервисом или behavior Lesson Aggregate;
- нужен ли отдельный HomeworkCompletionEvaluator;
- являются ли Policies отдельными artifacts или embedded definitions services;
- какой contract использовать для immutable snapshots;
- сохранять ли полный Decision artifact;
- где хранить Decision artifacts;
- какие Decisions должны быть atomic с Aggregate mutation;
- какие Decisions допускают delayed application;
- какой default ValidUntil использовать;
- как объявлять invalidating events;
- как выполнять automatic reevaluation;
- нужен ли central Evaluation Registry;
- нужен ли schema registry для evaluator inputs;
- как version service behavior;
- как связать Service Version и Policy Version;
- как обновлять golden fixtures;
- кто утверждает Policy behavioral changes;
- нужен ли DSL для criteria;
- какие criteria могут быть configuration-driven;
- какие rules должны оставаться code-defined;
- как предотвратить arbitrary scripting;
- нужен ли explainability renderer;
- где хранить localized Decision explanations;
- какие Reason Codes показывать Student;
- какие Reason Codes доступны только Teacher;
- какие evaluations требуют Human Review;
- как назначать Human Reviewer;
- нужен ли Review Process Manager;
- может ли один Decision иметь несколько authorized reviewers;
- нужен ли approval quorum;
- как импортировать historical Decisions;
- как обозначать Decisions с incomplete historical inputs;
- как моделировать approximate timestamps;
- какие Evidence Types принимать ProgressUpdateEvaluator;
- нужен ли отдельный Evidence Registry;
- как разрешать conflicting Teacher Evidence;
- должен ли Progress Update допускать negative movement;
- как моделировать Progress confidence;
- как version Goal criteria;
- могут ли Goal criteria ссылаться на future evidence;
- как поддержать composite Goals;
- какие Achievements repeatable;
- как проверять duplicate awards;
- можно ли Award после retired definition;
- как моделировать revoked Achievement;
- какие Song Readiness Areas обязательны;
- требуется ли Human Review для Safety Area;
- как вычислять ValidUntil Song Readiness;
- как связывать readiness с performance date;
- какие Concert Requirements входят в MVP;
- отделяется ли eligibility от consent;
- должен ли absence of consent приводить к NotEligible или Deferred;
- когда organizational capacity влияет на Approval, а не Eligibility;
- нужен ли ConcertProgramConstraintEvaluator;
- как моделировать duet/group eligibility;
- как оценивать ensemble readiness;
- как моделировать Homework deadline before next Lesson;
- как Active Blocker влияет на Expiration;
- какой maximum Grace Period;
- кто может override Expiration Decision;
- сколько reminders допускается;
- как учитывать Student age и Guardian preferences;
- какие notifications обязательны;
- нужен ли NotificationBundlingEvaluator;
- какие Privacy Levels разрешены для каждого channel;
- как выполнять send-time revalidation;
- какие Evidence validity rules общие;
- какие rules принадлежат consumer Policy;
- может ли Evidence быть valid без Source Aggregate;
- как обрабатывать удаленный source;
- как хранить disputed Evidence;
- кто может invalidate Evidence;
- какие AI analyses могут стать confirmed Evidence;
- нужен ли human approval для каждого AI-derived evidence;
- как сохранять AI Model Version;
- как тестировать AI model regression отдельно от deterministic evaluator;
- какие evaluations выполняются synchronously;
- какие должны запускаться Process Manager;
- как обрабатывать дорогие evaluations;
- допускается ли evaluation cache;
- какой cache retention;
- как предотвращать stale cached Decision;
- нужна ли digital signature Decisions;
- какие architecture tests должны запрещать repository dependency;
- как автоматически проверять deterministic behavior;
- нужны ли property-based tests;
- какие services могут быть реализованы как pure functions;
- какие Domain Services могут быть объединены без нарушения responsibility.

# History

| Version | Description |
| --- | --- |
| 1.0.0 | Определен канонический каталог чистых Domain Services Belcanto Product, включая правила детерминизма, versioning, snapshot inputs, Decisions, Evidence, privacy, AI boundaries, testing и десять основных evaluators. |
