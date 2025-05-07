package response

import (
	"fmt"
	"net"
	"strconv"

	"github.com/Ciobi0212/httpfromtcp/headers"
)

type ResponseWriter struct {
	Conn    net.Conn
	Headers headers.Headers
}

func (w *ResponseWriter) writeStatusLine(statusCode StatusCode) error {
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

func (w *ResponseWriter) WriteHeaders(statusCode StatusCode) error {
	err := w.writeStatusLine(statusCode)
	if err != nil {
		return fmt.Errorf("failed to write status line when sending headers: %w", err)
	}

	for key, value := range w.Headers {
		str := fmt.Sprintf("%s: %s%s", key, value, crlf)

		_, err := w.Conn.Write([]byte(str))
		if err != nil {
			return fmt.Errorf("failed to write header '%s': %w", key, err)
		}
	}

	_, err = w.Conn.Write([]byte(crlf))
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
	body := e.Error()

	bytes := []byte(body)
	w.Headers = headers.GetDefaultHeaders(len(bytes))
	w.WriteHeaders(e.StatusCode)

	w.WriteBody(bytes)

	return nil
}

func (w *ResponseWriter) WriteChunkedBody(p []byte) error {
	lengthInHex := fmt.Sprintf("%02x", len(p))

	_, err := w.Conn.Write([]byte(lengthInHex + crlf))
	if err != nil {
		return fmt.Errorf("failed to write length of chunked body: %w", err)
	}

	_, err = w.Conn.Write(p)
	if err != nil {
		return fmt.Errorf("failed to write chunked body : %w", err)
	}

	_, err = w.Conn.Write([]byte(crlf))
	if err != nil {
		return fmt.Errorf("failed to write crlf after chunked body: %w", err)
	}

	return nil
}

func (w *ResponseWriter) WriteChunkedBodyDone() error {
	str := "0" + crlf

	_, err := w.Conn.Write([]byte(str))
	if err != nil {
		return fmt.Errorf("failed to write chunked body done: %w", err)
	}

	return nil
}

func (w *ResponseWriter) WriteTrailers(trailers headers.Headers) error {

	for key, value := range trailers {
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
