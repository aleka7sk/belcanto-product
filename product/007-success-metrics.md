---
Document Id: PRD-METRICS
Title: Метрики успеха
Layer: L2 · Product — Продукт
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 2.4
Authority Rank: 10
Authority Scope: Смысл, границы, цели и правила продукта. Высшая содержательная власть.

Owners:
  - Product Owner

Depends On:
  - product/006-business-goals.md
  - product/000-product-overview.md

Required By:
  - features/000-feature-catalog.md
  - roadmap/003-milestones.md

Defines:
  - Определение каждой метрики
  - Способ расчёта
  - Целевые значения

Must Not Define:
  - Формулировку самих целей (принадлежит product/006)
  - Реализацию аналитики
---

# Метрики успеха

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Определить измеримые признаки успеха продукта: какие показатели мы наблюдаем, как они считаются и какие значения считаются нормой.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`product/006-business-goals.md`](../product/006-business-goals.md)
- [`product/000-product-overview.md`](../product/000-product-overview.md)

**Зависит от этого документа (обновить при изменении):**

- [`features/000-feature-catalog.md`](../features/000-feature-catalog.md)
- [`roadmap/003-milestones.md`](../roadmap/003-milestones.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L2 · Product — Продукт**
- Позиция в порядке чтения: **2.4**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждая метрика ссылается ровно на одну бизнес-цель BG-XX.
2. Каждая метрика имеет источник данных, выраженный через доменные события из domain/events/.
3. Метрика не измеряет сравнение учеников между собой (запрет product/000).

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
