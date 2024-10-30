# 0. 设计思路

见README.md

# 1. Install

安装go-1.20，执行如下命令：
 go env -w GO111MODULE=on
 go env -w GOPROXY=https://goproxy.cn,direct

# 2. Build

Use Makefile to make

go mod tidy
make

# 3. Run

配置文件 etc/goapptpl.toml

./gosvc

# 4. 防火墙

需要打开的防火墙端口如下，使用firewal-cmd表述，端口可以参考配置文件etc/goapptpl.toml.tpl

	port = 3000

firewall-cmd --permanent --add-port=3000/tcp
firewall-cmd --reload


# 5. 数据库采用 mysql8.0


# 6. JSON

Mysql 查询结果转成JSON
一种是一行一条JSON，类似PG的row_to_json()，一种是多行合并成一个数组，类似PG的array_to_json()
时间格式化成RFC3339，使用date_format(log_time, '%Y-%m-%dT%T+08:00')

PG row_to_json() 相对于JSON_OBJECT()
PG array_to_json() JSON_ARRAYAGG(JSON_OBJECT()

例如：
```
SELECT JSON_ARRAYAGG(JSON_OBJECT(
'log_id', log_id ,
'log_time', date_format(log_time, '%Y-%m-%dT%T+08:00'),
'request', request,
'response', response ,
'tag', tag,
'ext', ext)) FROM goapptpl_log_http;
```

反之把JSON串转换成表用 JSON_TABLE()

多表联合查询转成JSON

```
SELECT JSON_OBJECT(
'log_id', a.log_id ,
'list_id', a.list_id ,
'list_name', dlh.list_name ,
'log_time', date_format(a.log_time, '%Y-%m-%dT%T+08:00'),
'request', request,
'response', response ,
'tag', a.tag,
'ext', a.ext) FROM goapptpl_log_http a, goapptpl_list_http dlh where a.list_id = dlh.list_id ;
```


# 7. 如何使用实时重新加载？
Air 是一款便捷的工具，每当源代码发生更改时，它都会自动重启你的 Go 应用程序，从而使你的开发过程更快、更有效率。

要在 Fiber 项目中使用 Air，请按照以下步骤操作

通过从 GitHub 发布页面下载适用于你操作系统的相应二进制文件或直接从源代码构建该工具来安装 Air。
在你的项目目录中为 Air 创建一个配置文件。此文件可以命名为 .air.toml 或 air.conf。以下是一个与 Fiber 配合使用的示例配置文件

```
# .air.toml
root = "."
tmp_dir = "tmp"
[build]
  cmd = "go build -o ./tmp/main ."
  bin = "./tmp/main"
  delay = 1000 # ms
  exclude_dir = ["assets", "tmp", "vendor"]
  include_ext = ["go", "tpl", "tmpl", "html"]
  exclude_regex = ["_test\\.go"]
```

使用 Air 启动你的 Fiber 应用程序，在终端中运行以下命令

```
air
```

当你对源代码进行更改时，Air 会检测到它们并自动重启应用程序。
