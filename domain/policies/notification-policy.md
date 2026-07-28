---
Status: Approved
Version: 1.0.0
Last Updated: 2026-07-27

Policy Id: NOTIFICATION_POLICY

Policy Type:
  - Communication Policy
  - Delivery Policy
  - Scheduling Policy
  - Privacy Policy
  - Reaction Policy
  - Escalation Policy
  - Reliability Policy

Owners:
  - Product Owner
  - Education Lead
  - Technical Lead
  - Privacy Owner

Related Aggregates:
  - Notification
  - NotificationIntent
  - NotificationDelivery
  - NotificationPreference
  - NotificationTemplate
  - NotificationBundle
  - Student
  - Teacher
  - Administrator
  - Guardian
  - Lesson
  - Homework
  - Goal
  - Achievement
  - Song
  - Concert
  - LearningPause

Observed Events:
  - NotificationRequested
  - NotificationPreferenceChanged
  - NotificationChannelEnabled
  - NotificationChannelDisabled
  - NotificationConsentGranted
  - NotificationConsentWithdrawn
  - StudentTimezoneChanged
  - QuietHoursChanged
  - LearningPauseStarted
  - LearningPauseEnded
  - HomeworkReminderSendingRequested
  - HomeworkStatusChanged
  - LessonScheduled
  - LessonRescheduled
  - LessonCancelled
  - GoalCompleted
  - GoalReviewRequested
  - AchievementAwarded
  - AchievementRevoked
  - SongReadinessChanged
  - ConcertParticipationChanged
  - ConcertProgramChanged
  - PerformanceSlotChanged
  - NotificationDeliverySucceeded
  - NotificationDeliveryFailed
  - NotificationOpened
  - NotificationActionCompleted
  - NotificationRetryRequested
  - NotificationRecalculationRequested

Produced Commands:
  - CreateNotificationIntent
  - ApproveNotificationIntent
  - RejectNotificationIntent
  - ScheduleNotification
  - RescheduleNotification
  - CancelNotification
  - SuppressNotification
  - BundleNotifications
  - RenderNotification
  - SendNotification
  - RetryNotificationDelivery
  - StopNotificationRetry
  - SwitchNotificationChannel
  - MarkNotificationDelivered
  - MarkNotificationDeliveryFailed
  - MarkNotificationOpened
  - ExpireNotification
  - ArchiveNotification
  - NotifyAdministratorAboutDeliveryIncident
  - RequestNotificationReview

Related Documents:
  - 000-domain-policy-overview.md
  - lesson-completion-policy.md
  - progress-update-policy.md
  - goal-completion-policy.md
  - achievement-award-policy.md
  - song-readiness-policy.md
  - concert-eligibility-policy.md
  - homework-reminder-policy.md
  - homework-expiration-policy.md
  - periodic-review-policy.md
  - ../student.md
  - ../lesson.md
  - ../homework.md
  - ../progress.md
---

# Notification Policy

> Notification Policy определяет, может ли доменное сообщение быть доставлено получателю, когда его следует доставить, через какой канал, в какой форме и с какими ограничениями.
>
> Политика не создает образовательный смысл сообщения. Она принимает разрешенный Notification Intent от другой доменной политики и отвечает за безопасную, своевременную, ненавязчивую и аудируемую доставку.

---

# Purpose

Belcanto Product создает коммуникацию в разных контекстах:

- Lesson назначен или перенесен;
- Homework создано;
- Homework Reminder должен быть отправлен;
- Homework стало Overdue или Expired;
- Teacher запросил уточнение;
- Goal завершена;
- требуется Goal Review;
- Achievement присужден;
- Song Readiness изменилась;
- Concert Participation подтверждена;
- Performance Slot изменен;
- программа Concert опубликована;
- требуется согласие;
- возник технический сбой;
- необходимо действие Administrator.

Без единой Notification Policy каждая часть продукта может самостоятельно:

- отправлять слишком много сообщений;
- игнорировать Quiet Hours;
- дублировать уведомления;
- обходить пользовательские настройки;
- раскрывать private данные на экране блокировки;
- отправлять устаревшие сообщения;
- путать образовательное событие с маркетинговой коммуникацией;
- использовать неподходящий канал;
- повторно отправлять одно и то же;
- создавать ложную срочность;
- подменять Delivery фактом прочтения;
- превращать уведомления в инструмент давления.

Политика должна объединить все коммуникационные ограничения в одном месте.

---

# Core Principle

Notification разрешено отправлять только тогда, когда одновременно выполняются три условия:

```text
Valid Domain Intent
        +
Permitted Recipient Communication
        +
Current Delivery Context
        =
Deliverable Notification
```

Даже корректный Notification Intent не гарантирует немедленную отправку.

Перед доставкой необходимо повторно проверить:

- актуальность события;
- получателя;
- consent;
- preferences;
- канал;
- Quiet Hours;
- timezone;
- frequency limits;
- privacy;
- наличие более новой версии;
- отсутствие отмены;
- возможность bundling;
- срок полезности сообщения.

---

## Domain Intent and Delivery Are Separate

Доменная политика отвечает:

> Какое сообщение требуется в связи с доменным событием?

Notification Policy отвечает:

> Разрешено ли его доставить, когда, куда и в каком виде?

Пример:

```text
Homework Reminder Policy
        |
        v
HomeworkReminderSendingRequested
        |
        v
Notification Policy
        |
        +--> Send now
        +--> Schedule
        +--> Bundle
        +--> Suppress
        +--> Change channel
        +--> Expire
```

Notification Policy не должна менять исходный педагогический смысл.

---

# Notification Intent

Рекомендуемая структура:

```text
NotificationIntent
├── NotificationIntentId
├── IntentType
├── SourceDomain
├── SourceEntityType
├── SourceEntityId
├── SourceEntityVersion
├── RecipientType
├── RecipientId
├── AudienceScope
├── Purpose
├── Priority
├── Urgency
├── RequiredAction
├── RequestedChannels
├── DeliveryWindow
├── ExpiresAt
├── TemplateReference
├── RenderingParameters
├── PrivacyLevel
├── DeduplicationKey
├── CorrelationId
├── CausationId
├── PolicyId
├── PolicyVersion
└── Version
```

---

# Notification Delivery

```text
NotificationDelivery
├── NotificationDeliveryId
├── NotificationIntentId
├── RecipientId
├── Channel
├── DestinationReference
├── Status
├── ScheduledFor
├── RenderedContentReference
├── TemplateVersion
├── IdempotencyKey
├── AttemptCount
├── LastAttemptAt
├── DeliveredAt
├── OpenedAt
├── ActionCompletedAt
├── FailedAt
├── FailureCode
├── SuppressedAt
├── SuppressionReason
├── ExpiresAt
├── ProviderReference
└── Version
```

---

# Notification Status

Допустимые состояния:

- Draft
- Pending Review
- Approved
- Scheduled
- Queued
- Rendering
- Sending
- Delivered
- Opened
- Action Completed
- Delivery Failed
- Retry Scheduled
- Suppressed
- Cancelled
- Expired
- Archived

## Draft

Intent создан, но еще не прошел проверку.

## Pending Review

Требуется Human Review или дополнительная проверка.

## Approved

Intent разрешен к дальнейшей обработке.

## Scheduled

Уведомление запланировано на конкретное время.

## Queued

Delivery передано в инфраструктурную очередь.

## Rendering

Создается channel-specific представление.

## Sending

Провайдер принял запрос на доставку или отправка выполняется.

## Delivered

Получено технически доступное подтверждение доставки.

Это не означает, что сообщение прочитано.

## Opened

Получен сигнал открытия, если канал и настройки это поддерживают.

