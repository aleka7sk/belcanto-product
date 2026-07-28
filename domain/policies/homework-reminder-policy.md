---
Status: Approved
Version: 1.0.0
Last Updated: 2026-07-27

Policy Id: HOMEWORK_REMINDER_POLICY

Policy Type:
  - Reaction Policy
  - Scheduling Policy
  - Recommendation Policy
  - Escalation Policy
  - Communication Policy

Owners:
  - Product Owner
  - Education Lead
  - Technical Lead

Related Aggregates:
  - Homework
  - HomeworkAssignment
  - Student
  - Teacher
  - Lesson
  - NotificationPreference
  - ReminderSchedule
  - LearningActivity

Observed Events:
  - HomeworkAssigned
  - HomeworkUpdated
  - HomeworkDueDateChanged
  - HomeworkSubmitted
  - HomeworkReviewed
  - HomeworkCancelled
  - HomeworkExpired
  - LessonCompleted
  - LessonRescheduled
  - LessonCancelled
  - StudentNotificationPreferencesChanged
  - StudentTimezoneChanged
  - StudentLearningPauseStarted
  - StudentLearningPauseEnded
  - ReminderDeliveryFailed
  - ReminderDeliveryConfirmed
  - HomeworkReminderRecalculationRequested

Produced Commands:
  - ScheduleHomeworkReminder
  - RescheduleHomeworkReminder
  - CancelHomeworkReminder
  - SendHomeworkReminder
  - SuppressHomeworkReminder
  - RequestHomeworkClarification
  - NotifyTeacherAboutHomeworkBlocker
  - RecalculateHomeworkReminderPlan
  - MarkHomeworkReminderDeliveryFailed
  - MarkHomeworkReminderDelivered

Related Documents:
  - 000-domain-policy-overview.md
  - lesson-completion-policy.md
  - progress-update-policy.md
  - homework-expiration-policy.md
  - notification-policy.md
  - periodic-review-policy.md
  - ../homework.md
  - ../lesson.md
  - ../student.md
---

# Homework Reminder Policy

> Homework Reminder Policy определяет, когда, кому и в какой форме следует напоминать о назначенном Homework.
>
> Политика поддерживает самостоятельную работу Student, но не превращает обучение в систему давления, наказаний или навязчивых уведомлений.

---

# Purpose

Homework может быть назначено между Lesson, но Student может:

- забыть о задании;
- не понять формулировку;
- не знать срок;
- откладывать начало;
- не понимать приоритет;
- потерять ссылку на материал;
- не иметь возможности выполнить задание вовремя;
- завершить работу, но не отметить Submission;
- получить слишком много уведомлений;
- находиться в паузе или отпуске;
- получать напоминания в неподходящее время.

Без отдельной политики система может:

- отправлять напоминания после выполнения;
- уведомлять ночью;
- дублировать сообщения;
- напоминать о просроченном или отмененном задании;
- создавать чувство вины;
- подменять Teacher;
- игнорировать предпочтения Student;
- раскрывать чувствительную информацию в push-тексте;
- автоматически считать отсутствие Submission признаком отсутствия практики.

Homework Reminder Policy должна обеспечивать своевременную, уважительную и объяснимую коммуникацию.

---

# Core Principle

Reminder — это поддержка следующего полезного действия, а не наказание за бездействие.

```text
Active Homework
      +
Relevant Timing
      +
Student Preferences
      +
Current Homework State
      +
Communication Constraints
      =
Permitted Reminder
```

Напоминание не должно отправляться только потому, что технически наступило запланированное время.

Перед отправкой необходимо повторно проверить актуальное состояние Homework.

---

## Reminder Is Not an Assessment

Отсутствие реакции на Reminder не означает:

- отказ выполнять Homework;
- отсутствие мотивации;
- низкую дисциплину;
- ухудшение Progress;
- неуважение к Teacher;
- отсутствие самостоятельной практики.

Reminder Delivery и образовательная оценка являются разными процессами.

---

# Homework Reminder Plan

Рекомендуемая концептуальная структура:

```text
HomeworkReminderPlan
├── ReminderPlanId
├── HomeworkAssignmentId
├── StudentId
├── TeacherId
├── HomeworkVersion
├── Status
├── Strategy
├── ScheduledReminders
├── MaximumReminderCount
├── ReminderWindow
├── QuietHours
├── Timezone
├── PreferredChannels
├── EscalationRules
├── CreatedAt
├── RecalculatedAt
├── PolicyId
├── PolicyVersion
└── Version
```

---

# Scheduled Reminder

```text
ScheduledHomeworkReminder
├── ReminderId
├── ReminderPlanId
├── HomeworkAssignmentId
├── StudentId
├── ReminderType
├── ScheduledFor
├── Timezone
├── Channel
├── Status
├── MessageTemplateId
├── HomeworkVersion
├── DueDateReference
├── TriggerReference
├── IdempotencyKey
├── DeliveredAt
├── SuppressedAt
├── SuppressionReason
└── Version
```

---

# Reminder Status

Допустимые состояния:

- Planned
- Scheduled
- Due
- Sending
- Delivered
- Delivery Failed
- Suppressed
- Cancelled
- Expired

## Planned

Reminder рассчитан политикой, но еще не передан в scheduling infrastructure.

## Scheduled

Reminder поставлен на выполнение в определенное время.

## Due

Наступило время повторной проверки и возможной отправки.

Due не означает автоматическую доставку.

## Sending

Команда передачи уведомления принята коммуникационной инфраструктурой.

## Delivered

Канал подтвердил техническую доставку в пределах доступной модели подтверждения.

Это не означает, что Student прочитал сообщение.

## Delivery Failed

Попытка доставки неуспешна.

Повтор зависит от Notification Policy и Retry Policy.

## Suppressed

Reminder сознательно не отправлен по доменной причине.

## Cancelled

Reminder отменен до наступления времени.

## Expired

Время Reminder прошло, а отправка больше не имеет образовательного смысла.

---

# Reminder Types

## Assignment Created Reminder

Сообщает о новом Homework после его назначения.

Обычно это не отдельное «напоминание», а подтверждение назначения.

## Start Reminder

Помогает Student начать работу заранее.

Пример:

> До следующего занятия осталось несколько дней. Можно начать с короткого первого шага.

## Midpoint Reminder

Отправляется между назначением и сроком, если это оправдано продолжительностью задания.

## Due Soon Reminder

Сообщает о приближении срока.

## Due Today Reminder

Может использоваться, если Student разрешил такой тип коммуникации и задание еще актуально.

## Pre-Lesson Reminder

Напоминает о Homework перед следующим связанным Lesson.

## Clarification Reminder

Не напоминает «сделать», а предлагает уточнить непонятное задание.

## Resume Reminder

Может быть создан после окончания Learning Pause, если Homework все еще актуально.

## Teacher Attention Reminder

Это не сообщение Student.

Оно информирует Teacher, что Homework требует внимания, например:

- Student сообщил о блокере;
- срок прошел;
- задание невозможно выполнить;
- отсутствует необходимый материал;
- повторно не удалось доставить важное уведомление.

---

# Reminder Strategy

