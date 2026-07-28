---
Status: Accepted
Version: 1.0.0
Last Updated: 2026-07-27

Policy Id: HOMEWORK_EXPIRATION_POLICY

Policy Type:
  - Lifecycle Policy
  - Validation Policy
  - Reaction Policy
  - Escalation Policy
  - Recommendation Policy

Owners:
  - Product Owner
  - Education Lead
  - Technical Lead

Related Aggregates:
  - Homework
  - HomeworkAssignment
  - Student
  - Teacher
  - Lesson
  - Goal
  - Song
  - Concert
  - LearningPause
  - Assessment
  - Progress

Observed Events:
  - HomeworkAssigned
  - HomeworkUpdated
  - HomeworkDueDateChanged
  - HomeworkSubmitted
  - HomeworkReviewed
  - HomeworkCorrectionRequested
  - HomeworkCancelled
  - HomeworkReplaced
  - HomeworkDueDateReached
  - HomeworkExpirationEvaluationRequested
  - LessonRescheduled
  - LessonCancelled
  - StudentLearningPauseStarted
  - StudentLearningPauseEnded
  - GoalCompleted
  - GoalCancelled
  - SongReadinessChanged
  - ConcertCompleted
  - ConcertCancelled
  - TeacherHomeworkReviewCompleted

Produced Commands:
  - MarkHomeworkOverdue
  - ExpireHomework
  - ExtendHomeworkDueDate
  - KeepHomeworkActive
  - RequestHomeworkExpirationReview
  - RequestTeacherHomeworkDecision
  - CancelHomework
  - ReplaceHomework
  - ReopenHomework
  - SuppressHomeworkReminders
  - RecalculateHomeworkReminderPlan
  - NotifyStudentAboutHomeworkStatus
  - ArchiveHomework

Related Documents:
  - 000-domain-policy-overview.md
  - lesson-completion-policy.md
  - progress-update-policy.md
  - goal-completion-policy.md
  - song-readiness-policy.md
  - concert-eligibility-policy.md
  - homework-reminder-policy.md
  - notification-policy.md
  - periodic-review-policy.md
  - ../homework.md
  - ../lesson.md
  - ../student.md
  - ../progress.md
---

# Homework Expiration Policy

> Homework Expiration Policy определяет, что происходит с Homework после наступления Due Date или после утраты его образовательной актуальности.
>
> Наступление срока не означает автоматический провал, наказание или удаление задания.

---

# Purpose

Homework может иметь срок, связанный с:

- следующим Lesson;
- конкретным этапом Goal;
- работой над Song;
- репетицией;
- Concert;
- периодом самостоятельной практики;
- проверкой определенного Skill;
- временно доступным материалом.

После наступления срока система должна определить:

- остается ли Homework полезным;
- можно ли принять позднюю Submission;
- следует ли изменить Due Date;
- нужно ли запросить решение Teacher;
- нужно ли заменить Assignment;
- потеряло ли задание образовательный смысл;
- следует ли закрыть его как Expired;
- нужно ли прекратить Reminder;
- должен ли статус влиять на Progress или Goal;
- требуется ли объяснение Student.

Без отдельной политики система может ошибочно:

- считать любое опоздание неуспехом;
- закрывать полезное задание;
- бесконечно хранить старые Assignments активными;
- продолжать навязчивые Reminders;
- терять Submission, отправленную после срока;
- наказывать Student за перенос Lesson;
- считать Learning Pause нарушением;
- завершать Goal на основании технического срока;
- переписывать историю Homework;
- автоматически ухудшать Progress.

---

# Core Principle

Due Date является точкой принятия решения, а не автоматическим приговором.

```text
Due Date Reached
      |
      v
Evaluate Current Educational Context
      |
      +--> Keep Active
      |
      +--> Mark Overdue
      |
      +--> Extend Due Date
      |
      +--> Teacher Review Required
      |
      +--> Replace
      |
      +--> Cancel
      |
      +--> Expire
```

Результат зависит не только от времени, но и от цели Assignment.

---

## Due Date Is Not Expiration Date

Необходимо различать:

### Due Date

Ожидаемый срок выполнения или Submission.

### Grace Period End

Конец разрешенного периода поздней Submission, если он существует.

### Educational Validity End

Момент, после которого Homework теряет образовательную актуальность.

### Expiration Evaluation Time

Момент запуска политики.

### Expired At

Момент фактического перехода Homework в Expired.

Эти значения могут совпадать, но не обязаны.

---

# Homework Lifecycle

Рекомендуемая модель:

```text
Draft
  |
  v
Assigned
  |
  +--> In Progress
  |
  +--> Submitted
  |
  +--> Clarification Required
  |
  +--> Overdue
  |
  +--> Correction Requested
  |
  +--> Completed
  |
  +--> Cancelled
  |
  +--> Replaced
  |
  +--> Expired
  |
  v
Archived
```

---

# Relevant Homework States

## Assigned

Homework доступно Student, но выполнение не подтверждено.

## In Progress

Student явно начал работу или сохранил промежуточный результат.

Этот статус не должен выводиться только из открытия экрана.

## Submitted

Student отправил результат.

Submission может быть сделана до или после Due Date.

## Overdue

Due Date прошло, но Homework остается активным и может быть выполнено или проверено.

Overdue не означает Expired.

## Clarification Required

Homework временно невозможно корректно выполнить без уточнения.

## Correction Requested

Teacher запросил исправление после Submission.

Для исправления может устанавливаться новый Due Date.

## Completed

Homework проверено или иным способом завершено согласно Homework Completion model.

## Cancelled

Teacher или уполномоченный Actor отменил Homework.

Отмена не означает, что Student не справился.

## Replaced

Homework заменено другим Assignment.

## Expired

Homework больше не считается активным и не должно ожидать обычного выполнения.

История сохраняется.

## Archived

Homework остается доступным в истории, но исключено из активных представлений.

---

# Homework Expiration State

Рекомендуемая структура:

```text
HomeworkExpirationState
├── HomeworkAssignmentId
├── StudentId
├── HomeworkVersion
├── DueDate
├── GracePeriodEnd
├── EducationalValidityEnd
├── CurrentStatus
├── ExpirationDecision
├── ReasonCodes
├── BlockingIssues
├── RelatedLessonId
├── RelatedGoalId
├── RelatedSongId
├── RelatedConcertId
├── EvaluatedBy
├── EvaluatedAt
├── EffectiveAt
├── PolicyId
├── PolicyVersion
└── Version
```

---

# Expiration Decision

```text
HomeworkExpirationDecision
├── DecisionId
├── HomeworkAssignmentId
├── HomeworkVersion
├── Outcome
├── PreviousStatus
├── NewStatus
├── DueDate
├── NewDueDate
├── GracePeriodEnd
├── EducationalContext
├── EvidenceReferences
├── ReasonCodes
├── StudentExplanation
├── TeacherExplanation
├── EvaluatedBy
├── HumanReviewRequired
├── PolicyId
├── PolicyVersion
├── EvaluatedAt
└── Version
```

---

# Trigger

Политика применяется при:

```text
HomeworkDueDateReached
HomeworkExpirationEvaluationRequested
HomeworkDueDateChanged
HomeworkSubmitted
HomeworkCorrectionRequested
HomeworkCancelled
HomeworkReplaced
LessonRescheduled
LessonCancelled
StudentLearningPauseStarted
StudentLearningPauseEnded
GoalCompleted
GoalCancelled
SongReadinessChanged
ConcertCompleted
ConcertCancelled
TeacherHomeworkReviewCompleted
```

