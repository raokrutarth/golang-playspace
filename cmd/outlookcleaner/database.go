package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// GormDB The database object that can be used by middleware to get data
var GormDB *gorm.DB

type Message struct {
	gorm.Model
	MessageID       string `gorm:"unique"`
	From            string `gorm:"size:255,index"`
	FromName        string
	To              string `gorm:"size:255"`
	Body            string
	Subject         string
	ReceivedAt      time.Time
	RemoteDeletedAt time.Time
	OpenedAt        sql.NullTime
	MailBoxFolder   string
	SizeBytes       uint32
	IsSeen          bool
	IsFlagged       bool
	IsReceipt       bool
	AttachmentNames string
	Attributes      datatypes.JSON
}

// SetupDatabase - Connects the database
func SetupDatabase(logMode bool) error {
	DbConfig := getConfig().Database
	connectionString := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		DbConfig.Hostname,
		DbConfig.Port,
		DbConfig.User,
		DbConfig.Database,
		DbConfig.Password,
	)
	log.Info().Str("hostname", DbConfig.Hostname).Msgf("DB Connected")
	db, errConnect := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if errConnect != nil {
		return errConnect
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to ")
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)

	GormDB = db
	return nil
}

func init() {
	if err := SetupDatabase(true); err != nil {
		log.Fatal().Err(err).Msg("DB connection init failed with error")
	}
	log.Info().Msg("Running auto migrations.")
	GormDB.AutoMigrate(Message{})
}
