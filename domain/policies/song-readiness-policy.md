---
Status: Accepted
Version: 1.0.0
Last Updated: 2026-07-27

Policy Id: SONG_READINESS_POLICY

Policy Type:
  - Eligibility Policy
  - Calculation Policy
  - Validation Policy
  - Recommendation Policy

Owners:
  - Product Owner
  - Education Lead
  - Technical Lead

Related Aggregates:
  - Song
  - Repertoire
  - Student
  - Teacher
  - Skill
  - Progress
  - Assessment
  - Lesson
  - Homework
  - Goal
  - Concert

Observed Events:
  - SongAddedToRepertoire
  - SongWorkStarted
  - LessonCompleted
  - HomeworkReviewed
  - AssessmentPublished
  - AssessmentSuperseded
  - AssessmentWithdrawn
  - ProgressUpdated
  - SongReadinessReviewRequested
  - TeacherSongReviewCompleted
  - ConcertParticipationProposed
  - SongPerformed

Produced Commands:
  - UpdateSongReadiness
  - RecordSongReadinessEvidence
  - RequestTeacherSongReview
  - MarkSongReadinessConflict
  - ProposeSongForConcert
  - ReturnSongToPreviousStage
  - ArchiveSongFromActiveRepertoire

Related Documents:
  - 000-domain-policy-overview.md
  - lesson-completion-policy.md
  - progress-update-policy.md
  - goal-completion-policy.md
  - achievement-award-policy.md
  - ../assessment.md
  - ../progress.md
  - ../lesson.md
  - ../homework.md
  - ../student.md
---

# Song Readiness Policy

> Song Readiness Policy определяет текущую стадию готовности произведения в репертуаре конкретного ученика.
>
> Политика показывает, насколько песня освоена в заданном контексте, но не принимает окончательное решение о допуске ученика к Concert.

---

# Purpose

Работа над песней проходит через несколько этапов.

Ученик может:

- только выбрать произведение;
- знакомиться с мелодией и текстом;
- разбирать технические сложности;
- работать над интерпретацией;
- исполнять песню на уроке;
- готовить ее к сцене;
- уже выступить с ней;
- поддерживать произведение в активном репертуаре.

Без формализованной модели разные пользователи могут понимать готовность по-разному.

Например:

- Student считает песню готовой, потому что знает слова;
- Teacher считает ее готовой для исполнения на Lesson;
- Administrator видит, что песня указана в заявке на Concert;
- Owner видит статус `Ready`, но не понимает, готова ли она технически или сценически.

Song Readiness Policy создает единый язык для описания состояния произведения.

---

# Core Principle

Готовность принадлежит не Song как музыкальному произведению вообще, а связи:

`Student + Song + Learning Context`

Одна и та же Song может:

- быть полностью готовой у одного Student;
- находиться на этапе разбора у другого;
- быть готовой для Lesson, но не для Concert;
- ранее исполняться, но требовать восстановления.

---

## Song Readiness Is Contextual

Запрещено использовать один универсальный флаг:

`IsReady = true`

без уточнения контекста.

Необходимо различать как минимум:

- Learning Readiness;
- Technical Readiness;
- Memory Readiness;
- Interpretation Readiness;
- Rehearsal Readiness;
- Performance Readiness;
- Maintenance Readiness.

---

# Readiness State

Рекомендуемая концептуальная структура:

```text
SongReadinessState
├── SongReadinessId
├── StudentId
├── SongId
├── RepertoireEntryId
├── CurrentStage
├── Context
├── Confidence
├── Summary
├── Strengths
├── BlockingAreas
├── EvidenceReferences
├── ConflictingEvidenceReferences
├── ConfirmedBy
├── ConfirmedAt
├── EffectiveFrom
├── PolicyId
├── PolicyVersion
└── Version
```

---

# Song Readiness Stages

Базовый жизненный путь:

```text
Selected
    |
    v
Exploring
    |
    v
Learning
    |
    v
Technical Work
    |
    v
Interpretation
    |
    v
Memorized
    |
    v
Rehearsal Ready
    |
    v
Performance Ready
    |
    v
Performed
    |
    v
Maintained
```

Этот путь не является строго линейным.

Song может возвращаться на предыдущий этап.

## Selected

Произведение добавлено в репертуар Student.

Это означает только намерение работать над ним.

На этом этапе:

- работа может еще не начаться;
- тональность может быть не утверждена;
- версия или аранжировка может быть не выбрана;
- педагогическая пригодность может еще проверяться.

## Exploring

Student и Teacher знакомятся с произведением.

Может происходить:

- прослушивание;
- определение диапазона;
- предварительная проверка сложности;
- выбор тональности;
- обсуждение смысла;
- выбор версии;
- решение о продолжении работы.

Переход в Exploring не означает окончательное утверждение Song.

## Learning

Student изучает основную структуру произведения.

Обычно включает:

- мелодию;
- текст;
- ритм;
- форму;
- вступления;
- окончания;
- основные музыкальные ориентиры.

## Technical Work

Основная структура известна, но требуется работа над техническим исполнением.

Например:

- дыхание;
- интонация;
- регистровые переходы;
- дикция;
- ритмическая точность;
- сложные интервалы;
- длинные фразы;
- динамика;
- атака звука.

## Interpretation

Student работает над художественным содержанием.

