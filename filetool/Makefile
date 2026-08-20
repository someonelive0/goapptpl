# Makefile for filetool
#

build_time := $(shell date +%FT%T%z)

# go command for linux and windows.
GO=CGO_ENABLED=0 go
#PARAMS=-ldflags '-s -w -extldflags "-static"'
PARAMS=-ldflags '-X main.BUILD_TIME=${build_time} -s -w -extldflags "-static"'

# upx is a tool to compress executable program.
UPX=upx

PRGS=filetool


all:	$(PRGS)

filetool:
	$(GO) build $(PARAMS) -o $@ .


clean:
	rm -f $(PRGS)


PREFIX=../target
install: $(PRGS)
	#$(UPX) $(PRGS) || echo $?
	mkdir -p $(PREFIX)
	cp filetool antiword $(PREFIX)
	cp -a ~/.antiword $(PREFIX)

