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



// Types

type OtpInfo struct {
	enabled bool
	secret  string
}

type SearchResultRow struct {
	Username string `json:"username"`
	Interest string `json:"interest"`
}

type MessageHistory struct {
	Message []string `json:"message"`
	Side    string   `json:"side"` // string of 1's and 0's. 1 means the message at this index was sent by ourselves, 0 means it was sent by the other user.
}



// Global vars

var db *sql.DB

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrSessionExpired    = errors.New("session expired")
var ErrSessionNotFound   = errors.New("session not found")



// Functions

func init_db() error {
	var err error
	db, err = sql.Open("sqlite3", "showcase.sqlite")
	if err != nil { return err }

	init_exec("create Users table", `
CREATE TABLE "Users" (
	"Id"                    INTEGER,
	"Username"              TEXT     NOT NULL UNIQUE,
	"PasswordHash"          TEXT     NOT NULL,
	"SessionToken"          TEXT     NOT NULL DEFAULT "",
	"SessionExpirationDate" DATETIME NOT NULL DEFAULT "0001-01-01 00:00:00+00:00",
	"Interest"              TEXT     NOT NULL DEFAULT "",
	"Banned"                INTEGER  NOT NULL DEFAULT 0,
	"Admin"                 INTEGER  NOT NULL DEFAULT 0,
	"OtpEnabled"            INTEGER  NOT NULL DEFAULT 0,
	"OtpSecret"             TEXT     NOT NULL DEFAULT "",
	PRIMARY KEY("Id" AUTOINCREMENT)
);
`)
// NOTE: this table cannot be STRICT because then reading dates becomes more of a hassle. Because...
// DATETIME isn't actually a datatype (it's actually converts to TEXT), so STRICT tables don't allow it.
// However, the sqlite driver only allows scanning a time.Time if the table has DATETIME as the type.
// If you make the table STRICT, then that column must be TEXT and you have to parse the strings back to time.Time manually.
// So you either have nice scanning and no strictness, or strictness but annoying scanning.

	init_exec("create Username index", `
CREATE UNIQUE INDEX "idx_Username" ON "Users" ("Username");
`)

	init_exec("make Messages table", `
CREATE TABLE "Messages" (
	"Text"	TEXT	NOT NULL,
	"To"	INTEGER NOT NULL,
	"From"	INTEGER NOT NULL,
	FOREIGN KEY("To") REFERENCES "Users"("Id")
) STRICT;
`)

	return nil
}

func init_exec(description, sql string) {
	_, err := db.Exec(sql)
	if err != nil {
		log.Println("Could not", description, "because", err)
	} else {
		log.Println("init_db:")
	}
}

func deinit_db() {
	log.Print("Closing db...")
	if err := db.Close(); err != nil {
		log.Println("Error closing db:", err)
	} else {
		log.Print("Successfully closed db.")
	}
}



// Simple queries

// could change the majority of these to take an id instead of name but that's for later. (not all, at login time we don't have the id yet.)
func set_session(name, token string, exp time.Time) error {
	_, err := db.Exec(`UPDATE "Users" SET "SessionToken" = ?, "SessionExpirationDate" = ? WHERE "Username" = ?;`, token, exp, name)
	return err
}
func set_interest(name, interest string) error {
	_, err := db.Exec(`UPDATE "Users" SET "Interest" = ? WHERE "Username" = ?;`, interest, name)
	return err
}
func set_banned(name string, banned bool) error {
	_, err := db.Exec(`UPDATE "Users" SET "Banned" = ? WHERE "Username" = ?;`, banned, name)
	return err
}
func set_otp_enabled(name string, enabled bool) error {
	_, err := db.Exec(`UPDATE "Users" SET "OtpEnabled" = ? WHERE "Username" = ?;`, enabled, name)
	return err
}
func set_otp_secret(name string, secret string) error {
	_, err := db.Exec(`UPDATE "Users" SET "OtpSecret" = ? WHERE "Username" = ?;`, secret, name)
	return err
}
func save_message(my_id, other_id int, text string) (err error) {
	_, err = db.Exec(`INSERT INTO "Messages" ("From", "To", "Text") VALUES (?, ?, ?)`, my_id, other_id, text)
	return
}

