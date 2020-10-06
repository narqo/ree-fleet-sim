package fleetstate

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestVehicleHandler_HandleUpdatePosition(t *testing.T) {
	store := NewMemStore()
	handler := NewVehicleHandler(store)

	v := url.Values{
		"lat": []string{"52.520008"},
		"lon": []string{"13.404954"},
	}
	r := httptest.NewRequest(http.MethodPost, "/the1vin", strings.NewReader(v.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	if err := handler.HandleUpdatePosition(w, r); err != nil {
		t.Fatal(err)
	}

	if want := http.StatusCreated; want != w.Code {
		t.Fatalf("HandleUpdatePosition: unexpected response status: want %v got %v", want, w.Code)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reader, err := store.Reader(ctx, "THE1VIN")
	if err != nil {
		t.Fatal(err)
	}

	ts, lat, lon, err := reader.Read()
	if err != nil {
		t.Fatal(err)
	}
	if ts.IsZero() {
		t.Errorf("unexpected ts %v", ts)
	}
	if want := 13.404954; want != lon {
		t.Errorf("lon: want %v got %v", want, lon)
	}
	if want := 52.520008; want != lat {
		t.Errorf("lat: want %v got %v", want, lat)
	}
}

func TestVehicleHandler_HandleUpdatePosition_BadRequest(t *testing.T) {
	store := NewMemStore()
	handler := NewVehicleHandler(store)

	t.Run("bad vin", func(t *testing.T) {
		v := url.Values{
			"lat": []string{"52.520008"},
			"lon": []string{"13.404954"},
		}
		r := httptest.NewRequest(http.MethodPost, "/the1vin.jpg", strings.NewReader(v.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		if err := handler.HandleUpdatePosition(w, r); err == nil {
			t.Fatal("HandleUpdatePosition: want error, got nil")
		}
	})

	t.Run("missing lat", func(t *testing.T) {
		v := url.Values{
			"lon": []string{"13.404954"},
		}
		r := httptest.NewRequest(http.MethodPost, "/the1vin", strings.NewReader(v.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		if err := handler.HandleUpdatePosition(w, r); err == nil {
			t.Fatal("HandleUpdatePosition: want error, got nil")
		}
	})

	t.Run("bad lat", func(t *testing.T) {
		v := url.Values{
			"lat": []string{"abc"},
			"lon": []string{"13.404954"},
		}
		r := httptest.NewRequest(http.MethodPost, "/the1vin", strings.NewReader(v.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		if err := handler.HandleUpdatePosition(w, r); err == nil {
			t.Fatal("HandleUpdatePosition: want error, got nil")
		}
	})
}

func TestVehicleHandler_HandleStreamPosition(t *testing.T) {
	store := NewMemStore()
	handler := NewVehicleHandler(store)

	now := time.Now().UTC()

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	// (Berlin, Cathedral)
	if err := store.Write(ctx, "THE1VIN", now, 52.518898, 13.401797); err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodGet, "/the1vin/stream", nil)
	r = r.WithContext(ctx)

	w := httptest.NewRecorder()
	respReader := bufio.NewReader(w.Body)

	var wg sync.WaitGroup

	// start consuming the stream in the background
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := handler.HandleStreamPosition(w, r); err != nil {
			t.Fatal(err)
		}
	}()

	// give handler some time to start processing
	time.Sleep(time.Second)

	// same vin, same coords (Berlin, Cathedral)
	if err := store.Write(ctx, "THE1VIN", now.Add(time.Second), 52.518898, 13.401797); err != nil {
		t.Fatal(err)
	}
	// same vin, different coords (Berlin, Fernsehturm)
	if err := store.Write(ctx, "THE1VIN", now.Add(2*time.Second), 52.520645, 13.409779); err != nil {
		t.Fatal(err)
	}
	// different vin, same coords (Berlin, Fernsehturm)
	if err := store.Write(ctx, "ANOTHER1VIN", now.Add(3*time.Second), 52.520645, 13.409779); err != nil {
		t.Fatal(err)
	}

	// give handler extra time to progress the stream
	time.Sleep(time.Second)

	// closes the client connection
	cancelCtx()
	wg.Wait()

	for i, want := range []string{
		`{"lat":52.518898,"lon":13.401797,"speed":0}`,
		`{"lat":52.520645,"lon":13.409779,"speed":2066.191265042517}`,
	} {
		got, _ := respReader.ReadString('\n')
		if want != strings.TrimSpace(got) {
			t.Errorf("HandleStreamPosition: line %d want %s got %s", i, want, got)
		}
	}
	if line, err := respReader.ReadString('\n'); err != io.EOF {
		t.Fatalf("HandleStreamPosition: want EOF got %v, %v", line, err)
	}
}

func TestVehicleHandler_HandleStreamPosition_UnknownVIN(t *testing.T) {
	store := NewMemStore()
	handler := NewVehicleHandler(store)

	r := httptest.NewRequest(http.MethodGet, "/the1vin/stream", nil)
	w := httptest.NewRecorder()

	if err := handler.HandleStreamPosition(w, r); err == nil {
		t.Fatal("HandleStreamPosition: want error, got nil")
	}
}
