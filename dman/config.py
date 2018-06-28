"""config"""

class DefaultConfig:
    def __init__(self):
        pass

config = DefaultConfig()

config.host='127.0.0.1'
config.port=8333

config.db_host='127.0.0.1'
config.db_port=3306
config.db_user='root'
config.db_passwd=''
config.db_name='dman'

config.file_account='data/account.json'
