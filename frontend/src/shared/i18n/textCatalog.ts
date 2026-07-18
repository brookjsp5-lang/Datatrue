export type TextCatalog = Record<string, string>;

const getWhitespaceBounds = (text: string) => {
  const leadingWhitespace = text.match(/^\s*/)?.[0] ?? '';
  const trailingWhitespace = text.match(/\s*$/)?.[0] ?? '';

  return { leadingWhitespace, trailingWhitespace };
};

const COUNT_SUMMARY_PATTERN = /^(\d+) of (\d+) (logs|users)$/;

const normalizeInlineWhitespace = (text: string) => text.trim().replace(/\s+/g, ' ');

const decodeCatalogEntities = (text: string) =>
  text
    .replace(/&apos;/g, "'")
    .replace(/&quot;/g, '"')
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>');

const PREFIX_DELIMITERS = new Set([' ', ':', '：', '(', '[', '"', "'", '“', '&']);

const hasTranslatablePrefix = (text: string, sourceText: string) => {
  const nextCharacter = text[sourceText.length];

  return (
    text.startsWith(sourceText) && Boolean(nextCharacter && PREFIX_DELIMITERS.has(nextCharacter))
  );
};

const getPrefixTranslation = (text: string, catalog: TextCatalog): string | undefined => {
  const matchingSource = Object.keys(catalog)
    .map((sourceText) => ({ sourceText, lookupText: decodeCatalogEntities(sourceText) }))
    .filter(({ lookupText }) => hasTranslatablePrefix(text, lookupText))
    .sort(
      (firstSource, secondSource) => secondSource.lookupText.length - firstSource.lookupText.length,
    )[0];

  if (!matchingSource) {
    return undefined;
  }

  return `${catalog[matchingSource.sourceText]}${text.slice(matchingSource.lookupText.length)}`;
};

const getExactTranslation = (text: string, catalog: TextCatalog): string | undefined => {
  const directTranslation = catalog[text];
  if (directTranslation) {
    return directTranslation;
  }

  const matchingSource = Object.keys(catalog).find(
    (sourceText) => decodeCatalogEntities(sourceText) === text,
  );

  return matchingSource ? catalog[matchingSource] : undefined;
};

const getNormalizedTranslation = (text: string, catalog: TextCatalog): string | undefined => {
  const normalizedText = normalizeInlineWhitespace(text);
  if (!normalizedText) {
    return undefined;
  }

  const matchingSource = Object.keys(catalog).find(
    (sourceText) => normalizeInlineWhitespace(decodeCatalogEntities(sourceText)) === normalizedText,
  );

  return matchingSource ? catalog[matchingSource] : undefined;
};

const getCountSummaryTranslation = (text: string, catalog: TextCatalog): string | undefined => {
  const match = text.match(COUNT_SUMMARY_PATTERN);
  if (!match) {
    return undefined;
  }

  const [, loadedCount, totalCount, noun] = match;
  const translatedNoun = catalog[noun];
  if (!translatedNoun || translatedNoun === noun) {
    return undefined;
  }

  return `已加载 ${loadedCount} / ${totalCount} ${translatedNoun}`;
};

const getCatalogTranslation = (
  text: string,
  catalog: TextCatalog,
  options: { allowNormalized?: boolean } = {},
): string | undefined => {
  const allowNormalized = options.allowNormalized ?? true;

  return (
    getExactTranslation(text, catalog) ??
    (allowNormalized ? getNormalizedTranslation(text, catalog) : undefined) ??
    getPrefixTranslation(text, catalog) ??
    getCountSummaryTranslation(text, catalog)
  );
};

export const translateText = (text: string, catalog: TextCatalog): string => {
  const translatedText = getCatalogTranslation(text, catalog, {
    allowNormalized: text === text.trim(),
  });
  if (translatedText) {
    return translatedText;
  }

  const trimmedText = text.trim();
  const translatedTrimmedText = getCatalogTranslation(trimmedText, catalog);
  if (!translatedTrimmedText) {
    return text;
  }

  const { leadingWhitespace, trailingWhitespace } = getWhitespaceBounds(text);

  return `${leadingWhitespace}${translatedTrimmedText}${trailingWhitespace}`;
};
