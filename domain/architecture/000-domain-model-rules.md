---
Status: Approved
Version: 1.0.0
Last Updated: 2026-07-27

Document Id: DOMAIN_MODEL_RULES

Document Type:
  - Domain Architecture Constitution
  - Modeling Standard
  - Consistency Rules
  - Design Review Standard
  - Implementation Governance

Owners:
  - Product Owner
  - Education Lead
  - Domain Architecture Lead
  - Technical Lead

Applies To:
  - Ubiquitous Language
  - Aggregate Roots
  - Entities
  - Value Objects
  - Domain Services
  - Policies
  - Decisions
  - Commands
  - Domain Events
  - Process Managers
  - Projections
  - Repositories
  - Application Services
  - Integrations
  - AI-Assisted Operations
  - Domain Tests

Related Directories:
  - ../aggregates/
  - ../commands/
  - ../events/
  - ../policies/
  - ../value-objects/
  - ../entities/
  - ../services/
  - ../processes/
  - ../../product/
  - ../../school/

Related Documents:
  - ../aggregates/000-aggregate-catalog.md
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

# Domain Model Rules

> Этот документ является конституцией доменной модели Belcanto Product.
>
> Он определяет обязательные правила проектирования, изменения, реализации и проверки доменных компонентов.
>
> Любой новый Aggregate, Entity, Value Object, Policy, Command, Event, Domain Service, Process Manager или Projection должен соответствовать этим правилам.

---

# Purpose

Belcanto Product строится как learning product, а не как набор CRUD-экранов и таблиц.

Система должна сохранять смысл образовательных процессов:

- Lesson действительно завершен;
- Homework действительно назначено и имеет понятный lifecycle;
- Progress подтвержден evidence;
- Goal завершена по критериям;
- Achievement присуждено обоснованно;
- Song Readiness относится к конкретной версии песни и типу выступления;
- Concert Eligibility не смешивается с административным Approval;
- Reminder не превращается в давление;
- Notification не становится источником образовательной истины;
- AI не становится скрытым владельцем решения.

Без общих архитектурных правил модель постепенно деградирует:

```text
Explicit Domain Model
        |
        v
Generic Services
        |
        v
Status Flags
        |
        v
CRUD
        |
        v
Business Rules in Controllers and SQL
```

Этот документ предотвращает такую деградацию.

---

# Normative Language

В документе используются следующие значения:

- MUST — обязательное правило;
- MUST NOT — запрещенное действие;
- SHOULD — рекомендуемое правило, отклонение требует причины;
- SHOULD NOT — обычно запрещено, исключение требует обоснования;
- MAY — допустимый вариант;
- REQUIRES DECISION — изменение требует отдельного архитектурного решения.

---

# Architectural Priorities

При конфликте проектных целей применяется следующий приоритет:

1. Educational correctness
2. Student safety and dignity
3. Domain truth
4. Privacy and authorization
5. Explainability and auditability
6. Consistency of terminology
7. Correct aggregate boundaries
8. Operational reliability
9. Implementation simplicity
10. Performance optimization
11. Developer convenience

Производительность не должна достигаться путем скрытого нарушения доменных инвариантов.

Удобство UI не должно определять Aggregate boundary.

---

# Core Domain Flow

Канонический путь изменения состояния:

```text
Actor / Policy / Scheduler / Integration
                    |
                    v
                 Command
                    |
                    v
            Application Service
                    |
                    +--> authentication
                    +--> authorization
                    +--> idempotency
                    +--> load aggregate
                    +--> load approved decision/snapshots
                    |
                    v
              Aggregate Root
                    |
                    +--> validates owned invariants
                    +--> changes state
                    +--> emits Domain Events
                    |
                    v
       State + Outbox persisted atomically
```

Для сложного решения:

```text
Current State
+
Evidence
+
Configuration
+
Evaluation Time
        |
        v
      Policy
        |
        v
     Decision
        |
        v
     Command
        |
        v
   Aggregate Root
        |
        v
   Domain Event
```

---

# Modeling Layers

## Domain Layer

Содержит:

- Aggregate Roots;
- Entities;
- Value Objects;
- Policies;
- Decisions;
- Domain Services;
- domain invariants;
- Domain Events;
- domain-specific Reason Codes.

Domain Layer не должен зависеть от:

- HTTP;
- database driver;
- message broker;
- UI;
- framework;
- provider SDK;
- system clock;
- environment variables;
- application container;
- external network.

## Application Layer

Оркестрирует use case.

Содержит:

- Command Handlers;
- Application Services;
- Query Handlers;
- Process Managers;
- transaction coordination;
- repository calls;
- authorization orchestration;
- idempotency;
- Outbox coordination;
- integration mapping.

Application Layer не должен создавать доменную истину в обход Domain Layer.

## Infrastructure Layer

Содержит:

- repository implementation;
- database;
- broker;
- Outbox publisher;
- email/SMS/push providers;
- file storage;
- external API adapters;
- clock implementation;
- identifier generator;
- telemetry.

Infrastructure Layer не определяет бизнес-правила.

## Presentation Layer

Содержит:

- API;
- mobile client;
- web interface;
- staff interface;
- input mapping;
- human-readable error rendering.

Presentation Layer не изменяет Aggregate напрямую.

---

# Ubiquitous Language

## UL-001: Canonical terms are mandatory

Во всех спецификациях, командах, событиях, коде и интерфейсах должны использоваться утвержденные термины.

Канонические примеры:

```text
Student
Teacher
Lesson
Homework Assignment
Submission
Homework Review
Progress Evidence
Goal
Achievement Definition
Achievement Award
Student Song
Song Version
Song Readiness
Concert
Concert Participation
Concert Eligibility
Approval
Performance Slot
Reminder Plan
Notification Intent
Notification Delivery
Periodic Review
```

