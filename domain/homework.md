---
Status: Approved
Version: 1.0.0

Entity: Homework

Owners:
  - Product Owner
  - Technical Lead

Aggregate: Lesson

Related:
  - lesson.md
  - student.md
  - teacher.md
  - skill.md
  - song.md
  - attachment.md
---

# Homework

> Homework представляет собой индивидуальный план самостоятельной работы ученика между занятиями.
>
> Его цель — продолжить образовательный процесс за пределами урока.

---

# Purpose

Домашнее задание фиксирует действия, которые ученик должен выполнить до следующего занятия.

Homework помогает:

- закрепить материал;
- развить необходимые навыки;
- подготовиться к следующему уроку;
- сформировать привычку регулярной практики.

Homework является продолжением Lesson.

---

# Responsibilities

Homework отвечает за:

- список заданий;
- рекомендации преподавателя;
- срок выполнения;
- статус выполнения;
- результат выполнения;
- обратную связь преподавателя.

Homework не отвечает за:

- расписание;
- оценку ученика;
- историю занятий;
- концертную деятельность.

---

# Lifecycle

```
Draft

↓

Assigned

↓

In Progress

↓

Submitted

↓

Reviewed

↓

Completed
```

Дополнительно:

```
Cancelled

Expired
```

---

# States

## Draft

Черновик домашнего задания.

Доступен только преподавателю.

---

## Assigned

Домашнее задание выдано ученику.

---

## In Progress

Ученик приступил к выполнению.

---

## Submitted

Ученик сообщил о завершении.

При необходимости прикрепил материалы.

---

## Reviewed

Преподаватель проверил выполнение.

Добавил комментарии.

---

## Completed

Домашнее задание считается завершенным.

---

## Expired

Срок выполнения истек.

Домашнее задание остается в истории.

---

## Cancelled

Домашнее задание отменено преподавателем.

Причина отмены сохраняется.

---

# Structure

Homework состоит из набора Task.

Каждый Task представляет отдельное действие.

Например:

- выполнить дыхательное упражнение;
- записать видео исполнения;
- выучить куплет;
- прослушать запись;
- повторить распевку.

---

# Homework Task

Каждый Task содержит:

- название;
- описание;
- рекомендуемое время;
- связанный Skill (необязательно);
- связанную Song (необязательно);
- материалы;
- статус.

---

# Business Rules

Homework всегда принадлежит Lesson.

Homework всегда имеет автора.

Homework всегда связан минимум с одним Student.

Completed Homework нельзя изменить.

Каждое изменение фиксируется.

---

# Commands

CreateHomework

AssignHomework

StartHomework

SubmitHomework

ReviewHomework

CompleteHomework

CancelHomework

AddTask

RemoveTask

UpdateTask

AttachFile

---

# Domain Events

HomeworkCreated

HomeworkAssigned

HomeworkStarted

HomeworkSubmitted

HomeworkReviewed

HomeworkCompleted

HomeworkCancelled

TaskAdded

TaskCompleted

AttachmentUploaded

---

# Relationships

```
Homework

├── Lesson
├── Teacher
├── Student
├── Task
├── Skill
├── Song
└── Attachment
```

---

# Attachments

Homework может содержать:

- видео;
- аудио;
- фотографии;
- PDF;
- ссылки;
- текстовые комментарии.

---

# Permissions

Student

- просмотр;
- выполнение;
- загрузка материалов.

Teacher

- создание;
- изменение;
- проверка;
- завершение.

Administrator

- просмотр.

Owner

- просмотр.

---

# Invariants

Homework никогда не существует без Lesson.

Homework никогда не удаляется.

История выполнения сохраняется полностью.

Статус Completed является финальным.

---

# Future Extensions

В будущем возможны:

- AI-проверка выполнения;
- автоматическое распознавание вокала;
- напоминания;
- адаптивные домашние задания;
- персональные рекомендации;
- автоматическая генерация упражнений.

---

# Open Questions

Можно ли назначать одно Homework нескольким ученикам?

Допускается ли совместное выполнение?

Можно ли копировать Homework между Lesson?

Как учитывать частичное выполнение?

---

# History

| Version | Description |
|---------|-------------|
|1.0.0|Первое описание Homework.|