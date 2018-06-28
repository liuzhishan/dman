CREATE DATABASE IF NOT EXISTS dman;

use dman;

CREATE TABLE IF NOT EXISTS dbaccount (
       hostname VARCHAR(200) NOT NULL COMMENT 'hostname',
       dbname VARCHAR(100) NOT NULL COMMENT 'dbname',
       port INT NOT NULL COMMENT 'port',
       username VARCHAR(100) NOT NULL COMMENT 'username',
       password VARCHAR(100) NOT NULL COMMENT 'password',
       PRIMARY KEY (hostname, dbname, username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS user_apply (
       applyid INT AUTO_INCREMENT COMMENT 'applyid',
       appkey VARCHAR(100) NOT NULL COMMENT 'appkey',
       hostname VARCHAR(200) NOT NULL COMMENT 'hostname',
       dbname VARCHAR(100) NOT NULL COMMENT 'dbname',
       username VARCHAR(100) NOT NULL COMMENT 'username',
       status INT NOT NULL DEFAULT 0 COMMENT '1 for allow, 0 for deny',
       PRIMARY KEY (appkey, hostname, dbname),
       KEY (applyid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