## UL-002: Similar concepts must remain distinct

Следующие понятия нельзя смешивать:

- Eligibility != Approval
- Approval != Program Placement
- Program Placement != Slot Assignment
- Homework Submitted != Homework Completed
- Notification Delivered != Notification Opened
- Notification Opened != Action Completed
- Reminder Delivered != Homework Started
- Lesson Completed != Progress Updated
- Goal Completed != Achievement Awarded
- Song Ready != Concert Participation Approved
- Cancelled != Expired
- Expired != Archived
- Reopened != Status Reset

## UL-003: Avoid generic business words

Следует избегать доменных имен:

```text
Item
Object
Record
Data
EntityData
ProcessData
StatusManager
CommonService
BusinessService
UpdateRequest
Action
HandlerService
```

Если термин не выражает доменный смысл, он не должен становиться частью публичной модели.

## UL-004: One term must have one meaning

Например, Review не должно одновременно означать:

- Teacher review Homework;
- Periodic system review;
- moderation;
- product feedback.

Используются уточненные имена:

```text
HomeworkReview
GoalReview
SongReadinessReview
PeriodicReview
DomainIntegrityReview
```

## UL-005: Translation must preserve semantics

Русский UI может отображать локализованный текст, но внутренний канонический термин должен сохранять один смысл.

Например:

Homework Assignment

может отображаться как:

Домашнее задание

Но в модели оно не должно случайно превратиться в Task, если Task имеет другое значение.

## UL-006: Renaming requires compatibility review

Изменение канонического термина требует проверки:

- commands;
- events;
- payload;
- aggregate names;
- policies;
- projections;
- analytics;
- external integration contracts;
- historical documentation.

---

# Aggregate Rules

## AG-001: Aggregate is a consistency boundary

Aggregate определяется инвариантами, а не таблицами, UI-страницей или названием сервиса.

## AG-002: Every Aggregate has one Root

Все изменения проходят через Aggregate Root.

## AG-003: Internal entities are not mutated independently

Запрещено:

```text
repository.SaveHomeworkSubmission(submission)
```

если Submission является внутренней Entity HomeworkAssignment.

Корректно:

```text
homework.Submit(...)
repository.Save(homework)
```

## AG-004: Aggregate owns only its own invariants

Lesson может проверить собственный lifecycle, но не может самостоятельно решить, что Student достиг Goal.

## AG-005: Aggregate references foreign aggregates by identity

Допустимо:

```text
StudentId
GoalId
LessonId
SongVersionId
ConcertId
```

Не допускается хранение mutable foreign Aggregate object.

## AG-006: Aggregate must be deterministic

Результат зависит только от:

- текущего состояния;
- команды;
- переданных Value Objects;
- переданного времени;
- переданного Policy Decision;
- переданных immutable snapshots.

Aggregate не должен скрыто обращаться к внешнему миру.

## AG-007: Aggregate must not call repositories

Aggregate не загружает и не сохраняет себя самостоятельно.

## AG-008: Aggregate must not call external services

Запрещены внутри Aggregate:

- HTTP;
- broker publication;
- email;
- file storage;
- AI model;
- database query;
- system environment;
- push provider.

## AG-009: Aggregate must not read global time

Неправильно:

```text
time.Now()
```

внутри доменного метода.

Корректно:

```text
homework.Expire(decision, evaluatedAt)
```

## AG-010: Aggregate methods express domain actions

Предпочтительно:

```text
lesson.Complete(...)
homework.Submit(...)
goal.Reopen(...)
participation.AssignPerformanceSlot(...)
```

Запрещенный публичный API:

```text
SetStatus(...)
SetField(...)
UpdateData(...)
Patch(...)
Save(...)
```

## AG-011: Generic status mutation is prohibited

Статус является следствием доменного перехода.

Нельзя выполнять:

```text
homework.SetStatus("completed")
```

Нужно:

```text
homework.Complete(completionDecision, completedAt)
```

## AG-012: Aggregate validates lifecycle transitions

Каждый переход должен быть явно разрешен.

Пример:

```text
Assigned -> In Progress
Assigned -> Submitted
Submitted -> Under Review
Under Review -> Correction Requested
Under Review -> Completed
Overdue -> Submitted
Expired -> Reopened
```

Неизвестный переход отклоняется.

## AG-013: Terminal history is immutable

При Reopen старый terminal record сохраняется.

Нельзя просто очистить:

```text
ExpiredAt = null
Status = Assigned
```

Нужно создать:

```text
ReopenRecord
New Aggregate Version
HomeworkReopened event
```

## AG-014: Aggregate emits events only for facts

Event создается после принятого изменения.

## AG-015: Rejected operation does not create success event

При отклоненном CompleteGoal событие GoalCompleted не создается.

## AG-016: No state change usually means no mutation event

Повторная идемпотентная команда может вернуть Already Processed или No Change Required.

## AG-017: Aggregate version is monotonic

Версия не уменьшается и не переиспользуется.

## AG-018: Aggregate must remain reasonably small

Неограниченные коллекции должны быть вынесены:

- delivery attempt history;
- full audit log;
- thousands of reminders;
- all historical assessments;
- large attachment content.

## AG-019: Aggregate cannot be designed as a report

Aggregate содержит authoritative state, а не удобный UI snapshot всех связанных данных.

## AG-020: Cross-aggregate mutation is prohibited

Один Root не меняет другой Aggregate.

## AG-021: Cross-aggregate invariant requires explicit coordination

Используются:

- Policy;
- Domain Service;
- Process Manager;
- reservation;
- eventual consistency;
- compensation.

