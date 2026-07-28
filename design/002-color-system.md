---
Document Id: DSG-COLOR
Title: Цветовая система
Layer: L7 · Design — Форма
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 7.3
Authority Rank: 60
Authority Scope: Визуальный язык, движение, доступность, тон.

Owners:
  - Design Lead

Depends On:
  - design/001-design-principles.md
  - design/009-accessibility.md

Required By:
  - design/011-design-tokens.md
  - design/006-component-language.md

Defines:
  - Смысловые роли цвета
  - Поведение в темах

Must Not Define:
  - Значения токенов (принадлежит design/011)
---

# Цветовая система

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Определить смысловую роль цвета: какие роли существуют, что каждая означает, и как цвет ведёт себя в светлой и тёмной теме.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Design Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`design/001-design-principles.md`](../design/001-design-principles.md)
- [`design/009-accessibility.md`](../design/009-accessibility.md)

**Зависит от этого документа (обновить при изменении):**

- [`design/011-design-tokens.md`](../design/011-design-tokens.md)
- [`design/006-component-language.md`](../design/006-component-language.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L7 · Design — Форма**
- Позиция в порядке чтения: **7.3**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждая роль цвета имеет определённый смысл, а не только значение.
2. Каждая пара «текст / фон» соответствует требованиям design/009-accessibility.md.
3. Ни одна роль не выражена только цветом без второго признака.

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
