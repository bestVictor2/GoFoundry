package main

import (
	"GoGorm/gorm"
	"GoGorm/session"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type User struct {
	Name string `foundry:"PRIMARY KEY"`
	Age  int
}

func main() {
	engine, err := gorm.NewEngine("sqlite3", "gee.db")
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Close()

	s := engine.NewSession().Model(&User{})
	_ = s.DropTable()
	_ = s.AutoMigrate()
	if _, err := s.Insert(&User{Name: "Tom", Age: 18}, &User{Name: "Sam", Age: 21}); err != nil {
		log.Fatal(err)
	}

	if err := engine.Transaction(func(tx *session.Session) error {
		tx.Model(&User{})
		_, err := tx.Insert(&User{Name: "Alice", Age: 26})
		return err
	}); err != nil {
		log.Fatal(err)
	}

	var users []User
	if err := s.Select("Name", "Age").OrderBy("Age DESC").Find(&users); err != nil {
		log.Fatal(err)
	}
	fmt.Println("query users:")
	for _, user := range users {
		fmt.Printf("- %s (%d)\n", user.Name, user.Age)
	}
}
