---
Document Id: META-READING-ORDER
Title: Порядок чтения документации
Layer: L0 · Meta — Система документации
Status: Approved
Version: 1.0.0
Last Updated: 2026-07-28
Reading Order: 0.2
Authority Rank: 5
Authority Scope: Порядок чтения. Не имеет власти над содержанием документов.

Owners:
  - Documentation Lead

Depends On:
  - meta/000-documentation-system.md

Required By:
  - domain/001-domain-reading-order.md
  - README.md

Defines:
  - Канонический порядок чтения документации
  - Маршруты чтения по ролям
  - Минимальный набор для конкретной задачи

Must Not Define:
  - Содержание документов
  - Иерархию авторитета (принадлежит meta/002)
---

# Порядок чтения документации

> Документов много. Читать их в произвольном порядке бессмысленно: нижние слои не имеют смысла без верхних.
>
> Этот документ задаёт **единственный канонический порядок** и короткие маршруты для тех, кому нужна часть.

---

## 1. Purpose · Назначение

Гарантировать, что любой читатель — человек или AI-агент — получает знание в том порядке, в котором оно было построено, и никогда не читает следствие раньше причины.

---

## 2. Owner · Владелец

- Documentation Lead

---

## 3. Dependencies · Зависимости

**Читать до этого документа:**

- [`meta/000-documentation-system.md`](000-documentation-system.md)

**Зависит от этого документа:**

- [`domain/001-domain-reading-order.md`](../domain/001-domain-reading-order.md)
- [`README.md`](../README.md)

---

## 4. Reading Order · Место в порядке чтения

Позиция **0.2** — второй документ репозитория.

---

## 5. Validation Rules · Правила валидации

1. Каждый документ репозитория присутствует ровно в одной позиции полного порядка (раздел 6.1).
2. Позиция документа в этом файле совпадает с полем `Reading Order` в самом документе.
3. Ни один документ не стоит раньше документа из своего поля `Depends On`.
4. Каждый маршрут раздела 6.3 является подмножеством полного порядка с сохранением последовательности.
5. При добавлении файла в репозиторий этот документ обновляется в том же изменении.

---

## 6. Content · Содержание

### 6.1. Полный порядок чтения

Читать сверху вниз. Это порядок построения знания, а не порядок важности.

