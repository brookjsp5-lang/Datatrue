import { describe, expect, it } from 'vitest';

import { translateText } from './textCatalog';

describe('textCatalog', () => {
  it('translates exact catalog matches', () => {
    const translatedText = translateText('Settings', { Settings: 'Preferences' });

    expect(translatedText).toBe('Preferences');
  });

  it('preserves surrounding whitespace when translating text nodes', () => {
    const translatedText = translateText('  Settings\n', { Settings: 'Preferences' });

    expect(translatedText).toBe('  Preferences\n');
  });

  it('translates a known prefix without changing the dynamic suffix', () => {
    const translatedText = translateText('Last backup 2 hours ago', {
      'Last backup': 'Previous backup',
    });

    expect(translatedText).toBe('Previous backup 2 hours ago');
  });

  it('translates a known prefix followed by punctuation without changing the suffix', () => {
    const translatedText = translateText('Agent ID: backup-agent', {
      'Agent ID': 'Agent 编号',
    });

    expect(translatedText).toBe('Agent 编号: backup-agent');
  });

  it('matches catalog entries when JSX text contains internal newlines and indentation', () => {
    const translatedText = translateText(
      'When enabled, new users can register accounts in Datatrue. If disabled, new\n users can only register via invitation',
      {
        'When enabled, new users can register accounts in Datatrue. If disabled, new users can only register via invitation':
          '启用后，新用户可以在 Datatrue 注册账号。禁用后，新用户只能通过邀请注册',
      },
    );

    expect(translatedText).toBe(
      '启用后，新用户可以在 Datatrue 注册账号。禁用后，新用户只能通过邀请注册',
    );
  });

  it('prefers a normalized full sentence over a shorter translated prefix', () => {
    const translatedText = translateText('  Workspace is a place where you group:', {
      Workspace: '工作区',
      'Workspace is a place where you group:': '工作区用于归类管理：',
    });

    expect(translatedText).toBe('  工作区用于归类管理：');
  });

  it('formats common count summaries in Chinese when the noun is translated', () => {
    const translatedText = translateText('4 of 4 logs', {
      logs: '条日志',
    });

    expect(translatedText).toBe('已加载 4 / 4 条日志');
  });

  it('matches catalog keys that contain JSX-decoded HTML entities', () => {
    const translatedText = translateText(
      "Use this endpoint to monitor your Datatrue system's availability",
      {
        'Use this endpoint to monitor your Datatrue system&apos;s availability':
          '使用此端点监控 Datatrue 系统可用性',
      },
    );

    expect(translatedText).toBe('使用此端点监控 Datatrue 系统可用性');
  });

  it('returns the original text when no catalog entry matches', () => {
    const translatedText = translateText('Unmapped text', { Settings: 'Preferences' });

    expect(translatedText).toBe('Unmapped text');
  });
});