---

# Inputs

Для оценки могут потребоваться:

- Homework Assignment;
- Homework Version;
- Homework Status;
- Assignment Type;
- Due Date;
- Grace Period;
- Educational Validity End;
- Required or Optional status;
- Student Submission;
- Student-reported blocker;
- Clarification state;
- Teacher instructions;
- related Lesson;
- related Goal;
- related Song;
- related Concert;
- Learning Pause;
- Student Timezone;
- reminder history;
- previous extensions;
- correction history;
- Teacher decision;
- Policy Version.

---

# Expiration Outcomes

## Keep Active

Homework остается активным без изменения Due Date.

Используется осторожно, если Due Date носила ориентировочный характер.

## Mark Overdue

Homework остается выполнимым, но срок прошел.

## Extend Due Date

Устанавливается новый срок.

## Grace Period Active

Homework временно остается доступным в пределах заранее разрешенного периода.

## Teacher Review Required

Автоматического решения недостаточно.

## Student Clarification Required

Необходимо уточнить намерение или проблему Student.

## Replace Homework

Текущее Assignment заменяется новым.

## Cancel Homework

Homework прекращается без статуса Expired.

## Expire Homework

Homework теряет активность и переводится в Expired.

## Reopen Homework

Ранее закрытое Homework снова становится активным через отдельное решение.

## No Change Required

Текущее состояние уже соответствует правилам.

## Rejected

Запрос недействителен или неавторизован.

---

# Expiration Categories

## Deadline Expiration

Решение связано с наступлением Due Date.

## Context Expiration

Homework потеряло актуальность из-за изменения образовательного контекста.

Примеры:

- Concert уже завершен;
- Song больше не изучается;
- Goal отменена;
- Lesson sequence изменена;
- задание заменено;
- материал устарел.

## Resource Expiration

Необходимый ресурс больше недоступен или недействителен.

## Administrative Expiration

Homework закрывается в рамках миграции, завершения учебного периода или архивирования.

## Safety Expiration

Продолжение задания признано неуместным или потенциально небезопасным.

Политика не ставит медицинские диагнозы.

---

# Decision Rules

## HE-001: Due date does not automatically expire homework

Наступление Due Date само по себе не должно безусловно переводить Homework в Expired.

Reason Code: `DUE_DATE_NOT_AUTOMATIC_EXPIRATION`

---

## HE-002: Expiration requires an explicit policy decision

Каждый переход в Expired должен иметь:

- Decision;
- Reason Code;
- EffectiveAt;
- Policy Version;
- Audit metadata.

Reason Code: `HOMEWORK_EXPIRATION_DECISION_REQUIRED`

---

## HE-003: Homework version is required

Решение должно относиться к конкретной версии Assignment.

Reason Code: `HOMEWORK_VERSION_REQUIRED`

---

## HE-004: Completed homework cannot expire as incomplete

Если Homework уже Completed, последующая Expiration Evaluation не должна менять его в Expired.

Decision: No Change Required

Reason Code: `HOMEWORK_ALREADY_COMPLETED`

---

## HE-005: Submitted homework must not expire before review by default

Если Student отправил Homework до Expiration Decision, Assignment обычно переходит в Review flow.

Reason Code: `HOMEWORK_SUBMITTED_PENDING_REVIEW`

---

## HE-006: Late submission may be accepted

Submission после Due Date может быть допустима, если:

- Homework остается педагогически актуальным;
- нет жесткого event deadline;
- Teacher разрешил late submission;
- Grace Period активен;
- Assignment не заменено;
- связанные материалы действительны.

---

## HE-007: Late submission must preserve actual timestamps

Система сохраняет:

- Due Date;
- SubmittedAt;
- lateness duration;
- applicable timezone;
- Grace Period;
- status at submission time.

Нельзя переписывать Due Date задним числом только для сокрытия опоздания.

---

## HE-008: Lateness is not progress evidence

Факт поздней Submission не должен автоматически:

- снижать Progress;
- снижать Assessment;
- отменять Goal;
- отзывать Achievement;
- влиять на Song Readiness.

Reason Code: `LATENESS_NOT_PROGRESS_EVIDENCE`

---

## HE-009: Overdue and expired are distinct states

Overdue означает:

- срок прошел;
- Assignment активно;
- Submission еще может быть принята.

Expired означает:

- обычное выполнение больше не ожидается;
- Reminder прекращаются;
- требуется отдельное действие для Reopen.

---

## HE-010: Mark Overdue when work remains useful

Homework может перейти в Overdue, если:

- его цель остается актуальной;
- Student может выполнить его позже;
- связанные Goal или Song активны;
- материалы доступны;
- Teacher не отменил Assignment.

---

## HE-011: Optional homework may remain active without overdue pressure

Optional Homework может:

- остаться Assigned;
- получить статус Available;
- не отображаться как просроченное;
- не создавать escalation.

Reason Code: `OPTIONAL_HOMEWORK_NOT_PUNITIVELY_OVERDUE`

---

## HE-012: Required and optional homework must be distinguished

Политика не должна одинаково обрабатывать:

- обязательное задание;
- рекомендацию;
- дополнительную практику;
- справочный материал;
- добровольное упражнение.

---

## HE-013: Hard deadlines require explicit configuration

Жесткий срок должен быть указан в Assignment или связанном контексте.

Примеры:

- подготовка к конкретной репетиции;
- загрузка материала до Concert;
- задание, которое Teacher должен проверить до следующего Lesson;
- участие в ограниченном по времени проекте.

Reason Code: `HARD_DEADLINE_CONFIGURATION_REQUIRED`

---

## HE-014: Hard deadline may cause context expiration

После жесткого срока Homework может стать Expired, если выполнение больше не достигает исходной цели.

Пример:

> Подготовить вступление к генеральной репетиции, которая уже завершилась.

---

## HE-015: Soft deadline normally allows overdue state

При Soft Deadline предпочтительный результат: Mark Overdue

а не автоматический Expired.

---

## HE-016: Grace period must be explicit

Grace Period не должен придумывается динамически без утвержденного правила.

Он может быть задан:

- Assignment;
- Homework Type;
- Teacher;
- school policy;
- related event.

---

## HE-017: Grace period starts from the effective due date

Если Due Date была изменена до наступления срока, Grace Period рассчитывается от новой даты.

---

## HE-018: Grace period does not rewrite due date

Homework может быть:

```text
Status: Overdue
Grace Period: Active
```

Это позволяет сохранить правдивую историю.

---

## HE-019: Grace period end triggers reevaluation

После окончания Grace Period система снова применяет Policy.

Результатом может быть:

- Keep Active;
- Extend;
- Teacher Review;
- Expire.

---

## HE-020: Extensions must be explicit and auditable

Due Date нельзя молча сдвигать.

Для Extension сохраняются:

- previous Due Date;
- new Due Date;
- reason;
- Actor;
- requested by;
- approved by;
- timestamp;
- Homework Version.

Reason Code: `HOMEWORK_EXTENSION_REASON_REQUIRED`

---

## HE-021: Student may request an extension

Student может запросить перенос срока.

Запрос не означает автоматическое одобрение.

---

## HE-022: Extension request must not require sensitive disclosure

Student может указать:

- Need more time
- Temporarily unavailable
- Material unclear
- Technical problem

без раскрытия лишних личных или медицинских подробностей.

---

## HE-023: Teacher may extend homework

Authorized Teacher может изменить Due Date, если:

- Homework остается актуальным;
- новый срок выполним;
- изменение не конфликтует с жестким событием;
- Student получает обновленную информацию.

---

## HE-024: Administrator cannot make pedagogical extensions by default

Administrator может исправить техническую ошибку, но не должен самостоятельно менять образовательный срок без соответствующего полномочия.

---

## HE-025: Extension after expiration requires reopen flow

Если Homework уже Expired, обычное изменение Due Date недостаточно.

Необходимо:

`ReopenHomework`

или создать новое Assignment.

---

## HE-026: Reopen preserves expiration history

Reopen не удаляет:

- ExpiredAt;
- исходную причину;
- предыдущий Due Date;
- Decision;
- Actor;
- Audit.

---

## HE-027: Reopen requires current educational relevance

Перед Reopen проверяется:

- цель еще существует;
- материалы доступны;
- Assignment понятно;
- Student должен выполнять именно это задание;
- не существует замены;
- срок реалистичен.

---

## HE-028: Replaced homework must not remain active

Если Assignment заменено другим, старое получает Replaced.

Оно не должно дополнительно считаться Expired, если модель не требует исторической классификации.

Reason Code: `HOMEWORK_REPLACED`

---

## HE-029: Replacement must reference successor

Replaced Homework должно содержать ссылку на новое Assignment.

---

## HE-030: Replacement must not duplicate active obligations

Student не должен видеть старое и новое Homework как два независимых обязательных задания, если новое заменяет старое.

---

## HE-031: Cancelled homework is not expired homework

Cancellation означает, что Assignment прекращено решением Actor.

Expiration означает утрату актуальности по lifecycle rules.

История должна различать эти причины.

---

## HE-032: Cancellation requires explanation

Student-visible explanation может быть кратким:

> Задание отменено преподавателем и больше не требует выполнения.

---

## HE-033: Learning pause can suspend expiration

Если Learning Pause началась до Due Date, политика может:

- отложить Expiration Evaluation;
- продлить Due Date;
- запросить Teacher Review;
- сохранить Homework активным без Overdue;
- отменить задание, если оно потеряет актуальность.

---

## HE-034: Learning pause must not automatically mark homework overdue

Reason Code: `LEARNING_PAUSE_PREVENTS_AUTOMATIC_OVERDUE`

---

## HE-035: Pause ending triggers relevance review

После завершения Learning Pause нельзя автоматически возобновлять все старые Assignments.

Необходимо проверить:

- актуальность;
- нагрузку;
- новую программу;
- связанные Lessons;
- наличие replacement;
- новые сроки.

---

## HE-036: Long pause may justify replacement

После длительного перерыва правильнее создать новое Homework, чем возвращать старое без изменений.

---

## HE-037: Lesson rescheduling may require due date recalculation

Если Due Date связана со следующим Lesson, его перенос запускает Review.

Возможные решения:

- перенос Due Date;
- сохранение срока;
- Mark Overdue;
- Teacher Review.

---

## HE-038: Lesson rescheduling must not silently penalize student

Если школа или Teacher перенесли Lesson, Student не должен автоматически считаться просрочившим Homework, связанное с этим Lesson.

Reason Code: `RELATED_LESSON_RESCHEDULED`

---

## HE-039: Lesson cancellation does not automatically expire homework

Cancellation Lesson может привести к:

- переносу Assignment;
- привязке к новому Lesson;
- сохранению как самостоятельной практики;
- отмене Homework;
- Expiration;
- Teacher Review.

---

## HE-040: Lesson completion does not always expire homework

Homework может оставаться актуальным после связанного Lesson, если:

- Teacher продолжает работу;
- оно поддерживает Goal;
- требуется поздняя Submission;
- назначена Correction.

---

## HE-041: Goal completion may make homework redundant

Если Homework существовало только для завершенной Goal, требуется оценить:

- нужно ли еще закрепление;
- требуется ли Review;
- можно ли Cancel;
- следует ли Expire;
- имеет ли Submission самостоятельную ценность.

---

## HE-042: Goal completion does not automatically delete homework

История Assignment сохраняется.

---

## HE-043: Goal cancellation triggers context review

Homework, связанное исключительно с отмененной Goal, обычно:

- Cancelled;
- Replaced;
- Expired.

Решение зависит от педагогической актуальности.

---

## HE-044: Song change may invalidate homework

Если Student больше не работает над конкретной Song, связанное Homework может потерять актуальность.

Reason Code: `RELATED_SONG_CONTEXT_CHANGED`

---

## HE-045: Song version changes may require replacement

Задание по старой тональности или аранжировке нельзя автоматически считать применимым к новой версии.

---

## HE-046: Concert-linked homework may have hard contextual expiry

Homework, цель которого существовала только до Concert, может стать Expired после:

- завершения Concert;
- отмены Concert;
- снятия Participation;
- замены Song;
- окончания технического deadline.

---

## HE-047: Concert completion does not invalidate general skill practice

Если Assignment также развивает общий Skill, Teacher может:

- сохранить;
- заменить;
- преобразовать в Optional;
- назначить новый Due Date.

---

## HE-048: Concert cancellation triggers review, not automatic deletion

После отмены Concert Homework может оставаться полезным для:

- следующего выступления;
- Song Readiness;
- Goal;
- записи;
- общего развития.

---

## HE-049: Missing required material can suspend expiration

Если Homework нельзя было выполнить из-за отсутствующего school-provided material:

- Student не должен автоматически получать Overdue;
- Reminder подавляются;
- Teacher Attention создается;
- Due Date может быть изменена после восстановления доступа.

Reason Code: `HOMEWORK_REQUIRED_MATERIAL_UNAVAILABLE`

---

## HE-050: Technical failure may justify extension

Если подтвержденная техническая проблема препятствовала Submission, возможны:

- Grace Period;
- Extension;
- альтернативный способ Submission;
- Teacher Review.

---

## HE-051: Student-reported blocker must be considered

При активном Blocker Policy не должна автоматически переводить Homework в Expired без проверки.

---

## HE-052: Clarification-required homework should not expire as ordinary non-completion

Если Student своевременно запросил Clarification, срок может быть:

- приостановлен;
- перенесен;
- отправлен на Teacher Review.

Reason Code: `HOMEWORK_CLARIFICATION_PENDING`

---

## HE-053: Teacher delay must not penalize student

Если Teacher не ответил на Clarification или не предоставил материал, Student не должен считаться виновным в просрочке.

---

## HE-054: Submitted homework awaiting review remains submitted

Длительное отсутствие Teacher Review не переводит Homework в Expired.

---

## HE-055: Correction request creates a new completion window

После HomeworkCorrectionRequested должно быть определено:

- что исправить;
- новый срок;
- является ли срок hard или soft;
- можно ли отправить повторно;
- какая версия Submission активна.

---

## HE-056: Old due date must not govern correction automatically

Срок первоначального Assignment не должен безусловно применяться к исправлению.

---

## HE-057: Correction can expire independently

Correction Request может иметь собственный Due Date и Expiration Decision.

---

## HE-058: One missed homework must not cancel learning progress

Expiration одного Assignment не отменяет:

- прошлый Progress;
- завершенные Goals;
- Achievements;
- Lesson history;
- Song Readiness.

---

## HE-059: Expired homework is not negative assessment by itself

Reason Code: `HOMEWORK_EXPIRATION_NOT_ASSESSMENT`

---

## HE-060: Expiration may create review evidence, not skill evidence

Teacher может учитывать историю Assignment в образовательном обсуждении.

