package main

import (
	"encoding/json"
	"net/http"

	"github.com/graphql-go/graphql"
)

// Server serves the GraphQL API and an embedded GraphiQL explorer.
type Server struct {
	schema graphql.Schema
}

// NewServer compiles the API's schema and returns an http.Handler.
func NewServer(a *API) (*Server, error) {
	schema, err := a.BuildSchema()
	if err != nil {
		return nil, err
	}
	return &Server{schema: schema}, nil
}

func (s *Server) cors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// Handler wires the routes: GraphiQL at /, the GraphQL endpoint at /graphql.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/graphql", s.graphql)
	mux.HandleFunc("/", s.root)
	return mux
}

type gqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
	Operation string         `json:"operationName"`
}

func (s *Server) graphql(w http.ResponseWriter, r *http.Request) {
	s.cors(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	var req gqlRequest
	switch r.Method {
	case http.MethodGet:
		req.Query = r.URL.Query().Get("query")
	case http.MethodPost:
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.writeJSON(w, http.StatusBadRequest, map[string]any{"errors": []any{map[string]string{"message": "invalid JSON body"}}})
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if req.Query == "" {
		s.writeJSON(w, http.StatusBadRequest, map[string]any{"errors": []any{map[string]string{"message": "no query provided"}}})
		return
	}
	result := graphql.Do(graphql.Params{
		Schema:         s.schema,
		RequestString:  req.Query,
		VariableValues: req.Variables,
		OperationName:  req.Operation,
		Context:        r.Context(),
	})
	s.writeJSON(w, http.StatusOK, result)
}

func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	s.cors(w)
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(graphiQL))
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(body)
}

// graphiQL is a zero-dependency explorer that posts queries to /graphql. It
// keeps the runtime a single binary with no asset pipeline.
const graphiQL = `<!doctype html>
<html><head><meta charset="utf-8"><title>Fabric Runtime · GraphQL</title>
<style>
 body{margin:0;font:14px/1.5 ui-monospace,SFMono-Regular,Menlo,monospace;background:#161616;color:#f4f4f4}
 header{padding:12px 16px;background:#262626;border-bottom:1px solid #393939}
 header b{color:#78a9ff}
 main{display:grid;grid-template-columns:1fr 1fr;gap:1px;background:#393939;height:calc(100vh - 49px)}
 textarea,pre{margin:0;padding:16px;border:0;background:#161616;color:#f4f4f4;font:inherit;overflow:auto}
 textarea{resize:none;outline:none}
 button{position:absolute;right:16px;top:9px;background:#0f62fe;color:#fff;border:0;padding:6px 14px;cursor:pointer}
</style></head>
<body>
<header><b>Fabric</b> data fabric runtime — GraphQL <button onclick="run()">Run ▶</button></header>
<main>
 <textarea id="q">{
  health
  classes { id name question }
  graph { edges { from rel to } }
}</textarea>
 <pre id="out">// results appear here</pre>
</main>
<script>
async function run(){
 const r = await fetch('/graphql',{method:'POST',headers:{'Content-Type':'application/json'},
   body:JSON.stringify({query:document.getElementById('q').value})});
 document.getElementById('out').textContent = JSON.stringify(await r.json(),null,2);
}
document.addEventListener('keydown',e=>{if((e.metaKey||e.ctrlKey)&&e.key==='Enter')run()});
run();
</script>
</body></html>`
