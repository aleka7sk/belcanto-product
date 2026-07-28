---
Status: Accepted
Version: 1.0.0
Last Updated: 2026-07-27

Policy Id: ACHIEVEMENT_AWARD_POLICY

Policy Type:
  - Eligibility Policy
  - Reaction Policy
  - Validation Policy

Owners:
  - Product Owner
  - Education Lead
  - Technical Lead

Related Aggregates:
  - Achievement
  - Student
  - Goal
  - Progress
  - Lesson
  - Homework
  - Song
  - Concert
  - BelCoinAccount

Observed Events:
  - GoalCompleted
  - ProgressUpdated
  - LessonCompleted
  - HomeworkReviewed
  - SongReadinessChanged
  - ConcertPerformanceAssessed
  - TeacherProgressReviewCompleted
  - AchievementAwardRequested
  - AchievementRevocationRequested

Produced Commands:
  - AwardAchievement
  - RequestAchievementReview
  - RevokeAchievement
  - RestoreAchievement
  - GrantBelCoins
  - PublishAchievement
  - SuppressAchievementNotification

Related Documents:
  - 000-domain-policy-overview.md
  - progress-update-policy.md
  - goal-completion-policy.md
  - ../progress.md
  - ../student.md
  - ../lesson.md
  - ../homework.md
---

# Achievement Award Policy

> Achievement Award Policy определяет, когда ученик действительно заслуживает достижение, кто может его подтвердить и какие последствия возникают после выдачи.
>
> Achievement фиксирует значимый факт образовательного пути, а не используется как случайная награда за любое действие в приложении.

---

# Purpose

Achievements помогают ученику:

- замечать собственное развитие;
- сохранять важные этапы;
- видеть завершенные цели;
- отмечать первые значимые события;
- поддерживать мотивацию;
- формировать личную историю обучения.

При неправильной модели Achievements быстро превращаются в:

- бессмысленные уведомления;
- искусственную геймификацию;
- соревнование между учениками;
- награду за нажатия;
- способ манипуляции вовлеченностью;
- замену реальной педагогической обратной связи.

Achievement Award Policy должна предотвращать такое поведение.

---

# Core Principle

Achievement выдается только за подтвержденный, понятный и значимый факт.

```text
Confirmed Domain Fact
        |
        v
Achievement Award Policy
        |
        +--> Not Eligible
        |
        +--> More Evidence Required
        |
        +--> Human Review Required
        |
        +--> Eligible
        |
        v
Achievement Awarded
```

Achievement не должен создаваться только потому, что системе необходимо увеличить пользовательскую активность.

---

# Achievement Definition

Achievement представляет собой подтвержденную запись о значимом событии, результате или устойчивом поведении Student.

Концептуальная структура:

```text
Achievement
├── AchievementId
├── DefinitionId
├── DefinitionVersion
├── StudentId
├── Category
├── Title
├── Description
├── Status
├── EvidenceReferences
├── AwardReason
├── AwardedBy
├── AwardedAt
├── EffectiveAt
├── Visibility
├── BelCoinReward
├── PolicyId
├── PolicyVersion
└── AuditMetadata
```

Achievement Definition и выданный Achievement являются разными объектами.

---

# Achievement Definition

Achievement Definition описывает правило и отображение достижения.

```text
AchievementDefinition
├── DefinitionId
├── Version
├── Name
├── Description
├── Category
├── EligibilityCriteria
├── RequiredEvidence
├── Repeatability
├── Visibility
├── BelCoinReward
├── IsActive
├── ValidFrom
└── ValidUntil
```

Изменение Definition не должно бесследно менять уже выданные Achievements.

---

# Achievement Categories

## Milestone Achievement

Фиксирует значимый этап.

Примеры:

- первое завершенное произведение;
- первое выступление;
- первая достигнутая Goal;
- первая опубликованная запись исполнения;
- завершение образовательного этапа.

## Skill Achievement

Связано с подтвержденным развитием конкретного Skill.

Примеры:

- устойчивое сохранение ритма;
- уверенное использование дыхательной опоры;
- стабильное исполнение сложного фрагмента.

Skill Achievement не должно заменять Progress.

## Consistency Achievement

Отмечает устойчивое образовательное поведение.

Примеры:

- регулярная самостоятельная практика;
- последовательное выполнение Homework;
- участие в серии занятий без длительных перерывов.

