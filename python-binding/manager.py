import MySQLdb
from DBUtils.PooledDB import PooledDB
from collections import namedtuple
import fire
import json
import tornado

from tool import *
from config import *

def write_account(filename=config.file_account):
    with open(filename) as f:
        accounts = json.load(f)
        for d in accounts:
            try:
                sql = "replace into dbaccount (hostname, dbname, port, username, password) values (%s, %s, %s, %s, %s)"
                params = [d['hostname', d['dbname'], d['port'], d['username'], d['password']]]
                db_manager.execute_sql(sql=sql, params=params)
            except Exception as e:
                logging.info(e)

def show_apply():
    pass

def approve_apply(applyid=0):
    pass

class HandlerUserApply:
    pass

class HandlerApplyResult:
    pass

def start_server():
    pass

if __name__ == '__main__':
    fire.Fire()

