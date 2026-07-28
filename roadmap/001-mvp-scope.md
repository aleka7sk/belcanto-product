---
Document Id: RMP-MVP
Title: Границы первой версии
Layer: L8 · Delivery — Поставка
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 8.2
Authority Rank: 70
Authority Scope: Объём и последовательность релизов.

Owners:
  - Product Owner

Depends On:
  - features/000-feature-catalog.md
  - features/003-feature-dependency-map.md
  - experience/004-journey-student.md

Required By:
  - roadmap/002-release-plan.md
  - roadmap/004-backlog-and-out-of-scope.md

Defines:
  - Состав первой версии
  - Осознанные исключения

Must Not Define:
  - Сроки (принадлежит roadmap/003)
  - Описание фич (принадлежит features/)
---

# Границы первой версии

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Зафиксировать, что входит в первую работающую версию продукта и что осознанно из неё исключено.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`features/000-feature-catalog.md`](../features/000-feature-catalog.md)
- [`features/003-feature-dependency-map.md`](../features/003-feature-dependency-map.md)
- [`experience/004-journey-student.md`](../experience/004-journey-student.md)

**Зависит от этого документа (обновить при изменении):**

- [`roadmap/002-release-plan.md`](../roadmap/002-release-plan.md)
- [`roadmap/004-backlog-and-out-of-scope.md`](../roadmap/004-backlog-and-out-of-scope.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L8 · Delivery — Поставка**
- Позиция в порядке чтения: **8.2**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждая включённая фича имеет id F-XXX из features/000.
2. Состав первой версии обеспечивает хотя бы один полный путь из experience/004-journey-student.md.
3. Каждое исключение имеет причину и запись в roadmap/004.
4. Состав не нарушает features/003-feature-dependency-map.md.

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
