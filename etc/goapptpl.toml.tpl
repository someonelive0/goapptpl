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
    # 连接池里面的连接最大空闲时长。
    # 当连接持续空闲时长达到maxIdleTime后，该连接就会被关闭并从连接池移除，
    # 哪怕当前空闲连接数已经小于SetMaxIdleConns(maxIdleConns)设置的值
    maxidletime = "60s"

    # Data Source Name
    dsn = [
      "username:password@tcp(localhost:3306)/dbname?charset=utf8&parseTime=True&loc=Local&timeout=10s&readTimeout=0s&writeTimeout=0s",
      "username:password@tcp(localhost:3306)/dbname?charset=utf8&parseTime=True&loc=Local&timeout=10s&readTimeout=0s&writeTimeout=0s"
    ]

[minio]
    addr = "localhost:9000"
    user = "username"
    password = "password"
    ssl = false
    # timeout of seconds, default 10s
    timeout = 10


[redis]
    addr = "localhost:6379"
    password = ""
    db = 0
    timeout = 10


[clickhouse]
    dbtype = "clickhouse"
    # 设置最大开放连接数，注意该值为小于0或等于0指的是无限制连接数
    maxopenconns = 0
    # 设置空闲连接数，将此值设置为小于或等于0将意味着不保留空闲连接，即立即关闭连接
    maxidleconns = 10
    # 连接池里面的连接最大空闲时长。
    # 当连接持续空闲时长达到maxIdleTime后，该连接就会被关闭并从连接池移除，
    # 哪怕当前空闲连接数已经小于SetMaxIdleConns(maxIdleConns)设置的值
    maxidletime = "60s"

    # Data Source Name
    dsn = [
      "clickhouse://default:password@localhost:9000/default?dial_timeout=2000ms&max_execution_time=60s",
      "clickhouse://default:password@localhost:9000/default?dial_timeout=2000ms&max_execution_time=60s"
    ]


[log]
    # log level = trace|debug|info|warn|error|fatal|panic, default info
    level = "info"
    path = "log"
    filename = "goapptpl.log"
    rotate_files = 7
    rotate_mbytes = 10

