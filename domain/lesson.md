---
Status: Approved
Version: 1.0.0

Entity: Lesson

Owners:
  - Product Owner
  - Technical Lead

Aggregate Root: Yes

Related:
  - student.md
  - teacher.md
  - homework.md
  - assessment.md
  - skill.md
  - song.md
  - attachment.md
---

# Lesson

> Lesson — центральный агрегат образовательного процесса Belcanto Product.
>
> Практически все действия преподавателя происходят в рамках конкретного занятия.
>
> Lesson является главным источником образовательной истории ученика.

---

# Purpose

Lesson представляет собой завершенное образовательное взаимодействие между преподавателем и учеником (или группой учеников).

Lesson существует для того, чтобы:

- сохранить историю обучения;
- зафиксировать образовательный результат;
- определить дальнейшие действия;
- стать точкой отсчета для последующего развития.

Lesson — это не просто запись в расписании.

Это образовательное событие.

---

# Responsibilities

Lesson отвечает за:

- проведение занятия;
- образовательный контекст;
- комментарии преподавателя;
- результаты занятия;
- домашнее задание;
- рекомендации;
- историю изменений.

Lesson НЕ отвечает за:

- профиль ученика;
- глобальный прогресс;
- программу обучения;
- расписание школы.

---

# Lifecycle

```
Scheduled

↓

Started

↓

Completed

↓

Archived
```

Дополнительно возможны состояния

```
Cancelled

Missed

Rescheduled
```

---

# Scheduled

Занятие запланировано.

Еще не началось.

---

# Started

Преподаватель начал занятие.

---

# Completed

Занятие завершено.

Зафиксированы результаты.

---

# Cancelled

Занятие отменено.

История причины сохраняется.

---

# Missed

Ученик отсутствовал.

---

# Rescheduled

Занятие перенесено.

История первоначальной даты сохраняется.

---

# Архивирование

После завершения Lesson никогда не удаляется.

Он становится частью образовательной истории.

---

# Participants

Lesson обязательно содержит.

Teacher

минимум одного Student

дату проведения

---

# Main Sections

Каждый Lesson состоит из нескольких логических частей.

## Preparation

Подготовка преподавателя.

История ученика.

Предыдущие рекомендации.

Домашнее задание.

---

## Session

Что происходило во время занятия.

Какие упражнения выполнялись.

Какие песни разбирались.

Какие проблемы возникли.

---

## Assessment

Педагогическая оценка занятия.

Что получилось.

Что требует внимания.

---

## Homework

Домашнее задание.

---

## Recommendations

Советы преподавателя.

---

## Attachments

Видео.

Фото.

Аудио.

PDF.

Файлы.

---

# Business Rules

Каждый Lesson должен иметь преподавателя.

Каждый Lesson должен иметь минимум одного ученика.

Completed Lesson обязан содержать образовательный результат.

Completed Lesson может существовать без домашнего задания.

Cancelled Lesson не может иметь результат занятия.

Archived Lesson нельзя редактировать.

---

# Commands

CreateLesson

StartLesson

CompleteLesson

CancelLesson

RescheduleLesson

AttachHomework

AttachAssessment

AttachComment

UploadAttachment

ArchiveLesson

---

# Domain Events

LessonCreated

LessonStarted

LessonCompleted

LessonCancelled

LessonRescheduled

HomeworkAssigned

AssessmentAdded

CommentAdded

AttachmentUploaded

LessonArchived

---

# Relationships

```
Lesson

├── Student

├── Teacher

├── Homework

├── Assessment

├── Song

├── Skill

├── Comment

├── Attachment

└── Notification
```

---

# Audit

Все изменения Lesson логируются.

Необходимо сохранять.

Автор изменения.

Дата.

Предыдущее значение.

Новое значение.

Причину изменения.

---

# Permissions

Student

может просматривать собственные Lesson.

Teacher

может создавать и изменять Lesson своих учеников.

Administrator

имеет организационный доступ.

Owner

имеет доступ только для просмотра.

---

# Invariants

Lesson всегда принадлежит образовательному процессу.

Lesson никогда не удаляется.

Completed Lesson нельзя перевести обратно в Started без отдельного административного действия.

Каждый Lesson имеет образовательную ценность.

---

# Non Goals

Lesson не хранит.

Платежи.

CRM-информацию.

Маркетинговые данные.

Финансовые операции.

---

# Future Extensions

В будущем Lesson может содержать.

AI-анализ занятия.

Автоматическую транскрипцию.

Распознавание речи.

Анализ вокала.

Генерацию домашнего задания.

Подготовку материалов.

---

# Open Questions

Поддерживаются ли групповые занятия.

Можно ли менять преподавателя после завершения занятия.

Как учитывать совместные мастер-классы.

Какие ограничения существуют для переноса занятия.

---

# History

| Version | Description |
|----------|-------------|
|1.0.0|Первое описание Lesson Aggregate.|