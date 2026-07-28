---
Document Id: EXP-INTERACTION
Title: Правила взаимодействия
Layer: L5 · Experience — Опыт
Status: Scaffold
Version: 0.1.0
Last Updated: 2026-07-28
Reading Order: 5.11
Authority Rank: 40
Authority Scope: Структура информации, навигация, пользовательские сценарии.

Owners:
  - Design Lead
  - Product Owner

Depends On:
  - product/009-product-rules.md
  - experience/003-screen-catalog.md

Required By:
  - design/007-motion-language.md
  - features/001-feature-template.md

Defines:
  - Сквозные правила взаимодействия
  - Требования к обратимости и подтверждениям

Must Not Define:
  - Параметры анимации (принадлежит design/007)
  - Логику конкретных фич
---

# Правила взаимодействия

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Задать сквозные правила поведения интерфейса: подтверждения, обратимость, обратная связь, задержки, обработка ошибок.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Design Lead
- Product Owner

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`product/009-product-rules.md`](../product/009-product-rules.md)
- [`experience/003-screen-catalog.md`](../experience/003-screen-catalog.md)

**Зависит от этого документа (обновить при изменении):**

- [`design/007-motion-language.md`](../design/007-motion-language.md)
- [`features/001-feature-template.md`](../features/001-feature-template.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L5 · Experience — Опыт**
- Позиция в порядке чтения: **5.11**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждое правило проверяемо на конкретном экране.
2. Для каждого необратимого действия определено подтверждение.
3. Правила не противоречат product/009-product-rules.md.

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