Например:

- эмоция;
- смысл текста;
- фразировка;
- динамический план;
- образ;
- сценическая подача;
- индивидуальная трактовка.

Техническая и интерпретационная работа могут идти параллельно.

## Memorized

Student может исполнить произведение без зависимости от текста или постоянных подсказок в утвержденном объеме.

Memorized не означает:

- техническую готовность;
- сценическую готовность;
- устойчивость под давлением;
- готовность к Concert.

## Rehearsal Ready

Song достаточно подготовлена для полноценной репетиции в целевом формате.

Например:

- с микрофоном;
- под минус;
- с концертмейстером;
- с ансамблем;
- с движением;
- на сценической площадке.

На этом этапе могут выявиться новые проблемы.

## Performance Ready

Teacher подтвердил, что Song может рассматриваться для публичного исполнения в определенном контексте.

Это не означает автоматический допуск к конкретному Concert.

Concert Eligibility Policy отдельно учитывает:

- формат мероприятия;
- длительность программы;
- состав участников;
- организационные ограничения;
- актуальность готовности;
- решение ответственного Teacher;
- состояние Student;
- требования конкретной площадки.

## Performed

Student исполнил Song в подтвержденном performance context.

Например:

- Concert;
- открытый урок;
- отчетное выступление;
- студийная запись;
- внутренний показ.

Сам факт исполнения не означает, что результат был успешным или что Song остается Performance Ready.

## Maintained

Song ранее была подготовлена или исполнена и сейчас поддерживается в активном состоянии.

Для сохранения этого статуса могут требоваться:

- периодическое повторение;
- недавняя репетиция;
- актуальный Assessment;
- подтверждение Teacher.

---

# Additional States

При необходимости могут использоваться:

- Paused
- Dropped
- Needs Rework
- Not Suitable
- Archived

## Paused

Работа временно приостановлена.

Причины могут включать:

- смену приоритета;
- подготовку другого произведения;
- паузу Student;
- отсутствие подходящего материала;
- временные организационные ограничения.

## Dropped

Решено прекратить работу над Song.

История сохраняется.

Dropped не должно восприниматься как неудача.

## Needs Rework

Song ранее находилась на более высокой стадии, но теперь требует существенного возвращения к работе.

## Not Suitable

Teacher определил, что текущая версия Song не подходит Student в данном контексте.

Причина должна быть профессионально и уважительно сформулирована.

Возможны альтернативы:

- другая тональность;
- сокращенная версия;
- другая аранжировка;
- возвращение позже;
- выбор другой Song.

## Archived

Repertoire Entry исключена из активной работы, но остается в истории.

---

# Context

Readiness всегда оценивается в контексте.

Примеры:

- Regular Lesson;
- Group Lesson;
- Home Practice;
- Rehearsal;
- Studio Recording;
- Internal Performance;
- Public Concert;
- Competition;
- Online Performance.

Song может иметь разные состояния по разным контекстам.

Пример:

```text
Regular Lesson:
  Stage: Performance Ready

Public Concert:
  Stage: Rehearsal Ready
```

---

# Trigger

Политика применяется при:

```text
SongAddedToRepertoire
SongWorkStarted
LessonCompleted
HomeworkReviewed
AssessmentPublished
AssessmentSuperseded
AssessmentWithdrawn
ProgressUpdated
SongReadinessReviewRequested
TeacherSongReviewCompleted
ConcertParticipationProposed
SongPerformed
```

Не каждое событие приводит к изменению Stage.

---

# Inputs

Для оценки могут потребоваться:

- StudentId;
- SongId;
- Repertoire Entry;
- текущая Readiness State;
- контекст;
- выбранная версия Song;
- тональность;
- аранжировка;
- Lesson Results;
- Homework Reviews;
- Assessments;
- Skill Progress;
- Goal Progress;
- rehearsal history;
- performance history;
- Teacher confirmations;
- текстовая и техническая подготовка;
- актуальность Evidence;
- Actor;
- Policy Version.

---

# Evidence

Song Readiness может опираться на:

- Teacher Assessment;
- Lesson Result;
- Homework Review;
- Rehearsal Assessment;
- Concert Assessment;
- Recording Review;
- Goal Completion;
- Skill Progress;
- Student Self Assessment;
- Teacher Song Review.

---

# Evidence Structure

```text
SongReadinessEvidence
├── EvidenceId
├── EvidenceType
├── SourceId
├── StudentId
├── SongId
├── RepertoireEntryId
├── Context
├── TargetStage
├── ObservedAreas
├── Strength
├── Confidence
├── AuthorId
├── AuthorRole
├── ObservedAt
├── Visibility
├── IsHumanConfirmed
├── IsAIInvolved
└── SourceVersion
```

---

# Readiness Areas

Готовность может оцениваться по нескольким направлениям.

## Material Knowledge

- мелодия;
- текст;
- структура;
- вступления;
- окончания;
- переходы.

## Technical Stability

- интонация;
- дыхание;
- ритм;
- дикция;
- регистры;
- диапазон;
- устойчивость сложных фрагментов.

## Interpretation

- понимание текста;
- образ;
- эмоция;
- фразировка;
- динамика;
- индивидуальность исполнения.

## Performance Stability

