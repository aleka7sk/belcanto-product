---
Document Id: GOV-CHANGE
Title: Управление изменениями
Layer: L9 · Governance — Управление
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 9.2
Authority Rank: 0
Authority Scope: Решения и изменения. Принятое решение (PDR) изменяет любой документ.

Owners:
  - Documentation Lead
  - Product Owner

Depends On:
  - meta/008-lifecycle-and-versioning.md
  - governance/000-governance-overview.md

Required By:
  - features/002-feature-lifecycle.md
  - governance/002-review-checklist.md

Defines:
  - Процесс изменения документа
  - Правила каскадного обновления зависимых документов

Must Not Define:
  - Правила версионирования (принадлежит meta/008)
---

# Управление изменениями

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Определить, что происходит при изменении документа: какие документы обязаны обновиться, кто проверяет, как повышается версия.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Documentation Lead
- Product Owner

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`meta/008-lifecycle-and-versioning.md`](../meta/008-lifecycle-and-versioning.md)
- [`governance/000-governance-overview.md`](../governance/000-governance-overview.md)

**Зависит от этого документа (обновить при изменении):**

- [`features/002-feature-lifecycle.md`](../features/002-feature-lifecycle.md)
- [`governance/002-review-checklist.md`](../governance/002-review-checklist.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L9 · Governance — Управление**
- Позиция в порядке чтения: **9.2**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Для каждого типа изменения указан список обязательно обновляемых документов.
2. Изменение документа с Authority Rank ≤ 30 требует записи PDR.
3. Процесс применим и к человеку, и к AI-агенту.

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
