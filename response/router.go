package response

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Ciobi0212/httpfromtcp/request"
)

type HttpMethod string

const (
	GET     HttpMethod = "GET"
	PUT     HttpMethod = "PUT"
	POST    HttpMethod = "POST"
	DELETE  HttpMethod = "DELETE"
	OPTIONS HttpMethod = "OPTIONS"
)

type RouterNode struct {
	// E.g : "/{userId}, /{apiKey}"
	IsParamater bool

	// E.g : "/users, /api, /auth"
	Segment string

	// E.g : "userId, apiKey"
	ParamaterName string

	// E.g : "/users -> /profile (/user/profile)"
	Children map[string]*RouterNode

	// E.g : "/users -> /{userId} (/user/{userId})"
	ParamChildren *RouterNode

	// Callbacks for endpoint
	Handlers map[HttpMethod]Handler
}

func NewRouterNode(segment string) *RouterNode {
	return &RouterNode{
		IsParamater:   false,
		Segment:       segment,
		ParamaterName: "",
		Children:      make(map[string]*RouterNode),
		ParamChildren: nil,
		Handlers:      make(map[HttpMethod]Handler),
	}
}

func NewRouterParamNode(paramName string) *RouterNode {
	return &RouterNode{
		IsParamater:   true,
		Segment:       "",
		ParamaterName: paramName,
		Children:      make(map[string]*RouterNode),
		ParamChildren: nil,
		Handlers:      make(map[HttpMethod]Handler),
	}
}

type Router struct {
	Root             *RouterNode
	GlobalMiddleware []Middleware
}

func NewRouter() *Router {
	return &Router{
		// Root doesn't represent any valid endpoint, just used to serve as root of the tree
		Root:             NewRouterNode(""),
		GlobalMiddleware: []Middleware{},
	}
}

func isPathParam(segment string) bool {
	return len(segment) > 2 && (segment[0] == '{' && segment[len(segment)-1] == '}')
}

func (r *Router) Use(mw Middleware) {
	r.GlobalMiddleware = append(r.GlobalMiddleware, mw)
}

func (r *Router) AddHandler(method HttpMethod, path string, h Handler) {
	path = strings.ToLower(path)

	segments := strings.Split(path, "/")
	segments[0] = "/"

	// Start traversing the tree, only adding nodes if they don't exist
	// (Note: A node can exist, but can have a nil handler. e.g: adding endpoint /users/{userId} without already having /users endpoint)
	curNode := r.Root
	for _, segment := range segments {
		if isPathParam(segment) {
			cleanParamName := segment[1 : len(segment)-1]
			if curNode.ParamChildren != nil {
				if curNode.ParamChildren.ParamaterName != cleanParamName {
					log.Panicf("routing conflict when adding handler for path: %s, conflict with paramater path: %s != %s", path, segment, cleanParamName)
				}
				curNode = curNode.ParamChildren
			} else {
				newNode := NewRouterParamNode(cleanParamName)
				curNode.ParamChildren = newNode
				curNode = newNode
			}
			continue
		}

		child, ok := curNode.Children[segment]

		if !ok {
			newNode := NewRouterNode(segment)
			curNode.Children[segment] = newNode
			curNode = newNode
		} else {
			curNode = child
		}
	}

	curNode.Handlers[method] = h
}

func (r *Router) GetHandlerAndPathParamsForPath(method HttpMethod, path string) (Handler, map[string]string, bool) {
	pathParams := make(map[string]string)
	path = strings.ToLower(path)

	segments := strings.Split(path, "/")
	segments[0] = "/"

	curNode := r.Root
	for _, segment := range segments {
		child, ok := curNode.Children[segment]

		if ok {
			curNode = child
		} else if curNode.ParamChildren != nil && !isPathParam(segment) {
			curNode = curNode.ParamChildren
			pathParams[curNode.ParamaterName] = segment
		} else {
			return nil, nil, false
		}

	}

	h, ok := curNode.Handlers[method]
	if !ok {
		pathParams = nil
	}
	return h, pathParams, ok
}

func (r *Router) Handle(res ResponseWriter) {
	defer res.Conn.Close()

	req, err := request.RequestFromReader(res.Conn)
	if err != nil {
		log.Printf("Error handling connection from %s: %v", res.Conn.RemoteAddr(), err)
		return
	}

	// Wrap serveHttp in middleware
	currentHandler := r.serveHttp
	for i := len(r.GlobalMiddleware) - 1; i >= 0; i-- {
		mw := r.GlobalMiddleware[i]
		currentHandler = mw(currentHandler)
	}

	hErr := currentHandler(res, req)

	if hErr != nil {
		res.RespondWithHandleError(hErr)
		return
	}
}

func (r *Router) serveHttp(res ResponseWriter, req *request.Request) *HandlerError {
	method := HttpMethod(req.RequestLine.Method)
	path := strings.ToLower(req.RequestLine.RequestTarget)

	handler, pathParams, ok := r.GetHandlerAndPathParamsForPath(method, path)

	// Retarded default message if path not find, TODO: change this
	if !ok {
		Custom404Response(res)
		return nil
	}

	req.PathParams = pathParams
	return handler(res, req)
}

func PrintRouterTree(node *RouterNode, indent string) {
	segmentDisplay := node.Segment
	if node.IsParamater {
		segmentDisplay = fmt.Sprintf("{%s}", node.ParamaterName)
	}
	if segmentDisplay == "" {
		segmentDisplay = "(root)" // Special case for the actual root
	}

	fmt.Printf("%s%s", indent, segmentDisplay)

	// Print handlers for this node
	if len(node.Handlers) > 0 {
		methods := []string{}
		for method := range node.Handlers {
			methods = append(methods, string(method))
		}
		fmt.Printf(" [%s]", strings.Join(methods, ", "))
	}
	fmt.Println() // Newline after node info

	// Print static children
	for _, child := range node.Children {
		PrintRouterTree(child, indent+"  ") // Increase indent
	}

	// Print parameter child
	if node.ParamChildren != nil {
		PrintRouterTree(node.ParamChildren, indent+"  ") // Increase indent
	}
}

// Custom 404 if path not found
func Custom404Response(res ResponseWriter) {
	body := `<html>
  <head>
  </head>
  <body>
    <h1>404 NOT FOUND</h1>
  </body>
</html>`

	bytes := []byte(body)

	res.Headers.Add("Content-type", "text/html")
	res.Headers.Add("Content-length", strconv.Itoa(len(bytes)))

	res.WriteHeaders(BadRequest)
	res.WriteBody(bytes)
}