- исполнение целиком;
- восстановление после ошибки;
- работа с микрофоном;
- сохранение результата под давлением;
- устойчивость в непривычной среде.

## Format Readiness

- подходящий минус;
- утвержденная тональность;
- готовая аранжировка;
- согласованная длительность;
- технические требования;
- вступление и завершение;
- сценическая постановка.

Format Readiness частично относится к Concert context и не должна смешиваться с вокальной оценкой.

---

# Confidence

Допустимые значения:

- Low;
- Medium;
- High.

Confidence показывает надежность вывода, а не качество исполнения.

---

# Decision Outcomes

## Evidence Recorded

Evidence сохранено, Stage не меняется.

## No Change Required

Новые данные подтверждают текущее состояние.

## Stage Advancement Proposed

Предлагается переход на следующий этап.

## Stage Advancement Approved

Переход может быть применен согласно утвержденным правилам.

## Stage Regression Proposed

Появились основания вернуться на предыдущий этап.

Требует осторожного объяснения.

## Teacher Review Required

Необходимо профессиональное решение Teacher.

## Conflicting Evidence

Разные источники дают несовместимые выводы.

## Insufficient Evidence

Недостаточно данных для изменения.

## Context Split Required

Необходимо разделить Readiness по разным контекстам.

## Rejected

Evidence или запрос недействительны.

---

# Decision Rules

## SR-001: Readiness belongs to a Student Song relationship

Нельзя хранить готовность только на уровне глобальной Song.

Reason Code: `SONG_READINESS_STUDENT_CONTEXT_REQUIRED`

---

## SR-002: Active repertoire entry is required

Обычное изменение Readiness требует существующей Repertoire Entry.

Исторический пересчет может работать с архивной записью.

Reason Code: `REPERTOIRE_ENTRY_REQUIRED`

---

## SR-003: Song version must be identifiable

Если существуют разные версии произведения, Readiness должна относиться к определенной версии.

Например:

- оригинал;
- сокращенная версия;
- другая тональность;
- акустическая аранжировка;
- дуэт;
- измененный текст;
- концертная версия.

Reason Code: `SONG_VERSION_REQUIRED`

---

## SR-004: Readiness is not inferred from elapsed time

Количество недель работы не определяет Stage автоматически.

Reason Code: `TIME_SPENT_NOT_READINESS_EVIDENCE`

---

## SR-005: Lesson count is not readiness evidence by itself

Пять проведенных Lesson не означают, что Song готова.

Reason Code: `LESSON_COUNT_NOT_SONG_READINESS`

---

## SR-006: Song selection does not imply suitability

Добавление Song в Repertoire не подтверждает:

- соответствие диапазону;
- педагогическую целесообразность;
- готовность к работе;
- готовность к Concert.

---

## SR-007: Stage advancement requires relevant Evidence

Каждый переход должен иметь Evidence, относящееся к требованиям целевого Stage.

Reason Code: `SONG_READINESS_EVIDENCE_REQUIRED`

---

## SR-008: Evidence must refer to the same song version

Наблюдение по одной аранжировке нельзя автоматически использовать для другой существенно отличающейся версии.

Reason Code: `SONG_VERSION_EVIDENCE_MISMATCH`

---

## SR-009: Knowing lyrics is not performance readiness

Memorized не может автоматически переводить Song в Performance Ready.

Reason Code: `MEMORIZATION_NOT_PERFORMANCE_READINESS`

---

## SR-010: Technical quality alone is not full performance readiness

Технически стабильное исполнение может не иметь:

- интерпретации;
- сценической устойчивости;
- готового формата;
- полной структуры;
- актуальной репетиции.

---

## SR-011: Interpretation alone is insufficient

Эмоциональная выразительность не компенсирует обязательные технические и структурные требования.

---

## SR-012: Homework completion does not prove readiness

Самостоятельное выполнение задания без Review не меняет Readiness автоматически.

Reason Code: `HOMEWORK_COMPLETION_NOT_READINESS_EVIDENCE`

---

## SR-013: Reviewed homework may provide limited evidence

Teacher-reviewed Homework может подтверждать конкретную область:

- знание текста;
- сложный фрагмент;
- ритмику;
- интонацию;
- самостоятельность.

Но не должно автоматически подтверждать целостную Performance Readiness.

---

## SR-014: One successful fragment is not whole-song readiness

Успешное выполнение отдельного фрагмента не означает готовность всего произведения.

Reason Code: `PARTIAL_FRAGMENT_NOT_FULL_SONG_READINESS`

---

## SR-015: Full run-through is required for higher stages

Переход в Rehearsal Ready или Performance Ready обычно требует хотя бы одного подтвержденного целостного исполнения в релевантном формате.

Конкретные исключения утверждаются Education Lead.

Reason Code: `FULL_RUN_THROUGH_REQUIRED`

---

## SR-016: Repetition increases confidence

Несколько устойчивых исполнений могут повышать Confidence без изменения Stage.

---

## SR-017: One poor performance does not automatically lower stage

Единичная ошибка не должна автоматически переводить Song назад.

Причинами могут быть:

- волнение;
- техническая проблема;
- усталость;
- незнакомая площадка;
- ошибка минуса;
- недостаточная разминка.

Reason Code: `SINGLE_POOR_PERFORMANCE_NOT_STAGE_REGRESSION`

---

## SR-018: Sustained instability may require review

