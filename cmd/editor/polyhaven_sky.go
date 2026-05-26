package main

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"sort"
	"sync"
)

const polyhavenHdrisURL = "https://api.polyhaven.com/assets?type=hdris"

type polyhavenHdriResolution struct {
	Hdr *polyhavenIncludeFile `json:"hdr"`
}

type polyhavenHdriFiles struct {
	Hdri map[string]polyhavenHdriResolution `json:"hdri"`
}

// PendingHDRI is a decoded equirectangular HDRI ready to be uploaded to the GPU.
type PendingHDRI struct {
	DisplayName string
	Width       int
	Height      int
	Pixels      []byte
}

// PolyhavenSky fetches the Poly Haven HDRI index and downloads individual
// environment maps, decoding them off the main thread.
type PolyhavenSky struct {
	mu         sync.Mutex
	status     polyhavenIndexStatus
	entries    []PolyhavenEntry
	loading    string
	pending    *PendingHDRI
	wantRandom bool
}

func NewPolyhavenSky() *PolyhavenSky {
	return &PolyhavenSky{}
}

// TakePending removes and returns a finished download, or nil if none is ready.
func (b *PolyhavenSky) TakePending() *PendingHDRI {
	b.mu.Lock()
	defer b.mu.Unlock()
	pending := b.pending
	b.pending = nil
	return pending
}

// FetchRandom downloads a random HDRI, kicking off the index fetch first if it
// has not loaded yet and deferring the pick until the index arrives.
func (b *PolyhavenSky) FetchRandom() {
	b.mu.Lock()
	switch b.status {
	case polyhavenIdle:
		b.status = polyhavenLoading
		b.wantRandom = true
		b.mu.Unlock()
		go b.fetchIndex()
		return
	case polyhavenLoading:
		b.wantRandom = true
		b.mu.Unlock()
		return
	case polyhavenFailed:
		b.mu.Unlock()
		return
	}
	entries := b.entries
	b.mu.Unlock()
	b.fetchRandomFrom(entries)
}

func (b *PolyhavenSky) fetchRandomFrom(entries []PolyhavenEntry) {
	if len(entries) == 0 {
		return
	}
	b.fetchEntry(entries[rand.IntN(len(entries))])
}

func (b *PolyhavenSky) fetchEntry(entry PolyhavenEntry) {
	if !b.beginLoading(entry.Name) {
		return
	}
	go b.fetchHdri(entry)
}

func (b *PolyhavenSky) beginLoading(displayName string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.loading != "" {
		return false
	}
	b.loading = displayName
	return true
}

func (b *PolyhavenSky) clearLoading() {
	b.mu.Lock()
	b.loading = ""
	b.mu.Unlock()
}

func (b *PolyhavenSky) publishPending(pending *PendingHDRI) {
	b.mu.Lock()
	b.pending = pending
	b.loading = ""
	b.mu.Unlock()
}

func (b *PolyhavenSky) fetchIndex() {
	data, ok := httpGetBytes(polyhavenHdrisURL)
	if !ok {
		b.failIndex("fetch index")
		return
	}
	var raw map[string]polyhavenRawAsset
	if err := json.Unmarshal(data, &raw); err != nil {
		b.failIndex("parse: " + err.Error())
		return
	}
	entries := polyhavenEntriesFromRaw(raw)
	b.mu.Lock()
	b.entries = entries
	b.status = polyhavenLoaded
	wantRandom := b.wantRandom
	b.wantRandom = false
	b.mu.Unlock()
	if wantRandom {
		b.fetchRandomFrom(entries)
	}
}

func (b *PolyhavenSky) failIndex(message string) {
	log.Printf("polyhaven sky: index %s", message)
	b.mu.Lock()
	b.status = polyhavenFailed
	b.mu.Unlock()
}

func (b *PolyhavenSky) fetchHdri(entry PolyhavenEntry) {
	data, ok := httpGetBytes(polyhavenFilesURL + entry.Slug)
	if !ok {
		b.clearLoading()
		return
	}
	var files polyhavenHdriFiles
	if err := json.Unmarshal(data, &files); err != nil {
		b.clearLoading()
		return
	}
	url, ok := pickHdriURL(files.Hdri, polyhavenPreferredResolution)
	if !ok {
		b.clearLoading()
		return
	}
	hdrBytes, ok := httpGetBytes(url)
	if !ok {
		b.clearLoading()
		return
	}
	width, height, pixels, err := decodeHDR(hdrBytes)
	if err != nil {
		log.Printf("polyhaven sky: decode %s: %v", entry.Name, err)
		b.clearLoading()
		return
	}
	b.publishPending(&PendingHDRI{
		DisplayName: entry.Name,
		Width:       width,
		Height:      height,
		Pixels:      pixels,
	})
}

func pickHdriURL(resolutions map[string]polyhavenHdriResolution, preferred uint32) (string, bool) {
	type candidate struct {
		value uint32
		url   string
	}
	candidates := make([]candidate, 0, len(resolutions))
	for key, resolution := range resolutions {
		if resolution.Hdr == nil {
			continue
		}
		candidates = append(candidates, candidate{value: polyhavenResolutionValue(key), url: resolution.Hdr.URL})
	}
	if len(candidates) == 0 {
		return "", false
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].value < candidates[j].value })
	for _, item := range candidates {
		if item.value == preferred {
			return item.url, true
		}
	}
	chosen := candidates[0]
	for _, item := range candidates {
		if item.value <= preferred {
			chosen = item
		}
	}
	return chosen.url, true
}
