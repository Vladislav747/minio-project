package shutdown

import (
	"github.com/Vladislav747/minio-project/pkg/logging"
	"io"
	"os"
	"os/signal"
)

func GracefulShutdown(signals []os.Signal, closeItems ...io.Closer) {
	logger := logging.GetLogger()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, signals...)
	sig := <-sigc
	logger.Infof("Received signal: %s. Shutting down...", sig)

	// Here we can do graceful shutdown (close connections and etc)
	for _, closer := range closeItems {
		if err := closer.Close(); err != nil {
			logger.Errorf("Error closing %v: %v", closer, err)
		}
	}
}
