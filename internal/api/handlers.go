package api

import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"
    "strings"

    "go-demo-crud/internal/model"
    "go-demo-crud/internal/repo"
)

type Server struct {
    Repo *repo.ProductRepo
}

func NewServer(r *repo.ProductRepo) *Server { return &Server{Repo: r} }

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("/health", s.healthHandler)
    mux.HandleFunc("/products", s.productsHandler)
    mux.HandleFunc("/products/", s.productByIDHandler)
}

func respondJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if v == nil {
        return
    }
    _ = json.NewEncoder(w).Encode(v)
}

func parseIDFromPath(path string) (int64, error) {
    parts := strings.Split(strings.Trim(path, "/"), "/")
    if len(parts) < 2 {
        return 0, errors.New("missing id")
    }
    return strconv.ParseInt(parts[1], 10, 64)
}

func (s *Server) productsHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        items, err := s.Repo.ListProducts()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        respondJSON(w, http.StatusOK, items)
    case http.MethodPost:
        var in model.Product
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            http.Error(w, "invalid JSON", http.StatusBadRequest)
            return
        }
        if strings.TrimSpace(in.Name) == "" {
            http.Error(w, "name is required", http.StatusBadRequest)
            return
        }
        out, err := s.Repo.CreateProduct(in)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        respondJSON(w, http.StatusCreated, out)
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
}

func (s *Server) productByIDHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        id, err := parseIDFromPath(r.URL.Path)
        if err != nil {
            http.Error(w, "invalid id", http.StatusBadRequest)
            return
        }
        p, err := s.Repo.GetProduct(id)
        if errors.Is(err, repo.ErrNotFound) || errors.Is(err, strconv.ErrSyntax) {
            http.NotFound(w, r)
            return
        }
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        respondJSON(w, http.StatusOK, p)
    case http.MethodPut:
        id, err := parseIDFromPath(r.URL.Path)
        if err != nil {
            http.Error(w, "invalid id", http.StatusBadRequest)
            return
        }
        var in model.Product
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
            http.Error(w, "invalid JSON", http.StatusBadRequest)
            return
        }
        if strings.TrimSpace(in.Name) == "" {
            http.Error(w, "name is required", http.StatusBadRequest)
            return
        }
        ok, err := s.Repo.UpdateProduct(id, in)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        if !ok {
            http.NotFound(w, r)
            return
        }
        in.ProductID = id
        respondJSON(w, http.StatusOK, in)
    case http.MethodDelete:
        id, err := parseIDFromPath(r.URL.Path)
        if err != nil {
            http.Error(w, "invalid id", http.StatusBadRequest)
            return
        }
        ok, err := s.Repo.DeleteProduct(id)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        if !ok {
            http.NotFound(w, r)
            return
        }
        respondJSON(w, http.StatusNoContent, nil)
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
    respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

