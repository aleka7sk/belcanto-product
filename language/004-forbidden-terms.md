---
Document Id: LNG-FORBIDDEN
Title: Запрещённые термины
Layer: L3 · Language — Язык продукта
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 3.5
Authority Rank: 20
Authority Scope: Определения терминов. Единственный источник терминологии.

Owners:
  - Education Lead
  - Product Owner

Depends On:
  - language/001-ubiquitous-language.md
  - product/001-vision-and-principles.md

Required By:
  - language/003-ui-terminology.md
  - design/010-content-and-voice.md
  - ai/003-writing-rules.md

Defines:
  - Список запрещённых слов
  - Разрешённую замену
  - Причину запрета

Must Not Define:
  - Определения разрешённых терминов (принадлежит language/001)
---

# Запрещённые термины

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Перечислить слова, которые продукт не использует — ни в документации, ни в интерфейсе — и указать разрешённую замену для каждого.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Education Lead
- Product Owner

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`language/001-ubiquitous-language.md`](../language/001-ubiquitous-language.md)
- [`product/001-vision-and-principles.md`](../product/001-vision-and-principles.md)

**Зависит от этого документа (обновить при изменении):**

- [`language/003-ui-terminology.md`](../language/003-ui-terminology.md)
- [`design/010-content-and-voice.md`](../design/010-content-and-voice.md)
- [`ai/003-writing-rules.md`](../ai/003-writing-rules.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L3 · Language — Язык продукта**
- Позиция в порядке чтения: **3.5**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждое запрещённое слово имеет замену и причину.
2. Причина ссылается на product/001-vision-and-principles.md или product/009-product-rules.md.
3. Список применим и к документации, и к интерфейсу.

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
