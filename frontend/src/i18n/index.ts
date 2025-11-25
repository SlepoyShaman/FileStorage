// i18n.js
import { createI18n } from 'vue-i18n';

import en from './en.json';
import ru from './ru.json';

type LocaleMap = { [key: string]: string };

export const internalToStandardLocaleMap: { [key: string]: string } = {
  nlBE: 'nl-be',
  ptBR: 'pt-br',
  svSE: 'sv-se',
  zhCN: 'zh-cn',
  zhTW: 'zh-tw',
  cz: 'cs',
  ua: 'uk',
};

export function toStandardLocale(locale: string): string {
  return internalToStandardLocaleMap[locale] || locale;
}

export function detectLocale(): string {
  const locale = navigator.language.toLowerCase();
  const localeMap: LocaleMap = {
    'en': 'en',
    'ru': 'ru',
  };

  for (const key in localeMap) {
    if (locale.startsWith(key)) {
      return localeMap[key];
    }
  }
  return 'en-us'; // Default fallback
}


export function setLocale(locale: string) {
  //@ts-ignore
  i18n.global.locale.value = locale;
}


const i18n = createI18n({
  locale: detectLocale(),
  fallbackLocale: 'en',
  legacy: true,
  messages: {
    en,
    ru,
  },
});

export default i18n;
