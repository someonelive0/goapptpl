package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"goapptol/utils"
)

var (
	param_debug   = flag.Bool("D", false, "debug")
	param_version = flag.Bool("v", false, "version")
	param_config  = flag.String("f", "etc/goapptpl.toml", "config filename")
	START_TIME    = time.Now()
	myconfig      *MyConfig
)

func init() {
	// parse command line
	flag.Parse()
	if *param_version {
		fmt.Print(utils.APP_BANNER)
		fmt.Printf("%s\n", "1.0.0")
		os.Exit(0)
	}
	utils.ShowBannerForApp("goapptpl", utils.APP_VERSION, utils.BUILD_TIME)
	utils.Chdir2PrgPath()
	pwd, _ := utils.GetPrgDir()
	fmt.Println("pwd:", pwd)

	// load config
	config, err := LoadConfig(*param_config)
	if err != nil {
		fmt.Printf("loadConfig error %s\n", err)
		os.Exit(1)
	}
	myconfig = config

	// init log
	err = utils.InitLogRotate(myconfig.LogConfig.Path, myconfig.LogConfig.Filename,
		myconfig.LogConfig.Level,
		myconfig.LogConfig.Rotate_files,
		myconfig.LogConfig.Rotate_mbytes)
	if err != nil {
		fmt.Printf("InitLogRotate error %s\n", err)
		os.Exit(1)
	}

	log.Infof("BEGIN... %v, config=%v, debug=%v",
		START_TIME.Format("2006-01-02 15:04:05"), *param_config, *param_debug)
	log.Debugf("MyConfig: %s", myconfig.Dump())
}

func main() {
	var done = make(chan bool, 2)
	var wg sync.WaitGroup

	// start api server of gofiber
	var apiServer = &ApiServer{Myconfig: myconfig}
	wg.Add(1)
	go func() {
		defer wg.Done()
		apiServer.Start()
	}()

	// set signal, when signaled then sent a message to done channel
	setSignal(done)

	// GracefullyExit, stop and wait all routines
	gracefullyExit := func() {
		log.Info("GracefullyExit")
		apiServer.Stop()
		wg.Wait()
	}

	// 主程序阻塞在这里，等待API Server启动失败，或者是Ctrl-C
	<-done
	gracefullyExit()
	log.Infof("END... %v", time.Now().Format("2006-01-02 15:04:05"))
}

func setSignal(done chan bool) {
	var signchan = make(chan os.Signal, 1)
	signal.Notify(signchan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		s := <-signchan
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			signal.Stop(signchan)
			done <- true
			log.Info("receive SIGNAL: ", s)
		default:
			log.Info("receive other signal, ignore", s)
		}
	}()
}
