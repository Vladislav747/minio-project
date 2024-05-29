package main

import (
	"errors"
	"fmt"
	"github.com/Vladislav747/minio-project/internal/config"
	"github.com/Vladislav747/minio-project/internal/file"
	"github.com/Vladislav747/minio-project/internal/file/storage/minio"
	"github.com/Vladislav747/minio-project/pkg/logging"
	"github.com/Vladislav747/minio-project/pkg/shutdown"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	logging.Init()
	logger := logging.GetLogger()
	logger.Println("logger initialized")

	logger.Println("config initializing")
	cfg := config.GetConfig()

	logger.Println("router initializing")
	router := httprouter.New()

	//metricHandler := metrics.NewMetricHandler()
	//metricHandler.Register(router);

	fileStorage, err := minio.NewStorage(cfg.MinIO.Endpoint, cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, logger)
	if err != nil {
		logger.Fatal(err)
	}
	fileService, err := file.NewService(fileStorage, logger)
	if err != nil {
		logger.Fatal(err)
	}
	filesHandler := file.Handler{
		Logger:      logger,
		FileService: fileService,
	}

	filesHandler.Register(router)

	logger.Println("start application")
	start(router, logger, cfg)
}

func start(router *httprouter.Router, logger logging.Logger, cfg *config.Config) {
	var server *http.Server
	var listener net.Listener

	if cfg.Listen.Type == "sock" {
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			logger.Fatal(err)
		}
		socketPath := path.Join(appDir, "server.sock")
		logger.Infof("socket path: %s", socketPath)
		listener, err = net.Listen("unix", socketPath)
		if err != nil {
			logger.Fatal(err)
		}
	} else {
		logger.Infof("bind application to host: %s and port %s", cfg.Listen.BindIP, cfg.Listen.Port)

		var err error
		listener, err = net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Listen.BindIP, cfg.Listen.Port))
		if err != nil {
			logger.Fatal(err)
		}
	}

	server = &http.Server{
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go shutdown.GracefulShutdown([]os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM}, server)

	logger.Println("application initialized and started")

	if err := server.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			logger.Warn("server shutdown")
		default:
			logger.Fatal(err)
		}
	}
}
