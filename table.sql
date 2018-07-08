CREATE DATABASE IF NOT EXISTS dman;

use dman;

CREATE TABLE IF NOT EXISTS dbaccount (
       dbkey varchar(200) NOT NULL DEFAULT '' COMMENT 'dbkey, xxx.dbname',
       hostname VARCHAR(200) NOT NULL DEFAULT '' COMMENT 'hostname',
       dbname VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'dbname',
       port INT NOT NULL COMMENT DEFAULT 0 'port',
       username VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'username',
       password VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'password',
       PRIMARY KEY (dbkey, username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS user_apply (
       applyid INT AUTO_INCREMENT COMMENT 'applyid',
       appkey VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'appkey',
       dbkey VARCHAR(200) NOT NULL DEFAULT '' COMMENT 'dbkey',
       workername VARCHAR(30) NOT NULL DEFAULT '' COMMENT 'worker name',
       status INT NOT NULL DEFAULT 0 COMMENT '1 for allow, 0 for deny',
       info VARCHAR(500) DEFAULT '' COMMENT 'info',
       secretkey VARCHAR(100) DEFAULT '' COMMENT 'secretkey',
       username VARCHAR(100) DEFAULT '' COMMENT 'username for db',
       PRIMARY KEY (appkey, dbkey),
       KEY (applyid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
