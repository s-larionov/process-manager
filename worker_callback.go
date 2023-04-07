package process

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type CallbackFunc func(ctx context.Context) error

type RetryOnErrorOpt struct {
	CallbackOpt

	Timeout     time.Duration
	MaxAttempts int
}

type CallbackOpt interface {
	isOption() bool
}

type CallbackWorker struct {
	ctx          context.Context
	cancel       context.CancelFunc
	name         string
	cb           CallbackFunc
	isRunning    bool
	lock         sync.Locker
	RetryOnError bool
	Retries      uint
	RetryTimeout time.Duration
	errors       uint
}

func NewCallbackWorker(name string, cb CallbackFunc, opts ...CallbackOpt) *CallbackWorker {
	ctx, cancel := context.WithCancel(context.Background())

	w := &CallbackWorker{
		ctx:    ctx,
		cancel: cancel,
		name:   name,
		cb:     cb,
		lock:   &sync.Mutex{},
	}

	w.applyOpts(opts...)

	return w
}

func (w *CallbackWorker) applyOpts(opts ...CallbackOpt) {
	for _, opt := range opts {
		switch o := opt.(type) {
		case RetryOnErrorOpt:
			w.RetryOnError = true
			w.RetryTimeout = o.Timeout
			w.Retries = uint(o.MaxAttempts)
		}
	}
}

func (w *CallbackWorker) Start() error {
	defer log.Info("callback worker has been stopped", LogFields{"worker": w.name})

	w.lock.Lock()
	if w.isRunning {
		w.lock.Unlock()
		return fmt.Errorf("worker %s is already run", w.name)
	}

	w.isRunning = true
	w.lock.Unlock()

	log.Info("start callback worker", LogFields{"worker": w.name})

	for w.isRunning {
		err := w.start()

		if err == nil {
			return nil
		}

		if err == context.Canceled {
			return nil
		}

		w.errors++
		if w.Retries > 0 && w.errors >= w.Retries {
			return err
		}

		if !w.RetryOnError {
			return err
		}

		log.Error("retrying execution of callback during error", err, LogFields{"worker": w.name})
		<-time.After(w.RetryTimeout)
	}

	return nil
}

func (w *CallbackWorker) start() (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case string:
				err = errors.New(v)
			case error:
				err = v
			default:
				err = errors.New("unknown error")
			}
		}
	}()

	return w.cb(w.ctx)
}

func (w *CallbackWorker) Stop() error {
	w.lock.Lock()
	defer w.lock.Unlock()

	if !w.isRunning {
		return fmt.Errorf("worker %s isn't run", w.name)
	}

	w.cancel()
	w.isRunning = false

	log.Info("worker has got signal for stopping", LogFields{"worker": w.name})

	return nil
}