Открытие не означает понимание или выполнение действия.

## Action Completed

Получатель выполнил целевое действие.

Например:

- подтвердил участие;
- открыл Homework и отправил Submission;
- подтвердил новый Lesson time;
- дал consent.

## Delivery Failed

Отправка не удалась.

## Retry Scheduled

Запланирована повторная попытка.

## Suppressed

Уведомление сознательно не доставлено по правилу Policy.

## Cancelled

Intent или Delivery отменены до отправки.

## Expired

Сообщение потеряло актуальность.

## Archived

Delivery сохранено только в истории.

---

# Recipient Types

Политика поддерживает:

- Student;
- Teacher;
- Administrator;
- Owner;
- Concert Coordinator;
- Guardian;
- External Collaborator;
- System Operator.

Каждый Recipient Type должен иметь отдельные:

- permission rules;
- privacy boundaries;
- allowed message categories;
- channel rules;
- default preferences.

---

# Notification Categories

## Educational

Сообщения, связанные с обучением:

- Lesson;
- Homework;
- Assessment;
- Goal;
- Progress;
- Song Readiness.

## Operational

Сообщения, необходимые для организации процесса:

- Lesson reschedule;
- Concert schedule;
- room change;
- technical requirement;
- approval request;
- consent request.

## Transactional

Подтверждают уже совершенное действие:

- Homework submitted;
- consent recorded;
- Concert Participation withdrawn;
- notification preference changed.

## Safety and Security

Сообщения о:

- подозрительном входе;
- изменении контакта;
- изменении важных consent;
- возможной ошибке данных;
- security incident.

## Administrative

Сообщения для Staff:

- unresolved blocker;
- delivery incident;
- pending review;
- unassigned responsibility.

## Informational

Не требуют обязательного действия:

- Achievement Awarded;
- Goal Completed;
- новый материал;
- итог Lesson.

## Marketing

Маркетинговая коммуникация должна существовать как отдельная категория с отдельным consent.

Educational и Transactional сообщения нельзя маскировать под Marketing и наоборот.

---

# Priority

Допустимые значения:

- Low
- Normal
- High
- Critical

Critical применяется только к действительно критическим сообщениям.

Homework Reminder, Goal Completion и Achievement обычно не являются Critical.

---

# Urgency

Priority и Urgency различаются.

## Priority

Насколько важно обработать сообщение системой.

## Urgency

Насколько быстро получателю нужно совершить действие.

Пример:

- изменение Concert Slot за два часа до мероприятия: High Priority, High Urgency;
- Achievement Awarded: Normal Priority, Low Urgency;
- security warning: Critical Priority, High Urgency;
- Homework Reminder за три дня: Normal Priority, Low Urgency.

---

# Delivery Channels

Базовые каналы:

- In-App;
- Push;
- Email;
- SMS;
- Messenger Integration;
- Staff Dashboard;
- System Inbox.

MVP может поддерживать только часть каналов.

## In-App

Основной безопасный канал для полной информации.

Преимущества:

- управляемая privacy;
- глубокие ссылки;
- доступ к актуальному состоянию;
- отсутствие раскрытия полного текста на lock screen.

## Push

Используется для короткого сигнала.

Push не должен содержать чувствительные детали, если preview может быть виден третьим лицам.

## Email

Подходит для:

- длинных сообщений;
- summary;
- отчетов;
- подтверждений;
- материалов;
- административной коммуникации.

Email destination должен быть подтвержден.

## SMS

Используется только при отдельном разрешении и для ограниченных сценариев.

SMS не подходит для передачи чувствительных педагогических данных.

## Messenger Integration

Требует отдельной интеграционной и privacy-модели.

Нельзя считать наличие номера телефона согласием на сообщения в Messenger.

## Staff Dashboard

Может использоваться вместо внешнего уведомления для Teacher и Administrator.

## System Inbox

Хранит Notification независимо от внешней доставки.

---

# Notification Preference

```text
NotificationPreference
├── PreferenceId
├── RecipientId
├── Category
├── Enabled
├── AllowedChannels
├── PreferredChannelOrder
├── QuietHours
├── AllowedWeekdays
├── FrequencyLimit
├── BundlingPreference
├── Locale
├── Timezone
├── PreviewPrivacyLevel
├── ConsentReference
├── EffectiveFrom
└── Version
```

---

# Notification Consent

Consent и Preference различаются.

## Consent

Имеет ли продукт право использовать конкретный канал или категорию.

## Preference

Как пользователь хочет получать разрешенную коммуникацию.

Отключенная Preference не должна автоматически трактоваться как отзыв юридически значимого Consent, если модель разделяет эти сущности.

---

# Decision Outcomes

## Deliver Now

Сообщение можно доставить немедленно.

## Schedule Delivery

Сообщение следует отправить позже.

## Reschedule Delivery

Ранее выбранное время больше не подходит.

## Bundle

Несколько Intent объединяются.

## Change Channel

Первичный канал недоступен или запрещен, но разрешен fallback.

## In-App Only

Внешнее сообщение подавляется, но Notification остается в приложении.

## Suppress

Сообщение не следует отправлять.

## Cancel

Intent отменен источником или потерял актуальность.

## Expire

Окно полезности завершилось.

## Retry

Допустима повторная техническая попытка.

## Stop Retry

Дальнейшие попытки не имеют смысла или превышен лимит.

## Human Review Required

Требуется ручная проверка.

## Rejected

Intent недействителен, неавторизован или небезопасен.

---

# Preconditions

Перед обработкой Notification Intent должны существовать:

- Источник Intent.
- Source Entity и Version.
- Recipient.
- Recipient Type.
- Purpose.
- Notification Category.
- Priority.
- ExpiresAt или правило срока актуальности, если применимо.
- Deduplication Key.
- Template Reference или утвержденный безопасный текст.
- Privacy Level.
- Policy Version.
- Авторизованный trigger.

---

# Decision Rules

## NP-001: Notification requires a valid domain intent

Notification нельзя создавать только на основании желания интерфейса показать сообщение.

Reason Code: `VALID_NOTIFICATION_INTENT_REQUIRED`

---

## NP-002: Intent must reference its source

Должны быть указаны:

- Source Domain;
- Source Entity;
- Source Entity Id;
- Source Entity Version;
- Causation.

Reason Code: `NOTIFICATION_SOURCE_REFERENCE_REQUIRED`

---

## NP-003: Notification must have a defined recipient

Reason Code: `NOTIFICATION_RECIPIENT_REQUIRED`

---

## NP-004: Recipient identity must be verified

Система должна предотвращать:

- отправку не тому Student;
- использование устаревшего контакта;
- подмену RecipientId;
- смешивание Guardian и Student destinations.

Reason Code: `NOTIFICATION_RECIPIENT_IDENTITY_MISMATCH`

Security Audit обязателен.

---

## NP-005: Intent must be current

Перед отправкой загружается актуальное состояние источника.

Например, Reminder не отправляется, если Homework уже Submitted.

Reason Code: `NOTIFICATION_INTENT_NO_LONGER_CURRENT`

---

## NP-006: Source version must be checked

Устаревший Intent не должен отправляться после изменения:

- Lesson time;
- Homework Due Date;
- Concert Slot;
- Goal state;
- consent request;
- Song Version;
- recipient contact.

Reason Code: `NOTIFICATION_SOURCE_VERSION_CHANGED`

---

## NP-007: Notification must have an expiration rule

Time-sensitive сообщение должно иметь ExpiresAt или вычисляемое окно актуальности.

Reason Code: `NOTIFICATION_EXPIRATION_REQUIRED`

---

## NP-008: Expired messages must not be delivered late

Пример:

