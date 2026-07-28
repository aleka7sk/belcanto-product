---
Document Id: PRD-BUSINESS-GOALS
Title: Бизнес-цели
Layer: L2 · Product — Продукт
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 2.8
Authority Rank: 10
Authority Scope: Смысл, границы, цели и правила продукта. Высшая содержательная власть.

Owners:
  - Product Owner

Depends On:
  - product/000-product-overview.md
  - product/001-product-vision.md
  - school/005-current-owner-workflow.md

Required By:
  - product/007-success-metrics.md
  - roadmap/000-roadmap-overview.md
  - features/000-feature-catalog.md

Defines:
  - Перечень бизнес-целей
  - Горизонт каждой цели
  - Владельца каждой цели

Must Not Define:
  - Метрики и их формулы (принадлежит product/007)
  - Сроки релизов (принадлежит roadmap/)
---

# Бизнес-цели

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Зафиксировать, зачем школе нужен продукт с точки зрения бизнеса: какие цели он преследует, в каком горизонте и по каким признакам цель считается достигнутой.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`product/000-product-overview.md`](../product/000-product-overview.md)
- [`school/005-current-owner-workflow.md`](../school/005-current-owner-workflow.md)

**Зависит от этого документа (обновить при изменении):**

- [`product/007-success-metrics.md`](../product/007-success-metrics.md)
- [`roadmap/000-roadmap-overview.md`](../roadmap/000-roadmap-overview.md)
- [`features/000-feature-catalog.md`](../features/000-feature-catalog.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L2 · Product — Продукт**
- Позиция в порядке чтения: **2.3**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждая цель имеет id вида BG-XX.
2. Каждая цель связана хотя бы с одной метрикой из product/007-success-metrics.md.
3. Ни одна цель не противоречит product/002-product-boundaries.md.
4. Число целей не превышает 7.

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
