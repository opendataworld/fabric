// Command runtime is the Fabric data-fabric runtime: an in-memory multi-model
// graph store, seeded self-describingly from the canonical Fabric model
// (gen/model.json), exposed over a GraphQL API.
//
// Usage:
//
//	runtime                # serve GraphQL on :8088 (override with PORT)
//	runtime --selftest     # in-process checks, no socket
//
// Environment:
//
//	PORT             listen port (default 8088)
//	FABRIC_ROOT      repo root holding gen/ (default: discovered upward)
//	FABRIC_DATA_DIR  durable graph log dir (default: <root>/api/data)
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func dataDir(model *Model) string {
	if d := os.Getenv("FABRIC_DATA_DIR"); d != "" {
		return d
	}
	// gen/ lives at <root>/gen; keep runtime data beside the Python server's.
	return filepath.Join(filepath.Dir(model.genDir), "api", "data")
}

// build assembles the model, durable log, seeded graph, and GraphQL API.
func build() (*API, *Log, error) {
	model, err := LoadModel()
	if err != nil {
		return nil, nil, fmt.Errorf("load model: %w", err)
	}
	logFile, err := OpenLog(filepath.Join(dataDir(model), "graph.jsonl"))
	if err != nil {
		return nil, nil, fmt.Errorf("open log: %w", err)
	}
	g := NewGraph(logFile)
	model.Seed(g)                             // self-describing meta layer (not persisted)
	if err := logFile.Replay(g); err != nil { // durable instance data
		return nil, nil, fmt.Errorf("replay log: %w", err)
	}
	return &API{Model: model, Graph: g}, logFile, nil
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--selftest" {
		if err := selftest(); err != nil {
			log.Fatalf("selftest FAILED: %v", err)
		}
		return
	}

	api, logFile, err := build()
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	// Agent-native transport: speak MCP over stdio so any agent can use the
	// fabric as tools. Diagnostics must stay on stderr (stdout is the channel).
	if len(os.Args) > 1 && os.Args[1] == "--mcp" {
		st := api.Graph.Stats()
		log.Printf("Fabric runtime (MCP/stdio): %d primitives, %d records, %d edges",
			len(api.Model.Classes), st["records"], st["edges"])
		if err := NewMCPServer(api).Serve(); err != nil {
			log.Fatal(err)
		}
		return
	}

	srv, err := NewServer(api)
	if err != nil {
		log.Fatalf("build schema: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}
	st := api.Graph.Stats()
	log.Printf("Fabric runtime: %d primitives, %d records, %d edges",
		len(api.Model.Classes), st["records"], st["edges"])
	log.Printf("GraphQL on http://localhost:%s/graphql  (explorer at /)", port)
	if err := http.ListenAndServe(":"+port, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
