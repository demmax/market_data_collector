package utils

import (
	"log"
	"os"
)

var Logger = log.New(os.Stdout, "[invest-openapi-go-sdk]", log.LstdFlags)
