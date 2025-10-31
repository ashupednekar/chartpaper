package main

import (
	"chartpaper/pkg/server"
	"log"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

var (
	databaseURL string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "chartpaper",
		Short: "Chartpaper API server and CLI",
	}

	listenCmd := &cobra.Command{
		Use:   "listen",
		Short: "Start the Chartpaper API server",
		Run: func(cmd *cobra.Command, args []string) {
			s, err := server.NewServer()
			if err != nil {
				log.Fatal(err)
			}
			if err := s.Start(); err != nil {
				log.Fatal(err)
			}
		},
	}

	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Run: func(cmd *cobra.Command, args []string) {
			s, err := server.NewServer()
			if err != nil{
				log.Fatalf("couldn't initialize state")
			}
			db := stdlib.OpenDBFromPool(s.GetPool())
			defer db.Close()
			if err := goose.SetDialect("pgx"); err != nil {
				log.Fatalf("goose: failed to set dialect: %v\n", err)
			}
			arguments := []string{}
			if len(args) > 0 {
				arguments = append(arguments, args...)
			}

			if err := goose.RunContext(cmd.Context(), "up", db, "internal/db/migrations", arguments...); err != nil {
				log.Fatalf("goose: migrate failed: %v\n", err)
			}
		},
	}
	migrateCmd.Flags().StringVarP(&databaseURL, "database-url", "d", "", "Database connection URL (defaults to DATABASE_URL environment variable)")

	rootCmd.AddCommand(listenCmd)
	rootCmd.AddCommand(migrateCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

