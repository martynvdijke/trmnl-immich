package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var immichUserAgent = "trmnl-immich/" + Version

type immichClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func newImmichClient(baseURL, apiKey string) *immichClient {
	return &immichClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *immichClient) do(ctx context.Context, method, path string, query url.Values, body any) (*http.Response, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, err
	}
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", immichUserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.http.Do(req)
}

// asset is a subset of Immich's AssetResponseDto.
type asset struct {
	ID               string   `json:"id"`
	OriginalFileName string   `json:"originalFileName"`
	FileCreatedAt    string   `json:"fileCreatedAt"`
	Width            int      `json:"width"`
	Height           int      `json:"height"`
	ExifInfo         exifInfo `json:"exifInfo"`
}

type exifInfo struct {
	Description    string  `json:"description"`
	Make           string  `json:"make"`
	Model          string  `json:"model"`
	LensModel      string  `json:"lensModel"`
	FocalLength    float64 `json:"focalLength"`
	ISO            int     `json:"iso"`
	ExposureTime   string  `json:"exposureTime"`
	FNumber        float64 `json:"fNumber"`
	FileSizeInByte int64   `json:"fileSizeInByte"`
	City           string  `json:"city"`
	State          string  `json:"state"`
	Country        string  `json:"country"`
}

// photo is the JSON payload polled by the TRMNL device.
type photo struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Date         string  `json:"date"`
	Make         string  `json:"make"`
	Model        string  `json:"model"`
	LensModel    string  `json:"lens_model"`
	FocalLength  float64 `json:"focal_length"`
	ISO          int     `json:"iso"`
	ExposureTime string  `json:"exposure_time"`
	FNumber      float64 `json:"f_number"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	FileSize     int64   `json:"file_size"`
	City         string  `json:"city"`
	State        string  `json:"state"`
	Country      string  `json:"country"`
	Image        string  `json:"image"`
}

func (a *asset) toPhoto() photo {
	title := a.OriginalFileName
	if a.ExifInfo.Description != "" {
		title = a.ExifInfo.Description
	}
	return photo{
		ID:           a.ID,
		Title:        title,
		Date:         a.FileCreatedAt,
		Make:         a.ExifInfo.Make,
		Model:        a.ExifInfo.Model,
		LensModel:    a.ExifInfo.LensModel,
		FocalLength:  a.ExifInfo.FocalLength,
		ISO:          a.ExifInfo.ISO,
		ExposureTime: a.ExifInfo.ExposureTime,
		FNumber:      a.ExifInfo.FNumber,
		Width:        a.Width,
		Height:       a.Height,
		FileSize:     a.ExifInfo.FileSizeInByte,
		City:         a.ExifInfo.City,
		State:        a.ExifInfo.State,
		Country:      a.ExifInfo.Country,
		Image:        "/api/trmnl/photo/" + url.PathEscape(a.ID),
	}
}

// randomSearchDto mirrors Immich's RandomSearchDto.
type randomSearchDto struct {
	Size      int      `json:"size"`
	AlbumIDs  []string `json:"albumIds,omitempty"`
	PersonIDs []string `json:"personIds,omitempty"`
}

func (c *immichClient) randomPhoto(ctx context.Context, albumID, personID string) (*photo, error) {
	body := randomSearchDto{Size: 1}
	if albumID != "" {
		body.AlbumIDs = []string{albumID}
	}
	if personID != "" {
		body.PersonIDs = []string{personID}
	}
	resp, err := c.do(ctx, http.MethodPost, "/api/search/random", nil, body)
	if err != nil {
		return nil, fmt.Errorf("search random: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search random: unexpected status %d", resp.StatusCode)
	}
	var assets []asset
	if err := json.NewDecoder(resp.Body).Decode(&assets); err != nil {
		return nil, fmt.Errorf("search random: decode: %w", err)
	}
	if len(assets) == 0 {
		return nil, fmt.Errorf("no photos found in Immich library")
	}
	p := assets[0].toPhoto()
	return &p, nil
}

func (c *immichClient) handlePhotoOfTheDay(w http.ResponseWriter, r *http.Request) {
	albumID := r.URL.Query().Get("album_id")
	personID := r.URL.Query().Get("person_id")
	p, err := c.randomPhoto(r.Context(), albumID, personID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// handlePhoto proxies the Immich thumbnail so the TRMNL device can load
// images without holding an Immich API key.
func (c *immichClient) handlePhoto(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing photo id")
		return
	}
	query := url.Values{"size": []string{"preview"}}
	resp, err := c.do(r.Context(), http.MethodGet, "/api/assets/"+url.PathEscape(id)+"/thumbnail", query, nil)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		writeError(w, resp.StatusCode, fmt.Sprintf("upstream returned %d", resp.StatusCode))
		return
	}
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("handlePhoto: copy: %v", err)
	}
}
