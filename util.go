package ecs

import (
	"cmp"
	"reflect"
	"slices"
)

func (storage *Storage[ID]) componentEnsure(v any) int {
	name := reflect.TypeOf(v).String()
	for idx, v := range storage.Components {
		if v.Name == name {
			return idx
		}
	}
	storage.Components = append(storage.Components, Component{Name: name})
	return len(storage.Components) - 1
}

func (storage *Storage[ID]) compoundEnsure(components, hashes []int) int {
CompoundLoop:
	for idx, v := range storage.Compounds {
		if len(v.Components) != len(components) {
			continue
		}
	EqualityLoop:
		for ai, av := range components {
			for _, b := range v.Components {
				if av != b.ID {
					continue
				}
				if hashes[ai] != b.Hash {
					continue CompoundLoop
				}
				continue EqualityLoop
			}
			continue CompoundLoop
		}
		return idx
	}
	var cmps []CompoundComponent
	for idx, component := range components {
		cmps = append(cmps, CompoundComponent{ID: component, Hash: hashes[idx]})
	}
	storage.Compounds = append(storage.Compounds, &Compound[ID]{Components: cmps})
	return len(storage.Compounds) - 1
}

func (storage *Storage[ID]) getComponent(name string) (int, bool) {
	for id, cmp := range storage.Components {
		if cmp.Name == name {
			return id, true
		}
	}
	return 0, false
}

func typeName[T any]() string {
	var z [0]T
	return reflect.TypeOf(z).Elem().String()
}

func sliceRemove[V any](s []V, i int) []V {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func getOptional[T any](slice []T, idx int) *T {
	if slice == nil {
		return nil
	}
	return &slice[idx]
}

func sliceInsertOrdered[T cmp.Ordered](ts []T, t T) []T {
	i, _ := slices.BinarySearch(ts, t) // find slot
	// Make room for new value and add it
	ts = append(ts, *new(T))
	copy(ts[i+1:], ts[i:])
	ts[i] = t
	return ts
}

func componentHash(v any) int {
	if h, ok := v.(Hashable); ok {
		return h.Hash()
	}
	return 0
}

func sliceFind[V comparable](slice []V, value V) (int, bool) {
	for i, v := range slice {
		if v == value {
			return i, true
		}
	}
	return 0, false
}
