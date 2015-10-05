package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
)

type Node struct {
	Value    interface{}
	Children map[string]*Node
}

var Root *Node

/*
 * FindNode: finds a node in the config
 * takes:
 *   config: a "path" string made of node names
 * returns:
 *   Node pointer to the desired node or nil
 *   error if unable to find node or nil
 */
func FindNode(config string) (*Node, error) {
	names := strings.Split(config, "/")

	node := Root
	if node != nil {
		for _, name := range names {
			if len(name) > 0 { // this check lets us handle extra /'s
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
func Create(url string, body io.Reader) int {
	node := new(Node)
	decoder := json.NewDecoder(body)
	err := decoder.Decode(node)

	if err != nil {
		return http.StatusBadRequest
	} else if url == "" { // creating root node so we can't find a parent
		if Root != nil {
			return http.StatusConflict
		} else {
			Root = node
			return http.StatusOK
		}
	} else {
		url, name := path.Split(url)
		parent, err := FindNode(url)

		if err != nil {
			return http.StatusNotFound
		} else {
			if parent.Children[name] != nil {
				return http.StatusConflict
			} else {
				parent.Children[name] = node
				return http.StatusOK
			}
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
func Read(url string) ([]byte, int) {
	node, err := FindNode(url)

	if err != nil {
		return nil, http.StatusNotFound
	} else {
		result, err := json.Marshal(node)
		if err != nil {
			return nil, http.StatusBadRequest
		} else {
			return result, http.StatusOK
		}
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
func Update(url string, body io.Reader) int {
	node := new(Node)
	decoder := json.NewDecoder(body)
	err := decoder.Decode(node)

	if err != nil {
		return http.StatusBadRequest
	} else if url == "" {  // updating root node so we can't find a parent
		if Root == nil {
			return http.StatusNotFound
		} else {
			Root = node
			return http.StatusOK
		}
	} else {
		url, name := path.Split(url)
		parent, err := FindNode(url)

		if err != nil {
			return http.StatusNotFound
		} else if parent.Children[name] == nil {
			return http.StatusNotFound
		} else {
			delete(parent.Children, name)
			parent.Children[name] = node
			return http.StatusOK
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
func Delete(url string) int {
	if url == "" {
		Root = nil
		return http.StatusOK
	} else {
		url, name := path.Split(url)
		parent, err := FindNode(url)

		if err != nil {
			return http.StatusNotFound
		} else {
			if parent.Children[name] == nil {
				return http.StatusNotFound
			} else {
				delete(parent.Children, name)
				return http.StatusOK
			}
		}
	}
}

func Handle(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[1:]
	body := r.Body

	switch r.Method {
	case "PUT":
		w.WriteHeader(Create(url, body))
	case "POST":
		w.WriteHeader(Update(url, body))
	case "DELETE":
		w.WriteHeader(Delete(url))
	default:
		result, code := Read(url)
		w.WriteHeader(code)
		if result != nil {
			w.Write(result)
		}
	}
}

func main() {
	http.HandleFunc("/", Handle)
	http.ListenAndServe(":8080", nil)
}
