package main

import (
	"chartpaper/pkg/server"
	"log"
)

func main(){
	s, err := server.NewServer()
	if err != nil{
		log.Fatal(err)
	}
	if err := s.Start(); err != nil{
		log.Fatal(err)
	}	
}

