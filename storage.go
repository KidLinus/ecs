package ecs

import (
	"sync"
	"sync/atomic"
)

//go:generate go run ./cmd/generate/main.go -depth 20

type Int interface {
	int | uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~int8 | ~int16 | ~int32 | ~int64
}

type Storage[ID Int] struct {
	lock       sync.RWMutex
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
	Components     []CompoundComponent
	Entitys        []ID
	EntitysRemoved []ID
	cleanupTime    atomic.Bool
}

type CompoundComponent struct {
	ID   int
	Hash int
	Data Slice
}

type Hashable interface{ Hash() int }

type ComponentHash struct{ ID, Hash int }

func New[ID Int]() *Storage[ID] {
	return &Storage[ID]{Entitys: map[ID]Entity{}}
}

func ComponentLookup[T any, ID Int](storage *Storage[ID]) (int, bool) {
	storage.lock.RLock()
	defer storage.lock.RUnlock()
	name := typeName[T]()
	for id, cmp := range storage.Components {
		if cmp.Name == name {
			return id, true
		}
	}
	return 0, false
}
