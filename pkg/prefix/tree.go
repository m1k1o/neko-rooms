package prefix

import (
	"strings"
)

type Tree[T any] interface {
	Insert(prefix string, value T)
	Find(prefix string) (value T, ok bool)
	Remove(prefix string)
	Match(path string) (value T, prefix string, ok bool)
}

type tree[T any] struct {
	Value    T
	IsLeaf   bool
	Children map[string]*tree[T]
}

func NewTree[T any]() *tree[T] {
	return &tree[T]{}
}

func (p *tree[T]) Insert(prefix string, value T) {
	arr := strings.Split(prefix, "/")
	l := len(arr)

	for i, a := range arr {
		if a == "" {
			continue
		}

		if p.Children == nil {
			p.Children = map[string]*tree[T]{}
		}

		dat, ok := p.Children[a]
		if !ok {
			dat = &tree[T]{}
		}

		if i == l-1 {
			dat.Value = value
			dat.IsLeaf = true
			dat.Children = nil
		}

		p.Children[a] = dat
		p = dat
	}
}

func (p *tree[T]) Find(prefix string) (value T, ok bool) {
	arr := strings.Split(prefix, "/")
	l := len(arr)

	for i, a := range arr {
		if a == "" {
			continue
		}

		if p.Children == nil {
			p.Children = map[string]*tree[T]{}
		}

		dat, found := p.Children[a]
		if !found {
			dat = &tree[T]{}
		}

		if i == l-1 {
			value = dat.Value
			ok = dat.IsLeaf
			return
		}

		p.Children[a] = dat
		p = dat
	}

	return
}

func (p *tree[T]) Remove(prefix string) {
	arr := strings.Split(prefix, "/")
	l := len(arr)

	ptrs := []*tree[T]{p}
	for i, a := range arr {
		if a == "" {
			continue
		}

		if i == l-1 {
			delete(p.Children, a)
		}

		dat, found := p.Children[a]
		if !found {
			break
		}
		p = dat
		ptrs = append(ptrs, p)
	}

	// remove all empty references
	rm := false
	for i := len(ptrs) - 1; i >= 0; i-- {
		if len(ptrs[i].Children) == 0 {
			rm = true
			continue
		}
		if rm {
			ptrs[i].Children = nil
		}
	}
}

func (p *tree[T]) Match(path string) (value T, prefix string, ok bool) {
	arr := strings.Split(path, "/")
	prefixArr := []string{}

	for _, a := range arr {
		if a == "" {
			continue
		}

		pointer, found := p.Children[a]
		if !found {
			return
		}

		p = pointer
		prefixArr = append(prefixArr, a)

		if p.IsLeaf {
			break
		}
	}

	value = p.Value
	prefix = "/" + strings.Join(prefixArr, "/")
	ok = p.IsLeaf
	return
}