## AG-022: Aggregate cannot silently trust foreign state

Если решение зависит от Song Readiness, необходимо сохранить:

```text
SongReadinessId
SongReadinessVersion
EvaluationId
EvaluatedAt
```

## AG-023: Stale decisions cannot be applied

Policy Decision должна быть связана с версиями входного состояния.

## AG-024: Aggregate persistence format is not domain API

Database columns не должны определять публичный доменный интерфейс.

## AG-025: Archive is not delete

Архивный переход является отдельным доменным действием.

---

# Entity Rules

## EN-001: Entity has identity

Entity отличается от другой Entity идентификатором, даже если остальные поля равны.

## EN-002: Entity lifecycle belongs to its Aggregate

Внутренняя Entity не должна иметь независимый Application Service без отдельного решения о выделении Aggregate.

## EN-003: Entity identity scope must be explicit

Идентификатор может быть:

- глобальным;
- уникальным внутри Aggregate.

Это должно быть зафиксировано.

## EN-004: Entity cannot escape as mutable object

Repository или API не должны возвращать внутренний mutable объект для изменения вне Root.

## EN-005: Entity transitions use domain methods

Например:

```text
submission.Withdraw(...)
review.Complete(...)
blocker.Resolve(...)
```

вызванные и контролируемые Aggregate Root.

## EN-006: Entity history may be externalized

Если collection становится большой, исторические записи могут храниться отдельно, но Root должен сохранять необходимые authoritative references.

---

# Value Object Rules

## VO-001: Value Object has no independent identity

Value Object определяется набором значений.

## VO-002: Value Object is immutable

После создания значения не изменяются.

Изменение означает создание нового Value Object.

## VO-003: Equality is by value

Два TimeWindow равны, если равны их значения и semantics.

## VO-004: Value Object validates itself at creation

Нельзя создать:

- неизвестный Timezone;
- отрицательный Duration;
- некорректный TimeWindow;
- пустой required identifier;
- invalid PrivacyLevel.

## VO-005: Invalid Value Object must not exist

Предпочтительно отклонить создание, а не распространять valid=false.

## VO-006: Value Object may contain behavior

Value Object не обязан быть пассивной структурой.

Примеры:

```text
timeWindow.Contains(moment)
gracePeriod.HasEnded(at)
deliveryWindow.NextAllowedTime(at)
version.IsExpected(current)
```

## VO-007: Value Object behavior is deterministic

Он не обращается к внешнему миру.

## VO-008: Serialization is not the domain definition

JSON или database representation может отличаться от внутренней реализации, но смысл должен сохраняться.

## VO-009: Primitive obsession should be avoided

Вместо:

```text
timezone string
privacy string
version int
dueDate time
```

предпочтительно:

```text
Timezone
PrivacyLevel
AggregateVersion
DueDate
```

## VO-010: Do not create meaningless wrappers

Value Object оправдан, если он добавляет:

- validation;
- semantics;
- behavior;
- type safety;
- privacy classification;
- formatting rules.

---

# Policy Rules

## PO-001: Policy evaluates, but does not mutate Aggregate

Policy получает факты и возвращает Decision.

```text
PolicyInput
    |
    v
Policy
    |
    v
Decision
```

## PO-002: Policy must be deterministic

Одинаковый input и Policy Version дают одинаковый Decision.

## PO-003: Policy has an explicit version

Каждое значимое решение сохраняет Policy Version.

## PO-004: Policy input must be explicit

Нельзя скрыто читать:

- database;
- clock;
- configuration service;
- AI;
- current user;
- global variable.

## PO-005: Policy returns structured Decision

Decision должна содержать:

```text
DecisionId
Outcome
PolicyId
PolicyVersion
EvaluatedAt
InputReferences
InputVersions
ReasonCodes
Conditions
BlockingConditions
EvidenceReferences
ValidUntil
HumanReviewRequired
```

## PO-006: Policy Decision is immutable

После использования Decision не переписывается.

При исправлении создается новая Decision.

## PO-007: Policy does not send notifications

Она может решить:

```text
Outcome: Request Human Review
```

или создать основание для Notification Intent, но не обращается к provider.

## PO-008: Policy does not execute commands

Application Service или Process Manager интерпретирует Decision и отправляет разрешенную Command.

## PO-009: Policy must explain negative and positive outcomes

Не только отказ, но и принятие должны быть обоснованы.

## PO-010: Policy cannot infer missing facts as truth

Отсутствие Evidence не равно отрицательному Evidence.

## PO-011: AI cannot replace deterministic policy

AI может помочь классифицировать input или подготовить proposal, но authoritative Decision принадлежит утвержденной Policy или Human Review.

## PO-012: Policy must define stale-input behavior

Если input устарел, результат:

- Re-evaluate
- Defer
- Human Review
- Reject stale decision

## PO-013: Policy owns its Reason Codes

Reason Code имеет стабильный смысл внутри Policy.

## PO-014: Policy must distinguish conditions and blockers

Condition

означает требование, при котором ограниченное действие допустимо.

Blocking Condition

означает препятствие, не позволяющее применить положительный outcome.

---

# Decision Rules

## DE-001: Decision is a first-class immutable artifact

Решение нельзя хранить только как лог или boolean.

## DE-002: Decision must reference exact inputs

Для каждого значимого источника сохраняются:

- type;
- id;
- version;
- timestamp;
- evidence reference.

## DE-003: Decision outcome must be canonical

Не допускается произвольный текст вместо структурированного результата.

## DE-004: Decision expiration must be explicit

Если решение может устареть, используется:

```text
ValidUntil
ReevaluationTrigger
```

## DE-005: Human and policy decisions must be distinguishable

