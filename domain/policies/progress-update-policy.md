---
Status: Accepted
Version: 1.0.0
Last Updated: 2026-07-27

Policy Id: PROGRESS_UPDATE_POLICY
Policy Type:
  - Reaction Policy
  - Calculation Policy
  - Recommendation Policy
  - Escalation Policy

Owners:
  - Product Owner
  - Education Lead
  - Technical Lead

Related Aggregates:
  - Progress
  - Student
  - Skill
  - Assessment
  - Lesson
  - Homework
  - Goal
  - Song
  - Concert

Observed Events:
  - AssessmentPublished
  - AssessmentSuperseded
  - AssessmentWithdrawn
  - LessonCompleted
  - HomeworkReviewed
  - GoalCompleted
  - GoalReopened
  - ConcertPerformanceAssessed
  - StudentSelfAssessmentAdded
  - TeacherProgressReviewCompleted
  - ProgressRecalculationRequested

Produced Commands:
  - CreateProgressReview
  - UpdateSkillProgress
  - UpdateGoalProgress
  - UpdateSongProgress
  - RecordProgressEvidence
  - MarkProgressConflict
  - RequestTeacherProgressReview
  - RecalculateProgressProjection
  - PublishStudentProgressSummary

Related Documents:
  - 000-domain-policy-overview.md
  - lesson-completion-policy.md
  - ../progress.md
  - ../assessment.md
  - ../lesson.md
  - ../homework.md
  - ../student.md
---

# Progress Update Policy

> Progress Update Policy определяет, когда новые образовательные факты могут изменить представление системы о развитии ученика.
>
> Политика не оценивает ценность или способности человека. Она формирует объяснимую, версионируемую и подтверждаемую картину изменений во времени.

---

# Purpose

В Belcanto Product регулярно появляются новые факты:

- преподаватель публикует Assessment;
- завершается Lesson;
- проверяется Homework;
- ученик выступает на Concert;
- достигается Goal;
- ученик добавляет Self Assessment;
- преподаватель проводит периодический обзор.

Не каждый новый факт означает изменение Progress.

Например:

- одно успешное упражнение не доказывает устойчивое развитие Skill;
- один неудачный Lesson не означает регресс;
- посещение занятия само по себе не означает улучшение;
- выполненное Homework не всегда подтверждает качество результата;
- мнение AI не является педагогическим фактом;
- самооценка ученика не должна автоматически менять подтвержденный Progress.

Progress Update Policy отделяет новые Evidence от обоснованных изменений Progress.

---

# Core Principle

Progress изменяется только тогда, когда доступная история Evidence позволяет сделать объяснимый вывод о значимом изменении.

```text
New Evidence
    |
    v
Policy Evaluation
    |
    +--> No Progress Change
    |
    +--> Evidence Recorded
    |
    +--> Progress Update Proposed
    |
    +--> Teacher Review Required
    |
    +--> Progress Updated
```

Новый факт всегда может быть сохранен как Evidence.

Но он не обязан менять текущее представление Progress.

---

## Progress Is Not a Score

Progress не должен сводиться к одному числу.

Запрещенная модель:

> Progress = 73%

без объяснения:

- что именно измеряется;
- как рассчитано значение;
- какие данные использовались;
- насколько вывод надежен;
- кто подтвердил интерпретацию;
- какая версия правил применена.

Progress должен описывать развитие по отдельным направлениям.

Например:

```text
Breathing Support:
  Trend: Improving
  Confidence: Medium

Rhythmic Accuracy:
  Trend: Stable
  Confidence: High

Stage Confidence:
  Trend: Needs Review
  Confidence: Low
```

---

# Progress Dimensions

Progress может отслеживаться по следующим объектам.

## Skill Progress

Изменение конкретного Skill.

Примеры:

- дыхание;
- интонация;
- дикция;
- ритм;
- артикуляция;
- резонанс;
- сценическая уверенность.

## Goal Progress

Продвижение к определенной Goal.

## Song Progress

Текущая стадия работы над Song.

## Concert Readiness

Динамика подготовки к выступлению.

Не является окончательным решением о допуске.

## Practice Progress

История самостоятельной практики и выполнения Homework.

Не должна автоматически интерпретироваться как рост вокального Skill.

## General Development Summary

Периодическое профессиональное резюме развития Student.

Формируется Teacher или подтверждается им.

---

# Trigger

Политика применяется при появлении нового Evidence или запросе на пересмотр Progress.

Основные события:

