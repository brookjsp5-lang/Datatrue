export type LanguageCode = 'en' | 'zh-CN';

const SUPPORTED_LANGUAGES = new Set<LanguageCode>(['en', 'zh-CN']);

export const DEFAULT_LANGUAGE: LanguageCode = 'en';
export const LANGUAGE_STORAGE_KEY = 'databasus-language';

export const isLanguageCode = (value: unknown): value is LanguageCode => {
  return typeof value === 'string' && SUPPORTED_LANGUAGES.has(value as LanguageCode);
};

export const resolveInitialLanguage = (
  storedLanguage: string | null | undefined,
  browserLanguage: string | undefined,
): LanguageCode => {
  if (isLanguageCode(storedLanguage)) {
    return storedLanguage;
  }

  if (browserLanguage?.toLowerCase().startsWith('zh')) {
    return 'zh-CN';
  }

  return DEFAULT_LANGUAGE;
};
