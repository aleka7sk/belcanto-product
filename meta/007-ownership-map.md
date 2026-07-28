---
Document Id: META-OWNERSHIP
Title: Карта владения знанием
Layer: L0 · Meta — Система документации
Status: Approved
Version: 1.0.0
Last Updated: 2026-07-28
Reading Order: 0.8
Authority Rank: 5
Authority Scope: Реестр владельцев понятий. Не имеет власти над содержанием понятий.

Owners:
  - Documentation Lead

Depends On:
  - meta/004-anti-duplication-rules.md

Required By:
  - meta/006-validation-rules.md
  - ai/001-agent-reading-protocol.md
  - governance/002-review-checklist.md

Defines:
  - Реестр «понятие → документ-владелец»
  - Реестр «роль → зона ответственности»

Must Not Define:
  - Содержание понятий (принадлежит документам-владельцам)
---

# Карта владения знанием

> Один вопрос — один адрес. Этот документ отвечает: «Где живёт знание о X?»
>
> Прежде чем что-то писать, посмотри сюда. Если понятие уже имеет владельца — писать нужно туда.

---

## 1. Purpose · Назначение

Дать один поисковый указатель от понятия к документу-владельцу, чтобы автор не создавал вторую копию знания просто потому, что не нашёл первую.

---

## 2. Owner · Владелец

- Documentation Lead

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`meta/004-anti-duplication-rules.md`](004-anti-duplication-rules.md)

**Зависит от этого документа:**

- [`meta/006-validation-rules.md`](006-validation-rules.md)
- [`ai/001-agent-reading-protocol.md`](../ai/001-agent-reading-protocol.md)
- [`governance/002-review-checklist.md`](../governance/002-review-checklist.md)

---

## 4. Reading Order · Место в порядке чтения

Позиция **0.8**. Используется как справочник, а не читается целиком.

---

## 5. Validation Rules · Правила валидации

1. Каждая строка раздела 6.1 указывает ровно один документ-владелец.
2. Каждое понятие раздела 6.1 присутствует в поле `Defines` указанного документа.
3. Каждое понятие из поля `Defines` любого документа присутствует в разделе 6.1.
4. Реестр обновляется в том же изменении, что и появление нового понятия.

---

## 6. Content · Содержание

### 6.1. Реестр: понятие → документ-владелец

**Смысл и границы продукта**

| Понятие | Владелец |
|---------|----------|
| Что такое Belcanto Product, миссия, определение успеха | `product/000-product-overview.md` |
| Видение, продуктовые принципы | `product/001-vision-and-principles.md` |
| Границы продукта, что продуктом не является | `product/002-product-boundaries.md` |
| Роли и персоны | `product/003-personas.md` |
| Ключевые концепции продукта | `product/004-core-concepts.md` |
| Бизнес-цели (`BG-XX`) | `product/006-business-goals.md` |
| Метрики успеха | `product/007-success-metrics.md` |
| Ценность по ролям | `product/008-value-propositions.md` |
| Продуктовые правила (`PR-XX`) | `product/009-product-rules.md` |
| Риски и допущения | `product/010-risks-and-assumptions.md` |

**Реальность школы**

| Понятие | Владелец |
|---------|----------|
| Как устроена школа сегодня | `school/000-school-overview.md` |
| Образовательная модель | `school/001-education-model.md` |
| Текущие процессы ролей | `school/002…005-*` |
| Текущие инструменты школы | `school/006-current-tooling-landscape.md` |
| Неподтверждённые предположения | `school/007-open-questions.md` |

**Язык**

| Понятие | Владелец |
|---------|----------|
| Определение любого термина продукта | `language/001-ubiquitous-language.md` |
| Правила именования и идентификаторов | `language/002-naming-rules.md` |
| Подписи в интерфейсе | `language/003-ui-terminology.md` |
| Запрещённые слова | `language/004-forbidden-terms.md` |
| Языки продукта | `language/005-localization-policy.md` |

**Доменная модель**

