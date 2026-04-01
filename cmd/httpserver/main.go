package main

import (
	"Batman/internal/request"
	"Batman/internal/response"
	"Batman/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handlerFunc)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handlerFunc(w *response.Writer, req *request.Request) {
	endPoint := req.RequestLine.RequestTarget

	if strings.HasPrefix(endPoint, "/httpbin/") {
		endPoint = strings.TrimPrefix(endPoint, "/httpbin")
		res, err := http.Get("https://httpbin.org/" + endPoint)
		if err != nil {
			h := response.GetDefaultHeaders(len(err.Error()))
			w.WriteStatusLine(response.StatusBadRequest)
			w.WriteHeaders(h)
			w.WriteBody([]byte(err.Error()))
			return
		}
		defer res.Body.Close()
		w.WriteStatusLine(response.StatusOK)
		h := response.GetDefaultHeaders(0)
		h.Delete("content-length")
		h.Delete("content-type")
		h.Set("Transfer-Encoding", "chunked")
		w.WriteHeaders(h)
		buf := make([]byte, 32)
		for {
			n, err := res.Body.Read(buf)
			if n > 0 {
				w.WriteChunkedBody(buf[:n])
			}
			if err != nil {
				break
			}
		}
		w.WriteChunkedBodyDone()
	} else {
		switch endPoint {
		case "/yourproblem":
			body := []byte(
				`<html>
  			<head><title>400 Bad Request</title></head>
  			<body>
    			<h1>Bad Request</h1>
    			<p>Your request honestly kinda sucked.</p>
  			</body>
		</html>`)
			h := response.GetDefaultHeaders(len(body))
			h.Override("content-type", "text/html")
			w.WriteStatusLine(response.StatusBadRequest)
			w.WriteHeaders(h)
			w.WriteBody(body)
		case "/myproblem":
			body := []byte(`
		<html>
  			<head>
    			<title>500 Internal Server Error</title>
  			</head>
  			<body>
    			<h1>Internal Server Error</h1>
    			<p>Okay, you know what? This one is on me.</p>
  			</body>
		</html>`)
			h := response.GetDefaultHeaders(len(body))
			h.Override("content-type", "text/html")
			w.WriteStatusLine(response.StatusInternalServerError)
			w.WriteHeaders(h)
			w.WriteBody(body)
		default:
			body := []byte(`
		<html>
  			<head>
    			<title>200 OK</title>  			
			</head>
  			<body>
    			<h1>Success!</h1>
				<p>Your request was an absolute banger.</p> 
			</body>
		</html>`)
			h := response.GetDefaultHeaders(len(body))
			h.Override("content-type", "text/html")
			w.WriteStatusLine(response.StatusOK)
			w.WriteHeaders(h)
			w.WriteBody(body)
		}
	}
}
