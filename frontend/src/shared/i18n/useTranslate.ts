import { useCallback } from 'react';

import { translateText } from './textCatalog';
import { useLanguage } from './useLanguage';

export const useTranslate = () => {
  const { catalog } = useLanguage();

  return useCallback((text: string) => translateText(text, catalog), [catalog]);
};
