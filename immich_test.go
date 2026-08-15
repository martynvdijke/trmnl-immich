package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testKey = "test-api-key"

func fakeImmich(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != testKey {
			t.Errorf("expected x-api-key %q, got %q", testKey, r.Header.Get("x-api-key"))
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/api/search/random":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var body randomSearchDto
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode body: %v", err)
			}
			if body.Size != 1 {
				t.Errorf("expected size 1, got %d", body.Size)
			}
			_ = body
			assets := []asset{
				{
					ID:               "asset-1",
					OriginalFileName: "sunset.jpg",
					FileCreatedAt:    "2026-08-01T19:30:00Z",
					Width:            2400,
					Height:           1600,
					ExifInfo: exifInfo{
						Description:    "Sunset over the dunes",
						Make:           "Fujifilm",
						Model:          "X-T5",
						LensModel:      "XF35mmF2 R WR",
						FocalLength:    35,
						ISO:            160,
						ExposureTime:   "1/500",
						FNumber:        2,
						FileSizeInByte: 512000,
						City:           "Amsterdam",
						State:          "Noord-Holland",
						Country:        "Netherlands",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(assets)
		case "/api/assets/asset-1/thumbnail":
			w.Header().Set("Content-Type", "image/webp")
			_, _ = w.Write([]byte("fake-thumbnail-bytes"))
		default:
			http.NotFound(w, r)
		}
	}))
}

func newTestClient(t *testing.T, server *httptest.Server) *immichClient {
	t.Helper()
	return newImmichClient(server.URL, testKey)
}

func TestRandomPhoto(t *testing.T) {
	server := fakeImmich(t)
	defer server.Close()
	client := newTestClient(t, server)

	p, err := client.randomPhoto(t.Context(), "", "")
	if err != nil {
		t.Fatalf("randomPhoto: %v", err)
	}
	if p.ID != "asset-1" {
		t.Errorf("expected id asset-1, got %q", p.ID)
	}
	if p.Title != "Sunset over the dunes" {
		t.Errorf("expected description as title, got %q", p.Title)
	}
	if p.Make != "Fujifilm" || p.Model != "X-T5" {
		t.Errorf("expected camera metadata, got %q %q", p.Make, p.Model)
	}
	if p.City != "Amsterdam" || p.Country != "Netherlands" {
		t.Errorf("expected location metadata, got %q %q", p.City, p.Country)
	}
	if p.Image != "/api/trmnl/photo/asset-1" {
		t.Errorf("expected relative image path, got %q", p.Image)
	}
}

func TestRandomPhotoWithFilters(t *testing.T) {
	server := fakeImmich(t)
	defer server.Close()
	client := newTestClient(t, server)

	p, err := client.randomPhoto(t.Context(), "album-9", "person-4")
	if err != nil {
		t.Fatalf("randomPhoto: %v", err)
	}
	if p.ID != "asset-1" {
		t.Errorf("expected id asset-1, got %q", p.ID)
	}
}

func TestPhotoOfTheDayHandler(t *testing.T) {
	server := fakeImmich(t)
	defer server.Close()
	client := newTestClient(t, server)

	req := httptest.NewRequest(http.MethodGet, "/api/trmnl/photo-of-the-day", nil)
	rec := httptest.NewRecorder()
	client.handlePhotoOfTheDay(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var p photo
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if p.Image != "/api/trmnl/photo/asset-1" {
		t.Errorf("unexpected image: %q", p.Image)
	}
}

func TestPhotoHandler(t *testing.T) {
	server := fakeImmich(t)
	defer server.Close()
	client := newTestClient(t, server)

	req := httptest.NewRequest(http.MethodGet, "/api/trmnl/photo/asset-1", nil)
	req.SetPathValue("id", "asset-1")
	rec := httptest.NewRecorder()
	client.handlePhoto(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "fake-thumbnail-bytes" {
		t.Errorf("unexpected body: %q", rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/webp" {
		t.Errorf("expected image/webp, got %q", ct)
	}
}

func TestPhotoHandlerMissingID(t *testing.T) {
	server := fakeImmich(t)
	defer server.Close()
	client := newTestClient(t, server)

	req := httptest.NewRequest(http.MethodGet, "/api/trmnl/photo/", nil)
	req.SetPathValue("id", "")
	rec := httptest.NewRecorder()
	client.handlePhoto(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "missing photo id") {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}
