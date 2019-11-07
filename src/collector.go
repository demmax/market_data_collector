package main

import (
	"log"
	"os"
	"time"

	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
)

func handleEvent(event interface{}) error {
	log.Println(event)
	return nil
}

func main() {
	const token = "t.C7I8dvLkSGpdLsJYdOeLsqsQ_drTdHmXVo7mF4lzRrrFKZxafy0lPOfRhAet3qyhV08mYRSdKA9MO7C0Cf4-Gw"

	logger := log.New(os.Stdout, "[invest-openapi-go-sdk]", log.LstdFlags)

	streamClient, err := sdk.NewStreamingClient(logger, token)
	if err != nil {
		logger.Fatalln(err)
	}

	go streamClient.RunReadLoop(handleEvent)

	restClient := sdk.NewRestClient(token)
	etfs, err := restClient.SearchInstrumentByTicker("YNDX")
	if err != nil {
		logger.Fatalln(err)
	}

	logger.Println(etfs[0].FIGI)

	streamClient.SubscribeInstrumentInfo(etfs[0].FIGI, "YNDX")
	time.Sleep(3 * time.Second)
	logger.Println("end")
}