| Понятие | Владелец |
|---------|----------|
| Правила моделирования домена | `domain/architecture/000-domain-model-rules.md` |
| Агрегаты | `domain/aggregates/000-aggregate-catalog.md` |
| Сущности | `domain/entities/000-entity-catalog.md` |
| Объекты-значения | `domain/value-objects/000-value-object-catalog.md` |
| Команды | `domain/commands/000-domain-command-catalog.md` |
| Доменные события | `domain/events/000-domain-event-catalog.md` |
| Доменные сервисы | `domain/services/000-domain-service-catalog.md` |
| Правила расчёта и реакции | `domain/policies/*` |

**Опыт**

| Понятие | Владелец |
|---------|----------|
| Иерархия разделов продукта | `experience/001-information-architecture.md` |
| Навигация, точки входа, переходы | `experience/002-navigation-model.md` |
| Экраны (`SCR-XXX`) | `experience/003-screen-catalog.md` |
| Целевые пути ролей | `experience/004…007-journey-*.md` |
| Состояния и пустые состояния | `experience/008-states-and-empty-states.md` |
| Опыт уведомлений | `experience/009-notification-experience.md` |
| Сквозные правила взаимодействия | `experience/010-interaction-rules.md` |

**Функциональность**

| Понятие | Владелец |
|---------|----------|
| Реестр фич (`F-XXX`) | `features/000-feature-catalog.md` |
| Форма описания фичи | `features/001-feature-template.md` |
| Статусы фичи | `features/002-feature-lifecycle.md` |
| Зависимости фич | `features/003-feature-dependency-map.md` |
| Поведение конкретной фичи | `features/catalog/F-XXX-*.md` |

**Форма**

| Понятие | Владелец |
|---------|----------|
| Принципы дизайна | `design/001-design-principles.md` |
| Смысл ролей цвета | `design/002-color-system.md` |
| Роли текста, шкала | `design/003-typography.md` |
| Отступы, сетка, компоновка | `design/004-spacing-and-layout.md` |
| Иконки | `design/005-iconography.md` |
| Компоненты интерфейса | `design/006-component-language.md` |
| Смысл движения | `design/007-motion-language.md` |
| Звук и тактильная отдача | `design/008-sound-and-haptics.md` |
| Требования доступности | `design/009-accessibility.md` |
| Тон и правила текста | `design/010-content-and-voice.md` |
| **Все конкретные значения формы** | `design/011-design-tokens.md` |

**Поставка, управление, AI, материалы**

| Понятие | Владелец |
|---------|----------|
| Принципы приоритизации | `roadmap/000-roadmap-overview.md` |
| Состав первой версии | `roadmap/001-mvp-scope.md` |
| Состав релизов | `roadmap/002-release-plan.md` |
| Контрольные точки | `roadmap/003-milestones.md` |
| Отложенное и вне границ | `roadmap/004-backlog-and-out-of-scope.md` |
| Процесс решений | `governance/000-governance-overview.md` |
| Процесс изменений и каскад | `governance/001-change-management.md` |
| Проверочный лист приёмки | `governance/002-review-checklist.md` |
| Решения (`PDR-XXXX`) | `governance/decisions/000-decision-log.md` |
| Протокол чтения агента | `ai/001-agent-reading-protocol.md` |
| Запреты агента | `ai/005-prohibited-actions.md` |
| Границы AI внутри продукта | `ai/006-ai-in-product-boundary.md` |
| Внешние источники | `assets/references/000-references.md` |
| Структура документации | `meta/000-documentation-system.md` |
| Порядок чтения | `meta/001-reading-order.md` |
| Авторитет и конфликты | `meta/002-authority-model.md` |

**Производство документации (стандарт `meta/standard/`)**