```text
AssessmentPublished
AssessmentSuperseded
AssessmentWithdrawn
LessonCompleted
HomeworkReviewed
GoalCompleted
GoalReopened
ConcertPerformanceAssessed
StudentSelfAssessmentAdded
TeacherProgressReviewCompleted
ProgressRecalculationRequested
```

Каждый Trigger должен содержать ссылку на источник, но не обязан передавать полный чувствительный контент.

---

# Inputs

Для применения политики могут потребоваться:

- StudentId;
- Progress Area;
- текущее состояние Progress;
- история Progress;
- новое Evidence;
- связанные Assessment;
- связанные Lesson;
- связанные Homework;
- связанные Goal;
- связанные Song;
- Concert results;
- Self Assessments;
- Teacher confirmations;
- даты и контексты наблюдений;
- авторы Evidence;
- Visibility;
- Confidence;
- версия политики;
- текущие права Actor;
- действующая образовательная программа.

Политика должна загружать только данные, относящиеся к оцениваемому направлению.

---

# Evidence Model

Каждый источник Progress должен быть представлен как Evidence Reference.

```text
ProgressEvidence
├── EvidenceId
├── EvidenceType
├── SourceId
├── StudentId
├── ProgressAreaId
├── AuthorId
├── AuthorRole
├── ObservedAt
├── RecordedAt
├── ContextType
├── Direction
├── Strength
├── Confidence
├── Visibility
├── IsHumanConfirmed
├── IsAIInvolved
└── SourceVersion
```

---

# Evidence Types

## Teacher Assessment Evidence

Наиболее значимый источник профессионального педагогического наблюдения.

## Lesson Result Evidence

Может использоваться, если содержит конкретное наблюдение о Progress Area.

Сам факт завершения Lesson не является доказательством роста.

## Homework Review Evidence

Может отражать:

- самостоятельность;
- стабильность;
- качество выполнения;
- перенос навыка в самостоятельную практику;
- регулярность.

## Concert Evidence

Отражает поведение навыка в условиях выступления.

Контекст Concert нельзя автоматически считать более важным, чем работа на Lesson.

## Self Assessment Evidence

Показывает восприятие развития самим Student.

Хранится как отдельный вид Evidence.

## Goal Evidence

Подтверждает выполнение конкретного критерия Goal.

## Practice Evidence

Показывает факт или регулярность практики.

Не доказывает автоматически улучшение качества исполнения.

## AI Observation Evidence

Может использоваться только как неподтвержденное вспомогательное Evidence.

Без Teacher confirmation не может самостоятельно изменить подтвержденный Progress.

---

# Evidence Direction

Evidence может указывать направление наблюдения.

Допустимые концептуальные значения:

- Positive;
- Neutral;
- Negative;
- Mixed;
- Unknown.

Negative не означает регресс автоматически.

Это описание отдельного наблюдения.

---

# Evidence Strength

Evidence может иметь относительную силу.

Допустимые значения:

- Weak;
- Moderate;
- Strong.

Сила определяется не эмоциональностью формулировки, а качеством основания.

Например:

- короткое наблюдение в одном упражнении — Weak;
- повторяющийся результат на нескольких Lesson — Moderate;
- устойчивый результат в разных контекстах — Strong.

---

# Progress State

Для каждого Progress Area рекомендуется хранить:

```text
ProgressState
├── ProgressAreaId
├── CurrentStage
├── Trend
├── Confidence
├── Summary
├── ActiveEvidenceReferences
├── ConflictingEvidenceReferences
├── ConfirmedBy
├── ConfirmedAt
├── PolicyVersion
├── EffectiveFrom
└── Version
```

---

# Trend

Допустимые концептуальные значения:

- Improving;
- Stable;
- Needs Attention;
- Declining;
- Insufficient Evidence;
- Conflicting Evidence;
- Not Recently Observed.

Использование Declining требует особенно строгого подтверждения.

Один отрицательный Assessment не должен автоматически создавать такой Trend.

---

# Confidence

Confidence показывает надежность текущего вывода.

Допустимые значения:

- Low;
- Medium;
- High.

Confidence не является оценкой уровня Student.

---

# Decision Outcomes

Политика возвращает один из следующих результатов.

## Evidence Recorded

Новое Evidence сохранено, но Progress не изменяется.

## No Change Required

Новое Evidence соответствует текущему состоянию и не требует новой версии Progress.

## Update Proposed

Политика сформировала предложение об изменении.

Требуется подтверждение Teacher, если правило не допускает автоматическое применение.

## Update Approved

