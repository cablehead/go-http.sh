package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

type Packet struct {
	App     string      `json:"app"`
	Content interface{} `json:"content"`
}

func emitPacket(app string, content interface{}) {
	packet := &Packet{App: app, Content: content}
	json.NewEncoder(os.Stdout).Encode(packet)
}

type Request struct {
	Method     string      `json:"method"`
	Header     http.Header `json:"header"`
	RemoteAddr string      `json:"request_addr"`
	RequestURI string      `json:"request_uri"`
	Body       []byte      `json:"body"`
	RequestID  uuid.UUID   `json:"request_id"`
}

type Response struct {
	Body      []byte    `json:"body"`
	RequestID uuid.UUID `json:"request_id"`
}

type Responses struct {
	waiters map[uuid.UUID]chan *Response
}

func NewResponses() *Responses {
	return &Responses{waiters: make(map[uuid.UUID]chan *Response)}
}

// TODO, need to lock waiters
// add timeout
// status codes
func (r *Responses) Get(request_id uuid.UUID) *Response {
	c := make(chan *Response)
	r.waiters[request_id] = c
	return <-c
}

func (r *Responses) Respond(request_id uuid.UUID, response *Response) {
	if c, ok := r.waiters[request_id]; ok {
		c <- response
		return
	}

	// action expected? none, just informational, warning actually
	// TODO: should log something structured
	fmt.Println("unknown request:", request_id)
}

func main() {
	responses := NewResponses()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			response := new(Response)
			err := json.Unmarshal(scanner.Bytes(), response)
			if err != nil {
				// TODO should log something structured
				// action expected? none, just informational, warning actually
				panic(err)
			}
			fmt.Println("GOT RESPONSE", response)
			responses.Respond(response.RequestID, response)
		}
		if err := scanner.Err(); err != nil {
			fmt.Println(err)
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		requestID := uuid.New()

		req := &Request{
			Method:     r.Method,
			Header:     r.Header,
			RemoteAddr: r.RemoteAddr,
			RequestURI: r.RequestURI,
			Body:       body,
			RequestID:  requestID,
		}
		emitPacket("http.request", req)

		response := responses.Get(requestID)
		// TODO: should log that we've responded
		w.Write(response.Body)
	})

	log.Fatal(http.ListenAndServe(":80", nil))
}
