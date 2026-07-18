import { describe, expect, it } from 'vitest';

import { isLanguageCode, resolveInitialLanguage } from './language';

describe('language', () => {
  it('accepts only supported language codes', () => {
    expect(isLanguageCode('en')).toBe(true);
    expect(isLanguageCode('zh-CN')).toBe(true);
    expect(isLanguageCode('zh')).toBe(false);
    expect(isLanguageCode('fr')).toBe(false);
    expect(isLanguageCode(null)).toBe(false);
  });

  it('prefers a stored supported language over the browser language', () => {
    const language = resolveInitialLanguage('zh-CN', 'en-US');

    expect(language).toBe('zh-CN');
  });

  it('falls back to simplified Chinese for Chinese browser locales', () => {
    const language = resolveInitialLanguage('fr', 'zh-Hans-CN');

    expect(language).toBe('zh-CN');
  });

  it('falls back to English when no supported stored or browser language is available', () => {
    const language = resolveInitialLanguage(undefined, 'de-DE');

    expect(language).toBe('en');
  });
});
