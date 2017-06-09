package main

import "net/http"
import "io"
import "time"

//Must use Newserver in order to instantiate.
type server struct {
	http_server *http.Server
	listen_addr string
}

func index(writer http.ResponseWriter, req *http.Request) {
	io.WriteString(writer, "<h1>Hello world!</h1>")
}

func NewServer(listen_addr string) server {
	http_server := &http.Server {
		Addr: listen_addr,
		ReadTimeout: 5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	serv := server{
		http_server: http_server,
		listen_addr: listen_addr,
	}
	serv.initHandlers()
	return serv
}

func (serv *server) initHandlers() {
	http.HandleFunc("/", index)
}

func (serv *server) start() {
	serv.http_server.ListenAndServe()
}