Изменение может быть применено без дополнительного ручного решения в рамках утвержденной политики.

## Teacher Review Required

Недостаточно оснований для автоматического решения или существуют противоречия.

## Insufficient Evidence

Данных недостаточно для изменения.

## Conflicting Evidence

Источники дают существенные противоречивые выводы.

## Recalculation Required

Изменение или отзыв исторического Evidence требует пересчета.

## Rejected

Evidence не может использоваться в текущем Progress.

Причина обязательна.

---

# Decision Rules

## PU-001: Every progress conclusion requires Evidence

Progress не может быть создан или изменен без минимум одного Evidence Reference.

Reason Code: `PROGRESS_EVIDENCE_REQUIRED`

---

## PU-002: Lesson completion alone does not prove progress

Событие LessonCompleted не должно автоматически менять Skill Progress.

Оно может:

- добавить факт участия;
- зарегистрировать Lesson Result;
- запустить анализ связанных Assessment;
- обновить историю активности.

Reason Code: `LESSON_COMPLETION_NOT_PROGRESS_EVIDENCE`

---

## PU-003: Attendance is not skill development

Регулярное посещение может быть полезным показателем вовлеченности.

Однако Attendance не подтверждает улучшение вокального Skill.

Reason Code: `ATTENDANCE_NOT_SKILL_EVIDENCE`

---

## PU-004: One observation is normally insufficient for stable trend change

Одно наблюдение обычно не должно менять устойчивый Trend.

Исключения:

- диагностическое подтверждение очевидного нового состояния;
- значимое завершение Goal;
- подтвержденный результат Concert;
- решение Teacher в рамках Progress Review;
- исправление ошибочного текущего состояния.

Reason Code: `SINGLE_OBSERVATION_NOT_SUFFICIENT`

---

## PU-005: Multiple observations must be meaningfully independent

Несколько записей не считаются независимыми Evidence, если они:

- скопированы из одного источника;
- относятся к одному и тому же короткому наблюдению;
- созданы автоматически из одного события;
- дублируют друг друга;
- являются повторной публикацией одной версии.

Reason Code: `EVIDENCE_NOT_INDEPENDENT`

---

## PU-006: Recency matters

Недавнее Evidence обычно имеет большую актуальность, чем старое.

Однако старое Evidence не удаляется и не считается ошибочным.

Политика должна различать:

- историческое состояние;
- текущий вывод;
- длительную устойчивость;
- отсутствие недавних наблюдений.

Reason Code: `RECENT_EVIDENCE_REQUIRED`

---

## PU-007: Context diversity increases confidence

Evidence из разных контекстов может увеличивать Confidence.

Пример:

- Lesson;
- Homework;
- Rehearsal;
- Concert.

Но разнообразие контекста не является обязательным для каждого обновления.

---

## PU-008: Repeated performance increases confidence

Устойчивое повторение результата может увеличить Confidence без изменения Trend.

Пример:

```text
Trend: Improving
Confidence: Medium
```

после дополнительных подтверждений может стать:

```text
Trend: Improving
Confidence: High
```

---

## PU-009: Positive evidence does not always mean trend change

Новое положительное наблюдение может только подтвердить текущее состояние.

Результат: No Change Required

Reason Code: `EVIDENCE_CONFIRMS_CURRENT_STATE`

---

## PU-010: Negative evidence does not automatically mean decline

Единичная сложность может быть вызвана:

- новым материалом;
- усталостью;
- сложным упражнением;
- непривычным контекстом;
- техническими условиями;
- сценическим волнением.

Для Declining требуется устойчивая подтвержденная динамика или Teacher Review.

Reason Code: `NEGATIVE_OBSERVATION_NOT_DECLINE`

---

## PU-011: Conflicting assessments require explicit handling

Если подтвержденные Assessment дают несовместимые выводы, политика не должна молча усреднять их.

Результат: Teacher Review Required или Conflicting Evidence

Reason Code: `CONFLICTING_ASSESSMENTS`

---

## PU-012: Different contexts may legitimately produce different results

Разница между Lesson и Concert не всегда является конфликтом.

Пример:

Student стабильно выполняет навык на Lesson, но теряет его на сцене.

В этом случае могут существовать разные контекстные выводы:

```text
Technical Skill in Lesson Context: Stable
Stage Application: Needs Attention
```

---

## PU-013: Assessment must target the relevant progress area

Assessment о дикции не может использоваться для изменения Progress дыхания без явно указанной связи.

Reason Code: `EVIDENCE_AREA_MISMATCH`