Повторяющиеся сложности в релевантных условиях могут привести к:

- снижению Confidence;
- Needs Rework;
- возврату на предыдущую Stage;
- созданию нового Goal;
- Teacher Review.

---

## SR-019: Context differences must not be silently averaged

Song может быть стабильной на Lesson и нестабильной на сцене.

В таком случае требуется Context Split.

Reason Code: `SONG_READINESS_CONTEXT_VARIATION`

---

## SR-020: Performance Ready requires Teacher confirmation

Окончательное присвоение Performance Ready требует подтверждения Teacher или другого уполномоченного Education Actor.

Reason Code: `PERFORMANCE_READY_REQUIRES_TEACHER`

---

## SR-021: AI cannot confirm performance readiness

AI может предложить Stage или обнаружить проблемы, но не может подтвердить Performance Ready.

Reason Code: `AI_CANNOT_CONFIRM_SONG_READINESS`

---

## SR-022: Student can submit self-readiness

Student может указать:

- насколько уверенно знает материал;
- какие части вызывают трудности;
- готов ли попробовать полное исполнение;
- как ощущает сценическую готовность.

Self Assessment хранится отдельно и не меняет подтвержденную Stage самостоятельно.

---

## SR-023: Student cannot self-approve performance readiness

Student не может самостоятельно присвоить Performance Ready.

Reason Code: `SELF_APPROVED_PERFORMANCE_READY_NOT_ALLOWED`

---

## SR-024: Administrator cannot make pedagogical readiness decisions by default

Administrator может управлять организационными данными, но не подтверждает вокальную готовность без отдельного полномочия.

---

## SR-025: Format readiness does not prove pedagogical readiness

Наличие минуса, микрофона и заявки на Concert не означает, что Song подготовлена педагогически.

Reason Code: `FORMAT_READY_NOT_PEDAGOGICALLY_READY`

---

## SR-026: Pedagogical readiness does not guarantee event eligibility

Performance Ready не означает автоматическое участие в конкретном Concert.

Reason Code: `SONG_READY_NOT_CONCERT_ELIGIBILITY`

---

## SR-027: Stage progression may skip stages only explicitly

Некоторые Song могут быть знакомы Student заранее.

Teacher может подтвердить переход сразу на более поздний Stage.

Пропущенные этапы должны быть объяснены в Audit.

Reason Code: `SONG_STAGE_SKIP_REQUIRES_CONFIRMATION`

---

## SR-028: Stage regression must preserve history

Возврат на предыдущую Stage создает новую версию Readiness.

Старое состояние не переписывается.

---

## SR-029: Regression wording must be supportive

Student Visible explanation не должна звучать как наказание.

Недопустимо:

> Песня больше не готова, потому что вы стали исполнять хуже.

Допустимо:

> После перерыва песне требуется несколько повторных репетиций, чтобы восстановить прежнюю устойчивость.

---

## SR-030: Long inactivity lowers current confidence

После длительного отсутствия практики Song может сохранить исторический Stage, но получить:

```text
Confidence: Low
Current Status: Needs Reconfirmation
```

или перейти в Maintained / Needs Rework согласно утвержденной модели.

---

## SR-031: Performed does not imply maintained

После Concert Song получает факт Performed.

Она не становится автоматически активной и готовой для будущих выступлений.

---

## SR-032: Maintained requires recent confirmation

Для статуса Maintained должно существовать недавнее подтверждение, что Student сохраняет способность исполнить Song на необходимом уровне.

---

## SR-033: Change of key may require partial reassessment

Изменение тональности может повлиять на:

- диапазон;
- регистровые переходы;
- дыхание;
- интерпретацию;
- техническую устойчивость.

Политика должна определить, какие Evidence остаются применимыми.

---

## SR-034: Significant arrangement change creates a new readiness context

Если аранжировка существенно изменена, прежняя Performance Readiness не переносится автоматически.

Reason Code: `SIGNIFICANT_ARRANGEMENT_CHANGE_REQUIRES_REVIEW`

---

## SR-035: Duet or ensemble version requires participant context

Готовность сольной версии не доказывает готовность дуэта или ансамбля.

Необходимо учитывать:

- конкретных участников;
- партии;
- взаимодействие;
- совместные репетиции.

---

## SR-036: Evidence withdrawal triggers recalculation

Если Assessment или другое активное Evidence отозвано, Readiness должна быть пересмотрена.

Reason Code: `SONG_READINESS_RECALCULATION_REQUIRED`

---

## SR-037: Superseded evidence must not count twice

Старая и новая версии одного Assessment не являются независимыми Evidence.

---

## SR-038: Readiness updates are versioned

Каждое значимое изменение создает новую версию Song Readiness State.

---

## SR-039: Duplicate triggers are idempotent

Повторная обработка одного события не создает новую идентичную версию.

Reason Code: `SONG_READINESS_TRIGGER_ALREADY_PROCESSED`

---

## SR-040: Concurrent evidence must not be lost

При одновременной публикации нескольких Assessment система должна повторно оценить Policy на актуальном состоянии.

Reason Code: `SONG_READINESS_VERSION_CONFLICT`

---

## SR-041: Visibility of evidence must be preserved

Student Visible Summary не должен раскрывать закрытые Teacher notes.

---

## SR-042: Readiness cannot be used to rank students

