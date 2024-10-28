# config file for goapptpl
#

version = "1.0"

host = "0.0.0.0"
port = 3000


[mysql]
  dbtype = "mysql"
  maxopenconns = 0
  maxidleconns = 0
  maxidletime = "60s"
  # Data Source Name
  dsn = [
    "username:password@tcp(localhost:3306)/dbname?charset=utf8&parseTime=True&loc=Local",
    "username:password@tcp(localhost:3306)/dbname?charset=utf8&parseTime=True&loc=Local"
  ]

[minio]
    host = "localhost"
    port = 3306
    user = "root"
    password = "root"
    database = "goapptpl"



[redis]
    host = "localhost"
    port = 6379
    password = ""

[log]
    level = "debug"
    path = "log"
    filename = "goapptpl.log"
    rotate_files = 7
    rotate_mbytes = 10

