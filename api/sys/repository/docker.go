package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/fortytw2/dockertest"
)

func NewDockerRepo() (*RDBMSRepository, func(), error) {
	cfg := Config{
		Driver: "postgres",
	}
	var db *sql.DB
	container, runErr := dockertest.RunContainer("postgres:alpine", "5432", func(addr string) error {
		hostPort := strings.Split(addr, ":")
		if len(hostPort) != 2 {
			return errors.New("wrong addr format")
		}

		port, err := strconv.Atoi(hostPort[1])
		if err != nil {
			return err
		}

		cfg.DSN = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			"postgres", "postgres", hostPort[0], port, "postgres")

		db, err = connect(cfg)
		return err
	})
	if runErr != nil {
		return nil, func() {}, runErr
	}

	repo := &RDBMSRepository{
		Name: "crawler",
		db:   db,
		Cfg:  cfg,
	}

	closer := func() {
		_ = db.Close()
		container.Shutdown()
	}

	if err := repo.MigrateUp(); err != nil {
		return nil, closer, err
	}

	if err := repo.Init(); err != nil {
		return nil, closer, err
	}

	return repo, closer, nil
}
