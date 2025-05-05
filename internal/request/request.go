package request

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/Ciobi0212/httpfromtcp/internal/headers"
)

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	PathParams  map[string]string
	State       int
}

func NewRequest() *Request {
	return &Request{
		RequestLine: RequestLine{},
		Headers:     headers.NewHeaders(),
		State:       ReadingRequestLine,
		Body:        []byte{},
	}
}

const bufferSize = 1024

const crlf = "\r\n"

// States of a request
const (
	ReadingRequestLine = iota
	ReadingHeader
	ReadingBody
	Done
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	buffer := make([]byte, bufferSize)
	var accumulatedBytes []byte
	req := NewRequest()
	eofFlag := false

	for req.State != Done {
		bytesConsumed, newState, err := parse(accumulatedBytes, req, eofFlag)

		if err != nil {
			return nil, fmt.Errorf("error parsing: %w", err)
		}

		if bytesConsumed > 0 {
			accumulatedBytes = accumulatedBytes[bytesConsumed:]
			req.State = newState
			continue
		}

		nBytesRead, err := reader.Read(buffer)

		if nBytesRead > 0 {
			accumulatedBytes = append(accumulatedBytes, buffer[:nBytesRead]...)
		}

		if errors.Is(err, io.EOF) && eofFlag {
			return nil, fmt.Errorf("EOF with incomplete data: %w", err)
		}

		if errors.Is(err, io.EOF) && !eofFlag {
			eofFlag = true
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("error reading from io.Reader: %w", err)
		}
	}

	return req, nil
}

func parse(bytes []byte, req *Request, eofFlag bool) (int, int, error) {
	switch req.State {
	case ReadingRequestLine:
		requestLine, bytesConsumed, err := parseRequestLine(bytes)
		if err != nil {
			return 0, ReadingRequestLine, err
		}

		if bytesConsumed > 0 {
			req.RequestLine = requestLine
			return bytesConsumed, ReadingHeader, nil
		}

		if eofFlag {
			return 0, ReadingRequestLine, errors.New("incomplete request line at EOF")
		}

		return 0, ReadingRequestLine, nil

	case ReadingHeader:
		bytesConsumed, doubleCrlfFlag, err := req.Headers.ParseHeader(bytes)

		if err != nil {
			return 0, ReadingHeader, err
		}

		if doubleCrlfFlag {
			if req.Headers.Get("content-length") != "" {
				return len(crlf), ReadingBody, nil
			}

			return len(crlf), Done, nil
		}

		if bytesConsumed > 0 {
			return bytesConsumed, ReadingHeader, nil
		}

		if eofFlag {
			return 0, ReadingRequestLine, errors.New("incomplete header at EOF")
		}

		return 0, ReadingHeader, nil

	case ReadingBody:
		contentLength, err := strconv.Atoi(req.Headers.Get("content-length"))
		if err != nil {
			return 0, ReadingBody, err
		}

		bodyPiece, bytesConsumed := parseBody(bytes)

		if bytesConsumed > 0 {
			req.Body = append(req.Body, bodyPiece...)

			if len(req.Body) > contentLength {
				return bytesConsumed, ReadingBody, errors.New("EOF when reading body, content-length > body length")
			}

			if len(req.Body) == contentLength {
				return bytesConsumed, Done, nil
			}
			return bytesConsumed, ReadingBody, nil
		}

		if eofFlag {
			return 0, ReadingBody, errors.New("EOF when reading body, content-length < body length")
		}

		return 0, ReadingBody, nil

	default:
		return -1, -1, errors.New("unsuported state")
	}
}

func parseBody(bytes []byte) (body []byte, bytesConsumed int) {
	if len(bytes) > 0 {
		return bytes, len(bytes)
	}

	return nil, 0
}

func parseRequestLine(bytes []byte) (RequestLine, int, error) {
	str := string(bytes)

	idx := strings.Index(str, crlf)

	if idx == -1 {
		return RequestLine{}, 0, nil
	}

	substr := str[:idx]

	requestLinesParts := strings.Split(substr, " ")

	if len(requestLinesParts) != 3 {
		return RequestLine{}, 0, errors.New("http line start doesn't have 3 parts: " + substr)
	}

	method, requestPath, httpVersion := requestLinesParts[0], requestLinesParts[1], requestLinesParts[2]

	isAlpha := regexp.MustCompile(`^[A-Za-z]+$`).MatchString
	if !isAlpha(method) {
		return RequestLine{}, 0, errors.New("method contains non-letters: " + method)
	}

	if method != strings.ToUpper(method) {
		return RequestLine{}, 0, errors.New("method contains lower letters: " + method)
	}

	if httpVersion != "HTTP/1.1" {
		return RequestLine{}, 0, errors.New("not supported http version: " + httpVersion)
	}

	reqLine := RequestLine{
		Method:        method,
		RequestTarget: requestPath,
		HttpVersion:   "1.1",
	}

	return reqLine, idx + len(crlf), nil
}