```text
DecisionSource:
  DeterministicPolicy
  AuthorizedHuman
  ApprovedMigration
```

AI не является authoritative source.

## DE-006: Overridden decision preserves both records

При override сохраняются:

- original Decision;
- override Decision;
- Actor;
- reason;
- authority;
- time;
- affected command.

## DE-007: Decision cannot mutate state by itself

Изменение выполняется Command через Aggregate.

---

# Command Rules

## CO-001: Command expresses intent

Название используется в повелительной форме:

```text
CompleteLesson
SubmitHomework
EvaluateGoalCompletion
```

## CO-002: Command has one primary target

Multi-target команда допустима только как явно описанный batch/workflow contract.

## CO-003: Command has an authoritative Actor

Actor не определяется только полем payload.

## CO-004: Command must carry tenant context

## CO-005: User-driven mutations usually require ExpectedAggregateVersion

Исключения документируются.

## CO-006: Technical retry preserves CommandId

## CO-007: Changed intent requires a new CommandId

## CO-008: IdempotencyKey cannot be reused with different payload

Это считается Conflict или Security issue.

## CO-009: Generic update commands are restricted

Команда UpdateHomework не должна скрывать:

- Expire;
- Cancel;
- Replace;
- Complete;
- Reopen;
- Change Due Date.

## CO-010: Authorization precedes mutation

## CO-011: Authorization does not replace domain validation

Teacher может иметь право отправить команду, но текущее состояние может запрещать ее выполнение.

## CO-012: Command must not carry authoritative foreign snapshots from client

Client не может доказать текущий Progress или Eligibility, передав JSON.

## CO-013: Scheduled command does not automatically own decision authority

Scheduler запускает evaluation, а не принимает педагогическое решение.

## CO-014: Policy-generated command requires DecisionReference

## CO-015: Command result must distinguish rejection and failure

```text
Rejected = domain/security/validation reason
Failed = technical processing failure
```

## CO-016: Batch commands preserve per-item validation

## CO-017: Command must be auditable

## CO-018: UI action is not necessarily one command

Одна кнопка MAY запустить Application Service, который создаст последовательность команд, но границы и результаты должны оставаться явными.

---

# Domain Event Rules

## EV-001: Event describes a past fact

```text
LessonCompleted
HomeworkSubmitted
GoalReopened
```

## EV-002: Event is immutable

## EV-003: Event is owned by one producer domain

Consumer не расширяет смысл Event.

## EV-004: Event contains minimal sufficient payload

Не следует публиковать полный Aggregate snapshot.

## EV-005: Event must not contain secrets

## EV-006: Event references command and causation

Если применимо:

```text
CommandId
CorrelationId
CausationId
```

## EV-007: Event contains Aggregate Version

## EV-008: Event publication follows successful commit

## EV-009: State and Outbox are persisted atomically

## EV-010: At-least-once delivery is assumed

Consumers MUST be idempotent.

## EV-011: Global ordering is not assumed

Порядок гарантируется только в заявленном scope.

## EV-012: Event cannot authorize mutation

Получение Event не означает, что consumer вправе изменить Aggregate без собственной проверки.

## EV-013: Event consumer sends a Command

Consumer не выполняет raw update чужого состояния.

## EV-014: Domain Event and Integration Event are different contracts

## EV-015: Replay must suppress external side effects

При replay нельзя автоматически повторять:

- email;
- SMS;
- push;
- external provider call;
- human escalation;
- reward delivery.

## EV-016: Correction uses a new Event

Старое событие не переписывается.

## EV-017: CRUD events are prohibited

Нежелательные события:

```text
EntityUpdated
RecordChanged
DataSaved
StatusModified
```

Нужен доменный факт.

## EV-018: Event cycles must be prevented

Система должна обнаруживать и останавливать бесконечные реакции.

## EV-019: Event time semantics are explicit

```text
OccurredAt != RecordedAt
```

## EV-020: Event is not an audit dump

Audit может содержать дополнительные технические данные отдельно.

---

# Domain Service Rules

## DS-001: Domain Service is used only when behavior belongs to the domain but not naturally to one Aggregate or Value Object

## DS-002: Domain Service must have a domain-specific name

Предпочтительно:

```text
ConcertEligibilityEvaluator
HomeworkExpirationEvaluator
GoalCompletionEvaluator
```

Нежелательно:

```text
DomainService
CommonService
BusinessService
Manager
Helper
```

## DS-003: Domain Service must remain stateless unless explicitly modeled as an Aggregate or Process Manager

## DS-004: Domain Service must be deterministic

## DS-005: Domain Service does not persist state

## DS-006: Domain Service does not call infrastructure

## DS-007: Domain Service does not become a dumping ground

Если Service содержит все бизнес-правила, а Aggregate является DTO, модель признана анемичной.

## DS-008: Cross-aggregate evaluation may use a Domain Service

Он получает prepared snapshots и возвращает Decision или calculation result.

## DS-009: Orchestration belongs to Application Service or Process Manager

Domain Service не управляет retries, transactions и broker delivery.

---

# Application Service Rules

## AS-001: Application Service coordinates one use case

## AS-002: Application Service may load multiple sources

Он может загрузить:

- target Aggregate;
- Policy configuration;
- evidence snapshots;
- authorization context;
- current time;
- Decision.

## AS-003: Application Service must not reimplement aggregate invariants

## AS-004: Application Service must not decide domain truth through conditionals that belong to Policy

Неправильно:

```text
if score > 80:
    mark goal completed
```

Корректно:

```text
decision = goalCompletionPolicy.Evaluate(input)
goal.Complete(decision)
```

## AS-005: Transaction scope is explicit

## AS-006: Application Service persists through repositories

