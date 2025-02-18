package main

type config struct {
	addr     string
	env      string
	dbConfig dbConfig
}

type dbConfig struct {
	addr               string
	maxOpenConnections int
	maxIdleConnections int
	maxIdleTime        string
}
