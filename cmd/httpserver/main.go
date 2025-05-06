package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Ciobi0212/httpfromtcp/internal/headers"
	"github.com/Ciobi0212/httpfromtcp/internal/request"
	"github.com/Ciobi0212/httpfromtcp/internal/response"
	"github.com/Ciobi0212/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	server.Router.AddHandler(response.GET, "/yourproblem", handleYourProblem)
	server.Router.AddHandler(response.GET, "/myproblem", handleMyProblem)
	server.Router.AddHandler(response.GET, "/httpbin/html", handleHttpBin)
	server.Router.AddHandler(response.GET, "/video", handleVideo)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handleVideo(res response.ResponseWriter, req *request.Request) *response.HandlerError {
	video, err := os.ReadFile("/home/ciobi/projects/boot.dev/httpfromtcp/assets/vim.mp4")
	if err != nil {
		log.Println(err)
		return &response.HandlerError{
			StatusCode: response.InternalServerError,
			Message:    "error reading video into memory",
		}
	}

	headers := headers.NewHeaders()
	headers.Add("Content-type", "video/mp4")
	res.WriteHeaders(response.Ok, headers)

	res.WriteBody(video)

	return nil
}

func handleYourProblem(res response.ResponseWriter, req *request.Request) *response.HandlerError {
	h := headers.NewHeaders()

	body := `<html>
  <head>
	<title>400 Bad Request</title>
  </head>
  <body>
	<h1>Bad Request</h1>
	<p>Your request honestly kinda sucked. <3</p>
  </body>
</html>`
	bytes := []byte(body)

	h.Add("Content-type", "text/html")
	h.Add("Content-length", strconv.Itoa(len(bytes)))

	res.WriteHeaders(response.BadRequest, h)
	res.WriteBody(bytes)

	return nil
}

func handleHttpBin(res response.ResponseWriter, req *request.Request) *response.HandlerError {
	resHttpBin, err := http.Get("https://httpbin.org/html")
	if err != nil {
		return &response.HandlerError{
			StatusCode: response.InternalServerError,
			Message:    "couldn't get response from http bin server",
		}
	}

	internalHeaders := headers.NewHeaders()
	for key, values := range resHttpBin.Header {
		for _, value := range values {
			internalHeaders.Add(key, value)
		}
	}

	internalHeaders.Del("Content-length")
	internalHeaders.Add("Transfer-encoding", "chunked")

	internalHeaders.Add("Trailer", "X-Content-SHA256")
	internalHeaders.Add("Trailer", "X-Content-Length")

	res.WriteHeaders(response.Ok, internalHeaders)

	buffer := make([]byte, 1024)
	data := []byte{}

	for {
		n, err := resHttpBin.Body.Read(buffer)
		if errors.Is(err, io.EOF) {
			res.WriteChunkedBodyDone()
			break
		}

		if err != nil {
			return &response.HandlerError{
				StatusCode: response.InternalServerError,
				Message:    "couldn't read bytes from http bin response",
			}
		}

		data = append(data, buffer[:n]...)
		res.WriteChunkedBody(buffer[:n])
	}

	sha := sha256.Sum256(data)

	strSha := hex.EncodeToString(sha[:])

	trailers := headers.NewHeaders()
	trailers.Add("X-Content-SHA256", strSha)
	trailers.Add("X-Content-Length", strconv.Itoa(len(data)))

	res.WriteTrailers(trailers)

	return nil
}

func handleMyProblem(res response.ResponseWriter, req *request.Request) *response.HandlerError {
	h := headers.NewHeaders()

	body := `<html>
  <head>
	<title>500 Internal Server Error</title>
  </head>
  <body>
	<h1>Internal Server Error</h1>
	<p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

	bytes := []byte(body)

	h.Add("Content-type", "text/html")
	h.Add("Content-length", strconv.Itoa(len(bytes)))

	res.WriteHeaders(response.InternalServerError, h)
	res.WriteBody(bytes)

	return nil
}
