---
Document Id: LNG-L10N
Title: Политика локализации
Layer: L3 · Language — Язык продукта
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 3.6
Authority Rank: 20
Authority Scope: Определения терминов. Единственный источник терминологии.

Owners:
  - Product Owner
  - Design Lead

Depends On:
  - language/001-ubiquitous-language.md
  - language/003-ui-terminology.md

Required By:
  - design/010-content-and-voice.md
  - experience/003-screen-catalog.md

Defines:
  - Список поддерживаемых языков
  - Исходный язык
  - Правила перевода терминов

Must Not Define:
  - Сами переводы строк интерфейса
---

# Политика локализации

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Определить, на каких языках существует продукт, какой язык является исходным, и как термины переводятся без потери смысла.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner
- Design Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`language/001-ubiquitous-language.md`](../language/001-ubiquitous-language.md)
- [`language/003-ui-terminology.md`](../language/003-ui-terminology.md)

**Зависит от этого документа (обновить при изменении):**

- [`design/010-content-and-voice.md`](../design/010-content-and-voice.md)
- [`experience/003-screen-catalog.md`](../experience/003-screen-catalog.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L3 · Language — Язык продукта**
- Позиция в порядке чтения: **3.6**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Указан ровно один исходный язык.
2. Для каждого языка указан статус поддержки.
3. Правила перевода терминов не допускают появления нового смысла.

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
