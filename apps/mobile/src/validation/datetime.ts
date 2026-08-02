const RFC3339_PATTERN =
  /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.\d{1,9})?(?:Z|[+-](\d{2}):(\d{2}))$/;

function leapYear(year: number): boolean {
  return year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0);
}

function daysInMonth(year: number, month: number): number {
  if (month === 2) return leapYear(year) ? 29 : 28;
  return [4, 6, 9, 11].includes(month) ? 30 : 31;
}

export function parseStrictRfc3339(value: string): number | null {
  const match = RFC3339_PATTERN.exec(value);
  if (match === null) return null;
  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
  const hour = Number(match[4]);
  const minute = Number(match[5]);
  const second = Number(match[6]);
  const offsetHour = match[7] === undefined ? 0 : Number(match[7]);
  const offsetMinute = match[8] === undefined ? 0 : Number(match[8]);
  if (
    month < 1 ||
    month > 12 ||
    day < 1 ||
    day > daysInMonth(year, month) ||
    hour > 23 ||
    minute > 59 ||
    second > 59 ||
    offsetHour > 23 ||
    offsetMinute > 59
  ) {
    return null;
  }
  const timestamp = Date.parse(value);
  return Number.isFinite(timestamp) ? timestamp : null;
}

const HUMAN_DATE_PATTERN = /^(\d{2})\.(\d{2})\.(\d{4})$/;
const HUMAN_TIME_PATTERN = /^(\d{2}):(\d{2})$/;
const BELCANTO_TIME_ZONE = "Asia/Almaty";

function zonedParts(timestamp: number): Record<string, number> {
  const parts = new Intl.DateTimeFormat("en-CA", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hourCycle: "h23",
    timeZone: BELCANTO_TIME_ZONE,
  }).formatToParts(new Date(timestamp));
  return Object.fromEntries(
    parts
      .filter((part) => part.type !== "literal")
      .map((part) => [part.type, Number(part.value)]),
  );
}

/** Converts human Belcanto date/time fields to an instant in Asia/Almaty. */
export function parseAlmatyLocalDateTime(
  dateValue: string,
  timeValue: string,
): number | null {
  const date = HUMAN_DATE_PATTERN.exec(dateValue.trim());
  const time = HUMAN_TIME_PATTERN.exec(timeValue.trim());
  if (date === null || time === null) return null;
  const day = Number(date[1]);
  const month = Number(date[2]);
  const year = Number(date[3]);
  const hour = Number(time[1]);
  const minute = Number(time[2]);
  if (
    month < 1 ||
    month > 12 ||
    day < 1 ||
    day > daysInMonth(year, month) ||
    hour > 23 ||
    minute > 59
  ) {
    return null;
  }
  const desiredAsUtc = Date.UTC(year, month - 1, day, hour, minute, 0);
  let timestamp = desiredAsUtc;
  for (let attempt = 0; attempt < 3; attempt += 1) {
    const parts = zonedParts(timestamp);
    const representedAsUtc = Date.UTC(
      parts.year!,
      parts.month! - 1,
      parts.day!,
      parts.hour!,
      parts.minute!,
      parts.second!,
    );
    timestamp += desiredAsUtc - representedAsUtc;
  }
  const verified = zonedParts(timestamp);
  return verified.year === year &&
    verified.month === month &&
    verified.day === day &&
    verified.hour === hour &&
    verified.minute === minute
    ? timestamp
    : null;
}
