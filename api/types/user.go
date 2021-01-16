package types

import (
	"errors"

	"gorm.io/gorm"
)

//User : Struct for a user (to access to the API)
type User struct {
	ID       int      `gorm:"primaryKey"`
	Email    string   `gorm:"not null;"`
	Username string   `gorm:"not null;"`
	Password string   `gorm:"not null;"`
	Token    string   `gorm:"not null;"`
	IsAdmin  bool     `gorm:"not null;"`
	Domains  []Domain `gorm:"-"` //Dont save this in the DB.
}

//IsOwner : Check if user owns domain
func (u User) IsOwner(d Domain) bool {
	if d.OwnerID != u.ID && !u.IsAdmin {
		return false
	}
	return true
}

//GetUser : get user from gorm database (by id)
func (u *User) GetUser(db *gorm.DB) error {
	result := db.First(&u, u.ID)
	return result.Error
}

//GetUserByUsername : get user from gorm database (by username)
func (u *User) GetUserByUsername(db *gorm.DB) error {
	result := db.Where("username = ?", u.Username).First(&u)
	return result.Error
}

//CreateUser : create user in gorm database
func (u *User) CreateUser(db *gorm.DB) error {
	if u.EmailExists(db) || u.UsernameExists(db) {
		return gorm.ErrRegistered
	}
	result := db.Create(&u)
	return result.Error
}

//UpdateUser : update user from gorm database (by id)
func (u *User) UpdateUser(db *gorm.DB) error {
	result := db.Save(&u)
	return result.Error
}

//DeleteUser : delete user from gorm database (by id)
func (u *User) DeleteUser(db *gorm.DB) error {
	result := db.Delete(&u)
	return result.Error
}

//EmailExists : check if user with the same EMAIL already exists
func (u *User) EmailExists(db *gorm.DB) bool {
	result := db.Where("email = ?", u.Email).First(&u)
	return !errors.Is(result.Error, gorm.ErrRecordNotFound)
}

//UsernameExists : check if user with the same USERNAME already exists
func (u *User) UsernameExists(db *gorm.DB) bool {
	result := db.Where("username = ?", u.Username).First(&u)
	return !errors.Is(result.Error, gorm.ErrRecordNotFound)
}
