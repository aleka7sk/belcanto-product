---
Status: Accepted
Version: 1.0.0
Last Updated: 2026-07-27

Policy Id: LESSON_COMPLETION_POLICY
Policy Type:
  - Validation Policy
  - Reaction Policy

Owners:
  - Product Owner
  - Education Lead
  - Technical Lead

Related Aggregates:
  - Lesson
  - Student
  - Teacher
  - Homework
  - Assessment
  - Attendance

Observed Events:
  - LessonStarted
  - LessonResultRecorded
  - AttendanceRecorded
  - AssessmentPublished
  - HomeworkAssigned
  - LessonCompletionRequested

Produced Commands:
  - CompleteLesson
  - RequestLessonResult
  - RequestAttendanceRecord
  - RequestTeacherConfirmation
  - CreateLessonFollowUp
  - RecalculateStudentProgress

Related Documents:
  - 000-domain-policy-overview.md
  - ../lesson.md
  - ../homework.md
  - ../assessment.md
  - ../progress.md
---

# Lesson Completion Policy

> Lesson Completion Policy определяет, при каких условиях занятие может быть переведено в состояние `Completed`.
>
> Политика гарантирует, что завершенный Lesson содержит достаточный образовательный и организационный результат и может стать надежной частью истории ученика.

---

# Purpose

В реальной работе преподаватель может завершить занятие сразу после его проведения, не заполнив все необходимые данные.

Если система просто разрешает сменить статус на `Completed`, история обучения постепенно становится неполной.

В ней могут отсутствовать:

- факт присутствия;
- образовательный результат;
- рекомендации;
- информация о следующем шаге;
- причина досрочного завершения;
- сведения о фактическом преподавателе;
- связь с выполненной работой.

Lesson Completion Policy предотвращает появление формально завершенных, но содержательно пустых занятий.

---

# Core Principle

Статус `Completed` означает:

> Занятие фактически состоялось, его результат зафиксирован, а образовательная история содержит достаточно информации для продолжения работы.

`Completed` не означает только то, что наступило время окончания занятия.

---

# Trigger

Политика применяется при команде `CompleteLesson` или событии `LessonCompletionRequested`.

Политика также может повторно применяться после устранения блокирующего условия.

Например:

- `LessonResultRecorded`;
- `AttendanceRecorded`;
- `TeacherConfirmationProvided`.

---

# Inputs

Для оценки политики требуются:

- LessonId;
- текущий статус Lesson;
- запланированное время;
- фактическое время начала;
- фактическое время окончания;
- TeacherId;
- список участников;
- Attendance;
- Lesson Result;
- тип Lesson;
- формат Lesson;
- причина отклонения от расписания;
- связанные Assessment;
- связанные Homework;
- Actor;
- права Actor;
- версия политики.

Дополнительные данные могут требоваться для отдельных типов занятий.

---

# Lesson Types

Политика должна учитывать тип Lesson.

Поддерживаемые концептуальные типы:

- Individual Lesson;
- Group Lesson;
- Trial Lesson;
- Masterclass;
- Rehearsal;
- Concert Preparation;
- Diagnostic Lesson;
- Online Lesson;
- Replacement Lesson.

Конкретный перечень утверждается отдельно.

Требования к завершению могут различаться в зависимости от типа занятия.

---

# Preconditions

Перед применением политики должны выполняться базовые условия.

- Lesson существует.
- Actor аутентифицирован.
- Actor имеет право завершать данный Lesson.
- Lesson не находится в финальном состоянии.
- Lesson не архивирован.
- Lesson не отменен.
- Версия Lesson не конфликтует с текущим состоянием.
- Все обязательные участники известны.

При нарушении precondition команда отклоняется без дальнейшей оценки образовательных правил.

---

# Final States

К финальным операционным состояниям относятся:

- Completed;
- Cancelled;
- Missed.

Archived является состоянием хранения, а не результатом проведения.

Lesson Completion Policy применяется только для перевода в Completed.

