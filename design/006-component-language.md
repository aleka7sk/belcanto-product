---
Document Id: DSG-COMPONENTS
Title: Язык компонентов
Layer: L7 · Design — Форма
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 7.7
Authority Rank: 60
Authority Scope: Визуальный язык, движение, доступность, тон.

Owners:
  - Design Lead

Depends On:
  - design/002-color-system.md
  - design/004-spacing-and-layout.md
  - experience/003-screen-catalog.md

Required By:
  - experience/008-states-and-empty-states.md
  - design/007-motion-language.md

Defines:
  - Перечень компонентов
  - Назначение и правила выбора компонента

Must Not Define:
  - Реализацию компонентов в коде
  - Содержимое экранов
---

# Язык компонентов

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Определить набор смысловых компонентов интерфейса, назначение каждого и правила выбора между ними.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Design Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`design/002-color-system.md`](../design/002-color-system.md)
- [`design/004-spacing-and-layout.md`](../design/004-spacing-and-layout.md)
- [`experience/003-screen-catalog.md`](../experience/003-screen-catalog.md)

**Зависит от этого документа (обновить при изменении):**

- [`experience/008-states-and-empty-states.md`](../experience/008-states-and-empty-states.md)
- [`design/007-motion-language.md`](../design/007-motion-language.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L7 · Design — Форма**
- Позиция в порядке чтения: **7.7**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждый компонент решает ровно одну задачу.
2. Для каждого компонента определены все состояния из experience/008.
3. Нет двух компонентов с одинаковым назначением.
4. Каждый компонент используется хотя бы на одном экране из experience/003.

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
