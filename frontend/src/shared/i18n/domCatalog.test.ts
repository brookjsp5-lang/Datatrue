import { describe, expect, it } from 'vitest';

import { isExternalCatalogValueChange, translateCatalogValue } from './domCatalog';

describe('domCatalog', () => {
  it('keeps catalog-owned translations tied to the cached original text', () => {
    expect(isExternalCatalogValueChange('选择工作区', 'Select a workspace', '选择工作区')).toBe(
      false,
    );
  });

  it('detects React updates that reuse a translated text node', () => {
    expect(isExternalCatalogValueChange('test', 'Select a workspace', '选择工作区')).toBe(true);
  });

  it('restores cached original text when the catalog is empty', () => {
    expect(translateCatalogValue('Select a workspace', {})).toBe('Select a workspace');
  });

  it('translates through the active catalog', () => {
    expect(
      translateCatalogValue('Select a workspace', {
        'Select a workspace': '选择工作区',
      }),
    ).toBe('选择工作区');
  });
});