- Lesson уже начался;
- Concert Slot уже прошел;
- consent deadline завершен;
- Reminder потерял смысл.

Decision: Expire

Reason Code: `NOTIFICATION_DELIVERY_WINDOW_EXPIRED`

---

## NP-009: Delivery does not create domain truth

Успешная доставка не означает:

- Homework выполнено;
- Student согласен;
- Goal принята;
- Lesson подтвержден;
- Concert Participation подтверждена.

---

## NP-010: Opening does not prove understanding

Opened не является подтверждением, что Recipient понял сообщение.

---

## NP-011: Notification interaction is not educational evidence

Нельзя использовать как Progress Evidence:

- Delivery;
- Open;
- Click;
- time-to-open;
- ignored notification;
- channel preference.

Reason Code: `NOTIFICATION_INTERACTION_NOT_PROGRESS_EVIDENCE`

---

## NP-012: Preferences must be applied by category

Student может отключить Achievement notifications, но оставить Lesson changes.

Настройки не должны применяться как один глобальный флаг без категорий, если модель поддерживает различия.

---

## NP-013: Consent must be validated before external delivery

Reason Code: `NOTIFICATION_CHANNEL_CONSENT_REQUIRED`

---

## NP-014: Contact existence is not consent

Наличие email или номера не дает автоматического права использовать канал.

---

## NP-015: Transactional and essential operational messages require explicit classification

Если сообщение нельзя отключить из-за обязательности процесса, это должно быть:

- документировано;
- ограничено;
- объяснено;
- отделено от Marketing;
- разрешено соответствующей policy.

---

## NP-016: Marketing consent cannot authorize educational surveillance

Согласие на Marketing не дает права:

- отправлять Homework details;
- раскрывать Assessment;
- использовать Progress;
- вовлекать Guardian;
- обходить educational preferences.

---

## NP-017: Withdrawal of channel consent stops future delivery

После отзыва consent будущие внешние Deliveries по этому каналу подавляются.

Уже отправленные сообщения не удаляются автоматически.

---

## NP-018: Quiet hours must be respected

Обычные Educational, Informational и Reminder notifications не отправляются в Quiet Hours.

Reason Code: `RECIPIENT_QUIET_HOURS_ACTIVE`

---

## NP-019: Quiet hours are evaluated in recipient timezone

Reason Code: `RECIPIENT_TIMEZONE_REQUIRED`

---

## NP-020: Quiet-hour messages should be rescheduled when still useful

Если после переноса сообщение остается актуальным: Schedule Delivery

Если нет: Expire

---

## NP-021: High priority does not automatically bypass quiet hours

Bypass должен быть разрешен конкретной категорией.

---

## NP-022: Critical classification must be protected

Teacher, Administrator или AI не должны произвольно ставить Critical.

Reason Code: `NOTIFICATION_PRIORITY_ESCALATION_NOT_AUTHORIZED`

---

## NP-023: Homework reminders are not critical

Они не должны обходить:

- Quiet Hours;
- disabled channel;
- frequency limit;
- learning pause.

---

## NP-024: Last-minute schedule changes may justify urgent delivery

Например:

- Lesson отменен незадолго до начала;
- Concert Slot существенно изменен;
- площадка изменилась;
- Student должен не приезжать.

Но причина срочности должна быть проверяема.

---

## NP-025: Notification frequency must be bounded

Ограничения применяются:

- на Intent type;
- на Category;
- на Recipient;
- на канал;
- на временное окно;
- на весь продукт.

Reason Code: `NOTIFICATION_FREQUENCY_LIMIT_REACHED`

---

## NP-026: Multiple domains must share the same fatigue budget

Homework, Lesson, Goal и Concert не должны независимо исчерпывать внимание Recipient.

---

## NP-027: Lack of response must not increase pressure automatically

Система не должна:

- увеличивать частоту;
- добавлять каналы;
- усиливать тон;
- уведомлять Guardian;
- помечать сообщение Critical;

только потому, что Notification не открыта.

---

## NP-028: Repeated notifications require new value

Повтор оправдан, только если:

- изменилось состояние;
- приближается реальный deadline;
- предыдущая доставка не состоялась;
- требуется подтвержденный следующий шаг;
- повтор разрешен исходной policy.

---

## NP-029: Duplicate intents must be deduplicated

Deduplication Key должен учитывать:

```text
RecipientId
IntentType
SourceEntityId
SourceEntityVersion
RelevantTimeWindow
```

Reason Code: `NOTIFICATION_ALREADY_PROCESSED`

---

## NP-030: Delivery must be idempotent

Повторная обработка одного Intent не должна создавать несколько одинаковых сообщений.

---

## NP-031: Provider callback must be idempotent

Повторный webhook или callback не должен повторно менять доменное состояние.

---

## NP-032: Bundling should reduce fatigue

Можно объединять:

- несколько Homework reminders;
- несколько informational updates;
- Teacher review summary;
- дневной operational digest.

---

## NP-033: Bundling must preserve meaning

Каждый элемент должен сохранять:

- источник;
- действие;
- deadline;
- deep link;
- privacy;
- status.

---

## NP-034: Urgent and non-urgent messages should not be bundled blindly

Срочное изменение Lesson не следует скрывать внутри общего weekly digest.

---

## NP-035: Messages for different privacy levels must not be bundled

Private pedagogical update нельзя объединять с публичной информацией в небезопасном канале.

---

## NP-036: Bundling must not alter domain outcomes

Notification Bundle — представление доставки, а не новый доменный факт.

---

## NP-037: Notification text must be rendered from current data

При отправке следует использовать актуальные безопасные параметры.

Нельзя отправлять stale rendered text после значимого изменения.

---

## NP-038: Rendered content must use versioned templates

Сохраняются:

- TemplateId;
- TemplateVersion;
- locale;
- channel;
- parameters;
- rendered content reference.

---

## NP-039: Template must match channel

Push, email и in-app сообщение могут иметь разную длину и privacy level.

---

## NP-040: Templates must preserve source intent

Шаблон не может:

- усиливать обязательность;
- менять deadline;
- обещать несуществующий результат;
- добавлять наказание;
- искажать Teacher instruction;
- придумывать причину.

---

## NP-041: Artificial urgency is prohibited

Недопустимо:

> Последний шанс!

если это не соответствует реальному deadline.

Reason Code: `ARTIFICIAL_NOTIFICATION_URGENCY_NOT_ALLOWED`

---

## NP-042: Blame and shame language are prohibited

Недопустимо:

- «Вы снова проигнорировали задание»;
- «Вы подводите преподавателя»;
- «Все уже сделали, кроме вас»;
- «Плохая дисциплина».

Reason Code: `NOTIFICATION_TONE_NOT_APPROPRIATE`

---

## NP-043: Student comparison is prohibited

Нельзя сообщать:

- сколько других Student выполнили Homework;
- место в рейтинге;
- кто подтвердил Concert;
- кто получил Achievement;
- кто открыл Notification.

---

## NP-044: Lock-screen previews must minimize data

Push preview может скрывать:

- Song title;
- Assessment;
- Homework details;
- Concert Participation;
- Guardian relationship;
- Teacher private note.

---

## NP-045: Privacy level controls rendering

Допустимые значения:

- Public
- Low Sensitivity
- Private
- Sensitive
- Highly Restricted

Внешние каналы могут поддерживать не все уровни.

---

## NP-046: Sensitive content should default to in-app

Push может содержать только безопасный сигнал:

> У вас новое сообщение в Belcanto.

Полный текст доступен после авторизации.

---

## NP-047: Notification must not expose internal reason codes

Recipient-facing text использует понятное объяснение.

---

## NP-048: Teacher-only data must not appear in student delivery

