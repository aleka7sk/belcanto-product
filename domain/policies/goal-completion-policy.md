---
Status: Accepted
Version: 1.0.0
Last Updated: 2026-07-27

Policy Id: GOAL_COMPLETION_POLICY

Policy Type:
  - Validation Policy
  - Reaction Policy
  - Recommendation Policy

Owners:
  - Product Owner
  - Education Lead
  - Technical Lead

Related Aggregates:
  - Goal
  - Student
  - Progress
  - Assessment
  - Lesson
  - Homework

Observed Events:
  - AssessmentPublished
  - ProgressUpdated
  - HomeworkReviewed
  - LessonCompleted
  - GoalCompletionRequested
  - TeacherProgressReviewCompleted

Produced Commands:
  - CompleteGoal
  - ReopenGoal
  - RequestTeacherGoalReview
  - CreateFollowUpGoal
  - AwardAchievement

Related Documents:
  - 000-domain-policy-overview.md
  - progress-update-policy.md
  - ../goal.md
  - ../progress.md
---

# Goal Completion Policy

> Goal Completion Policy определяет, когда образовательная цель действительно считается достигнутой.

Goal представляет собой ожидаемый образовательный результат.

Политика гарантирует, что завершение Goal основано на подтвержденных фактах, а не на субъективном ощущении.

---

# Purpose

Необходимо исключить ситуации, когда Goal считается выполненной потому что:

- прошло достаточно времени;
- проведено много занятий;
- ученик считает, что цель достигнута;
- преподаватель случайно нажал кнопку.

Goal должна завершаться только при наличии подтвержденного достижения критериев.

---

# Core Principle

Goal считается завершенной только тогда, когда существуют достаточные подтвержденные Evidence того, что критерии Goal выполнены.

Количество проведенных Lesson само по себе не является критерием достижения.

---

# Goal States

```
Draft

↓

Active

↓

Completed
```

Дополнительно:

```
Cancelled

Reopened

Archived
```

---

# Trigger

Политика применяется при:

- GoalCompletionRequested
- TeacherProgressReviewCompleted
- ProgressUpdated
- AssessmentPublished

---

# Inputs

Политика использует:

- Goal
- Goal Criteria
- Goal Progress
- Skill Progress
- Assessment
- Homework Review
- Teacher Review
- Lesson History
- Student
- Policy Version

---

# Decision Outcomes

Возможные результаты:

- Goal Completed
- Goal Not Yet Completed
- Teacher Review Required
- Insufficient Evidence
- Goal Reopened

---

# Decision Rules

## GC-001

Goal должна иметь минимум один критерий.

Пустая Goal завершена быть не может.

Reason Code:

GOAL_HAS_NO_CRITERIA

---

## GC-002

Все обязательные критерии должны быть выполнены.

Если хотя бы один обязательный критерий не подтвержден —

Goal не завершается.

Reason Code:

GOAL_CRITERIA_NOT_MET

---

## GC-003

Progress может быть Evidence достижения Goal.

Но сам по себе Progress не завершает Goal автоматически.

---

## GC-004

Teacher может подтвердить достижение Goal.

Подтверждение сохраняется в Audit.

---

## GC-005

Student Self Assessment не завершает Goal.

Она может использоваться как дополнительный источник информации.

---

## GC-006

AI не завершает Goal.

AI может лишь предложить:

- достигнута ли Goal;
- какие Evidence использованы;
- какие критерии еще отсутствуют.

---

## GC-007

После завершения Goal история не удаляется.

Даже если позже Goal будет Reopened.

---

## GC-008

Goal может быть Reopened.

Например:

- ошибочное завершение;
- изменение требований;
- отмена Assessment;
- отзыв Teacher.

История предыдущего Completion сохраняется.

---

## GC-009

Goal должна иметь автора завершения.

Сохраняются:

- Teacher
- дата
- версия Policy
- использованные Evidence

---

## GC-010

Goal Completion должна быть объяснимой.

Student должен понимать:

- почему Goal достигнута;
- какие Evidence использованы;
- что делать дальше.

---

# Commands

- CompleteGoal
- ReopenGoal
- RequestTeacherGoalReview
- CreateFollowUpGoal
- AwardAchievement

---

# Domain Events

- GoalCompleted
- GoalReopened
- GoalCompletionRejected
- GoalReviewRequested

---

# Effects

После GoalCompleted могут быть запущены:

- Achievement Award Policy
- Progress Update Policy
- Notification Policy

---

# Explainability

Вместо:

> Goal достигнута.

Система должна уметь объяснить:

> За последние пять занятий преподаватель подтвердил устойчивое выполнение всех критериев Goal. Навык успешно проявился также во время концертной репетиции.

---

# Failure Modes

## Недостаточно Evidence

Decision:

Insufficient Evidence

---

## Не выполнен критерий

Decision:

Goal Not Yet Completed

---

## Конфликтующие Assessment

Decision:

Teacher Review Required

---

## Goal уже завершена

Decision:

No Action

---

# Audit

Сохраняются:

- Policy Version
- Goal Version
- Teacher
- Decision
- Evidence
- Time
- Reason Codes

---

# AI Assistance

AI может:

- проверить соответствие критериям;
- найти отсутствующие Evidence;
- предложить Summary.

AI не может:

- завершать Goal;
- менять Goal Status;
- подтверждать педагогическое решение.

---

# Tests

Обязательные тесты:

- Goal без критериев;
- Goal с одним критерием;
- Goal с несколькими критериями;
- недостаточно Evidence;
- Teacher подтверждает Completion;
- AI не может завершить Goal;
- Reopen Goal;
- повторное Completion идемпотентно;
- конфликтующие Assessment;
- отзыв Assessment после Completion.

---

# Non-Goals

Политика не определяет:

- Progress Calculation;
- Lesson Completion;
- Homework Completion;
- расписание занятий;
- Achievement Rules.

---

# Open Questions

Необходимо определить:

- обязательны ли количественные критерии;
- допускаются ли несколько Teachers;
- могут ли существовать автоматические Goal;
- может ли Student предлагать завершение Goal;
- нужна ли периодическая проверка старых Completed Goal.

---

# History

| Version | Description |
|---------|-------------|
|1.0.0|Первое описание Goal Completion Policy.|