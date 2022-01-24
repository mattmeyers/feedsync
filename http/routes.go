package http

import "net/http"

func (s *Server) loadRoutes() {
	s.router.Handler(http.MethodGet, "/v1/feeds", s.handleFeedsList())
	s.router.Handler(http.MethodPost, "/v1/feeds", s.handleFeedsInsert())
}
