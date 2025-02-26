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
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	dbutils "github.com/anoaland/xgo/db/utils"
)

type GooseMigrator struct {
	db      *sql.DB
	fsys    fs.FS
	dialect database.Dialect
	dir     string
	logger  goose.Logger
}

func NewMigrator(dialect database.Dialect, dsn string, fsys fs.FS, dir string) GooseMigrator {
	db, err := goose.OpenDBWithDriver(string(dialect), dsn)
	if err != nil {
		panic(err)
	}

	defaultLogger := zerolog.New(zerolog.ConsoleWriter{Out: log.Writer()})

	return GooseMigrator{
		db:     db,
		fsys:   fsys,
		dir:    dir,
		logger: NewZeroLogGooseLogger(&defaultLogger),
	}
}

func (g GooseMigrator) SetLogger(logger *zerolog.Logger) {
	goose.SetLogger(NewZeroLogGooseLogger(logger))
}

func (g GooseMigrator) setup() {
	goose.SetLogger(g.logger)
	goose.SetBaseFS(g.fsys)
}

func (g GooseMigrator) provider() *goose.Provider {
	g.setup()
	provider, err := goose.NewProvider(
		g.dialect,
		g.db,
		g.fsys,
	)

	if err != nil {
		if err == goose.ErrNoMigrations {
			log.Println("No migration found!")
			return nil
		} else {
			panic(err)
		}
	}

	return provider
}

func (g GooseMigrator) generateName() string {
	randomNumber := rand.Intn(3) + 1
	name := petname.Generate(randomNumber, "_")

	return name
}

func (g GooseMigrator) Create() {
	name := g.generateName()
	g.setup()
	err := goose.Create(g.db, g.dir, name, "sql")

	if err != nil {
		panic(err)
	}
}

func (g GooseMigrator) CreateWithStatement(up string, down string) {
	name := g.generateName()
	g.setup()

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
	provider := g.provider()
	if provider == nil {
		return
	}

	res, err := provider.Up(context.Background())
	if err != nil {
		panic(err)
	}

	log.Printf("%v", res)
}

func (g GooseMigrator) Down() {
	provider := g.provider()
	if provider == nil {
		return
	}

	res, err := provider.Down(context.Background())
	if err != nil {
		g.logger.Fatalf("goose: %v", err)
		panic(err)
	}

	log.Printf("%v", res)
}
