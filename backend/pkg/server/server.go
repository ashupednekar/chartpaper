package server

import (
	"fmt"
)

func (s *Server) Start() error {
	s.engine.Run(fmt.Sprintf(":%d", s.port))
	return nil
}