Homework может использовать одну из стратегий:

- No Reminders
- Single Reminder
- Due-Based
- Lesson-Based
- Adaptive
- Teacher Defined
- Student Defined

## No Reminders

Система не создает автоматические напоминания.

Это может быть:

- выбор Student;
- тип Homework;
- решение Teacher;
- короткий срок;
- чувствительный контекст;
- ручное задание без цифрового сопровождения.

## Single Reminder

Создается одно напоминание в выбранное время.

## Due-Based

Расписание определяется относительно Due Date.

Пример:

- 48 hours before due date
- 6 hours before due date

## Lesson-Based

Расписание определяется относительно следующего связанного Lesson.

Пример:

- Evening before lesson
- Morning of lesson

## Adaptive

Система может предложить расписание на основе:

- длительности до срока;
- предполагаемого объема;
- предпочтений Student;
- истории взаимодействия с Reminder;
- типа Homework;
- расписания Lesson.

Adaptive Strategy не должна использовать психологическое давление или скрытую оптимизацию вовлеченности.

## Teacher Defined

Teacher явно задает Reminder Plan.

## Student Defined

Student выбирает удобное время или количество напоминаний.

---

# Trigger

Политика применяется при:

```text
HomeworkAssigned
HomeworkUpdated
HomeworkDueDateChanged
HomeworkSubmitted
HomeworkReviewed
HomeworkCancelled
HomeworkExpired
LessonCompleted
LessonRescheduled
LessonCancelled
StudentNotificationPreferencesChanged
StudentTimezoneChanged
StudentLearningPauseStarted
StudentLearningPauseEnded
ReminderDeliveryFailed
ReminderDeliveryConfirmed
HomeworkReminderRecalculationRequested
```

---

# Inputs

Для оценки могут потребоваться:

- Homework Assignment;
- Homework Status;
- Homework Version;
- Student;
- Teacher;
- Assignment Date;
- Due Date;
- связанный Lesson;
- следующий Lesson;
- ожидаемая длительность Homework;
- Homework Type;
- обязательность;
- Student Timezone;
- Quiet Hours;
- Preferred Channels;
- Notification Preferences;
- Learning Pause;
- уже существующие Reminders;
- Reminder delivery history;
- Student-reported blocker;
- Teacher-defined strategy;
- Policy Version.

---

# Preconditions

- Homework Assignment существует.
- Student существует.
- Homework относится к Student.
- Homework находится в состоянии, допускающем напоминание.
- Homework Version известна.
- Timezone определена или применяется утвержденное значение по умолчанию.
- Reminder Strategy разрешена.
- Канал коммуникации доступен.
- Actor или Event имеет право инициировать расчет.
- Policy Version доступна.

---

# Homework States Relevant to Reminders

Напоминания могут рассматриваться для:

- Assigned;
- In Progress;
- Clarification Required;
- Submitted with Correction Requested.

Напоминания обычно не отправляются для:

- Reviewed;
- Completed;
- Cancelled;
- Expired;
- Replaced;
- Archived.

---

# Decision Outcomes

## Reminder Plan Created

Создан новый план.

## Reminder Scheduled

Конкретный Reminder запланирован.

## Reminder Rescheduled

Время изменено из-за новых данных.

## Reminder Cancelled

Напоминание больше не актуально.

## Reminder Suppressed

Напоминание не отправляется из-за доменного ограничения.

## Reminder Sent

Отправка разрешена.

## No Reminder Required

Напоминание не требуется.

## Clarification Required

Вместо обычного напоминания необходимо предложить уточнение.

## Teacher Attention Required

Необходимо информировать Teacher.

## Deferred

Решение отложено до появления данных или завершения паузы.

## Rejected

Запрос недействителен или неавторизован.

---

# Decision Rules

## HR-001: Reminder requires an active homework assignment

Напоминание нельзя создавать без активного Homework Assignment.

Reason Code: `ACTIVE_HOMEWORK_REQUIRED`

---

## HR-002: Reminder must reference a concrete homework version

Каждый Reminder должен быть связан с конкретной версией Homework.

Reason Code: `HOMEWORK_VERSION_REQUIRED`

---

## HR-003: Completed homework must not receive completion reminders

После подтвержденного Submission или Completion обычные напоминания о выполнении отменяются.

Reason Code: `HOMEWORK_ALREADY_COMPLETED`

Если Submission требует исправления, должен быть создан новый Reminder Context.

---

## HR-004: Reviewed homework must not receive pending reminders

После HomeworkReviewed все Reminders, относящиеся к предыдущему состоянию, отменяются.

---

## HR-005: Cancelled homework must not receive reminders

Reason Code: `HOMEWORK_CANCELLED`

---

## HR-006: Expired homework must not receive ordinary reminders

После HomeworkExpired дальнейшая коммуникация определяется Homework Expiration Policy и Notification Policy.

Reason Code: `HOMEWORK_EXPIRED`

---

## HR-007: Replaced homework invalidates old reminders

Если Homework заменено новой версией или новым Assignment, старые Reminders отменяются.

Reason Code: `HOMEWORK_REPLACED`

---

## HR-008: Reminder must be revalidated at send time

Непосредственно перед отправкой необходимо повторно проверить:

- Homework Status;
- Homework Version;
- Due Date;
- Student preferences;
- Quiet Hours;
- Learning Pause;
- существование Assignment;
- отсутствие недавней Submission;
- отсутствие Cancellation.

Reason Code при отмене: `REMINDER_NO_LONGER_CURRENT`

---

## HR-009: Assignment confirmation and reminder are separate

Сообщение о создании Homework не должно считаться одним из ограниченного числа напоминаний, если политика не определяет иначе.

---

## HR-010: Due date is optional only when strategy supports it

Для Due-Based Strategy Due Date обязательна.

Если Due Date отсутствует, можно использовать:

- Lesson-Based;
- Teacher Defined;
- Student Defined;
- No Reminders.

Reason Code: `HOMEWORK_DUE_DATE_REQUIRED_FOR_STRATEGY`

---

## HR-011: Reminder timing must allow meaningful action

Напоминание не следует отправлять, если Student уже не может разумно выполнить задание до срока.

Вместо:

> Срочно выполните задание за 15 минут.

система может предложить:

> До занятия осталось мало времени. Можно сообщить преподавателю, что задание требует переноса или уточнения.

Reason Code: `INSUFFICIENT_TIME_FOR_MEANINGFUL_ACTION`

---

## HR-012: Immediate reminders should not duplicate assignment delivery

Если Homework назначено только что, отдельный Reminder не должен отправляться сразу после Assignment notification.

Reason Code: `REMINDER_TOO_CLOSE_TO_ASSIGNMENT`

---

## HR-013: Reminders must respect quiet hours

Напоминание не отправляется в Quiet Hours, кроме явно разрешенного критического сценария.

Homework Reminder не является критическим уведомлением.

Reason Code: `STUDENT_QUIET_HOURS_ACTIVE`

---

## HR-014: Quiet-hour reminders must be shifted, not silently lost

Если время попадает в Quiet Hours, Reminder переносится на ближайшее допустимое окно, если после переноса он сохраняет смысл.

