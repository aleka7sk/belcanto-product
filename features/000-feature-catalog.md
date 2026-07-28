---
Document Id: FEAT-CATALOG
Title: Каталог функциональности
Layer: L6 · Features — Функциональность
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 6.1
Authority Rank: 50
Authority Scope: Поведение конкретных возможностей продукта.

Owners:
  - Product Owner

Depends On:
  - product/008-value-propositions.md
  - experience/003-screen-catalog.md
  - product/006-business-goals.md

Required By:
  - roadmap/001-mvp-scope.md
  - features/003-feature-dependency-map.md

Defines:
  - Реестр фич и их идентификаторов
  - Статус каждой фичи
  - Владельца каждой фичи

Must Not Define:
  - Подробное поведение фичи (принадлежит features/catalog/F-XXX-*.md)
---

# Каталог функциональности

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Единый реестр всех возможностей продукта: идентификатор, статус, роль, ценность, документ-описание. Точка входа во всё, что продукт умеет делать.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`product/008-value-propositions.md`](../product/008-value-propositions.md)
- [`experience/003-screen-catalog.md`](../experience/003-screen-catalog.md)
- [`product/006-business-goals.md`](../product/006-business-goals.md)

**Зависит от этого документа (обновить при изменении):**

- [`roadmap/001-mvp-scope.md`](../roadmap/001-mvp-scope.md)
- [`features/003-feature-dependency-map.md`](../features/003-feature-dependency-map.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L6 · Features — Функциональность**
- Позиция в порядке чтения: **6.1**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждая фича имеет id вида F-XXX и файл в features/catalog/.
2. Каждая фича связана хотя бы с одной бизнес-целью BG-XX.
3. Каждая фича связана хотя бы с одним экраном SCR-XXX.
4. Каждая фича имеет статус из features/002-feature-lifecycle.md.
5. Нет двух фич с пересекающимся назначением.

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
