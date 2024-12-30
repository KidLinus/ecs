package main

import (
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/KidLinus/ecs"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type A struct{}
type B struct{}
type C struct{}
type D struct{}

func main() {
	storage := ecs.New[uint32]()
	cpus := runtime.NumCPU()
	var ops uint64
	var idCreate uint32
	for i := 0; i < cpus; i++ {
		go func() {
			for {
				id := atomic.AddUint32(&idCreate, 1)
				ecs.Set4(storage, id, A{}, B{}, C{}, D{})
				atomic.AddUint64(&ops, 1)
				storage.Remove(id)
				atomic.AddUint64(&ops, 1)
			}
		}()
	}
	for i := 0; i < cpus; i++ {
		q := ecs.Query1[A](storage)
		go func() {
			for {
				q.Each(func(u uint32, a *A) {})
				atomic.AddUint64(&ops, 1)
			}
		}()
	}
	for i := 0; i < cpus; i++ {
		q := ecs.Query1[B](storage)
		go func() {
			for {
				q.Each(func(u uint32, a *B) {})
				atomic.AddUint64(&ops, 1)
			}
		}()
	}
	for i := 0; i < cpus; i++ {
		q := ecs.Query1[C](storage)
		go func() {
			for {
				q.Each(func(u uint32, a *C) {})
				atomic.AddUint64(&ops, 1)
			}
		}()
	}
	for i := 0; i < cpus; i++ {
		q := ecs.Query1[D](storage)
		go func() {
			for {
				q.Each(func(u uint32, a *D) {})
				atomic.AddUint64(&ops, 1)
			}
		}()
	}
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	ticker := time.NewTicker(time.Second)
	p := message.NewPrinter(language.English)
	for {
		select {
		case <-ticker.C:
			p.Printf("Calls %d\n", atomic.LoadUint64(&ops))
		case <-ch:
			return
		}
	}
}