| Позиция | Документ | Отвечает на вопрос |
|---------|----------|--------------------|
| **Project** | | **Ради чего и в каком порядке ведётся проект?** |
| 0.0 | `docs/project/project-constitution.md` | Конституция проекта — нормативный контекст |
| **0 · Meta** | | **Как устроена документация?** |
| 0.1 | `meta/000-documentation-system.md` | Как устроены слои |
| 0.2 | `meta/001-reading-order.md` | В каком порядке читать |
| 0.3 | `meta/002-authority-model.md` | Кто главнее при конфликте |
| 0.4 | `meta/003-document-template.md` | Как выглядит документ |
| 0.5 | `meta/004-anti-duplication-rules.md` | Как не дублировать знание |
| 0.6 | `meta/005-anti-contradiction-rules.md` | Как не допустить противоречий |
| 0.7 | `meta/006-validation-rules.md` | Как проверить документ |
| 0.8 | `meta/007-ownership-map.md` | Где живёт каждое понятие |
| 0.9 | `meta/008-lifecycle-and-versioning.md` | Как документ живёт и версионируется |
| 0.10 | `meta/009-contribution-workflow.md` | Как внести изменение |
| **0 · Standard** | | **Как документы производятся?** · читается перед первым авторством |
| 0.11 | `meta/standard/000-standard-overview.md` | Что такое стандарт |
| 0.12 | `meta/standard/001-document-lifecycle.md` | Состояния документа |
| 0.13 | `meta/standard/002-entry-criteria.md` | Можно ли создавать документ |
| 0.14 | `meta/standard/003-exit-criteria.md` | Можно ли считать готовым |
| 0.15 | `meta/standard/004-validation-checklist.md` | Как проверять |
| 0.16 | `meta/standard/005-cross-reference-checklist.md` | Как ссылаться |
| 0.17 | `meta/standard/006-traceability.md` | Откуда взято утверждение |
| 0.18 | `meta/standard/007-ownership-and-transfer.md` | Кто отвечает и как передаёт |
| 0.19 | `meta/standard/008-agent-change-protocol.md` | Как агент предлагает правки |
| 0.20 | `meta/standard/009-conflict-detection.md` | Как находится расхождение |
| 0.21 | `meta/standard/010-responsibility-budget.md` | Сколько документ может нести |
| 0.22 | `meta/standard/011-decomposition-rules.md` | Как делить и объединять |
| 0.23 | `meta/standard/012-diagram-rules.md` | Правила диаграмм |
| 0.24 | `meta/standard/013-example-rules.md` | Правила примеров |
| 0.25 | `meta/standard/014-terminology-rules.md` | Правила терминов |
| 0.26 | `meta/standard/015-migration-rules.md` | Как менять сами соглашения |
| 0.27 | `meta/standard/016-conformance-and-binding.md` | Как стандарт подключается |
| 0.28 | `meta/standard/017-authoring-pipeline.md` | Как документ производится |
| 0.29 | `meta/standard/018-document-class-profiles.md` | Чем различаются классы документов |
| **0 · Binding** | | **Что из этого действует здесь?** |
| 0.30 | `meta/010-migration-register.md` | Какие соглашения сейчас меняются |
| 0.31 | `meta/011-conformance-statement.md` | Чему репозиторий соответствует сегодня |
| 0.32 | `meta/012-authoring-binding.md` | Кто ведёт какую стадию здесь |
| **1 · Reality** | | **Как школа работает сегодня?** |
| 1.1 | `school/000-school-overview.md` | Что такое школа Belcanto |
| 1.2 | `school/001-education-model.md` | Как устроено обучение |
| 1.3 | `school/002-current-student-journey.md` | Путь ученика сегодня |
| 1.4 | `school/003-teacher-workflow.md` | Работа преподавателя сегодня |
| 1.5 | `school/004-current-admin-workflow.md` | Работа администратора сегодня |
| 1.6 | `school/005-current-owner-workflow.md` | Работа руководителя сегодня |
| 1.7 | `school/006-current-tooling-landscape.md` | Какими инструментами пользуются |
| 1.8 | `school/007-open-questions.md` | Что ещё не подтверждено |
| **2 · Product** | | **Зачем существует продукт?** |
| 2.1 | `product/000-product-overview.md` | Что такое продукт (главный документ) |
| 2.2 | `product/001-product-vision.md` | **Видение продукта** — корень цепочки вывода |
| 2.3 | `product/001-vision-and-principles.md` | Продуктовые принципы |
| 2.4 | `product/002-product-boundaries.md` | Границы продукта |
| 2.5 | `product/003-personas.md` | Для кого продукт |
| 2.6 | `product/004-core-concepts.md` | Ключевые концепции |
| 2.7 | `product/005-domain-glossary.md` | *(указатель на `language/`)* |
| 2.8 | `product/006-business-goals.md` | Бизнес-цели |
| 2.9 | `product/007-success-metrics.md` | Метрики успеха |
| 2.10 | `product/008-value-propositions.md` | Ценность по ролям |
| 2.11 | `product/009-product-rules.md` | Продуктовые правила |
| 2.12 | `product/010-risks-and-assumptions.md` | Риски и допущения |
| **3 · Language** | | **Что означает каждое слово?** |
| 3.1 | `language/000-language-overview.md` | Зачем единый язык |
| 3.2 | `language/001-ubiquitous-language.md` | Каталог терминов |
| 3.3 | `language/002-naming-rules.md` | Правила именования |
| 3.4 | `language/003-ui-terminology.md` | Слова в интерфейсе |
| 3.5 | `language/004-forbidden-terms.md` | Запрещённые слова |
| 3.6 | `language/005-localization-policy.md` | Языки продукта |
| **4 · Domain** | | **Как устроена предметная область?** |
| 4.1 | `domain/000-domain-overview.md` | Что такое домен |
| 4.2 | `domain/001-domain-reading-order.md` | Порядок чтения домена |
| 4.3 | `domain/architecture/000-domain-model-rules.md` | Правила моделирования |
| 4.4 | `domain/aggregates/000-aggregate-catalog.md` | Агрегаты |
| 4.5 | `domain/entities/000-entity-catalog.md` | Сущности |
| 4.6 | `domain/value-objects/000-value-object-catalog.md` | Объекты-значения |
| 4.7 | `domain/commands/000-domain-command-catalog.md` | Команды |
| 4.8 | `domain/events/000-domain-event-catalog.md` | События |
| 4.9 | `domain/services/000-domain-service-catalog.md` | Доменные сервисы |
| 4.10 | `domain/policies/000-domain-policy-overview.md` | Политики (далее — по каталогу) |
| **5 · Experience** | | **Как продукт устроен для пользователя?** |
| 5.1 | `experience/000-experience-overview.md` | Состав слоя |
| 5.2 | `experience/001-information-architecture.md` | Информационная архитектура |
| 5.3 | `experience/002-navigation-model.md` | Навигация |
| 5.4 | `experience/003-screen-catalog.md` | Каталог экранов |
| 5.5–5.8 | `experience/004..007-journey-*.md` | Целевые сценарии по ролям |
| 5.9 | `experience/008-states-and-empty-states.md` | Состояния |
| 5.10 | `experience/009-notification-experience.md` | Уведомления |
| 5.11 | `experience/010-interaction-rules.md` | Правила взаимодействия |
| **6 · Features** | | **Что продукт умеет?** |
| 6.1 | `features/000-feature-catalog.md` | Реестр возможностей |
| 6.2 | `features/001-feature-template.md` | Форма описания фичи |
| 6.3 | `features/002-feature-lifecycle.md` | Жизненный цикл фичи |
| 6.4 | `features/003-feature-dependency-map.md` | Зависимости фич |
| 6.5 | `features/catalog/F-*.md` | Отдельные фичи (по необходимости) |
| **7 · Design** | | **Как продукт выглядит и звучит?** |
| 7.1 | `design/000-design-overview.md` | Состав слоя |
| 7.2 | `design/001-design-principles.md` | Принципы дизайна |
| 7.3–7.6 | `design/002..005-*` | Цвет, типографика, компоновка, иконки |
| 7.7 | `design/006-component-language.md` | Компоненты |
| 7.8 | `design/007-motion-language.md` | Движение |
| 7.9 | `design/008-sound-and-haptics.md` | Звук |
| 7.10 | `design/009-accessibility.md` | Доступность |
| 7.11 | `design/010-content-and-voice.md` | Тон и текст |
| 7.12 | `design/011-design-tokens.md` | Токены |
| **8 · Delivery** | | **В каком порядке делаем?** |
| 8.1 | `roadmap/000-roadmap-overview.md` | Как планируем |
| 8.2 | `roadmap/001-mvp-scope.md` | Первая версия |
| 8.3 | `roadmap/002-release-plan.md` | Релизы |
| 8.4 | `roadmap/003-milestones.md` | Контрольные точки |
| 8.5 | `roadmap/004-backlog-and-out-of-scope.md` | Отложенное |
| **9 · Governance** | | **Кто и как решает?** |
| 9.1 | `governance/000-governance-overview.md` | Процесс решений |
| 9.2 | `governance/001-change-management.md` | Изменения |
| 9.3 | `governance/002-review-checklist.md` | Приёмка |
| 9.4 | `governance/decisions/000-decision-log.md` | Журнал решений |
| 9.5 | `governance/decisions/001-decision-template.md` | Форма решения |
| 9.6 | `governance/decisions/002-pending-decisions.md` | Отложенные решения (PD-XXXX) |
| 9.7 | `governance/decisions/records/000-records-readme.md` | Правила каталога записей решений |
| **10 · AI** | | **Как работает агент?** |
| 10.1 | `ai/000-ai-overview.md` | Границы AI-слоя |
| 10.2 | ~~`ai/001-agent-reading-protocol.md`~~ | **`Superseded`** (PD-0026) — история |
| 10.3 | ~~`ai/002-authority-resolution.md`~~ | **`Superseded`** (PD-0026) — история |
| 10.4 | ~~`ai/003-writing-rules.md`~~ | **`Superseded`** (PD-0026) — история |
| 10.5 | ~~`ai/004-task-playbooks.md`~~ | **`Superseded`** (PD-0026) — история |
| 10.6 | ~~`ai/005-prohibited-actions.md`~~ | **`Superseded`** (PD-0026) — история |
| 10.7 | `ai/006-ai-in-product-boundary.md` | AI внутри продукта |
| **11 · Assets** | | **Где материалы?** |
| 11.1 | `assets/000-assets-overview.md` | Состав материалов |
| 11.2+ | `assets/**/000-*.md` | Правила подкаталогов |