Нельзя сравнивать скорость подготовки Song между учениками для определения качества или ценности Student.

Reason Code: `SONG_READINESS_RANKING_NOT_ALLOWED`

---

## SR-043: Not Suitable requires explanation and next step

Решение Not Suitable должно включать:

- профессиональную причину;
- контекст;
- возможную адаптацию;
- альтернативу;
- возможность пересмотра.

Оно не должно содержать категоричных выводов о способностях Student.

---

## SR-044: Readiness is not a medical judgment

Политика не должна диагностировать голосовые или медицинские состояния.

При наличии риска Teacher может рекомендовать остановить работу и обратиться к профильному специалисту вне доменной оценки готовности.

---

## SR-045: Student consent may be required for sensitive performance contexts

Готовность к публичному выступлению не заменяет согласие Student на участие.

---

# Stage Transition Guidance

## Selected → Exploring

Возможные основания:

- Teacher начал предварительную работу;
- определена цель знакомства;
- начато обсуждение версии.

## Exploring → Learning

Обычно требуется:

- подтверждение, что Song подходит для текущей работы;
- выбрана рабочая версия;
- начато изучение материала.

## Learning → Technical Work

Обычно требуется:

- Student ориентируется в основной мелодии и структуре;
- известны ключевые сложные места;
- работа сфокусирована на качестве исполнения.

## Technical Work → Interpretation

Обычно требуется:

- основной материал достаточно стабилен;
- техническая работа позволяет перейти к художественной задаче.

Техническая работа при этом может продолжаться.

## Interpretation → Memorized

Обычно требуется:

- подтвержденное исполнение без существенной зависимости от текста;
- знание структуры;
- способность продолжить после небольшой ошибки.

## Memorized → Rehearsal Ready

Обычно требуется:

- целостное исполнение;
- утвержденная версия;
- готовый рабочий формат;
- достаточная техническая стабильность;
- Teacher confirmation.

## Rehearsal Ready → Performance Ready

Обычно требуется:

- минимум одна релевантная полная репетиция;
- подтвержденная устойчивость;
- приемлемая техническая и художественная готовность;
- отсутствие нерешенных блокирующих областей;
- Teacher confirmation.

## Performance Ready → Performed

Требуется подтвержденное событие исполнения.

## Performed → Maintained

Требуется решение продолжать хранить Song в активном репертуаре и подтверждение актуального состояния.

---

# Blocking Areas

Переход на следующий Stage может блокироваться конкретными областями.

Примеры:

- lyrics incomplete;
- melody unstable;
- rhythm unstable;
- key not approved;
- accompaniment missing;
- difficult section unresolved;
- full run-through missing;
- interpretation incomplete;
- microphone rehearsal missing;
- stage stability unconfirmed;
- teacher confirmation missing.

Blocking Area должна быть объяснима и иметь следующий шаг.

---

# Decision Flow

```text
Trigger received
      |
      v
Validate Student, Song and Repertoire Entry
      |
      v
Resolve Song version and context
      |
      v
Load current Readiness State
      |
      v
Normalize Evidence
      |
      v
Check relevance, authorship and visibility
      |
      +--> Rejected
      |
      v
Record Evidence
      |
      v
Evaluate target stage requirements
      |
      +--> No Change Required
      |
      +--> Insufficient Evidence
      |
      +--> Conflicting Evidence
      |
      +--> Context Split Required
      |
      +--> Teacher Review Required
      |
      +--> Stage Advancement Proposed
      |
      +--> Stage Regression Proposed
      |
      v
Apply confirmed update
      |
      v
SongReadinessChanged
      |
      +--> Concert Eligibility Policy
      |
      +--> Goal Completion Policy
      |
      +--> Achievement Award Policy
      |
      +--> Notification Policy
```

---

# Commands Produced

## RecordSongReadinessEvidence

Сохраняет допустимую ссылку на Evidence.

## UpdateSongReadiness

Создает новую версию Readiness State.

## RequestTeacherSongReview

Создает запрос на профессиональный Review.

## MarkSongReadinessConflict

Фиксирует несовместимые Evidence.

## ProposeSongForConcert

Может быть сформирована после подтвержденного Performance Ready.

Это только предложение для Concert Eligibility Policy.

## ReturnSongToPreviousStage

Применяется после подтвержденного Review.

## ArchiveSongFromActiveRepertoire

Переводит Repertoire Entry в архив без удаления истории.

---

# Domain Events

После операций могут создаваться:

```text
SongReadinessEvidenceRecorded
SongReadinessReviewRequested
SongReadinessStageProposed
SongReadinessChanged
SongReadinessConfidenceChanged
SongReadinessConflictDetected
SongReadinessContextSplit
SongMarkedPerformanceReady
SongReadinessReassessmentRequested
SongReturnedToPreviousStage
SongPerformed
SongArchivedFromRepertoire
```

## SongReadinessChanged Event

Событие должно содержать:

- SongReadinessId;
- StudentId;
- SongId;
- RepertoireEntryId;
- SongVersion reference;
- Context;
- PreviousStage;
- NewStage;
- Confidence;
- Evidence References;
- ConfirmedBy;
- EffectiveFrom;
- PolicyId;
- PolicyVersion;
- CorrelationId;
- CausationId.

