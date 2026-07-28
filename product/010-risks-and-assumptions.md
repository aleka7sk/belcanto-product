---
Document Id: PRD-RISKS
Title: Риски и допущения
Layer: L2 · Product — Продукт
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 2.12
Authority Rank: 10
Authority Scope: Смысл, границы, цели и правила продукта. Высшая содержательная власть.

Owners:
  - Product Owner
  - Technical Lead

Depends On:
  - product/000-product-overview.md
  - school/007-open-questions.md

Required By:
  - governance/decisions/000-decision-log.md
  - roadmap/000-roadmap-overview.md

Defines:
  - Реестр рисков
  - Реестр допущений
  - Что зависит от каждого допущения

Must Not Define:
  - Решения по рискам (принадлежит governance/decisions/)
---

# Риски и допущения

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Зафиксировать продуктовые риски и допущения, на которых построены решения, чтобы при изменении допущения можно было найти всё, что от него зависит.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner
- Technical Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`product/000-product-overview.md`](../product/000-product-overview.md)
- [`school/007-open-questions.md`](../school/007-open-questions.md)

**Зависит от этого документа (обновить при изменении):**

- [`governance/decisions/000-decision-log.md`](../governance/decisions/000-decision-log.md)
- [`roadmap/000-roadmap-overview.md`](../roadmap/000-roadmap-overview.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L2 · Product — Продукт**
- Позиция в порядке чтения: **2.7**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждый риск имеет id, вероятность, влияние и владельца.
2. Каждое допущение перечисляет документы, которые на нём построены.
3. Опровергнутое допущение обязано породить запись в governance/decisions/.

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
