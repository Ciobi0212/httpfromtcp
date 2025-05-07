# Custom Go HTTP Server From Scratch

[![Go Version](https://img.shields.io/badge/go-1.24.2+-blue.svg)](https://golang.org/dl/)

 A custom HTTP/1.1 server from the scratch using TCP. It leverages only the standard `net` package for TCP communication. Made for learning purposes.

## ✨ Core Features

*   **Low-Level TCP Handling:** Directly manages TCP connections for HTTP communication.
*   **Concurrent Request Processing:** Handles each incoming client connection in a separate goroutine for concurrent request processing.
*   **HTTP Request Parsing:**
    *   Parses request lines (Method, Target, Version).
    *   Handles HTTP headers.
    *   Processes request bodies with `Content-Length`.
    *   Extracts URL query parameters.
*   **Dynamic Routing:**
    *   Define handlers for static and parameterized paths (e.g., `/users/{id}`).
    *   Differentiate handlers by HTTP method (GET, POST, etc.).
*   **Middleware Support:**
    *   Global middleware that runs for all requests
    *   Built-in CORS middleware with configurable options
    *   Extensible design for custom middleware implementation
*   **Flexible HTTP Response Generation:**
    *   Construct status lines and HTTP headers.
    *   Support for fixed-length and chunked transfer encoding for response bodies.
    *   Ability to send trailers after chunked responses.

## 🛠️ Conceptual Usage

To build an HTTP server using this toolkit, you would typically:

1.  **Initialize the Server Core (`server.Serve(port)`):**
    *   Starts a TCP listener on a specified port.
    *   The core component manages incoming connections, launching a new goroutine for each to handle requests concurrently.
    *   It utilizes a router for dispatching requests.

2.  **Define Request Handlers:**
    *   Create functions that process incoming requests and generate responses.
    *   These handlers receive a `request.Request` object (containing parsed data) and a `response.ResponseWriter` (for sending the reply).
    *   Signature: `func(response.ResponseWriter, *request.Request) *response.HandlerError`

3.  **Configure Global Middleware:**
    *   Apply middleware that should run for all requests using `router.Use(middleware)`.
    *   The toolkit provides built-in middleware like CORS support.
    *   Create your own middleware for logging, authentication, etc.

4.  **Configure Routing:
    *   Use `router.AddHandler(method, path, handlerFunc)` to map URL paths and HTTP methods to your defined handler functions.
    *   The router will parse path parameters (e.g., `{id}`) and query paramaters (e.g., `?name=andrew&isAdmin=false`) and make them available in `request.Request.PathParams` and `request.Request.QueryParams`.

4.    **The `request.Request` object provided to your handler contains:**
        *   `RequestLine`: Method, target URL, HTTP version.
        *   `Headers`: Parsed request headers.
        *   `Body`: Raw request body (if present).
        *   `QueryParams`: Parsed URL query parameters.
        *   `PathParams`: Parameters extracted from the URL path by the router.

5.  **Send Responses (within your handlers using `response` and `headers` packages):**
    *   Use the `response.ResponseWriter` to send the HTTP response:
        *   `WriteHeaders(responseCode, headersMap)`: Send status line with resposne code and headers.
        *   `WriteBody(bodyBytes)`: Send a fixed-length body.
        *   `WriteChunkedBody(chunkBytes)`, `WriteChunkedBodyDone()`, `WriteTrailers(trailerMap)`: For streaming chunked responses.
    *   Use the `headers.NewHeaders()` utility to build your response header map.

## 🚀 Setting Up and Using as a Library

**1. Installation (for users of your library):**
   ```bash
   go get github.com/Ciobi0212/httpfromtcp
   ```

**2. Importing and Using in Their Code:**
   ```go
   // Example main.go in another project
   package main

   import (
       "fmt"
       "strconv"
       "log"
       "os"
       "os/signal"
       "syscall"

       // Import paths for your library packages
       "github.com/Ciobi0212/httpfromtcp/headers"
       "github.com/Ciobi0212/httpfromtcp/request"
       "github.com/Ciobi0212/httpfromtcp/response"
       "github.com/Ciobi0212/httpfromtcp/server"
   )

   // Example Handler
   func handleGreeting(res response.ResponseWriter, req *request.Request) *response.HandlerError {
       name := "Guest"
       if qName, ok := req.QueryParams["name"]; ok {
           name = qName
       }
       body := fmt.Sprintf("Greetings, %s!", name)
       
       h := headers.NewHeaders()
       h.Add("Content-Type", "text/plain")
       h.Add("Content-Length", strconv.Itoa(len(body)))
       
       res.WriteHeaders(response.Ok, h)
       res.WriteBody([]byte(body))
       return nil
   }

   // Example middleware for logging
   func loggingMiddleware(next response.Handler) response.Handler {
       return func(res response.ResponseWriter, req *request.Request) *response.HandlerError {
           log.Printf("%s %s", req.RequestLine.Method, req.RequestLine.RequestTarget)
           return next(res, req)
       }
   }

   func main() {
       srv, err := server.Serve(8080) 
       if err != nil {
           fmt.Printf("Failed to start server: %v\n", err)
           return
       }
       defer srv.Close() // Assuming a Close method exists
       
       // Configure CORS middleware
       corsOptions := response.CorsOptions{
           AllowAllOrigins: true,
           AllowedMethods:  []string{"GET", "POST", "OPTIONS"},
           AllowedHeaders:  []string{"Content-Type", "Authorization"},
       }
       
       // Apply global middleware
       srv.Router.Use(response.NewCORSMiddleware(corsOptions))
       srv.Router.Use(loggingMiddleware)
      
       srv.Router.AddHandler(response.GET, "/greet", handleGreeting)
       // Add more routes...
       
       fmt.Println("Custom server starting on port 8080...")
       // The server.Serve function starts listening in a goroutine.
       // Keep the main goroutine alive, e.g., by waiting for a signal or another mechanism.
       sigChan := make(chan os.Signal, 1)
       signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
       <-sigChan
       log.Println("Server gracefully stopped")
   }
   ```
