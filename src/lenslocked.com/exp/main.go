package main

// this is what we used initially
// _ "github.com/lib/pq"
import (
	"bufio"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

const (
	host   = "localhost"
	port   = 5432
	user   = "eitah"
	dbname = "lenslocked_itah"
)

type User struct {
	gorm.Model
	Name  string
	Email string `gorm:"not null;unique_index"`
}

func main() {
	// db := openSQLConnection()

	db := openGormConnection()
	defer db.Close()

	// With logging enabled we will start to see output like the following when SQL statements are run.
	// CREATE TABLE "users"...
	// db.LogMode(true)

	// getAUser(db)
	// getAMinUser(db)
	searchByUserExample(db)

	// AutoMigrate will only create things that dont already exists, so if you already had a table named users it would not delete that table and attempt to make a new one. Likewise, it will not delete a column or replace it with a new type as these both have the potential to delete data unintentionally. Instead you will need to handle those types of migrations on your own.
	// db.AutoMigrate(&User{})

	// seedUsers(db)

	// populateSingleRowData(db)
	// populateOrderData(db)
	// readSingleRowOfData(db)
	// readMultipleRowsOfData(db)
	// joinMultipleTables(db)

	// dropAllTables(db)
}

func getAUser(db *gorm.DB) {
	var u User
	db.First(&u)
	if db.Error != nil {
		panic(db.Error)
	}
	fmt.Println(u)
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

func openGormConnection() *gorm.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, port, user, dbname)

	db, err := gorm.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	return db
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

func openSQLConnection() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, port, user, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	fmt.Println("Successfully Connected")
	return db
}

func readSingleRowOfData(db *sql.DB) {
	var id int
	var name, email string
	row := db.QueryRow(`
	Select id, name, email
	FROM users_pk
	WHERE id=$1
	`, 1)
	must(row.Scan(&id, &name, &email))
	fmt.Printf("ID %d. Name %s, email %s\n", id, name, email)
}

func populateSingleRowData(db *sql.DB) {
	var id int
	row := db.QueryRow(`
	INSERT INTO users_pk(name, email)
	VALUES($1, $2) RETURNING id`,
		"Jon Calhoun", "jon@calhoun.io")

	err := row.Scan(&id)
	if err != nil {
		panic(err)
	}
	fmt.Printf("inserted %d\n", id)
	db.Close()
}

func readMultipleRowsOfData(db *sql.DB) {
	var id int
	var name, email string
	rows, err := db.Query(`
SELECT id,name,email
FROM users_pk
WHERE email = $1
OR ID > $2`,
		"jon#calhoun.io", 2)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		rows.Scan(&id, &name, &email)
		fmt.Printf("ID %d. Name %s, email %s\n", id, name, email)
	}
	db.Close()
}

func populateOrderData(db *sql.DB) {
	var id int
	for i := 1; i < 6; i++ {
		// Create some fake data
		userId := 1
		if i > 3 {
			userId = 2
		}
		amount := 1000 * i
		description := fmt.Sprintf("USB-C Adapter x%d", i)
		err := db.QueryRow(`
INSERT INTO orders (user_id, amount, description) VALUES ($1, $2, $3)
RETURNING id`,
			userId, amount, description).Scan(&id)
		if err != nil {
			panic(err)
		}
		fmt.Println("Created an order with the ID:", id)
	}
	db.Close()
}

func joinMultipleTables(db *sql.DB) {
	var id, order_id int
	var name, email, order_amount, order_description string
	rows, err := db.Query(`
	SELECT u.id, u.email, u.name, orders.id AS order_id,
	orders.amount AS order_amount,
	orders.description AS order_description
	FROM users_pk u
	INNER JOIN orders
	ON u.id = orders.user_id;`)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		// eli found that the scan doesnt print unless teh same number of columns that you select are what you assign which is neat
		rows.Scan(&id, &email, &name, &order_id, &order_amount, &order_description)
		fmt.Printf("ID %d, Name %s, email %s, oid %d, order_amount %s, order_description %s\n", id, name, email, order_id, order_amount, order_description)
	}
	db.Close()
}

func dropAllTables(db *sql.DB) {
	_, err := db.Exec("Drop table orders;")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("Drop table users_pk;")
	if err != nil {
		panic(err)
	}
	fmt.Println("Dropped all tables goodbye")
	db.Close()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
