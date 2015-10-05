package main

import (
	"encoding/json"
	"fmt"
	hr "github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"path"
	"strings"
)

type Node struct {
	Value    string
	Children map[string]*Node
}

type ConfigServer struct {
	Router *hr.Router
	Root *Node
}

// general error handler to save repeating code
func (cs *ConfigServer) ErrorHandler(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	fmt.Fprintf(w, "Error: %s\n", err)
}

/*
 * FindNode: finds a node in the config
 * takes:
 *   config: a "path" string made of node names
 * returns:
 *   Node pointer to the desired node or nil
 *   error if unable to find node or nil
 */
func (cs *ConfigServer) FindNode(config string) (*Node, error) {
	names := strings.Split(config, "/")

	node := cs.Root
	log.Printf("node: %v", node)
	if node != nil {
		log.Printf("len names: %d  names: %v", len(names), names)
		for _, name := range names {
			if len(name) > 0 {
				node = node.Children[name]
				if node == nil {
					return nil, fmt.Errorf("Config %s not found", config)
				}
			}
		}
		return node, nil
	} else {
		return nil, fmt.Errorf("Config %s not found", config)
	}
}

/*
 * Create: CRUD function to handle creating new node, will fail if node already exists
 * takes:
 *   URI path is used as identifier to node
 * returns:
 *   200 if successful
 *   400 if unknown error
 *   404 if parent node is not found
 *   409 if node already exists
 */
func (cs *ConfigServer) Create(w http.ResponseWriter, r *http.Request, p hr.Params) {
	config := p.ByName("config")[1:]
	node := new(Node)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(node)

	if err == nil {
		cs.ErrorHandler(w, http.StatusBadRequest, err)
	} else if config == "" {
		if cs.Root != nil {
			cs.ErrorHandler(w, http.StatusConflict, err)
		} else {
			cs.Root = node
			w.WriteHeader(http.StatusOK)
		}
	} else {
		config, name := path.Split(config)
		parent, err := cs.FindNode(config)

		if err == nil {
			if parent.Children[name] != nil {
				cs.ErrorHandler(w, http.StatusConflict, err)
			} else {
				parent.Children[name] = node
				w.WriteHeader(http.StatusOK)
			}
		} else {
			cs.ErrorHandler(w, http.StatusNotFound, err)
		}
	}
}

/*
 * Read: CRUD function to handle reading a node
 * takes:
 *   URI path is used as identifier to node
 * returns:
 *   JSON encoded node tree
 *   400 if unknown error
 *   404 if node is not found
 */
func (cs *ConfigServer) Read(w http.ResponseWriter, r *http.Request, p hr.Params) {
	config := p.ByName("config")[1:]
	node, err := cs.FindNode(config)

	if err == nil {
		encoder := json.NewEncoder(w)
		encoder.Encode(node)
	} else {
		cs.ErrorHandler(w, http.StatusNotFound, err)
	}
}

/*
 * Update: CRUD function to handle updating a node, will fail if node does not exist
 * takes:
 *   URI path is used as identifier to node
 * returns:
 *   200 if successful
 *   400 if unknown error
 *   404 if node is not found
 */
func (cs *ConfigServer) Update(w http.ResponseWriter, r *http.Request, p hr.Params) {
	config := p.ByName("config")[1:]
	node := new(Node)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(node)

	if err != nil {
		cs.ErrorHandler(w, http.StatusBadRequest, err)
	} else if config == "" {
		if cs.Root == nil {
			cs.ErrorHandler(w, http.StatusNotFound, err)
		} else {
			cs.Root = node
			w.WriteHeader(http.StatusOK)
		}
	} else {
		config, name := path.Split(config)
		parent, err := cs.FindNode(config)

		if err != nil {
			cs.ErrorHandler(w, http.StatusNotFound, err)
		} else if parent.Children[name] == nil {
			cs.ErrorHandler(w, http.StatusNotFound, err)
		} else {
			delete(parent.Children, name)
			parent.Children[name] = node
			w.WriteHeader(http.StatusOK)
		}
	}
}

/*
 * Delete: http function to handle deleting a node, will fail if node does not exist
 * takes:
 *   URI path is used as identifier to node
 * returns:
 *   200 if successful
 *   400 if unknown error
 *   404 if node is not found
 */
func (cs *ConfigServer) Delete(w http.ResponseWriter, r *http.Request, p hr.Params) {
	config := p.ByName("config")[1:]

	if config == "" {
		cs.Root = nil
		w.WriteHeader(http.StatusOK)
	} else {
		config, name := path.Split(config)
		parent, err := cs.FindNode(config)

		if err == nil {
			if parent.Children[name] == nil {
				cs.ErrorHandler(w, http.StatusNotFound, err)
			} else {
				delete(parent.Children, name)
				w.WriteHeader(http.StatusOK)
			}
		} else {
			cs.ErrorHandler(w, http.StatusNotFound, err)
		}
	}
}

/*
 * Start: run the damn thing
 */
func (cs *ConfigServer) Start() {
	cs.Router.POST("/*config", cs.Create)
	cs.Router.GET("/*config", cs.Read)
	cs.Router.PUT("/*config", cs.Update)
	cs.Router.DELETE("/*config", cs.Delete)

	http.ListenAndServe(":8080", cs.Router)
}

func main() {
	log.Printf("Starting up config server.\n")

	root := &Node {
		Value: "root val",
		Children: map[string]*Node {
			"test1": &Node {
				Value: "whatever",
				Children: map[string]*Node {
					"test1child1": &Node {
						Children: map[string]*Node {
							"test1child2": &Node {
								Value: "blah",
							},
						},
						Value: "another value",
					},
				},
			},
		},
	}

	server := ConfigServer {
		Router: hr.New(),
		Root: root,
	}

	server.start()
}
