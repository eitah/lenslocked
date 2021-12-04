package main

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
	us, err := models.NewUserService(psqlInfo)
	if err != nil {
		panic(err)
	}
	defer us.Close()
	us.DestructiveReset()

	if err := us.Create(&models.User{Name: "fred", Email: "fred@gmail.com"}); err != nil {
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

}