---

## HR-015: Student timezone is authoritative for student-facing timing

Напоминание рассчитывается в актуальном Timezone Student.

Reason Code: `STUDENT_TIMEZONE_REQUIRED`

---

## HR-016: Timezone changes trigger recalculation

После StudentTimezoneChanged будущие Reminders должны быть пересчитаны.

Уже доставленные сообщения не изменяются.

---

## HR-017: Learning pause suppresses ordinary reminders

Во время активной Learning Pause обычные Reminders не отправляются.

Reason Code: `STUDENT_LEARNING_PAUSE_ACTIVE`

---

## HR-018: Learning pause does not automatically cancel homework

Homework и Reminder имеют разные жизненные циклы.

Политика должна определить:

- сохранить Assignment;
- перенести Due Date;
- подавить Reminders;
- запросить решение Teacher;
- пересчитать после окончания паузы.

---

## HR-019: Resume reminders require current relevance

После окончания паузы Reminder создается только если Homework все еще:

- активно;
- педагогически актуально;
- понятно;
- выполнимо;
- связано с текущей программой.

---

## HR-020: Student preferences must be respected

Student может настроить:

- разрешенные каналы;
- количество Reminders;
- preferred hours;
- отключение отдельных типов;
- объединение уведомлений;
- отсутствие автоматических Reminders.

Reason Code: `HOMEWORK_REMINDERS_DISABLED`

---

## HR-021: Mandatory policy cannot silently override preferences

Если школа считает определенное уведомление обязательным, оно должно быть отдельно классифицировано и объяснено.

Обычное Homework Reminder не должно обходить пользовательские настройки.

---

## HR-022: Teacher-defined schedule must respect communication constraints

Teacher может задать желаемое время, но система все равно применяет:

- Quiet Hours;
- Timezone;
- channel availability;
- максимальную частоту;
- consent и preferences;
- текущий Homework Status.

---

## HR-023: Maximum reminder count must be bounded

Каждый Reminder Plan должен иметь максимальное число автоматических Student-facing Reminders.

Reason Code: `HOMEWORK_REMINDER_LIMIT_REQUIRED`

---

## HR-024: Repeated reminders must not become harassment

Запрещено:

- отправлять сообщения каждый час;
- бесконечно повторять одно и то же;
- увеличивать давление при отсутствии реакции;
- использовать угрожающие формулировки;
- отправлять сообщения по всем каналам одновременно без необходимости.

Reason Code: `EXCESSIVE_REMINDER_FREQUENCY_NOT_ALLOWED`

---

## HR-025: Lack of interaction does not justify escalating pressure

Отсутствие открытия Notification не должно автоматически увеличивать частоту.

---

## HR-026: Delivery confirmation is not reading confirmation

Технически доставленный Reminder не означает, что Student его прочитал или понял.

---

## HR-027: Read receipts must not affect educational assessment

Даже если канал поддерживает Read Receipt, оно не является Progress Evidence.

---

## HR-028: Reminder must contain a useful next action

Сообщение должно помогать Student:

- открыть Homework;
- начать с первого шага;
- посмотреть инструкцию;
- задать вопрос;
- сообщить о блокере;
- перенести срок, если разрешено;
- перейти к Submission.

Недопустимо сообщение без следующего действия:

> Не забудьте домашнее задание.

---

## HR-029: Reminder should preserve teacher intent

Текст не должен искажать:

- цель Homework;
- обязательность;
- срок;
- критерии;
- ожидаемый объем;
- инструкции Teacher.

---

## HR-030: Reminder cannot invent missing instructions

Если Homework неполно или неоднозначно, система не должна дополнять его самостоятельно.

Decision: Clarification Required

Reason Code: `HOMEWORK_INSTRUCTIONS_INCOMPLETE`

---

## HR-031: Missing material changes reminder type

Если отсутствует файл, ссылка, запись или другой обязательный ресурс, обычное напоминание о выполнении подавляется.

Вместо него:

- Student получает предложение сообщить о проблеме;
- Teacher получает запрос исправить Assignment.

Reason Code: `HOMEWORK_REQUIRED_MATERIAL_MISSING`

---

## HR-032: Student-reported blocker suppresses repetitive reminders

Если Student сообщил о блокере, одинаковые Reminders не отправляются до разрешения ситуации.

Reason Code: `HOMEWORK_BLOCKER_REPORTED`

---

## HR-033: Blocker should create an actionable teacher signal

Система может создать:

`NotifyTeacherAboutHomeworkBlocker`

с минимально необходимой информацией.

---

## HR-034: Reminder tone must be supportive

Допустимо:

> До следующего занятия два дня. Можно начать с первого куплета и отметить места, где нужна помощь.

Недопустимо:

> Вы снова не выполнили задание.

---

## HR-035: Reminder must not assign blame

Сообщение не должно предполагать, что Student:

- ленится;
- игнорирует Teacher;
- плохо учится;
- нарушает обязательства;
- подводит группу.

Reason Code: `REMINDER_TONE_NOT_APPROPRIATE`

---

## HR-036: Reminder must not use artificial urgency

Нельзя использовать:

- ложный countdown;
- угрозу потери Progress;
- угрозу потери Achievement;
- выдуманный штраф;
- ложное утверждение о последней возможности.

---

## HR-037: Homework importance must come from the assignment

Система не должна самостоятельно повышать приоритет Homework ради вовлеченности.

---

## HR-038: Reminder does not change homework status

Отправка, доставка или чтение Reminder не переводят Homework в In Progress.

---

## HR-039: Reminder cannot mark homework incomplete

Только отсутствие подтвержденного Completion означает, что система не знает о завершении.

Она не должна утверждать, что Homework точно не выполнено.

Корректная формулировка:

> В системе пока нет отметки о выполнении.

---

## HR-040: Offline completion must be possible

Если Student мог выполнить Homework вне приложения, Reminder должен позволять:

- отметить выполнение;
- отправить результат;
- сообщить Teacher;
- скрыть дальнейшие напоминания до Review.

---

## HR-041: Submission immediately cancels pending reminders

После HomeworkSubmitted все обычные pending Reminders отменяются транзакционно или через надежную event-driven реакцию.

---

## HR-042: Correction requested creates a new reminder context

Если Teacher просит исправление, новый Reminder должен ссылаться на:

- Review;
- конкретные исправления;
- новый срок;
- новую версию Assignment state.

Старый Reminder Plan не переиспользуется бесследно.

---

## HR-043: Due-date changes trigger full recalculation

После изменения Due Date:

- старые Scheduled Reminders отменяются;
- создается новый план;
- предотвращается дублирование;
- сохраняется история изменения.

Reason Code: `HOMEWORK_DUE_DATE_CHANGED`

---

## HR-044: Lesson rescheduling may change lesson-based reminders

Если Homework привязано к следующему Lesson, его перенос запускает перерасчет.

---

## HR-045: Lesson cancellation requires pedagogical relevance review

Если связанный Lesson отменен, система не должна автоматически считать Homework отмененным.

Возможные решения:

