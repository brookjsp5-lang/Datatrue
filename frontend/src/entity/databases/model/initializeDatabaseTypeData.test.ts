import { describe, expect, it } from 'vitest';

import type { Database } from './Database';
import { DatabaseType } from './DatabaseType';
import { initializeDatabaseTypeData } from './initializeDatabaseTypeData';

const baseDatabase = {
  id: '',
  workspaceId: 'workspace-id',
  name: 'Source',
  type: DatabaseType.POSTGRES_LOGICAL,
  notifiers: [],
} satisfies Database;

describe('initializeDatabaseTypeData', () => {
  it('initializes MySQL physical data and clears other engine data', () => {
    const database = initializeDatabaseTypeData({
      ...baseDatabase,
      type: DatabaseType.MYSQL_PHYSICAL,
      mysql: {
        id: 'mysql-id',
        version: '8.0',
        host: 'mysql.local',
        port: 3306,
        username: 'logical',
        password: 'secret',
        isHttps: false,
      },
    } as Database);

    expect(database.mysql).toBeUndefined();
    expect(database.mysqlPhysical).toEqual({
      dataDir: '/var/lib/mysql',
    });
  });
});