| Понятие | Владелец |
|---------|----------|
| Назначение стандарта, модальности, уровни соответствия | `meta/standard/000-standard-overview.md` |
| Состояния документа и переходы | `meta/standard/001-document-lifecycle.md` |
| Условия создания документа | `meta/standard/002-entry-criteria.md` |
| Условия завершённости | `meta/standard/003-exit-criteria.md` |
| Обязательные проверки | `meta/standard/004-validation-checklist.md` |
| Типы ссылок и правила ссылок | `meta/standard/005-cross-reference-checklist.md` |
| Прослеживаемость и классы утверждений | `meta/standard/006-traceability.md` |
| Владение документом и передача | `meta/standard/007-ownership-and-transfer.md` |
| Классы изменений агента, форма предложения, запреты | `meta/standard/008-agent-change-protocol.md` |
| Классы конфликтов и детекторы | `meta/standard/009-conflict-detection.md` |
| Бюджет ответственности документа | `meta/standard/010-responsibility-budget.md` |
| Стратегии и процедура разделения | `meta/standard/011-decomposition-rules.md` |
| Правила диаграмм | `meta/standard/012-diagram-rules.md` |
| Правила примеров | `meta/standard/013-example-rules.md` |
| Правила терминов (обращение, не состав) | `meta/standard/014-terminology-rules.md` |
| Классы и порядок миграций | `meta/standard/015-migration-rules.md` |
| Требования к привязке, уровни соответствия | `meta/standard/016-conformance-and-binding.md` |

**Привязка стандарта**

| Понятие | Владелец |
|---------|----------|
| Пороги сроков и интервалы ревизии | `meta/008-lifecycle-and-versioning.md` |
| Реестр миграций репозитория | `meta/010-migration-register.md` |
| Заявленный уровень соответствия, реестр отклонений | `meta/011-conformance-statement.md` |

### 6.2. Понятия с несколькими аспектами

Одно слово — несколько владельцев, по одному на каждый вопрос. Это не дублирование ([`meta/004`](004-anti-duplication-rules.md), правило D2).

| Понятие | Что означает слово | Как устроено в модели | Где пользователь видит | Как выглядит | Когда сделаем |
|---------|--------------------|------------------------|------------------------|--------------|---------------|
| Навык | `language/001` | `domain/` | `experience/003` | `design/006` | `roadmap/002` |
| Домашнее задание | `language/001` | `domain/homework.md` | `experience/003` | `design/006` | `roadmap/002` |
| Прогресс | `language/001` | `domain/progress.md`, `domain/policies/progress-update-policy.md` | `experience/003` | `design/006` | `roadmap/002` |
| Достижение | `language/001` | `domain/policies/achievement-award-policy.md` | `experience/003` | `design/006` | `roadmap/002` |
| Уведомление | `language/001` | `domain/policies/notification-policy.md` | `experience/009` | `design/010` | `roadmap/002` |
| Концерт | `language/001` | `domain/policies/concert-eligibility-policy.md` | `experience/003` | `design/006` | `roadmap/002` |

### 6.3. Роли и зоны ответственности

| Роль | Владеет слоями | Право принимать решения |
|------|----------------|-------------------------|
| Product Owner | `product/`, `roadmap/`, `governance/` | Все продуктовые решения |
| Education Lead | `school/`, `language/`, педагогическая часть `domain/` | Педагогические решения |
| Domain Architecture Lead | `domain/` | Решения о модели |
| Design Lead | `design/`, `experience/` совместно с Product Owner | Решения о форме |
| Technical Lead | техническая часть `domain/`, `assets/diagrams/` | Технические ограничения |
| Documentation Lead | `meta/` | Решения о форме документации |
| AI Lead | `ai/` | Правила работы агентов |

Роль — не человек. Один человек может исполнять несколько ролей; документ принадлежит роли.

### 6.4. Если владельца нет

Понятие без владельца — незавершённое состояние системы. Порядок действий:

1. Определить вопрос, на который отвечает понятие.
2. Найти слой по [`meta/000`](000-documentation-system.md), раздел 6.1.
3. Выбрать документ слоя; при отсутствии подходящего — [`meta/004`](004-anti-duplication-rules.md), правило D7.
4. Добавить понятие в `Defines` документа **и** в раздел 6.1 этого реестра — одним изменением.

---

## История изменений

| Версия | Дата | Изменение |
|--------|------|-----------|
| 1.0.0 | 2026-07-28 | Первая редакция карты владения. |
