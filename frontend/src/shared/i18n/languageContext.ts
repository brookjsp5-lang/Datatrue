import type { Locale } from 'antd/es/locale';
import { createContext } from 'react';

import type { LanguageCode } from './language';
import type { TextCatalog } from './textCatalog';

export interface LanguageContextValue {
  language: LanguageCode;
  setLanguage: (language: LanguageCode) => void;
  catalog: TextCatalog;
  antdLocale: Locale;
}

export const LanguageContext = createContext<LanguageContextValue | undefined>(undefined);
