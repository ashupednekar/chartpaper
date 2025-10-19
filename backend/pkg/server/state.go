package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
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
	databaseURL := os.Getenv("DATABASE_URL")
	var err error
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("error parsing DATABASE_URL: %w", err)
	}
	config.MaxConns = 5
	config.ConnConfig.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	config.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
	  _, err := c.Exec(ctx, "create schema if not exists chartpaper;") 	
		if err != nil{
			return fmt.Errorf("error craeting db schema: %s", err)
		}
	  _, err = c.Exec(ctx, "set search_path to chartpaper;")	
		if err != nil{
			return fmt.Errorf("error setting search_path: %s", err)
		}
		return nil
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	s.db = pool
	log.Printf("Database initialized\n")
	return nil
}

func (s *Server) GetPool() *pgxpool.Pool {
	return s.db
}
