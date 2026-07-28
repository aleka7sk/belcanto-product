# Назначение

Эта политика отвечает не за отдельное доменное решение, а за регулярную переоценку состояния системы.

Если остальные Policy реагируют на события (LessonCompleted, HomeworkSubmitted, ConcertCompleted и т.д.), то Periodic Review отвечает на вопрос:

> "Что необходимо переоценить, даже если никаких новых событий не произошло?"

Это фактически "heartbeat" всей доменной модели.

---

# Purpose

Даже если пользователь ничего не делает, время продолжает идти.

Из-за этого могут измениться:

- актуальность Homework;
- Song Readiness;
- Concert Eligibility;
- Reminder Plans;
- Goal Review;
- Learning Pause;
- Notification Schedules;
- архивирование;
- просроченные проверки;
- ожидание Teacher Review.

Поэтому система должна регулярно переоценивать состояние.

---

# Core Principle

Время само по себе является доменным фактором.

```text
Current Time
      |
      v
Periodic Review
      |
      +--> Homework
      |
      +--> Goals
      |
      +--> Songs
      |
      +--> Concerts
      |
      +--> Notifications
      |
      +--> Reviews
      |
      +--> Archive
```

---

# Что НЕ делает Policy

Она не принимает доменные решения самостоятельно.

Она только инициирует повторную оценку соответствующей Policy.

Например:

```text
Homework became old
↓
Periodic Review
↓
Request Homework Expiration Evaluation
↓
Homework Expiration Policy
↓
Expire / Keep Active / Extend
```

---

# Что проверяется

## Homework

Проверяется:

- долгое Overdue;
- окончание Grace Period;
- Homework без активности;
- устаревшие Assignment;
- просроченный Review.

## Goal

Проверяется:

- давно нет Progress;
- давно нет Review;
- Goal ожидает Teacher;
- Goal потеряла актуальность.

## Song

Проверяется:

- давно не репетировалась;
- давно не оценивалась;
- требуется новая Readiness Evaluation.

## Concert

Проверяется:

- Eligibility устарела;
- изменился состав;
- требуется повторное подтверждение;
- Slot изменился.

## Reminder

Проверяется:

- Reminder уже потерял смысл;
- Reminder нужно пересчитать;
- Reminder завис.

## Notification

Проверяется:

- Scheduled Notification уже не актуальна;
- Delivery Window закончилась;
- Retry больше нельзя выполнять.

## Teacher Review

Проверяется:

- слишком долго Pending;
- нужен Escalation;
- Review забыта.

## Archive

Проверяется:

- старые Homework;
- старые Notifications;
- старые Decisions;
- Completed Aggregate;
- Expired Aggregate.

---

# Trigger

Policy запускается:

- каждые N минут;
- ежедневно;
- еженедельно;
- после startup;
- после восстановления после outage;
- вручную Administrator.

---

# Review Categories

- Homework Review
- Goal Review
- Song Review
- Concert Review
- Reminder Review
- Notification Review
- Archive Review
- Teacher Review
- Integrity Review

---

# Integrity Review

Проверяются инварианты.

Например:

- Homework Active без Student;
- Goal без Owner;
- Reminder без Homework;
- Notification без Recipient;
- Broken Reference;
- Duplicate Aggregate;
- Invalid Version.

---

# Commands

Policy генерирует только запросы.

Например:

- RequestHomeworkExpirationReview
- RequestGoalReview
- RequestSongReadinessReview
- RequestConcertEligibilityReview
- RequestReminderRecalculation
- RequestNotificationReview
- RequestArchiveEvaluation
- RequestIntegrityReview

Она не вызывает:

- ExpireHomework
- CompleteGoal
- AwardAchievement

напрямую.

---

# Decision Rules

## PR-001

Periodic Review никогда не изменяет Aggregate самостоятельно.

## PR-002

Каждый Review должен быть идемпотентным.

## PR-003

Review должен использовать текущее состояние Aggregate.

## PR-004

Review не должен использовать устаревший Snapshot.

## PR-005

Review не должен запускать одну и ту же проверку одновременно.

## PR-006

Если Aggregate уже находится в терминальном состоянии, повторный Review обычно не требуется.

## PR-007

Review не должен создавать бесконечные циклы.

Например:

```text
Periodic Review
↓
Homework Review
↓
Reminder Recalculation
↓
Notification
↓
Periodic Review
```

Такой цикл запрещен.

## PR-008

Review должен иметь собственный CorrelationId.

## PR-009

Review должен хранить время запуска.

## PR-010

Review должен хранить Policy Version.

---

# Scheduling

Рекомендуемые интервалы:

| Category | Suggested Frequency |
| --- | --- |
| Homework | Hourly |
| Reminder | Every 15–30 min |
| Notification | Every 5–15 min |
| Goal | Daily |
| Song | Daily |
| Concert | Daily |
| Archive | Daily / Weekly |
| Integrity | Daily |

Частота является конфигурацией, а не частью доменной логики.

---

# Failure Handling

Если Review завершился ошибкой:

- сохраняется причина;
- допускается Retry;
- Aggregate не считается обработанным;
- Review может быть перенесен.

---

# Retry

Retry допустим только при технических ошибках.

Не допускается бесконечный Retry.

---

# Audit

Для каждого запуска сохраняются:

- Review Id;
- Category;
- Aggregate Type;
- Aggregate Id;
- Started At;
- Finished At;
- Duration;
- Result;
- Requested Commands;
- Policy Version;
- CorrelationId.

---

# AI

AI может:

- предложить Aggregate для Review;
- обнаружить аномалию;
- предложить Priority;
- обнаружить забытые Homework.

AI не может:

- самостоятельно завершить Review;
- менять Aggregate;
- принимать доменные решения.

---

# Examples

## Homework

```text
Homework
↓
45 дней Overdue
↓
Periodic Review
↓
RequestHomeworkExpirationReview
```

## Song

```text
Song
↓
180 дней без оценки
↓
Periodic Review
↓
RequestSongReadinessReview
```

## Reminder

```text
Reminder
↓
Due прошло
↓
Periodic Review
↓
RequestReminderRecalculation
```

## Notification

```text
Notification Scheduled
↓
Delivery Window закончился
↓
Periodic Review
↓
Expire Notification
```

через Notification Policy.

## Teacher Review

```text
Teacher Review
↓
Pending 14 дней
↓
Periodic Review
↓
RequestTeacherReviewEscalation
```

---

# Test Requirements

Проверить:

- повторный запуск идемпотентен;
- Review не запускает цикл;
- Review использует актуальную версию Aggregate;
- Aggregate не изменяется напрямую;
- команды отправляются соответствующим Policy;
- Retry работает только при технической ошибке;
- CorrelationId сохраняется;
- Audit полный;
- терминальные Aggregate пропускаются;
- массовый Review масштабируется безопасно.

---

# Non-Goals

Periodic Review Policy не определяет:

- Homework Expiration;
- Goal Completion;
- Achievement Award;
- Reminder Schedule;
- Notification Delivery;
- Song Readiness;
- Concert Eligibility;
- Archive Rules;
- Retention Policy.

Она только инициирует повторную оценку соответствующих доменных политик.

---

# History

| Version | Description |
| --- | --- |
| 1.0.0 | Введена единая политика периодической переоценки состояния доменной модели. |
