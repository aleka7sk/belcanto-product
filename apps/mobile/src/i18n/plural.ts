import type { Locale } from "./messages";

/**
 * CLDR plural selection via Intl.PluralRules (HOF-15: ICU plural, no
 * runtime string concatenation). Russian distinguishes one/few/many;
 * Kazakh uses one/other. `other` is the mandatory fallback form.
 */
export type PluralForms = Readonly<
  Partial<Record<Intl.LDMLPluralRule, string>> & { other: string }
>;

const rulesCache = new Map<Locale, Intl.PluralRules>();

function rulesFor(locale: Locale): Intl.PluralRules {
  const cached = rulesCache.get(locale);
  if (cached) {
    return cached;
  }
  const rules = new Intl.PluralRules(locale);
  rulesCache.set(locale, rules);
  return rules;
}

export function selectPlural(
  locale: Locale,
  count: number,
  forms: PluralForms,
): string {
  const category = rulesFor(locale).select(count);
  return forms[category] ?? forms.other;
}

/**
 * Substitutes named `{param}` placeholders. Templates keep whole sentences
 * in the catalog; values are inserted, never concatenated around.
 */
export function formatTemplate(
  template: string,
  params: Readonly<Record<string, string | number>>,
): string {
  return template.replace(/\{([a-zA-Z0-9_]+)\}/g, (match, name: string) => {
    const value = params[name];
    return value === undefined ? match : String(value);
  });
}
