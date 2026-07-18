import { Button, Input, InputNumber, Switch } from 'antd';
import { useEffect, useState } from 'react';

import { type Database, databaseApi } from '../../../../entity/databases';
import { ToastHelper } from '../../../../shared/toast';

interface Props {
  database: Database;

  isShowCancelButton?: boolean;
  onCancel: () => void;

  isShowBackButton: boolean;
  onBack: () => void;

  saveButtonText?: string;
  isSaveToApi: boolean;
  onSaved: (database: Database) => void;

  isShowDbName?: boolean;
}

export const EditMySqlPhysicalSpecificDataComponent = ({
  database,

  isShowCancelButton,
  onCancel,

  isShowBackButton,
  onBack,

  saveButtonText,
  isSaveToApi,
  onSaved,
}: Props) => {
  const [editingDatabase, setEditingDatabase] = useState<Database>();
  const [isSaving, setIsSaving] = useState(false);

  const [isConnectionTested, setIsConnectionTested] = useState(false);
  const [isTestingConnection, setIsTestingConnection] = useState(false);
  const [isConnectionFailed, setIsConnectionFailed] = useState(false);

  const trimDatabase = (databaseToTrim: Database): Database => {
    if (!databaseToTrim.mysqlPhysical) return databaseToTrim;

    return {
      ...databaseToTrim,
      mysqlPhysical: {
        ...databaseToTrim.mysqlPhysical,
        host: databaseToTrim.mysqlPhysical.host?.trim(),
        username: databaseToTrim.mysqlPhysical.username?.trim(),
        password: databaseToTrim.mysqlPhysical.password?.trim(),
        dataDir: databaseToTrim.mysqlPhysical.dataDir?.trim(),
      },
    };
  };

  const testConnection = async () => {
    if (!editingDatabase?.mysqlPhysical) return;

    setIsTestingConnection(true);
    setIsConnectionFailed(false);

    try {
      const trimmedDatabase = trimDatabase(editingDatabase);
      await databaseApi.testDatabaseConnectionDirect(trimmedDatabase);
      setEditingDatabase(trimmedDatabase);
      setIsConnectionTested(true);
      ToastHelper.showToast({
        title: 'Connection test passed',
        description: 'MySQL physical backup prerequisites were verified',
      });
    } catch (e) {
      setIsConnectionFailed(true);
      alert((e as Error).message);
    }

    setIsTestingConnection(false);
  };

  const saveDatabase = async () => {
    if (!editingDatabase?.mysqlPhysical) return;

    const trimmedDatabase = trimDatabase(editingDatabase);

    if (isSaveToApi) {
      setIsSaving(true);

      try {
        await databaseApi.updateDatabase(trimmedDatabase);
      } catch (e) {
        alert((e as Error).message);
      }

      setIsSaving(false);
    }

    onSaved(trimmedDatabase);
  };

  useEffect(() => {
    setIsSaving(false);
    setIsConnectionTested(false);
    setIsTestingConnection(false);
    setIsConnectionFailed(false);
    setEditingDatabase({ ...database });
  }, [database]);

  if (!editingDatabase) return null;

  let isAllFieldsFilled = true;
  if (!editingDatabase.mysqlPhysical?.host) isAllFieldsFilled = false;
  if (!editingDatabase.mysqlPhysical?.port) isAllFieldsFilled = false;
  if (!editingDatabase.mysqlPhysical?.username) isAllFieldsFilled = false;
  if (!editingDatabase.id && !editingDatabase.mysqlPhysical?.password) isAllFieldsFilled = false;
  if (!editingDatabase.mysqlPhysical?.dataDir) isAllFieldsFilled = false;

  const isLocalhostDb =
    editingDatabase.mysqlPhysical?.host?.includes('localhost') ||
    editingDatabase.mysqlPhysical?.host?.includes('127.0.0.1');

  return (
    <div>
      <div className="mb-1 flex w-full items-center">
        <div className="min-w-[150px]">Host</div>
        <Input
          value={editingDatabase.mysqlPhysical?.host}
          onChange={(e) => {
            if (!editingDatabase.mysqlPhysical) return;

            setEditingDatabase({
              ...editingDatabase,
              mysqlPhysical: {
                ...editingDatabase.mysqlPhysical,
                host: e.target.value.trim().replace('https://', '').replace('http://', ''),
              },
            });
            setIsConnectionTested(false);
          }}
          size="small"
          className="max-w-[200px] grow"
          placeholder="Enter MySQL host"
        />
      </div>

      {isLocalhostDb && (
        <div className="mb-1 flex">
          <div className="min-w-[150px]" />
          <div className="max-w-[200px] text-xs text-gray-500 dark:text-gray-400">
            Please{' '}
            <a
              href="https://databasus.com/faq/localhost"
              target="_blank"
              rel="noreferrer"
              className="!text-blue-600 dark:!text-blue-400"
            >
              read this document
            </a>{' '}
            to study how to backup local database
          </div>
        </div>
      )}

      <div className="mb-1 flex w-full items-center">
        <div className="min-w-[150px]">Port</div>
        <InputNumber
          type="number"
          value={editingDatabase.mysqlPhysical?.port}
          onChange={(value) => {
            if (!editingDatabase.mysqlPhysical || value === null) return;

            setEditingDatabase({
              ...editingDatabase,
              mysqlPhysical: { ...editingDatabase.mysqlPhysical, port: value },
            });
            setIsConnectionTested(false);
          }}
          size="small"
          className="max-w-[200px] grow"
          placeholder="Enter MySQL port"
        />
      </div>

      <div className="mb-1 flex w-full items-center">
        <div className="min-w-[150px]">Username</div>
        <Input
          value={editingDatabase.mysqlPhysical?.username}
          onChange={(e) => {
            if (!editingDatabase.mysqlPhysical) return;

            setEditingDatabase({
              ...editingDatabase,
              mysqlPhysical: { ...editingDatabase.mysqlPhysical, username: e.target.value.trim() },
            });
            setIsConnectionTested(false);
          }}
          size="small"
          className="max-w-[200px] grow"
          placeholder="Enter MySQL username"
        />
      </div>

      <div className="mb-1 flex w-full items-center">
        <div className="min-w-[150px]">Password</div>
        <Input.Password
          value={editingDatabase.mysqlPhysical?.password}
          onChange={(e) => {
            if (!editingDatabase.mysqlPhysical) return;

            setEditingDatabase({
              ...editingDatabase,
              mysqlPhysical: { ...editingDatabase.mysqlPhysical, password: e.target.value },
            });
            setIsConnectionTested(false);
          }}
          size="small"
          className="max-w-[200px] grow"
          placeholder="Enter MySQL password"
          autoComplete="off"
          data-1p-ignore
          data-lpignore="true"
          data-form-type="other"
        />
      </div>

      <div className="mb-1 flex w-full items-center">
        <div className="min-w-[150px]">Data directory</div>
        <Input
          value={editingDatabase.mysqlPhysical?.dataDir}
          onChange={(e) => {
            if (!editingDatabase.mysqlPhysical) return;

            setEditingDatabase({
              ...editingDatabase,
              mysqlPhysical: { ...editingDatabase.mysqlPhysical, dataDir: e.target.value.trim() },
            });
            setIsConnectionTested(false);
          }}
          size="small"
          className="max-w-[200px] grow"
          placeholder="/var/lib/mysql"
        />
      </div>

      <div className="mb-3 flex w-full items-center">
        <div className="min-w-[150px]">Use HTTPS</div>
        <Switch
          checked={editingDatabase.mysqlPhysical?.isHttps}
          onChange={(checked) => {
            if (!editingDatabase.mysqlPhysical) return;

            setEditingDatabase({
              ...editingDatabase,
              mysqlPhysical: { ...editingDatabase.mysqlPhysical, isHttps: checked },
            });
            setIsConnectionTested(false);
          }}
          size="small"
        />
      </div>

      <div className="mt-5 flex">
        {isShowCancelButton && (
          <Button className="mr-1" danger ghost onClick={() => onCancel()}>
            Cancel
          </Button>
        )}

        {isShowBackButton && (
          <Button className="mr-auto" type="primary" ghost onClick={() => onBack()}>
            Back
          </Button>
        )}

        {!isConnectionTested && (
          <Button
            type="primary"
            onClick={() => testConnection()}
            loading={isTestingConnection}
            disabled={!isAllFieldsFilled}
            className="mr-5"
          >
            Test connection
          </Button>
        )}

        {isConnectionTested && (
          <Button
            type="primary"
            onClick={() => saveDatabase()}
            loading={isSaving}
            disabled={!isAllFieldsFilled}
            className="mr-5"
          >
            {saveButtonText || 'Save'}
          </Button>
        )}
      </div>

      {isConnectionFailed && (
        <div className="mt-3 text-sm text-gray-500 dark:text-gray-400">
          Make sure MySQL 8.0 or 8.4 is reachable, binary logging is enabled, and the configured
          user can inspect the server.
        </div>
      )}
    </div>
  );
};
