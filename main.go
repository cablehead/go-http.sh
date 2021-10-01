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
	Err     string      `json:"error,omitempty"`
	Content interface{} `json:"content"`
}

func emitPacket(app, err string, content interface{}) {
	packet := &Packet{App: app, Err: err, Content: content}
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
	emitPacket("http.response", "unknown request", response)
}

func main() {
	responses := NewResponses()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			response := new(Response)
			err := json.Unmarshal(scanner.Bytes(), response)
			if err != nil {
				emitPacket("http.response", fmt.Sprintf("malformed: %s", err), scanner.Text())
				continue
			}
			responses.Respond(response.RequestID, response)
		}
		fmt.Println("SCANNER DONE")
		if err := scanner.Err(); err != nil {
			panic(err)
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
		emitPacket("http.request", "", req)

		response := responses.Get(requestID)
		w.Write(response.Body)
		emitPacket("http.response", "", response)
	})

	log.Fatal(http.ListenAndServe(":80", nil))
}
