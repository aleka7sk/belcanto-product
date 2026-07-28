---
Document Id: EXP-IA
Title: Информационная архитектура
Layer: L5 · Experience — Опыт
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 5.2
Authority Rank: 40
Authority Scope: Структура информации, навигация, пользовательские сценарии.

Owners:
  - Product Owner
  - Design Lead

Depends On:
  - product/004-core-concepts.md
  - language/001-ubiquitous-language.md
  - domain/aggregates/000-aggregate-catalog.md

Required By:
  - experience/002-navigation-model.md
  - experience/003-screen-catalog.md

Defines:
  - Иерархию разделов продукта
  - Место каждой сущности в структуре
  - Правила группировки информации

Must Not Define:
  - Способ перехода между разделами (принадлежит experience/002)
  - Визуальное оформление (принадлежит design/)
---

# Информационная архитектура

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Определить, из каких смысловых разделов состоит продукт, как знание сгруппировано и где живёт каждая сущность с точки зрения пользователя.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner
- Design Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`product/004-core-concepts.md`](../product/004-core-concepts.md)
- [`language/001-ubiquitous-language.md`](../language/001-ubiquitous-language.md)
- [`domain/aggregates/000-aggregate-catalog.md`](../domain/aggregates/000-aggregate-catalog.md)

**Зависит от этого документа (обновить при изменении):**

- [`experience/002-navigation-model.md`](../experience/002-navigation-model.md)
- [`experience/003-screen-catalog.md`](../experience/003-screen-catalog.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L5 · Experience — Опыт**
- Позиция в порядке чтения: **5.2**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждая концепция из product/004-core-concepts.md имеет место в структуре.
2. Ни одна сущность не находится в двух разделах одновременно.
3. Глубина иерархии не превышает трёх уровней.
4. Названия разделов взяты из language/003-ui-terminology.md.

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
