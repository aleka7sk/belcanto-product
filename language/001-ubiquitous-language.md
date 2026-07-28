---
Document Id: LNG-UBIQUITOUS
Title: Единый язык (каталог терминов)
Layer: L3 · Language — Язык продукта
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 3.2
Authority Rank: 20
Authority Scope: Определения терминов. Единственный источник терминологии.

Owners:
  - Education Lead
  - Domain Architecture Lead

Depends On:
  - language/000-language-overview.md
  - product/004-core-concepts.md
  - school/001-education-model.md

Required By:
  - domain/000-domain-overview.md
  - experience/001-information-architecture.md
  - features/000-feature-catalog.md
  - design/010-content-and-voice.md

Defines:
  - Определение каждого термина продукта
  - Канонические имена на русском и английском

Must Not Define:
  - Структуру доменной модели (принадлежит domain/)
  - Пользовательские подписи в интерфейсе (принадлежит language/003)
---

# Единый язык (каталог терминов)

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Единственный авторитетный каталог терминов продукта. Любое слово, имеющее в Belcanto особый смысл, определяется здесь и нигде больше.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Education Lead
- Domain Architecture Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`language/000-language-overview.md`](../language/000-language-overview.md)
- [`product/004-core-concepts.md`](../product/004-core-concepts.md)
- [`school/001-education-model.md`](../school/001-education-model.md)

**Зависит от этого документа (обновить при изменении):**

- [`domain/000-domain-overview.md`](../domain/000-domain-overview.md)
- [`experience/001-information-architecture.md`](../experience/001-information-architecture.md)
- [`features/000-feature-catalog.md`](../features/000-feature-catalog.md)
- [`design/010-content-and-voice.md`](../design/010-content-and-voice.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L3 · Language — Язык продукта**
- Позиция в порядке чтения: **3.2**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждый термин имеет: каноническое имя (RU), техническое имя (EN), определение, статус, документ-владелец смысла.
2. Ни один термин не определяется дважды в репозитории.
3. Каждое доменное имя из domain/entities/, domain/events/, domain/commands/ присутствует в каталоге.
4. Определение не длиннее трёх предложений.
5. Синонимы перечислены явно и помечены как неканонические.

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
