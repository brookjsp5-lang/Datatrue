import type { MysqlVersion } from './MysqlVersion';

export interface MysqlPhysicalDatabase {
  id: string;
  version: MysqlVersion;

  host: string;
  port: number;
  username: string;
  password: string;
  isHttps: boolean;

  dataDir: string;
  serverUuid?: string;
  isBinlogEnabled?: boolean;
}