func get_otp_info(name string) (info OtpInfo, err error) {
	err = db.QueryRow(`SELECT "OtpSecret", "OtpEnabled" FROM "Users" WHERE "Username" = ?`, name).Scan(&info.secret, &info.enabled)
	return
}
func get_id(name string) (id int, err error) {
	err = db.QueryRow(`SELECT "Id" FROM "Users" WHERE "Username" = ?`, name).Scan(&id)
	return
}
func get_password_hash(name string) (hash string, err error) {
	err = db.QueryRow(`SELECT "PasswordHash" FROM "Users" WHERE "Username" = ?`, name).Scan(&hash)
	return
}



// More complex queries

func create_chatter(name, password string, admin bool) (int, error) {
	{ // check if exists, because generating a password hash is expensive.
		exists := false
		db.QueryRow(`SELECT 1 FROM "Users" WHERE "Username" = ?`, name).Scan(&exists)
		if exists { return 0, ErrUserAlreadyExists }
	}

	hash, err := pwhash.EncodePassword(password)
	if err != nil { log.Println("ERR: failed to generate encoded hash:", err); return 0, err }

	result, err := db.Exec(`INSERT INTO "Users" ("Username", "PasswordHash", "Admin") VALUES (?, ?, ?)`, name, hash, admin)
	if err != nil { log.Println("ERR: create user exec:", err); return 0, err }

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	} else {
		if id != 0 { log.Println("created user with id:", id) }
		return int(id), nil
	}
}

func get_info_from_session(session_token string) (info *UserInfo, err error) {

	var exp      time.Time
	var is_admin bool
	info = &UserInfo{}

	err = db.QueryRow(
		`SELECT "SessionExpirationDate", "Id", "Username", "Interest", "Admin" FROM "Users" WHERE "SessionToken" = ?`,
		session_token,
	).Scan(&exp, &info.id, &info.Name, &info.Interest, &is_admin)
	if err != nil {
		if err == sql.ErrNoRows { err = ErrSessionNotFound }
		return nil, err
	}

	if time.Now().After(exp) {
		return nil, ErrSessionExpired
	}

	if is_admin {
		info.Role = "admin"
	} else {
		info.Role = "chatter"
	}

	return info, nil
}

func search_by_interest(match string, exclude_id int, limit int) ([]SearchResultRow, error) {
	// yes this is slow, idgaf, it's a school project, it can suck my ass
	match = strings.ReplaceAll(match, `\`, `\\`)
	match = strings.ReplaceAll(match, `%`, `\%`)
	match = strings.ReplaceAll(match, `_`, `\_`)
	match = "%" + match + "%"

	rows, err := db.Query(`
SELECT "Username", Interest
FROM "Users"
WHERE "Banned" != TRUE
	AND "Interest" LIKE ? ESCAPE '\'
	AND "Id" != ?
ORDER BY "Username" ASC
LIMIT ?;
`, match, exclude_id, limit)
	if err != nil {
		log.Println("search_by_interest:", err)
		return nil, err
	}
	defer rows.Close()

	results := make([]SearchResultRow, 0, 10)

	for rows.Next() {
		var username, interest string
		err := rows.Scan(&username, &interest)
		if err != nil {
			log.Println("search_by_interest row scan error:", err)
			return nil, err
		}
		results = append(results, SearchResultRow{
			Username: username,
			Interest: interest,
		})
	}

	return results, nil
}

func get_chat_history(id1, id2, limit int) (history *MessageHistory, err error) {
	rows, err := db.Query(`
SELECT "Text", ("From" = ?1) as "FromMe"
FROM "Messages"
WHERE ("From" = ?1 AND "To" = ?2) OR ("From" = ?2 AND "To" = ?1)
ORDER BY ROWID ASC
LIMIT ?3;
`, id1, id2, limit)
	if err != nil {
		log.Println("get_chat_history:", err)
		return nil, err
	}
	defer rows.Close()

	history = &MessageHistory{
		Message: make([]string, 0, 100),
	}
	side := make([]byte, 0, 100)

	for rows.Next() {
		var msg string
		var sender byte
		err := rows.Scan(&msg, &sender)
		if err != nil {
			log.Println("get_chat_history row scan error:", err)
			return nil, err
		}
		history.Message = append(history.Message, msg)

		side = append(side, sender + '0')
	}

	history.Side = string(side)

	return history, nil
}
