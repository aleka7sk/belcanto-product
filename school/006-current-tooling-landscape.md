---
Document Id: SCH-TOOLING
Title: Текущий инструментальный ландшафт школы
Layer: L1 · Reality — Школа Belcanto
Status: Scaffold
Version: 0.1.1
Last Updated: 2026-07-29
Reading Order: 1.7
Authority Rank: 90
Authority Scope: Описание реальности школы. Документы описательные, а не нормативные.

Owners:
  - Product Owner
  - Technical Lead

Depends On:
  - school/000-school-overview.md
  - school/003-teacher-workflow.md
  - school/004-current-admin-workflow.md

Required By:
  - product/002-product-boundaries.md
  - roadmap/001-mvp-scope.md

Defines:
  - Перечень используемых инструментов
  - Где сегодня хранятся данные
  - Точки ручного переноса данных

Must Not Define:
  - Целевое состояние продукта
  - Технические решения по интеграциям
---

# Текущий инструментальный ландшафт школы

> **Статус: SCAFFOLD.**
> Документ создан как часть архитектуры документации Belcanto Product.
> Содержание ещё не написано. Разделы 1–5 являются контрактом документа и уже действуют.
> Раздел 6 заполняет владелец документа.

---

## 1. Purpose · Назначение

Зафиксировать, какими инструментами школа пользуется сегодня (таблицы, мессенджеры, журналы, платежи) и какие данные в них живут — чтобы продукт знал, что он заменяет, а что оставляет вовне.

---

## 2. Owner · Владелец

Ответственный за содержание и актуальность:

- Product Owner
- Technical Lead

Изменение этого документа без участия владельца недопустимо.

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`school/000-school-overview.md`](../school/000-school-overview.md)
- [`school/003-teacher-workflow.md`](../school/003-teacher-workflow.md)
- [`school/004-current-admin-workflow.md`](../school/004-current-admin-workflow.md)

**Зависит от этого документа (обновить при изменении):**

- [`product/002-product-boundaries.md`](../product/002-product-boundaries.md)
- [`roadmap/001-mvp-scope.md`](../roadmap/001-mvp-scope.md)

---

## 4. Reading Order · Место в порядке чтения

- Слой: **L1 · Reality — Школа Belcanto**
- Позиция в порядке чтения: **1.7**
- Полный порядок чтения: [`meta/001-reading-order.md`](../meta/001-reading-order.md)
- Правило авторитета: [`meta/002-authority-model.md`](../meta/002-authority-model.md)

---

## 5. Validation Rules · Правила валидации

Документ считается валидным, если выполнены все условия:

1. Каждый инструмент помечен: остаётся / заменяется / вне границ продукта.
2. Неподтверждённые сведения помечены как гипотеза.
3. Ни один пункт не описывает будущее — только текущее состояние.

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
| 0.1.1 | 2026-07-29 | Позиция чтения приведена к реестру-владельцу `meta/001` (1.1 → 1.7). Содержание не изменялось. |
| 0.1.0 | 2026-07-28 | Создан скелет документа. |
