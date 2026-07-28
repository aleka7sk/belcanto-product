---
Document Id: LNG-NAMING
Title: Правила именования
Layer: L3 · Language — Язык продукта
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 3.3
Authority Rank: 20
Authority Scope: Определения терминов. Единственный источник терминологии.

Owners:
  - Domain Architecture Lead
  - Technical Lead

Depends On:
  - language/001-ubiquitous-language.md
  - domain/architecture/000-domain-model-rules.md

Required By:
  - features/001-feature-template.md
  - meta/003-document-template.md

Defines:
  - Шаблоны имён и идентификаторов
  - Правила образования новых имён

Must Not Define:
  - Конкретные имена сущностей (принадлежит domain/)
---

# Правила именования

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Задать правила образования имён: сущностей, событий, команд, фич, экранов, документов, идентификаторов.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Domain Architecture Lead
- Technical Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`language/001-ubiquitous-language.md`](../language/001-ubiquitous-language.md)
- [`domain/architecture/000-domain-model-rules.md`](../domain/architecture/000-domain-model-rules.md)

**Зависит от этого документа (обновить при изменении):**

- [`features/001-feature-template.md`](../features/001-feature-template.md)
- [`meta/003-document-template.md`](../meta/003-document-template.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L3 · Language — Язык продукта**
- Позиция в порядке чтения: **3.3**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждое правило иллюстрировано парой «верно / неверно».
2. Правила покрывают все типы идентификаторов, используемые в репозитории.
3. Правила не противоречат domain/architecture/000-domain-model-rules.md.

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