Для Cancelled и Missed должны существовать отдельные правила завершения сценария.

---

# Decision Outcomes

Политика возвращает один из следующих результатов.

## Approved

Lesson может быть завершен.

## Rejected

Завершение запрещено.

Исправление текущих данных не может сделать операцию допустимой без изменения состояния Lesson.

## Missing Required Data

Для завершения не хватает обязательной информации.

Система должна указать конкретные поля или действия.

## Teacher Confirmation Required

Автоматически определить допустимость невозможно.

Требуется явное подтверждение преподавателя.

## Administrative Review Required

Обнаружена ситуация, требующая административного вмешательства.

## Already Completed

Lesson уже завершен.

Повторная команда не должна создавать новых последствий.

---

# Decision Rules

## LC-001: Lesson must be in a completable state

Lesson может быть завершен только из разрешенного состояния.

По умолчанию:

`Started -> Completed`

Дополнительно может быть разрешено:

`Scheduled -> Completed`

только если преподаватель подтверждает, что занятие состоялось, но запуск не был зарегистрирован в системе.

Прямой переход из Scheduled должен фиксироваться в Audit.

---

## LC-002: Cancelled Lesson cannot be completed

Lesson в состоянии Cancelled не может перейти в Completed.

Требуется отдельное административное восстановление или создание нового Lesson.

Reason Code: `LESSON_CANCELLED`

---

## LC-003: Missed Lesson cannot be completed

Lesson, зафиксированный как Missed, не может одновременно считаться Completed.

Если статус был установлен ошибочно, сначала выполняется корректирующее действие.

Reason Code: `LESSON_MARKED_MISSED`

---

## LC-004: Teacher is required

Completed Lesson обязан иметь фактического преподавателя.

Фактический Teacher может отличаться от запланированного при замене.

В таком случае сохраняются:

- planned teacher;
- actual teacher;
- reason for replacement;
- actor who confirmed replacement.

Reason Code: `ACTUAL_TEACHER_REQUIRED`

---

## LC-005: Participant list is required

Lesson обязан содержать минимум одного Student, если тип занятия предполагает участие учеников.

Для группового занятия должен быть зафиксирован список фактических участников.

Reason Code: `LESSON_PARTICIPANT_REQUIRED`

---

## LC-006: Attendance must be recorded

Перед завершением должен быть определен статус посещения каждого ожидаемого Student.

Допустимые концептуальные значения:

- Present;
- Late;
- Left Early;
- Absent;
- Excused;
- Unknown.

Unknown не допускается при завершении без подтверждения Teacher.

Reason Code: `ATTENDANCE_REQUIRED`

---

## LC-007: At least one participant must have attended

Обычный Lesson не может быть завершен, если ни один Student фактически не присутствовал.

В таком случае должен использоваться статус Missed, Cancelled или другой подходящий результат.

Исключения возможны для служебных или подготовительных типов событий, которые не являются Lesson ученика.

Reason Code: `NO_ATTENDING_STUDENTS`

---

## LC-008: Educational result is required

Completed Lesson должен содержать образовательный результат.

Минимальный результат должен отвечать хотя бы на один вопрос:

- над чем работали;
- что изменилось;
- что получилось;
- что требует внимания;
- что делать дальше.

Одной технической записи вроде:

> Урок проведен.

недостаточно.

Reason Code: `LESSON_RESULT_REQUIRED`

---

## LC-009: Result must be meaningful

Lesson Result должен пройти минимальную проверку содержательности.

Проверка может учитывать:

- минимальную длину;
- наличие структурированных элементов;
- отсутствие только шаблонного текста;
- отсутствие пустого копирования предыдущего результата;
- наличие связи с Lesson activity.

Автоматическая проверка не должна оценивать педагогическое качество текста как окончательное решение.

При сомнении система может запросить подтверждение Teacher.

Reason Code: `LESSON_RESULT_NOT_MEANINGFUL`

---

## LC-010: Homework is optional

Homework не является обязательным для завершения каждого Lesson.

