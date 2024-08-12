package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/raokrutarth/golang-playspace/pkg/logger"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// GormDB The database object that can be used by middleware to get data
var GormDB *gorm.DB

type Message struct {
	gorm.Model
	MessageID       string `gorm:"unique"`
	UID             uint32
	SeqNum          uint32
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

// override table name for gorm
func (Message) TableName() string {
	return "outlookcleaner_messages"
}

// SetupDatabase - Connects the database
func SetupDatabase(ctx context.Context) error {
	dbConfig := getConfig(ctx).Database
	l := logger.GetLoggerFromContext(ctx)
	connectionString := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		dbConfig.Hostname,
		dbConfig.Port,
		dbConfig.User,
		dbConfig.Database,
		dbConfig.Password,
	)
	l.Info("connecting to database", "hostname", dbConfig.Hostname, "port",
		dbConfig.Port, "user", dbConfig.User, "database", dbConfig.Database,
		"pwdLen", len(dbConfig.Password),
	)
	db, errConnect := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if errConnect != nil {
		return errConnect
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Error("failed to get sql db", "error", err)
		return err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)
	GormDB = db
	return nil
}

func initDB(ctx context.Context) error {
	if err := SetupDatabase(ctx); err != nil {
		return err
	}
	logger.GetLoggerFromContext(ctx).Info("running auto migrations")
	err := GormDB.AutoMigrate(Message{})
	if err != nil {
		return fmt.Errorf("failed to migrate database with error %w", err)
	}
	return nil
}
