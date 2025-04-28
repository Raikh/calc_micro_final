package app

import (
	"github.com/jmoiron/sqlx"
	"github.com/raikh/calc_micro_final/internal/config"
)

type App struct {
	Cfg *config.Config
	DB  *sqlx.DB
}
