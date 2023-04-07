# Process Manager

<a href="https://opensource.org/licenses/Apache-2.0" rel="nofollow"><img src="https://img.shields.io/badge/license-Apache%202-blue" alt="License" style="max-width:100%;"></a>
![unit-tests](https://github.com/s-larionov/process-manager/workflows/unit-tests/badge.svg)

## Installation

```bash
go get -u github.com/s-larionov/process-manager
```

## Example

```go
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/s-larionov/process-manager"
)


func main() {
	manager := process.NewManager()

	// Create a callback worker
	manager.AddWorker(process.NewCallbackWorker("test", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		}
	}))

	// Create a callback worker with retries
	opt := process.RetryOnErrorOpt{
		Timeout:     time.Second, // When it is omitted manager will try to run it immediately
		MaxAttempts: 10,          // When this param is missed manager will try restart in infinity loop
	}
	manager.AddWorker(process.NewCallbackWorker("test with error", func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Millisecond):
				return errors.New("test error")
			}
		}
	}, opt))

	// Create an example of server worker for prometheus
	handler := mux.NewRouter()
	handler.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:    ":2112",
		Handler: handler,
	}
	manager.AddWorker(process.NewServerWorker("prometheus", server))

	manager.StartAll()

	WaitShutdown(manager)
}

func WaitShutdown(manager *process.Manager) {
	go func(manager *process.Manager) {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		<-sigChan

		manager.StopAll()
	}(manager)

	manager.AwaitAll()
}
```
