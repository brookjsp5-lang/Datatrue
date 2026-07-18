import { GlobalOutlined } from '@ant-design/icons';
import { Dropdown } from 'antd';
import type { MenuProps } from 'antd';

import { type LanguageCode, useLanguage } from '../i18n';

const getLanguageLabel = (language: LanguageCode) => {
  switch (language) {
    case 'en':
      return 'English';
    case 'zh-CN':
      return 'Simplified Chinese';
  }
};

export const LanguageToggleComponent = () => {
  const { language, setLanguage } = useLanguage();

  const items: MenuProps['items'] = [
    {
      key: 'en',
      label: 'English',
      onClick: () => setLanguage('en'),
    },
    {
      key: 'zh-CN',
      label: 'Simplified Chinese',
      onClick: () => setLanguage('zh-CN'),
    },
  ];

  return (
    <Dropdown
      menu={{ items, selectedKeys: [language] }}
      trigger={['click']}
      placement="bottomRight"
    >
      <button
        className="flex cursor-pointer items-center gap-1.5 rounded-md border border-gray-200 bg-white px-2.5 py-1 text-sm text-gray-700 transition-colors hover:bg-gray-50 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-200 dark:hover:bg-gray-600"
        title={`Language: ${getLanguageLabel(language)}`}
      >
        <GlobalOutlined />
        <span className="hidden sm:inline">{getLanguageLabel(language)}</span>
      </button>
    </Dropdown>
  );
};
