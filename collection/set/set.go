package set

type SetString struct {
	inner map[string]struct{}
}

func NewSetString() *SetString {
	return &SetString{inner: make(map[string]struct{})}
}

func (s *SetString) Add(elem ...string) {
	for _, ele := range elem {
		s.inner[ele] = struct{}{}
	}
}

func (s *SetString) Remove(elem ...string) {
	for _, ele := range elem {
		delete(s.inner, ele)
	}
}

func (s *SetString) Contains(elem string) bool {
	_, has := s.inner[elem]
	return has
}
