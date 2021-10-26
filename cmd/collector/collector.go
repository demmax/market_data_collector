package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/demmax/market_data_collector/internal"
	"github.com/demmax/market_data_collector/internal/utils"
	"log"
	"os"
)

func main() {
	configFileName := flag.String("config", "config.json", "Config file to use")
	file, err := os.Open(*configFileName)

	logger := utils.Logger
	//logger.SetReportCaller(true)
	log.SetOutput(os.Stdout)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	var config map[string]interface{}
	if err := dec.Decode(&config); err != nil {
		s := fmt.Sprintf("Can't parse config file %s: %s", configFileName, err)
		logger.Fatalln(s)
	}

	logger.Printf("Using config version %s", config["version"])

	ctrlChan := make(chan string)
	dataChan := make(chan utils.MarketData)

	dataSourcesCfg := config["data_sources"].(map[string]interface{})
	dataSourceManager := internal.NewDataSourceManager(ctrlChan, dataChan, dataSourcesCfg)

	dataSourceManager.Run()
}