Но системный статус Expired не является автоматическим доказательством отсутствия Skill.

---

## HE-061: Homework expiration must not award or revoke achievements

Achievement Award Policy принимает отдельное решение.

---

## HE-062: Homework expiration must not alter Goal automatically

Goal Completion или Goal Cancellation определяются собственными критериями.

---

## HE-063: Expiration stops ordinary reminders

После перехода в Expired pending Homework Reminders отменяются или подавляются.

Command: `SuppressHomeworkReminders`

---

## HE-064: Overdue may still receive limited reminders

Для Overdue Reminder Policy может создавать:

- мягкое напоминание;
- предложение запросить перенос;
- предложение сообщить о Blocker;
- Teacher Attention.

Частота должна быть ограничена.

---

## HE-065: Expired homework should not generate repeated escalation

После Expiration система не должна продолжать:

- ежедневные Reminders;
- давление;
- множественные Teacher alerts;
- сообщения Guardian без отдельной политики.

---

## HE-066: Student-facing language must be neutral

Недопустимо:

> Вы провалили домашнее задание.

Допустимо:

> Срок задания завершился, и оно больше не считается активным. При необходимости преподаватель может открыть его повторно или назначить новую версию.

---

## HE-067: Overdue language must reflect uncertainty

Если нет Submission в системе, корректно:

> В системе пока нет отправленного результата.

Некорректно:

> Вы не выполняли задание.

---

## HE-068: Expiration reason must be explainable

Student должен понимать, было ли Homework закрыто потому что:

- завершился срок;
- закончился связанный Concert;
- Assignment заменено;
- Goal изменилась;
- Teacher отменил задание;
- материал устарел.

---

## HE-069: Private reasons must not leak

В Student explanation не должны попадать:

- private Teacher notes;
- внутренние оценки;
- медицинские сведения;
- служебные комментарии;
- технические Reason Codes.

---

## HE-070: Teacher can override automatic expiration where allowed

Teacher может сохранить Homework активным, если оно остается полезным.

Override должен содержать:

- reason;
- new status;
- optional new due date;
- expiration review date;
- Actor;
- timestamp.

---

## HE-071: Teacher override cannot rewrite history

Наступивший Due Date и предыдущие Decisions сохраняются.

---

## HE-072: AI cannot expire homework independently

AI может:

- предложить outcome;
- обнаружить потерю контекста;
- найти устаревшее Assignment;
- подготовить explanation.

AI не может:

- финально Expire;
- менять Due Date;
- отменять Homework;
- определять вину Student;
- создавать педагогический смысл без Teacher.

Reason Code: `AI_CANNOT_FINALIZE_HOMEWORK_EXPIRATION`

---

## HE-073: Automatic expiration requires deterministic criteria

Без Human Review Homework может быть Expired автоматически только при явно утвержденных условиях.

Примеры:

- Assignment уже Replaced;
- связанный временный ресурс закончился;
- hard event deadline прошел;
- Concert-specific task потеряло единственную цель;
- срок архивной кампании завершен.

---

## HE-074: Ambiguous cases require Teacher Review

Decision: Teacher Review Required

Reason Code: `HOMEWORK_EXPIRATION_CONTEXT_AMBIGUOUS`

---

## HE-075: Expiration must be idempotent

Повторная обработка события не должна создавать:

- второе Expired event;
- повторное Notification;
- повторную Reminder cancellation;
- новую идентичную version.

Reason Code: `HOMEWORK_EXPIRATION_ALREADY_PROCESSED`

---

## HE-076: Concurrent submission must win over stale expiration evaluation

Если Submission произошла до фиксации Expiration, система должна повторно загрузить актуальное состояние.

Старая Evaluation не должна перезаписать Submitted.

Reason Code: `HOMEWORK_EXPIRATION_VERSION_CONFLICT`

---

## HE-077: Backend is authoritative

Client не должен самостоятельно определять окончательный Expired status только по локальному времени.

---

## HE-078: Timezone must be explicit

Due Date должна интерпретироваться в указанной Timezone.

Reason Code: `HOMEWORK_DUE_TIMEZONE_REQUIRED`

---

## HE-079: Date-only deadlines require a defined end-of-day rule

Если Due Date содержит только дату, система должна иметь утвержденную интерпретацию.

Например:

`Due At: 23:59:59 Student Local Time`

или:

`Due At: Start of Related Lesson`

---

## HE-080: Timezone change does not rewrite historical deadlines

Изменение Timezone Student влияет на будущие расчеты согласно правилам, но не должно бесследно менять историю.

---

## HE-081: Daylight-saving changes must be handled deterministically

Для регионов с переходом времени хранится:

- timezone identifier;
- resolved UTC timestamp;
- local representation.

---

## HE-082: Scheduled evaluation is not the sole source of truth

Если background job запустился поздно, Policy оценивает актуальное состояние, а не предполагает, что Assignment все еще просрочено.

---

## HE-083: Expiration event must reference causation

Событие должно включать:

- Due Date event;
- context change event;
- Teacher decision;
- replacement event;
- related event completion.

---

## HE-084: Bulk expiration requires per-assignment decisions

Даже при массовом архивировании каждое Homework должно получить отдельный explainable result или применимый batch reason.

---

## HE-085: Bulk actions must protect active submissions

Homework со статусами:

- Submitted;
- Under Review;
- Correction Requested;
- Clarification Required;

не должны быть закрыты массово без специальных правил.

---

## HE-086: Archiving is separate from expiration

Expired Homework может быть Archived позже.

Archive отвечает за представление и retention, а не за педагогическое решение.

---

## HE-087: Archived homework remains auditable

Сохраняются:

- Assignment;
- versions;
- Due Dates;
- Submissions;
- Decisions;
- Reviews;
- Events;
- Actor metadata.

---

## HE-088: Deletion is not an expiration outcome

Удаление применяется только по отдельной retention/privacy policy.

---

# Deadline Types

## Soft Deadline

Предпочтительный срок.

После него Homework обычно становится Overdue, но остается активным.

## Hard Deadline

После срока исходная цель может быть недостижима.

Примеры:

- задача до Concert;
- материал до репетиции;
- заявка в ограниченное окно;
- упражнение до Assessment session.

## Lesson Deadline

Срок связан со стартом или завершением Lesson.

## Review Deadline

Срок нужен Teacher для своевременной проверки.

## Resource Deadline

Срок связан с доступностью внешнего ресурса.

## Open-Ended

Due Date отсутствует.

Такое Homework не может истечь только по времени без дополнительного правила.

---

# Expiration Strategy

Homework Assignment может иметь:

```text
ExpirationStrategy
├── StrategyType
├── DueDateBehavior
├── GracePeriod
├── AutoExpireAllowed
├── TeacherReviewRequired
├── RelatedContextBehavior
├── ReminderBehavior
├── LateSubmissionAllowed
└── ArchiveAfter
```

Допустимые Strategy Type:

- Manual Review;
- Soft Deadline;
- Hard Deadline;
- Lesson Bound;
- Event Bound;
- Goal Bound;
- Open Ended;
- Replace on Context Change.

---

# Soft Deadline Strategy

Обычно:

```text
Due Date Reached
      |
      v
Mark Overdue
      |
      v
Limited Reminder or Teacher Review
```

---

# Hard Deadline Strategy

Обычно:

```text
Due Date Reached
      |
      v
Check Submission and Context
      |
      +--> Submitted
      |
      +--> Grace Period
      |
      +--> Expire
      |
      +--> Teacher Review
```

