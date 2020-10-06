package fleetstate

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/narqo/ree-fleet-sim/internal/vehicle"
)

var (
	ErrReaderClosed = errors.New("reader is closed")
)

type Store interface {
	Write(ctx context.Context, vin vehicle.VIN, ts time.Time, lat, lon float64) error
	Reader(ctx context.Context, vin vehicle.VIN) (Reader, error)
}

type Reader interface {
	Read() (ts time.Time, lat, lon float64, err error)
}

// MemStore is an in-memory implementation of Store.
type MemStore struct {
	mu   sync.Mutex
	data map[vehicle.VIN]*Data
}

var _ Store = (*MemStore)(nil)

type Record struct {
	Ts  time.Time
	Lon float64
	Lat float64
}

type Data struct {
	// protects recs from concurrent access
	mu sync.Mutex
	// notifies readers about new records in recs
	cond sync.Cond
	recs []Record
}

func NewMemStore() *MemStore {
	return &MemStore{
		data: make(map[vehicle.VIN]*Data),
	}
}

func (store *MemStore) Write(ctx context.Context, vin vehicle.VIN, ts time.Time, lat, lon float64) error {
	store.mu.Lock()

	data := store.data[vin]
	if data == nil {
		data = &Data{}
		data.cond.L = &data.mu
		store.data[vin] = data
	}

	data.mu.Lock()
	defer data.mu.Unlock()

	// NOTE: unlock store-level lock only after we acquired the lock on vin-level data above
	store.mu.Unlock()

	if len(data.recs) > 0 {
		// don't bother back-filling a missing data points, to make things simpler
		lastTs := data.recs[len(data.recs)-1].Ts
		if lastTs.After(ts) {
			return fmt.Errorf("old record for vin %s: ts %d, lastTs %d, lat, %f, lon %f", vin, ts.UnixNano(), lastTs.UnixNano(), lon, lat)
		}
	}
	data.recs = append(data.recs, Record{ts, lon, lat})
	data.cond.Broadcast()

	return nil
}

func (store *MemStore) Reader(ctx context.Context, vin vehicle.VIN) (Reader, error) {
	store.mu.Lock()
	data := store.data[vin]
	store.mu.Unlock()
	if data == nil {
		return nil, fmt.Errorf("unknown vin %s", vin)
	}

	data.mu.Lock()
	offset := len(data.recs) - 1
	data.mu.Unlock()

	r := &reader{
		data:   data,
		offset: offset,
		closed: make(chan struct{}),
	}

	go func() {
		// Note, in this (demo) implementation caller guaranties the context is closed eventually,
		// i.e. no goroutine leak
		<-ctx.Done()

		close(r.closed)
		data.cond.Broadcast()
	}()

	return r, nil
}

type reader struct {
	data   *Data
	offset int
	closed chan struct{}
}

func (r *reader) Read() (ts time.Time, lat, lon float64, err error) {
	r.data.mu.Lock()
	defer r.data.mu.Unlock()
	for len(r.data.recs) <= r.offset {
		r.data.cond.Wait()

		select {
		case <-r.closed:
			return time.Time{}, 0, 0, ErrReaderClosed
		default:
		}
	}

	rec := r.data.recs[r.offset]
	r.offset++

	return rec.Ts, rec.Lat, rec.Lon, nil
}
