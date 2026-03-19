package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/tomassar/dioptra/internal/db"
)

type Server struct {
	db       *db.DB
	httpSrv  *http.Server
	listener net.Listener
}

type response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
	Meta  any    `json:"meta,omitempty"`
}

func New(database *db.DB, staticFS fs.FS) (*Server, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	s := &Server{
		db:       database,
		listener: listener,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/schemas", s.handleSchemas)
	mux.HandleFunc("GET /api/tables", s.handleTables)
	mux.HandleFunc("GET /api/tables/{schema}/{table}", s.handleTableData)
	mux.HandleFunc("POST /api/tables/{schema}/{table}/insert", s.handleInsert)
	mux.HandleFunc("POST /api/tables/{schema}/{table}/update", s.handleUpdate)
	mux.HandleFunc("POST /api/query", s.handleQuery)
	mux.HandleFunc("GET /api/status", s.handleStatus)
	mux.Handle("GET /", http.FileServerFS(staticFS))

	s.httpSrv = &http.Server{Handler: mux}
	return s, nil
}

func (s *Server) Port() int    { return s.listener.Addr().(*net.TCPAddr).Port }
func (s *Server) URL() string  { return fmt.Sprintf("http://127.0.0.1:%d", s.Port()) }
func (s *Server) Start() error { return s.httpSrv.Serve(s.listener) }

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}

func (s *Server) handleSchemas(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	schemas, err := s.db.Schemas(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Data: schemas})
}

func (s *Server) handleTables(w http.ResponseWriter, r *http.Request) {
	schema := r.URL.Query().Get("schema")
	if schema == "" {
		schema = "public"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	tables, err := s.db.Tables(ctx, schema)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response{Data: tables})
}

func (s *Server) handleTableData(w http.ResponseWriter, r *http.Request) {
	schema := r.PathValue("schema")
	table := r.PathValue("table")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	sortCol := r.URL.Query().Get("sortCol")
	sortDir := r.URL.Query().Get("sortDir")
	filterCol := r.URL.Query().Get("filterCol")
	filterVal := r.URL.Query().Get("filterVal")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	data, err := s.db.TableData(ctx, schema, table, page, pageSize, sortCol, sortDir, filterCol, filterVal)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Error: err.Error()})
		return
	}

	count, _ := s.db.TableCount(ctx, schema, table, filterCol, filterVal)
	pks, _ := s.db.TablePK(ctx, schema, table)

	writeJSON(w, http.StatusOK, response{
		Data: data,
		Meta: map[string]any{
			"page":        page,
			"pageSize":    pageSize,
			"totalRows":   count,
			"totalPages":  (count + int64(pageSize) - 1) / int64(pageSize),
			"primaryKeys": pks,
		},
	})
}

func (s *Server) handleInsert(w http.ResponseWriter, r *http.Request) {
	schema := r.PathValue("schema")
	table := r.PathValue("table")

	var data map[string]string
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := s.db.InsertRow(ctx, schema, table, data); err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, response{Data: "ok"})
}

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	schema := r.PathValue("schema")
	table := r.PathValue("table")

	var req struct {
		PKValues map[string]any `json:"pkValues"`
		Updates  map[string]any `json:"updates"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := s.db.UpdateRow(ctx, schema, table, req.PKValues, req.Updates); err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, response{Data: "ok"})
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SQL string `json:"sql"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "invalid request body"})
		return
	}
	if req.SQL == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "sql is required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := s.db.RunQuery(ctx, req.SQL)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data: result,
		Meta: map[string]int{"rowLimit": 1000},
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, response{
		Data: map[string]any{
			"readOnly": s.db.ReadOnly(),
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
