package fleetstate

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/narqo/ree-fleet-sim/internal/geoutil"
	"github.com/narqo/ree-fleet-sim/internal/vehicle"
)

var ErrNotFound = errors.New("not found")

type VehicleHandler struct {
	store Store
}

func NewVehicleHandler(store Store) *VehicleHandler {
	return &VehicleHandler{
		store: store,
	}
}

func (h *VehicleHandler) Handler() http.Handler {
	return errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			return h.HandleUpdatePosition(w, r)
		}
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/stream") {
			return h.HandleStreamPosition(w, r)
		}
		return ErrNotFound
	})
}

func errorHandler(handle func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := handle(w, r)
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (h *VehicleHandler) HandleUpdatePosition(w http.ResponseWriter, r *http.Request) error {
	ts := time.Now().UTC()

	vin, err := extractVINFromURLPath(r.URL.Path)
	if err != nil {
		return fmt.Errorf("bad vin: %w", err)
	}

	lat, err := strconv.ParseFloat(r.PostFormValue("lat"), 64)
	if err != nil {
		return fmt.Errorf("bad lat: %w", err)
	}

	lon, err := strconv.ParseFloat(r.PostFormValue("lon"), 64)
	if err != nil {
		return fmt.Errorf("bad lon: %w", err)
	}

	if err := h.store.Write(r.Context(), vin, ts, lat, lon); err != nil {
		return fmt.Errorf("could not write position for vin %q: %w", vin, err)
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

type PositionResponse struct {
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	Speed float64 `json:"speed"`
	Error string  `json:"error,omitempty"`
}

func (h *VehicleHandler) HandleStreamPosition(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("bad request: client doens't support streaming")
	}

	w.Header().Set("Transfer-Encoding", "chunked")

	vin, err := extractVINFromURLPath(r.URL.Path)
	if err != nil {
		return fmt.Errorf("bad vin: %w", err)
	}

	reader, err := h.store.Reader(ctx, vin)
	if err != nil {
		return err
	}

	sw := &streamWriter{
		f:   flusher,
		enc: json.NewEncoder(w),
	}

	ts0, lat0, lon0, err := reader.Read()
	if err != nil {
		resp := PositionResponse{Error: err.Error()}
		sw.WriteChunk(resp)
	}

	for {
		var resp PositionResponse
		ts1, lat1, lon1, err := reader.Read()
		if err != nil {
			resp.Error = err.Error()
		} else {
			resp.Lat = lat1
			resp.Lon = lon1
			d := geoutil.Distance(lat0, lon0, lat1, lon1)
			if d != 0 {
				resp.Speed = d / ts1.Sub(ts0).Hours()
			}

			ts0, lat0, lon0 = ts1, lat1, lon1
		}

		select {
		case <-ctx.Done():
			// client has gone, nothing left to do
			return nil
		default:
			sw.WriteChunk(resp)
		}
	}
}

type streamWriter struct {
	f   http.Flusher
	enc *json.Encoder
}

func (w *streamWriter) WriteChunk(resp PositionResponse) {
	if err := w.enc.Encode(resp); err != nil {
		log.Printf("streamWriter: failed to encode json: %s", err)
	} else {
		w.f.Flush()
	}
}

func extractVINFromURLPath(p string) (vehicle.VIN, error) {
	var s string
	p = strings.Trim(path.Clean(p), "/")
	chunks := strings.SplitN(p, "/", 2)
	if len(chunks) > 0 {
		s = chunks[0]
	}
	return vehicle.VINFromString(s)
}
