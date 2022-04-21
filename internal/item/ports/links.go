package ports

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type LinkType int

const (
	LinkTypePrev LinkType = iota + 1
	LinkTypeNext
	LinkTypeFirst
	LinkTypeLast
)

type Link struct {
	req       *http.Request
	paginator *Paginator
}

func (l *Link) Encode(linkType LinkType) string {
	query := l.req.URL.Query()
	var rel string
	switch linkType {
	case LinkTypeNext:
		query.Set("page", strconv.Itoa(l.paginator.Next()))
		rel = "next"
	case LinkTypePrev:
		query.Set("page", strconv.Itoa(l.paginator.Prev()))
		rel = "prev"
	case LinkTypeFirst:
		query.Set("page", strconv.Itoa(l.paginator.First()))
		rel = "first"
	case LinkTypeLast:
		query.Set("page", strconv.Itoa(l.paginator.Last()))
		rel = "last"
	}
	return fmt.Sprintf("<%s?%s>;rel=%s", l.req.URL.Path, query.Encode(), rel)
}

func links(req *http.Request, paginator *Paginator) string {
	links := make([]string, 0)
	link := Link{
		paginator: paginator,
		req:       req,
	}
	if paginator.HasNext() {
		links = append(links, link.Encode(LinkTypeNext))
	}
	if paginator.HasPrev() {
		links = append(links, link.Encode(LinkTypePrev))
	}
	links = append(links, link.Encode(LinkTypeFirst))
	links = append(links, link.Encode(LinkTypeLast))

	return strings.Join(links, ",")
}
