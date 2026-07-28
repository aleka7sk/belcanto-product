---
Document Id: RMP-RELEASES
Title: План релизов
Layer: L8 · Delivery — Поставка
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 8.3
Authority Rank: 70
Authority Scope: Объём и последовательность релизов.

Owners:
  - Product Owner

Depends On:
  - roadmap/001-mvp-scope.md
  - features/003-feature-dependency-map.md
  - features/002-feature-lifecycle.md

Required By:
  - roadmap/003-milestones.md

Defines:
  - Последовательность релизов
  - Состав каждого релиза

Must Not Define:
  - Даты (принадлежит roadmap/003)
---

# План релизов

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Определить последовательность релизов и состав каждого.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`roadmap/001-mvp-scope.md`](../roadmap/001-mvp-scope.md)
- [`features/003-feature-dependency-map.md`](../features/003-feature-dependency-map.md)
- [`features/002-feature-lifecycle.md`](../features/002-feature-lifecycle.md)

**Зависит от этого документа (обновить при изменении):**

- [`roadmap/003-milestones.md`](../roadmap/003-milestones.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L8 · Delivery — Поставка**
- Позиция в порядке чтения: **8.3**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждая фича релиза имеет id F-XXX.
2. Ни одна фича не выпускается раньше своих зависимостей.
3. Каждый релиз даёт пользователю завершённую ценность.

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
