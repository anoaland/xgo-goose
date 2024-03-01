package xgoose

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"text/template"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"gorm.io/gorm"

	dbutils "github.com/anoaland/xgo/db/utils"
)

type GooseMigrator struct {
	db       *sql.DB
	fsys     fs.FS
	provider *goose.Provider
	dir      string
}

func NewMigrator(dialect database.Dialect, dsn string, fsys fs.FS, dir string) GooseMigrator {
	db, err := goose.OpenDBWithDriver(string(goose.DialectMSSQL), dsn)
	if err != nil {
		panic(err)
	}

	provider, err := goose.NewProvider(
		goose.DialectMSSQL,
		db,
		fsys,
	)
	goose.SetBaseFS(fsys)

	if err != nil {
		panic(err)
	}

	return GooseMigrator{
		db:       db,
		fsys:     fsys,
		provider: provider,
		dir:      dir,
	}
}

func (g GooseMigrator) Create() {
	name := petname.Generate(1, "")
	goose.SetBaseFS(g.fsys)
	err := goose.Create(g.db, g.dir, name, "sql")

	if err != nil {
		panic(err)
	}
}

func (g GooseMigrator) CreateWithStatement(up string, down string) {
	name := petname.Generate(1, "")
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
	res, err := g.provider.Up(context.Background())
	if err != nil {
		panic(err)
	}

	log.Printf("%v", res)
}

func (g GooseMigrator) Down() {
	res, err := g.provider.Down(context.Background())
	if err != nil {
		panic(err)
	}

	log.Printf("%v", res)
}

// Create writes a new blank migration file.
