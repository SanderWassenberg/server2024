package main

import (
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
	/*
	statement, err := db.Exec(`
CREATE TABLE "Users" (
	"Id"	INTEGER,
	"Username"	TEXT NOT NULL UNIQUE,
	"PasswordHash"	TEXT NOT NULL,
	"SessionToken"	TEXT NOT NULL DEFAULT "",
	"SessionExpirationDate"	TEXT NOT NULL DEFAULT "0001-01-01 00:00:00+00:00",
	"Interest"	TEXT NOT NULL DEFAULT "",
	PRIMARY KEY("Id" AUTOINCREMENT)
) STRICT`)

	_ = statement
	if err != nil {
		log.Println("exec failed:", err)
	} else {
		log.Println("executed alright...")
	}
	*/
}

func deinit_db() {
	log.Print("Closing db...")
	if err := db.Close(); err != nil {
		log.Printf("Error closing db: %v", err)
	} else {
		log.Print("Successfully closed db.")
	}
}

func create_user(name, password string, ignore_fail bool) (int64, error) {
	hash, err := pwhash.EncodePassword(password)
	if err != nil { log.Println("ERR: failed to generate encoded hash:", err); return 0, err }

	var stmt string
	if ignore_fail {
		stmt = "INSERT OR IGNORE INTO Users (Username, PasswordHash) VALUES(?, ?)"
	} else {
		stmt =           "INSERT INTO Users (Username, PasswordHash) VALUES(?, ?)"
	}

	result, err := db.Exec(stmt, name, hash)
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


type Session struct {
	Token          string
	ExpirationDate string
}

func get_session(name string) (session Session, err error) {
	rows, err := db.Query("SELECT SessionToken, SessionExpirationDate FROM Users WHERE Username = ?", name)
	if err != nil { return session, err }
	defer rows.Close()

	rows.Next()
	// TODO: how to read time back and maybe it is null, how to deal with that?
	err = rows.Scan(&session.Token, &session.ExpirationDate)
	return session, err
}

func set_session(name, token string, exp time.Time) error {
	// it is important that Id not be in quotes here
	_, err := db.Exec("UPDATE Users SET SessionToken = ?, SessionExpirationDate = ? WHERE Username = ?;", token, exp, name)
	if err != nil { return err }
	return nil
}