import { type Database, MysqlVersion } from '../../../../entity/databases';

interface Props {
  database: Database;
}

const mysqlVersionLabels = {
  [MysqlVersion.MysqlVersion57]: '5.7',
  [MysqlVersion.MysqlVersion80]: '8.0',
  [MysqlVersion.MysqlVersion84]: '8.4',
};

export const ShowMySqlPhysicalSpecificDataComponent = ({ database }: Props) => {
  return (
    <div>
      <div className="mb-1 flex w-full items-center">
        <div className="min-w-[150px]">MySQL version</div>
        <div>
          {database.mysqlPhysical?.version
            ? mysqlVersionLabels[database.mysqlPhysical.version] || database.mysqlPhysical.version
            : ''}
        </div>
      </div>

      <div className="mb-1 flex w-full items-center">
        <div className="min-w-[150px] break-all">Host</div>
        <div>{database.mysqlPhysical?.host || ''}</div>
      </div>

      <div className="mb-1 flex w-full items-center">
        <div className="min-w-[150px]">Port</div>
        <div>{database.mysqlPhysical?.port || ''}</div>
      </div>

      <div className="mb-1 flex w-full items-center">
        <div className="min-w-[150px]">Username</div>
        <div>{database.mysqlPhysical?.username || ''}</div>
      </div>

      <div className="mb-1 flex w-full items-center">
        <div className="min-w-[150px]">Password</div>
        <div>{'*************'}</div>
      </div>

      <div className="mb-1 flex w-full items-center">
        <div className="min-w-[150px]">Data directory</div>
        <div>{database.mysqlPhysical?.dataDir || ''}</div>
      </div>

      <div className="mb-1 flex w-full items-center">
        <div className="min-w-[150px]">Use HTTPS</div>
        <div>{database.mysqlPhysical?.isHttps ? 'Yes' : 'No'}</div>
      </div>

      <div className="mb-1 flex w-full items-center">
        <div className="min-w-[150px]">Binary logging</div>
        <div>{database.mysqlPhysical?.isBinlogEnabled ? 'Enabled' : 'Not detected'}</div>
      </div>

      {database.mysqlPhysical?.serverUuid && (
        <div className="mb-1 flex w-full items-center">
          <div className="min-w-[150px]">Server UUID</div>
          <div>{database.mysqlPhysical.serverUuid}</div>
        </div>
      )}
    </div>
  );
};
