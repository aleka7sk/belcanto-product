---
Document Id: DSG-TOKENS
Title: Дизайн-токены
Layer: L7 · Design — Форма
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 7.12
Authority Rank: 60
Authority Scope: Визуальный язык, движение, доступность, тон.

Owners:
  - Design Lead
  - Technical Lead

Depends On:
  - design/002-color-system.md
  - design/003-typography.md
  - design/004-spacing-and-layout.md
  - design/007-motion-language.md

Required By:
  - (нет)

Defines:
  - Имена и значения токенов

Must Not Define:
  - Смысл ролей (принадлежит профильным документам design/)
---

# Дизайн-токены

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Единый реестр именованных значений формы: цвет, размер, отступ, длительность. Единственное место, где живут конкретные значения.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Design Lead
- Technical Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`design/002-color-system.md`](../design/002-color-system.md)
- [`design/003-typography.md`](../design/003-typography.md)
- [`design/004-spacing-and-layout.md`](../design/004-spacing-and-layout.md)
- [`design/007-motion-language.md`](../design/007-motion-language.md)

**Зависит от этого документа (обновить при изменении):**

- (нет)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L7 · Design — Форма**
- Позиция в порядке чтения: **7.12**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждый токен имеет имя по language/002-naming-rules.md.
2. Каждое значение существует ровно в одном токене.
3. Каждый токен ссылается на документ, определяющий его смысл.
4. Ни один документ, кроме этого, не содержит конкретных значений.

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
