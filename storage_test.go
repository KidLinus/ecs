package ecs

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"math/rand"
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

// Benchmark1mMove-8   	     349	   3526957 ns/op	       0 B/op	       0 allocs/op

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

// BenchmarkBouncing-8   	     206	   5323178 ns/op	       0 B/op	       0 allocs/op

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

// BenchmarkLoopEcsMixedQuery-8   	   13772	     93452 ns/op	    9303 B/op	       0 allocs/op
// BenchmarkLoopEcsMixedQuery-8   	   16664	     73580 ns/op	    9231 B/op	       3 allocs/op
// BenchmarkLoopEcsMixedQuery-8        17169         71048 ns/op        10243 B/op         3 allocs/op
// BenchmarkLoopEcsMixedQuery-8   	   17736	     69291 ns/op	    9921 B/op	       3 allocs/op
// BenchmarkLoopEcsMixedQuery-8   	   18907	     64818 ns/op	    9319 B/op	       3 allocs/op

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
	Set1(storage, 3, Position{100, 200})
	Set1(storage, 2, Position{200, 200})
	Set1(storage, 1, Momentum{10, -10})
	Query1[Position](storage).Each(func(id uint32, p *Position) {
		p.X += 10
		p.Y += 3
	})
	storage.Remove(3)
	storage.Remove(2)
	storage.Remove(1)
	Query1[Position](storage).Each(func(id uint32, p *Position) {})
	Query1[Momentum](storage).Each(func(id uint32, p *Momentum) {})
	js, err := json.Marshal(storage)
	if err != nil {
		panic(err)
	}
	t.Log(string(js))
}

// After mutex
// BenchmarkPut1m-8               	       5	 237636824 ns/op	193814433 B/op	 1038556 allocs/op
// Benchmark1mMove-8              	     333	   3690498 ns/op	       0 B/op	       0 allocs/op
// BenchmarkBouncing-8            	     220	   5439721 ns/op	       0 B/op	       0 allocs/op
// Benchmark1mMoveQuery-8         	     319	   3718086 ns/op	       0 B/op	       0 allocs/op
// BenchmarkLoopEcsMixedQuery-8   	   17371	     72591 ns/op	   10126 B/op	       3 allocs/op

// After PROPER mutex
// BenchmarkPut1m-8               	       4	 256931292 ns/op	193812948 B/op	 1038544 allocs/op
// Benchmark1mMove-8              	     265	   4460548 ns/op	       0 B/op	       0 allocs/op
// BenchmarkBouncing-8            	     190	   6312361 ns/op	       0 B/op	       0 allocs/op
// Benchmark1mMoveQuery-8         	     264	   4390330 ns/op	       0 B/op	       0 allocs/op
// BenchmarkLoopEcsMixedQuery-8   	   14545	     76777 ns/op	    8783 B/op	       4 allocs/op

// After fixing removal using map
// BenchmarkPut1m-8                	       4	 256507888 ns/op	193811240 B/op	 1038520 allocs/op
// Benchmark1mMove-8               	     279	   4177591 ns/op	       0 B/op	       0 allocs/op
// BenchmarkBouncing-8             	     194	   6392055 ns/op	       0 B/op	       0 allocs/op
// Benchmark1mMoveQuery-8          	     250	   4471713 ns/op	       0 B/op	       0 allocs/op
// BenchmarkLoopEcsMixedQuery-8    	    9624	    126884 ns/op	    8648 B/op	       4 allocs/op

// BenchmarkPut1m-8                	       4	 314491522 ns/op	193808216 B/op	 1038502 allocs/op
// Benchmark1mMove-8               	     242	   4932917 ns/op	       0 B/op	       0 allocs/op
// BenchmarkBouncing-8             	     182	   6495416 ns/op	       0 B/op	       0 allocs/op
// Benchmark1mMoveQuery-8          	     237	   4777694 ns/op	       0 B/op	       0 allocs/op
// BenchmarkLoopEcsMixedQuery-8    	   16764	     72545 ns/op	    9974 B/op	       4 allocs/op

// Fixin remove logic
// BenchmarkPut1m-8                	       4	 334963309 ns/op	193804428 B/op	 1038468 allocs/op
// Benchmark1mMove-8               	     226	   4861445 ns/op	       0 B/op	       0 allocs/op
// BenchmarkBouncing-8             	     183	   6497502 ns/op	       0 B/op	       0 allocs/op
// Benchmark1mMoveQuery-8          	     261	   4653538 ns/op	       0 B/op	       0 allocs/op
// BenchmarkLoopEcsMixedQuery-8    	   15938	     74471 ns/op	    8032 B/op	       4 allocs/op

// Also removing removal indexes
// BenchmarkPut1m-8                	       4	 275323257 ns/op	193824296 B/op	 1038647 allocs/op
// Benchmark1mMove-8               	     271	   4509466 ns/op	       0 B/op	       0 allocs/op
// BenchmarkBouncing-8             	     186	   6392391 ns/op	       0 B/op	       0 allocs/op
// Benchmark1mMoveQuery-8          	     260	   4655701 ns/op	       0 B/op	       0 allocs/op
// BenchmarkLoopEcsMixedQuery-8    	   15963	     77232 ns/op	    8018 B/op	       4 allocs/op

