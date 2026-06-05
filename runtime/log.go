package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// logEntry is one durable mutation in the append-only write-ahead log.
type logEntry struct {
	Op     string  `json:"op"` // "put" | "relate"
	Record *Record `json:"record,omitempty"`
	Edge   *Edge   `json:"edge,omitempty"`
}

// Log is a simple append-only JSONL store. Instance mutations are appended as
// they happen and replayed on boot, so the runtime survives restarts without a
// heavyweight embedded database. Seeded model (meta) data is not logged — it is
// deterministic from gen/model.json and re-seeded each boot.
type Log struct {
	mu   sync.Mutex
	path string
	f    *os.File
}

// OpenLog opens (creating parent dirs) the JSONL log at path for appending.
func OpenLog(path string) (*Log, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &Log{path: path, f: f}, nil
}

// Append writes one entry. Errors are intentionally swallowed: a durability
// hiccup must not take down request handling for an alpha runtime.
func (l *Log) Append(e logEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, err := json.Marshal(e)
	if err != nil {
		return
	}
	l.f.Write(append(b, '\n'))
	l.f.Sync()
}

// Replay reads the log from disk and applies each entry to g without
// re-persisting (persist=false), reconstructing prior runtime state.
func (l *Log) Replay(g *Graph) error {
	f, err := os.Open(l.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var e logEntry
		if err := json.Unmarshal(line, &e); err != nil {
			continue // skip a corrupt line rather than refuse to boot
		}
		switch e.Op {
		case "put":
			if e.Record != nil {
				g.Put(e.Record, false)
			}
		case "relate":
			if e.Edge != nil {
				g.Relate(e.Edge, false)
			}
		}
	}
	return sc.Err()
}

// Close flushes and closes the underlying file.
func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.f.Close()
}
