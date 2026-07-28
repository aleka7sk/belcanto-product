---
Document Id: FEAT-CATALOG-DIR
Title: Каталог фич — правила каталога
Layer: L6 · Features — Функциональность
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 6.5
Authority Rank: 50
Authority Scope: Поведение конкретных возможностей продукта.

Owners:
  - Product Owner

Depends On:
  - features/001-feature-template.md
  - language/002-naming-rules.md

Required By:
  - (нет)

Defines:
  - Правила именования файлов фич

Must Not Define:
  - Описание фич
  - Реестр фич (принадлежит features/000)
---

# Каталог фич — правила каталога

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Объяснить правила наполнения директории: одна фича — один файл, имя вида F-XXX-slug.md, форма по шаблону.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`features/001-feature-template.md`](../../features/001-feature-template.md)
- [`language/002-naming-rules.md`](../../language/002-naming-rules.md)

**Зависит от этого документа (обновить при изменении):**

- (нет)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L6 · Features — Функциональность**
- Позиция в порядке чтения: **6.5**
- Полный порядок чтения: [`meta/001-reading-order.md`](../../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждый файл директории соответствует шаблону features/001-feature-template.md.
2. Каждый файл зарегистрирован в features/000-feature-catalog.md.
3. Имя файла соответствует id фичи.

Дополнительно применяются общие правила: [`meta/006-validation-rules.md`](../../meta/006-validation-rules.md)

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
