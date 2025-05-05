package main

import (
	"fmt"
	"log"
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
	server.Router.AddHandler(response.GET, "/yourproblem/{name}", handleYourProblemWithName)
	response.PrintRouterTree(server.Router.Root, "")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handleYourProblem(res response.ResponseWriter, req *request.Request) *response.HandlerError {

	res.WriteStatusLine(response.BadRequest)
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

	res.WriteHeaders(h)
	res.WriteBody(bytes)

	return nil
}

func handleYourProblemWithName(res response.ResponseWriter, req *request.Request) *response.HandlerError {
	name := req.PathParams["name"]

	res.WriteStatusLine(response.BadRequest)
	h := headers.NewHeaders()

	body := fmt.Sprintf(`<html>
	<head>
	  <title>400 Bad Request</title>
	</head>
	<body>
	  <h1>Bad Request</h1>
	  <p>Your request honestly kinda sucked %s. </p> 
	</body>
  </html>`, name)

	bytes := []byte(body)

	h.Add("Content-type", "text/html")
	h.Add("Content-length", strconv.Itoa(len(bytes)))

	res.WriteHeaders(h)
	res.WriteBody(bytes)

	return nil
}

func handleMyProblem(res response.ResponseWriter, req *request.Request) *response.HandlerError {
	res.WriteStatusLine(response.InternalServerError)
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

	res.WriteHeaders(h)
	res.WriteBody(bytes)

	return nil
}
