---
Document Id: PRD-CORE-CONCEPTS
Title: Ключевые концепции продукта
Layer: L2 · Product — Продукт
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 2.1
Authority Rank: 10
Authority Scope: Смысл, границы, цели и правила продукта. Высшая содержательная власть.

Owners:
  - Product Owner
  - Education Lead

Depends On:
  - product/000-product-overview.md
  - product/001-vision-and-principles.md

Required By:
  - language/001-ubiquitous-language.md
  - domain/000-domain-overview.md
  - experience/001-information-architecture.md

Defines:
  - Список ключевых концепций продукта
  - Смысл каждой концепции для пользователя

Must Not Define:
  - Определения терминов (принадлежит language/001)
  - Структуру данных и атрибуты (принадлежит domain/)
---

# Ключевые концепции продукта

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Описать 7–12 несущих идей продукта (сопровождение, история, навык, прогресс, репертуар, достижение и т.д.) на продуктовом уровне — до того, как они превращаются в доменную модель.

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

**Зависит от этого документа (обновить при изменении):**

- [`language/001-ubiquitous-language.md`](../language/001-ubiquitous-language.md)
- [`domain/000-domain-overview.md`](../domain/000-domain-overview.md)
- [`experience/001-information-architecture.md`](../experience/001-information-architecture.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L2 · Product — Продукт**
- Позиция в порядке чтения: **2.1**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждая концепция объясняет ценность для пользователя, а не структуру данных.
2. Каждая концепция имеет соответствующий термин в language/001-ubiquitous-language.md.
3. Количество концепций не превышает 12.

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