Consistency Achievement не должно поощрять нездоровое или чрезмерное поведение.

## Performance Achievement

Связано с выступлением.

Примеры:

- первое сценическое выступление;
- успешное исполнение нового Song;
- участие в школьном Concert;
- завершение концертной программы.

## Community Achievement

Связано с конструктивным участием в сообществе.

Примеры:

- участие в общем творческом проекте;
- поддержка школьного события;
- совместное выступление.

Такие достижения требуют особенно осторожной модели, чтобы не создавать социальное давление.

## Personal Achievement

Создается Teacher для индивидуально значимого этапа, который невозможно выразить универсальным правилом.

Пример:

> Ученик впервые уверенно исполнил песню перед группой после длительной работы со сценическим волнением.

Personal Achievement всегда требует человеческого подтверждения.

---

# Achievement Status

Допустимые состояния:

- Draft
- Eligible
- Awarded
- Published
- Revoked
- Archived

## Draft

Achievement подготовлен, но еще не подтвержден.

## Eligible

Система определила, что критерии могут быть выполнены.

Статус не означает, что Achievement уже выдан.

## Awarded

Achievement подтвержден и зафиксирован.

## Published

Achievement доступен Student и другим разрешенным пользователям.

Awarded и Published разделяются, поскольку некоторые достижения могут требовать проверки текста или настройки видимости.

## Revoked

Achievement отозван из-за ошибки, недействительного Evidence или нарушения правил.

История выдачи сохраняется.

## Archived

Achievement остается частью истории, но не показывается в основном активном представлении.

---

# Trigger

Политика применяется при:

```text
GoalCompleted
ProgressUpdated
LessonCompleted
HomeworkReviewed
SongReadinessChanged
ConcertPerformanceAssessed
TeacherProgressReviewCompleted
AchievementAwardRequested
AchievementRevocationRequested
```

Не каждое событие должно приводить к созданию Achievement.

---

# Inputs

Для применения политики могут потребоваться:

- StudentId;
- Achievement Definition;
- Definition Version;
- Trigger Event;
- Evidence References;
- Goal state;
- Progress state;
- Lesson history;
- Homework history;
- Song history;
- Concert history;
- ранее выданные Achievements;
- Teacher confirmation;
- Student visibility settings;
- BelCoin reward configuration;
- Actor;
- Policy Version.

---

# Preconditions

- Student существует.
- Achievement Definition существует.
- Definition активна на момент оценки.
- Actor или Trigger имеет право инициировать проверку.
- Все Evidence относятся к тому же Student.
- Definition Version известна.
- Achievement не был ранее выдан, если он неповторяемый.
- Evidence не отозвано.
- Policy Version доступна.
- Eligibility Criteria могут быть проверены.

---

# Decision Outcomes

## Eligible

Критерии подтверждены.

Achievement может быть выдан.

## Not Eligible

Критерии не выполнены.

## Insufficient Evidence

Для решения недостаточно данных.

## Human Review Required

Требуется подтверждение Teacher или другого уполномоченного сотрудника.

## Already Awarded

Неповторяемое Achievement уже существует.

## Deferred

Решение следует повторить после появления дополнительных данных.

## Rejected

Запрос некорректен или нарушает права, структуру Definition либо доменные ограничения.

## Revocation Required

Выданное Achievement основано на Evidence, которое стало недействительным.

---

# Decision Rules

## AA-001: Achievement requires an active definition

Achievement нельзя выдавать без действующей Achievement Definition.

Reason Code: `ACHIEVEMENT_DEFINITION_REQUIRED`

---

## AA-002: Definition must be versioned

При выдаче сохраняется конкретная версия Definition.

Reason Code: `ACHIEVEMENT_DEFINITION_VERSION_REQUIRED`

---

## AA-003: Achievement requires meaningful criteria

Definition должна иметь конкретные Eligibility Criteria.

Недопустимый критерий:

> Ученик молодец.

Допустимый критерий:

> Student впервые принял участие в подтвержденном Concert Performance.

Reason Code: `ACHIEVEMENT_CRITERIA_REQUIRED`

---

## AA-004: Every award requires Evidence

Achievement не может быть выдан без Evidence References.

Reason Code: `ACHIEVEMENT_EVIDENCE_REQUIRED`

---

