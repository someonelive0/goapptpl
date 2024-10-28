# config file for goapptpl
#

version = "1.0"

host = "0.0.0.0"
port = 3000


[mysql]
    dbtype = "mysql"
    # 设置最大开放连接数，注意该值为小于0或等于0指的是无限制连接数
    maxopenconns = 0
    # 设置空闲连接数，将此值设置为小于或等于0将意味着不保留空闲连接，即立即关闭连接
    maxidleconns = 10
    maxidletime = "60s"
    # Data Source Name
    dsn = [
      "username:password@tcp(localhost:3306)/dbname?charset=utf8&parseTime=True&loc=Local",
      "username:password@tcp(localhost:3306)/dbname?charset=utf8&parseTime=True&loc=Local"
    ]

[minio]
    addr = "localhost:9000"
    user = "root"
    password = "root"
    ssl = false
    # timeout of seconds, default 10s
    timeout = 10


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

