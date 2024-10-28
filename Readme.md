
















如何使用实时重新加载？
Air 是一款便捷的工具，每当源代码发生更改时，它都会自动重启你的 Go 应用程序，从而使你的开发过程更快、更有效率。

要在 Fiber 项目中使用 Air，请按照以下步骤操作

通过从 GitHub 发布页面下载适用于你操作系统的相应二进制文件或直接从源代码构建该工具来安装 Air。
在你的项目目录中为 Air 创建一个配置文件。此文件可以命名为 .air.toml 或 air.conf。以下是一个与 Fiber 配合使用的示例配置文件
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

使用 Air 启动你的 Fiber 应用程序，在终端中运行以下命令
air

当你对源代码进行更改时，Air 会检测到它们并自动重启应用程序。
