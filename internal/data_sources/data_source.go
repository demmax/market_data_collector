package data_sources

type DataSource interface {
	Start() error
	Stop()
	AddTicker(ticker string) error
}