- сохранить Homework;
- перенести срок;
- запросить Teacher;
- отменить Lesson-based Reminder;
- заменить его Due-based Reminder.

---

## HR-046: Reminder should not arrive after the related lesson starts

Pre-Lesson Reminder, не отправленный вовремя, обычно получает Expired, а не доставляется с опозданием.

Reason Code: `REMINDER_WINDOW_EXPIRED`

---

## HR-047: Reminder schedule must account for task size when known

Большое Homework не следует впервые напоминать непосредственно перед сроком.

Но система не должна оценивать объем без данных.

---

## HR-048: Estimated effort is guidance, not surveillance

Если Homework содержит Estimated Effort, Reminder может использовать его:

> Для задания обычно достаточно около 15 минут.

Но не должен:

- контролировать фактическое время;
- требовать доказательства продолжительности;
- оценивать Student по скорости.

---

## HR-049: Adaptive scheduling must remain explainable

Если система выбрала время автоматически, должно быть возможно объяснить:

- почему Reminder создан;
- почему выбран этот момент;
- какие ограничения применены;
- какие предпочтения учтены.

---

## HR-050: AI may propose but not autonomously pressure

AI может предложить:

- более понятную формулировку;
- подходящий первый шаг;
- объединение Reminders;
- обнаружение неполного Assignment.

AI не может:

- увеличивать частоту;
- менять обязательность;
- придумывать сроки;
- отправлять сообщение без Policy decision;
- оценивать мотивацию Student;
- персонализировать давление на основе поведения.

Reason Code: `AI_CANNOT_OVERRIDE_REMINDER_POLICY`

---

## HR-051: Sensitive homework content must not appear in lock-screen text

Push preview должен использовать минимальный безопасный текст.

Например:

> У вас есть активное задание от Belcanto.

Вместо раскрытия закрытых педагогических деталей.

---

## HR-052: Reminder visibility follows homework visibility

Student-facing Reminder не должен включать Teacher-only notes.

---

## HR-053: Guardian notifications require separate permission

Homework Reminder не отправляется Guardian автоматически только потому, что Student несовершеннолетний.

Требуются:

- утвержденная модель;
- соответствующие права;
- настройки;
- подходящий тип Homework;
- privacy review.

Reason Code: `GUARDIAN_REMINDER_PERMISSION_REQUIRED`

---

## HR-054: Group homework must still produce individual reminder decisions

Общее Assignment для группы не означает одинаковую отправку всем.

Учитываются индивидуально:

- Completion;
- preferences;
- timezone;
- pause;
- blockers;
- channels.

---

## HR-055: Peer visibility is prohibited

Student не должен видеть:

- кто еще не выполнил Homework;
- кто получил Reminder;
- кто открыл сообщение;
- кто попросил перенос.

---

## HR-056: Reminder failure does not change educational state

Delivery Failed не влияет на:

- Homework Status;
- Progress;
- Teacher Assessment;
- Achievement;
- Goal.

---

## HR-057: Retry must be bounded

Повторные технические попытки должны иметь предел и не создавать дублирующие пользовательские сообщения.

---

## HR-058: Channel fallback must be authorized

Если push недоступен, переход к email или другому каналу разрешен только настройками и Notification Policy.

---

## HR-059: Duplicate reminder delivery must be prevented

Используется Idempotency Key, включающий как минимум:

```text
HomeworkAssignmentId
HomeworkVersion
ReminderType
ScheduledWindow
StudentId
```

Reason Code: `HOMEWORK_REMINDER_ALREADY_PROCESSED`

---

## HR-060: Concurrent homework changes require re-evaluation

Если одновременно произошли Submission и Due Date Change, отправка должна проверять актуальное состояние, а не доверять старому плану.

Reason Code: `HOMEWORK_REMINDER_VERSION_CONFLICT`

---

## HR-061: Backend is authoritative

Client может отображать локальное напоминание, но authoritative Reminder Plan и доменное решение хранятся на backend.

---

## HR-062: Local device reminders must synchronize safely

Если используются локальные notifications:

- изменение Homework отменяет старое расписание;
- Submission отменяет локальные Reminders;
- смена устройства не создает дубликаты;
- сервер хранит ожидаемое состояние;
- отсутствие синхронизации явно обрабатывается.

---

## HR-063: Reminder history must be auditable

Сохраняется:

- почему создан;
- когда запланирован;
- почему перенесен;
- был ли отправлен;
- почему подавлен;
- какая версия Homework использовалась.

---

## HR-064: Student can suppress an individual reminder plan

Student может отключить дальнейшие Reminders для конкретного Assignment, если политика продукта это разрешает.

Это не отменяет само Homework.

---

## HR-065: Teacher can suppress reminders for pedagogical reasons

Teacher может отключить Reminders, например если:

- Homework предназначено для свободной практики;
- задание чувствительное;
- Student должен работать без deadline pressure;
- дальнейшая коммуникация будет личной.

Решение аудируется.

---

## HR-066: Teacher cannot force excessive communication

Teacher-defined Reminder Plan не может превышать platform limits и privacy constraints.

---

## HR-067: Reminder escalation must not be punitive

Допустимая escalation:

```text
Student Reminder
        |
        v
Clarification Offer
        |
        v
Teacher Attention Required
```

Недопустимая escalation:

```text
Student Reminder
        |
        v
Repeated Pressure
        |
        v
Public Exposure
        |
        v
Penalty
```

---

## HR-068: Teacher attention requires educational relevance

Teacher не следует уведомлять о каждом непрочитанном Reminder.

Основания могут включать:

- Student сообщил blocker;
- обязательное Homework просрочено;
- необходимый материал отсутствует;
- срок требует педагогического решения;
- Student явно запросил помощь.

---

## HR-069: Reminder timing must not be optimized solely for engagement

Метрики открытия и кликов могут использоваться для качества доставки, но не должны становиться единственной целью стратегии.

---

## HR-070: Notification fatigue must be considered across the product

Homework Reminder Policy должна учитывать другие planned notifications.

При необходимости сообщения объединяются через Notification Policy.

Reason Code: `NOTIFICATION_FREQUENCY_LIMIT_REACHED`

---

## HR-071: Multiple homework reminders may be bundled

Если у Student несколько Homework, можно отправить одно summary-сообщение, если:

- задания не теряют ясность;
- сроки понятны;
- privacy сохраняется;
- у каждого действия есть ссылка;
- это разрешает Notification Policy.

---

## HR-072: Bundling must not hide urgent distinctions

Задания с разными сроками и приоритетами должны оставаться различимыми.

---

## HR-073: Homework reminder expiration is separate from homework expiration

Reminder может истечь, хотя Homework остается активным.

---

## HR-074: Missed reminder must not always be sent late

Если Reminder не был отправлен вовремя из-за сбоя, политика должна решить:

- отправить с задержкой;
- перенести;
- объединить;
- подавить;
- создать Teacher Attention.

Решение зависит от актуальной полезности.

---

## HR-075: Every suppression must have a reason

Reason Code обязателен для Suppressed.

---

# Default Reminder Guidance

Значения ниже являются guidance, а не окончательными неизменяемыми правилами.

## Homework due within 24 hours

