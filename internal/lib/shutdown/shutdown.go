package shutdown

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitForShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
}
