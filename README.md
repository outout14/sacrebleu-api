# Sacrebleu-API 

The API server which interacts with the [sacrebleu-dns DNS server](https://github.com/outout14/sacrebleu-dns) database to create/modify/delete records with token based authentification.

This software interacts directly with the SQL database and does not need to be on a server with ``sacrebleu-dns``.

The API documentation is accessible on ``[your_server]:[port]/``.

## Arguments 
You can show theses informations using ``./sacrebleu-api -h``.
``` 
    -config string
            the patch to the config file (default "extra/config.ini.example")
    -createadmin
            create admin user in the database
    -sqlmigrate
            initialize / migrate the database
``` 

## Configuration 
Variables names are case sensitives.
|Variable name|Type|Example|Informations|
|--|--|--|--|
| AppMode | string|``"production"``|Anything different than ``production`` will show debug messages
| App | Section |
|IP|string|``"127.0.0.1"``|IP address on which the HTTP server must listen. Blank to listen on all IPs 
|Port|int|``5001``|Port on which the HTTP server must listen
|Logfile|bool|``true``|Enable or disable file logs.
|Database|Section|
|Type|string|``"postgresql"``|SQL Database type. ``"postgresql"`` or ``"mysql"`` (anything different than ``"postgresql"`` will rollback to ``"mysql"``)
|Host|string|``"127.0.0.1"``  ``"/var/run/postgres"``|Can be either an IP or a path to a socket for Postgres
|Username|string|``"sacrebleu"``|SQL Database Username
|Password|string|``"superSecretPassword"``|SQL Database Password (optional)
|Port|string|``"5432"``|SQL Database port (``"5432"`` for postgres or ``"3306"`` for MySQL by default)
|Db|string|``"sacrebleudatabase"``|SQL Database Name

## Working 
- All API endpoints (domains, users and records)
- Automatic SOA generation when a record is edited or created 
- Swagger 

## ToDo
- XFR 
- DNSSEC 
- Unit tests 
- Clean up