Обычно:

- Assignment notification;
- не более одного дополнительного Reminder;
- Reminder только если остается meaningful action time.

## Homework due within 2–4 days

Обычно:

- Start Reminder;
- Due Soon Reminder;
- максимум два автоматических Reminders.

## Homework due within 5–10 days

Обычно:

- Start Reminder;
- Midpoint Reminder при необходимости;
- Due Soon Reminder;
- максимум три автоматических Reminders.

## Homework linked to next lesson without due date

Обычно:

- один Reminder вечером накануне;
- либо в предпочтительное Student time;
- без напоминания после начала Lesson.

## Optional practice

Обычно:

- мягкий Reminder;
- низкая частота;
- возможность полностью отключить;
- отсутствие формулировок об обязательности.

---

# Reminder Windows

Рекомендуемая модель:

```text
ReminderWindow
├── EarliestLocalTime
├── LatestLocalTime
├── AllowedWeekdays
├── MinimumSpacing
├── MaximumPerDay
├── MaximumPerAssignment
└── TimezoneSource
```

---

# Minimum Spacing

Между Student-facing Reminders должен существовать минимальный интервал.

Он применяется:

- внутри одного Homework;
- между разными Homework;
- совместно с Notification Policy.

---

# Message Composition

Reminder должен состоять из:

```text
Context
+
Current State
+
Useful Next Action
+
Optional Deadline
+
Support Path
```

Пример:

> К следующему занятию осталось подготовить первый куплет. Можно открыть текст, пройти мелодию один раз и отметить места, где нужна помощь.

---

# Message Requirements

Student-facing Reminder должен:

- быть кратким;
- использовать понятное название Homework;
- не раскрывать private notes;
- показывать срок, если он существует;
- содержать прямой переход к Assignment;
- позволять сообщить о проблеме;
- не обещать последствия, которых нет;
- соответствовать языку интерфейса;
- учитывать текущий Status.

---

# Message Template Model

```text
HomeworkReminderTemplate
├── TemplateId
├── Version
├── ReminderType
├── Locale
├── Channel
├── TitleTemplate
├── BodyTemplate
├── ActionDefinitions
├── PrivacyLevel
├── ApprovedBy
└── Status
```

---

# Example Message Templates

## Start Reminder

> Домашнее задание уже доступно. Можно начать с первого небольшого шага: открыть материал и пройти основную часть один раз.

## Due Soon

> До срока задания осталось два дня. В системе пока нет отметки о выполнении. Можно продолжить работу или задать вопрос преподавателю.

## Pre-Lesson

> Следующее занятие завтра. Можно проверить домашнее задание и отметить места, которые стоит разобрать вместе с преподавателем.

## Clarification

> В задании может не хватать пояснения или материала. Отправьте вопрос преподавателю, чтобы не тратить время на догадки.

## Blocker Acknowledgement

> Проблема с заданием зафиксирована. Повторные напоминания приостановлены, пока преподаватель не уточнит следующий шаг.

## Optional Practice

> Когда будет удобно, можно повторить упражнение из последнего занятия. Это дополнительная практика без обязательного срока.

---

# Reminder Evaluation Flow

```text
Homework event received
        |
        v
Load current Homework Assignment
        |
        v
Validate status and version
        |
        +--> Cancel existing reminders
        |
        v
Load Student preferences and timezone
        |
        v
Load due date and related lesson
        |
        v
Determine reminder strategy
        |
        v
Check pause, blockers and materials
        |
        +--> Suppress
        |
        +--> Clarification Required
        |
        +--> Teacher Attention Required
        |
        v
Calculate reminder windows
        |
        v
Apply quiet hours and frequency limits
        |
        v
Create Reminder Plan
        |
        v
Schedule reminders
```

Перед отправкой:

```text
Reminder becomes due
        |
        v
Reload Homework Assignment
        |
        v
Verify version and state
        |
        +--> Cancel
        |
        +--> Suppress
        |
        +--> Reschedule
        |
        v
Apply Notification Policy
        |
        +--> Bundle
        |
        +--> Channel fallback
        |
        +--> Frequency suppression
        |
        v
Send reminder
        |
        v
Record delivery result
```

---

# Commands Produced

## ScheduleHomeworkReminder

Создает конкретный scheduled Reminder.

## RescheduleHomeworkReminder

Переносит Reminder с сохранением причины.

## CancelHomeworkReminder

Отменяет Reminder, который больше не актуален.

## SendHomeworkReminder

Запрашивает доставку через Notification Policy.

## SuppressHomeworkReminder

Фиксирует доменное решение не отправлять Reminder.

## RequestHomeworkClarification

Создает путь Student → Teacher для уточнения Assignment.

## NotifyTeacherAboutHomeworkBlocker

Информирует Teacher о проблеме, требующей решения.

## RecalculateHomeworkReminderPlan

Перестраивает план после изменения Homework, Lesson, Student preferences или Timezone.

## MarkHomeworkReminderDeliveryFailed

Фиксирует техническую ошибку доставки.

## MarkHomeworkReminderDelivered

Фиксирует доступное подтверждение доставки.

---

# Domain Events

```text
HomeworkReminderPlanCreated
HomeworkReminderScheduled
HomeworkReminderRescheduled
HomeworkReminderDue
HomeworkReminderSendingRequested
HomeworkReminderDelivered
HomeworkReminderDeliveryFailed
HomeworkReminderSuppressed
HomeworkReminderCancelled
HomeworkReminderExpired
HomeworkClarificationRequested
HomeworkBlockerReported
TeacherHomeworkAttentionRequested
HomeworkReminderPlanRecalculationRequested
HomeworkReminderPlanRecalculated
```

## HomeworkReminderScheduled Event

Событие должно содержать:

- ReminderId;
- ReminderPlanId;
- HomeworkAssignmentId;
- HomeworkVersion;
- StudentId;
- ReminderType;
- ScheduledFor;
- Timezone;
- Channel;
- TemplateId;
- PolicyId;
- PolicyVersion;
- CorrelationId;
- CausationId.

Событие не должно содержать закрытый текст Teacher notes.

## HomeworkReminderSendingRequested Event

Должно содержать:

- ReminderId;
- HomeworkAssignmentId;
- StudentId;
- ReminderType;
- Channel preference;
- Message template reference;
- safe rendering parameters;
- IdempotencyKey;
- PolicyId;
- PolicyVersion;
- RequestedAt.

---

# Suppression Reasons

Допустимые Reason Codes включают:

```text
HOMEWORK_ALREADY_COMPLETED
HOMEWORK_CANCELLED
HOMEWORK_EXPIRED
HOMEWORK_REPLACED
REMINDER_NO_LONGER_CURRENT
STUDENT_QUIET_HOURS_ACTIVE
STUDENT_LEARNING_PAUSE_ACTIVE
HOMEWORK_REMINDERS_DISABLED
HOMEWORK_BLOCKER_REPORTED
HOMEWORK_REQUIRED_MATERIAL_MISSING
NOTIFICATION_FREQUENCY_LIMIT_REACHED
REMINDER_TOO_CLOSE_TO_ASSIGNMENT
REMINDER_WINDOW_EXPIRED
DUPLICATE_REMINDER
CHANNEL_NOT_ALLOWED
HOMEWORK_VERSION_CHANGED
```

