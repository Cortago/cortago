package router

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gookit/color"
)

// StartMessage Consist of startmessage of cortago
var (
	StartMessage = "Cortago is running"
	Port         = 1532
)

// Handle is just like "net/http" Handlers, only takes params.
type Handle func(http.ResponseWriter, *http.Request, url.Values)

// Router name says it all.
type Router struct {
	tree        *node
	rootHandler Handle
}

// New creates a new router. It takes the root (fall through) route
// like how the default mux works. The only difference, you get to specify one.
func New(rootHandler Handle) *Router {
	node := node{component: "/", isNamedParam: false}
	return &Router{tree: &node, rootHandler: rootHandler}
}

// Handle takes an http handler, method, and pattern for a route.
func (r *Router) Handle(path string, handler Handle) {
	if path[0] != '/' {
		panic("Path has to start with a /.")
	}
	r.tree.addNode(path, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	params := req.Form
	color.Red.Print(req.Method)
	fmt.Print("\t")
	color.Cyan.Println(req.URL.Path)
	node, _ := r.tree.traverse(strings.Split(req.URL.Path, "/")[1:], params)
	if handler := node.handles; handler != nil {
		handler(w, req, params)
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

// StartApp starts server on default server
func (r *Router) StartApp() {
	fmt.Println(StartMessage + " on " + strconv.FormatInt(int64(Port), 10))
	http.ListenAndServe(":"+strconv.FormatInt(int64(Port), 10), r)
}

// node represents a struct of each node in the tree.
type node struct {
	children     []*node
	component    string
	isNamedParam bool
	handles      Handle
}

// addNode - adds a node to our tree. Will add multiple nodes if path
// can be broken up into multiple components. Those nodes will have no
// handler implemented and will fall through to the default handler.
func (n *node) addNode(path string, handler Handle) {
	components := strings.Split(path, "/")[1:]
	count := len(components)

	for {
		aNode, component := n.traverse(components, nil)
		if aNode.component == component && count == 1 { // update an existing node.
			aNode.handles = handler
			return
		}
		newNode := node{component: component, isNamedParam: false}

		if len(component) > 0 && component[0] == ':' { // check if it is a named param.
			newNode.isNamedParam = true
		}
		if count == 1 { // this is the last component of the url resource, so it gets the handler.
			newNode.handles = handler
		}
		aNode.children = append(aNode.children, &newNode)
		count--
		if count == 0 {
			break
		}
	}
}

// traverse moves along the tree adding named params as it comes and across them.
// Returns the node and component found.
func (n *node) traverse(components []string, params url.Values) (*node, string) {
	component := components[0]
	if len(n.children) > 0 { // no children, then bail out.
		for _, child := range n.children {
			if component == child.component || child.isNamedParam {
				if child.isNamedParam && params != nil {
					params.Add(child.component[1:], component)
				}
				next := components[1:]
				if len(next) > 0 { // http://xkcd.com/1270/
					return child.traverse(next, params) // tail recursion is it's own reward.
				}
				return child, component
			}
		}
	}
	return n, component
}
