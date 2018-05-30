package main

import (
	"github.com/ollien/sms-pusher/server/config"
	"github.com/ollien/sms-pusher/server/db"
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	"github.com/ollien/sms-pusher/server/web"
	"github.com/sirupsen/logrus"
)

//Server represents a single instance of the sms-pusher server
type Server struct {
	databaseConnection db.DatabaseConnection
	logger             *logrus.Logger
	upstreamChannel    <-chan firebasexmpp.UpstreamMessage
	sendChannel        chan<- firebasexmpp.DownstreamPayload
	supervisor         XMPPSupervisor
	webserver          web.Webserver
}

//NewServer makes a new sms-pusher server
func NewServer() (Server, error) {
	config, err := config.GetConfig()
	if err != nil {
		return Server{}, err
	}

	logFormatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05-0700",
	}
	logger := logrus.New()
	logger.Formatter = logFormatter

	databaseConnection, err := db.InitDB(logger)
	if err != nil {
		return Server{}, err
	}

	_, err = setup(databaseConnection)
	if err != nil {
		logger.Fatal(err)
	}

	upstreamChannel := make(chan firebasexmpp.UpstreamMessage)
	sendChannel := make(chan firebasexmpp.DownstreamPayload)
	supervisor := NewXMPPSupervisor(upstreamChannel, sendChannel, logger)

	listenAddress := config.Web.GetListenAddress()
	webserver, err := web.NewWebserver(listenAddress, databaseConnection, sendChannel, logger)
	if err != nil {
		return Server{}, err
	}

	return Server{
		databaseConnection: databaseConnection,
		logger:             logger,
		upstreamChannel:    upstreamChannel,
		sendChannel:        sendChannel,
		supervisor:         supervisor,
		webserver:          webserver,
	}, nil
}

//Run starts the Server
func (server Server) Run() {
	go listenForSMS(server.upstreamChannel, server.logger)
	server.supervisor.SpawnClient()
	server.logger.Info("Listening for SMS")
	server.logger.Info("Starting Webserver")
	server.webserver.Server.ListenAndServe()
}

//Stop stops the Server
func (server Server) Stop() error {
	err := server.databaseConnection.Close()
	if err != nil {
		return err
	}

	err = server.webserver.Server.Close()
	if err != nil {
		return err
	}

	return nil
}