---

# Human Review

Human Review может потребоваться, если:

- Homework не имеет понятного срока;
- Assignment неполно;
- Teacher и Student по-разному понимают обязательность;
- Student сообщил повторяющийся blocker;
- Homework просрочено и дальнейшее действие неясно;
- связанный Lesson отменен;
- Learning Pause заканчивается после Due Date;
- требуется уведомление Guardian;
- автоматическая стратегия создает слишком много сообщений;
- AI предложил существенное изменение текста или расписания;
- Reminder должен быть отправлен после необычно долгой паузы.

---

# Human Review Result

Teacher или уполномоченный Actor может:

- уточнить Homework;
- добавить материал;
- изменить Due Date;
- отменить Assignment;
- отключить Reminders;
- создать индивидуальное расписание;
- разрешить позднюю Submission;
- заменить Homework;
- подтвердить Optional status;
- сообщить Student напрямую;
- закрыть Blocker;
- перенести обсуждение на Lesson.

---

# Notification Policy Integration

Homework Reminder Policy отвечает:

> Должно ли существовать Reminder и является ли он сейчас актуальным?

Notification Policy отвечает:

> Как, когда и через какой канал доставить разрешенное сообщение?

Notification Policy может:

- объединить несколько Reminders;
- выбрать канал;
- применить product-wide frequency limits;
- локализовать шаблон;
- отложить доставку;
- использовать fallback;
- управлять retry;
- подавить lock-screen details.

Notification Policy не должна изменять педагогический смысл Reminder.

---

# Homework Expiration Policy Integration

До истечения срока Homework Reminder Policy может создавать обычные Reminders.

После истечения:

```text
Homework Due Date Passed
        |
        v
Homework Expiration Policy
        |
        +--> Keep Active
        +--> Mark Overdue
        +--> Expire
        +--> Extend
        +--> Teacher Review
```

Только после решения Homework Expiration Policy определяется, допустима ли дальнейшая коммуникация.

Homework Reminder Policy не должна самостоятельно решать, что просроченное Homework потеряло смысл.

---

# Lesson Integration

Homework может быть связано с:

- предыдущим Lesson;
- следующим Lesson;
- конкретным Learning Goal;
- отдельной практикой;
- Concert preparation;
- Song Readiness.

Изменение Lesson schedule может влиять на Reminder timing, но не обязательно меняет содержание Assignment.

---

# Progress Integration

Reminder events не являются Progress Evidence.

Допустимо использовать:

- reviewed Homework;
- Teacher Assessment;
- подтвержденные результаты выполнения.

Недопустимо использовать:

- число доставленных Reminders;
- число открытий;
- скорость реакции;
- количество нажатий;
- отключение Notifications.

---

# Goal Integration

Homework может поддерживать Goal.

Reminder может объяснить связь:

> Это задание помогает закрепить переход, над которым вы работаете в текущей цели.

Но Reminder не должен утверждать, что невыполнение одного задания отменяет Goal или Progress.

---

# Student Presentation

Student должен видеть:

- активное Homework;
- Due Date;
- Reminder Plan;
- следующий Reminder, если это уместно;
- возможность изменить время;
- возможность отключить Reminders;
- возможность сообщить о Blocker;
- возможность запросить Clarification;
- историю отправленных сообщений в разумном объеме;
- текущий статус Submission.

Student не должен видеть внутренние delivery retries или технические Reason Codes.

---

# Teacher Presentation

Teacher должен видеть:

- Homework Assignment;
- Reminder Strategy;
- количество запланированных Reminders;
- Student preferences в минимально необходимом объеме;
- reported blockers;
- Clarification requests;
- Delivery failure, только если это требует действия;
- Suppression reasons;
- текущий Homework Status;
- возможность изменить Assignment или Reminder Plan.

Teacher не должен использовать Reminder analytics как прямую оценку мотивации.

---

# Administrator Presentation

Administrator может видеть:

- delivery state;
- channel failures;
- Reminder schedule;
- системные suppression reasons;
- массовые сбои;
- технические retries.

Administrator не должен видеть закрытые педагогические данные без необходимости.

---

# Owner Analytics

Owner может видеть агрегированные данные:

- долю Homework с Reminder Plans;
- количество Scheduled / Delivered / Suppressed;
- частоту Delivery Failed;
- число Clarification requests;
- частые причины Blocker;
- долю отключенных Reminders;
- количество дубликатов;
- уведомительную нагрузку;
- эффективность bundling;
- соблюдение Quiet Hours.

Owner Analytics не должна использоваться для:

- рейтинга Student;
- оценки дисциплины;
- начисления наказаний;
- автоматического давления;
- оценки Teacher только по Submission rate.

---

# AI Assistance

AI может:

- предложить понятный Reminder text;
- сократить длинную инструкцию;
- выделить первый шаг;
- обнаружить отсутствие Due Date;
- обнаружить противоречие;
- предложить объединение сообщений;
- классифицировать reported blocker;
- предложить Teacher clarification;
- проверить tone и privacy;
- предложить Reminder Strategy как Draft.

AI не может:

- придумывать содержание Homework;
- менять Due Date;
- увеличивать обязательность;
- определять мотивацию;
- выбирать манипулятивное время;
- обходить Quiet Hours;
- игнорировать preferences;
- отправлять сообщение без Policy decision;
- помечать Student как неответственного;
- передавать private Teacher notes в текст;
- использовать эмоциональную уязвимость для повышения реакции.

AI metadata должна содержать:

- model or mechanism;
- version;
- input references;
- proposed text or strategy;
- confidence;
- timestamp;
- human confirmation status, если требуется.

---

# Privacy

Homework Reminder может раскрывать:

- факт обучения;
- конкретную Song;
- педагогическую трудность;
- расписание;
- содержание Homework;
- активность Student.

Необходимо:

- минимизировать lock-screen preview;
- соблюдать channel consent;
- не показывать private notes;
- не отправлять данные чужому contact;
- защищать несовершеннолетних;
- отделять Guardian access;
- хранить только необходимые delivery metadata;
- ограничивать retention;
- аудитировать изменение contact destination.

---

# Security

Необходимо защищать:

- отправку Reminder чужому Student;
- подмену HomeworkAssignmentId;
- использование устаревшей Homework Version;
- повторную массовую доставку;
- обход Notification Preferences;
- изменение Timezone для манипуляции;
- подделку delivery confirmation;
- неавторизованное включение Guardian;
- раскрытие private content;
- произвольное изменение Reminder Plan.

---

# Audit Requirements

Для создания Reminder Plan сохраняются:

- PolicyId;
- PolicyVersion;
- HomeworkAssignmentId;
- HomeworkVersion;
- StudentId;
- TeacherId;
- Strategy;
- Due Date;
- related Lesson reference;
- Timezone;
- Quiet Hours;
- Preferred Channels;
- Maximum Reminder Count;
- ActorId;
- CreatedAt;
- CorrelationId;
- CausationId.

