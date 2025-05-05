package response

import (
	"fmt"
	"net"
	"strconv"

	"github.com/Ciobi0212/httpfromtcp/internal/headers"
	"github.com/Ciobi0212/httpfromtcp/internal/request"
)

type HandlerError struct {
	StatusCode StatusCode
	Message    string
}

func (h *HandlerError) Error() string {
	return h.Message
}

type ResponseWriter struct {
	Conn net.Conn
}

type Handler func(w ResponseWriter, req *request.Request) *HandlerError

func (w *ResponseWriter) WriteStatusLine(statusCode StatusCode) error {
	switch statusCode {
	case Ok:
		_, err := w.Conn.Write([]byte("HTTP/1.1 200 OK" + crlf))
		return err
	case BadRequest:
		_, err := w.Conn.Write([]byte("HTTP/1.1 400 Bad Request" + crlf))
		return err
	case InternalServerError:
		_, err := w.Conn.Write([]byte("HTTP/1.1 500 Internal Server Error" + crlf))
		return err
	default:
		_, err := w.Conn.Write([]byte("HTTP/1.1 " + strconv.Itoa(int(statusCode)) + crlf))
		return err

	}
}

func (w *ResponseWriter) WriteHeaders(h headers.Headers) error {
	for key, value := range h {
		str := fmt.Sprintf("%s: %s%s", key, value, crlf)

		_, err := w.Conn.Write([]byte(str))
		if err != nil {
			return fmt.Errorf("failed to write header '%s': %w", key, err)
		}
	}

	_, err := w.Conn.Write([]byte(crlf))
	if err != nil {
		return fmt.Errorf("failed to write final CRLF after headers: %w", err)
	}

	return nil
}

func (w *ResponseWriter) WriteBody(p []byte) error {
	_, err := w.Conn.Write(p)
	if err != nil {
		return fmt.Errorf("failed to write body: %w", err)
	}

	return nil
}

func (w *ResponseWriter) RespondWithHandleError(e *HandlerError) error {
	err := w.WriteStatusLine(e.StatusCode)
	if err != nil {
		return fmt.Errorf("failed to write status line when responding with handle error '%s': %w", e.Error(), err)
	}

	body := e.Error()

	bytes := []byte(body)

	w.WriteHeaders(headers.GetDefaultHeaders(len(bytes)))

	w.WriteBody(bytes)

	return nil
}
