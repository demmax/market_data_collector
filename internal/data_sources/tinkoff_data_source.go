package data_sources

import (
	"context"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"github.com/demmax/market_data_collector/internal/utils"
)

type tinkoffDataSource struct {
	dataChannel chan utils.MarketData
	token       string
	interval    int64
	tickers     []string
}

func NewTinkoffDataSource(dataChan chan utils.MarketData,
	cfg map[string]interface{}) DataSource {

	token := cfg["token"].(string)

	// For some reason, json.Decoder consider `interval` as float.
	// We assume it's 'int', so just perform explicit conversion.
	interval := int64(cfg["interval"].(float64))

	var tickers []string
	for _, ticker := range cfg["tickers"].([]interface{}) {
		tickers = append(tickers, ticker.(string))
	}

	return &tinkoffDataSource{
		dataChannel: dataChan,
		token:       token,
		interval:    interval,
		tickers:     tickers,
	}
}

func (source *tinkoffDataSource) Start() {
	logger := utils.Logger

	logger.Print(source.token)
	logger.Print(source.interval)
	logger.Print(source.tickers)

	streamClient, err := sdk.NewStreamingClient(logger, source.token)
	if err != nil {
		logger.Fatalln(err)
	}

	figies, err := retrieveFigiesByTickers(source.token, source.tickers)
	if err != nil {
		logger.Fatalln(err)
	}

	handleEvent := func(event interface{}) error {
		utils.Logger.Println(event)
		source.dataChannel <- eventToMarketData(event)
		return nil
	}
	go streamClient.RunReadLoop(handleEvent)

	for _, figi := range figies {
		streamClient.SubscribeInstrumentInfo(figi, figi)
	}
}

func retrieveFigiesByTickers(token string, tickers []string) ([]string, error) {

	var allFigies []string

	restClient := sdk.NewRestClient(token)
	for _, ticker := range tickers {
		instruments, err := restClient.SearchInstrumentByTicker(context.TODO(), ticker)
		if err != nil {
			return nil, err
		}

		for _, instrument := range instruments {
			allFigies = append(allFigies, instrument.FIGI)
		}
	}

	return allFigies, nil
}

func eventToMarketData(event interface{}) utils.MarketData {
	marketData := utils.MarketData{}
	return marketData
}