Для каждого Reminder сохраняются:

- ReminderId;
- ReminderType;
- ScheduledFor;
- original scheduled time;
- adjusted time;
- adjustment reason;
- Channel;
- TemplateId;
- Status;
- Homework Version;
- IdempotencyKey;
- delivery attempt references;
- DeliveredAt;
- SuppressedAt;
- Suppression Reason;
- CancelledAt;
- Cancel Reason.

Для recalculation:

- previous plan version;
- new plan version;
- triggering event;
- changed inputs;
- cancelled Reminders;
- new Reminders;
- Actor;
- timestamp.

---

# Failure Modes

## Homework not found

- Decision: Rejected
- Reason Code: HOMEWORK_NOT_FOUND

## Homework assignment inactive

- Decision: No Reminder Required
- Reason Code: ACTIVE_HOMEWORK_REQUIRED

## Homework version missing

- Decision: Rejected
- Reason Code: HOMEWORK_VERSION_REQUIRED

## Student mismatch

- Decision: Rejected
- Reason Code: HOMEWORK_REMINDER_STUDENT_MISMATCH

Security Audit обязателен.

## Missing timezone

- Decision: Deferred (или используется документированное school default timezone)
- Reason Code: STUDENT_TIMEZONE_REQUIRED

## Reminder falls in quiet hours

- Decision: Reminder Rescheduled or Reminder Suppressed
- Reason Code: STUDENT_QUIET_HOURS_ACTIVE

## Homework already submitted

- Decision: Reminder Cancelled
- Reason Code: HOMEWORK_ALREADY_COMPLETED

## Due date changed

- Decision: Reminder Plan Recalculated
- Reason Code: HOMEWORK_DUE_DATE_CHANGED

## Required material missing

- Decision: Clarification Required
- Commands: SuppressHomeworkReminder, NotifyTeacherAboutHomeworkBlocker
- Reason Code: HOMEWORK_REQUIRED_MATERIAL_MISSING

## Student reported blocker

- Decision: Teacher Attention Required
- Reason Code: HOMEWORK_BLOCKER_REPORTED

## Student disabled reminders

- Decision: Reminder Suppressed
- Reason Code: HOMEWORK_REMINDERS_DISABLED

## Frequency limit reached

- Decision: Reminder Suppressed (или bundling)
- Reason Code: NOTIFICATION_FREQUENCY_LIMIT_REACHED

## Delivery failure

- Decision: Delivery Failed
- Reason Code: HOMEWORK_REMINDER_DELIVERY_FAILED

Дальнейшее действие определяется Retry и Notification Policy.

## Duplicate trigger

- Decision: No Reminder Required
- Reason Code: HOMEWORK_REMINDER_ALREADY_PROCESSED

## Concurrent update

- Decision: Deferred
- Reason Code: HOMEWORK_REMINDER_VERSION_CONFLICT

Политика повторно оценивается на актуальном Homework state.

---

# Explainability Examples

## Reminder Scheduled

> Напоминание запланировано на вечер за два дня до следующего занятия. Время выбрано с учетом вашего часового пояса и настроек уведомлений.

## Quiet Hours

> Напоминание перенесено на утро, потому что первоначальное время попадало в период без уведомлений.

## Homework Submitted

> Напоминание отменено, потому что задание уже отправлено преподавателю.

## Blocker

> Повторные напоминания приостановлены, потому что вы сообщили о проблеме с материалом. Преподавателю отправлен запрос на уточнение.

## Due Date Changed

> Преподаватель изменил срок задания, поэтому старые напоминания отменены и создан новый план.

## No Time for Meaningful Action

> До занятия осталось слишком мало времени для полноценного выполнения. Вместо обычного напоминания можно сообщить преподавателю, какая часть требует дополнительного времени.

---

# Examples

## Example 1: Standard due-based homework

Дано:

- Homework назначено в понедельник;
- Due Date — пятница;
- Student разрешил два Reminder;
- Quiet Hours: 22:00–08:00.

Результат:

- Decision: Reminder Plan Created
- Reminders:
  - Wednesday 19:00 — Start Reminder
  - Friday 18:00 — Due Soon Reminder

## Example 2: Homework completed before reminder

Дано:

- Reminder назначен на 19:00;
- Student отправил Homework в 17:30.

Результат:

- Decision: Reminder Cancelled
- Reason Code: HOMEWORK_ALREADY_COMPLETED

## Example 3: Reminder falls at night

Дано:

- Teacher выбрал 23:30;
- Quiet Hours начинаются в 22:00;
- Due Date позволяет перенос.

Результат:

- Decision: Reminder Rescheduled
- New Time: 08:30 next day
- Reason Code: STUDENT_QUIET_HOURS_ACTIVE

## Example 4: Student reports missing file

Дано:

- Homework требует аудиозапись;
- файл недоступен;
- Student сообщил Blocker.

Результат:

- Decision: Teacher Attention Required
- Commands:
  - SuppressHomeworkReminder
  - NotifyTeacherAboutHomeworkBlocker
- Reason Code: HOMEWORK_REQUIRED_MATERIAL_MISSING

## Example 5: Optional practice

Дано:

- Homework помечено Optional;
- Due Date отсутствует;
- Student разрешил мягкие reminders.

Результат:

- Decision: Reminder Scheduled
- Reminder Type: Optional Practice
- Maximum Count: 1

Текст не должен создавать обязательность.

## Example 6: Lesson moved

Дано:

- Homework связано со следующим Lesson;
- Lesson перенесен на три дня;
- стратегия Lesson-Based.

Результат:

- Decision: Reminder Plan Recalculated
- Reason Code: RELATED_LESSON_RESCHEDULED

## Example 7: Student is on learning pause

Дано:

- Learning Pause активна;
- Reminder должен отправиться завтра.

Результат:

- Decision: Reminder Suppressed
- Reason Code: STUDENT_LEARNING_PAUSE_ACTIVE

После завершения паузы требуется проверка актуальности Homework.

## Example 8: Homework correction requested

Дано:

- Student отправил Homework;
- Teacher запросил исправление;
- назначен новый срок.

Результат:

- Decision: New Reminder Plan Created
- Context: Correction Requested

Старые Reminders остаются отмененными.

## Example 9: Multiple homework assignments

Дано:

- у Student три активных Homework;
- два Reminder приходятся на один вечер;
- bundling разрешен.

Результат:

- Decision: Bundle Through Notification Policy

Student получает одно summary-сообщение.

## Example 10: Delivery failure

Дано:

- push token недействителен;
- email fallback не разрешен.

Результат:

- Decision: Delivery Failed
- Reason Code: HOMEWORK_REMINDER_DELIVERY_FAILED

Homework Status не меняется.

## Example 11: Reminder due after lesson began

Дано:

- Pre-Lesson Reminder задержался из-за сбоя;
- Lesson уже начался.

Результат:

- Decision: Reminder Expired
- Reason Code: REMINDER_WINDOW_EXPIRED

## Example 12: AI proposes more urgent wording

Дано:

- AI предлагает: «Последний шанс выполнить задание»;
- Homework можно отправить позже;
- Teacher не устанавливал такую формулировку.