Отсутствие Homework не блокирует завершение.

Однако система может запросить объяснение или подтверждение, если:

- программа требует самостоятельной практики;
- предыдущий Homework требует продолжения;
- Teacher указал дальнейшие действия, но не оформил их;
- Lesson относится к концертной подготовке;
- активная Goal требует регулярной практики.

Результат в таком случае: Teacher Confirmation Required

Reason Code: `HOMEWORK_CONFIRMATION_REQUIRED`

---

## LC-011: Formal Assessment is optional

Для завершения Lesson не требуется отдельный Published Assessment по каждому Skill.

Краткий Lesson Result может быть достаточным.

Formal Assessment обязателен только если это предусмотрено:

- типом Lesson;
- программой;
- Periodic Review;
- Diagnostic Lesson;
- завершением Goal;
- концертным этапом;
- отдельной Domain Policy.

Reason Code: `REQUIRED_ASSESSMENT_MISSING`

---

## LC-012: Started time must be known

Если Lesson был начат через систему, фактическое время начала должно быть сохранено.

Если оно отсутствует, Teacher может подтвердить проведение вручную.

Reason Code: `LESSON_START_TIME_REQUIRED`

---

## LC-013: Completion time must be valid

Фактическое время окончания:

- не может быть раньше времени начала;
- не может быть необоснованно далеко в будущем;
- должно соответствовать часовому поясу школы;
- должно быть сохранено в едином формате времени.

Reason Codes: `INVALID_COMPLETION_TIME`, `COMPLETION_TIME_IN_FUTURE`

---

## LC-014: Large duration deviation requires confirmation

Если фактическая продолжительность значительно отличается от запланированной, требуется подтверждение Teacher.

Примеры:

- занятие длилось менее минимально допустимого времени;
- занятие продолжалось значительно дольше;
- Teacher завершает Lesson сразу после начала;
- приложение было оставлено открытым много часов.

Политика не должна автоматически считать отклонение нарушением.

Reason Code: `LESSON_DURATION_CONFIRMATION_REQUIRED`

---

## LC-015: Group Lesson requires per-student attendance

Для Group Lesson общий статус посещения недостаточен.

Attendance должен быть определен отдельно для каждого Student.

При этом Lesson Result может быть:

- общим;
- индивидуальным;
- сочетать оба вида.

Reason Code: `GROUP_ATTENDANCE_INCOMPLETE`

---

## LC-016: Individual sensitive notes are not required

Teacher не обязан создавать закрытые заметки о каждом Student.

Наличие Teacher Only Assessment не является условием завершения.

---

## LC-017: Required attachments must exist when mandated

По умолчанию Attachment не обязателен.

Он становится обязательным только если это определено конкретным сценарием.

Например:

- запись диагностического исполнения;
- контрольная видеозапись;
- материал концертной репетиции;
- подтверждение дистанционного задания.

Reason Code: `REQUIRED_ATTACHMENT_MISSING`

---

## LC-018: Active blocking action must be resolved

Lesson не может быть завершен, если существует блокирующая проблема:

- конфликт участников;
- ошибка идентичности Student;
- спорный фактический Teacher;
- незавершенное изменение расписания;
- security restriction;
- несогласованное объединение двух Lesson.

Reason Code: `LESSON_HAS_BLOCKING_ISSUE`

---

## LC-019: Permissions must be valid at completion time

Actor должен иметь право завершать Lesson в момент выполнения команды.

Право, существовавшее во время планирования, недостаточно.

Обычно Lesson может завершить:

- фактический Teacher;
- уполномоченный replacement Teacher;
- Administrator через специальный corrective flow;
- Owner только при наличии отдельного полномочия.

Reason Code: `LESSON_COMPLETION_NOT_AUTHORIZED`

---

## LC-020: Student cannot complete the Lesson

Student не может самостоятельно перевести Lesson в Completed.

Student может:

- подтвердить участие;
- оставить Self Assessment;
- отправить обратную связь;
- подтвердить отдельные факты, если предусмотрено.

