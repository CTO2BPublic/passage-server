package dbdriver

import (
	"fmt"

	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/models"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

type Database struct {
	Engine *gorm.DB
}

var Driver = new(Database)
var Config = config.GetConfig()
var Tracer = otel.Tracer("pkg/dbdriver")

func (d *Database) Connect() {

	var err error

	log.Info().Msg("Starting dbdriver")

	// PSQL engine
	if Config.Db.Engine == "psql" {

		dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s search_path=%s", Config.Db.Psql.Host, Config.Db.Psql.Port, Config.Db.Psql.Username, Config.Db.Psql.Database, Config.Db.Psql.Password, Config.Db.Psql.SSLMode, Config.Db.Psql.Schema)

		d.Engine, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	// MySQL engine
	if Config.Db.Engine == "mysql" {

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", Config.Db.Mysql.Username, Config.Db.Mysql.Password, Config.Db.Mysql.Host, Config.Db.Mysql.Port, Config.Db.Mysql.Database)

		d.Engine, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	// SQLite engine
	if Config.Db.Engine == "sqlite" {

		d.Engine, err = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	d.Engine.Use(tracing.NewPlugin())

	Driver = d
}

func (d *Database) AutoMigrate() {

	log.Info().Msg("Starting db migrations")
	err := d.Engine.AutoMigrate(
		models.AccessRequest{},
		models.AccessRole{},
		models.UserProfile{},
		models.Event{},
	)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	log.Info().Msg("Completed db migrations")
}

func GetDriver() *Database {
	return Driver
}
