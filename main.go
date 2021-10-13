package main

import (
	"os"
	"time"

	"github.com/mosajjal/aio-gw/conf"
	"github.com/mosajjal/aio-gw/web"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

func runDaemon() error {
	// apply the upstream settings
	err := conf.ApplyUpstreamSettings(conf.GlobalUpstreamSettings)
	if err != nil {
		return err
	}
	// apply container settings and start them
	err = conf.ApplyServiceSettings(conf.GlobalServiceSettings)
	if err != nil {
		return err
	}
	// apply web server settings and start the web server
	err = web.ApplyWebServerSettings(conf.GlobalWebserverSettings)
	if err != nil {
		return err
	}

	return nil
}

func main() {

	var dumpConfigFile bool
	var configPath string

	flag.StringVarP(&configPath, "config", "", "", "Use config file")
	flag.BoolVarP(&dumpConfigFile, "dumpConfigFile", "", false, "dump default configuration to stdout")
	flag.Parse()

	if configPath != "" {
		//loading settings from file
		log.Info("Loading settings from file: ", configPath)
		err := conf.LoadSettingsFromFile(configPath)
		if err != nil {
			log.Fatal(err)
		}
	}
	if dumpConfigFile {
		conf.GenerateDefaultSettings()
		os.Exit(0)
	}
	// main daemon
	ticker := time.NewTicker(time.Second * 10)
	go runDaemon()
	for {
		select {
		case <-ticker.C:
			log.Infof("tick")
		}
	}

}
