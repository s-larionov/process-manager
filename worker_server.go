package process

import (
	"context"
	"errors"
	"net/http"
	"sync"
)

var (
	ErrServerIsAlreadyRun = errors.New("server is already run")
	ErrServerIsNotRun     = errors.New("server isn't run")
)

type ServerWorker struct {
	name      string
	server    *http.Server
	isRunning bool
	lock      sync.Locker
}

func NewServerWorker(name string, server *http.Server) *ServerWorker {
	w := ServerWorker{
		name:   name,
		server: server,
		lock:   &sync.Mutex{},
	}

	return &w
}

func (w *ServerWorker) Start() error {
	w.lock.Lock()

	if w.isRunning {
		w.lock.Unlock()

		return ErrServerIsAlreadyRun
	}

	w.isRunning = true

	w.lock.Unlock()

	log.Info("start server worker", LogFields{"listen": w.server.Addr, "worker": w.name})

	err := w.server.ListenAndServe()

	log.Info("server worker has been stopped", LogFields{"worker": w.name})

	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (w *ServerWorker) Stop() error {
	w.lock.Lock()
	defer w.lock.Unlock()

	if !w.isRunning {
		return ErrServerIsNotRun
	}

	err := w.server.Shutdown(context.Background())
	if err == http.ErrServerClosed {
		err = nil
	}

	w.isRunning = false

	log.Info("worker has got signal for stopping", LogFields{"worker": w.name})

	return err
}
