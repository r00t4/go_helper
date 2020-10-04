package helper

import (
	"encoding/json"
	"github.com/djumanoff/amqp"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type AMQPResponse struct {
	Message string `json:"msg,omitempty"`
	Error error `json:"err,omitempty"`
	Body interface{} `json:"body,omitempty"`
}

func(amqpResp *AMQPResponse) AMQP() *amqp.Message {
	data, _ := json.Marshal(amqpResp)
	return &amqp.Message{Body: data}
}

func Err (err error) *amqp.Message {
	resp := AMQPResponse{
		Error:   err,
	}
	return resp.AMQP()
}

func OK (body interface{}) *amqp.Message {
	resp := AMQPResponse{
		Body:    body,
	}
	return resp.AMQP()
}

func ListenAndServe(srv amqp.Server) error {
	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)
		// sigterm signal sent from kubernetes
		signal.Notify(sigint, syscall.SIGTERM)

		<-sigint

		log.Println("Shuting down the server")
		// We received an interrupt signal, shut down.
		if err := srv.Stop(); err != nil {
			// Error from closing listeners, or context timeout:
			log.Println("AMQP server Shutdown: ", err.Error())
		}
		close(idleConnsClosed)
	}()

	if err := srv.Start(); err != nil {
		log.Println("AMQP server Start: ", err.Error())
		return err
	}

	<-idleConnsClosed

	return nil
}