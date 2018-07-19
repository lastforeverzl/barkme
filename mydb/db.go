package mydb

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lastforeverzl/barkme/message"
)

type Datastore interface {
	GetAllUsers(chan *AllUsers)
	CreateUser(chan *UserChan)
	UpdateUser(chan *UserChan, string, User)
	AddFavUser(chan *UserChan, string, User)
	RemoveFavUser(chan *UserChan, string, User)
	UpdateUserAction(message.Envelope)
}

type DB struct {
	*gorm.DB
}

type DbConfiguration struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func NewDB(filename string) (*DB, error) {
	file, _ := os.Open(filename)
	defer file.Close()
	decoder := json.NewDecoder(file)
	cfg := DbConfiguration{}
	err := decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	dbInfo := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		cfg.User, cfg.Password, cfg.Database, cfg.Host, cfg.Port)
	db, err := gorm.Open("postgres", dbInfo)
	if err != nil {
		return nil, err
	}
	if err = db.DB().Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) InitSchema() {
	db.AutoMigrate(&User{})
}
