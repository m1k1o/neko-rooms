package prefix

import (
	"fmt"
	"testing"
)

func TestMatch(t *testing.T) {
	// create a new prefix tree
	tree := NewTree[string]()

	// add some values to the tree
	tree.Insert("/users/1", "user1")
	tree.Insert("/users/2", "user2")

	// test matching a path that exists in the tree
	value, prefix, ok := tree.Match("/users/1")
	if !ok || value != "user1" || prefix != "/users/1" {
		t.Errorf("Match failed for /users/1")
	}

	// test matching a path that exists in the tree but has a trailing slash
	value, prefix, ok = tree.Match("/users/2/")
	if !ok || value != "user2" || prefix != "/users/2" {
		t.Errorf("Match failed for /users/2/")
	}

	// test matching a path that does not exist in the tree
	_, _, ok = tree.Match("/users/4")
	if ok {
		t.Errorf("Match should have failed for /users/4")
	}

	// test matching a path that partially exists in the tree
	value, prefix, ok = tree.Match("/users/1/posts")
	if !ok || value != "user1" || prefix != "/users/1" {
		t.Errorf("Match failed for /users/1/posts: %v %v %v", value, prefix, ok)
	}
}

func TestFind(t *testing.T) {
	// create a new prefix tree
	tree := NewTree[string]()

	// add some values to the tree
	tree.Insert("/users/1", "user1")
	tree.Insert("/users/2", "user2")

	// test finding a path that exists in the tree
	value, ok := tree.Find("/users/1")
	if !ok || value != "user1" {
		t.Errorf("Find failed for /users/1")
	}

	// test finding a path that exists in the tree but has a trailing slash
	_, ok = tree.Find("/users/2/")
	if ok {
		t.Errorf("Find should have failed for /users/2/")
	}

	// test finding a path that does not exist in the tree
	_, ok = tree.Find("/users/4")
	if ok {
		t.Errorf("Find should have failed for /users/4")
	}

	// test finding a path that partially exists in the tree
	_, ok = tree.Find("/users/1/posts")
	if ok {
		t.Errorf("Find should have failed for /users/1/posts")
	}
}

func TestRemove(t *testing.T) {
	// create a new prefix tree
	tree := &tree[string]{}

	// add some values to the tree
	tree.Insert("/users/1", "user1")
	tree.Insert("/users/2", "user2")

	// test matching a path that exists in the tree
	value, prefix, ok := tree.Match("/users/1")
	if !ok || value != "user1" || prefix != "/users/1" {
		t.Errorf("Match failed for /users/1")
	}

	// remove a path that exists in the tree
	tree.Remove("/users/1")

	// test matching a path that exists in the tree
	_, _, ok = tree.Match("/users/1")
	if ok {
		t.Errorf("Match should have failed for /users/1")
	}

	// test finding a path that exists in the tree
	_, ok = tree.Find("/users/1")
	if ok {
		t.Errorf("Find should have failed for /users/1")
	}

	// remove another path that exists in the tree
	tree.Remove("/users/2")

	// test matching a path that exists in the tree
	_, _, ok = tree.Match("/users/2")
	if ok {
		t.Errorf("Match should have failed for /users/2")
	}

	// check that the tree is empty
	if tree.Children != nil {
		t.Errorf("Tree should be empty, but contains %d children", len(tree.Children))
		for k, v := range tree.Children {
			t.Errorf("%s: %v", k, v)
		}
	}
}

// Benchmark the match function
func Benchmark(b *testing.B) {
	// create a new prefix tree
	tree := NewTree[string]()

	// insert 1 - 10 levels
	level := "/users"
	for i := 1; i <= 10; i++ {
		level += fmt.Sprintf("/%d", i)
		tree.Insert(level, fmt.Sprintf("level_%d", i))
	}

	// run the benchmark
	level = "/users"
	for i := 0; i <= 10; i++ {
		level += fmt.Sprintf("/%d", i)
		b.Run(fmt.Sprintf("level_%d", i), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tree.Match(level)
			}
		})
	}
}
