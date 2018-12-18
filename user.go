package main

import (
	"github.com/jinzhu/gorm"
	"strconv"
)

type User struct {
	gorm.Model
	Name string
}

func CreateUser(fields map[string]interface{}) (*User, error) {
	return nil, nil
}

func GetUser(id int32) (*User, error) {
	var user User
	DB.First(&user, "id = ?", id)
	return &user, nil
}

func (user *User) Update(changes map[string]interface{}) error {
	return nil
}

func (user *User) Delete() error {
	return nil
}

func (user *User) String() string {
	return strconv.Itoa(int(user.Model.ID)) + user.Name
}