---

# Lesson-Bound Strategy

```text
Related Lesson changed?
      |
      +--> Recalculate due date
      |
      +--> Keep original date
      |
      +--> Teacher Review
```

---

# Event-Bound Strategy

Для Concert, rehearsal или другого события.

После завершения Event проверяется, сохраняет ли Homework самостоятельную образовательную ценность.

---

# Open-Ended Strategy

Homework остается доступным до:

- Completion;
- Cancellation;
- Replacement;
- manual expiration;
- context loss.

---

# Decision Matrix

| Condition | Typical Outcome |
| --- | --- |
| Soft deadline reached, homework still useful | Mark Overdue |
| Hard deadline reached, context ended | Expire |
| Submitted before decision | Keep Submitted |
| Submitted during grace period | Accept Submission |
| Student requested clarification | Teacher Review |
| Required material unavailable | Extend or Review |
| Lesson moved by school | Recalculate Due Date |
| Learning Pause active | Defer or Extend |
| Homework replaced | Mark Replaced |
| Goal cancelled | Review Context |
| Concert completed | Expire or Convert |
| Optional practice | Keep Available |
| Teacher explicitly extends | Extend Due Date |
| Long-inactive open-ended homework | Periodic Review |
| Ambiguous educational value | Teacher Review |

---

# Expiration Evaluation Flow

```text
Expiration trigger received
        |
        v
Load current Homework Assignment
        |
        v
Validate Assignment and Version
        |
        +--> Rejected
        |
        v
Check terminal state
        |
        +--> Completed
        +--> Cancelled
        +--> Replaced
        +--> Expired
        |
        v
Check current Submission
        |
        +--> Submitted / Under Review
        |
        v
Resolve deadline strategy
        |
        v
Resolve Due Date, timezone and grace period
        |
        v
Evaluate blockers and learning pause
        |
        +--> Deferred
        +--> Teacher Review
        |
        v
Evaluate related context
        |
        +--> Lesson
        +--> Goal
        +--> Song
        +--> Concert
        |
        v
Determine outcome
        |
        +--> Keep Active
        +--> Mark Overdue
        +--> Extend
        +--> Replace
        +--> Cancel
        +--> Expire
        |
        v
Persist versioned decision
        |
        v
Publish domain event
        |
        +--> Reminder Policy
        +--> Notification Policy
        +--> Periodic Review Policy
```

---

# Commands Produced

## MarkHomeworkOverdue

Переводит Assignment в Overdue.

## ExpireHomework

Переводит Assignment в Expired.

## ExtendHomeworkDueDate

Создает новую версию срока.

## KeepHomeworkActive

Фиксирует решение сохранить Assignment активным.

## RequestHomeworkExpirationReview

Создает ручной Review.

## RequestTeacherHomeworkDecision

Запрашивает педагогическое решение.

## CancelHomework

Отменяет Assignment.

## ReplaceHomework

Связывает старое Homework с новым.

## ReopenHomework

Возвращает Expired Homework в активное состояние.

## SuppressHomeworkReminders

Останавливает pending Reminders после terminal outcome.

## RecalculateHomeworkReminderPlan

Перестраивает план после Extension или Reopen.

## NotifyStudentAboutHomeworkStatus

Передает разрешенное изменение в Notification Policy.

## ArchiveHomework

Переводит завершенное lifecycle-состояние в архивное представление.

---

# Domain Events

```text
HomeworkExpirationEvaluationStarted
HomeworkKeptActive
HomeworkMarkedOverdue
HomeworkGracePeriodStarted
HomeworkGracePeriodEnded
HomeworkDueDateExtensionRequested
HomeworkDueDateExtended
HomeworkExpirationReviewRequested
HomeworkTeacherDecisionRequested
HomeworkExpired
HomeworkReopened
HomeworkCancelled
HomeworkReplaced
HomeworkReminderSuppressionRequested
HomeworkReminderRecalculationRequested
HomeworkArchived
HomeworkLateSubmissionAccepted
HomeworkLateSubmissionRejected
```

## HomeworkMarkedOverdue Event

Событие должно содержать:

- HomeworkAssignmentId;
- HomeworkVersion;
- StudentId;
- DueDate;
- GracePeriodEnd;
- DeadlineType;
- ReasonCodes;
- PolicyId;
- PolicyVersion;
- EffectiveAt;
- CorrelationId;
- CausationId.

## HomeworkExpired Event

Событие должно содержать:

- HomeworkAssignmentId;
- HomeworkVersion;
- StudentId;
- PreviousStatus;
- ExpiredAt;
- ExpirationCategory;
- RelatedContextReferences;
- DecisionId;
- ReasonCodes;
- EvaluatedBy;
- PolicyId;
- PolicyVersion;
- CorrelationId;
- CausationId.

Событие не должно содержать private Teacher notes.

## HomeworkDueDateExtended Event

Должно содержать:

- HomeworkAssignmentId;
- PreviousDueDate;
- NewDueDate;
- Timezone;
- ExtensionReason;
- RequestedBy;
- ApprovedBy;
- PolicyId;
- PolicyVersion;
- EffectiveAt.

---

# Late Submission Handling

Late Submission требует отдельного решения.

Возможные outcomes:

- Accepted
- Accepted with Teacher Review
- Accepted as Practice Evidence Only
- Rejected Because Context Expired
- Replacement Required
- Manual Review Required

---

# Late Submission Rules

## Submission during active grace period

Обычно принимается.

## Submission after soft deadline

Может приниматься, если Assignment активно.

## Submission after hard contextual deadline

Может быть:

- отклонена как выполнение исходного Assignment;
- сохранена как Practice Evidence;
- направлена Teacher;
- использована для нового Homework;
- сохранена без Progress effect.

## Submission after Expired

Не должна автоматически менять Homework в Submitted.

Необходимо:

- Reopen;
- Teacher Review;
- или создать новое Assignment context.

## Late submission must not be discarded silently

Даже если Submission не может выполнить исходное Homework, Student должен получить объяснение.

---

# Teacher Review

Teacher Review обязателен, если:

- срок прошел, но ценность Homework неясна;
- Student запросил Extension;
- Learning Pause пересекает Due Date;
- связанный Lesson перенесен или отменен;
- Goal завершена или отменена;
- Song Version изменилась;
- Concert отменен или завершен;
- Student сообщил Blocker;
- отсутствовал необходимый материал;
- поздняя Submission поступила после Expired;
- требуется Reopen;
- Assignment является чувствительным;
- автоматическая Expiration может повлиять на образовательный план;
- AI предложил Expiration.

---

# Teacher Review Result

Teacher может:

- оставить активным;
- Mark Overdue;
- установить Grace Period;
- продлить срок;
- принять Late Submission;
- запросить Correction;
- заменить Homework;
- отменить;
- Expire;
- Reopen;
- преобразовать в Optional Practice;
- связать с новым Lesson;
- связать с новой Goal;
- дать Student объяснение.

---

# Extension Model

```text
HomeworkDueDateExtension
├── ExtensionId
├── HomeworkAssignmentId
├── PreviousDueDate
├── NewDueDate
├── Reason
├── RequestedBy
├── ApprovedBy
├── RequestedAt
├── ApprovedAt
├── Status
├── StudentVisibleExplanation
└── AuditMetadata
```

Допустимые статусы:

- Requested;
- Approved;
- Rejected;
- Cancelled;
- Superseded.

---

# Reopen Model