## AA-005: Evidence must match the achievement criteria

Evidence должно подтверждать именно тот факт, который описывает Achievement.

Завершенный Lesson не подтверждает автоматически освоение Skill.

Reason Code: `ACHIEVEMENT_EVIDENCE_MISMATCH`

---

## AA-006: Attendance alone is not an educational achievement

Attendance может использоваться для ограниченного Consistency Achievement.

Но количество посещений не должно автоматически считаться педагогическим результатом.

Reason Code: `ATTENDANCE_NOT_EDUCATIONAL_ACHIEVEMENT`

---

## AA-007: Goal completion can satisfy milestone criteria

GoalCompleted может быть достаточным Evidence для Achievement, если Definition прямо связана с завершением Goal.

Пример:

> Первая достигнутая образовательная цель.

---

## AA-008: Progress update does not automatically create an achievement

Не каждое изменение Progress является отдельным достижением.

Achievement должен представлять значимый этап, а не каждое небольшое изменение.

Reason Code: `PROGRESS_CHANGE_NOT_SIGNIFICANT_MILESTONE`

---

## AA-009: Skill achievement requires confirmed progress

Skill Achievement должно основываться на:

- подтвержденном Progress;
- Teacher Review;
- нескольких связанных Evidence;
- либо другом утвержденном критерии.

Одно AI-наблюдение не является достаточным основанием.

---

## AA-010: AI cannot award achievements

AI не может быть AwardedBy.

AI может:

- обнаружить потенциальное соответствие;
- создать Draft;
- предложить Evidence;
- сформировать описание.

Reason Code: `AI_CANNOT_AWARD_ACHIEVEMENT`

---

## AA-011: Personal achievements require human confirmation

Любое Personal Achievement требует подтверждения Teacher.

Reason Code: `PERSONAL_ACHIEVEMENT_REQUIRES_TEACHER`

---

## AA-012: Student cannot award an achievement to themselves

Student может:

- предложить Achievement;
- отправить Self Assessment;
- указать на пропущенный факт;
- запросить Review.

Student не может самостоятельно подтвердить выдачу.

Reason Code: `SELF_AWARD_NOT_ALLOWED`

---

## AA-013: Non-repeatable achievement is awarded once

Если Definition имеет:

`Repeatability: Once`

повторная выдача запрещена.

Результат: Already Awarded

Reason Code: `ACHIEVEMENT_ALREADY_AWARDED`

---

## AA-014: Repeatable achievements require occurrence identity

Для повторяемых Achievements каждый экземпляр должен относиться к отдельному факту.

Пример:

> Achievement за участие в Concert может повторяться, если каждый award связан с отдельным ConcertId.

Reason Code: `ACHIEVEMENT_OCCURRENCE_REQUIRED`

---

## AA-015: Duplicate evidence does not create multiple achievements

Повторная обработка одного события не должна создавать копии.

Reason Code: `ACHIEVEMENT_TRIGGER_ALREADY_PROCESSED`

---

## AA-016: Achievement must not reward arbitrary app usage

Запрещены достижения только за:

- открытие приложения;
- просмотр экрана;
- нажатие кнопок;
- ежедневный вход без образовательного смысла;
- чтение уведомления;
- покупку;
- оплату;
- увеличение времени в приложении.

Это не запрещает onboarding indicators, но они не должны называться образовательными Achievements.

Reason Code: `NON_EDUCATIONAL_ACTIVITY_NOT_ELIGIBLE`

---

## AA-017: Achievement must not encourage unhealthy behavior

Запрещено награждать за:

- чрезмерную длительность практики;
- занятия без отдыха;
- выполнение заданий ночью;
- посещение при плохом самочувствии;
- отказ от перерывов;
- нездоровые серии активности.

Reason Code: `UNHEALTHY_BEHAVIOR_REWARD_NOT_ALLOWED`

---

## AA-018: Achievement must not compare students

Критерии не должны зависеть от места Student среди других учеников.

Запрещено:

- лучший ученик недели;
- топ-1 по занятиям;
- больше всех Homework;
- быстрее всех достиг Goal.

Reason Code: `COMPETITIVE_RANKING_NOT_ALLOWED`

---

## AA-019: Achievement wording must be respectful

Название и описание не должны:

