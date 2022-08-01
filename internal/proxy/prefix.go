package proxy

import (
	"net/http/httputil"
	"strings"
)

type prefixHandler struct {
	Value    *httputil.ReverseProxy
	Children map[string]*prefixHandler
}

func (p *prefixHandler) Set(prefix string, value *httputil.ReverseProxy) {
	arr := strings.Split(prefix, "/")
	len := len(arr)

	for i, a := range arr {
		if a == "" {
			continue
		}

		if p.Children == nil {
			p.Children = map[string]*prefixHandler{}
		}

		dat, ok := p.Children[a]
		if !ok {
			dat = &prefixHandler{}
		}

		if i == len-1 {
			dat.Value = value
			dat.Children = nil
		}

		p.Children[a] = dat
		p = dat
	}
}

func (p *prefixHandler) Remove(prefix string) {
	arr := strings.Split(prefix, "/")
	len := len(arr)

	for i, a := range arr {
		if a == "" {
			continue
		}

		if i == len-1 {
			delete(p.Children, a)
		}

		dat, ok := p.Children[a]
		if !ok {
			break
		}
		p = dat
	}

	// TODO: Remove all empty references.
}

func (p *prefixHandler) Match(path string) (*httputil.ReverseProxy, string, bool) {
	arr := strings.Split(path, "/")
	prefixArr := []string{}

	for _, a := range arr {
		if a == "" {
			continue
		}

		pointer, ok := p.Children[a]
		if !ok {
			return nil, "", false
		}

		p = pointer
		prefixArr = append(prefixArr, a)

		// if leaf node
		if p.Children == nil {
			break
		}
	}

	// if not leaf node
	if p.Children != nil {
		return nil, "", false
	}

	prefix := "/" + strings.Join(prefixArr, "/")
	return p.Value, prefix, true
}
