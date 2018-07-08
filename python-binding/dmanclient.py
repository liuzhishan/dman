import MySQLdb
from DBUtils.PooledDB import PooledDB
from collections import namedtuple
import logging
import ctypes
from functools import wraps
from ctypes import *
import ctypes

FORMAT = '%(asctime)-15s [%(funcName)s] %(message)s'
logging.basicConfig(level=logging.INFO, format=FORMAT)

def singleton(cls):
    instances = {}
    @wraps(cls)
    def getinstance(*args, **kw):
        if cls not in instances:
            instances[cls] = cls(*args, **kw)
        return instances[cls]
    return getinstance

@singleton
class DbManager:
    def __init__(self, host='localhost', port=3306, user='root', \
                 passwd='', db='dman', maxconnections=10):
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

lib = cdll.LoadLibrary("./dbclient.so")

class GoString(ctypes.Structure):
    _fields_ = [("p", c_char_p), ("n", c_longlong)]
    def __init__(self, p='', n=0):
        self.p = p
        self.n = len(p)
        logging.info("p: %s, n: %s", self.p, self.n)

class CGetDbinfo_return(ctypes.Structure):
    _fields_ = [("r0", c_longlong), ("r1", c_char_p), ("r2",  c_char_p), ("r3",  c_char_p), ("r4",  c_char_p),
                ("r5",  c_char_p), ("r6", c_longlong)]

def c_apply(dbkey='', appkey='', workername='', info=''):
    lib.CApply.argtypes = [GoString, GoString, GoString, GoString]
    res = lib.CApply(GoString(dbkey), GoString(appkey),
                     GoString(workername), GoString(info))

    logging.info("c_apply, dbkey: %s, appkey: %s, workername: %s, info: %s, resut: %s",
                 dbkey, appkey, workername, info, res)

    return res

def c_check(dbkey='', appkey=''):
    lib.CCheck.argtypes = [GoString, GoString]
    res = lib.CCheck(GoString(dbkey), GoString(appkey))

    logging.info("dbkey: %s, appkey: %s, res: %s", dbkey, appkey, res)
    return res

def c_get_dbinfo(dbkey='', appkey=''):
    lib.CGetDbinfo.argtypes = [GoString, GoString]
    lib.CGetDbinfo.restype = CGetDbinfo_return
    res = lib.CGetDbinfo(GoString(dbkey), GoString(appkey))

    logging.info("dbkey: %s, appkey: %s, res status: %s, dbkey: %s, hostname: %s, dbname: %s, username: %s",
                 dbkey, appkey, res.r0, res.r1, res.r2, res.r3, res.r4)
    return res

if __name__ == '__main__':
    #c_apply("localhost.dman", "app", "a", "")
    #c_check("localhost.dman", "app")
    c_get_dbinfo("localhost.dman", "app")

