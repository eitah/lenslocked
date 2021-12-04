package exp

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

const (
	host   = "localhost"
	port   = 5432
	user   = "eitah"
	dbname = "lenslocked_itah" // this is not the prod db
)

type User struct {
	gorm.Model
	Name   string
	Email  string `gorm:"not null;unique_index"`
	Orders []Order
}

type Order struct {
	gorm.Model
	UserID      uint
	Amount      int
	Description string
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, port, user, dbname)

	db, err := gorm.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// With logging enabled we will start to see output like the following when SQL statements are run.
	// CREATE TABLE "users"...
	db.LogMode(true)

	// AutoMigrate will only create things that dont already exists, so if you already had a table named users it would not delete that table and attempt to make a new one. Likewise, it will not delete a column or replace it with a new type as these both have the potential to delete data unintentionally. Instead you will need to handle those types of migrations on your own.
	db.AutoMigrate(&User{}, &Order{})

	getAUser(db)
	getAMinUser(db)
	searchByUserExample(db)
	searchForMultipleUsers(db)
	seedUsers(db)
	seedOrders(db)
	preloadOrders(db)

}

func getAUser(db *gorm.DB) {
	var u User
	db.First(&u)
	if db.Error != nil {
		panic(db.Error)
	}
	fmt.Println(u)
}

func preloadOrders(db *gorm.DB) {
	var user User
	// Because user has an orders array and because we tell it to be preloaded it all just works.
	db.Preload("Orders").First(&user)
	if db.Error != nil {
		panic(db.Error)
	}
	fmt.Println("Email", user.Email)
	fmt.Println("Number of orders", len(user.Orders))
	fmt.Println("Orders", user.Orders)
}

func getAMinUser(db *gorm.DB) {
	var u User
	maxId := 3
	db.Where("id <= ?", maxId).First(&u)
	if db.Error != nil {
		panic(db.Error)
	}
	fmt.Println(u)
}

func seedUsers(db *gorm.DB) {
	name, email := getInfo()
	u := &User{
		Name:  name,
		Email: email,
	}
	must(db.Create(u).Error)
	fmt.Printf("%+v\n", u)
}

func seedOrders(db *gorm.DB) {
	var user User
	db.First(&user)
	if db.Error != nil {
		panic(db.Error)
	}
	createOrder(db, user, 1001, "Fake Description #1")
	createOrder(db, user, 9999, "Fake Description #2")
	createOrder(db, user, 8800, "Fake Description #3")

}

func createOrder(db *gorm.DB, user User, amount int, desc string) {
	db.Create(&Order{
		UserID: user.ID, Amount: amount, Description: desc,
	})
	if db.Error != nil {
		panic(db.Error)
	}
}

func getInfo() (name, email string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("What is your name?")
	name, _ = reader.ReadString('\n')
	name = strings.TrimSuffix(name, "\n")
	emailHosts := []string{"yahoo", "gmail", "hotmail", "comcast", "buzzmail", "tmail", "bmail"}
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	myRand := r1.Intn(6)
	fmt.Println(myRand)
	myHost := emailHosts[myRand]
	fillerNumber := strconv.Itoa(r1.Intn(1000))
	email = name + fillerNumber + "@" + myHost + ".com"
	return name, email
}

func searchByUserExample(db *gorm.DB) {
	validU := User{
		Name: "shera",
	}
	invalidU := User{
		Name: "notARealUser",
	}
	for _, u := range []User{validU, invalidU} {
		db.Where(u).First(&u)
		if db.Error != nil {
			panic(db.Error)
		}
		var isReal bool
		if u.ID != 0 {
			isReal = true
		}
		if isReal {
			fmt.Printf("User %s found %+v\n", u.Name, u)
		} else {
			fmt.Printf("User %s not found\n", u.Name)
		}
	}
}

func searchForMultipleUsers(db *gorm.DB) {
	var users []User
	db.Find(&users)
	if db.Error != nil {
		panic(db.Error)
	}

	// this found all users, or you can filter by passing in users.
	fmt.Printf("Retrieved %d users.", len(users))
	fmt.Println(users)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