- унижать;
- высмеивать;
- содержать медицинские выводы;
- фиксировать недостатки;
- раскрывать чувствительную информацию;
- создавать давление.

Reason Code: `ACHIEVEMENT_TEXT_NOT_APPROPRIATE`

---

## AA-020: Visibility must be explicit

Каждое Achievement должно иметь Visibility.

Допустимые значения:

- Student Only;
- Student and Teachers;
- School Community;
- Private Staff;
- Public with Consent.

По умолчанию используется минимальная видимость.

Reason Code: `ACHIEVEMENT_VISIBILITY_REQUIRED`

---

## AA-021: Public visibility requires consent

Achievement нельзя публиковать публично без явного согласия Student или законного представителя, если применимо.

Reason Code: `PUBLIC_ACHIEVEMENT_CONSENT_REQUIRED`

---

## AA-022: BelCoin reward must not determine educational truth

Наличие BelCoin Reward не должно влиять на решение о том, заслужено ли Achievement.

Сначала определяется Achievement Eligibility.

Только после подтверждения отдельно применяется BelCoin consequence.

---

## AA-023: BelCoin reward must be defined in advance

Размер награды должен быть указан в Achievement Definition или утвержденной Reward Policy.

Запрещено произвольно назначать BelCoins в момент выдачи.

Reason Code: `BELCOIN_REWARD_RULE_REQUIRED`

---

## AA-024: Award and BelCoin grant must be idempotent

Повторная обработка AchievementAwarded не должна повторно начислять BelCoins.

Для начисления используется уникальная ссылка на AchievementId.

Reason Code: `BELCOIN_REWARD_ALREADY_GRANTED`

---

## AA-025: Achievement can exist without BelCoins

Не каждое Achievement обязано иметь денежную или виртуальную награду.

Признание значимого этапа является самостоятельной ценностью.

---

## AA-026: BelCoins cannot be the sole purpose of achievement

Definition не должна существовать только для выдачи валюты без образовательного или общественного смысла.

---

## AA-027: Revoked evidence triggers review

Если Evidence было:

- Withdrawn;
- Superseded;
- признано ошибочным;
- связано с другим Student;
- создано без полномочий;

выданное Achievement должно пройти Review.

Reason Code: `ACHIEVEMENT_EVIDENCE_INVALIDATED`

---

## AA-028: Revocation does not delete history

Отозванное Achievement остается в Audit.

Сохраняются:

- первоначальная выдача;
- исходные Evidence;
- причина отзыва;
- Actor;
- дата;
- связанные BelCoin consequences.

---

## AA-029: Revocation requires explanation

Achievement нельзя отозвать без Reason Code и понятного объяснения.

Reason Code: `ACHIEVEMENT_REVOCATION_REASON_REQUIRED`

---

## AA-030: BelCoin consequences of revocation require separate policy

Отзыв Achievement не должен автоматически списывать уже начисленные BelCoins без отдельного правила.

Необходимо учитывать:

- были ли BelCoins потрачены;
- произошла ли техническая ошибка;
- виноват ли Student;
- создаст ли списание отрицательный баланс;
- допустима ли компенсационная операция.

---

## AA-031: Historical definition changes do not rewrite awards

Изменение названия, критериев или награды Definition не должно автоматически менять исторический Achievement.

При необходимости создается отдельная migration policy.

---

## AA-032: Award must preserve authorship and source

Сохраняются:

- кто инициировал;
- кто подтвердил;
- какое событие стало причиной;
- какие Evidence использованы;
- участвовал ли AI.

---

## AA-033: Achievement can be manually proposed

Teacher может предложить Achievement, которого нет в автоматическом потоке.

Но предложение должно:

- использовать существующую Definition;
- либо создать Draft Definition для отдельного утверждения;
- содержать Evidence;
- соблюдать права и видимость.

---

## AA-034: Administrator cannot create pedagogical achievements by default

Administrator может:

- исправить техническую ошибку;
- запустить Review;
- управлять Definition;
- восстановить ошибочно отозванное Achievement.

Administrator не должен самостоятельно подтверждать педагогический Skill Achievement без соответствующего полномочия.

---

## AA-035: Owner analytics does not affect eligibility

Популярность Achievement, количество выдач и маркетинговые показатели не должны менять критерии выдачи отдельному Student.

---

## AA-036: Award operation is versioned

Выдача должна учитывать:

- Student version;
- Achievement Definition version;
- Evidence versions;
- Policy version.

---

## AA-037: Concurrent awards must resolve to one valid result

Если несколько обработчиков одновременно пытаются выдать неповторяемое Achievement, должен быть создан только один экземпляр.

---

## AA-038: Achievement must remain explainable

Student должен понимать:

- за что выдано Achievement;
- когда это произошло;
- какой факт оно отмечает;
- кто подтвердил;
- является ли оно публичным;
- начислены ли BelCoins.

---

# Eligibility Criteria Model

Рекомендуемая структура:

```text
AchievementEligibilityCriteria
├── CriterionId
├── CriterionType
├── Required
├── TargetReference
├── Operator
├── ExpectedValue
├── RequiredEvidenceTypes
├── MinimumEvidenceCount
├── RequiresHumanConfirmation
└── ExplanationTemplate
```

---

# Criterion Types

## Event Occurred

Пример:

`ConcertPerformanceCompleted`

## Goal Completed

Пример:

`GoalCategory = FIRST_SONG_COMPLETION`

## Progress State Reached

Пример:

`SkillProgress.Confidence = High`

## Repeated Confirmed Behavior

Пример:

`ReviewedHomeworkCount >= 5`

Количество само по себе недостаточно, если Definition требует качества выполнения.

## Teacher Confirmed Milestone

Пример:

`TeacherReviewOutcome = MILESTONE_REACHED`

## Composite Criterion

Сочетает несколько условий.

Пример:

```text
GoalCompleted
AND
TeacherConfirmed
AND
SongReadiness = PerformanceReady
```

---

# Repeatability

Допустимые модели:

## Once

Выдается один раз за весь путь Student.

Пример:

- первое выступление;
- первая завершенная Goal.

## Once Per Subject

Выдается один раз для каждого Song, Skill или Goal.

Пример:

Song полностью подготовлена.

## Once Per Event

Выдается для каждого отдельного Concert.

## Repeatable With Cooldown

Может повторяться, но не чаще заданного периода.

Используется осторожно, чтобы не превратить Achievement в механическую серию.

## Manual Only

Может выдаваться повторно только через подтвержденный Human Review.

---

# Achievement Award Flow

```text
Domain Event received
        |
        v
Find applicable definitions
        |
        v
Validate definition and version
        |
        v
Load required evidence
        |
        v
Check existing achievements
        |
        v
Evaluate criteria
        |
        +--> Not Eligible
        |
        +--> Insufficient Evidence
        |
        +--> Human Review Required
        |
        +--> Already Awarded
        |
        v
Eligible
        |
        v
Execute AwardAchievement
        |
        v
AchievementAwarded
        |
        +--> Grant BelCoins
        |
        +--> Notification Policy
        |
        +--> Publish Achievement
        |
        +--> Update read models
```

---

# Commands Produced

## AwardAchievement

Создает подтвержденный Achievement.

## RequestAchievementReview

Создает задачу Teacher или другому уполномоченному Actor.

## RevokeAchievement

Переводит Achievement в Revoked.

## RestoreAchievement

Восстанавливает ошибочно отозванное Achievement через отдельный audited flow.

## GrantBelCoins

Создается только после успешного AchievementAwarded, если Definition содержит награду.

## PublishAchievement

Делает Achievement доступным согласно Visibility.

## SuppressAchievementNotification

Используется, если Achievement должно быть сохранено без отдельного уведомления.

---

# Domain Events

После операций могут создаваться:

```text
AchievementEligibilityDetected
AchievementReviewRequested
AchievementAwarded
AchievementPublished
AchievementNotificationSuppressed
AchievementRevocationRequested
AchievementRevoked
AchievementRestored
AchievementBelCoinRewardRequested
AchievementBelCoinRewardGranted
```

## AchievementAwarded Event

Событие должно содержать:

- AchievementId;
- DefinitionId;
- DefinitionVersion;
- StudentId;
- Category;
- Evidence References;
- AwardedBy;
- AwardedAt;
- Visibility;
- BelCoinReward reference;
- PolicyId;
- PolicyVersion;
- CorrelationId;
- CausationId.

Событие не должно раскрывать закрытые тексты Evidence.

---

# Human Review

Human Review обязателен для:

- Personal Achievement;
- спорного Skill Achievement;
- противоречивого Evidence;
- Achievement с чувствительной формулировкой;
- ручной выдачи;
- публичного Achievement без ранее заданного consent;
- значимой награды BelCoins;
- восстановления отозванного Achievement;
- Achievement, предложенного AI.

## Human Review Result

Reviewer может:

- подтвердить выдачу;
- отклонить;
- изменить описание;
- изменить Visibility;
- исключить неподходящее Evidence;
- запросить дополнительные данные;
- выбрать другую Definition;
- отменить BelCoin Reward;
- отложить публикацию.

Изменение критериев самой Definition не должно происходить внутри Review конкретного Student.

---

# BelCoins Integration

Achievement Award Policy определяет только право инициировать награду.

Фактическое движение BelCoins должно происходить через отдельную транзакционную модель.

```text
AchievementAwarded
        |
        v
Reward Policy
        |
        v
GrantBelCoins
        |
        v
BelCoinTransactionCreated
```

BelCoin transaction должна содержать:

- TransactionId;
- StudentId;
- Amount;
- Reason;
- AchievementId;
- DefinitionVersion;
- IdempotencyKey;
- CreatedAt.

---

# Notification Effects

Achievement Award Policy не отправляет уведомления напрямую.

Notification Policy решает:

- уведомлять ли Student;
- уведомлять ли Teacher;
- объединять ли уведомление с Goal completion;
- публиковать ли Achievement в Community;
- отложить ли уведомление;
- скрыть ли BelCoin amount;
- учитывать ли quiet hours.

---

# Student Presentation

Student должен видеть:

- название Achievement;
- понятное описание;
- дату;
- причину;
- связанные доступные Evidence;
- имя Teacher, если допустимо;
- BelCoin Reward;
- Visibility;
- возможность скрыть Achievement из Community.

Следует избегать:

- технических Reason Codes;
- скрытых педагогических заметок;
- сравнения с другими;
- искусственной редкости;
- давления сохранить серию;
- формулировок о потере достижения из-за перерыва.

---

# Teacher Presentation

Teacher должен видеть:

- Definition;
- критерии;
- Evidence;
- существующие Awards;
- необходимость подтверждения;
- историю Review;
- AI proposal metadata;
- BelCoin consequence;
- Visibility;
- возможные конфликты.

---

# Owner Analytics

Owner может видеть:

- количество выдач по Definition;
- процент ручных Review;
- частоту отзыва;
- неиспользуемые Definitions;
- концентрацию BelCoin Rewards;
- повторные ошибки;
- долю автоматических и ручных Awards.

Owner Analytics не должна использоваться для рейтинга Student или Teacher без отдельной утвержденной модели.

---

# AI Assistance

AI может:

- искать потенциальное соответствие Definition;
- предлагать Achievement;
- собирать Evidence References;
- создавать Draft Description;
- обнаруживать дубликаты;
- выявлять неподходящие формулировки;
- находить конфликтующее Evidence.

AI не может:

- выдавать Achievement;
- начислять BelCoins;
- изменять Visibility;
- публиковать Achievement;
- отзывать Achievement;
- утверждать Personal Achievement;
- создавать недостающие факты.

AI proposal должен сохранять:

- модель;
- версию;
- входные Evidence;
- Confidence;
- время;
- предложенную Definition;
- статус человеческого подтверждения.

---

# Privacy

Achievement может косвенно раскрывать чувствительную информацию.

Например:

- работу со сценическим страхом;
- длительную паузу;
- конкретную образовательную сложность;
- участие в закрытом мероприятии.

Необходимо:

- использовать минимальную видимость;
- проверять описание;
- не публиковать скрытое Evidence;
- хранить consent;
- позволять Student скрывать Community Achievement;
- отделять публичное описание от внутреннего основания.

---

# Audit Requirements

Для каждой оценки политики сохраняются:

- PolicyId;
- PolicyVersion;
- Trigger Event;
- StudentId;
- DefinitionId;
- DefinitionVersion;
- Evidence References;
- ActorId;
- Decision;
- Reason Codes;
- Existing Achievement check;
- Human Review requirement;
- BelCoin Reward;
- Visibility;
- EvaluatedAt;
- CorrelationId;
- CausationId.

Для выдачи дополнительно:

- AchievementId;
- AwardedBy;
- AwardedAt;
- final description;
- consent reference;
- idempotency key;
- BelCoin command reference.