---

## PU-014: Evidence author must be trusted for the context

Источник должен иметь право формировать соответствующее Evidence.

Например:

- Teacher может создавать педагогический Assessment;
- Student может создавать Self Assessment;
- AI может создавать только вспомогательное неподтвержденное наблюдение;
- Administrator не создает педагогическую оценку по умолчанию.

Reason Code: `EVIDENCE_AUTHOR_NOT_AUTHORIZED`

---

## PU-015: AI evidence requires human confirmation

AI-generated observation без подтверждения Teacher может быть сохранено как Draft или вспомогательное Evidence.

Оно не может самостоятельно:

- менять Trend;
- повышать Confidence до High;
- завершать Goal;
- определять Declining;
- формировать Student Visible conclusion.

Reason Code: `AI_EVIDENCE_REQUIRES_CONFIRMATION`

---

## PU-016: Self Assessment is preserved separately

Self Assessment не должен перезаписывать Teacher Assessment.

Он может:

- дополнить картину;
- выявить расхождение восприятия;
- инициировать обсуждение;
- повысить объяснимость Progress.

Reason Code: `SELF_ASSESSMENT_RECORDED_SEPARATELY`

---

## PU-017: Student disagreement does not erase professional evidence

Student может не согласиться с Progress Summary.

Несогласие должно быть сохранено.

Но оно не удаляет подтвержденное Teacher Evidence.

Возможные действия:

- добавить комментарий;
- запросить обсуждение;
- создать Self Assessment;
- инициировать Review.

---

## PU-018: Teacher review can confirm a progress change

Teacher Progress Review может быть достаточным основанием для обновления, если:

- Teacher имеет право оценивать Student;
- Review содержит Summary;
- указаны Evidence;
- определена область Progress;
- решение подтверждено явно.

Reason Code: `TEACHER_REVIEW_CONFIRMED`

---

## PU-019: Goal completion may affect related progress

Завершение Goal может быть сильным Evidence для связанного Progress Area.

Но Goal не должен автоматически менять несвязанные Skills.

Reason Code: `GOAL_EVIDENCE_APPLIED`

---

## PU-020: Homework completion alone is not enough

Статус HomeworkCompleted без Review подтверждает выполнение действия, но не качество результата.

Для изменения Skill Progress требуется:

- Teacher Review;
- подтвержденный результат;
- другое надежное Evidence.

Reason Code: `HOMEWORK_COMPLETION_NOT_QUALITY_EVIDENCE`

---

## PU-021: Concert evidence is context-sensitive

Успешное выступление может быть сильным Evidence.

Но оно не должно автоматически означать полное освоение всех связанных Skills.

Необходимо определить:

- что наблюдалось;
- какой Skill проявился;
- была ли стабильность;
- кто подтвердил результат;
- какие условия выступления существовали.

---

## PU-022: Withdrawn evidence must not remain active

После AssessmentWithdrawn соответствующее Evidence исключается из активного основания текущего Progress.

Историческая связь сохраняется.

Reason Code: `WITHDRAWN_EVIDENCE_EXCLUDED`

---

## PU-023: Superseded evidence must use the latest valid version

При AssessmentSuperseded старая версия остается в истории, но не должна одновременно учитываться как независимое активное Evidence.

Reason Code: `SUPERSEDED_EVIDENCE_REPLACED`

---

## PU-024: Evidence changes may require recalculation

Отзыв, исправление или замена Evidence может привести к:

- сохранению текущего Progress;
- снижению Confidence;
- изменению Trend;
- Teacher Review;
- созданию новой версии Progress.

Историческая версия не переписывается.

Reason Code: `PROGRESS_RECALCULATION_REQUIRED`

---

## PU-025: Progress updates are versioned

Любое значимое изменение создает новую версию Progress State.

Предыдущая версия остается доступной.

---

## PU-026: Historical progress must remain reproducible

Для каждой версии сохраняются:

- использованные Evidence;
- PolicyId;
- PolicyVersion;
- автор подтверждения;
- рассчитанный результат;
- время вступления в силу.

---

## PU-027: Policy changes do not silently rewrite history

Новая версия политики не должна автоматически заменять старые Progress State.

Пересчет должен быть явным и аудируемым.

Reason Code: `EXPLICIT_RECALCULATION_REQUIRED`

---

## PU-028: Progress summaries must be explainable

Student Visible Summary должен объяснять:

- что изменилось;
- на основании чего;
- за какой период;
- что можно делать дальше.

Недопустимо:

> Ваш прогресс снизился на 8%.

без контекста и объяснения.

---

## PU-029: Progress must not compare students

Политика не должна использовать относительное положение среди других учеников для определения индивидуального Progress.

Reason Code: `PEER_COMPARISON_NOT_ALLOWED`

---

## PU-030: Program expectations may contextualize progress

Учебная программа может определять ожидаемые этапы.

Но отклонение от среднего срока не должно автоматически считаться плохим Progress.

---

## PU-031: Long absence changes confidence, not history

После длительной паузы предыдущий подтвержденный Progress не удаляется.

Текущее состояние может получить:

```text
Trend: Not Recently Observed
Confidence: Low
```

до нового Assessment.

---

## PU-032: Progress cannot be updated by unauthorized clients

Mobile или web client не могут самостоятельно вычислять и сохранять доменный Progress.

Клиент может показывать предварительную визуализацию.

Финальное решение применяется backend.

---

## PU-033: Duplicate events must be idempotent

Повторная обработка одного Trigger не должна создавать повторные Progress versions.

Reason Code: `PROGRESS_TRIGGER_ALREADY_PROCESSED`

---

## PU-034: Concurrent updates require conflict resolution

Если одновременно обрабатываются несколько Evidence, система должна:

- применить определенный порядок;
- использовать optimistic concurrency;
- повторно оценить политику;
- не потерять ни одно Evidence.

Reason Code: `PROGRESS_VERSION_CONFLICT`

---

## PU-035: Sensitive evidence visibility must be preserved

Student Visible Progress может ссылаться только на Evidence, доступное Student, либо использовать безопасное обобщение.

Закрытый Teacher Assessment не должен раскрываться через текст Summary.

---

# Minimum Evidence Guidance

Конкретные пороги зависят от Progress Area и должны утверждаться отдельно.

Базовая рекомендация:

## Low Confidence Change

Может основываться на:

- одном подтвержденном Teacher Assessment;
- одном значимом Goal result;
- одном сильном Concert observation;
- одном Teacher Progress Review.

## Medium Confidence Change

Обычно требует:

- нескольких наблюдений;
- повторения результата;
- минимум двух временных точек;
- подтверждения Teacher.

## High Confidence Change

Обычно требует:

- устойчивой динамики во времени;
- нескольких независимых Evidence;
- повторения в сопоставимых или разных контекстах;
- отсутствия значимых необъясненных противоречий;
- подтверждения Teacher.

Это guidance, а не универсальная жесткая формула.

---

# Progress Update Flow

```text
Domain Event received
        |
        v
Validate trigger and source
        |
        v
Load current Progress State
        |
        v
Normalize Evidence reference
        |
        v
Check authorization and visibility
        |
        v
Evaluate relevance
        |
        +--> Reject Evidence
        |
        v
Record Evidence
        |
        v
Evaluate change rules
        |
        +--> No Change Required
        |
        +--> Insufficient Evidence
        |
        +--> Conflicting Evidence
        |
        +--> Teacher Review Required
        |
        +--> Update Proposed
        |
        v
Apply approved Progress update
        |
        v
ProgressUpdated
        |
        +--> Notification Policy
        |
        +--> Goal Completion Policy
        |
        +--> Achievement Award Policy
        |
        +--> Recommendation generation
```

---

# Commands Produced

## RecordProgressEvidence

Сохраняет ссылку на новое допустимое Evidence.

## UpdateSkillProgress

Обновляет состояние конкретного Skill.

## UpdateGoalProgress

Обновляет продвижение по Goal.

## UpdateSongProgress

Обновляет состояние работы над Song.

## CreateProgressReview

Создает периодический или ситуационный Review.

## MarkProgressConflict

Отмечает наличие противоречивых Evidence.

## RequestTeacherProgressReview

Запрашивает профессиональное решение Teacher.

## RecalculateProgressProjection

Пересчитывает аналитическое или прогнозное представление.

Не должно бесследно менять подтвержденный исторический Progress.

## PublishStudentProgressSummary

Публикует разрешенное резюме для Student после подтверждения.

---

# Domain Events

После успешных операций могут создаваться:

```text
ProgressEvidenceRecorded
ProgressUpdateProposed
ProgressUpdated
ProgressConfidenceChanged
ProgressTrendChanged
ProgressConflictDetected
ProgressReviewRequested
ProgressReviewCompleted
ProgressRecalculated
StudentProgressSummaryPublished
```

## ProgressUpdated Event

Событие должно содержать:

