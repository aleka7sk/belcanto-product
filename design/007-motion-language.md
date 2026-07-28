---
Document Id: DSG-MOTION
Title: Язык движения
Layer: L7 · Design — Форма
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 7.8
Authority Rank: 60
Authority Scope: Визуальный язык, движение, доступность, тон.

Owners:
  - Design Lead

Depends On:
  - design/001-design-principles.md
  - experience/010-interaction-rules.md
  - design/009-accessibility.md

Required By:
  - design/011-design-tokens.md
  - design/006-component-language.md

Defines:
  - Смысл каждого типа движения
  - Длительности и кривые
  - Запреты на движение

Must Not Define:
  - Реализацию анимаций
---

# Язык движения

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Определить, что означает движение в продукте: какие переходы существуют, что они сообщают пользователю, и где движение запрещено.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Design Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`design/001-design-principles.md`](../design/001-design-principles.md)
- [`experience/010-interaction-rules.md`](../experience/010-interaction-rules.md)
- [`design/009-accessibility.md`](../design/009-accessibility.md)

**Зависит от этого документа (обновить при изменении):**

- [`design/011-design-tokens.md`](../design/011-design-tokens.md)
- [`design/006-component-language.md`](../design/006-component-language.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L7 · Design — Форма**
- Позиция в порядке чтения: **7.8**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждое движение сообщает изменение состояния, а не украшает.
2. Для каждого движения определены длительность и кривая.
3. Определено поведение при включённом режиме уменьшенного движения.
4. Ни одно движение не задерживает пользователя дольше установленного предела.

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
