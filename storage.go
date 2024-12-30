package ecs

import (
	"log"
	"sync/atomic"
)

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
	name := typeName[T]()
	for id, cmp := range storage.Components {
		if cmp.Name == name {
			return id, true
		}
	}
	return 0, false
}

func (storage *Storage[ID]) compoundRebuild(id int) {
	compound := storage.Compounds[id]
	total := len(compound.EntitysRemoved)
	if len(compound.Entitys) == total {
		storage.Compounds = sliceRemove(storage.Compounds, id)
		return
	}
	idxRemove := make([]int, total)
	skipID, skipCount := compound.EntitysRemoved[0], 0
	for idx, id := range compound.Entitys {
		if id == skipID {
			idxRemove[skipCount] = idx
			skipCount++
			if skipCount == total {
				break
			}
			skipID = compound.EntitysRemoved[skipCount]
		}
	}
	log.Println("Cleanup", total, "/", len(compound.Entitys))
	panic("stop")
	for idx := range idxRemove {
		compound.Entitys = sliceRemove(compound.Entitys, idx)
		for _, component := range compound.Components {
			component.Data.remove(idx)
		}
	}
	compound.EntitysRemoved = nil
}