Для отзыва:

- RevokedBy;
- RevokedAt;
- reason;
- invalidated Evidence;
- financial consequence decision;
- replacement Achievement reference.

---

# Failure Modes

## Definition not found

- Decision: Rejected
- Reason Code: ACHIEVEMENT_DEFINITION_NOT_FOUND

## Definition inactive

- Decision: Not Eligible
- Reason Code: ACHIEVEMENT_DEFINITION_INACTIVE

## Evidence missing

- Decision: Insufficient Evidence
- Reason Code: ACHIEVEMENT_EVIDENCE_REQUIRED

## Evidence belongs to another Student

- Decision: Rejected
- Reason Code: ACHIEVEMENT_EVIDENCE_STUDENT_MISMATCH

Security Audit обязателен.

## Achievement already exists

- Decision: Already Awarded
- Reason Code: ACHIEVEMENT_ALREADY_AWARDED

## AI proposal without Teacher confirmation

- Decision: Human Review Required
- Reason Code: AI_ACHIEVEMENT_REQUIRES_CONFIRMATION

## Public visibility without consent

- Decision: Human Review Required
- Reason Code: PUBLIC_ACHIEVEMENT_CONSENT_REQUIRED

## Duplicate event

- Decision: Already Awarded or No Action

Новые последствия не создаются.

## Evidence withdrawn after award

- Decision: Human Review Required
- Reason Code: ACHIEVEMENT_EVIDENCE_INVALIDATED

## BelCoin service unavailable

Achievement может быть выдано, если сама выдача не зависит от доступности Reward infrastructure.

Команда начисления должна быть надежно повторена позже.

Пользователь не должен видеть BelCoins как начисленные до подтвержденной транзакции.

---

# Explainability Examples

## First Performance

> Вы впервые выступили на концерте Belcanto. Участие подтверждено преподавателем и связано с выступлением 24 июля.

## Goal Completed

> Вы достигли первой образовательной цели: уверенно исполнили выбранную песню полностью. Результат подтвержден преподавателем.

## Practice Consistency

> В течение четырех недель вы регулярно выполняли и отправляли задания, а преподаватель подтвердил их выполнение.

## Skill Milestone

> В нескольких занятиях и на репетиции вы стабильно сохраняли ритм в сложном фрагменте. Преподаватель подтвердил новый этап навыка.

---

# Examples

## Example 1: First completed goal

Дано:

- Student завершил первую Goal;
- Goal подтверждена Teacher;
- Definition активна;
- Achievement ранее не выдавалось.

Результат:

- Decision: Eligible
- Command: AwardAchievement
- Reason Code: FIRST_GOAL_COMPLETED

## Example 2: Duplicate first goal award

Дано:

- Achievement уже было выдано;
- повторно обработано событие GoalCompleted.

Результат:

- Decision: Already Awarded
- Reason Code: ACHIEVEMENT_ALREADY_AWARDED

## Example 3: Lesson attendance only

Дано:

- Student посетил десять Lesson;
- педагогического критерия нет;
- Definition описывает освоение Skill.

Результат:

- Decision: Not Eligible
- Reason Code: ATTENDANCE_NOT_EDUCATIONAL_ACHIEVEMENT

## Example 4: First concert

Дано:

- Concert Performance завершено;
- Student участвовал;
- Teacher подтвердил;
- Achievement повторяется только один раз за первое выступление.

Результат:

- Decision: Eligible
- Command: AwardAchievement
- Reason Code: FIRST_CONCERT_PERFORMANCE_CONFIRMED

## Example 5: AI suggests skill achievement

Дано:

- AI обнаружил устойчивые положительные наблюдения;
- Teacher еще не подтвердил;
- Definition требует профессионального подтверждения.

Результат:

- Decision: Human Review Required
- Command: RequestAchievementReview
- Reason Code: AI_ACHIEVEMENT_REQUIRES_CONFIRMATION

## Example 6: Repeatable concert achievement

Дано:

- Student участвует в новом Concert;
- Definition: Once Per Event;
- Achievement за этот Concert еще нет.

Результат:

- Decision: Eligible
- Occurrence Reference: ConcertId
- Reason Code: CONCERT_PARTICIPATION_CONFIRMED

## Example 7: Achievement with BelCoins