```text
HomeworkReopenDecision
├── ReopenDecisionId
├── HomeworkAssignmentId
├── PreviousExpiredAt
├── NewDueDate
├── NewHomeworkVersion
├── ReopenReason
├── RelatedContext
├── ReopenedBy
├── ReopenedAt
└── AuditMetadata
```

---

# Reminder Policy Integration

Homework Expiration Policy отвечает:

> Остается ли Assignment активным после Due Date?

Homework Reminder Policy отвечает:

> Следует ли создавать или отправлять Reminder для текущего состояния?

Типичные реакции:

```text
Mark Overdue
    |
    +--> limited overdue reminder

Extend Due Date
    |
    +--> recalculate reminder plan

Expire Homework
    |
    +--> suppress all ordinary reminders

Reopen Homework
    |
    +--> create new reminder plan
```

---

# Notification Policy Integration

Homework Expiration Policy не отправляет сообщения напрямую.

Notification Policy определяет:

- канал;
- время;
- bundling;
- localization;
- privacy-safe preview;
- retry;
- frequency limits.

Возможные Student notifications:

- Homework стало Overdue;
- срок продлен;
- Homework закрыто;
- требуется решение Teacher;
- Late Submission принята;
- Homework заменено;
- Homework открыто повторно.

---

# Periodic Review Policy Integration

Periodic Review Policy может находить:

- давно Overdue Homework;
- Open-Ended Homework без активности;
- просроченные Grace Period;
- Homework с активным Blocker;
- Assignments после завершенной Goal;
- старые Homework после смены Song;
- Expired Homework, готовые к Archive;
- Submission, слишком долго ожидающие Review.

Periodic Review не принимает педагогическое решение вместо Homework Expiration Policy.

---

# Progress Integration

Допустимые Progress inputs:

- reviewed Homework;
- accepted Submission;
- Assessment;
- Teacher observation.

Недопустимые автоматические inputs:

- Overdue;
- Expired;
- Extension request;
- Reminder count;
- Notification open rate;
- lateness duration.

---

# Goal Integration

Homework может быть одним из Evidence для Goal.

Но:

- Expiration Homework не отменяет Goal;
- Goal Completion не означает автоматический Completion Homework;
- Goal Cancellation не удаляет историю Homework;
- Reopened Homework не переоткрывает завершенную Goal автоматически.

---

# Song Readiness Integration

Homework может поддерживать Song Readiness.

Expired Assignment:

- перестает быть ожидаемой будущей работой;
- остается в истории;
- не удаляет ранее подтвержденное Evidence;
- не снижает Song Readiness автоматически.

---

# Concert Integration

Concert-specific Homework может иметь Hard Deadline.

Примеры:

- загрузить backing track;
- выучить финальную версию;
- подготовить сценический выход;
- записать demo до репетиции.

Но технические Concert tasks могут относиться к Concert domain, а не к образовательному Homework. Граница должна быть явно определена.

---

# Student Presentation

Student должен видеть:

- Homework title;
- Current Status;
- Due Date;
- Grace Period, если есть;
- допускается ли Late Submission;
- измененный срок;
- причину закрытия;
- следующий шаг;
- возможность запросить Extension;
- возможность сообщить о Blocker;
- возможность открыть replacement;
- Teacher explanation.

Следует избегать:

- красных наказательных индикаторов;
- слов «провал»;
- сравнений;
- потери streak;
- угроз;
- скрытого ухудшения Progress;
- неоднозначного статуса.

---

# Teacher Presentation

Teacher должен видеть:

- Assignment;
- Deadline Type;
- Due Date history;
- Grace Period;
- Submission history;
- Student blocker;
- Learning Pause;
- related Lesson;
- related Goal;
- related Song;
- related Concert;
- Reminder history;
- current Expiration Decision;
- recommended actions;
- pending Review;
- AI proposal metadata.

---

# Administrator Presentation

Administrator может видеть:

- lifecycle status;
- Due Date;
- pending Reviews;
- delivery-related consequences;
- bulk stale records;
- technical errors;
- archive eligibility.

Administrator не должен принимать педагогические решения без полномочий.

---

# Owner Analytics

Owner может видеть агрегированно:

- число Active / Overdue / Expired Homework;
- среднее время до Submission;
- частоту Extensions;
- причины Expiration;
- долю Late Submissions;
- число Clarification cases;
- число Assignment replacements;
- Teacher Review backlog;
- давность Overdue records;
- Reminder suppression after Expiration;
- количество Homework без deadline strategy.

Эти данные не должны использоваться для:

- публичного рейтинга Student;
- автоматического наказания;
- вывода о мотивации;
- оценки Teacher только по числу Expired Homework;
- давления на completion metrics.

---

# AI Assistance

AI может:

- находить устаревшие Assignment;
- определять возможную потерю контекста;
- предлагать outcome;
- предлагать Student explanation;
- группировать Reason Codes;
- находить конфликт сроков;
- предлагать replacement;
- обнаруживать Homework без стратегии;
- находить старые Overdue records;
- предлагать Teacher Review.

AI не может:

- финально Expire;
- Mark Overdue без Policy;
- менять Due Date;
- принимать Late Submission;
- определять виновность;
- ухудшать Progress;
- отменять Goal;
- отзывать Achievement;
- создавать medical conclusions;
- скрывать неопределенность.

AI proposal должна сохранять:

- model or mechanism;
- version;
- input references;
- proposed outcome;
- proposed reason;
- confidence;
- timestamp;
- human confirmation status.

---

# Privacy

Expiration data может раскрывать:

- Learning Pause;
- личные причины Extension;
- проблемы с материалом;
- расписание;
- участие в Concert;
- педагогические трудности;
- внутренние Teacher notes.

Необходимо:

- минимизировать Student-facing explanation;
- отделять private reason от public reason;
- ограничивать Administrator access;
- защищать minor data;
- не включать sensitive details в Notification;
- хранить Extension reason с подходящей visibility;
- не экспортировать личные причины в Owner analytics.

---

# Security

Необходимо защищать:

- Expiration чужого Homework;
- подмену StudentId;
- неавторизованный Extension;
- изменение Due Date задним числом;
- подделку Submission timestamp;
- неавторизованный Reopen;
- массовое Expire active Submission;
- обход version checks;
- удаление Audit;
- подмену Timezone;
- повторную обработку Events;
- раскрытие private reasons.

---

# Audit Requirements

Для каждой оценки сохраняются:

- PolicyId;
- PolicyVersion;
- DecisionId;
- HomeworkAssignmentId;
- HomeworkVersion;
- StudentId;
- PreviousStatus;
- CurrentStatus;
- Deadline Type;
- Due Date;
- Timezone;
- Grace Period;
- Educational Validity End;
- related Context References;
- active Submission Reference;
- Blocker Reference;
- Learning Pause Reference;
- ActorId;
- Outcome;
- Reason Codes;
- Student Explanation;
- Human Review flag;
- AI metadata;
- EvaluatedAt;
- EffectiveAt;
- CorrelationId;
- CausationId.

Для Extension:

- previous Due Date;
- new Due Date;
- requested by;
- approved by;
- reason;
- timestamps.

Для Expiration:

- ExpiredAt;
- Expiration Category;
- terminal reason;
- Reminder cancellation references;
- Notification reference;
- Archive eligibility.

Для Reopen:

- prior Expiration Decision;
- new Homework Version;
- new Due Date;
- Reopen reason;
- Actor;
- Reminder recalculation reference.

---

# Failure Modes

## Homework not found