---

## NP-049: Student data must not leak to other students

Особенно для:

- Group Homework;
- ensemble;
- Concert;
- shared Lesson;
- Goals;
- Achievements.

---

## NP-050: Guardian communication requires explicit scope

Guardian не получает автоматически все Student notifications.

Необходимо определить:

- допустимые категории;
- возрастные правила;
- consent;
- legal basis;
- Student visibility;
- privacy boundaries.

Reason Code: `GUARDIAN_NOTIFICATION_SCOPE_REQUIRED`

---

## NP-051: Guardian delivery must not expose unnecessary pedagogical detail

Можно сообщить:

> Изменилось время занятия.

Но подробные Assessment или private Goal notes требуют отдельного основания.

---

## NP-052: Learning pause may suppress selected categories

Во время Learning Pause могут подавляться:

- Homework Reminder;
- optional practice;
- engagement-oriented messages;
- nonessential Goal prompts.

При этом могут оставаться:

- Lesson cancellation;
- security;
- important operational changes;
- requested Teacher communication.

---

## NP-053: Learning pause must not suppress security notifications

---

## NP-054: Learning pause rules must be category-specific

Reason Code: `NOTIFICATION_SUPPRESSED_BY_LEARNING_PAUSE`

---

## NP-055: Notification channel fallback must be authorized

Fallback применяется только если:

- primary channel failed;
- secondary channel разрешен;
- consent существует;
- privacy level поддерживается;
- сообщение еще актуально;
- retry/fallback limit не превышен.

---

## NP-056: Failure does not authorize every available channel

Push failure не означает автоматическую отправку SMS.

---

## NP-057: Fallback must not duplicate successful delivery

Если primary channel позже подтвердил доставку, pending fallback должен быть отменен, если multi-channel delivery не была намеренной.

---

## NP-058: Multi-channel delivery requires explicit rule

Некоторые сообщения могут быть отправлены по нескольким каналам, например security warning.

Обычные Reminder не должны дублироваться во всех каналах.

---

## NP-059: Retry must be bounded

Retry policy должна ограничивать:

- число попыток;
- интервалы;
- общий срок;
- provider-specific ошибки;
- duplicate risk.

---

## NP-060: Retry requires current intent validation

Перед каждой повторной попыткой проверяется актуальность источника.

---

## NP-061: Permanent failures stop retry

Примеры:

- invalid destination;
- consent withdrawn;
- account removed;
- recipient mismatch;
- unsupported privacy level.

---

## NP-062: Temporary failures may retry

Примеры:

- provider unavailable;
- network timeout;
- rate limit;
- temporary mailbox issue.

---

## NP-063: Retry must not outlive notification expiration

---

## NP-064: Delivery failure does not imply recipient fault

Recipient-facing или Staff-facing объяснение не должно предполагать игнорирование.

---

## NP-065: Delivery incidents may create administrative alerts

Только если:

- затронута значимая категория;
- проблема массовая;
- требуется ручное действие;
- fallback невозможен;
- превышен допустимый failure threshold.

---

## NP-066: Staff alerts must also respect fatigue limits

---

## NP-067: Notification cancellation must propagate

Если исходное событие отменено:

- Scheduled Delivery отменяется;
- Retry прекращается;
- stale in-app action деактивируется;
- история сохраняется.

---

## NP-068: Cancellation after delivery cannot erase delivery history

Но in-app Notification может быть помечено как outdated или resolved.

---

## NP-069: Notification actions must validate current domain state

Кнопка:

`Подтвердить участие`

не должна работать, если Concert Participation уже Withdrawn или deadline завершен.

---

## NP-070: Deep links must be authorized

Recipient должен иметь право открыть целевой ресурс.

---

## NP-071: Notification content must not grant access

Наличие ссылки не заменяет authorization.

---

## NP-072: Open tracking should be minimal

Необходимо избегать избыточного поведенческого мониторинга.

---

## NP-073: Open events must not be used for student ranking

---

## NP-074: Read receipts require transparency

Если продукт показывает Teacher, что Student открыл сообщение, это должно быть:

- явно определено;
- минимально;
- не использоваться как Assessment;
- доступно только при необходимости.

---

## NP-075: AI may draft but not autonomously authorize

AI может:

- предложить текст;
- сократить сообщение;
- локализовать;
- проверить tone;
- обнаружить private data;
- предложить bundling;
- классифицировать Intent;
- рекомендовать delivery window.

AI не может:

- создать доменный смысл;
- изменить Recipient;
- повысить Priority;
- обойти Quiet Hours;
- изменить consent;
- выбрать запрещенный канал;
- добавить давление;
- отправить без Policy decision.

Reason Code: `AI_CANNOT_AUTHORIZE_NOTIFICATION`

---

## NP-076: AI-generated content requires traceability

Сохраняются:

- model or mechanism;
- version;
- input references;
- generated draft;
- confidence;
- validation result;
- human approval, если требуется.

---

## NP-077: AI must not infer sensitive facts for personalization

Нельзя персонализировать сообщение на основе предположений о:

- здоровье;
- эмоциях;
- семейной ситуации;
- мотивации;
- финансовом положении;
- личных отношениях.

---

## NP-078: Localization must preserve meaning

Перевод не должен менять:

- обязательность;
- deadline;
- outcome;
- consent scope;
- privacy warning;
- action semantics.

---

## NP-079: Locale fallback must be safe

Если локализованный template отсутствует:

- использовать утвержденный fallback;
- либо перейти к In-App generic message;
- либо запросить Review.

---

## NP-080: Date and time must include sufficient context

Особенно для:

- Lesson;
- Concert;
- deadlines;
- timezone-sensitive events.

Student должен понимать локальное время.

---

## NP-081: Ambiguous time wording is prohibited

Недопустимо:

> Завтра вечером

если Recipient может увидеть сообщение позже или находится в другом timezone.

Предпочтительно:

> 28 июля в 19:00 по вашему местному времени.

---

## NP-082: Historical notifications remain immutable

После доставки rendered content не переписывается бесследно.

Можно добавить:

- correction;
- updated state;
- superseding notification.

---

## NP-083: Correction notifications must reference the previous message

Если было отправлено неверное время, новое сообщение должно явно объяснить изменение.

---

## NP-084: Silent correction is insufficient for material changes

---

## NP-085: Owner analytics must be aggregated

Owner может видеть:

- delivery rate;
- failure rate;
- category volume;
- bundling rate;
- suppression reasons;
- notification load;
- channel health.

Owner не должен видеть unnecessary recipient-level behavioral details.

---

## NP-086: Notification metrics must not become education scores

---

## NP-087: Notification retention must be limited

Retention зависит от:

- category;
- legal requirements;
- audit need;
- privacy level;
- security need;
- user-facing history.

---

## NP-088: Deletion and archival are separate

Archived Notification может оставаться в Audit.

Удаление регулируется отдельной retention/privacy policy.

---

## NP-089: Backend is authoritative

Client может отображать локальные notifications, но authoritative status находится на backend.

---

## NP-090: Local notifications require synchronization

При изменении Source Entity локальное расписание должно быть отменено или обновлено.

---

## NP-091: Multiple devices must not create duplicate domain delivery

Устройство может показывать несколько push instances, но система должна иметь один authoritative Notification Intent.

---

## NP-092: Device token lifecycle must be managed securely

Invalid или чужой token должен быть отключен.

---

## NP-093: Device ownership change requires revalidation

---

## NP-094: Bulk notifications require additional safeguards

Необходимо:

- preview audience;
- validate category;
- estimate volume;
- verify template;
- validate consent;
- apply rate limits;
- support cancellation;
- preserve audit.

---