Полные тексты закрытых Assessment не включаются.

---

# Human Review

Teacher Review обязателен при:

- присвоении Performance Ready;
- значительном пропуске стадий;
- возврате на предыдущий Stage;
- конфликтующих Assessment;
- смене версии или аранжировки;
- длительном перерыве перед Concert;
- разнице между Lesson и stage context;
- AI-generated proposal;
- спорном Not Suitable;
- запросе Student на пересмотр.

## Review Result

Teacher может:

- подтвердить текущий Stage;
- повысить Stage;
- понизить Stage;
- разделить Context;
- изменить Confidence;
- указать Blocking Areas;
- определить следующий шаг;
- заменить Song version;
- приостановить работу;
- признать Song неподходящей;
- предложить Concert Eligibility review.

Решение должно иметь Explanation.

---

# Concert Integration

Song Readiness Policy отвечает только на вопрос:

> Насколько данная Song подготовлена Student в определенном контексте?

Concert Eligibility Policy отвечает на вопрос:

> Может ли Student исполнить эту Song на конкретном Concert?

Даже при Performance Ready могут существовать ограничения:

- неподходящий формат Concert;
- превышена длительность программы;
- нет необходимой техники;
- отсутствует согласие Student;
- конфликт репертуара;
- нет совместной репетиции;
- слишком давняя оценка;
- изменена аранжировка;
- мероприятие требует другого уровня подготовки.

---

# Progress Integration

Song Readiness может использовать Skill Progress как Evidence.

Но Skill Progress и Song Readiness не идентичны.

Пример:

Student имеет устойчивый Progress по интонации, но конкретная Song все еще содержит сложный переход.

И наоборот:

Student может подготовить одну Song за счет точечной работы, но это не доказывает устойчивое развитие Skill во всем репертуаре.

---

# Goal Integration

Goal может быть связана с Song.

Пример:

> Подготовить песню к школьному концерту.

Такая Goal не завершается только при присвоении Performance Ready, если критерии также включают:

- полную репетицию;
- участие в Concert;
- запись;
- сценическую задачу;
- Teacher Review.

Goal Completion Policy проверяет собственные критерии отдельно.

---

# Achievement Integration

После событий:

```text
SongMarkedPerformanceReady
SongPerformed
```

Achievement Award Policy может проверить:

- первая подготовленная Song;
- первое полное исполнение;
- первое сценическое произведение;
- новый этап репертуара.

Song Readiness Policy не выдает Achievement самостоятельно.

---

# Notification Effects

Политика не отправляет уведомления напрямую.

Notification Policy может решить:

- сообщить Student о новом Stage;
- уведомить Teacher о Review;
- предложить репетицию;
- не уведомлять о техническом пересчете;
- отложить сообщение до личного обсуждения;
- объединить обновление с новым Goal или Homework.

---

# Student Presentation

Student должен видеть:

- текущую Stage;
- понятное объяснение;
- что уже получается;
- что требует работы;
- следующий шаг;
- связанные доступные Evidence;
- дату последнего подтверждения;
- актуальность статуса.

Следует избегать:

- необъяснимых процентов;
- красных оценок;
- формулировки «не готов» без контекста;
- сравнения с другими;
- давления выступать;
- представления Stage как оценки личности.

---

# Teacher Presentation

Teacher должен видеть:

- полный Evidence timeline;
- Song version;
- контексты;
- текущую Stage;
- Confidence;
- Blocking Areas;
- историю переходов;
- Assessment;
- Lesson Results;
- Homework Reviews;
- AI proposals;
- ближайшие Concert requirements;
- необходимость подтверждения.

---

# Administrator Presentation

Administrator может видеть организационные данные:

- Song;
- Student;
- подтвержденный Stage;
- Concert proposal;
- длительность;
- версия минуса;
- техническая готовность;
- наличие Teacher confirmation.

Закрытые педагогические заметки не раскрываются по умолчанию.

---

# Owner Analytics

Owner может видеть агрегированные данные:

- количество активных Songs;
- распределение по Stages;
- среднее время между стадиями;
- количество конфликтов;
- долю давно не подтвержденных Readiness;
- количество Performance Ready Songs;
- частоту возврата на предыдущий Stage;
- полноту Teacher Reviews.

Эти данные не должны автоматически использоваться для оценки качества Student или Teacher.

---

# AI Assistance

AI может:

- структурировать Lesson notes;
- предлагать Readiness Area;
- находить повторяющиеся сложности;
- сравнивать Evidence по одной Song;
- выявлять возможные противоречия;
- создавать Draft Summary;
- предлагать Blocking Areas;
- проверять полноту Stage requirements;
- предложить Teacher Review.

AI не может:

- подтверждать Performance Ready;
- определять пригодность голоса медицински;
- автоматически отклонять Song;
- публиковать Stage;
- подменять Teacher Assessment;
- создавать факты;
- использовать неподтвержденные аудиоаналитические выводы как окончательную истину.

AI metadata должна содержать:

- модель или механизм;
- версию;
- входные Evidence;
- Confidence;
- время;
- предложенный Stage;
- подтверждение Teacher.

---

# Privacy

Song Readiness может содержать:

- закрытую обратную связь;
- записи исполнения;
- сведения о сценическом волнении;
- персональные педагогические наблюдения;
- информацию о будущих выступлениях.

