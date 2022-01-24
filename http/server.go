package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mattmeyers/feedsync/store"
)

type Server struct {
	router *httprouter.Router

	feedStore store.FeedStore
}

func NewServer(feedStore store.FeedStore) *Server {
	s := &Server{
		router:    httprouter.New(),
		feedStore: feedStore,
	}
	s.loadRoutes()

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) ListenAndServe(addr string) error {
	fmt.Println("Starting server on", addr)
	return http.ListenAndServe(addr, s)
}

func (s *Server) handleFeedsList() http.Handler {
	type responseBody struct {
		Response []store.Feed `json:"response"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		feeds, err := s.feedStore.List()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resBody := responseBody{Response: feeds}

		resBytes, err := json.Marshal(resBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(resBytes)
	})
}

func (s *Server) handleFeedsInsert() http.Handler {
	type requestBody struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}

	type responseBody store.Feed

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody requestBody
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		id, err := s.feedStore.Insert(store.Feed{Name: reqBody.Name, URL: reqBody.URL})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resBody := responseBody{
			ID:   id,
			Name: reqBody.Name,
			URL:  reqBody.URL,
		}

		resBytes, err := json.Marshal(resBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(resBytes)
	})
}
