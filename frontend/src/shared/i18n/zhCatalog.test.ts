import { describe, expect, it } from 'vitest';

import indexHtmlSource from '../../../index.html?raw';
import zhCN from '../../../public/locales/zh-CN.json';
import physicalConnectionErrorContentSource from '../../entity/databases/model/postgresql/physical/physicalConnectionErrorContent.ts?raw';
import editLogicalBackupConfigSource from '../../features/backups/logical/ui/EditLogicalBackupConfigComponent.tsx?raw';
import showLogicalBackupConfigSource from '../../features/backups/logical/ui/ShowLogicalBackupConfigComponent.tsx?raw';
import editPhysicalBackupConfigSource from '../../features/backups/physical/ui/EditPhysicalBackupConfigComponent.tsx?raw';
import createReadOnlySource from '../../features/databases/ui/edit/CreateReadOnlyComponent.tsx?raw';
import editMariaDbSource from '../../features/databases/ui/edit/EditMariaDbSpecificDataComponent.tsx?raw';
import editMongoDbSource from '../../features/databases/ui/edit/EditMongoDbSpecificDataComponent.tsx?raw';
import editMySqlSource from '../../features/databases/ui/edit/EditMySqlSpecificDataComponent.tsx?raw';
import editPostgreSqlLogicalSource from '../../features/databases/ui/edit/EditPostgreSqlLogicalSpecificDataComponent.tsx?raw';
import settingsSource from '../../features/settings/ui/SettingsComponent.tsx?raw';
import editNasStorageSource from '../../features/storages/ui/edit/storages/EditNASStorageComponent.tsx?raw';
import authNavbarSource from '../../features/users/ui/AuthNavbarComponent.tsx?raw';
import usersSource from '../../features/users/ui/UsersComponent.tsx?raw';
import verificationsSource from '../../features/verification/runs/ui/VerificationsComponent.tsx?raw';
import mainScreenSource from '../../widgets/main/MainScreenComponent.tsx?raw';

const requiredChineseCatalogKeys = [
  '+ Create new workspace',
  'Access key',
  'Account key',
  'Add database',
  'Add database for backup',
  'Add member',
  'Add storage',
  'After backup',
  'After installation',
  'Agent ID:',
  'Agent name',
  'Agent token',
  'Allow external registrations',
  'Allow member invitations',
  'Already have an account?',
  'Active:',
  'At least 8 characters',
  'Audit Logs',
  'Azure Blob Storage',
  'Back up now',
  'Backup interval',
  'Backup now',
  'Backups enabled',
  'Change password',
  'Changing password...',
  'Check status:',
  'Confirm New Password',
  'Confirm Password',
  'Container name',
  'Create user',
  'Create a user account with an initial password. The user can sign in immediately.',
  'Don&apos;t have an account?',
  'Email address',
  'Enter email address',
  'Enter a valid email address',
  'Forgot password?',
  'Failed to create user',
  'Full + incremental',
  'Full + incremental + WAL',
  'Full backup',
  'Full backups',
  'Full backups only',
  'Google Drive',
  'Initial password',
  'Language: English',
  'Language: Simplified Chinese',
  'Logout',
  'New Password',
  'Need an account? Ask an administrator to create one.',
  'No agents registered yet.',
  'No databases found matching &quot;',
  'No restore checks yet',
  'No restores yet',
  'No workspaces found',
  'Notifier settings',
  'OK',
  'Password is required',
  'Point-in-time restore',
  'Power Automate URL',
  'Private Key',
  'Public user registration is disabled',
  'Profile Information',
  'Queue check',
  'Remote path',
  'Reset Password',
  'Reset password',
  'Role:',
  'Role is required',
  'Restore from backup',
  'Restore from this backup',
  'Restore verifications',
  'Retention policy',
  'Retry backup if failed',
  'S3 Bucket',
  'Save changes',
  'Scheduled verification',
  'Search database',
  'Search notifier',
  'Search workspaces...',
  'Secret key',
  'Select a workspace',
  'Selected workspace',
  'Send reset code',
  'Send test notification',
  'My workspace',
  'Workspace created successfully',
  '\u2190 Back to databases',
] as const;

const requiredStorageCatalogKeys = [
  '- test connection via button below (even if you updated settings);',
  '- wait until the next save is done without errors;',
  '*content is hidden to not expose sensitive data. If you want to update existing config, put a new one here',
  '/backups (optional)',
  'backups (optional)',
  'backups (optional, no leading slash)',
  'Be careful: with local storage you may run out of ROM memory. It is recommended to use S3 or unlimited storages',
  'my-prefix/ (optional)',
  "Optional path prefix on the remote where backups will be stored (e.g., '/backups' or 'my-folder/backups')",
  "Optional prefix for all blob names (e.g., 'backups/' or 'my_team/')",
  'Paste your SSH private key (PEM format). Supports RSA, DSA, ECDSA, and Ed25519 keys.',
  'Share must be a single share name. Use the Path field for subdirectories (e.g. Share: Datatrue, Path: DB1)',
  'Skip SSH host key verification. Enable this if you trust the server. Warning: this reduces security.',
  'Local storage',
  'Storage - is a place where backups will be stored (local disk, S3, etc.)',
  'Storage - is a place where backups will be stored (local disk, S3, Google Drive, etc.)',
  'Type',
  'Type:',
  'local storage',
] as const;