## AS-007: Application Service handles idempotency and ExpectedVersion

## AS-008: Application Service maps domain result to external response

## AS-009: Application Service does not expose infrastructure errors as domain errors

---

# Process Manager Rules

## PM-001: Process Manager coordinates a long-running or multi-aggregate process

## PM-002: Process Manager owns workflow state, not educational truth

## PM-003: Process Manager reacts to Events and sends Commands

```text
Event
  |
  v
Process Manager
  |
  v
Command
```

## PM-004: Process Manager does not mutate Aggregate storage

## PM-005: Process Manager does not duplicate Aggregate invariants

## PM-006: Process Manager is idempotent

## PM-007: Process Manager preserves correlation and causation

## PM-008: Process Manager handles timeouts explicitly

## PM-009: Process Manager defines terminal process states

Пример:

```text
Running
Waiting
Completed
Failed
Cancelled
Compensating
```

## PM-010: Process Manager prevents duplicate commands

## PM-011: Process Manager must survive restart

Workflow state не должен существовать только в memory.

## PM-012: Process Manager supports compensation where required

Compensation не означает database rollback прошлого доменного факта.

## PM-013: Process Manager must not create hidden synchronous distributed transactions

## PM-014: Every Process Manager needs a documented responsibility

Кандидаты:

- Lesson Completion Process;
- Homework Lifecycle Process;
- Goal Completion Process;
- Concert Preparation Process;
- Notification Delivery Process;
- Periodic Review Process.

---

# Repository Rules

## RE-001: Repository is defined per Aggregate Root

## RE-002: Repository loads and persists Aggregate Root

Он не возвращает Projection вместо Aggregate.

## RE-003: Repository does not contain domain decisions

## RE-004: Repository does not expose generic table operations to Domain Layer

## RE-005: Save requires ExpectedVersion or equivalent concurrency protection

## RE-006: Repository persists state and pending events atomically

## RE-007: Repository does not publish events directly before commit

## RE-008: Repository should not return partially initialized Aggregate

## RE-009: Internal entities are not persisted through public independent repositories unless they become Aggregate Roots

## RE-010: Query repositories and domain repositories are distinct concepts

Read Projection repository MAY provide denormalized data.

Aggregate repository MUST return authoritative domain state.

## RE-011: Repository methods use domain language where useful

Предпочтительно:

```text
LoadHomeworkAssignment
SaveHomeworkAssignment
```

Необязательно создавать множество domain-specific search methods внутри Aggregate repository.

## RE-012: Bulk SQL must not bypass domain behavior for active business mutations

Bulk repair требует:

- approved migration;
- dry run;
- audit;
- provenance;
- validation;
- rollback or compensation plan.

---

# Projection Rules

## PRJ-001: Projection is a read model

Она не является Aggregate.

## PRJ-002: Projection may combine several Aggregates

## PRJ-003: Projection may be eventually consistent

## PRJ-004: Projection is disposable and rebuildable

## PRJ-005: Projection is not authoritative for mutations

Нельзя принимать критическое решение только потому, что dashboard показывает определенный status, без загрузки authoritative state.

## PRJ-006: Projection must expose freshness where relevant

Например:

```text
LastUpdatedAt
SourceVersion
ProjectionLag
```

## PRJ-007: Projection access requires authorization

Denormalization не отменяет privacy boundaries.

## PRJ-008: Projection should be designed for a use case

Примеры:

```text
StudentLearningDashboard
TeacherStudentOverview
HomeworkReviewQueue
ConcertPreparationBoard
OwnerOperationalDashboard
```

## PRJ-009: Projection failure does not change domain truth

## PRJ-010: Projection rebuild must not trigger domain side effects

---

# Command-Query Separation

## CQ-001: Command may change state

## CQ-002: Query must not change domain state

## CQ-003: Query may update purely technical cache only if it does not create a domain fact

## CQ-004: Query must not emit Domain Events

## CQ-005: Command response may include limited result data

Но не должна превращаться в произвольный reporting query.

## CQ-006: UI uses Queries for display and Commands for actions

## CQ-007: Query model may lag behind successful Command

Интерфейс должен уметь показать:

- Processing
- Pending projection update

при необходимости.

---

# Consistency Rules

## CS-001: Strong consistency is guaranteed inside one Aggregate

## CS-002: Cross-aggregate consistency is eventual by default

## CS-003: Eventual consistency must be explicit

Для процесса определяются:

- expected delay;
- intermediate state;
- retry;
- failure;
- user-visible behavior;
- compensation;
- monitoring.

## CS-004: Distributed transaction is not default

## CS-005: Critical uniqueness may require reservation

Пример:

- Concert performance slot;
- limited capacity;
- one active Primary Teacher assignment.

## CS-006: Compensation creates new facts

Например:

```text
ConcertPerformanceSlotAssigned
ConcertPerformanceSlotRemoved
```

Второе событие не удаляет первое.

## CS-007: Stale reads must be handled

## CS-008: Version conflict is a normal domain/application outcome

Это не всегда internal server error.

## CS-009: Read-your-write behavior must be defined per interface

## CS-010: Recovery cannot fabricate success

Если результат неизвестен, система должна проверить idempotency record или authoritative state.

---

# Time Rules

## TI-001: Time is explicit domain input

## TI-002: All persisted timestamps include timezone semantics

Предпочтительно хранить instant в UTC и отдельно сохранять relevant timezone where business meaning depends on it.

## TI-003: Local date and instant are not interchangeable

Например:

```text
DueDate as local school date
ScheduledFor as exact instant
```

## TI-004: OccurredAt and RecordedAt are distinct

## TI-005: Client time is not trusted automatically

