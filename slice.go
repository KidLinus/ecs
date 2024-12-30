package ecs

type Slice interface {
	remove(...int)
}

type slice[V any] struct {
	Data []V
}

func (s *slice[V]) remove(idxs ...int) {
	for _, idx := range idxs {
		s.Data = sliceRemove(s.Data, idx)
	}
}

func (s *slice[V]) append(vs ...V) {
	s.Data = append(s.Data, vs...)
}
