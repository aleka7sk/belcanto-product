---
Document Id: PRD-RULES
Title: Продуктовые правила
Layer: L2 · Product — Продукт
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 2.6
Authority Rank: 10
Authority Scope: Смысл, границы, цели и правила продукта. Высшая содержательная власть.

Owners:
  - Product Owner
  - Education Lead

Depends On:
  - product/000-product-overview.md
  - product/001-vision-and-principles.md
  - product/002-product-boundaries.md

Required By:
  - features/000-feature-catalog.md
  - experience/010-interaction-rules.md
  - design/000-design-overview.md
  - governance/002-review-checklist.md

Defines:
  - Инварианты продукта
  - Запреты продукта
  - Правила приёмки продуктового решения

Must Not Define:
  - Правила оформления документов (принадлежит meta/)
  - Правила доменного моделирования (принадлежит domain/architecture/)
---

# Продуктовые правила

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Свод инвариантов продукта: правила, которые обязаны соблюдаться в любой фиче, на любом экране, в любом релизе. Проверочный лист для приёмки любого решения.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner
- Education Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`product/000-product-overview.md`](../product/000-product-overview.md)
- [`product/001-vision-and-principles.md`](../product/001-vision-and-principles.md)
- [`product/002-product-boundaries.md`](../product/002-product-boundaries.md)

**Зависит от этого документа (обновить при изменении):**

- [`features/000-feature-catalog.md`](../features/000-feature-catalog.md)
- [`experience/010-interaction-rules.md`](../experience/010-interaction-rules.md)
- [`design/000-design-overview.md`](../design/000-design-overview.md)
- [`governance/002-review-checklist.md`](../governance/002-review-checklist.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L2 · Product — Продукт**
- Позиция в порядке чтения: **2.6**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждое правило сформулировано как проверяемое утверждение (да/нет).
2. Каждое правило имеет id вида PR-XX и ссылку на источник в product/000 или product/001.
3. Правила не дублируют друг друга и не противоречат друг другу.
4. Каждое правило можно нарушить — правило, которое невозможно нарушить, удаляется.

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
