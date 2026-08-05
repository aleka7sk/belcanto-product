import type { NavLabelKey } from "../navigation/tabs";
import {
  kkKZAccount,
  ruKZAccount,
  type AccountMessageKey,
} from "./accountMessages";

/**
 * Locale contract per HOF-15: ru-KZ and kk-KZ, ICU-style plural rules,
 * no runtime string concatenation. Catalogs are typed end to end: adding a
 * key to one locale without the other fails typecheck.
 *
 * Kazakh provenance: the design file materializes kk-KZ only on Page 34
 * (screen 370:107 «A11Y-03 · kk-KZ expansion» and board 372:200
 * «LOC-B · Localization contract»). Strings present there are carried
 * verbatim and marked `figma 370:107` / `figma 372:200`; the remaining
 * kk-KZ strings are implementation translations of the canonical Russian
 * copy and are flagged for human Kazakh review in the release report.
 */
export type Locale = "ru-KZ" | "kk-KZ";

export const SUPPORTED_LOCALES: readonly Locale[] = ["ru-KZ", "kk-KZ"];

export const DEFAULT_LOCALE: Locale = "ru-KZ";

export type MessageKey = NavLabelKey | AccountMessageKey;

export type Catalog = Readonly<Record<MessageKey, string>>;

/** Bottom Navigation labels, verbatim from Figma 310:20542 variants. */
const ruKZNav: Readonly<Record<NavLabelKey, string>> = {
  "nav.today": "Сегодня",
  "nav.schedule": "Расписание",
  "nav.practice": "Практика",
  "nav.community": "Сообщество",
  "nav.profile": "Профиль",
  "nav.students": "Ученики",
  "nav.review": "Проверка",
  "nav.operations": "Операции",
  "nav.people": "Люди",
  "nav.more": "Ещё",
  "nav.overview": "Обзор",
  "nav.analytics": "Аналитика",
  "nav.team": "Команда",
};

const kkKZNav: Readonly<Record<NavLabelKey, string>> = {
  "nav.today": "Бүгін",
  "nav.schedule": "Кесте", // figma 370:107 (shell override I370:139;310:289)
  "nav.practice": "Тәжірибе",
  "nav.community": "Қауымдастық", // figma 370:107 (I370:139;310:300)
  "nav.profile": "Профиль",
  "nav.students": "Оқушылар",
  "nav.review": "Тексеру",
  "nav.operations": "Операциялар", // figma 370:107 (I370:139;310:284)
  "nav.people": "Адамдар", // figma 370:107 (I370:139;310:295)
  "nav.more": "Тағы", // figma 370:107 (I370:139;310:307)
  "nav.overview": "Шолу",
  "nav.analytics": "Аналитика",
  "nav.team": "Команда",
};

export const ruKZ: Catalog = { ...ruKZNav, ...ruKZAccount };

export const kkKZ: Catalog = { ...kkKZNav, ...kkKZAccount };
