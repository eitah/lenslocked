package usexp

// package main

// this is what we used initially
// _ "github.com/lib/pq"
import (
	"fmt"

	"github.com/eitah/lenslocked/src/lenslocked.com/models"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

const (
	host   = "localhost"
	port   = 5432
	user   = "eitah"
	dbname = "lenslocked_dev" // this is the dev db
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, port, user, dbname)
	services, err := models.NewServices(psqlInfo)
	if err != nil {
		panic(err)
	}
	defer services.Close()
	services.DestructiveReset()

	us := services.User

	if err := us.Create(&models.User{Name: "fred", Email: "fred@gmail.com", Age: 32}); err != nil {
		panic(err)
	}
	if err := us.Create(&models.User{Name: "larry", Email: "larry@gmail.com", Age: 14}); err != nil {
		panic(err)
	}

	if err := us.Create(&models.User{Name: "barry", Email: "barry@gmail.com", Age: 24}); err != nil {
		panic(err)
	}

	if err := us.Create(&models.User{Name: "harry", Email: "harry@gmail.com", Age: 34}); err != nil {
		panic(err)
	}
	user, err := us.ByEmail("fred@gmail.com")
	if err != nil {
		panic(err)
	}

	user.Name = "Joe"
	if err := us.Update(user); err != nil {
		panic(err)
	}

	_, err = us.ByEmail("fred@gmail.com")
	if err != nil {
		panic(err)
	}

	if err := us.Delete(user.ID); err != nil {
		panic(err)
	}
	_, err = us.ByEmail("fred@gmail.com")

	fmt.Printf("as expected error is %s\n", err)

	var foundUser *models.User
	if foundUser, err = us.ByAge(14); err != nil {
		panic(err)
	}

	fmt.Printf("Found user is %s\n", foundUser.Name)

	var users []*models.User
	var min, max uint = 10, 40
	if users, err = us.InAgeRange(min, max); err != nil {
		panic(err)
	}
	if len(users) == 0 {
		fmt.Printf("no users found in age range %d to %d\n", min, max)
	} else {
		fmt.Printf("%d users found in age range %d to %d\n", len(users), min, max)

	}
	for idx, u := range users {
		fmt.Printf("%d user %s\n", idx+1, u.Name)
	}

}