## TI-006: Deadline evaluation uses explicit evaluation time

## TI-007: Quiet Hours use recipient timezone

## TI-008: Timezone changes trigger recalculation where required

## TI-009: Daylight saving and timezone rule changes must be supported

Даже если текущий tenant не использует DST.

## TI-010: Historical time meaning must not be recalculated silently

## TI-011: Future EffectiveAt requires explicit semantics

## TI-012: Clock is an application/infrastructure dependency

Domain receives value, not Clock service call.

---

# Identity Rules

## ID-001: Identifiers are typed by domain meaning

```text
StudentId
LessonId
GoalId
```

не должны случайно смешиваться.

## ID-002: Transport identifiers are not domain identifiers unless explicitly mapped

## ID-003: Provider reference is not AggregateId

## ID-004: Database primary key representation does not define external contract

## ID-005: Identifier reuse is prohibited

## ID-006: Historical reference remains resolvable or explainably unavailable

## ID-007: Tenant context must not be inferred solely from identifier

---

# Evidence Rules

## EVI-001: Evidence is a reference to a confirmed or classified source fact

## EVI-002: Evidence ownership remains with source domain

## EVI-003: Evidence acceptance is contextual

One evidence item may be valid for one Policy and insufficient for another.

## EVI-004: Evidence includes provenance

```text
SourceDomain
SourceEntityId
SourceEntityVersion
SourceEventId
OccurredAt
RecordedBy
```

## EVI-005: Evidence freshness is explicit

## EVI-006: Invalidated evidence is not deleted

## EVI-007: AI inference is not confirmed evidence by default

## EVI-008: Notification behavior is not educational evidence

Нельзя считать:

- open;
- click;
- reminder delivery;
- time on screen

доказательством Progress или motivation.

## EVI-009: Missing evidence is not negative evidence

## EVI-010: Private evidence uses restricted references

---

# Authorization Rules

## AU-001: Authentication and authorization are separate

## AU-002: Payload identity does not grant authority

## AU-003: Authorization uses relationship to target

Teacher authorization может зависеть от TeacherAssignment.

## AU-004: Tenant boundary is always checked

## AU-005: Policy Actor has narrow permissions

## AU-006: Scheduler has trigger authority, not educational authority

## AU-007: Integration permissions are explicitly mapped

## AU-008: Guardian authority is scope-limited

## AU-009: Delegation cannot expand original authority

## AU-010: Impersonation preserves original Actor identity

## AU-011: Sensitive commands may require stronger authentication

## AU-012: Unauthorized access must not reveal foreign object existence unnecessarily

## AU-013: Authorization result is auditable

---

# Privacy Rules

## PV-001: Collect and persist minimal data

## PV-002: Private content is stored by reference when possible

## PV-003: Event payload is more restricted than internal command payload

## PV-004: Analytics does not receive raw sensitive domain state by default

## PV-005: Student evaluation data is Sensitive

## PV-006: DomainIntegrityIssue may be Highly Restricted

## PV-007: Notification rendering must respect channel privacy

## PV-008: Logs must not contain full sensitive payload

## PV-009: Audit access is permission-controlled

## PV-010: Archive does not remove privacy requirements

## PV-011: Deletion, anonymization and retention are separate operations

## PV-012: Cross-student data leakage is a critical failure

---

# Notification and Communication Rules

## NC-001: Domain fact and communication are separate

## NC-002: Aggregate does not send notification directly

## NC-003: Domain process creates Notification Intent

## NC-004: Notification Policy decides delivery permission and scheduling

## NC-005: Delivery adapter does not reinterpret domain priority

## NC-006: Delivered does not mean read

## NC-007: Read does not mean understood

## NC-008: Interaction does not mean learning progress

## NC-009: Artificial urgency is prohibited

## NC-010: Educational reminders must not shame or compare students

## NC-011: Message content must not expose private assessment in unsafe channels

## NC-012: Send-time revalidation is required for stale-prone messages

---

# AI Rules

## AI-001: AI is advisory by default

## AI-002: AI is not an authoritative Actor

## AI-003: AI cannot create historical facts

## AI-004: AI cannot fabricate Evidence

## AI-005: AI cannot grant consent

## AI-006: AI cannot independently:

```text
CompleteGoal
AwardAchievement
ExpireHomework
ChangeProgress
FinalizeSongReadiness
ApproveConcertParticipation
RevokeAchievement
```

## AI-007: AI output must preserve provenance

```text
ProposalId
Model
ModelVersion
PromptOrInstructionReference
SourceReferences
GeneratedAt
Confidence
ValidationStatus
```

## AI-008: AI proposals require deterministic validation

## AI-009: Significant AI-assisted decisions require human confirmation or approved policy

## AI-010: AI cannot silently alter policy criteria

## AI-011: AI cannot invent missing domain values

## AI-012: AI explanation must distinguish fact from inference

## AI-013: AI access follows normal authorization and privacy boundaries

## AI-014: AI-generated content must be reviewable where it affects students

## AI-015: AI must not optimize engagement by increasing pressure

## AI-016: Model change requires behavioral regression review for authoritative workflows

---

# Error and Reason Code Rules

## ER-001: Technical and domain failures are distinct

## ER-002: Domain rejection uses stable Reason Codes

## ER-003: Internal codes are not shown directly to students

## ER-004: Human-readable explanation is derived from structured result

## ER-005: One Reason Code has one stable meaning

## ER-006: Free-text error must not replace structured classification

## ER-007: Sensitive details are omitted from unauthorized responses

## ER-008: Retryability is explicit

## ER-009: Conflict is not reported as generic failure

## ER-010: Human Review outcome is distinct from rejection

---

# Audit Rules

