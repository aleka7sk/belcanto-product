---
Document Id: PRD-VALUE
Title: Ценностные предложения по ролям
Layer: L2 · Product — Продукт
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 2.10
Authority Rank: 10
Authority Scope: Смысл, границы, цели и правила продукта. Высшая содержательная власть.

Owners:
  - Product Owner
  - Education Lead

Depends On:
  - product/003-personas.md
  - school/002-current-student-journey.md
  - school/003-teacher-workflow.md

Required By:
  - features/000-feature-catalog.md
  - experience/000-experience-overview.md

Defines:
  - Ценностное предложение для каждой роли
  - Снимаемую боль

Must Not Define:
  - Описание персон (принадлежит product/003)
  - Описание сценариев (принадлежит experience/)
---

# Ценностные предложения по ролям

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Сформулировать, какую конкретную ценность продукт даёт каждой роли — ученику, преподавателю, администратору, руководителю — и какую боль он снимает.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner
- Education Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`product/003-personas.md`](../product/003-personas.md)
- [`school/002-current-student-journey.md`](../school/002-current-student-journey.md)
- [`school/003-teacher-workflow.md`](../school/003-teacher-workflow.md)

**Зависит от этого документа (обновить при изменении):**

- [`features/000-feature-catalog.md`](../features/000-feature-catalog.md)
- [`experience/000-experience-overview.md`](../experience/000-experience-overview.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L2 · Product — Продукт**
- Позиция в порядке чтения: **2.5**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждая роль из product/003-personas.md имеет ровно один блок ценности.
2. Каждая боль подтверждена ссылкой на документ слоя school/.
3. Ни одно предложение не обещает того, что запрещено в product/002-product-boundaries.md.

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
