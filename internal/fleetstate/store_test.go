package fleetstate

import (
	"context"
	"testing"
	"time"

	"github.com/narqo/ree-fleet-sim/internal/vehicle"
)

func TestMemStore_Write_DropOldRecords(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := NewMemStore()

	now := time.Now().UTC()

	err := store.Write(ctx, "THE1VIN", now, 1, 1)
	if err != nil {
		t.Fatal(err)
	}

	err = store.Write(ctx, "THE1VIN", now.Add(-1*time.Second), 1, 1)
	if err == nil {
		t.Fatal("write old record: want err got nil")
	}

	// different vin is fine, it's its first data point
	err = store.Write(ctx, "ANOTHER1VIN", now.Add(-1*time.Second), 1, 1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMemStore_Reader(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := NewMemStore()

	vin := vehicle.VIN("THE1VIN")
	now := time.Now().UTC()

	for i := 1; i <= 3; i++ {
		ts := now.Add(time.Duration(i) * time.Second)
		if err := store.Write(ctx, vin, ts, float64(i)*10, float64(i)*10); err != nil {
			t.Fatal(err)
		}
	}

	// must read starting the last record
	reader, err := store.Reader(ctx, vin)
	if err != nil {
		t.Fatal(err)
	}
	testReaderRead(t, reader, now.Add(3*time.Second), 3*10, 3*10)

	now = now.Add(10 * time.Second)
	for i := 1; i <= 3; i++ {
		ts := now.Add(time.Duration(i) * time.Second)
		if err := store.Write(ctx, vin, ts, float64(i)*20, float64(i)*20); err != nil {
			t.Fatal(err)
		}
	}
	// must read all the new records, advancing its offset
	for i := 1; i <= 3; i++ {
		ts := now.Add(time.Duration(i) * time.Second)
		testReaderRead(t, reader, ts, float64(i)*20, float64(i)*20)
	}

	// every new instance must start reading from latest
	reader, err = store.Reader(ctx, vin)
	if err != nil {
		t.Fatal(err)
	}
	testReaderRead(t, reader, now.Add(3*time.Second), 3*20, 3*20)
}

func TestMemStore_Reader_ReadClosed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := NewMemStore()
	if err := store.Write(ctx, "THE1VIN", time.Now().UTC(), 1, 1); err != nil {
		t.Fatal(err)
	}

	reader, err := store.Reader(ctx, "THE1VIN")
	if err != nil {
		t.Fatal(err)
	}

	_, _, _, err = reader.Read()
	if err != nil {
		t.Fatal(err)
	}

	time.AfterFunc(300*time.Millisecond, cancel)

	_, _, _, err = reader.Read()
	if err != ErrReaderClosed {
		t.Fatalf("read closed: want err got %v", err)
	}
}

func TestMemStore_Reader_UnknownVIN(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := NewMemStore()
	if err := store.Write(ctx, "THE1VIN", time.Now().UTC(), 1, 1); err != nil {
		t.Fatal(err)
	}

	_, err := store.Reader(ctx, "ANOTHER1VIN")
	if err == nil {
		t.Fatal("read unknown vin: want err got nil")
	}
}

func testReaderRead(t *testing.T, reader Reader, wantTs time.Time, wantLat, wantLon float64) {
	t.Helper()

	ts, lat, lon, err := reader.Read()
	if err != nil {
		t.Fatal(err)
	}
	if !ts.Equal(wantTs) {
		t.Fatalf("ts: want %v got %v", wantTs, ts)
	}
	if lat != wantLat {
		t.Fatalf("lat: want %v got %v", wantLat, lat)
	}
	if lon != wantLon {
		t.Fatalf("lon: want %v got %v", wantLon, lon)
	}
}
