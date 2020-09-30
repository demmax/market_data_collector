package internal

import (
	"github.com/demmax/market_data_collector/internal/data_sources"
	"github.com/demmax/market_data_collector/internal/utils"
	"log"
)

type DataSourceManager interface {
	Run()
}

type dataSourceManager struct {
	controlChannel chan string
	dataChannel    chan utils.MarketData
	config         map[string]interface{}
	dataSources    []data_sources.DataSource
	logger         log.Logger
}

func NewDataSourceManager(ctrlChan chan string,
	dataChan chan utils.MarketData,
	cfg map[string]interface{}) DataSourceManager {
	return &dataSourceManager{
		controlChannel: ctrlChan,
		dataChannel:    dataChan,
		config:         cfg,
	}
}

func (sourceManager *dataSourceManager) Run() {
	for source, params := range sourceManager.config {
		switch source {
		case "tinkoff":
			sourceParams := params.(map[string]interface{})
			dataSource := data_sources.NewTinkoffDataSource(sourceManager.dataChannel, sourceParams)
			sourceManager.dataSources = append(sourceManager.dataSources, dataSource)
			err := dataSource.Start()
			if err != nil {
				utils.Logger.Panicf("Can't create source: %s", err)
			}
		default:
			panic("No such source!")
		}
	}

	for {
		cmd := <-sourceManager.controlChannel
		if cmd == "exit" {
			panic(sourceManager)
		}
	}
}
