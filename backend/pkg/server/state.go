package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	port   int
	engine *gin.Engine
	db     *pgxpool.Pool
}

func NewServer() (*Server, error) {
	s := Server{port: 8000, engine: gin.Default()}
	if err := s.initializeState(); err != nil {
		return nil, err
	}
	if err := s.buildRoutes(); err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *Server) initializeState() error {
	var err error
	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("error parsing DATABASE_URL: %w", err)
	}
	config.ConnConfig.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	s.db = pool
	log.Printf("Database initialized\n")
	return nil
}