## AD-001: Every meaningful mutation is traceable

Trace:

```text
Actor
  |
  v
Command
  |
  v
Decision
  |
  v
Aggregate Version
  |
  v
Domain Event
```

## AD-002: Audit does not depend only on application logs

## AD-003: Rejected sensitive operations are audited

## AD-004: Audit records are immutable or tamper-evident

## AD-005: Audit payload is minimized

## AD-006: Audit records preserve tenant

## AD-007: AI involvement is recorded

## AD-008: Migration provenance is recorded

## AD-009: Historical corrections preserve previous values or references

## AD-010: Audit and event stream are related but not identical

---

# Migration Rules

## MI-001: Migration is an explicit Actor type

## MI-002: Migration must not masquerade as Teacher or Student

## MI-003: Historical import must preserve provenance

## MI-004: Imported timestamps are distinguished from recorded timestamps

## MI-005: Migration cannot fabricate unknown history

Unknown data remains unknown.

## MI-006: Bulk migration requires dry run and validation report

## MI-007: Migration must respect tenant and privacy boundaries

## MI-008: Migration commands or records are idempotent

## MI-009: Historical imported facts may use separate event classification

## MI-010: Direct SQL migration of active domain state requires an approved exception

---

# Testing Rules

## TE-001: Domain tests are written in domain language

Пример:

```text
Given assigned homework
And due date has passed
And valid blocker remains active
When expiration policy is evaluated
Then homework is not expired
And teacher review is requested
```

## TE-002: Aggregate tests cover every lifecycle transition

## TE-003: Aggregate tests cover every invariant rejection

## TE-004: Policy tests cover decision table boundaries

## TE-005: Policy tests are deterministic

## TE-006: Command Handler tests cover:

- authorization;
- idempotency;
- ExpectedVersion;
- transaction;
- event production;
- failure mapping.

## TE-007: Event consumer tests cover duplicate delivery

## TE-008: Process Manager tests cover:

- duplicate event;
- restart;
- timeout;
- retry;
- compensation;
- terminal completion;
- cycle prevention.

## TE-009: Projection tests cover rebuild

## TE-010: Privacy tests are mandatory for sensitive flows

## TE-011: Cross-tenant tests are mandatory

## TE-012: AI boundary tests verify lack of authority

## TE-013: Time boundary tests include:

- exact deadline;
- before deadline;
- after deadline;
- timezone conversion;
- DST where applicable;
- clock skew;
- delayed processing.

## TE-014: Concurrency tests cover competing valid commands

Examples:

```text
SubmitHomework vs ExpireHomework
CompleteGoal vs CancelGoal
AssignSlot vs WithdrawParticipation
DeliverNotification vs CancelNotification
```

## TE-015: Replay tests ensure no external side effects

## TE-016: Tests must not depend on real system clock

## TE-017: Tests must not require network for Domain Layer

## TE-018: Every bug in a domain invariant should produce a regression test

---

# Documentation Rules

## DOC-001: Every domain artifact has an owner

## DOC-002: Every artifact has version and status

## DOC-003: Every Aggregate document defines:

- responsibility;
- identity;
- state;
- invariants;
- lifecycle;
- commands;
- events;
- references;
- privacy;
- tests;
- non-goals;
- open questions.

## DOC-004: Every Policy document defines:

- purpose;
- inputs;
- outcomes;
- rules;
- reason codes;
- decision model;
- human review;
- AI restrictions;
- examples;
- tests.

## DOC-005: Every Command defines:

- target;
- actor;
- payload;
- authorization;
- ExpectedVersion;
- idempotency;
- possible outcomes;
- produced events.

## DOC-006: Every Event defines:

- owner;
- producer;
- meaning;
- payload;
- version;
- privacy;
- consumers;
- ordering;
- idempotency.

## DOC-007: Open Questions do not silently become implementation decisions

## DOC-008: Resolved Open Question is recorded as Decision

## DOC-009: Duplicate definitions are replaced by references to canonical documents

## DOC-010: Human-readable explanation is preferred over implementation-line commentary

Документация должна описывать:

- что делает система;
- почему;
- по каким правилам;
- что происходит при исключениях.

Она не должна сводиться к пересказу строк кода.

---

# Review Rules

## RV-001: New domain artifact requires architecture review

## RV-002: Review checks ownership before implementation detail

## RV-003: Review must ask:

- What fact does this model represent?
- Who owns the invariant?
- What is the consistency boundary?
- What is the lifecycle?
- What is authoritative?
- What can become stale?
- What is the command?
- What event is produced?
- What is the privacy level?
- Can AI influence it?
- How is it tested?

## RV-004: Generic CRUD is a review warning

## RV-005: A service with many unrelated methods is a review warning

## RV-006: Multiple aggregates changed in one handler is a review warning

## RV-007: Boolean fields representing complex lifecycle are a review warning

Examples:

```text
is_done
is_ready
is_active
is_approved
```

без explicit state model.

## RV-008: Status strings without transition rules are prohibited

## RV-009: Events named Updated require special scrutiny

Допускаются только когда доменный смысл действительно состоит в значимом обновлении и changed fields ограничены контрактом.

## RV-010: Direct database mutation of active domain state is a critical review finding

---

# Architectural Smells

Следующие признаки требуют исправления или отдельного решения.

## Anemic Aggregate

```text
Aggregate = fields
Service = all rules
```

## God Service

Один service:

- загружает;
- авторизует;
- вычисляет;
- меняет несколько Aggregate;
- отправляет Notification;
- пишет audit;
- вызывает provider.

## Generic Update

```text
UpdateEntity(fields map)
```

для lifecycle-rich модели.

## Status-Driven Programming

