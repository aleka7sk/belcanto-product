---
Document Id: FEAT-LIFECYCLE
Title: Жизненный цикл фичи
Layer: L6 · Features — Функциональность
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 6.3
Authority Rank: 50
Authority Scope: Поведение конкретных возможностей продукта.

Owners:
  - Product Owner

Depends On:
  - features/000-feature-catalog.md
  - governance/001-change-management.md

Required By:
  - roadmap/002-release-plan.md

Defines:
  - Статусы фичи
  - Условия перехода между статусами

Must Not Define:
  - Сроки релизов (принадлежит roadmap/)
---

# Жизненный цикл фичи

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Определить статусы фичи (идея → принята → в разработке → выпущена → выведена) и условия перехода между ними.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`features/000-feature-catalog.md`](../features/000-feature-catalog.md)
- [`governance/001-change-management.md`](../governance/001-change-management.md)

**Зависит от этого документа (обновить при изменении):**

- [`roadmap/002-release-plan.md`](../roadmap/002-release-plan.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L6 · Features — Функциональность**
- Позиция в порядке чтения: **6.3**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждый статус имеет условие входа и выхода.
2. Переход в статус «принята» требует записи в governance/decisions/.
3. Ни одна фича в features/000 не имеет статуса вне этого списка.

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
