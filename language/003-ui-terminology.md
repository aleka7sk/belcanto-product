---
Document Id: LNG-UI-TERMS
Title: Терминология интерфейса
Layer: L3 · Language — Язык продукта
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 3.4
Authority Rank: 20
Authority Scope: Определения терминов. Единственный источник терминологии.

Owners:
  - Design Lead
  - Education Lead

Depends On:
  - language/001-ubiquitous-language.md
  - design/010-content-and-voice.md

Required By:
  - experience/003-screen-catalog.md
  - features/000-feature-catalog.md

Defines:
  - Соответствие «доменный термин → подпись в интерфейсе»
  - Список подписей, применяемых во всех экранах

Must Not Define:
  - Определения терминов (принадлежит language/001)
  - Тон и стиль текста (принадлежит design/010)
---

# Терминология интерфейса

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Сопоставить доменные термины и слова, которые видит пользователь. Внутреннее имя и подпись на экране — не одно и то же, и это соответствие должно быть зафиксировано.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Design Lead
- Education Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`language/001-ubiquitous-language.md`](../language/001-ubiquitous-language.md)
- [`design/010-content-and-voice.md`](../design/010-content-and-voice.md)

**Зависит от этого документа (обновить при изменении):**

- [`experience/003-screen-catalog.md`](../experience/003-screen-catalog.md)
- [`features/000-feature-catalog.md`](../features/000-feature-catalog.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L3 · Language — Язык продукта**
- Позиция в порядке чтения: **3.4**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждая подпись сопоставлена ровно одному термину из language/001.
2. Один термин не имеет двух разных подписей в одном контексте.
3. Подписи не содержат слов из language/004-forbidden-terms.md.

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
