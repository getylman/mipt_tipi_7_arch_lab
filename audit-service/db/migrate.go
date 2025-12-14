package db

import (
    "database/sql"
    "embed"
    "fmt"
    "io/fs"

    "github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func RunMigrations(db *sql.DB) error {
    goose.SetBaseFS(migrations)
    
    if err := goose.SetDialect("postgres"); err != nil {
        return fmt.Errorf("failed to set dialect: %w", err)
    }
    
    // Получаем список файлов миграций
    entries, err := fs.ReadDir(migrations, "migrations")
    if err != nil {
        return fmt.Errorf("failed to read migrations: %w", err)
    }
    
    if len(entries) == 0 {
        return fmt.Errorf("no migration files found")
    }
    
    if err := goose.Up(db, "migrations"); err != nil {
        return fmt.Errorf("failed to run migrations: %w", err)
    }
    
    return nil
}