Результат:

- Decision: Proposal Rejected
- Reason Code: ARTIFICIAL_URGENCY_NOT_ALLOWED

---

# Test Requirements

## Plan Creation Tests

- active Homework creates Reminder Plan;
- completed Homework does not create Plan;
- cancelled Homework does not create Plan;
- Due-Based Strategy requires Due Date;
- Lesson-Based Strategy can use related Lesson;
- maximum Reminder count is enforced;
- plan stores Homework Version.

## Scheduling Tests

- Reminder is scheduled in Student timezone;
- Quiet Hours shift Reminder;
- minimum spacing is enforced;
- excessive frequency is rejected;
- short assignments do not receive unnecessary reminders;
- long assignments can receive staged reminders;
- pre-lesson Reminder expires after Lesson begins.

## State Change Tests

- Submission cancels pending Reminders;
- Review cancels old Reminders;
- Correction Request creates new context;
- Due Date Change recalculates Plan;
- Homework Update invalidates old version;
- Homework Cancellation cancels all Reminders;
- Homework Expiration stops ordinary reminders.

## Lesson Integration Tests

- Lesson reschedule recalculates lesson-based Reminder;
- Lesson cancellation triggers review;
- unrelated Lesson change does not affect Plan;
- new linked Lesson updates schedule;
- Reminder is not sent after related Lesson starts.

## Preference Tests

- disabled reminders are suppressed;
- preferred channel is respected;
- forbidden fallback is not used;
- Student-defined time is applied;
- Quiet Hours override Teacher-defined time;
- individual Assignment suppression works.

## Pause Tests

- active Learning Pause suppresses Reminder;
- pause ending triggers relevance check;
- expired Homework is not automatically resumed;
- moved Due Date can restore Plan;
- pause history remains auditable.

## Blocker Tests

- missing material suppresses ordinary Reminder;
- Student blocker creates Teacher Attention;
- duplicate blocker does not spam Teacher;
- resolved blocker permits recalculation;
- incomplete instructions trigger Clarification.

## Delivery Tests

- successful delivery is recorded;
- failed delivery does not change Homework;
- retry is bounded;
- duplicate delivery is prevented;
- invalid channel does not silently fallback;
- late delivery may expire.

## Idempotency Tests

- duplicate HomeworkAssigned event creates one Plan;
- duplicate Reminder Due event creates one send request;
- duplicate delivery callback is harmless;
- concurrent Submission cancels stale send;
- stale Homework Version cannot send;
- duplicate recalculation does not duplicate Reminders.

## Permission Tests

- authorized Teacher can define strategy;
- Student can manage their preferences;
- Student cannot modify another Assignment;
- Administrator cannot expose private content;
- Guardian receives nothing without permission;
- AI cannot send independently.

## Privacy Tests

- lock-screen text is minimal;
- private Teacher notes are excluded;
- wrong contact is rejected;
- guardian scope is respected;
- group Homework does not reveal peer state;
- event payload excludes sensitive content.

## Tone Tests

- supportive text is accepted;
- blame language is rejected;
- false urgency is rejected;
- threats are rejected;
- ranking language is rejected;
- Optional Homework is not described as mandatory;
- Student-visible explanation contains no internal Reason Codes.

## AI Tests

- AI can propose Draft wording;
- AI cannot invent Due Date;
- AI cannot change Homework importance;
- AI cannot override Quiet Hours;
- AI cannot infer laziness;
- hallucinated material is rejected;
- AI metadata is stored where required.

## Explainability Tests

- scheduled time is explainable;
- suppression reason is explainable;
- reschedule reason is explainable;
- cancellation after Submission is explainable;
- bundling is explainable;
- Delivery Failed does not imply Student fault.

---

# Non-Goals

Homework Reminder Policy не определяет:

- содержание Homework;
- педагогическую корректность Assignment;
- Homework Completion;
- Homework Review;
- Homework Expiration outcome;
- Progress Calculation;
- Goal Completion;
- Achievement Award;
- наказания;
- оплату;
- CRM-статусы;
- посещаемость;
- расписание Lesson;
- push infrastructure;
- email transport;
- retry implementation;
- маркетинговые уведомления;
- рейтинг Student;
- оценку мотивации;
- медицинские рекомендации.

---

# Open Questions

Необходимо определить:

- какие Reminder Strategies входят в MVP;
- сколько автоматических Reminders разрешено;
- какой Minimum Spacing использовать;
- какие Quiet Hours применяются по умолчанию;
- может ли Student полностью отключить Homework Reminders;
- существуют ли обязательные уведомления;
- нужен ли отдельный Assignment Created message;
- должен ли Teacher видеть, что Reminder доставлен;
- нужны ли Read Receipts;
- какие каналы поддерживаются;
- разрешен ли SMS;
- когда применять email fallback;
- будут ли local device notifications;
- как синхронизировать локальные и серверные Reminders;
- как обрабатывать несколько устройств;
- как учитывать offline Submission;
- как быстро Submission отменяет scheduled push;
- нужен ли adaptive schedule в MVP;
- какие данные adaptive strategy может использовать;
- как измерять качество Reminder без оптимизации на engagement;
- можно ли объединять Homework разных Teachers;
- кто утверждает Message Templates;
- как локализовать тексты;
- можно ли Teacher редактировать Reminder text;
- нужно ли проверять Teacher text на tone и privacy;
- как обрабатывать Optional Homework;
- как напоминать о Correction Request;
- что делать с Homework без Due Date;
- что делать при отмене Lesson;
- что делать при переносе Lesson на длительный срок;
- как взаимодействовать с Learning Pause;
- кто может создавать Learning Pause;
- нужна ли отдельная Notification Preference aggregate;
- нужна ли отдельная Reminder Schedule aggregate;
- как моделировать Guardian notifications;
- какие данные доступны Guardian;
- как хранить notification consent;
- сколько хранить delivery history;
- нужно ли сохранять rendered message;
- как безопасно хранить message parameters;
- как отменять большое количество Reminders;
- как обрабатывать timezone daylight-saving changes;
- какой timezone использовать при неизвестном Student timezone;
- как обрабатывать путешествие Student;
- следует ли напоминать в выходные;
- может ли Student задавать allowed weekdays;
- как учитывать школьные праздники;
- нужны ли teacher attention summaries;
- как избежать уведомления Teacher о каждом Overdue Homework;
- какая политика определяет Overdue;
- когда Reminder Plan становится Archived;
- как обрабатывать удаление Student account;
- как обрабатывать transfer к другому Teacher;
- может ли новый Teacher видеть старые Reminder history;
- нужно ли отделять образовательные и технические delivery events;
- нужна ли отдельная Communication Policy;
- нужна ли отдельная Notification Fatigue Policy;
- должен ли Reminder содержать estimated effort;
- как формировать первый полезный шаг;
- может ли AI предлагать этот шаг;
- когда AI proposal требует Teacher confirmation;
- как тестировать отсутствие манипулятивного текста.

---

# History

| Version | Description |
| --- | --- |
| 1.0.0 | Определены правила планирования, отправки, подавления и пересчета Homework Reminders. |