Окончательное завершение выполняется уполномоченным сотрудником.

---

## LC-021: AI cannot complete the Lesson independently

AI не может быть Actor команды `CompleteLesson`.

AI может:

- предложить Lesson Result;
- структурировать заметки;
- выявить пропущенные поля;
- предложить Homework;
- создать Draft Assessment.

Перед завершением Teacher должен подтвердить AI-generated содержание.

Reason Code: `TEACHER_CONFIRMATION_REQUIRED_FOR_AI_CONTENT`

---

## LC-022: Published information must preserve authorship

Если при завершении создаются или публикуются:

- Assessment;
- Homework;
- Recommendation;
- Lesson Result;

каждый объект должен сохранять фактического автора и источник.

AI-assisted текст не должен записываться как полностью созданный Teacher без соответствующей отметки происхождения.

---

## LC-023: Completion is idempotent

Повторная обработка одной команды завершения не должна:

- повторно публиковать события;
- создавать копии Homework;
- дублировать Notification;
- повторно запускать Progress recalculation;
- менять время завершения.

Результат повторной команды: Already Completed

Reason Code: `LESSON_ALREADY_COMPLETED`

---

## LC-024: Completion must use optimistic concurrency

Команда должна учитывать текущую версию Lesson.

Если Lesson изменился после загрузки формы завершения, команда отклоняется с конфликтом версии.

Причина: Teacher может завершать устаревшее состояние, не видя изменения Attendance, состава участников или расписания.

Reason Code: `LESSON_VERSION_CONFLICT`

---

## LC-025: Completion creates immutable historical facts

После успешного завершения должны быть зафиксированы:

- фактический Teacher;
- участники;
- Attendance;
- фактическое время;
- Lesson Result;
- версия политики;
- Actor;
- примененные Reason Codes;
- созданные связанные объекты.

Эти факты не должны бесследно изменяться.

Корректировка выполняется через отдельный versioned flow.

---

# Minimum Completion Dataset

Минимально завершенный Individual Lesson содержит:

```text
LessonId
StudentId
ActualTeacherId
ActualStartTime or ManualOccurrenceConfirmation
ActualEndTime
Attendance
LessonResult
CompletedBy
CompletedAt
PolicyId
PolicyVersion
```

Для Group Lesson дополнительно:

```text
ParticipantAttendance[]
```

Дополнительные объекты:

```text
Homework?
Assessment?
Attachments?
Recommendations?
SkillReferences?
SongReferences?
```

---

# Lesson Result Structure

Рекомендуемая концептуальная структура:

```text
LessonResult
├── Summary
├── Activities
├── PositiveObservations
├── DevelopmentAreas
├── Recommendations
├── SkillReferences
├── SongReferences
├── GoalReferences
└── Visibility
```

Не все элементы обязательны для каждого Lesson.

Однако Summary или эквивалентный содержательный результат обязателен.

---

# Completion Flow

```text
Teacher requests completion
          |
          v
Load Lesson and related state
          |
          v
Validate actor and version
          |
          v
Evaluate Lesson Completion Policy
          |
          +--> Rejected
          |
          +--> Missing Required Data
          |
          +--> Confirmation Required
          |
          +--> Administrative Review
          |
          v
Approved
          |
          v
Execute CompleteLesson
          |
          v
LessonCompleted
          |
          +--> Progress Update Policy
          |
          +--> Notification Policy
          |
          +--> Achievement Award Policy
          |
          +--> Periodic Review Policy
```

---

# Commands Produced

В зависимости от результата политика может сформировать следующие команды.

## CompleteLesson

Формируется при Approved.

## RequestLessonResult

Формируется при отсутствии образовательного результата.

## RequestAttendanceRecord

Формируется при незаполненном Attendance.

## RequestTeacherConfirmation

Формируется при допустимом, но нетипичном сценарии.

Примеры:

- прямое завершение из Scheduled;
- слишком короткое занятие;
- отсутствие Homework в ожидаемом контексте;
- использование AI-generated результата.

