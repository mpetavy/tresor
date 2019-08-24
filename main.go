package main

import (
	"context"
	"flag"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/service"
)

// https://github.com/creamdog/gonfig
// https://github.com/sadlil/go-trigger
// https://github.com/antonholmquist/jason

var (
	serverAddress      *string
	serverReadTimeout  *int
	serverWriteTimeout *int

	server *http.Server
)

func init() {
	common.Init("tresor", "0.0.1", "2018", "archive solution", "mpetavy", common.APACHE, "https://github.com/golang/mpetavy/golang/tresor", true, start, stop, nil, 0)

	serverAddress = flag.String("serverport", ":8100", "Server address:port")
	serverReadTimeout = flag.Int("serverreadtimeout", 5000, "Server READ timeout")
	serverWriteTimeout = flag.Int("serverwritetimeout", 5000, "Server READ timeout")
}

func start() error {
	router := mux.NewRouter()
	http.Handle("/", router)

	err := service.InitServices(router)
	if err != nil {
		return err
	}

	server = &http.Server{
		Addr:           *serverAddress,
		Handler:        router,
		ReadTimeout:    time.Duration(*serverReadTimeout) * time.Millisecond,
		WriteTimeout:   time.Duration(*serverWriteTimeout) * time.Millisecond,
		MaxHeaderBytes: 1 << 20,
	}

	pathPrefix := "/static/"
	router.PathPrefix(pathPrefix).Handler(http.StripPrefix(pathPrefix, http.FileServer(http.Dir("./"))))

	go func(err *error) {
		*err = server.ListenAndServe()
	}(&err)

	time.Sleep(time.Millisecond * 500)

	return err
}

func stop() error {
	err := service.StopServices()
	if err != nil {
		return err
	}

	if server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := server.Shutdown(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	defer common.Done()

	common.Run(nil)
}
