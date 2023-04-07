package process

var log Logger = &DummyLog{}

type Logger interface {
	Info(msg string, fields ...LogFields)
	Error(msg string, err error, fields ...LogFields)
}

type LogFields map[string]interface{}

type DummyLog struct {
}

func (DummyLog) Info(_ string, _ ...LogFields)           {}
func (DummyLog) Error(_ string, _ error, _ ...LogFields) {}

func SetLogger(logger Logger) {
	if logger == nil {
		log = &DummyLog{}
		return
	}

	log = logger
}