## CreateLessonFollowUp

Может формироваться после завершения, если существуют действия, которые не блокируют Completion.

Например:

- завершить расширенный Assessment позднее;
- прикрепить дополнительную запись;
- проверить Homework;
- уточнить Song readiness.

## RecalculateStudentProgress

Не должна формироваться напрямую до успешного LessonCompleted.

Обычно запускается другой политикой после публикации соответствующего события.

---

# Events Produced After Approval

Успешное выполнение команды должно создать:

## LessonCompleted

Событие должно содержать:

- LessonId;
- StudentIds;
- ActualTeacherId;
- CompletedBy;
- CompletedAt;
- actual duration;
- Attendance summary;
- Lesson Result reference;
- Homework references;
- Assessment references;
- PolicyId;
- PolicyVersion;
- CorrelationId.

Событие не обязано содержать полный текст чувствительных заметок.

---

# Human Confirmation

Human confirmation требуется, когда правило допускает исключение, но не может безопасно принять решение автоматически.

Подтверждение должно включать:

- ActorId;
- подтверждаемое условие;
- причину;
- дату и время;
- версию Lesson;
- версию политики.

Пример:

> Condition: Lesson duration is below expected minimum.
>
> Confirmation: The lesson ended early because the student felt unwell, but meaningful educational work was completed.

Подтверждение не должно использоваться для обхода абсолютных ограничений безопасности или прав доступа.

---

# Administrative Correction

После завершения могут обнаружиться ошибки.

Примеры:

- неверный Student;
- неверный Teacher;
- ошибочный Attendance;
- неправильное время;
- Lesson завершен вместо Missed;
- результат относится к другому занятию.

Прямое редактирование исторических данных запрещено.

Используется отдельный flow:

```text
RequestLessonCorrection
    |
    v
AdministrativeReview
    |
    v
ApproveLessonCorrection
    |
    v
LessonCorrectionApplied
```

Должны сохраняться:

- старое значение;
- новое значение;
- причина;
- инициатор;
- подтверждающий Actor;
- дата;
- связь с исходным LessonCompleted.

---

# Notification Effects

Lesson Completion Policy не отправляет уведомления самостоятельно.

После LessonCompleted Notification Policy может решить:

- уведомить Student о результате;
- уведомить о новом Homework;
- не отправлять уведомление, если данные еще не опубликованы ученику;
- объединить несколько уведомлений;
- отложить уведомление до разрешенного времени.

---

# Progress Effects

Сам факт LessonCompleted не обязан менять Progress.

Progress Update Policy должна отдельно оценить:

- наличие значимого Evidence;
- Assessment;
- Skill references;
- историю предыдущих Lesson;
- качество и устойчивость наблюдений.

Это предотвращает автоматическое создание ложного роста после каждого посещения.

---

# Failure Modes

## Lesson does not exist

- Decision: Rejected
- Reason Code: LESSON_NOT_FOUND

## Lesson already completed

- Decision: Already Completed

Новые последствия не создаются.

## Attendance is missing

- Decision: Missing Required Data
- Produced Command: RequestAttendanceRecord

## Lesson Result is empty

- Decision: Missing Required Data
- Produced Command: RequestLessonResult

## Lesson has no attendees

- Decision: Rejected

Рекомендуемый следующий сценарий:

- MarkLessonMissed;
- CancelLesson;
- CorrectAttendance.

## Lesson duration is suspiciously short

- Decision: Teacher Confirmation Required

## Actor has no permission

- Decision: Rejected

Security Audit обязателен при подозрительной попытке.

## Lesson version changed

- Decision: Rejected

Клиент должен загрузить актуальное состояние.

## Required service is unavailable

Если недоступен необязательный асинхронный сервис, Lesson может быть завершен, а последствие обработано позже.

Если недоступен источник обязательных данных или авторизации, операция откладывается либо отклоняется.

Инфраструктурная ошибка не должна маскироваться как бизнес-решение.

