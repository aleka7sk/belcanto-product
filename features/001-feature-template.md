---
Document Id: FEAT-TEMPLATE
Title: Шаблон описания фичи
Layer: L6 · Features — Функциональность
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 6.2
Authority Rank: 50
Authority Scope: Поведение конкретных возможностей продукта.

Owners:
  - Product Owner
  - Documentation Lead

Depends On:
  - meta/003-document-template.md
  - features/000-feature-catalog.md

Required By:
  - features/000-feature-catalog.md

Defines:
  - Обязательные разделы описания фичи

Must Not Define:
  - Описание конкретных фич
---

# Шаблон описания фичи

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Обязательная форма файла фичи. Все файлы в features/catalog/ создаются копированием этого шаблона.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner
- Documentation Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`meta/003-document-template.md`](../meta/003-document-template.md)
- [`features/000-feature-catalog.md`](../features/000-feature-catalog.md)

**Зависит от этого документа (обновить при изменении):**

- [`features/000-feature-catalog.md`](../features/000-feature-catalog.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L6 · Features — Функциональность**
- Позиция в порядке чтения: **6.2**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Шаблон содержит обязательные разделы: назначение, роли, сценарий, состояния, команды, события, правила приёмки.
2. Шаблон соответствует meta/003-document-template.md.
3. Шаблон не содержит примеров реальных фич.

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
