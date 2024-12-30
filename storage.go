package ecs

//go:generate go run ./cmd/generate/main.go -depth 20

type Int interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 | ~int8 | ~int16 | ~int32 | ~int64
}

type Storage[ID Int] struct {
	Entitys    map[ID]Entity
	Components []Component
	Compounds  []*Compound[ID]
}

type Entity struct {
	Compound int
}

type Component struct {
	Name string
}

type Compound[ID Int] struct {
	Components []int
	Hashes     []int
	IDs        []ID
	Values     []any
	Removed    []ID
}

type Hashable interface{ Hash() int }

type ComponentHash struct{ ID, Hash int }

func New[ID Int]() *Storage[ID] {
	return &Storage[ID]{Entitys: map[ID]Entity{}}
}

func ComponentLookup[T any, ID Int](storage *Storage[ID]) (int, bool) {
	name := typeName[T]()
	for id, cmp := range storage.Components {
		if cmp.Name == name {
			return id, true
		}
	}
	return 0, false
}
