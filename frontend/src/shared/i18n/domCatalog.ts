import { type TextCatalog, translateText } from './textCatalog';

export const translateCatalogValue = (value: string, catalog: TextCatalog): string =>
  Object.keys(catalog).length === 0 ? value : translateText(value, catalog);

export const isExternalCatalogValueChange = (
  currentValue: string,
  cachedOriginalValue: string | undefined,
  lastAppliedValue: string | undefined,
) => {
  if (cachedOriginalValue === undefined) {
    return true;
  }

  return currentValue !== cachedOriginalValue && currentValue !== lastAppliedValue;
};