Дано:

- Achievement успешно выдано;
- Definition содержит 100 BelCoins;
- Reward ранее не начислялась.

Результат:

- Command: GrantBelCoins
- Idempotency Key: AchievementId

## Example 8: Withdrawn evidence

Дано:

- Achievement было выдано по ошибочному Assessment;
- Assessment отозван.

Результат:

- Decision: Human Review Required
- Reason Code: ACHIEVEMENT_EVIDENCE_INVALIDATED

Achievement не удаляется автоматически.

---

# Test Requirements

## Definition Tests

- active Definition can be evaluated;
- inactive Definition is rejected;
- missing Definition is rejected;
- Definition Version is stored;
- historical Award preserves old Definition Version;
- Definition without criteria is invalid.

## Eligibility Tests

- valid Goal Completion creates eligibility;
- Lesson attendance alone does not satisfy Skill Achievement;
- confirmed Concert satisfies applicable criteria;
- insufficient Evidence does not award;
- mismatched Evidence is rejected;
- withdrawn Evidence triggers Review.

## Repeatability Tests

- Once achievement is awarded once;
- Once Per Event uses EventId;
- duplicate trigger is idempotent;
- repeatable achievement requires new occurrence;
- concurrent requests create one valid Award.

## Permission Tests

- authorized Teacher can confirm;
- Student cannot self-award;
- AI cannot award;
- Administrator cannot confirm pedagogical Achievement by default;
- Owner cannot bypass criteria;
- blocked Actor is rejected.

## BelCoin Tests

- eligible Achievement creates one reward command;
- duplicate event does not duplicate reward;
- Achievement without reward creates no transaction;
- unavailable reward infrastructure does not duplicate later grant;
- revoked Achievement does not automatically debit without policy.

## Visibility Tests

- default visibility is minimal;
- public publication requires consent;
- private Evidence is not exposed;
- Student can hide Community Achievement where allowed;
- internal Personal Achievement is not published automatically.

## Revocation Tests

- revoked Achievement remains in history;
- reason is required;
- unauthorized revocation is rejected;
- invalidated Evidence requests Review;
- restoration creates Audit event;
- revocation does not silently delete BelCoins.

## AI Tests

- AI proposal remains Draft;
- AI sources must exist;
- hallucinated Evidence is rejected;
- Teacher confirmation is recorded;
- AI cannot publish or grant rewards.

## Explainability Tests

- Award contains readable reason;
- Award references valid Evidence;
- Student explanation does not expose private notes;
- Already Awarded result is explainable;
- Not Eligible result has Reason Code.

---

# Non-Goals

Achievement Award Policy не определяет:

- Progress Calculation;
- Goal Completion;
- Song Readiness;
- Concert Eligibility;
- стоимость BelCoins;
- каталог наград;
- правила покупки;
- финансовые операции;
- расписание;
- CRM-статусы;
- рейтинги;
- оплату Teacher;
- маркетинговые акции;
- push delivery.

---

# Open Questions

Необходимо определить:

- стартовый каталог Achievement Definitions;
- какие достижения входят в MVP;
- какие Achievements могут выдаваться автоматически;
- какие всегда требуют Teacher;
- какие категории получают BelCoins;
- максимальный размер BelCoin Reward;
- допускаются ли скрытые Achievements;
- может ли Student отказаться от Achievement;
- может ли Student скрывать его из профиля;
- разрешены ли Community Achievements;
- требуется ли согласие Teacher на автоматический Award;
- как выдавать Achievements группе;
- нужны ли временные Achievements;
- как мигрировать Definitions;
- как обрабатывать объединение дублирующих Achievements;
- кто может создавать Personal Achievement;
- могут ли родители видеть Achievements несовершеннолетнего;
- когда Achievement становится Published;
- нужно ли откладывать публикацию до Teacher feedback;
- можно ли отзывать Achievement без Student notification;
- как компенсировать ошибочно начисленные BelCoins;
- как предотвратить инфляцию BelCoins;
- как измерять полезность Achievements без оптимизации на вовлеченность;
- должны ли Achievement titles локализоваться;
- как сохранять текст исторической локализации;
- нужна ли отдельная Reward Policy.

---

# History

| Version | Description |
| --- | --- |
| 1.0.0 | Определены правила проверки, выдачи, публикации и отзыва Achievement. |