- StudentId;
- ProgressAreaId;
- PreviousStateReference;
- NewStateReference;
- Trend;
- Confidence;
- EffectiveFrom;
- EvidenceReferences;
- ConfirmedBy;
- PolicyId;
- PolicyVersion;
- CorrelationId;
- CausationId.

Оно не должно содержать полный текст закрытых Assessment без необходимости.

---

# Manual Review

Teacher Review требуется при:

- конфликтующих Assessment;
- возможном Declining;
- значительном изменении на основании ограниченных данных;
- расхождении AI и Teacher observation;
- отсутствии наблюдений после долгой паузы;
- изменении методики;
- спорном Goal completion;
- жалобе Student на некорректный Summary;
- необходимости объединить Evidence разных Teachers.

## Review Result

Teacher может:

- подтвердить предложенное изменение;
- отклонить изменение;
- изменить Trend;
- изменить Confidence;
- запросить дополнительные наблюдения;
- разделить Progress по контекстам;
- исправить связь Evidence;
- создать новое Assessment.

Решение должно иметь объяснение.

---

# Multi-Teacher Scenarios

Если Student работает с несколькими Teachers:

- каждый Assessment сохраняет авторство;
- различия мнений не усредняются автоматически;
- текущий ответственный Teacher не должен бесследно отменять чужое Evidence;
- спорные выводы требуют Review;
- права просмотра зависят от Scope;
- возможно разделение Progress по специализациям.

---

# Progress and Goals

Goal Progress и Skill Progress связаны, но не идентичны.

Пример:

Goal:

> Уверенно исполнить песню на ближайшем концерте.

Связанные Progress Areas:

- Intonation;
- Rhythm;
- Stage Confidence;
- Song Readiness.

Завершение Goal не означает автоматического полного освоения каждого Skill.

---

# Progress and Song

Song Progress может иметь собственные стадии.

Например:

```text
Selected
Learning
Technical Work
Interpretation
Memorized
Rehearsal Ready
Performance Ready
Performed
Maintained
```

Переходы должны регулироваться отдельной Song Readiness Policy.

---

# Progress and Achievements

Achievement может быть реакцией на подтвержденный Progress.

Однако Achievement Award Policy должна отдельно проверить:

- существует ли реальный факт;
- не выдавалось ли Achievement ранее;
- соответствует ли оно философии продукта;
- не превращает ли Progress в соревнование.

---

# Notification Effects

Progress Update Policy не отправляет уведомления напрямую.

Notification Policy может решить:

- уведомить Student о значимом изменении;
- не уведомлять о техническом пересчете;
- объединить обновление с Teacher feedback;
- скрыть чувствительный Trend до обсуждения;
- уведомить Teacher о необходимости Review.

---

# Student Presentation

Student должен видеть Progress в поддерживающей форме.

Рекомендуется показывать:

- что стало получаться лучше;
- подтверждающие примеры;
- текущий фокус;
- следующий понятный шаг;
- историю развития;
- уровень уверенности только при необходимости.

Следует избегать:

- красных предупреждений без объяснения;
- унизительных формулировок;
- сравнения с другими;
- необъяснимых процентов;
- категоричных прогнозов;
- псевдонаучной точности.

---

# Teacher Presentation

Teacher должен видеть:

- полный Evidence timeline;
- фильтрацию по Skill;
- контексты;
- авторов;
- Confidence;
- конфликтующие наблюдения;
- историю версий;
- причины изменения;
- AI involvement;
- запросы на Review.

---

# Owner Analytics

Owner может видеть агрегированные тенденции:

- частоту Progress Review;
- полноту Evidence;
- области без наблюдений;
- динамику по программам;
- качество заполнения;
- длительность между значимыми обновлениями.

Owner Analytics не должна:

- создавать рейтинги учеников;
- автоматически оценивать качество Teacher только по Progress;
- раскрывать закрытые заметки без основания;
- смешивать образовательную и финансовую ценность.

---

# AI Assistance

AI может:

- группировать Evidence;
- находить возможные повторяющиеся наблюдения;
- выявлять противоречия;
- предлагать Trend;
- создавать Draft Summary;
- предлагать вопросы для Teacher Review;
- находить отсутствующие наблюдения;
- объяснять динамику простым языком.

AI не может:

- самостоятельно подтверждать Progress;
- присваивать Declining без Teacher;
- скрывать неопределенность;
- создавать Evidence из отсутствующих фактов;
- считать регулярность практики доказательством качества;
- сравнивать Student с другими;
- публиковать чувствительный Summary без проверки.

