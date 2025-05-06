package response

import (
	"github.com/Ciobi0212/httpfromtcp/internal/request"
)

type HandlerError struct {
	StatusCode StatusCode
	Message    string
}

func (h *HandlerError) Error() string {
	return h.Message
}

type Handler func(w ResponseWriter, req *request.Request) *HandlerError
