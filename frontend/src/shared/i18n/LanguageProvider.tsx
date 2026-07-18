import enUS from 'antd/locale/en_US';
import zhCN from 'antd/locale/zh_CN';
import dayjs from 'dayjs';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { ReactNode } from 'react';

import { isExternalCatalogValueChange, translateCatalogValue } from './domCatalog';
import { LANGUAGE_STORAGE_KEY, type LanguageCode, resolveInitialLanguage } from './language';
import { LanguageContext } from './languageContext';
import { type TextCatalog } from './textCatalog';

const TRANSLATED_ATTRIBUTES = ['aria-label', 'placeholder', 'title'];

interface LanguageProviderProps {
  children: ReactNode;
}

const getInitialLanguage = (): LanguageCode => {
  if (typeof window === 'undefined') {
    return 'en';
  }

  return resolveInitialLanguage(
    localStorage.getItem(LANGUAGE_STORAGE_KEY),
    navigator.language || navigator.languages?.[0],
  );
};

const loadLanguageCatalog = async (language: LanguageCode): Promise<TextCatalog> => {
  if (language === 'en') {
    return {};
  }

  const response = await fetch(`/locales/${language}.json`);
  if (!response.ok) {
    return {};
  }

  return (await response.json()) as TextCatalog;
};

const applyTextNodeCatalog = (
  textNode: Text,
  catalog: TextCatalog,
  originalTextByNode: WeakMap<Text, string>,
  lastAppliedTextByNode: WeakMap<Text, string>,
) => {
  const currentText = textNode.textContent ?? '';
  let originalText = originalTextByNode.get(textNode);
  const lastAppliedText = lastAppliedTextByNode.get(textNode);

  if (isExternalCatalogValueChange(currentText, originalText, lastAppliedText)) {
    originalText = currentText;
    originalTextByNode.set(textNode, originalText);
  }

  const translatedText = translateCatalogValue(originalText ?? currentText, catalog);
  if (textNode.textContent !== translatedText) {
    textNode.textContent = translatedText;
  }

  lastAppliedTextByNode.set(textNode, translatedText);
};

const getOriginalAttributeName = (attributeName: string) => {
  return `data-i18n-original-${attributeName}`;
};

const getLastAppliedAttributeName = (attributeName: string) => {
  return `data-i18n-applied-${attributeName}`;
};

const applyElementCatalog = (element: Element, catalog: TextCatalog) => {
  for (const attributeName of TRANSLATED_ATTRIBUTES) {
    const originalAttributeName = getOriginalAttributeName(attributeName);
    const lastAppliedAttributeName = getLastAppliedAttributeName(attributeName);
    const currentAttributeValue = element.getAttribute(attributeName);

    if (!currentAttributeValue) {
      continue;
    }

    let originalAttributeValue = element.getAttribute(originalAttributeName) ?? undefined;
    const lastAppliedAttributeValue = element.getAttribute(lastAppliedAttributeName) ?? undefined;

    if (
      isExternalCatalogValueChange(
        currentAttributeValue,
        originalAttributeValue,
        lastAppliedAttributeValue,
      )
    ) {
      originalAttributeValue = currentAttributeValue;
      element.setAttribute(originalAttributeName, originalAttributeValue);
    }

    const translatedAttributeValue = translateCatalogValue(
      originalAttributeValue ?? currentAttributeValue,
      catalog,
    );

    if (currentAttributeValue !== translatedAttributeValue) {
      element.setAttribute(attributeName, translatedAttributeValue);
    }

    if (element.getAttribute(lastAppliedAttributeName) !== translatedAttributeValue) {
      element.setAttribute(lastAppliedAttributeName, translatedAttributeValue);
    }
  }
};

const applyCatalogToRoot = (
  root: ParentNode,
  catalog: TextCatalog,
  originalTextByNode: WeakMap<Text, string>,
  lastAppliedTextByNode: WeakMap<Text, string>,
) => {
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT | NodeFilter.SHOW_ELEMENT);

  let currentNode = walker.nextNode();
  while (currentNode) {
    if (currentNode.nodeType === Node.TEXT_NODE) {
      const textNode = currentNode as Text;
      if (textNode.textContent?.trim()) {
        applyTextNodeCatalog(textNode, catalog, originalTextByNode, lastAppliedTextByNode);
      }
    }

    if (currentNode.nodeType === Node.ELEMENT_NODE) {
      applyElementCatalog(currentNode as Element, catalog);
    }

    currentNode = walker.nextNode();
  }
};

export const LanguageProvider = ({ children }: LanguageProviderProps) => {
  const [language, setLanguageState] = useState<LanguageCode>(getInitialLanguage);
  const [catalog, setCatalog] = useState<TextCatalog>({});
  const originalTextByNode = useRef(new WeakMap<Text, string>());
  const lastAppliedTextByNode = useRef(new WeakMap<Text, string>());

  const setLanguage = useCallback((newLanguage: LanguageCode) => {
    setLanguageState(newLanguage);
    localStorage.setItem(LANGUAGE_STORAGE_KEY, newLanguage);
  }, []);

  useEffect(() => {
    let isCancelled = false;

    loadLanguageCatalog(language).then((loadedCatalog) => {
      if (!isCancelled) {
        setCatalog(loadedCatalog);
      }
    });

    document.documentElement.lang = language;
    dayjs.locale(language === 'zh-CN' ? 'zh-cn' : 'en');

    return () => {
      isCancelled = true;
    };
  }, [language]);

  useEffect(() => {
    if (typeof MutationObserver === 'undefined') {
      return;
    }

    const applyCurrentCatalog = () => {
      if (document.body) {
        applyCatalogToRoot(
          document.body,
          catalog,
          originalTextByNode.current,
          lastAppliedTextByNode.current,
        );
      }
    };

    applyCurrentCatalog();

    const observer = new MutationObserver(applyCurrentCatalog);
    observer.observe(document.body, {
      attributes: true,
      childList: true,
      characterData: true,
      subtree: true,
    });

    return () => observer.disconnect();
  }, [catalog]);

  const antdLocale = useMemo(() => (language === 'zh-CN' ? zhCN : enUS), [language]);

  const value = useMemo(
    () => ({
      language,
      setLanguage,
      catalog,
      antdLocale,
    }),
    [language, setLanguage, catalog, antdLocale],
  );

  return <LanguageContext.Provider value={value}>{children}</LanguageContext.Provider>;
};