## NP-095: Bulk action must not bypass per-recipient rules

Каждый Recipient оценивается отдельно.

---

## NP-096: Test recipients must be isolated

Тестовое сообщение не должно случайно уйти реальным Student.

---

## NP-097: Staff impersonation is prohibited

Notification не должна выглядеть как личное сообщение Teacher, если Teacher его не отправлял и template не объясняет автоматический характер.

---

## NP-098: System-authored messages should be identifiable

Recipient должен понимать, где:

- автоматическое уведомление;
- сообщение Teacher;
- сообщение Administrator;
- AI-assisted draft.

---

## NP-099: Notification should include a useful next action

Если действие требуется, оно должно быть:

- понятным;
- доступным;
- авторизованным;
- актуальным;
- ограниченным по сроку, если применимо.

---

## NP-100: No-action messages should not pretend action is required

Achievement Awarded или Goal Completed могут быть purely informational.

---

# Notification Intent Types

Базовый набор:

```text
LESSON_CREATED
LESSON_RESCHEDULED
LESSON_CANCELLED
LESSON_STARTING_SOON

HOMEWORK_ASSIGNED
HOMEWORK_REMINDER
HOMEWORK_CLARIFICATION_REQUIRED
HOMEWORK_OVERDUE
HOMEWORK_DUE_DATE_EXTENDED
HOMEWORK_EXPIRED
HOMEWORK_REPLACED
HOMEWORK_REVIEWED
HOMEWORK_CORRECTION_REQUESTED

GOAL_CREATED
GOAL_PROGRESS_UPDATED
GOAL_REVIEW_REQUIRED
GOAL_COMPLETED
GOAL_CANCELLED

ACHIEVEMENT_AWARDED
ACHIEVEMENT_REVOKED

SONG_READINESS_UPDATED
SONG_REASSESSMENT_REQUIRED

CONCERT_PARTICIPATION_PROPOSED
CONCERT_CONSENT_REQUIRED
CONCERT_ELIGIBILITY_UPDATED
CONCERT_REHEARSAL_REQUIRED
CONCERT_SLOT_ASSIGNED
CONCERT_SLOT_CHANGED
CONCERT_PROGRAM_PUBLISHED
CONCERT_PARTICIPATION_WITHDRAWN
CONCERT_CANCELLED

TEACHER_REVIEW_REQUIRED
ADMINISTRATOR_ATTENTION_REQUIRED

SECURITY_ALERT
CONTACT_CHANGED
CONSENT_CHANGED
```

---

# Delivery Window

```text
NotificationDeliveryWindow
├── EarliestDeliveryAt
├── LatestDeliveryAt
├── PreferredLocalTime
├── Timezone
├── QuietHoursBehavior
├── AllowedWeekdays
├── Urgency
└── ExpirationBehavior
```

---

# Default Channel Guidance

Это guidance, а не неизменяемое правило.

| Intent | Preferred Channel |
| --- | --- |
| Lesson rescheduled | Push + In-App |
| Lesson cancelled soon | Push + In-App, optional fallback |
| Homework assigned | In-App, optional Push |
| Homework reminder | Push or In-App |
| Goal completed | In-App |
| Achievement awarded | In-App, optional Push |
| Concert consent required | Push + In-App |
| Concert slot changed | Push + In-App |
| Teacher review required | Staff Dashboard, optional Email |
| Security alert | Multi-channel according to security policy |

---

# Bundling Model

```text
NotificationBundle
├── NotificationBundleId
├── RecipientId
├── BundleType
├── IncludedIntentIds
├── Category
├── ScheduledFor
├── Channel
├── TemplateId
├── PrivacyLevel
├── Status
└── Version
```

---

# Bundle Types

- Homework Summary;
- Daily Educational Summary;
- Teacher Review Digest;
- Concert Preparation Summary;
- Administrative Incident Summary.

---

# Bundling Rules

Bundle допустим, если:

- Intent относятся к одному Recipient;
- сроки совместимы;
- privacy level совместим;
- ни один Intent не потеряет срочность;
- действия остаются различимыми;
- сообщение не становится слишком длинным;
- deep links сохраняются.

---

# Retry Model

```text
NotificationRetryPlan
├── RetryPlanId
├── NotificationDeliveryId
├── FailureCategory
├── MaximumAttempts
├── AttemptIntervals
├── RetryUntil
├── FallbackAllowed
├── FallbackChannels
├── CurrentAttempt
└── Status
```

---

# Failure Categories

## Temporary Provider Failure

Retry разрешен.

## Rate Limited

Retry после разрешенного интервала.

## Invalid Destination

Retry прекращается.

## Consent Missing

Delivery подавляется.

## Channel Disabled

Можно использовать разрешенный fallback.

## Rendering Failed

Требуется исправление template или parameters.

## Authorization Failure

Security Review.

## Expired Intent

Retry запрещен.

## Unknown Failure

Ограниченный Retry и возможный administrative alert.

---

# Notification Evaluation Flow

```text
Notification Intent received
        |
        v
Validate source and recipient
        |
        +--> Reject
        |
        v
Load current source version
        |
        +--> Cancel or Expire
        |
        v
Classify category, priority and privacy
        |
        v
Load consent and preferences
        |
        +--> In-App Only
        +--> Suppress
        |
        v
Resolve timezone and delivery window
        |
        +--> Schedule
        +--> Expire
        |
        v
Apply learning pause rules
        |
        +--> Suppress
        |
        v
Apply frequency and fatigue limits
        |
        +--> Bundle
        +--> Reschedule
        +--> Suppress
        |
        v
Select permitted channel
        |
        v
Render current template
        |
        +--> Review Required
        |
        v
Create idempotent Delivery
        |
        v
Send
        |
        +--> Delivered
        +--> Retry
        +--> Fallback
        +--> Failed
```

---

# Send-Time Revalidation

Непосредственно перед отправкой:

```text
Scheduled notification becomes due
        |
        v
Reload source entity
        |
        v
Check source version and current state
        |
        +--> Cancel
        +--> Expire
        |
        v
Reload recipient preferences and consent
        |
        +--> Suppress
        +--> Change Channel
        |
        v
Check quiet hours and frequency limits
        |
        +--> Reschedule
        +--> Bundle
        |
        v
Render or validate rendered content
        |
        v
Send idempotently
```

---

# Commands Produced

## CreateNotificationIntent

Создает доменный запрос на коммуникацию.

## ApproveNotificationIntent

Подтверждает Intent после Policy evaluation или Human Review.

## RejectNotificationIntent

Отклоняет недействительный или небезопасный Intent.

## ScheduleNotification

Планирует Delivery.

## RescheduleNotification

Изменяет время с обязательной причиной.

## CancelNotification

Отменяет Intent или Delivery.

## SuppressNotification

Фиксирует решение не отправлять сообщение.

## BundleNotifications

Объединяет совместимые Intent.

## RenderNotification

Создает channel-specific content.

## SendNotification

Передает Delivery инфраструктурному adapter.

## RetryNotificationDelivery

Создает следующую попытку.

## StopNotificationRetry

Останавливает Retry.

## SwitchNotificationChannel

Переходит на разрешенный fallback.

## MarkNotificationDelivered

Фиксирует техническое подтверждение.

## MarkNotificationDeliveryFailed

Фиксирует ошибку.

## MarkNotificationOpened

Фиксирует доступный сигнал открытия.

## ExpireNotification

Закрывает потерявшее актуальность сообщение.

## ArchiveNotification

Переводит Notification в историческое представление.

## NotifyAdministratorAboutDeliveryIncident

Создает Staff alert при значимом сбое.

## RequestNotificationReview

Создает Human Review.

---

# Domain Events

