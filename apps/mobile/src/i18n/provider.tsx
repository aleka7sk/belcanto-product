import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";

import { DEFAULT_LOCALE, kkKZ, ruKZ, type Catalog, type Locale, type MessageKey } from "./messages";
import { formatTemplate } from "./plural";

type LocaleContextValue = {
  locale: Locale;
  setLocale: (locale: Locale) => void;
};

const LocaleContext = createContext<LocaleContextValue | null>(null);

const catalogs: Readonly<Record<Locale, Catalog>> = {
  "ru-KZ": ruKZ,
  "kk-KZ": kkKZ,
};

export function catalogFor(locale: Locale): Catalog {
  return catalogs[locale];
}

export function LocaleProvider({
  children,
  initialLocale = DEFAULT_LOCALE,
}: {
  children: ReactNode;
  initialLocale?: Locale;
}) {
  const [locale, setLocale] = useState<Locale>(initialLocale);
  const value = useMemo(() => ({ locale, setLocale }), [locale]);
  return <LocaleContext.Provider value={value}>{children}</LocaleContext.Provider>;
}

export function useLocale(): LocaleContextValue {
  const value = useContext(LocaleContext);
  if (value === null) {
    throw new Error("useLocale must be used within LocaleProvider");
  }
  return value;
}

export type MessageFormatter = (
  key: MessageKey,
  params?: Readonly<Record<string, string | number>>,
) => string;

export function useMessage(): MessageFormatter {
  const { locale } = useLocale();
  const catalog = catalogFor(locale);
  return useCallback<MessageFormatter>(
    (key, params) => {
      const template = catalog[key];
      return params === undefined ? template : formatTemplate(template, params);
    },
    [catalog],
  );
}
