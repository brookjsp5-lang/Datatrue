import { LoadingOutlined } from '@ant-design/icons';
import { App, Button, Input, Modal } from 'antd';
import { Spin } from 'antd';
import { useEffect, useState } from 'react';

import { type UserProfile, UserRole, type UsersSettings } from '../../../entity/users';
import type { WorkspaceResponse } from '../../../entity/workspaces';
import { workspaceApi } from '../../../entity/workspaces';
import { useTranslate } from '../../../shared/i18n';

interface Props {
  user: UserProfile;
  globalSettings: UsersSettings;

  onClose: () => void;
  onWorkspaceCreated: (workspace: WorkspaceResponse) => void;

  workspacesCount: number;
}

export const CreateWorkspaceDialogComponent = ({
  user,
  globalSettings,
  onClose,
  onWorkspaceCreated,
  workspacesCount,
}: Props) => {
  const { message } = App.useApp();
  const t = useTranslate();
  const [isCreating, setIsCreating] = useState(false);
  const [isWorkspaceNameEdited, setIsWorkspaceNameEdited] = useState(false);
  const [workspaceName, setWorkspaceName] = useState(
    workspacesCount === 0 ? t('My workspace') : '',
  );

  const isAllowedToCreateWorkspaces =
    globalSettings.isMemberAllowedToCreateWorkspaces || user.role === UserRole.ADMIN;

  useEffect(() => {
    if (workspacesCount === 0 && !isWorkspaceNameEdited) {
      setWorkspaceName(t('My workspace'));
    }
  }, [isWorkspaceNameEdited, t, workspacesCount]);

  const handleCreateWorkspace = async () => {
    if (!workspaceName.trim()) {
      message.error(t('Please enter a workspace name'));
      return;
    }

    setIsCreating(true);

    try {
      const newWorkspace = await workspaceApi.createWorkspace({
        name: workspaceName.trim(),
      });

      message.success(t('Workspace created successfully'));
      onWorkspaceCreated(newWorkspace);
      onClose();
    } catch (error) {
      message.error((error as Error).message || t('Failed to create workspace'));
    } finally {
      setIsCreating(false);
    }
  };

  if (!isAllowedToCreateWorkspaces) {
    return (
      <Modal
        title={t('Permission denied')}
        open
        onCancel={onClose}
        footer={[
          <Button key="ok" type="primary" onClick={onClose}>
            {t('OK')}
          </Button>,
        ]}
      >
        <p>
          {t(
            "You don't have permission to create workspaces. Please ask the administrator to create the workspace for you.",
          )}
        </p>
      </Modal>
    );
  }

  return (
    <Modal
      title={t('Create workspace')}
      open
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose} disabled={isCreating}>
          {t('Cancel')}
        </Button>,

        <Button
          key="create"
          type="primary"
          onClick={handleCreateWorkspace}
          disabled={isCreating || !workspaceName.trim()}
          className="border-blue-600 bg-blue-600 hover:border-blue-700 hover:bg-blue-700"
        >
          {isCreating ? (
            <Spin indicator={<LoadingOutlined spin />} size="small" />
          ) : (
            t('Create workspace')
          )}
        </Button>,
      ]}
    >
      <div className="mb-4">
        <div className="dark:text-gray-300">
          {t('Workspace is a place where you group:')}
          <br />
          {t('- your databases;')}
          <br />
          {t('- storages (like local drive, S3, Google Drive, etc.)')}
          <br />
          {t('- notifiers (like email, Slack, Telegram, etc.);')}
          <br />
          {t('- access control (if you have team);')}
        </div>

        <label className="mt-5 mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {t('Workspace name')}
        </label>
        <Input
          value={workspaceName}
          onChange={(e) => {
            setIsWorkspaceNameEdited(true);
            setWorkspaceName(e.target.value);
          }}
          placeholder={t('Enter workspace name')}
          disabled={isCreating}
          onPressEnter={handleCreateWorkspace}
          autoFocus
        />
      </div>
    </Modal>
  );
};
