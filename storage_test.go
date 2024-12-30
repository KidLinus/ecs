package ecs

import (
	"encoding/json"
	"math"
	"math/rand"
	"testing"
)

type Position struct{ X, Y int }
type Momentum struct{ HS, VS int }
type Walking struct{ Speed int }

func TestBasic(t *testing.T) {
	storage := New[uint32]()
	Set1(storage, 1, Position{100, 200})
	Set1(storage, 2, Position{200, 200})
	Set1(storage, 3, Momentum{10, -10})
	Query1[Position](storage).Each(func(id uint32, p *Position) {
		t.Log("RUN", id, p)
		p.X += 10
		p.Y += 3
	})
	storage.Remove(1)
	storage.Remove(2)
	storage.Remove(3)
	js, err := json.Marshal(storage)
	if err != nil {
		panic(err)
	}
	t.Log(string(js))
}

func BenchmarkPut1m(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		storage := New[uint32]()
		var id uint32
		var v Position
		b.StartTimer()
		for i := 0; i < 1_000_000; i++ {
			Set1(storage, id, v)
			id++
		}
	}
}

func Benchmark1mMove(b *testing.B) {
	storage := New[uint32]()
	for i := 0; i < 1_000_000; i++ {
		Set1(storage, uint32(i), Position{100, 200})
	}
	q := Query1[Position](storage)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		q.Each(func(id uint32, p *Position) {
			p.X++
			p.Y++
		})
	}
}

func BenchmarkBouncing(b *testing.B) {
	type Position struct{ X, Y int }
	type Physics struct {
		Hs, Vs            int
		Gravity, Friction int
	}
	storage := New[uint32]()
	for i := 0; i < 1_000; i++ {
		Set2(storage, uint32(i), Position{rand.Intn(500), rand.Intn(500)}, Physics{Gravity: 2})
	}
	collisionQ := Query1[Position](storage)
	collisionAt := func(id uint32, x, y int) (res bool) {
		collisionQ.Each(func(i uint32, p *Position) {
			if i == id {
				return
			}
			dx, dy := x-p.X, y-p.Y
			if dx > 30 || dx < -30 || dy > 30 || dy < -30 {
				res = true
				return
			}
			if dist(x, y, p.X, p.Y) <= 30 {
				res = true
				return
			}
		})
		return res
	}
	qAll := Query2[Position, Physics](storage)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		qAll.Each(func(id uint32, position *Position, physics *Physics) {
			physics.Vs += physics.Gravity
			if collisionAt(id, position.X, position.Y) {
				physics.Hs = -physics.Hs
			}
			if position.Y > 500 {
				position.Y = 500
				physics.Vs = min(0, -physics.Vs)
			}
			position.X += physics.Hs
			position.Y += physics.Vs
		})
	}
}

func dist(x, y, x2, y2 int) int {
	xx, yy := x2-x, y2-y
	return int(math.Sqrt(float64(xx*xx + yy*yy)))
}

func TestBasicQuery(t *testing.T) {
	storage := New[uint32]()
	Set1(storage, 1, Position{100, 200})
	Set1(storage, 2, Position{200, 200})
	Set1(storage, 3, Momentum{10, -10})
	Query1[Position](storage).Each(func(id uint32, p *Position) {
		t.Log("RUN", id, p)
		p.X += 10
		p.Y += 3
	})
	storage.Remove(1)
	storage.Remove(2)
	storage.Remove(3)
	js, err := json.Marshal(storage)
	if err != nil {
		panic(err)
	}
	t.Log(string(js))
}

func Benchmark1mMoveQuery(b *testing.B) {
	storage := New[uint32]()
	for i := 0; i < 1_000_000; i++ {
		Set1(storage, uint32(i), Position{100, 200})
	}
	q := Query1[Position](storage)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		q.Each(func(id uint32, p *Position) {
			p.X++
			p.Y++
		})
	}
}

func BenchmarkLoopEcsMixedQuery(b *testing.B) {
	type Value int
	type Thing int
	var store = New[uint32]()
	var id uint32
	for i := 0; i < 1_000_000; i++ {
		if id%100 == 0 {
			Set1(store, id, Thing(0))
		} else {
			Set1(store, id, Value(0))
		}
		id++
	}
	var idRemove uint32
	q := Query1[Thing](store)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// Loop Things
		q.Each(func(_ uint32, v *Thing) {
			*v++
		})
		// Remove
		for i := 0; i < 100; i++ {
			store.Remove(idRemove)
			idRemove++
		}
		// Add
		for i := 0; i < 100; i++ {
			if id%100 == 0 {
				Set1(store, id, Thing(0))
			} else {
				Set1(store, id, Value(0))
			}
			id++
		}
	}
}

func TestBasicRemoveQuery(t *testing.T) {
	storage := New[uint32]()
	Set1(storage, 1, Position{100, 200})
	Set1(storage, 2, Position{200, 200})
	Set1(storage, 3, Momentum{10, -10})
	Query1[Position](storage).Each(func(id uint32, p *Position) {
		p.X += 10
		p.Y += 3
	})
	storage.Remove(3)
	storage.Remove(1)
	storage.Remove(2)
	Query1[Position](storage).Each(func(id uint32, p *Position) {})
	Query1[Momentum](storage).Each(func(id uint32, p *Momentum) {})
	js, err := json.Marshal(storage)
	if err != nil {
		panic(err)
	}
	t.Log(string(js))
}
