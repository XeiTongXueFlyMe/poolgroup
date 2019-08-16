package poolgroup

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestGroup(t *testing.T) {
	counts, m := 0, sync.Mutex{}

	g := NewErrGroup()
	g.Go(func() error {
		time.Sleep(100 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return errors.New("err")
	})
	g.Go(func() error {
		time.Sleep(200 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return errors.New("err")
	})
	g.Go(func() error {
		time.Sleep(300 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return errors.New("err")
	})

	g.Wait()
	for _, err := range g.errs {
		assert.Error(t, err)
	}
	assert.EqualValues(t, 3, g.GetGoroutineNum())
	assert.EqualValues(t, 3, counts)
}

func TestGroupWithTimeout(t *testing.T) {
	counts, m := 0, sync.Mutex{}

	g := NewErrGroup()
	g.WithTimeout(context.Background(), 1000*150) //150ms
	g.Go(func() error {
		time.Sleep(100 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return errors.New("err")
	})
	g.Go(func() error {
		time.Sleep(100 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return errors.New("err")
	})
	g.Go(func() error {
		time.Sleep(200 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return errors.New("err")
	})

	g.Wait()
	for _, err := range g.errs {
		assert.Error(t, err)
	}
	assert.EqualValues(t, 3, g.GetGoroutineNum())
	assert.EqualValues(t, 2, counts)
}

func TestGroupWithContext(t *testing.T) {
	counts, m := 0, sync.Mutex{}

	g := NewErrGroup()
	g.WithContext(context.Background())
	g.Go(func() error {
		time.Sleep(10 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return nil
	})
	g.Go(func() error {
		time.Sleep(100 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return errors.New("err")
	})
	g.Go(func() error {
		time.Sleep(200 * time.Millisecond)
		m.Lock()
		defer m.Unlock()
		counts++
		return errors.New("err")
	})

	g.Wait()
	for _, err := range g.errs {
		assert.Error(t, err)
	}
	assert.EqualValues(t, 3, g.GetGoroutineNum())
	assert.EqualValues(t, 2, counts)
}