---

# Explainability

При невозможности завершения пользователь должен получить конкретное объяснение.

Недопустимо:

> Невозможно завершить урок.

Допустимо:

> Укажите посещаемость ученика перед завершением занятия.

или:

> Добавьте краткий результат занятия: над чем вы работали и что делать дальше.

или:

> Занятие длилось значительно меньше запланированного. Подтвердите причину досрочного завершения.

Для технических клиентов одновременно возвращается Reason Code.

---

# Audit Requirements

Для каждой оценки политики сохраняются:

- PolicyId;
- PolicyVersion;
- LessonId;
- LessonVersion;
- ActorId;
- ActorRole;
- evaluation time;
- decision;
- Reason Codes;
- missing fields;
- confirmation requirements;
- Evidence references;
- produced commands;
- CorrelationId;
- CausationId.

Для успешного Completion дополнительно:

- snapshot обязательных данных;
- confirmation records;
- AI assistance metadata;
- resulting Lesson version.

---

# Security Requirements

Политика должна проверять:

- принадлежность Teacher к школе;
- доступ к конкретному Student;
- доступ к филиалу или программе;
- полномочия replacement Teacher;
- административные ограничения;
- отсутствие заблокированного аккаунта;
- допустимость работы с чувствительными данными.

Backend является единственным доверенным местом применения политики.

Мобильный клиент может выполнять предварительную проверку только для UX.

---

# Privacy Requirements

В событие LessonCompleted не должен без необходимости включаться:

- полный текст Teacher Only notes;
- медицинская информация;
- закрытые Assessment;
- персональные вложения;
- приватные комментарии Student.

Потребители события должны получать только минимально необходимую информацию.

---

# AI Assistance

AI может помочь Teacher перед завершением:

- собрать Summary из черновых заметок;
- выделить Activities;
- предложить Development Areas;
- предложить Homework;
- проверить наличие обязательных элементов;
- найти потенциально противоречивые записи.

AI output должен:

- оставаться Draft;
- быть доступен Teacher для проверки;
- иметь происхождение;
- не добавлять факты, отсутствующие во входных данных;
- не публиковаться автоматически.

Teacher несет ответственность за подтвержденный Lesson Result.

---

# Read Models

## Teacher Completion Form

Показывает:

- участников;
- Attendance;
- фактическое время;
- предыдущий Homework;
- текущие Goals;
- Skills;
- Songs;
- поля Lesson Result;
- Draft Assessment;
- Draft Homework;
- блокирующие ошибки;
- предупреждения;
- требования подтверждения.

## Student Lesson History

После публикации показывает только разрешенные данные:

- дату;
- Teacher;
- тему;
- краткий результат;
- рекомендации;
- Homework;
- доступные вложения.

## Administrator Lesson Status View

Показывает:

- статус;
- время;
- Teacher;
- участников;
- Attendance completeness;
- наличие блокирующих организационных проблем.

Не должен по умолчанию раскрывать закрытые педагогические заметки.

---

# Examples

## Example 1: Standard completion

Дано:

- Lesson имеет статус Started;
- Teacher авторизован;
- Student присутствовал;
- указаны фактические времена;
- заполнен Lesson Result;
- Homework не требуется.

Результат:

- Decision: Approved
- Command: CompleteLesson
- Reason Code: COMPLETION_REQUIREMENTS_SATISFIED

## Example 2: Missing attendance

Дано:

- Lesson состоялся;
- Result заполнен;
- Attendance не указан.

Результат:

- Decision: Missing Required Data
- Command: RequestAttendanceRecord
- Reason Code: ATTENDANCE_REQUIRED

## Example 3: No homework

Дано:

- обычный Lesson завершен;
- Teacher указал, что новая самостоятельная работа не требуется.

Результат:

- Decision: Approved
- Reason Code: HOMEWORK_NOT_REQUIRED

## Example 4: Very short lesson

Дано:

- Lesson был запланирован на 60 минут;
- завершен через 12 минут;
- Attendance указан;
- результат заполнен.