Необходимо:

- соблюдать Visibility;
- не включать private notes в события;
- контролировать доступ к записям;
- хранить consent на публикацию;
- ограничивать доступ после смены Teacher;
- использовать безопасные Student Summaries;
- разделять внутреннюю оценку и публичный статус.

---

# Audit Requirements

Для каждой оценки политики сохраняются:

- PolicyId;
- PolicyVersion;
- TriggerEventId;
- StudentId;
- SongId;
- RepertoireEntryId;
- SongVersion;
- Context;
- current Readiness version;
- Evidence References;
- ActorId;
- Decision;
- Reason Codes;
- proposed Stage;
- final Stage;
- Confidence;
- Blocking Areas;
- Teacher confirmation;
- AI metadata;
- EvaluatedAt;
- CorrelationId;
- CausationId.

Для изменения дополнительно:

- PreviousStage;
- NewStage;
- explanation;
- EffectiveFrom;
- skipped stages;
- confirmation reason;
- resulting version.

---

# Failure Modes

## Song not found

- Decision: Rejected
- Reason Code: SONG_NOT_FOUND

## Repertoire Entry not found

- Decision: Rejected
- Reason Code: REPERTOIRE_ENTRY_REQUIRED

## Song belongs to another Student context

- Decision: Rejected
- Reason Code: SONG_READINESS_STUDENT_MISMATCH

Security Audit обязателен.

## Song version is unknown

- Decision: Insufficient Evidence
- Reason Code: SONG_VERSION_REQUIRED

## Evidence targets another Song

- Decision: Rejected
- Reason Code: SONG_EVIDENCE_MISMATCH

## Homework completed but not reviewed

- Decision: Evidence Recorded
- Reason Code: HOMEWORK_COMPLETION_NOT_READINESS_EVIDENCE

Stage не меняется.

## One successful fragment

- Decision: Evidence Recorded
- Reason Code: PARTIAL_FRAGMENT_NOT_FULL_SONG_READINESS

## Conflicting Teacher Assessments

- Decision: Teacher Review Required
- Reason Code: CONFLICTING_SONG_ASSESSMENTS

## AI proposes Performance Ready

- Decision: Teacher Review Required
- Reason Code: AI_CANNOT_CONFIRM_SONG_READINESS

## Long break after previous performance

- Decision: Teacher Review Required or Stage Regression Proposed
- Reason Code: SONG_READINESS_NOT_RECENTLY_CONFIRMED

## Duplicate trigger

- Decision: No Change Required
- Reason Code: SONG_READINESS_TRIGGER_ALREADY_PROCESSED

## Concurrent update

- Decision: Deferred
- Reason Code: SONG_READINESS_VERSION_CONFLICT

Политика повторно оценивается на новой версии.

---

# Explainability Examples

## Learning

> Вы уже знакомы с основной мелодией и структурой песни. Следующий шаг — закрепить текст и отдельно разобрать сложные переходы.

## Technical Work

> Основной материал выучен. Сейчас главный фокус — стабильная интонация в припеве и распределение дыхания в длинных фразах.

## Interpretation

> Техническая основа стала устойчивее, поэтому работа перешла к смыслу текста, фразировке и эмоциональной подаче.

## Rehearsal Ready

> Песня исполняется целиком и готова к репетиции в концертном формате. На репетиции будет проверена работа с микрофоном и устойчивость исполнения без остановок.

## Performance Ready

> Преподаватель подтвердил, что песня стабильно исполняется целиком в выбранной версии и может рассматриваться для ближайшего выступления.

## Needs Rework

> После перерыва некоторые части требуют восстановления. Предыдущий результат сохраняется в истории, а сейчас рекомендуется провести несколько повторных репетиций.

## Context Variation

> На занятиях песня исполняется устойчиво, но в сценическом формате пока требуется дополнительная работа с волнением и микрофоном.

---

# Examples

## Example 1: Song added to repertoire

Дано:

- Song добавлена Student;
- работа еще не началась.

Результат:

- Decision: Stage Advancement Approved
- New Stage: Selected
- Reason Code: SONG_ADDED_TO_REPERTOIRE

## Example 2: Student learned the lyrics

Дано:

- Student знает текст;
- Teacher подтвердил;
- полного исполнения еще не было.

Результат:

- Decision: Evidence Recorded
- Possible Stage: Memorized
- Performance Ready: No
- Reason Code: MEMORIZATION_NOT_PERFORMANCE_READINESS

Фактический Stage зависит от остальных требований.

## Example 3: Stable full lesson performance

Дано:

- Student исполнил Song полностью;
- Teacher отметил техническую устойчивость;
- концертного формата еще не было.

Результат:

- Decision: Stage Advancement Proposed
- Proposed Stage: Rehearsal Ready
- Reason Code: FULL_LESSON_RUN_THROUGH_CONFIRMED

## Example 4: Ready on lesson, unstable with microphone

Дано:

- на Lesson Song стабильна;
- на репетиции с микрофоном появились существенные трудности.

Результат:

- Decision: Context Split Required

```text
Regular Lesson:
  Stage: Performance Ready

Public Performance:
  Stage: Rehearsal Ready
```

- Reason Code: SONG_READINESS_CONTEXT_VARIATION

## Example 5: AI analyzes recording

Дано:

