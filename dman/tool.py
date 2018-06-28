import MySQLdb
from DBUtils.PooledDB import PooledDB
from collections import namedtuple
import logging

from config import config

FORMAT = '%(asctime)-15s [%(funcName)s] %(message)s'
logging.basicConfig(level=logging.INFO, format=FORMAT)

@singleton
class DbManager:
    def __init__(self, host=config.db_host, port=config.db_port, user=config.db_user, \
                 passwd=config.db_passwd, db=config.db_name, maxconnections=10):
        self._pool = PooledDB(MySQLdb, mincached=0, maxcached=10, maxshared=10, maxusage=10000, maxconnections=maxconnections, \
                              host=host, port=port, user=user, passwd=passwd, db=name, charset='utf8')

    def get_conn(self):
        return self._pool.connection()

    def execute_sql(self, sql='', params=[]):
        db = self._pool.connection()
        cursor = db.cursor()

        if len(params) == 0:
            cursor.execute(sql)
        else:
            cursor.execute(sql, params)

        db.commit()

        cursor.close()
        db.close()

    def get_sql_result(self, sql='', params=[]):
        db = self._pool.connection()
        cursor = db.cursor()

        if len(params) == 0:
            cursor.execute(sql)
        else:
            cursor.execute(sql, params)

        db.commit()

        res = []
        if cursor != None:
            for row in cursor:
                res.append(row)

        cursor.close()
        db.close()

        return res

db_manager = DbManager()

def open_db():
    return db_manager.get_conn()

