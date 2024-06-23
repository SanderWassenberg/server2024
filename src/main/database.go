package main

import (
	"errors"
	"log"
	"strings"
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
`)
// NOTE: this cannot be a strict table because then reading dates becomes more of a hassle.
// datetime isn't actually a datatype (it's actually converts to TEXT), strict tables don't allow it.
// however, the sqlite driver only allows scanning a date if the table has DATETIME as the type.
// the alternative is to have the field be TEXT and always parse the strings manually after scanning them as strings
// So you either have nice scanning and no strictness, or strictness but annoying scanning.
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
		result, err := db.Exec(`
CREATE UNIQUE INDEX "idx_Username" ON "Users" ("Username");
`)
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
	var hash string
	err := db.QueryRow("SELECT PasswordHash FROM Users WHERE Username = ?", name).Scan(&hash)
	if err != nil { log.Printf("check_login: %v", err); return false }

	return pwhash.VerifyPassword(password, hash)
}

func set_session(name, token string, exp time.Time) error {
	_, err := db.Exec("UPDATE Users SET SessionToken = ?, SessionExpirationDate = ? WHERE Username = ?;", token, exp, name)
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
	err = db.QueryRow("SELECT Admin FROM Users WHERE Username = ?", name).Scan(&isadmin)
	return
}

var ErrSessionExpired = errors.New("session expired")

func get_name_from_session(session_token string) (name string, err error) {
	var exp time.Time
	err = db.QueryRow("SELECT Username, SessionExpirationDate FROM Users WHERE SessionToken = ?", session_token).Scan(&name, &exp)
	if err != nil { return }

	if time.Now().After(exp) {
		err = ErrSessionExpired
	}

	return
}

func get_interest(name string) (interest string, err error) {
	err = db.QueryRow("SELECT Interest FROM Users WHERE Username = ?", name).Scan(&interest)
	return
}

type SearchResultRow struct {
	Name     string `json:"name"`
	Interest string `json:"interest"`
}
func search_by_interest(match, name string, limit int) ([]SearchResultRow, error) {
	// yes this is slow, idgaf, it's a school project, it can suck my ass
	match = strings.ReplaceAll(match, `\`, `\\`)
	match = strings.ReplaceAll(match, `%`, `\%`)
	match = strings.ReplaceAll(match, `_`, `\_`)
	match = "%" + match + "%"

	rows, err := db.Query(`
SELECT Username, Interest
FROM Users
WHERE Banned != TRUE
	AND Interest LIKE ? ESCAPE '\'
	AND Username != ?
ORDER BY Username ASC
LIMIT ?;
`, match, name, limit)
	if err != nil {
		log.Printf("search_by_interest: %v", err)
		return nil, err
	}
	defer rows.Close()

	results := make([]SearchResultRow, 0, 10)

	for rows.Next() {
		var name, interest string
		err := rows.Scan(&name, &interest)
		if err != nil {
			log.Printf("search_by_interest row scan error: %v", err)
			return nil, err
		}
		results = append(results, SearchResultRow{
			Name:     name,
			Interest: interest,
		})
	}

	return results, nil
}