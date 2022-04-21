package ports

import "math"

type Paginator struct {
	page    int
	perPage int
	count   int
}

func NewPaginator(page, perPage, count int) *Paginator {
	return &Paginator{page, perPage, count}
}

func (p *Paginator) Page() int {
	return p.page
}

func (p *Paginator) PerPage() int {
	return p.perPage
}

func (p *Paginator) Pages() int {
	return int(math.Ceil(float64(p.count) / float64(p.perPage)))
}

func (p *Paginator) Last() int {
	return p.Pages()
}

func (p *Paginator) First() int {
	return 1
}

func (p *Paginator) HasPrev() bool {
	return (p.page - 1) > 0
}

func (p *Paginator) Prev() int {
	return p.page - 1
}

func (p *Paginator) HasNext() bool {
	return (p.page + 1) <= p.Last()
}

func (p *Paginator) Next() int {
	return p.page + 1
}
