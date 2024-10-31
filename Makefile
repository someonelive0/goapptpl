
PREFIX=goapptpl-1.0.0

# go command for linux and windows.
GO=CGO_ENABLED=0 go
PARAMS=-ldflags '-s -w -extldflags "-static"'

# upx is a tool to compress executable program.
UPX=upx

PRGS=goapptpl


all:	$(PRGS)

goapptpl:
	$(GO) build $(PARAMS) -o $@ ./apptpl


clean:
	rm -f $(PRGS)

install:
	#$(UPX) $(PRGS) || echo $?
	mkdir -p $(PREFIX)/etc $(PREFIX)/tool
	cp -a $(PRGS) $(PREFIX)
	cp -a etc/*.tpl $(PREFIX)/etc
	cp -a tool/*.sql tool/*.service tool/Dockerfile $(PREFIX)/tool

tar: install
	tar cvfz goapptpl-1.0.0.tar.gz $(PREFIX)

# GO 交叉编译说明
# GOOS：目标平台的操作系统（darwin、freebsd、linux、windows）
# GOARCH：目标平台的体系架构（386、amd64、arm）
# 交叉编译不支持 CGO 所以要禁用它

# Mac 下编译 Linux 和 Windows 64位可执行程序
# CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
# CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build main.go

# Linux 下编译 Mac 和 Windows 64位可执行程序
# CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build main.go
# CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build main.go

# Windows 下编译 Mac 和 Linux 64位可执行程序
# SET CGO_ENABLED=0
# SET GOOS=darwin
# SET GOARCH=amd64
# go build main.go

# SET CGO_ENABLED=0
# SET GOOS=linux
# SET GOARCH=amd64
# go build main.go