Результат:

- Decision: Teacher Confirmation Required
- Reason Code: LESSON_DURATION_CONFIRMATION_REQUIRED

После подтверждения допустим Approved.

## Example 5: No student attended

Дано:

- Group Lesson;
- все Students отмечены как Absent.

Результат:

- Decision: Rejected
- Reason Code: NO_ATTENDING_STUDENTS

Система предлагает отметить занятие как Missed или применить другой утвержденный сценарий.

## Example 6: AI-generated result

Дано:

- AI создал Draft Lesson Result;
- Teacher не подтвердил текст.

Результат:

- Decision: Teacher Confirmation Required
- Reason Code: TEACHER_CONFIRMATION_REQUIRED_FOR_AI_CONTENT

## Example 7: Repeated command

Дано:

- Lesson уже Completed;
- повторно получена та же команда.

Результат:

- Decision: Already Completed
- Reason Code: LESSON_ALREADY_COMPLETED

Никакие события и уведомления повторно не создаются.

---

# Test Requirements

Минимальный набор тестов должен включать:

## State Tests

- Started Lesson can be completed;
- Cancelled Lesson cannot be completed;
- Missed Lesson cannot be completed;
- Completed Lesson is idempotent;
- Archived Lesson cannot be changed;
- Scheduled Lesson requires manual confirmation.

## Required Data Tests

- missing Teacher;
- missing Student;
- missing Attendance;
- missing Result;
- invalid completion time;
- end time before start time;
- future end time;
- incomplete group attendance.

## Permission Tests

- assigned Teacher completes Lesson;
- replacement Teacher with permission completes Lesson;
- unrelated Teacher is rejected;
- Student is rejected;
- Administrator uses corrective flow;
- blocked Actor is rejected.

## Educational Rules Tests

- Homework is optional;
- required Assessment is enforced for diagnostic type;
- empty templated result is rejected;
- meaningful summary is accepted;
- required attachment is enforced only in applicable scenarios.

## Confirmation Tests

- suspicious duration requires confirmation;
- AI-generated result requires confirmation;
- direct Scheduled-to-Completed transition requires confirmation;
- confirmation is stored in Audit.

## Concurrency Tests

- stale Lesson version is rejected;
- simultaneous completion produces one result;
- duplicate event does not duplicate effects.

## Privacy Tests

- private notes are excluded from public event;
- Administrator view does not expose Teacher Only Assessment;
- Student receives only visible result.

## Policy Version Tests

- decision records policy version;
- repeated evaluation with same inputs is deterministic;
- new policy version does not rewrite old completion record.

---

# Non-Goals

Lesson Completion Policy не определяет:

- алгоритм Progress;
- качество работы Teacher;
- размер оплаты;
- правила списания занятий;
- финансовую компенсацию;
- CRM-статусы;
- расписание следующего Lesson;
- готовность к Concert;
- выдачу Achievement;
- дисциплинарные меры.

Эти вопросы относятся к другим политикам или продуктам.

---

# Open Questions

Необходимо определить:

- какие Lesson Types существуют в первой версии;
- допускается ли завершение из Scheduled без Start;
- минимально достаточную структуру Lesson Result;
- нужна ли минимальная продолжительность для разных типов Lesson;
- кто может исправлять завершенный Lesson;
- какие изменения требуют двойного подтверждения;
- обязателен ли Homework в конкретных программах;
- когда Formal Assessment обязателен;
- как завершать занятия при техническом сбое приложения;
- как учитывать частичное участие в Group Lesson;
- как обрабатывать замену Teacher;
- должен ли Student подтверждать факт дистанционного занятия;
- какие данные сразу видимы Student;
- нужен ли период отложенной публикации результата;
- как учитывать Lesson, проведенный вне расписания;
- как завершать совместное занятие нескольких Teachers.

---

# History

| Version | Description |
| --- | --- |
| 1.0.0 | Определены условия, решения и последствия завершения Lesson. |
