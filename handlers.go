package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"

	"github.com/google/go-github/v29/github"
)

func close(r io.ReadCloser) {
	io.Copy(ioutil.Discard, r)
	r.Close()
}

type handler struct {
	handler         interface{}
	numContext      int
	numInstallation int
	numEvent        int
	returnError     bool
}

func (h *handler) Handle(ctx context.Context, installation *Installation, event interface{}) error {
	size := 0
	if h.numContext > -1 {
		size++
	}
	if h.numInstallation > -1 {
		size++
	}
	if h.numEvent > -1 {
		size++
	}
	args := make([]reflect.Value, size)
	if h.numContext > -1 {
		args[h.numContext] = reflect.ValueOf(ctx)
	}
	if h.numInstallation > -1 {
		args[h.numInstallation] = reflect.ValueOf(installation)
	}
	if h.numEvent > -1 {
		args[h.numEvent] = reflect.ValueOf(event)
	}

	result := reflect.ValueOf(h.handler).Call(args)
	if h.returnError && len(result) > 0 && result[0].Type().Implements(errorType) {
		r := result[0].Interface()
		if r == nil {
			return nil
		}
		return r.(error)
	}
	return nil
}

type Handlers struct {
	AppId        int64
	handlers     map[string]*handler
	ParseWebhook func(eventType string, bs []byte) (interface{}, error)
}

func New() *Handlers {
	return &Handlers{handlers: map[string]*handler{}, ParseWebhook: github.ParseWebHook}
}

func (h *Handlers) On(eventType string, handler interface{}) error {
	err := Validate(handler)
	if err != nil {
		return err
	}
	h.handlers[eventType] = createHandler(handler)
	return nil
}

func (h *Handlers) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	defer close(req.Body)
	eventType := req.Header.Get("X-Github-Event")
	if eventType == "" {
		res.WriteHeader(400)
		return
	}
	bs, err := ioutil.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(500)
		return
	}
	err = h.Handle(req.Context(), eventType, bs)
	if err != nil {
		log.Println(err)
		res.WriteHeader(500)
		return
	}
}

func (h *Handlers) Handle(ctx context.Context, eventType string, bs []byte) error {
	handler, ok := h.handlers[eventType]
	if !ok {
		return fmt.Errorf("handler for %s is not found", eventType)
	}
	i, event, err := h.ParseInstallationWebHook(eventType, bs)
	if err != nil {
		return err
	}
	return handler.Handle(ctx, i, event)
}

func (h *Handlers) ParseInstallationWebHook(eventType string, bytes []byte) (*Installation, interface{}, error) {
	i := &InstallationPart{}
	err := json.Unmarshal(bytes, i)
	if err != nil {
		return nil, nil, err
	}
	event, err := h.ParseWebhook(eventType, bytes)
	return i.Installation, event, err
}

var contextType reflect.Type = reflect.TypeOf((*context.Context)(nil)).Elem()
var errorType reflect.Type = reflect.TypeOf((*error)(nil)).Elem()

func createHandler(h interface{}) *handler {
	typ := reflect.TypeOf(h)
	handler := &handler{
		handler:         h,
		numContext:      -1,
		numInstallation: -1,
		numEvent:        -1,
		returnError:     typ.NumOut() == 1 && typ.Out(0).Implements(errorType),
	}
	for i := 0; i < typ.NumIn(); i++ {
		t := typ.In(i)
		if t.Implements(contextType) {
			handler.numContext = i
		} else if t == reflect.TypeOf((*Installation)(nil)) {
			handler.numInstallation = i
		} else {
			handler.numEvent = i
		}
	}
	return handler
}

func Validate(handler interface{}) error {
	typ := reflect.TypeOf(handler)
	if typ.Kind() != reflect.Func {
		return errors.New("handler must be function")
	}
	if typ.NumIn() == 0 {
		return errors.New("handler has no argument. not supported")
	}
	if typ.NumIn() > 3 {
		return errors.New("handler has too many argument. not supported")
	}
	if typ.NumIn() == 1 && typ.In(0).Implements(contextType) {
		return errors.New("handler has only context argument. not supported")
	}

	if typ.NumOut() > 1 {
		return errors.New("handler return too many value. handler must return error or not")
	}
	if typ.NumOut() == 1 && !typ.Out(0).Implements(errorType) {
		return errors.New("handler must return error or not")
	}

	return nil
}
