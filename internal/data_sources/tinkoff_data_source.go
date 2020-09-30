package data_sources

import (
	"context"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"github.com/demmax/market_data_collector/internal/utils"
	"strconv"
	"time"
)

const (
	ReadLoopRestartTimeout = 10 * time.Second
)

var logger = utils.Logger

type tinkoffDataSource struct {
	dataChannel  chan utils.MarketData
	token        string
	interval     int64
	tickers      []string
	requestId    int64
	stopping     bool
	streamClient *sdk.StreamingClient
}

func NewTinkoffDataSource(dataChan chan utils.MarketData,
	cfg map[string]interface{}) DataSource {

	token := cfg["token"].(string)

	// For some reason, json.Decoder considers `interval` as float.
	// We assume it is 'int', so just perform explicit conversion.
	interval := int64(cfg["interval"].(float64))

	var tickers []string
	for _, ticker := range cfg["tickers"].([]interface{}) {
		tickers = append(tickers, ticker.(string))
	}

	logger.Print("Initializing tinkoff data source:")
	logger.Printf("token: %s", token)
	logger.Printf("candles interval: %d", interval)
	logger.Print("tickers:")
	logger.Print(tickers)

	return &tinkoffDataSource{
		dataChannel: dataChan,
		token:       token,
		interval:    interval,
		tickers:     tickers,
		stopping:    false,
	}
}

func (source *tinkoffDataSource) Start() error {

	source.stopping = false

	figies, err := getFigiesByTickers(source.token, source.tickers)
	if err != nil {
		logger.Error(err)
		return err
	}

	source.streamClient, err = sdk.NewStreamingClient(logger, source.token)
	if err != nil {
		logger.Error(err)
		return err
	}

	go source.runReadLoop()

	errorCnt := source.subscribeToFigies(figies)
	if errorCnt != 0 {
		logger.Warningf("%d errors were occurred during subscription.", errorCnt)
		logger.Warning("Results may be incomplete.")
	}

	return nil
}

func (source *tinkoffDataSource) Stop() {
	logger.Info("Stopping..")
	source.stopping = true
	_ = source.streamClient.Close()
}

func (source *tinkoffDataSource) AddTicker(ticker string) error {
	figies, err := getFigiesByTicker(source.token, ticker)
	if err != nil {
		logger.Error(err)
		return err
	}

	errorCnt := source.subscribeToFigies(figies)
	if errorCnt != 0 {
		logger.Warningf("%d errors were occurred during subscription.", errorCnt)
		logger.Warning("Results may be incomplete.")
	}

	return nil
}

func (source *tinkoffDataSource) runReadLoop() {
	for {
		logger.Info("Starting read loop.")
		err := source.streamClient.RunReadLoop(eventHandler)
		if source.stopping {
			return
		}

		if err != nil {
			logger.Error(err)
		}

		logger.Warningf("Error while running read loop. Try again after %v...", ReadLoopRestartTimeout)
		time.Sleep(ReadLoopRestartTimeout)
	}
}

func eventHandler(event interface{}) error {

	logger.Println(event)

	switch event.(type) {
	case sdk.InstrumentInfoEvent:
		logger.Println("InstrumentInfoEvent")
		//ev := event.(sdk.InstrumentInfoEvent)
	case sdk.CandleEvent:
		logger.Println("CandleEvent")
	case sdk.OrderBookEvent:
		logger.Println("OrderBookEvent")
	default:
		logger.Println("Unknown event received")
	}
	//source.dataChannel <- eventToMarketData(&event)
	return nil
}

func (source *tinkoffDataSource) subscribeToFigies(figies []string) int {

	logger.Infof("Subscribing to %v...", figies)
	errorsCount := 0

	for _, figi := range figies {
		err := source.streamClient.SubscribeInstrumentInfo(figi, source.nextRequestId())
		if err != nil {
			logger.Error(err)
			errorsCount++
			continue
		}

		err = source.streamClient.SubscribeCandle(figi, secsToCandleInterval(source.interval), source.nextRequestId())
		if err != nil {
			logger.Error(err)
			errorsCount++
			continue
		}

		err = source.streamClient.SubscribeOrderbook(figi, sdk.MaxOrderbookDepth, source.nextRequestId())
		if err != nil {
			logger.Error(err)
			errorsCount++
			continue
		}
	}

	return errorsCount
}

func (source *tinkoffDataSource) nextRequestId() string {
	strId := strconv.FormatInt(source.requestId, 10)
	source.requestId++
	return strId
}

func getFigiesByTicker(token string, ticker string) ([]string, error) {

	var figies []string

	restClient := sdk.NewRestClient(token)
	instruments, err := restClient.SearchInstrumentByTicker(context.TODO(), ticker)
	if err != nil {
		return nil, err
	}

	for _, instrument := range instruments {
		figies = append(figies, instrument.FIGI)
	}

	logger.Infof("FIGIes for ticker '%s': %v", ticker, figies)
	return figies, nil
}

func getFigiesByTickers(token string, tickers []string) ([]string, error) {

	var allFigies []string
	for _, ticker := range tickers {
		figies, err := getFigiesByTicker(token, ticker)
		if err != nil {
			return nil, err
		}

		logger.Infof("FIGIes for ticker '%s': %v", ticker, figies)
		allFigies = append(allFigies, figies...)
	}

	return allFigies, nil
}

func eventToMarketData(event *interface{}) utils.MarketData {
	marketData := utils.MarketData{}
	return marketData
}

func secsToCandleInterval(secs int64) sdk.CandleInterval {
	type interval struct {
		seconds     int64
		sdkInterval sdk.CandleInterval
	}

	intervals := []interval{
		{60, sdk.CandleInterval1Min},
		{120, sdk.CandleInterval2Min},
		{180, sdk.CandleInterval3Min},
		{300, sdk.CandleInterval5Min},
		{600, sdk.CandleInterval10Min},
		{900, sdk.CandleInterval15Min},
		{1800, sdk.CandleInterval30Min},
		{3600, sdk.CandleInterval1Hour},
		{7200, sdk.CandleInterval2Hour},
		{14400, sdk.CandleInterval4Hour},
		{3600 * 24, sdk.CandleInterval1Day},
		{3600 * 24 * 7, sdk.CandleInterval1Week},
		// Month interval seems to be useless, so omit it.
	}

	previousSec := int64(0)
	previousInterval := sdk.CandleInterval("")

	for _, intervalPair := range intervals {
		if secs < intervalPair.seconds {
			mid := (intervalPair.seconds + previousSec) / 2
			if secs < mid {
				if previousInterval != "" {
					return previousInterval
				}
			}
			return intervalPair.sdkInterval
		}
		previousInterval = intervalPair.sdkInterval
		previousSec = intervalPair.seconds
	}

	logger.Println("Too big interval. Use Weekly one.")
	return sdk.CandleInterval1Week
}
