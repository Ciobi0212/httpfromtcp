package response

type StatusCode int

const (
	Ok                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
	NoContent           StatusCode = 204
)

const crlf = "\r\n"