```text
NotificationIntentCreated
NotificationIntentApproved
NotificationIntentRejected
NotificationScheduled
NotificationRescheduled
NotificationBundled
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
NotificationSuppressed
NotificationCancelled
NotificationExpired
NotificationArchived
NotificationReviewRequested
NotificationDeliveryIncidentDetected
```

## NotificationIntentCreated Event

Должно содержать:

- NotificationIntentId;
- IntentType;
- SourceDomain;
- SourceEntityType;
- SourceEntityId;
- SourceEntityVersion;
- RecipientType;
- RecipientId;
- Category;
- Priority;
- Urgency;
- PrivacyLevel;
- TemplateReference;
- ExpiresAt;
- DeduplicationKey;
- PolicyId;
- PolicyVersion;
- CorrelationId;
- CausationId.

Event не должен содержать готовый private message body, если достаточно template reference и параметров.

## NotificationDelivered Event

Должно содержать:

- NotificationDeliveryId;
- NotificationIntentId;
- RecipientId;
- Channel;
- ProviderReference;
- DeliveredAt;
- AttemptCount;
- TemplateVersion;
- IdempotencyKey;
- CorrelationId;
- CausationId.

## NotificationSuppressed Event

Должно содержать:

- NotificationIntentId;
- RecipientId;
- Category;
- SuppressionReason;
- SuppressedAt;
- SourceVersion;
- PolicyId;
- PolicyVersion.

---

# Suppression Reasons

```text
NOTIFICATION_INTENT_NO_LONGER_CURRENT
NOTIFICATION_SOURCE_VERSION_CHANGED
NOTIFICATION_ALREADY_PROCESSED
NOTIFICATION_CHANNEL_CONSENT_REQUIRED
NOTIFICATION_CATEGORY_DISABLED
NOTIFICATION_CHANNEL_DISABLED
RECIPIENT_QUIET_HOURS_ACTIVE
NOTIFICATION_SUPPRESSED_BY_LEARNING_PAUSE
NOTIFICATION_FREQUENCY_LIMIT_REACHED
NOTIFICATION_DELIVERY_WINDOW_EXPIRED
NOTIFICATION_PRIVACY_LEVEL_UNSUPPORTED
NOTIFICATION_RECIPIENT_UNAVAILABLE
NOTIFICATION_BUNDLED
NOTIFICATION_CANCELLED_BY_SOURCE
NOTIFICATION_TEMPLATE_INVALID
NOTIFICATION_DESTINATION_INVALID
```

---

# Human Review

Human Review может потребоваться, если:

- Priority повышена до Critical;
- Intent содержит Sensitive content;
- Guardian добавляется как Recipient;
- template отсутствует;
- AI существенно изменил смысл;
- bulk audience большой;
- recipient scope неоднозначен;
- сообщение связано с конфликтом;
- исправляется ранее отправленная ошибка;
- требуется multi-channel delivery;
- consent state неоднозначен;
- communication bypasses Quiet Hours;
- delivery incident затрагивает важное мероприятие;
- Staff хочет отправить нестандартный текст;
- source data содержит противоречия.

---

# Human Review Result

Reviewer может:

- Approve;
- Reject;
- изменить Recipient scope;
- изменить Privacy Level;
- изменить Priority;
- выбрать канал;
- потребовать In-App Only;
- исправить template;
- изменить Delivery Window;
- разрешить multi-channel delivery;
- снять bundling;
- заменить message body;
- запросить новый domain intent.

Reviewer не должен изменять исходный доменный факт без соответствующей доменной команды.

---

# Source Policy Integration

## Lesson

Lesson Policy создает Intent при:

- создании Lesson;
- переносе;
- отмене;
- изменении Teacher;
- изменении места или формата.

Notification Policy не решает, допустим ли перенос Lesson.

## Homework Reminder

Homework Reminder Policy определяет необходимость Reminder.

Notification Policy применяет:

- канал;
- Quiet Hours;
- bundling;
- frequency limits;
- privacy;
- retry.

## Homework Expiration

Homework Expiration Policy создает Intent для:

- Overdue;
- Extension;
- Expiration;
- Replacement;
- Reopen.

Notification Policy не определяет lifecycle Homework.

## Goal Completion

Goal Completion Policy решает, завершена ли Goal.

Notification Policy только доставляет результат.

## Achievement Award

Achievement Award Policy создает уведомление после подтвержденного Award.

Notification Policy не может Award или Revoke Achievement.

## Song Readiness

Song Readiness Policy определяет readiness outcome.

Notification Policy не должна отправлять внутренние педагогические оценки без разрешенного Student-facing explanation.

## Concert Eligibility

Concert Eligibility Policy создает Intent для:

- consent;
- rehearsal;
- eligibility status;
- Slot;
- Program;
- withdrawal;
- cancellation.

Notification Policy повторно проверяет актуальность Participation перед доставкой.

---

# Student Presentation

Student должен видеть:

- понятное сообщение;
- источник;
- дату и время;
- статус;
- целевое действие;
- deep link;
- возможность управлять category preferences;
- Quiet Hours;
- каналы;
- history в разумном объеме;
- причину обязательности, если сообщение нельзя отключить.

Student не должен видеть:

- provider details;
- retry count;
- internal Reason Codes;
- private Teacher notes;
- скрытые segmentation attributes;
- AI confidence;
- чужие Notification states.

---

# Teacher Presentation

Teacher должен видеть:

- отправленные им сообщения;
- system-generated Teacher alerts;
- delivery failure, если это требует действия;
- pending review;
- category и recipient scope;
- scheduled communication;
- cancellation;
- текущий источник Intent.

Teacher не должен видеть поведенческую аналитику Student без образовательной необходимости.

---

# Administrator Presentation

Administrator может видеть:

- delivery state;
- provider failure;
- invalid destinations;
- bulk incidents;
- suppression reason;
- retry state;
- queue delay;
- template error;
- channel health.

Доступ к содержимому Sensitive notifications ограничивается.

---

# Owner Analytics

Допустимые агрегаты:

- количество Intent;
- количество Delivered;
- Delivery failure rate;
- suppression rate;
- bundling rate;
- frequency-limit activations;
- category volume;
- channel distribution;
- template failure rate;
- Quiet Hours reschedules;
- notification load per day;
- incident count.

Недопустимо использовать:

- open rate как оценку Student;
- ignored notifications как мотивационный score;
- notification behavior для наказания;
- ranking Teacher только по click-through rate;
- скрытую психологическую сегментацию.

---

# AI Assistance

AI может:

- создавать Draft template;
- сокращать текст;
- локализовать;
- проверять tone;
- искать private data;
- определять возможный duplicate;
- предлагать bundle;
- предлагать safe preview;
- классифицировать urgency;
- находить неоднозначную дату;
- предлагать Human Review.

AI не может:

- авторизовать Intent;
- выбрать Recipient без source data;
- придумывать consent;
- менять source outcome;
- ставить Critical;
- обходить frequency limits;
- отправлять самостоятельно;
- персонализировать давление;
- использовать скрытые уязвимости;
- делать медицинские или психологические выводы.

---

# Privacy

Notification может раскрывать:

- факт обучения;
- Lesson schedule;
- Homework;
- Goal;
- Assessment context;
- Song;
- Concert participation;
- Guardian relationship;
- contact details;
- behavior metadata.

Необходимо:

- минимизировать внешние previews;
- хранить destination безопасно;
- отделять rendered content от audit metadata;
- ограничивать Staff access;
- скрывать private параметры;
- защищать minor data;
- контролировать Guardian delivery;
- не включать personal reasons в analytics;
- поддерживать retention limits;
- аудитировать изменение контактов.

---

# Security

