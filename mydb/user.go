package mydb

import (
	"fmt"

	"github.com/lastforeverzl/barkme/message"

	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	DeviceName string  `json:"deviceName"`
	Latitude   float32 `json:"latitude"`
	Longitude  float32 `json:"longitude"`
	Barks      int32   `json:"barks"`
	Favorites  []*User `gorm:"many2many:friendships;association_jointable_foreignkey:friend_id"`
}

type AllUsers struct {
	Users []User
	Err   error
}

type UserChan struct {
	User User
	Err  error
}

func (db *DB) GetAllUsers(c chan *AllUsers) {
	users := make([]User, 0)
	if err := db.Find(&users).Error; err != nil {
		c <- &AllUsers{Err: err}
	}
	c <- &AllUsers{Users: users}
	close(c)
}

func (db *DB) CreateUser(c chan *UserChan) {
	user := User{}
	if err := db.Create(&user).Error; err != nil {
		c <- &UserChan{Err: err}
	}
	c <- &UserChan{User: user}
	close(c)
}

func (db *DB) UpdateUser(c chan *UserChan, id string, userUpdate User) {
	user := User{}
	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		c <- &UserChan{Err: err}
	}
	if err := db.Model(&user).Updates(userUpdate).Error; err != nil {
		c <- &UserChan{Err: err}
	}
	c <- &UserChan{User: user}
	close(c)
}

func (db *DB) AddFavUser(c chan *UserChan, id string, favoriteUser User) {
	user := User{}
	favUser := User{}
	db.Preload("Favorites").First(&user, "id = ?", id)
	if err := db.Where("id = ?", favoriteUser.ID).First(&favUser).Error; err != nil {
		c <- &UserChan{Err: err}
	}
	if err := db.Model(&user).Association("Favorites").Append(favUser).Error; err != nil {
		c <- &UserChan{Err: err}
	}
	c <- &UserChan{User: user}
	close(c)
}

func (db *DB) RemoveFavUser(c chan *UserChan, id string, rmFavUser User) {
	user := User{}
	rmUser := User{}
	db.Preload("Favorites").First(&user, "id = ?", id)
	if err := db.Where("id = ?", rmFavUser.ID).First(&rmUser).Error; err != nil {
		c <- &UserChan{Err: err}
	}
	if err := db.Model(&user).Association("Favorites").Delete(rmUser).Error; err != nil {
		c <- &UserChan{Err: err}
	}
	c <- &UserChan{User: user}
	close(c)
}

func (user *User) AfterCreate(tx *gorm.DB) (err error) {
	tx.Model(user).Update("DeviceName", fmt.Sprintf("#%d", user.ID))
	return
}

func (db *DB) UpdateUserAction(msg message.Envelope) {
	user := User{}
	db.Where("id = ?", msg.ID).First(&user)
	switch msg.Msg {
	case "barks":
		user.Barks++
	}
	db.Save(&user)
	fmt.Println(user)
}
