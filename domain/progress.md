---
Status: Approved
Version: 1.0.0

Entity: Progress

Owners:
  - Product Owner
  - Technical Lead

Related:
  - assessment.md
  - lesson.md
  - student.md
  - skill.md
  - goal.md
  - homework.md
  - concert.md
---

# Progress

> Progress представляет собой накопленную картину развития ученика во времени.
>
> Он показывает не единичное событие, а устойчивую динамику обучения.

---

# Purpose

Progress существует для ответа на вопросы:

- Как развивается ученик?
- Что изменилось за последний месяц?
- Какие навыки улучшаются?
- Где развитие остановилось?
- Какие цели уже достигнуты?
- Какие направления требуют внимания?

---

# Core Principle

Progress никогда не строится на одном событии.

Он является интерпретацией истории.

Progress формируется из множества источников.

---

# Sources

Progress может использовать следующие источники.

- Assessment
- Lesson
- Homework
- Concert
- Goal
- Self Assessment
- Teacher Review
- Practice History

Ни один источник не обладает абсолютным приоритетом.

---

# Responsibilities

Progress отвечает за:

- историю развития;
- изменение навыков;
- изменение целей;
- накопление Evidence;
- временную динамику;
- визуализацию роста.

Progress не отвечает за:

- создание Assessment;
- проведение Lesson;
- выполнение Homework;
- изменение Student.

---

# Lifecycle

```
Initialized

↓

Tracking

↓

Archived
```

---

# Initialized

Создан профиль развития.

История еще отсутствует.

---

# Tracking

Прогресс активно обновляется.

---

# Archived

Прогресс переведен в архив вместе с Student.

---

# Progress Areas

Прогресс может отслеживаться по различным направлениям.

Например.

- Breathing
- Rhythm
- Intonation
- Articulation
- Resonance
- Stage Presence
- Emotional Expression
- Performance Confidence

Каждое направление развивается независимо.

---

# Timeline

Каждое изменение Progress должно быть привязано ко времени.

История должна позволять ответить:

Что было известно на любую дату?

---

# Trend

Для каждого направления определяется тенденция.

Например.

Improving

Stable

Needs Attention

Unknown

Trend не является оценкой ученика.

Trend отражает направление изменения.

---

# Evidence

Каждый вывод Progress обязан иметь Evidence.

Например.

```
Progress

↓

Assessment #104

↓

Lesson #82

↓

Homework #51

↓

Concert #12
```

Пользователь должен иметь возможность увидеть источник вывода.

---

# Business Rules

Progress существует только для Student.

Progress никогда не удаляется.

Все изменения объяснимы.

Каждое изменение имеет историю.

Progress не может противоречить имеющимся Evidence.

---

# Commands

InitializeProgress

UpdateProgress

AttachEvidence

ArchiveProgress

RecalculateProgress

---

# Domain Events

ProgressInitialized

ProgressUpdated

EvidenceAttached

ProgressArchived

ProgressRecalculated

---

# Relationships

```
Progress

├── Student
├── Assessment
├── Lesson
├── Homework
├── Goal
├── Skill
├── Concert
└── Timeline
```

---

# Invariants

Каждый Progress принадлежит одному Student.

Progress всегда содержит историю изменений.

Progress никогда не строится без Evidence.

Все изменения должны быть воспроизводимыми.

Progress не может содержать необъяснимых выводов.

---

# Read Models

## Student View

Показывает:

- текущую динамику;
- достижения;
- рекомендации;
- историю роста.

---

## Teacher View

Показывает:

- Evidence;
- историю Assessment;
- изменение Skills;
- изменение Goals.

---

## Owner Analytics

Показывает агрегированные данные по школе.

Без раскрытия лишней персональной информации.

---

# AI Assistance

AI может:

- искать закономерности;
- выявлять изменение динамики;
- предлагать возможные рекомендации;
- помогать визуализировать историю.

AI не может самостоятельно изменять Progress.

Любое изменение должно быть подтверждено утвержденной Progress Policy или преподавателем.

---

# Privacy

Progress содержит чувствительную образовательную информацию.

Необходимо обеспечить:

- ролевой доступ;
- аудит просмотра;
- контроль экспорта;
- безопасное хранение.

---

# Non-Goals

Progress не предназначен для:

- сравнения учеников между собой;
- формирования публичных рейтингов;
- определения ценности ученика;
- расчета оплаты преподавателей;
- принятия педагогических решений без участия человека.

---

# Failure Cases

## Недостаточно Evidence

Progress не обновляется.

---

## Конфликтующие Assessment

Progress отмечает необходимость ручного анализа.

---

## Попытка изменить историю

Операция запрещается.

История является неизменяемой.

---

# Open Questions

Необходимо определить:

- минимальное количество Evidence для изменения Trend;
- срок хранения старых данных;
- допустимые алгоритмы агрегации;
- правила отображения ученику;
- какие изменения требуют подтверждения преподавателя.

---

# Future Extensions

В будущем могут появиться:

- прогноз достижения целей;
- AI-анализ динамики;
- сравнение разных периодов обучения;
- визуальная карта развития;
- персональные рекомендации;
- прогноз готовности к концерту.

---

# History

| Version | Description |
|---------|-------------|
|1.0.0|Первое описание Progress.|