Необходимо защищать:

- подмену Recipient;
- отправку на старый contact;
- использование чужого device token;
- mass-send ошибку;
- template injection;
- unsafe deep links;
- повторную доставку;
- подделку provider callback;
- unauthorized priority escalation;
- consent bypass;
- PII leakage;
- cross-tenant delivery;
- неавторизованное чтение Staff;
- изменение исторического rendered content.

---

# Audit Requirements

Для Intent сохраняются:

- IntentId;
- Source Domain;
- Source Entity;
- Source Version;
- Recipient;
- Category;
- Purpose;
- Priority;
- Urgency;
- Privacy Level;
- requested channels;
- Template Reference;
- parameters reference;
- ExpiresAt;
- Deduplication Key;
- Actor;
- Policy Version;
- CreatedAt;
- CorrelationId;
- CausationId.

Для Delivery:

- Channel;
- destination reference;
- ScheduledFor;
- Template Version;
- rendered content hash or reference;
- Idempotency Key;
- provider;
- attempt history;
- DeliveredAt;
- FailedAt;
- Failure Code;
- fallback;
- suppression;
- cancellation;
- expiration.

Для Human Review:

- Reviewer;
- previous state;
- decision;
- changed fields;
- reason;
- timestamp.

Для AI:

- model;
- version;
- input references;
- generated proposal;
- validation result;
- human confirmation.

---

# Failure Modes

## Intent not found

- Decision: Rejected
- Reason Code: NOTIFICATION_INTENT_NOT_FOUND

## Recipient not found

- Decision: Rejected
- Reason Code: NOTIFICATION_RECIPIENT_NOT_FOUND

## Recipient mismatch

- Decision: Rejected
- Reason Code: NOTIFICATION_RECIPIENT_IDENTITY_MISMATCH

Security Audit обязателен.

## Source entity missing

- Decision: Cancelled or Rejected
- Reason Code: NOTIFICATION_SOURCE_NOT_FOUND

## Source version changed

- Decision: Cancelled (или создается новый Intent)
- Reason Code: NOTIFICATION_SOURCE_VERSION_CHANGED

## Consent missing

- Decision: In-App Only or Suppressed
- Reason Code: NOTIFICATION_CHANNEL_CONSENT_REQUIRED

## Category disabled

- Decision: Suppressed
- Reason Code: NOTIFICATION_CATEGORY_DISABLED

## Quiet Hours active

- Decision: Reschedule Delivery or Expire
- Reason Code: RECIPIENT_QUIET_HOURS_ACTIVE

## Frequency limit reached

- Decision: Bundle, Reschedule, or Suppress
- Reason Code: NOTIFICATION_FREQUENCY_LIMIT_REACHED

## Delivery window expired

- Decision: Expire
- Reason Code: NOTIFICATION_DELIVERY_WINDOW_EXPIRED

## Template missing

- Decision: Human Review Required
- Reason Code: NOTIFICATION_TEMPLATE_NOT_FOUND

## Template rendering failed

- Decision: Delivery Failed
- Reason Code: NOTIFICATION_RENDERING_FAILED

## Privacy level unsupported by channel

- Decision: Change Channel or In-App Only
- Reason Code: NOTIFICATION_PRIVACY_LEVEL_UNSUPPORTED

## Invalid destination

- Decision: Stop Retry
- Reason Code: NOTIFICATION_DESTINATION_INVALID

## Temporary provider failure

- Decision: Retry
- Reason Code: NOTIFICATION_PROVIDER_TEMPORARILY_UNAVAILABLE

## Permanent provider failure

- Decision: Stop Retry
- Reason Code: NOTIFICATION_PROVIDER_PERMANENT_FAILURE

## Duplicate intent

- Decision: No Change Required
- Reason Code: NOTIFICATION_ALREADY_PROCESSED

## Concurrent update

- Decision: Deferred
- Reason Code: NOTIFICATION_VERSION_CONFLICT

Policy повторно оценивает актуальное состояние.

---

# Explainability Examples

## Quiet Hours

> Уведомление перенесено на утро, потому что текущее время входит в ваш период без уведомлений.

## Bundled

> Несколько напоминаний объединены в одно сообщение, чтобы уменьшить количество уведомлений.

## Channel Disabled

> Push-уведомления отключены, поэтому сообщение доступно только в приложении.

## Source Changed

> Уведомление отменено, потому что время занятия уже изменилось повторно.

## Expired

> Сообщение больше не отправляется, потому что связанное событие уже началось.

## Privacy-Safe Push

> Подробности доступны в приложении Belcanto.

## Delivery Failed

> Сообщение не удалось доставить через выбранный канал. Это не влияет на ваш учебный статус.

---

# Examples

## Example 1: Homework reminder during quiet hours

Дано:

- Reminder разрешен;
- ScheduledFor: 23:30;
- Quiet Hours: 22:00–08:00;
- Homework актуально до следующего вечера.

Результат:

- Decision: Reschedule Delivery
- New Time: 08:30
- Reason Code: RECIPIENT_QUIET_HOURS_ACTIVE

## Example 2: Homework completed before send

Дано:

- Reminder уже Scheduled;
- Student отправил Homework;
- Delivery еще не началась.

Результат:

- Decision: Cancel Notification
- Reason Code: NOTIFICATION_INTENT_NO_LONGER_CURRENT

## Example 3: Two homework reminders

Дано:

- два Intent на один вечер;
- одинаковая privacy category;
- deadlines не срочные;
- bundling разрешен.

Результат:

- Decision: Bundle
- Bundle Type: Homework Summary

## Example 4: Concert slot changed urgently

Дано:

- Concert через три часа;
- Slot перенесен;
- Push разрешен;
- Student доступен;
- сообщение актуально.

Результат:

- Decision: Deliver Now
- Priority: High
- Urgency: High
- Channels:
  - Push
  - In-App

## Example 5: Sensitive assessment update

Дано:

- Notification содержит private pedagogical detail;
- Push preview доступен на lock screen.

Результат:

- Decision: In-App Only
- Push: "У вас новое сообщение в Belcanto."

## Example 6: Push token invalid

Дано:

- Push failed permanently;
- email fallback разрешен;
- consent существует;
- Intent еще актуален.

Результат:

- Decision: Change Channel
- New Channel: Email
- Reason Code: NOTIFICATION_DESTINATION_INVALID

## Example 7: Email fallback forbidden

Дано:

- Push failed;
- email существует;
- Student не разрешил email category.

Результат:

- Decision: Stop Retry
- Reason Code: NOTIFICATION_CHANNEL_CONSENT_REQUIRED

Notification остается в In-App Inbox.

## Example 8: Goal completed

Дано:

- Goal Completion Policy подтвердила Goal;
- Notification Informational;
- Student отключил push для достижений.

Результат:

- Decision: In-App Only

## Example 9: Duplicate provider callback

Дано:

- provider дважды сообщил Delivered;
- Delivery уже в статусе Delivered.

Результат:

- Decision: No Change Required
- Reason Code: NOTIFICATION_ALREADY_PROCESSED

## Example 10: Guardian not authorized

Дано:

- Student несовершеннолетний;
- Homework reminder предназначен Student;
- Guardian notification scope отсутствует.

Результат:

- Decision: Suppress Guardian Delivery
- Reason Code: GUARDIAN_NOTIFICATION_SCOPE_REQUIRED

## Example 11: Late lesson cancellation

Дано:

- Lesson через 30 минут;
- Teacher отменил;
- Quiet Hours активны;
- сообщение предотвращает бесполезную поездку.

Результат:

- Decision: Deliver Now
- Priority: High
- Urgency: High
- Reason: Operational exception configured for imminent cancellation