### 6.2. Правило порядка

**Документ нельзя читать раньше любого документа из его поля `Depends On`.**

Полный порядок выше уже удовлетворяет этому правилу. Любой сокращённый маршрут обязан его сохранять.

### 6.3. Маршруты чтения

Не всем нужно всё. Маршрут — это подмножество полного порядка **без нарушения последовательности**.

**Новый инженер (обязательный минимум, ~1 день)**

```
meta/000 → meta/002 → school/000 → school/001
→ product/000 → product/001 → product/002 → product/003
→ language/001
→ domain/000 → domain/001
→ experience/000 → experience/001 → experience/002 → experience/003
→ features/000
→ design/000 → design/009
→ roadmap/001
```

**Дизайнер**

```
product/000 → product/001 → product/003 → language/003
→ experience/001 → experience/002 → experience/003 → experience/008 → experience/010
→ design/ (весь слой)
```

**Преподаватель / методист (Education Lead)**

```
school/001 → product/000 → product/004 → language/001
→ domain/000 → domain/policies/ → experience/004 → experience/005
→ ai/006
```

**Руководитель**

```
product/000 → product/006 → product/007 → roadmap/ (весь слой)
→ governance/decisions/000
```

**AI-агент**

Не использует маршруты из этого раздела. Агент следует [`meta/standard/008-agent-change-protocol.md`](standard/008-agent-change-protocol.md), раздел 6.3, который задаёт обязательный набор чтения **от задачи**, а не от роли.

### 6.4. Минимальный набор для задачи

Для конкретной задачи читать не всё, а по формуле:

```
обязательный набор =
      meta/002-authority-model.md
    + документ-владелец затрагиваемого понятия   (см. meta/007-ownership-map.md)
    + всё из его поля Depends On
    + всё из его поля Required By
```

Поле `Required By` читается не для понимания, а чтобы знать, **что придётся обновить**.

---

## История изменений

| Версия | Дата | Изменение |
|--------|------|-----------|
| 1.0.0 | 2026-07-28 | Первая редакция порядка чтения. |