AI output должен содержать:

- источник;
- модель или версию механизма;
- Confidence;
- входные Evidence References;
- время;
- факт Teacher confirmation.

---

# Privacy

Progress содержит чувствительную образовательную историю.

Необходимо обеспечить:

- ролевой доступ;
- разграничение Student Visible и Staff Only данных;
- аудит просмотра;
- защиту экспортов;
- минимизацию содержимого событий;
- сохранение Visibility исходного Evidence;
- контроль доступа при смене Teacher;
- правила работы с несовершеннолетними, если применимо.

---

# Audit Requirements

Для каждой оценки политики сохраняются:

- PolicyId;
- PolicyVersion;
- TriggerEventId;
- StudentId;
- ProgressAreaId;
- current Progress version;
- Evidence References;
- ActorId;
- Decision;
- Reason Codes;
- proposed state;
- final state;
- Teacher confirmation;
- AI metadata;
- EvaluatedAt;
- CorrelationId;
- CausationId.

Для ручного Review дополнительно:

- ReviewerId;
- ReviewedAt;
- Review Summary;
- accepted and rejected Evidence;
- explanation;
- resulting Progress version.

---

# Failure Modes

## Evidence source not found

- Decision: Rejected
- Reason Code: EVIDENCE_SOURCE_NOT_FOUND

## Evidence belongs to another Student

- Decision: Rejected
- Reason Code: EVIDENCE_STUDENT_MISMATCH

Security Audit обязателен.

## Evidence is unrelated to Progress Area

- Decision: Rejected
- Reason Code: EVIDENCE_AREA_MISMATCH

## Single weak observation

- Decision: Evidence Recorded
- Reason Code: SINGLE_OBSERVATION_NOT_SUFFICIENT

## Conflicting Teacher Assessments

- Decision: Teacher Review Required
- Reason Code: CONFLICTING_ASSESSMENTS

## AI-only evidence

- Decision: Evidence Recorded
- Reason Code: AI_EVIDENCE_REQUIRES_CONFIRMATION

Progress не меняется.

## Withdrawn Assessment affected current state

- Decision: Recalculation Required
- Reason Code: WITHDRAWN_EVIDENCE_EXCLUDED

## Duplicate trigger

- Decision: No Change Required
- Reason Code: PROGRESS_TRIGGER_ALREADY_PROCESSED

## Version conflict

- Decision: Deferred
- Reason Code: PROGRESS_VERSION_CONFLICT

Политика повторно оценивается на актуальном состоянии.

---

# Explainability Examples

## Improving

> За последние четыре занятия преподаватель несколько раз отметил более стабильную опору в длинных фразах. Результат также сохранился при самостоятельном выполнении задания.

## Stable

> Навык сохраняется на текущем уровне в нескольких последних наблюдениях. Следующий фокус — перенос результата на более сложный репертуар.

## Needs Attention

> В последних наблюдениях результат был нестабильным в быстром темпе. Это не означает потерю навыка: преподаватель рекомендует продолжить работу в более медленном темпе.

## Insufficient Evidence

> Пока недостаточно новых наблюдений, чтобы определить динамику. Прогресс будет уточнен после следующих занятий.

## Conflicting Evidence

> Результат проявляется по-разному в зависимости от контекста. Преподавателю предложено провести дополнительный обзор.

---

# Examples

## Example 1: One positive Assessment

Дано:

- Published Assessment;
- связан с Skill;
- автор — Teacher;
- наблюдение положительное;
- это первое Evidence.

Результат:

- Decision: Evidence Recorded
- Reason Code: SINGLE_OBSERVATION_NOT_SUFFICIENT
- Progress Change: None

## Example 2: Repeated positive evidence

Дано:

- три подтвержденных Assessment;
- разные даты;
- один Skill;
- наблюдение повторяется;
- значимых конфликтов нет.

Результат:

- Decision: Update Proposed
- Proposed Trend: Improving
- Proposed Confidence: Medium
- Reason Code: REPEATED_CONFIRMED_EVIDENCE

В зависимости от политики требуется Teacher confirmation.

## Example 3: Homework completed without review

Дано:

- Student отметил Homework как Completed;
- Teacher еще не проверил;
- Homework связан с Skill.

Результат:

- Decision: Evidence Recorded
- Reason Code: HOMEWORK_COMPLETION_NOT_QUALITY_EVIDENCE
- Progress Change: None

## Example 4: Concert difficulty

Дано:

- на Lesson Skill стабилен;
- на Concert возникли трудности;
- Teacher подтвердил наблюдение.