- AI проанализировал запись;
- предложил Performance Ready;
- Teacher еще не подтвердил.

Результат:

- Decision: Teacher Review Required
- Reason Code: AI_CANNOT_CONFIRM_SONG_READINESS

## Example 6: Changed key

Дано:

- Song ранее была Performance Ready;
- тональность изменена;
- диапазон и переходы существенно отличаются.

Результат:

- Decision: Teacher Review Required
- Reason Code: SONG_KEY_CHANGE_REQUIRES_REASSESSMENT

## Example 7: First performance completed

Дано:

- Student выступил с Song на Concert;
- участие подтверждено.

Результат:

- Decision: Stage Advancement Approved
- New Stage: Performed
- Reason Code: SONG_PERFORMANCE_CONFIRMED

## Example 8: Long inactivity

Дано:

- Song исполнялась год назад;
- с тех пор не повторялась;
- новых Evidence нет.

Результат:

- Decision: Stage Regression Proposed
- Proposed Stage: Needs Rework
- Confidence: Low
- Reason Code: SONG_READINESS_NOT_RECENTLY_CONFIRMED

Teacher Review может сохранить Maintained, если есть иные основания.

---

# Test Requirements

## Stage Tests

- new repertoire entry starts at Selected;
- stage can advance with valid Evidence;
- stage does not advance from elapsed time;
- stage can skip only with confirmation;
- Performance Ready requires Teacher;
- Performed requires confirmed performance;
- Maintained requires recent evidence;
- regression preserves history.

## Evidence Tests

- relevant Assessment is accepted;
- unrelated Song Assessment is rejected;
- wrong Song version is rejected;
- Homework completion without review does not advance Stage;
- reviewed Homework can support a limited area;
- one fragment does not prove full readiness;
- withdrawn Evidence triggers recalculation;
- superseded Evidence does not count twice.

## Context Tests

- Lesson and Concert contexts can differ;
- context difference does not become false conflict;
- duet readiness is separate from solo readiness;
- studio recording does not automatically prove public stage readiness;
- format readiness is separate from pedagogical readiness.

## Permission Tests

- assigned Teacher can confirm;
- authorized replacement Teacher can confirm;
- Student cannot self-approve Performance Ready;
- AI cannot confirm;
- Administrator cannot make pedagogical decision by default;
- unauthorized Teacher is rejected.

## Version Tests

- every change creates a new version;
- historical versions remain reproducible;
- duplicate trigger is idempotent;
- stale state causes retry;
- concurrent Evidence is preserved;
- Policy Version is stored.

## Review Tests

- conflicting Assessment requires Review;
- key change can require Review;
- arrangement change can require Review;
- long inactivity can require Review;
- stage regression requires explanation;
- Not Suitable requires a next step.

## Privacy Tests

- Student Summary excludes private notes;
- Administrator does not see Teacher Only Assessment;
- event does not contain recording contents;
- public visibility requires consent;
- former Teacher access follows scope.

## AI Tests

- AI suggestion remains unconfirmed;
- hallucinated Evidence is rejected;
- AI cannot publish Stage;
- AI cannot mark Not Suitable independently;
- AI Confidence is not domain Confidence;
- Teacher confirmation metadata is stored.

## Explainability Tests

- each stage change has Reason Codes;
- Student receives a readable explanation;
- blocking areas contain next actions;
- regression language is supportive;
- Context Split is understandable;
- No Change decision is explainable.

---

# Non-Goals

Song Readiness Policy не определяет:

- окончательный допуск к Concert;
- расписание репетиций;
- финансовые условия выступления;
- CRM-статусы;
- авторские права на Song;
- лицензирование музыки;
- техническую доставку минусов;
- медицинскую безопасность голоса;
- автоматический выбор репертуара;
- рейтинг Song;
- рейтинг Student;
- оплату Teacher;
- Achievement Award;
- BelCoin Reward.

---

# Open Questions

Необходимо определить:

- официальный перечень Song Stages;
- обязательны ли все стадии;
- можно ли хранить несколько активных Readiness Contexts;
- как версионировать тональность и аранжировку;
- какие Stage видит Student;
- может ли Student предложить переход;
- кто подтверждает Performance Ready при нескольких Teachers;
- сколько полных исполнений требуется;
- какой срок считать Readiness актуальным;
- как учитывать простые и сложные Songs;
- нужна ли отдельная шкала сложности;
- как работать с дуэтами и ансамблями;
- как учитывать совместную готовность участников;
- является ли Memorized отдельной Stage или характеристикой;
- нужна ли отдельная Technical Readiness;
- как учитывать телесуфлер или текст на сцене;
- как учитывать импровизационные произведения;
- как работать с сокращенными концертными версиями;
- как переносить Readiness при смене минуса;
- что считать существенным изменением аранжировки;
- кто может присвоить Not Suitable;
- как корректно показывать возврат на предыдущий Stage;
- должна ли Song автоматически архивироваться после долгой паузы;
- как связывать Readiness с Goal;
- какие события запускают Concert Eligibility Policy;
- может ли Performance Ready истекать автоматически;
- нужен ли обязательный Review перед каждым Concert;
- как хранить требования конкретного performance context.

---

# History

| Version | Description |
| --- | --- |
| 1.0.0 | Определена модель стадий, Evidence и правил готовности Song. |
