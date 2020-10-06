package vehicle

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFleetStateClient_UpdatePosition(t *testing.T) {
	done := make(chan struct{})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if want := http.MethodPost; want != r.Method {
			t.Fatalf("method: want %s got %s", want, r.Method)
		}
		if want, got := "/vehicle/THE1VIN", r.URL.Path; want != got {
			t.Fatalf("url: want %s got %s", want, r.URL.Path)
		}
		if want, got := "52.520008", r.PostFormValue("lat"); want != got {
			t.Fatalf("lat: want %s got %s", want, got)
		}
		if want, got := "13.401797", r.PostFormValue("lon"); want != got {
			t.Fatalf("lon: want %s got %s", want, got)
		}

		w.WriteHeader(http.StatusCreated)

		close(done)
	}))
	defer ts.Close()

	client := NewFleetStateClient(ts.URL)
	client.Client = ts.Client()

	if err := client.UpdatePosition(context.Background(), "THE1VIN", 52.520008, 13.401797); err != nil {
		t.Fatal(err)
	}

	<-done
}

func TestFleetStateClient_UpdatePosition_BadStatus(t *testing.T) {
	done := make(chan struct{})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)

		close(done)
	}))
	defer ts.Close()

	client := NewFleetStateClient(ts.URL)
	client.Client = ts.Client()

	if err := client.UpdatePosition(context.Background(), "THE1VIN", 52.520008, 13.401797); err == nil {
		t.Fatal("UpdatePosition: want err for nil")
	}

	<-done
}
