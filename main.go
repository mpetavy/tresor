package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/mpetavy/tresor/service/database"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/service"
)

// https://github.com/creamdog/gonfig
// https://github.com/sadlil/go-trigger

// Using generics for database models
// tesseract OCR with word coordinates
// thumbnail generation into DB

var (
	serverAddress      *string
	serverReadTimeout  *int
	serverWriteTimeout *int

	server *http.Server

	codegen = flag.Bool("codegen", false, "code generation")
)

func init() {
	common.Init("0.0.1", "", "", "2018", "archive solution", "mpetavy", fmt.Sprintf("https://github.com/mpetavy/%s", common.Title()), common.GPL2, nil, start, stop, run, 0)

	serverAddress = flag.String("serverport", ":8100", "Server address:port")
	serverReadTimeout = flag.Int("serverreadtimeout", 5000, "Server READ timeout")
	serverWriteTimeout = flag.Int("serverwritetimeout", 5000, "Server READ timeout")
}

func run() error {
	if *codegen {
		return common.ExitOrError(database.Codegen())
	}

	return nil
}

func start() error {
	router := mux.NewRouter()
	http.Handle("/", router)

	err := service.StartServices(router)
	if common.Error(err) {
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
	router.PathPrefix(pathPrefix).Handler(http.StripPrefix(pathPrefix, http.FileServer(http.Dir("./static"))))

	go func(err *error) {
		defer common.UnregisterGoRoutine(common.RegisterGoRoutine(1))

		*err = server.ListenAndServe()
	}(&err)

	time.Sleep(time.Millisecond * 500)

	common.Info(fmt.Sprintf("%s on %s ready", common.Title(), *serverAddress))

	return err
}

func stop() error {
	err := service.StopServices()
	if common.Error(err) {
		return err
	}

	if server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		err := server.Shutdown(ctx)
		if common.Error(err) {
			return err
		}
	}

	return nil
}

func main() {
	defer common.Done()

	common.Run(nil)
}