Вся модель строится вокруг произвольной строки status без методов и правил переходов.

## Event as Command

Consumer получает GoalCompleted и напрямую обновляет Award table без команды и проверки.

## Command as Event

CompleteGoal публикуется как будто факт уже произошел.

## Projection as Source of Truth

Mutation выполняется по данным dashboard без загрузки Aggregate.

## Cross-Aggregate Transaction by Default

Несколько Aggregate изменяются в одной transaction без явной причины.

## Hidden Clock

Domain behavior зависит от текущего системного времени, которое невозможно контролировать в тесте.

## AI Authority Leakage

AI response напрямую записывается как Progress или Eligibility.

## Notification Coupling

Homework Aggregate вызывает push provider.

## Evidence Without Provenance

Progress изменяется на основании arbitrary text или score без source reference.

## Boolean Explosion

```text
is_completed
is_cancelled
is_expired
is_archived
is_reopened
```

могут образовать противоречивые комбинации.

---

# Exception Process

Отклонение от обязательного правила допускается только через отдельный Architecture Decision.

Decision должна содержать:

```text
Decision Id
Rule Being Overridden
Context
Reason
Alternatives
Consequences
Risk
Scope
Expiration or Review Date
Approved By
Implementation Constraints
Required Tests
```

Временное исключение не становится новым правилом автоматически.

---

# Definition of Domain-Ready

Доменный компонент считается готовым к реализации, если определены:

- Каноническое имя.
- Responsibility.
- Owner.
- Identity.
- State.
- Lifecycle.
- Invariants.
- Commands.
- Events.
- Policy dependencies.
- Cross-aggregate references.
- Consistency model.
- Authorization.
- Privacy classification.
- Idempotency.
- Concurrency behavior.
- Audit.
- Failure outcomes.
- AI restrictions.
- Tests.
- Non-goals.
- Open Questions.

---

# Definition of Implementation-Ready

Компонент считается готовым к программной реализации, если дополнительно определены:

- Command schema.
- Event schema.
- Value Objects.
- Repository interface.
- ExpectedVersion behavior.
- Transaction boundary.
- Outbox behavior.
- Idempotency storage.
- Projection impacts.
- Migration approach.
- Observability.
- Performance constraints.
- Security tests.
- Acceptance scenarios.

---

# Canonical Compliance Checklist

Перед merge нового доменного изменения проверить:

- [ ] Используется каноническая терминология.
- [ ] Определен владелец инварианта.
- [ ] Выбран корректный Aggregate Root.
- [ ] Нет прямого изменения внутренней Entity.
- [ ] Нет generic status update.
- [ ] Команда выражает намерение.
- [ ] Событие выражает прошедший факт.
- [ ] Policy возвращает Decision.
- [ ] Decision содержит версии входов.
- [ ] Время передается явно.
- [ ] ExpectedVersion проверяется.
- [ ] Idempotency определена.
- [ ] Cross-tenant доступ запрещен.
- [ ] Sensitive payload минимизирован.
- [ ] AI не является authoritative Actor.
- [ ] State и Outbox сохраняются атомарно.
- [ ] Consumer идемпотентен.
- [ ] Projection не используется как source of truth.
- [ ] Cross-aggregate процесс имеет coordinator.
- [ ] Определены failure outcomes.
- [ ] Есть regression и invariant tests.
- [ ] Документация обновлена.

---

# Non-Goals

Этот документ не определяет:

- конкретный язык программирования;
- структуру Go packages;
- framework;
- database engine;
- ORM;
- broker;
- deployment model;
- microservice boundaries;
- API transport;
- mobile architecture;
- UI design;
- cloud provider;
- exact retention periods;
- legal policy;
- CRM domain;
- billing;
- payments;
- payroll;
- marketing automation.

---

# Open Questions

Необходимо определить:

- какие bounded contexts будут выделены физически;
- будет ли modular monolith основным deployment model;
- какие Aggregate войдут в MVP;
- какие Domain Services нужны отдельно от Policies;
- какие Process Managers реализуются первыми;
- будет ли использоваться CQRS физически или только логически;
- какие Aggregate будут event-sourced;
- какие используют state + Outbox;
- какой формат Decision artifact использовать;
- где хранить Policy Decision;
- как version Policy соотносится с deployment version;
- нужен ли централизованный schema registry;
- какой формат identifiers выбрать;
- какие команды допускают mergeable conflict resolution;
- нужен ли reservation pattern для Concert Slot;
- как моделировать group Lesson;
- как моделировать group Homework;
- как моделировать ensemble participation;
- какие Aggregate допускают offline mobile commands;
- как хранить AI proposal provenance;
- какие действия AI может выполнять без human confirmation;
- какие projection freshness guarantees нужны UI;
- как реализовать read-your-write;
- какой срок idempotency retention;
- как обнаруживать event loops;
- нужен ли process manager framework;
- как выполнять migration legacy data;
- как моделировать imported history;
- как проводить anonymization;
- какие audit records являются обязательными долгосрочно;
- какие architecture rules будут проверяться автоматически;
- нужен ли static linter для naming и boundaries;
- как отражать архитектурные правила в repository structure;
- какой шаблон использовать для отдельных Aggregate specifications;
- какой шаблон использовать для Value Object specifications;
- какой шаблон использовать для Process Manager specifications;
- кто имеет право утверждать исключения;
- как часто пересматривать этот документ.

---

# History

| Version | Description |
| --- | --- |
| 1.0.0 | Создана единая конституция доменной модели Belcanto: правила для Aggregate, Entity, Value Object, Policy, Decision, Command, Event, Domain Service, Process Manager, Repository, Projection, времени, evidence, authorization, privacy, AI, тестирования и архитектурного review. |
