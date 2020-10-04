package helper

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

func UnmarshalJSON(r *http.Request, item interface{}) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	err := decoder.Decode(item)
	if err != nil {
		return err
	}
	log.Println(item)
	return nil
}

type Response struct {
	Body       interface{}
	StatusCode int
}

type Handler struct {
	Tkn bool
	H   func(w http.ResponseWriter, r *http.Request) (*Response, error)
}

func NewHandler(H func(w http.ResponseWriter, r *http.Request) (*Response, error)) Handler {
	return Handler{
		Tkn: false,
		H:   H,
	}
}

func NewHandlerWithACL(H func(w http.ResponseWriter, r *http.Request) (*Response, error)) Handler {
	return Handler{
		Tkn: true,
		H:   H,
	}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	resp, err := &Response{}, EmptyErr
	if h.Tkn {
		resp, err = Middleware(w, r, h.H)
	} else {
		resp, err = h.H(w, r)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err != nil {
		parseErr := json.NewEncoder(w).Encode(struct {
			Msg string `json:"message"`
		}{err.Error()})
		if parseErr != nil {
			http.Error(w, `{"message":"`+ parseErr.Error()+`"}`, http.StatusInternalServerError)
		}
		switch e := err.(type) {

		case MiddleHttpError:
			// We can retrieve the status here and write out a specific
			// HTTP status code.
			log.Printf("HTTP %d - %s", e.Status(), err.Error())
			w.WriteHeader(e.Status())
		default:
			// Any error types we don't specifically look out for default
			// to serving a HTTP 500
			log.Printf("HTTP %d - %s", 500, err.Error())
			w.WriteHeader(500)
		}
		return
	}
	w.WriteHeader(resp.StatusCode)
	log.Printf("HTTP %d - %s", resp.StatusCode, resp.Body)
	err = json.NewEncoder(w).Encode(resp.Body)
	if err != nil {
		http.Error(w, `{"message":"`+ err.Error()+`"}`, http.StatusInternalServerError)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// returns our new websocket connection
	return ws, nil
}

func WebSocketMDW(conn *websocket.Conn, ticker time.Duration, cmd interface{}, fn func(interface{}) (interface{}, error)) {
	// we want to kick off a for loop that runs for the
	// duration of our websockets connection
	for {
		// we create a new ticker that ticks every 5 seconds
		ticker := time.NewTicker(ticker)

		// every time our ticker ticks
		for t := range ticker.C {
			// print out that we are updating the stats
			log.Printf("Updating Stats: %+v\n", t)
			// and retrieve the subscribers

			err := conn.ReadJSON(cmd)
			if err != nil {
				log.Println(err)
				return
			}

			items, err := fn(cmd)
			if err != nil {
				log.Println(err)
				continue
			}
			// next we marshal our response into a JSON string
			// and finally we write this JSON string to our WebSocket
			// connection and record any errors if there has been any
			if err := conn.WriteJSON(items); err != nil {
				log.Println(err)
				return
			}
		}
	}
}
