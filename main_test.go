package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

var test_child = &Node {
	Value: 12345,
	Children: map[string]*Node {
		"test1child2": &Node {
			Value: []string{"test","testing"},
		},
	},
}

var test_root = &Node {
	Value: "root val",
	Children: map[string]*Node {
		"child1": &Node {
			Value: "I'm a child",
			Children: map[string]*Node {
				"child2": test_child,
			},
		},
	},
}

func TestCreate(t *testing.T) {
	enc_root, _ := json.Marshal(test_root)
	code := Create("", bytes.NewReader(enc_root))
	if code != http.StatusOK {
		t.Fail()
	}
}

func TestRead(t *testing.T) {
	// Test reading root
	_, code := Read("")
	if code != http.StatusOK {
		t.Fail()
	}

	// Test reading a node
	_, code = Read("child1/child2")
	if code != http.StatusOK {
		t.Fail()
	}
}

func TestDelete(t *testing.T) {
	// Test deleting a node
	code := Delete("child1/child2")

	if code != http.StatusOK {
		t.Fail()
	}

	_, code = Read("child1/child2")
	if code != http.StatusNotFound {
		t.Fail()
	}
}

func TestUpdate(t *testing.T) {
	enc_child, _ := json.Marshal(test_child)
	code := Create("child1/child2", bytes.NewReader(enc_child))
	if code != http.StatusOK {
		t.Fail()
	}

	_, code = Read("child1/child2")
	if code != http.StatusOK {
		t.Fail()
	}
}