Результат:

- Decision: Update Proposed
- Action: Create context-specific progress state
- Lesson Context: Stable
- Stage Context: Needs Attention
- Reason Code: CONTEXT_SPECIFIC_VARIATION

## Example 5: Conflicting Teachers

Дано:

- два Teachers;
- один Assessment указывает Improving;
- другой — Declining;
- контексты сопоставимы;
- оба Evidence подтверждены.

Результат:

- Decision: Teacher Review Required
- Reason Code: CONFLICTING_ASSESSMENTS

## Example 6: Long absence

Дано:

- последний Assessment был восемь месяцев назад;
- новых Evidence нет.

Результат:

- Decision: Update Approved
- Trend: Not Recently Observed
- Confidence: Low
- Reason Code: CURRENT_STATE_NOT_RECENTLY_OBSERVED

Исторический Progress сохраняется.

## Example 7: Assessment withdrawn

Дано:

- текущий Trend использовал Assessment;
- Assessment отозван как ошибочный.

Результат:

- Decision: Recalculation Required
- Reason Code: WITHDRAWN_EVIDENCE_EXCLUDED

---

# Test Requirements

## Evidence Tests

- valid Teacher Assessment is recorded;
- unrelated Assessment is rejected;
- withdrawn Assessment is excluded;
- superseded Assessment does not count twice;
- AI-only Evidence cannot update Progress;
- Self Assessment is stored separately;
- duplicate Evidence is idempotent.

## Trend Tests

- one positive observation does not create Improving;
- repeated confirmed observations can propose Improving;
- one negative observation does not create Declining;
- conflicting observations require Review;
- old Evidence lowers current confidence when no recent data exists;
- stable repeated Evidence may increase Confidence.

## Context Tests

- Lesson and Concert contexts can produce separate conclusions;
- context differences are not automatically conflicts;
- Evidence from wrong context does not update unrelated state.

## Permission Tests

- authorized Teacher can confirm update;
- unrelated Teacher cannot confirm;
- Student cannot publish professional Progress;
- Administrator cannot create pedagogical conclusion;
- AI cannot act as confirmer.

## Versioning Tests

- every update creates a new version;
- historical versions remain reproducible;
- policy version is stored;
- recalculation does not overwrite history;
- stale version causes retry;
- concurrent Evidence is not lost.

## Explainability Tests

- every Progress change has Reason Codes;
- every conclusion references Evidence;
- Student Summary does not expose private notes;
- No Change decision is explainable;
- Conflicting Evidence result is explainable.

## Privacy Tests

- Student sees only permitted Evidence;
- owner analytics is aggregated;
- events exclude private text;
- previous Teacher access respects scope;
- private Assessment is not leaked through Summary.

## AI Tests

- AI suggestion remains Draft;
- Teacher confirmation is required;
- hallucinated source reference is rejected;
- AI Confidence is not treated as domain Confidence;
- AI cannot create Declining automatically.

---

# Non-Goals

Progress Update Policy не определяет:

- расписание;
- оплату;
- CRM-статусы;
- заработную плату Teacher;
- публичные рейтинги;
- медицинские выводы;
- талант или потенциал человека;
- окончательную готовность к Concert;
- выдачу Achievement;
- содержание учебной программы;
- автоматический выбор репертуара.

---

# Open Questions

Необходимо определить:

- официальный перечень Progress Areas;
- допустимые Trend для первой версии;
- формальные значения Confidence;
- когда Teacher confirmation обязательно;
- можно ли автоматически повышать Confidence;
- минимальный набор Evidence для разных Skills;
- как учитывать Evidence разных Teachers;
- как часто проводить Periodic Review;
- какие Progress данные видит Student;
- показывать ли Student Reason Codes;
- как отображать Declining или Needs Attention;
- нужен ли Progress по отдельным контекстам;
- как учитывать групповые Lesson;
- как связывать Goal и Skill Progress;
- как обрабатывать смену образовательной программы;
- какие части Progress могут рассчитываться автоматически;
- требуется ли отдельный Progress Aggregate на каждый Skill;
- как хранить исторические Policy Decisions;
- когда пересчитывать старые данные после изменения методики;
- как обрабатывать удаление вложения, использованного как Evidence;
- допускается ли подтверждение несколькими Teachers;
- как объяснять отсутствие Progress без демотивации Student.

---

# History

| Version | Description |
| --- | --- |
| 1.0.0 | Определены правила использования Evidence и изменения Progress. |
