---
Document Id: EXP-NOTIFY
Title: Опыт уведомлений
Layer: L5 · Experience — Опыт
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 5.10
Authority Rank: 40
Authority Scope: Структура информации, навигация, пользовательские сценарии.

Owners:
  - Product Owner
  - Education Lead

Depends On:
  - domain/policies/notification-policy.md
  - product/001-vision-and-principles.md
  - design/010-content-and-voice.md

Required By:
  - features/000-feature-catalog.md
  - experience/004-journey-student.md

Defines:
  - Правила опыта уведомлений
  - Ограничения по частоте
  - Права пользователя на управление

Must Not Define:
  - Доменные правила отправки (принадлежит domain/policies/notification-policy.md)
---

# Опыт уведомлений

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Определить, как продукт обращается к пользователю вне приложения: поводы, частота, тон, право пользователя отказаться.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner
- Education Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`domain/policies/notification-policy.md`](../domain/policies/notification-policy.md)
- [`product/001-vision-and-principles.md`](../product/001-vision-and-principles.md)
- [`design/010-content-and-voice.md`](../design/010-content-and-voice.md)

**Зависит от этого документа (обновить при изменении):**

- [`features/000-feature-catalog.md`](../features/000-feature-catalog.md)
- [`experience/004-journey-student.md`](../experience/004-journey-student.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L5 · Experience — Опыт**
- Позиция в порядке чтения: **5.10**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждый повод для уведомления соответствует доменному событию.
2. Определён верхний предел частоты уведомлений.
3. Ни одно уведомление не мотивирует через чувство вины (запрет product/000).
4. Документ не переопределяет domain/policies/notification-policy.md, а ссылается на него.

Дополнительно применяются общие правила: [`meta/006-validation-rules.md`](../meta/006-validation-rules.md)

---

## 6. Content · Содержание

<!-- SCAFFOLD: содержание не написано.
     Перед заполнением прочитать документы из раздела 3.
     После заполнения: Status → Draft, обновить meta/007-ownership-map.md. -->

_Содержание не написано._

---

## История изменений

| Версия | Дата | Изменение |
|--------|------|-----------|
| 0.1.0 | 2026-07-28 | Создан скелет документа. |