Exception должна быть аудирована.

## Example 12: AI-generated urgent wording

Дано:

- AI предлагает «Срочно! Последний шанс подтвердить участие»;
- deadline через пять дней;
- срочность не подтверждена.

Результат:

- Decision: Draft Rejected
- Reason Code: ARTIFICIAL_NOTIFICATION_URGENCY_NOT_ALLOWED

---

# Test Requirements

## Intent Validation Tests

- valid Intent is accepted;
- missing Recipient is rejected;
- missing Source is rejected;
- stale Source Version is cancelled;
- expired Intent is not delivered;
- duplicate Intent is idempotent;
- unauthorized Intent is rejected.

## Preference Tests

- enabled category can deliver;
- disabled category is suppressed;
- channel order is respected;
- consent is checked;
- withdrawal stops future delivery;
- essential category follows explicit rule;
- Marketing and Educational consent remain separate.

## Quiet Hours Tests

- ordinary Notification is rescheduled;
- expired message is not sent after Quiet Hours;
- critical security alert can bypass when configured;
- Homework Reminder cannot bypass;
- recipient timezone is used;
- timezone change recalculates scheduled delivery.

## Frequency Tests

- product-wide limit applies across domains;
- repeated Reminder is suppressed;
- new meaningful state change can create new Intent;
- lack of open does not increase frequency;
- multiple compatible Intent are bundled;
- urgent message is not hidden in digest.

## Bundling Tests

- compatible Homework reminders bundle;
- different Recipients never bundle;
- different privacy levels do not bundle;
- urgent and low-priority messages remain separate;
- every bundled action remains accessible;
- bundle delivery remains idempotent.

## Channel Tests

- allowed Push is used;
- disabled Push falls back only when authorized;
- invalid Push token stops further Push attempts;
- email consent is required;
- SMS is not used implicitly;
- Sensitive content defaults to In-App;
- multi-channel delivery requires explicit configuration.

## Rendering Tests

- correct Template Version is used;
- source intent is preserved;
- false urgency is rejected;
- private parameters are excluded from push;
- locale is applied;
- date formatting is unambiguous;
- missing template creates Review.

## Delivery Tests

- successful provider response marks Delivered;
- Delivered does not mark Opened;
- Opened does not complete domain action;
- duplicate callback is harmless;
- temporary failure retries;
- permanent failure stops retry;
- retry does not outlive ExpiresAt.

## Cancellation Tests

- source cancellation cancels scheduled Delivery;
- changed Lesson time cancels old Intent;
- Homework Submission cancels Reminder;
- withdrawn Concert Participation cancels related actions;
- delivered history remains immutable;
- stale deep link action is rejected.

## Learning Pause Tests

- Homework reminders are suppressed;
- security alert remains deliverable;
- imminent Lesson cancellation remains deliverable;
- optional practice is suppressed;
- learning pause ending does not resend expired messages;
- category-specific rules are applied.

## Guardian Tests

- Guardian receives only allowed categories;
- missing scope suppresses delivery;
- Guardian contact requires consent;
- private Assessment is not disclosed;
- Student and Guardian destinations remain separate;
- Guardian withdrawal stops delivery.

## Privacy Tests

- lock-screen preview is minimal;
- Teacher notes are excluded;
- Student data does not leak to peers;
- sensitive message uses In-App;
- analytics are aggregated;
- rendered content access is restricted;
- wrong destination is rejected.

## Security Tests

- RecipientId cannot be substituted;
- device token ownership is validated;
- forged provider callback is rejected;
- unauthorized Critical priority is rejected;
- bulk send requires safeguards;
- deep links require authorization;
- cross-tenant delivery is impossible.

## AI Tests

- AI can draft text;
- AI cannot authorize delivery;
- AI cannot invent consent;
- AI cannot raise Priority;
- AI cannot bypass Quiet Hours;
- inferred sensitive personalization is rejected;
- AI metadata is stored;
- source meaning remains unchanged.

## Explainability Tests

- suppression reason is understandable;
- reschedule reason is understandable;
- channel fallback is explainable;
- delivery failure does not blame Recipient;
- bundled Notification explains included actions;
- expired Intent explains why no delivery occurred;
- internal Reason Codes are hidden from recipient.

---

# Non-Goals

Notification Policy не определяет:

- педагогический смысл события;
- Lesson lifecycle;
- Homework lifecycle;
- Homework Reminder necessity;
- Homework Expiration outcome;
- Progress Calculation;
- Goal Completion;
- Achievement Award;
- Song Readiness;
- Concert Eligibility;
- CRM marketing campaigns;
- provider implementation;
- billing;
- SMS pricing;
- email reputation strategy;
- message copywriting brand guide in full;
- data deletion policy;
- legal consent model in full;
- emergency services communication;
- medical notifications.

---

# Open Questions

Необходимо определить:

- какие каналы входят в MVP;
- является ли In-App Inbox обязательным;
- поддерживаются ли local device notifications;
- нужен ли email;
- нужен ли SMS;
- какие categories Student может отключать;
- какие сообщения считаются essential;
- какие сообщения могут обходить Quiet Hours;
- стандартные Quiet Hours;
- стандартный frequency limit;
- как считать fatigue budget;
- какие Intent можно bundle;
- нужен ли daily digest;
- нужен ли weekly digest;
- как Teacher управляет уведомлениями;
- может ли Teacher отправлять свободный текст;
- кто утверждает Templates;
- как версионировать Templates;
- какие locale поддерживаются;
- какой fallback locale;
- нужно ли хранить rendered body;
- сколько хранить Notification history;
- нужны ли open receipts;
- видит ли Teacher open status;
- нужна ли прозрачность tracking;
- какие delivery providers использовать;
- как обрабатывать provider outage;
- какие retry интервалы;
- максимальное число retry;
- когда использовать fallback channel;
- разрешена ли intentional multi-channel delivery;
- как синхронизировать несколько устройств;
- как удалять invalid device token;
- как проверять ownership нового устройства;
- нужен ли web push;
- как реализовать desktop notifications;
- как обрабатывать account logout;
- должны ли notifications исчезать после logout;
- как обрабатывать смену email;
- нужна ли повторная верификация email;
- как обрабатывать bounced email;
- как управлять Guardian scope;
- какие категории доступны Guardian;
- видит ли Student сообщения Guardian;
- нужны ли Staff digests;
- как эскалировать массовый delivery failure;
- кто получает operational incident alerts;
- как отделить Marketing;
- нужен ли отдельный Marketing Consent aggregate;
- нужна ли отдельная Template Policy;
- нужна ли отдельная Notification Fatigue Policy;
- нужна ли отдельная Communication Consent Policy;
- как классифицировать Security notifications;
- нужна ли отдельная Security Notification Policy;
- может ли AI локализовать сообщения автоматически;
- какие AI drafts требуют Human Review;
- как тестировать tone;
- как предотвращать artificial urgency;
- как проводить bulk-send preview;
- нужен ли staged rollout;
- можно ли отменить bulk send;
- как хранить Deduplication Key;
- как моделировать Notification Bundle;
- является ли Notification отдельным aggregate;
- является ли Delivery отдельным aggregate;
- нужен ли Outbox pattern;
- нужен ли Inbox pattern;
- какие события публикуются наружу;
- как избежать утечки rendered content через event bus;
- как обрабатывать deleted Source Entity;
- что происходит с Notification после удаления аккаунта;
- какие данные остаются в Audit;
- как разделить Audit и user-facing history.

---

# History

| Version | Description |
| --- | --- |
| 1.0.0 | Определены правила создания Intent, доставки, scheduling, bundling, consent, privacy, retry и suppression для всех доменных уведомлений. |