const requiredSettingsAndBrandCatalogKeys = [
  'Datatrue settings',
  'Datatrue users',
  'Health-check',
  'Read more about settings you can',
  'here',
  'When enabled, new users can register accounts in Datatrue. If disabled, new users can only register via invitation',
  'When enabled, existing members can invite new users to join Datatrue. If not - only admins can invite users.',
  'When enabled, members (non-admin users) can create new workspaces. If not - only admins can create workspaces.',
  'Use this endpoint to monitor your Datatrue system&apos;s availability',
  'Administrators create users from the Datatrue users page. New users cannot create their own accounts from the sign-in screen.',
  'No verification agents registered - please add it in Datatrue settings tab in',
  'Admin',
  'Member',
  'User',
  'User created successfully',
  'Message',
  'Workspace',
  'Created',
  'logs',
  'users',
  'System',
  'User signed in with email',
  'Workspace created',
] as const;

const requiredWorkspaceCatalogKeys = [
  'Workspace is a place where you group:',
  '- your databases;',
  '- storages (like local drive, S3, Google Drive, etc.)',
  '- notifiers (like email, Slack, Telegram, etc.);',
  '- access control (if you have team);',
  'Workspace name',
  'Failed to create workspace',
  'You don&apos;t have permission to create workspaces. Please ask the administrator to create the workspace for you.',
] as const;

const visibleBrandSources = [
  [
    'src/entity/databases/model/postgresql/physical/physicalConnectionErrorContent.ts',
    physicalConnectionErrorContentSource,
  ],
  [
    'src/features/backups/logical/ui/EditLogicalBackupConfigComponent.tsx',
    editLogicalBackupConfigSource,
  ],
  [
    'src/features/backups/logical/ui/ShowLogicalBackupConfigComponent.tsx',
    showLogicalBackupConfigSource,
  ],
  [
    'src/features/backups/physical/ui/EditPhysicalBackupConfigComponent.tsx',
    editPhysicalBackupConfigSource,
  ],
  ['src/features/databases/ui/edit/CreateReadOnlyComponent.tsx', createReadOnlySource],
  ['src/features/databases/ui/edit/EditMariaDbSpecificDataComponent.tsx', editMariaDbSource],
  ['src/features/databases/ui/edit/EditMongoDbSpecificDataComponent.tsx', editMongoDbSource],
  ['src/features/databases/ui/edit/EditMySqlSpecificDataComponent.tsx', editMySqlSource],
  [
    'src/features/databases/ui/edit/EditPostgreSqlLogicalSpecificDataComponent.tsx',
    editPostgreSqlLogicalSource,
  ],
  ['src/features/settings/ui/SettingsComponent.tsx', settingsSource],
  ['src/features/storages/ui/edit/storages/EditNASStorageComponent.tsx', editNasStorageSource],
  ['src/features/users/ui/AuthNavbarComponent.tsx', authNavbarSource],
  ['src/features/users/ui/UsersComponent.tsx', usersSource],
  ['src/features/verification/runs/ui/VerificationsComponent.tsx', verificationsSource],
  ['src/widgets/main/MainScreenComponent.tsx', mainScreenSource],
  ['index.html', indexHtmlSource],
] as const;

describe('zh-CN catalog coverage', () => {
  it('covers high-visibility frontend text found by the i18n audit', () => {
    const missing = requiredChineseCatalogKeys.filter((key) => !zhCN[key]);

    expect(missing).toEqual([]);
  });

  it('does not leave high-visibility entries in English', () => {
    const untranslated = requiredChineseCatalogKeys.filter((key) => zhCN[key] === key);

    expect(untranslated).toEqual([]);
  });

  it('covers storage setup text found by the storage i18n audit', () => {
    const missing = requiredStorageCatalogKeys.filter((key) => !zhCN[key]);

    expect(missing).toEqual([]);
  });

  it('does not leave storage setup entries in English', () => {
    const untranslated = requiredStorageCatalogKeys.filter((key) => zhCN[key] === key);

    expect(untranslated).toEqual([]);
  });

  it('covers settings and Datatrue brand text found by the i18n audit', () => {
    const missing = requiredSettingsAndBrandCatalogKeys.filter((key) => !zhCN[key]);

    expect(missing).toEqual([]);
  });

  it('does not leave Databasus in the Simplified Chinese catalog', () => {
    const leaks = Object.entries(zhCN).filter(
      ([key, value]) => key.includes('Databasus') || value.includes('Databasus'),
    );

    expect(leaks).toEqual([]);
  });

  it('uses Datatrue in high-visibility frontend source strings', () => {
    const leaks = visibleBrandSources
      .filter(([, source]) => source.includes('Databasus'))
      .map(([relativePath]) => relativePath);

    expect(leaks).toEqual([]);
  });

  it('covers workspace dialog text found by the i18n audit', () => {
    const missing = requiredWorkspaceCatalogKeys.filter((key) => !zhCN[key]);

    expect(missing).toEqual([]);
  });

  it('does not leave workspace dialog entries in English', () => {
    const untranslated = requiredWorkspaceCatalogKeys.filter((key) => zhCN[key] === key);

    expect(untranslated).toEqual([]);
  });
});
