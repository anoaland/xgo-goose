package xgoose

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"text/template"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"gorm.io/gorm"

	dbutils "github.com/anoaland/xgo/db/utils"
)

type GooseMigrator struct {
	db         *sql.DB
	fsys       fs.FS
	provider   *goose.Provider
	dir        string
	noMigraion bool
}

func NewMigrator(dialect database.Dialect, dsn string, fsys fs.FS, dir string) GooseMigrator {
	db, err := goose.OpenDBWithDriver(string(dialect), dsn)
	if err != nil {
		panic(err)
	}

	provider, err := goose.NewProvider(
		dialect,
		db,
		fsys,
	)
	goose.SetBaseFS(fsys)

	noMigraion := false
	if err != nil {
		if err == goose.ErrNoMigrations {
			noMigraion = true
		} else {
			panic(err)
		}
	}

	return GooseMigrator{
		db:         db,
		fsys:       fsys,
		provider:   provider,
		noMigraion: noMigraion,
		dir:        dir,
	}
}

func (g GooseMigrator) generateName() string {
	randomNumber := rand.Intn(3) + 1
	name := petname.Generate(randomNumber, "_")

	return name
}

func (g GooseMigrator) Create() {
	name := g.generateName()
	goose.SetBaseFS(g.fsys)
	err := goose.Create(g.db, g.dir, name, "sql")

	if err != nil {
		panic(err)
	}
}

func (g GooseMigrator) CreateWithStatement(up string, down string) {
	name := g.generateName()
	goose.SetBaseFS(g.fsys)

	sqlMigrationTemplate := template.Must(template.New("goose.sql-migration").Parse(fmt.Sprintf(`-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query for %s';
%s
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query for %s';
%s
-- +goose StatementEnd
`, name, up, name, down)))

	err := goose.CreateWithTemplate(g.db, g.dir, sqlMigrationTemplate, name, "sql")

	if err != nil {
		panic(err)
	}
}

func (g GooseMigrator) CreateFromGormModels(db *gorm.DB, dst ...interface{}) {
	sql := dbutils.PrintAutoMigrateSql(db, dst...)
	g.CreateWithStatement(sql, "-- TODO: Create your own down migration")
}

func (g GooseMigrator) Up() {
	if g.noMigraion {
		log.Fatal("No migration found!")
	}

	res, err := g.provider.Up(context.Background())
	if err != nil {
		panic(err)
	}

	log.Printf("%v", res)
}

func (g GooseMigrator) Down() {
	if g.noMigraion {
		log.Fatal("No migration found!")
	}

	res, err := g.provider.Down(context.Background())
	if err != nil {
		panic(err)
	}

	log.Printf("%v", res)
}

// Create writes a new blank migration file.
