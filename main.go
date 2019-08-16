package main

import (
	"api/cfg"
	"api/router"
	"api/service"
	"log"
	"net/http"
	"time"
)

func main() {

	conf := cfg.GetConfig() // Get configs from config file
	go func(count time.Duration) {
		for {
			time.Sleep((time.Minute * count) + service.Check())
		}
	}(conf.Timeout)

	srv := &http.Server{
		WriteTimeout: time.Millisecond * 20,
		Addr:         conf.Port,
		Handler:      router.New(),
	}

	log.Fatal(srv.ListenAndServe())
}
