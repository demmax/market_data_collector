package main

import (
	"encoding/json"
	"fmt"
	"github.com/demmax/market_data_collector/internal"
	"github.com/demmax/market_data_collector/internal/utils"
	"log"
	"os"
	"time"
)

func main() {
	configFileName := "./config.json"
	file, err := os.Open(configFileName)

	logger := utils.Logger

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

	log.Printf("Using config version %s", config["version"])

	ctrlChan := make(chan string)
	dataChan := make(chan utils.MarketData)

	dataSourcesCfg := config["data_sources"].(map[string]interface{})
	dataManager := internal.NewDataSourceManager(ctrlChan, dataChan, dataSourcesCfg)

	go dataManager.Run()
	//ctrlChan <- "exit"
	time.Sleep(3 * time.Second)
	logger.Println("end")
}