- Decision: Rejected
- Reason Code: HOMEWORK_NOT_FOUND

## Homework version missing

- Decision: Rejected
- Reason Code: HOMEWORK_VERSION_REQUIRED

## Student mismatch

- Decision: Rejected
- Reason Code: HOMEWORK_EXPIRATION_STUDENT_MISMATCH

Security Audit обязателен.

## Due date missing for deadline strategy

- Decision: Teacher Review Required
- Reason Code: HOMEWORK_DUE_DATE_REQUIRED

## Timezone unknown

- Decision: Deferred (или используется документированный default)
- Reason Code: HOMEWORK_DUE_TIMEZONE_REQUIRED

## Homework already completed

- Decision: No Change Required
- Reason Code: HOMEWORK_ALREADY_COMPLETED

## Homework already expired

- Decision: No Change Required
- Reason Code: HOMEWORK_EXPIRATION_ALREADY_PROCESSED

## Submission exists

- Decision: Keep Active (или переход в Review state)
- Reason Code: HOMEWORK_SUBMITTED_PENDING_REVIEW

## Active learning pause

- Decision: Deferred or Teacher Review Required
- Reason Code: STUDENT_LEARNING_PAUSE_ACTIVE

## Required material unavailable

- Decision: Teacher Review Required
- Reason Code: HOMEWORK_REQUIRED_MATERIAL_UNAVAILABLE

## Related lesson moved

- Decision: Teacher Review Required or Extend Due Date
- Reason Code: RELATED_LESSON_RESCHEDULED

## Related concert completed

- Decision: Expire Homework or Teacher Review Required
- Reason Code: RELATED_CONCERT_CONTEXT_COMPLETED

## Related goal cancelled

- Decision: Teacher Review Required
- Reason Code: RELATED_GOAL_CANCELLED

## Homework replaced

- Decision: No Change Required
- Current Status: Replaced
- Reason Code: HOMEWORK_REPLACED

## Ambiguous context

- Decision: Teacher Review Required
- Reason Code: HOMEWORK_EXPIRATION_CONTEXT_AMBIGUOUS

## Duplicate trigger

- Decision: No Change Required
- Reason Code: HOMEWORK_EXPIRATION_ALREADY_PROCESSED

## Concurrent submission

- Decision: Deferred
- Reason Code: HOMEWORK_EXPIRATION_VERSION_CONFLICT

Политика повторно оценивается.

## Unauthorized extension

- Decision: Rejected
- Reason Code: HOMEWORK_EXTENSION_NOT_AUTHORIZED

## Reopen without relevance

- Decision: Rejected
- Reason Code: HOMEWORK_REOPEN_CONTEXT_INVALID

---

# Explainability Examples

## Marked Overdue

> Срок задания прошел, но оно остается активным и его все еще можно отправить преподавателю.

## Grace Period

> Основной срок завершился, но задание можно отправить до 30 июля включительно.

## Extended

> Срок задания перенесен на 2 августа после изменения даты следующего занятия.

## Learning Pause

> Срок не считается просроченным во время учебной паузы. После возвращения преподаватель подтвердит, остается ли задание актуальным.

## Expired After Concert

> Задание было связано с подготовкой к прошедшему концерту и больше не считается активным. Результаты предыдущей работы сохранены.

## Replaced

> Преподаватель заменил это задание новой версией. Старое задание осталось в истории, но выполнять его больше не требуется.

## Late Submission Accepted

> Задание отправлено после первоначального срока, но преподаватель все еще может его проверить.

## Late Submission Requires Review

> Задание было отправлено после закрытия. Преподаватель решит, открыть ли его повторно или использовать результат как дополнительную практику.

## Optional Homework

> Это дополнительная практика без жесткого срока. Она остается доступной и не отмечается как просроченная.

---

# Examples

## Example 1: Soft deadline reached

Дано:

- Homework Required;
- Deadline Type: Soft;
- Due Date прошла;
- Submission отсутствует;
- Assignment остается полезным.

Результат:

- Decision: Mark Overdue
- New Status: Overdue
- Reason Code: SOFT_DEADLINE_REACHED

## Example 2: Student submitted shortly after due date

Дано:

- Due Date: 18:00;
- Submission: 19:30;
- Grace Period: 24 hours;
- Homework активно.

Результат:

- Decision: Late Submission Accepted
- Status: Submitted
- Reason Code: SUBMITTED_WITHIN_GRACE_PERIOD

## Example 3: Concert preparation task after concert

Дано:

- Homework: подготовить финальный сценический выход;
- Concert завершен;
- Submission отсутствует;
- другой образовательной цели нет.

Результат:

- Decision: Expire Homework
- New Status: Expired
- Reason Code: RELATED_CONCERT_CONTEXT_COMPLETED

## Example 4: Lesson rescheduled by teacher

Дано:

- Due Date была связана со следующим Lesson;
- Lesson перенесен на неделю;
- Student еще не отправил Homework.

Результат:

- Decision: Extend Due Date
- Reason Code: RELATED_LESSON_RESCHEDULED

Student не получает негативный Overdue status.

## Example 5: Student on learning pause

Дано:

- Learning Pause началась до Due Date;
- Pause все еще активна;
- Homework не имеет hard event deadline.

Результат:

- Decision: Deferred
- Reason Code: STUDENT_LEARNING_PAUSE_ACTIVE

## Example 6: Missing backing track

Дано:

- Homework требует backing track;
- файл был недоступен;
- Student сообщил об этом до срока.

Результат:

- Decision: Teacher Review Required
- Reason Code: HOMEWORK_REQUIRED_MATERIAL_UNAVAILABLE

Автоматический Overdue не применяется.

## Example 7: Homework replaced

Дано:

- Teacher создал новое Assignment;
- старое Homework явно заменено.

Результат:

- Decision: Replace Homework
- New Status: Replaced
- Reason Code: HOMEWORK_REPLACED

## Example 8: Submission arrives after expiration

Дано:

- Homework уже Expired;
- Student отправил результат;
- Assignment все еще может быть педагогически полезным.

Результат:

- Decision: Teacher Review Required
- Reason Code: LATE_SUBMISSION_AFTER_EXPIRATION

Teacher может Reopen или сохранить Submission как Practice Evidence.

## Example 9: Optional exercise without due date

Дано:

- Homework Optional;
- Due Date отсутствует;
- материал актуален.

Результат:

- Decision: Keep Active
- Reason Code: OPEN_ENDED_OPTIONAL_PRACTICE

## Example 10: Goal cancelled

Дано:

- Homework существовало только для конкретной Goal;
- Goal отменена;
- Assignment больше не имеет самостоятельной цели.

Результат:

- Decision: Cancel Homework
- Reason Code: RELATED_GOAL_CANCELLED

## Example 11: Correction requested

Дано:

- Student отправил Homework;
- Teacher запросил исправление;
- новый Due Date еще не задан.

Результат:

- Decision: Teacher Review Required
- Reason Code: CORRECTION_DUE_DATE_REQUIRED

## Example 12: AI proposes expiration

Дано:

- AI обнаружил старое Homework;
- Assignment открыто 90 дней;
- Teacher decision отсутствует;
- Context неоднозначен.

Результат:

- Decision: Teacher Review Required
- Reason Code: AI_CANNOT_FINALIZE_HOMEWORK_EXPIRATION

---

# Test Requirements

## Basic Expiration Tests

- Due Date alone does not automatically expire;
- Soft Deadline produces Overdue;
- Hard contextual deadline can produce Expired;
- Completed Homework remains Completed;
- Cancelled Homework does not become Expired;
- Replaced Homework stays Replaced;
- duplicate evaluation is idempotent.

