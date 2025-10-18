package server

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"github.com/gin-gonic/gin"
)

type Server struct{
	port int
	engine *gin.Engine
	db *sql.DB
}

func NewServer() (*Server, error){
	s := Server{port: 8000, engine: gin.Default()}
	if err := s.initializeState(); err != nil{
		return nil, err
	}
	if err := s.buildRoutes(); err != nil{
		return nil, err
	}
	return &s, nil
}

func (s *Server) Start() error {
	s.engine.Run(fmt.Sprintf(":%d", s.port))
	return nil
}

func (s *Server) initializeState() error{
	var err error
	database, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(fmt.Sprintf("Failed to open database: %v", err))
	}
	defer database.Close()
	s.db = database
	log.Printf("Database initialized\n")
	return nil
}


