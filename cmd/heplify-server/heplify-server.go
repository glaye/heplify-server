package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"

	//"net"
	//_ "net/http/pprof"

	"github.com/koding/multiconfig"
	"glaye/heplify-server/config"
	"glaye/heplify-server/logp"
	input "glaye/heplify-server/server"
)

type server interface {
	Run()
	End()
}

func init() {
	var err error
	var logging logp.Logging
	var fileRotator logp.FileRotator

	c := multiconfig.New()
	cfg := new(config.HeplifyServer)
	c.MustLoad(cfg)
	config.Setting = *cfg

	if tomlExists(config.Setting.Config) {
		cf := multiconfig.NewWithPath(config.Setting.Config)
		err := cf.Load(cfg)
		if err == nil {
			config.Setting = *cfg
		} else {
			fmt.Println("Syntax error in toml config file, use flag defaults.", err)
		}
	} else {
		fmt.Println("Could not find toml config file, use flag defaults.", err)
	}

	logp.DebugSelectorsStr = &config.Setting.LogDbg
	logging.Level = config.Setting.LogLvl
	logp.ToStderr = &config.Setting.LogStd
	fileRotator.Path = "./"
	fileRotator.Name = "heplify-server.log"
	logging.Files = &fileRotator

	err = logp.Init("heplify-server", &logging)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func tomlExists(f string) bool {
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		return false
	} else if !strings.Contains(f, ".toml") {
		return false
	}
	return err == nil
}

func main() {
	if config.Setting.Version {
		fmt.Println(config.Version)
		os.Exit(0)
	}
	var wg sync.WaitGroup
	var sigCh = make(chan os.Signal, 1)

	//go http.ListenAndServe(":8181", http.DefaultServeMux)
	debug.SetGCPercent(50)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	hep := input.NewHEPInput()
	servers := []server{hep}
	/* 	autopprof.Capture(autopprof.CPUProfile{
		Duration: 15 * time.Second,
	}) */

	for _, srv := range servers {
		wg.Add(1)
		go func(s server) {
			defer wg.Done()
			s.Run()
		}(srv)
	}

	<-sigCh

	for _, srv := range servers {
		wg.Add(1)
		go func(s server) {
			defer wg.Done()
			s.End()
		}(srv)
	}
	wg.Wait()
}