## Submission Tests

- Submission before Due Date prevents Expiration;
- Submission during Grace Period is accepted;
- Submission after Soft Deadline can be accepted;
- Submission after Expired requires Review;
- Submitted Homework awaiting Teacher Review remains Submitted;
- late submission timestamp is preserved;
- stale Expiration cannot overwrite concurrent Submission.

## Grace Period Tests

- configured Grace Period starts after Due Date;
- Grace Period does not rewrite Due Date;
- Submission within Grace Period is accepted;
- Grace Period end triggers reevaluation;
- absent Grace Period is not invented;
- changed Due Date recalculates Grace Period.

## Extension Tests

- Student can request Extension;
- authorized Teacher can approve;
- unauthorized Actor is rejected;
- previous Due Date is preserved;
- new Due Date is versioned;
- Extension after Expired requires Reopen;
- Extension recalculates Reminder Plan.

## Learning Pause Tests

- active Pause suppresses automatic Overdue where configured;
- pause ending triggers relevance review;
- old Homework is not automatically resumed;
- hard event deadline can still require Review;
- pause reason is not exposed unnecessarily;
- long pause can lead to replacement.

## Lesson Integration Tests

- Lesson reschedule can extend Due Date;
- school-caused reschedule does not penalize Student;
- Lesson cancellation triggers Review;
- completed Lesson does not always expire Homework;
- new related Lesson can restore relevance;
- unrelated Lesson change has no effect.

## Goal Integration Tests

- completed Goal triggers context review;
- cancelled Goal can cancel associated Homework;
- Homework Expiration does not cancel Goal;
- Homework Reopen does not reopen Goal;
- Goal-independent practice remains active.

## Song Integration Tests

- Song removal triggers context review;
- changed key can invalidate version-specific Homework;
- archived Song can expire associated Assignment;
- general Skill Homework can remain active;
- Expiration does not lower Song Readiness.

## Concert Integration Tests

- Concert-specific hard deadline expires when context ends;
- cancelled Concert triggers Review;
- completed Concert does not automatically expire general practice;
- changed Concert date can extend Due Date;
- withdrawn Participation can invalidate concert-only Homework.

## Blocker Tests

- missing material prevents automatic negative outcome;
- pending Clarification triggers Review;
- Teacher delay does not penalize Student;
- technical failure can justify Extension;
- resolved Blocker triggers reevaluation;
- duplicate Blocker does not duplicate Review requests.

## Reminder Integration Tests

- Expired Homework suppresses pending Reminders;
- Overdue Homework can retain limited Reminder;
- Extended Homework recalculates Plan;
- Reopened Homework creates new Plan;
- Replaced Homework cancels old Reminders;
- duplicate suppression is idempotent.

## Permission Tests

- authorized Teacher can extend;
- authorized Teacher can Expire;
- Student can request Extension;
- Student cannot self-approve Extension where approval is required;
- Administrator cannot make pedagogical decision by default;
- AI cannot finalize Expiration;
- Owner cannot rewrite Homework history.

## Versioning Tests

- every decision stores Homework Version;
- old decision remains reproducible;
- Due Date history is preserved;
- Reopen creates a new version;
- concurrent updates trigger retry;
- stale Evaluation cannot overwrite current state;
- Policy Version is stored.

## Privacy Tests

- Student explanation excludes private notes;
- Extension reason visibility is respected;
- Owner analytics excludes personal reasons;
- Notification preview is minimal;
- Guardian does not receive data without permission;
- event payload excludes sensitive content.

## Explainability Tests

- Overdue outcome is understandable;
- Expired outcome states why;
- Extension states new date and reason;
- Replacement links to new Assignment;
- Late Submission outcome is explained;
- Learning Pause explanation does not reveal sensitive details;
- internal Reason Codes are not shown directly.

## AI Tests

- AI can propose stale Homework review;
- AI cannot Expire directly;
- AI cannot infer Student laziness;
- AI cannot invent related context;
- hallucinated Evidence is rejected;
- AI proposal stores source references;
- ambiguous AI proposal results in Human Review.

---

# Non-Goals

Homework Expiration Policy не определяет:

- содержание Homework;
- качество Submission;
- Homework Assessment;
- Homework Completion criteria;
- Progress Calculation;
- Goal Completion;
- Song Readiness;
- Concert Eligibility;
- Achievement Award;
- Notification transport;
- Reminder delivery implementation;
- CRM status;
- финансовые штрафы;
- оплату;
- посещаемость;
- дисциплинарные меры;
- рейтинг Student;
- медицинские решения;
- legal data retention deletion.

---

# Open Questions

Необходимо определить:

- какие Deadline Types входят в MVP;
- используется ли Grace Period по умолчанию;
- какова стандартная продолжительность Grace Period;
- какие Homework могут Auto Expire;
- какие всегда требуют Teacher Review;
- показывается ли статус Overdue Student;
- должны ли Optional Homework иметь Due Date;
- нужен ли статус Available;
- может ли Student отправлять после Expired;
- как хранить Practice Evidence после Expired;
- кто может Reopen Homework;
- требуется ли новая версия при Reopen;
- можно ли Reopen несколько раз;
- максимальное число Extensions;
- может ли Extension быть автоматическим;
- как обрабатывать праздники;
- как учитывать выходные;
- как интерпретировать date-only Due Date;
- какой timezone является authoritative;
- что происходит при смене Timezone;
- нужно ли переносить сроки при Lesson reschedule автоматически;
- что происходит при отмене Lesson;
- что происходит после Learning Pause;
- как долго хранить Overdue Homework активным;
- когда Periodic Review должен эскалировать старое Homework;
- когда Expired Homework архивируется;
- сколько хранить Expiration Decisions;
- может ли Teacher массово Expire Homework;
- какие защиты нужны для bulk action;
- как обрабатывать Group Homework;
- может ли часть группы иметь Extension;
- как обрабатывать дуэтные или совместные Assignments;
- что происходит при смене Teacher;
- может ли новый Teacher Reopen старое Homework;
- как обрабатывать Homework, связанное с несколькими Goals;
- как обрабатывать Homework, связанное с несколькими Songs;
- является ли Concert task образовательным Homework;
- нужна ли отдельная Task domain model;
- как отличать Hard Deadline от Soft Deadline;
- кто определяет Deadline Type;
- может ли AI предложить Deadline Type;
- когда AI proposal требует Teacher confirmation;
- как уведомлять о Late Submission;
- нужно ли уведомлять Teacher о каждом Overdue;
- когда создавать Teacher Attention;
- как избежать большого Review backlog;
- как учитывать отсутствие school-provided material;
- нужно ли автоматически продлевать срок после восстановления материала;
- можно ли отменить Expiration Decision;
- чем Reopen отличается от нового Assignment;
- как связывать replacement Homework;
- что происходит с Reminder history после Reopen;
- влияет ли Expired Homework на Owner analytics;
- как предотвратить использование Expired count как наказательной метрики;
- нужен ли отдельный Homework Lifecycle document;
- нужна ли отдельная Late Submission Policy;
- нужна ли отдельная Extension Policy;
- нужна ли отдельная Learning Pause Policy;
- нужна ли отдельная Archive and Retention Policy.

---

# History

| Version | Description |
| --- | --- |
| 1.0.0 | Определены правила Overdue, Grace Period, Extension, Expiration, Replacement и Reopen для Homework. |
