---
Document Id: EXP-NAV
Title: Модель навигации
Layer: L5 · Experience — Опыт
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 5.3
Authority Rank: 40
Authority Scope: Структура информации, навигация, пользовательские сценарии.

Owners:
  - Product Owner
  - Design Lead

Depends On:
  - experience/001-information-architecture.md
  - product/003-personas.md

Required By:
  - experience/003-screen-catalog.md
  - experience/004-journey-student.md
  - design/006-component-language.md

Defines:
  - Модель навигации для каждой роли
  - Точки входа
  - Правила переходов и возврата

Must Not Define:
  - Иерархию разделов (принадлежит experience/001)
  - Внешний вид навигационных элементов (принадлежит design/006)
---

# Модель навигации

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Определить, как пользователь перемещается по продукту: точки входа, основная навигация, переходы, возвраты, глубинные ссылки.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner
- Design Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`experience/001-information-architecture.md`](../experience/001-information-architecture.md)
- [`product/003-personas.md`](../product/003-personas.md)

**Зависит от этого документа (обновить при изменении):**

- [`experience/003-screen-catalog.md`](../experience/003-screen-catalog.md)
- [`experience/004-journey-student.md`](../experience/004-journey-student.md)
- [`design/006-component-language.md`](../design/006-component-language.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L5 · Experience — Опыт**
- Позиция в порядке чтения: **5.3**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждая роль из product/003-personas.md имеет описанную навигацию.
2. Из любого экрана существует путь на верхний уровень.
3. Каждый раздел из experience/001 достижим не более чем за три перехода.
4. Нет навигационных элементов, ведущих за границы product/002-product-boundaries.md.

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