// Updated generator to match
// BenchmarkPut1m-8                	       4	 260577882 ns/op	193804004 B/op	 1038464 allocs/op
// Benchmark1mMove-8               	     270	   4618143 ns/op	       0 B/op	       0 allocs/op
// BenchmarkBouncing-8             	     190	   6335551 ns/op	       0 B/op	       0 allocs/op
// Benchmark1mMoveQuery-8          	     264	   4441990 ns/op	       0 B/op	       0 allocs/op
// BenchmarkLoopEcsMixedQuery-8    	   15588	     78227 ns/op	    8206 B/op	       4 allocs/op

func Benchmark1000x10Array(b *testing.B) {
	var items []int
	for i := 0; i < 1000; i++ {
		items = append(items, i)
	}
	var remove []int
	for i := 0; i < 10; i++ {
		remove = append(remove, rand.Intn(len(items)))
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, i := range items {
			for _, r := range remove {
				if i == r {
					break
				}
			}
		}
	}
}

func Benchmark1000x10MapData(b *testing.B) {
	var items []int
	for i := 0; i < 1000; i++ {
		items = append(items, i)
	}
	var remove []int
	for i := 0; i < 10; i++ {
		remove = append(remove, rand.Intn(len(items)))
	}
	var target int
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		m := map[int]int{}
		for idx, i := range items {
			m[i] = idx
		}
		for _, r := range remove {
			target = m[r]
		}
	}
	fmt.Sprint(target)
}

func Benchmark1000x10MapTarget(b *testing.B) {
	var items []int
	for i := 0; i < 1000; i++ {
		items = append(items, i)
	}
	var remove []int
	for i := 0; i < 10; i++ {
		remove = append(remove, rand.Intn(len(items)))
	}
	var target int
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		m := map[int]struct{}{}
		for _, i := range remove {
			m[i] = struct{}{}
		}
		for _, r := range items {
			if _, ok := m[r]; ok {
				target++
			}
		}
	}
	fmt.Sprint(target)
}

func Benchmark10000x10Array(b *testing.B) {
	var items []int
	for i := 0; i < 10000; i++ {
		items = append(items, i)
	}
	var remove []int
	for i := 0; i < 10; i++ {
		remove = append(remove, rand.Intn(len(items)))
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, i := range items {
			for _, r := range remove {
				if i == r {
					break
				}
			}
		}
	}
}

func Benchmark10000x10MapData(b *testing.B) {
	var items []int
	for i := 0; i < 10000; i++ {
		items = append(items, i)
	}
	var remove []int
	for i := 0; i < 10; i++ {
		remove = append(remove, rand.Intn(len(items)))
	}
	var target int
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		m := map[int]int{}
		for idx, i := range items {
			m[i] = idx
		}
		for _, r := range remove {
			target = m[r]
		}
	}
	fmt.Sprint(target)
}

func Benchmark10000x10MapTarget(b *testing.B) {
	var items []int
	for i := 0; i < 1000; i++ {
		items = append(items, i)
	}
	var remove []int
	for i := 0; i < 10; i++ {
		remove = append(remove, rand.Intn(len(items)))
	}
	var target int
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		m := map[int]struct{}{}
		for _, i := range remove {
			m[i] = struct{}{}
		}
		for _, r := range items {
			if _, ok := m[r]; ok {
				target++
			}
		}
	}
	fmt.Sprint(target)
}

func Benchmark10000x100Array(b *testing.B) {
	var items []int
	for i := 0; i < 10000; i++ {
		items = append(items, i)
	}
	var remove []int
	for i := 0; i < 100; i++ {
		remove = append(remove, rand.Intn(len(items)))
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, i := range items {
			for _, r := range remove {
				if i == r {
					break
				}
			}
		}
	}
}

func Benchmark10000x100MapData(b *testing.B) {
	var items []int
	for i := 0; i < 10000; i++ {
		items = append(items, i)
	}
	var remove []int
	for i := 0; i < 100; i++ {
		remove = append(remove, rand.Intn(len(items)))
	}
	var target int
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		m := map[int]int{}
		for idx, i := range items {
			m[i] = idx
		}
		for _, r := range remove {
			target = m[r]
		}
	}
	fmt.Sprint(target)
}

func Benchmark10000x100MapTarget(b *testing.B) {
	var items []int
	for i := 0; i < 1000; i++ {
		items = append(items, i)
	}
	var remove []int
	for i := 0; i < 100; i++ {
		remove = append(remove, rand.Intn(len(items)))
	}
	var target int
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		m := map[int]struct{}{}
		for _, i := range remove {
			m[i] = struct{}{}
		}
		for _, r := range items {
			if _, ok := m[r]; ok {
				target++
			}
		}
	}
	fmt.Sprint(target)
}

func Benchmark1x1MapTarget(b *testing.B) {
	var items []int
	for i := 0; i < 1; i++ {
		items = append(items, i)
	}
	var remove []int
	for i := 0; i < 1; i++ {
		remove = append(remove, rand.Intn(len(items)))
	}
	var target int
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		m := map[int]struct{}{}
		for _, i := range remove {
			m[i] = struct{}{}
		}
		for _, r := range items {
			if _, ok := m[r]; ok {
				target++
			}
		}
	}
	fmt.Sprint(target)
}
