package datastore

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	LayoutISO      = "2006-01-02T15:04:05+07:00"
	LayoutTime     = "15:04"
	LayoutHour     = "15"
	LayoutMinute   = "04"
	LayoutTanggal  = "2 January 2006"
	LayoutSQL      = "2006-01-02"
	LayoutFullSQL  = "2006-01-02 15:04:05"
	LayoutFullUser = "2 January 2006 15:04"
)

var Db *gorm.DB

func NewDB() {

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // Disable color
		},
	)

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "root:@tcp(127.0.0.1:3306)/spera?charset=utf8&parseTime=True&loc=Local", // data source name
		DefaultStringSize:         256,                                                                     // default size for string fields
		DisableDatetimePrecision:  true,                                                                    // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,                                                                    // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,                                                                    // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false,                                                                   // auto configure based on currently MySQL version
	}), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		panic(err)
	}
	Db = db
}
