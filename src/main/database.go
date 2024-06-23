package main

import (
	"errors"
	"log"
	"time"

	"src/pwhash"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	// we don't use any exports of this package, but importing it here makes it run the package init func,
	// which adds sqlite3 as a driver for database/sql to use
)

var db *sql.DB

func init_db() {
	var err error
	db, err = sql.Open("sqlite3", "showcase.sqlite")
	if err != nil {
		panic(err)
	}

	// Don't use prepared statements for things that modify tables that may be removed.
	// To prepare the statement, it checks if the table it modifies exists, if it doesn't the statment fails to be prepared.

	{
		result, err := db.Exec(`
CREATE TABLE "Users" (
	"Id"					INTEGER,
	"Username"				TEXT NOT NULL UNIQUE,
	"PasswordHash"			TEXT NOT NULL,
	"SessionToken"			TEXT NOT NULL DEFAULT "",
	"SessionExpirationDate"	DATETIME NOT NULL DEFAULT "0001-01-01 00:00:00+00:00",
	"Interest"				TEXT NOT NULL DEFAULT "",
	"Banned"				INTEGER NOT NULL DEFAULT 0,
	"Admin"					INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY("Id" AUTOINCREMENT)
);
`) // I'd li
		_ = result
		if err != nil {
			log.Println("Failed to create Users table:", err)
		} else {
			log.Println("Created Users table.")
		}
	}

	{
		result, err := db.Exec(`
CREATE TABLE "Messages" (
	"Text"	TEXT	NOT NULL,
	"To"	INTEGER NOT NULL,
	"From"	INTEGER NOT NULL,
	"Time"	TEXT	NOT NULL DEFAULT 'datetime("now", "subsec")',
	FOREIGN KEY("To") REFERENCES "Users"("Id")
) STRICT;
`)
		_ = result
		if err != nil {
			log.Println("Failed to create Messages table:", err)
		} else {
			log.Println("Created Messages table.")
		}
	}

	{
		result, err := db.Exec(`CREATE UNIQUE INDEX "idx_Username" ON "Users" ("Username")`)
		_ = result
		if err != nil {
			log.Println("Failed to create Username index:", err)
		} else {
			log.Println("Created Username index.")
		}
	}
	// */
}

func deinit_db() {
	log.Print("Closing db...")
	if err := db.Close(); err != nil {
		log.Printf("Error closing db: %v", err)
	} else {
		log.Print("Successfully closed db.")
	}
}

var ErrUserAlreadyExists = errors.New("user already exists")

func create_chatter(name, password string, admin bool) (int64, error) {
	{
		exists := false
		db.QueryRow("SELECT 1 FROM Users where Username = ?", name).Scan(&exists)
		if exists { return 0, ErrUserAlreadyExists }
	}

	hash, err := pwhash.EncodePassword(password)
	if err != nil { log.Println("ERR: failed to generate encoded hash:", err); return 0, err }

	result, err := db.Exec("INSERT INTO Users (Username, PasswordHash, Admin) VALUES(?, ?, ?)", name, hash, admin)
	if err != nil { log.Println("ERR: create user exec:", err); return 0, err }

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	} else {
		if id != 0 { log.Println("created user with id:", id) }
		return id, nil
	}
}

func check_login(name, password string) bool {
	rows, err := db.Query("SELECT PasswordHash FROM Users WHERE Username = ?", name)
	if err != nil { log.Print(err); return false }
	defer rows.Close()

	var hash string
	rows.Next()
	if err := rows.Scan(&hash); err != nil { log.Print(err); return false }

	return pwhash.VerifyPassword(password, hash)
}


// type Session struct {
// 	Token          string
// 	ExpirationDate string
// }

// func get_session(name string) (session Session, err error) {
// 	rows, err := db.Query("SELECT SessionToken, SessionExpirationDate FROM Users WHERE Username = ?", name)
// 	if err != nil { return session, err }
// 	defer rows.Close()

// 	rows.Next()
// 	err = rows.Scan(&session.Token, &session.ExpirationDate)
// 	return session, err
// }

func set_session(name, token string, exp time.Time) error {
	_, err := db.Exec("UPDATE Users SET SessionToken = ?, SessionExpirationDate = ? WHERE Username = ?;", token, exp.Format(time.RFC3339), name)
	return err
}
func set_interest(name, interest string) error {
	_, err := db.Exec("UPDATE Users SET Interest = ? WHERE Username = ?;", interest, name)
	return err
}
func set_banned(name string, banned bool) error {
	_, err := db.Exec("UPDATE Users SET Banned = ? WHERE Username = ?;", banned, name)
	return err
}
func is_admin(name string) (isadmin bool, err error) {
	rows, err := db.Query("SELECT Admin FROM Users WHERE Username = ?", name)
	if err != nil { return false, err }
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&isadmin)
	return isadmin, err
}

var ErrSessionExpired  = errors.New("session expired")
var ErrSessionNotFound = errors.New("session not found")

func get_name_from_session(session_token string) (name string, err error) {
	rows, err := db.Query("SELECT Username, SessionExpirationDate FROM Users WHERE SessionToken = ?", session_token)
	if err != nil { return "", err }
	defer rows.Close()

	var exp time.Time

	if !rows.Next() {
		return "", ErrSessionNotFound
	}
	err = rows.Scan(&name, &exp) // if this isn't able to scan the date then ensure the table column has type DATETIME (doesn't work on strict tables)
								 // and all fields are filled in with proper dates (not null and proper default values)
	if err != nil { return "", err }

	if time.Now().After(exp) {
		err = ErrSessionExpired
	}

	return name, err
}