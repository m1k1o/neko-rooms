package proxy

import (
	"strings"
)

type prefixHandler[T any] struct {
	Value    T
	IsLeaf   bool
	Children map[string]*prefixHandler[T]
}

func NewPrefixHandler[T any]() *prefixHandler[T] {
	return &prefixHandler[T]{}
}

func (p *prefixHandler[T]) Set(prefix string, value T) {
	arr := strings.Split(prefix, "/")
	len := len(arr)

	for i, a := range arr {
		if a == "" {
			continue
		}

		if p.Children == nil {
			p.Children = map[string]*prefixHandler[T]{}
		}

		dat, ok := p.Children[a]
		if !ok {
			dat = &prefixHandler[T]{}
		}

		if i == len-1 {
			dat.Value = value
			dat.IsLeaf = true
			dat.Children = nil
		}

		p.Children[a] = dat
		p = dat
	}
}

func (p *prefixHandler[T]) Get(prefix string) (value T, ok bool) {
	arr := strings.Split(prefix, "/")
	len := len(arr)

	for i, a := range arr {
		if a == "" {
			continue
		}

		if p.Children == nil {
			p.Children = map[string]*prefixHandler[T]{}
		}

		dat, found := p.Children[a]
		if !found {
			dat = &prefixHandler[T]{}
		}

		if i == len-1 {
			value = dat.Value
			ok = dat.IsLeaf
			return
		}

		p.Children[a] = dat
		p = dat
	}

	return
}

func (p *prefixHandler[T]) Remove(prefix string) {
	arr := strings.Split(prefix, "/")
	len := len(arr)

	for i, a := range arr {
		if a == "" {
			continue
		}

		if i == len-1 {
			delete(p.Children, a)
		}

		dat, found := p.Children[a]
		if !found {
			break
		}
		p = dat
	}

	// TODO: Remove all empty references.
}

func (p *prefixHandler[T]) Match(path string) (value T, prefix string, ok bool) {
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
