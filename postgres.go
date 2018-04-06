package main

import "database/sql"
import "fmt"
import "log"
import _ "github.com/lib/pq"

type PostgresClient struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
	Db       *sql.DB
}

func (p *PostgresClient) GetDB() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.User, p.Password, p.Dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return db
}

type Lore struct {
	UserID  string
	Message string
}

func (p *PostgresClient) RecentLore() []Lore {
	sqlStatement := `
    SELECT user_id, message FROM lores ORDER BY timestamp_added DESC LIMIT 3`
	rows, err := p.Db.Query(sqlStatement)
	defer rows.Close()
	if err != nil {
		panic(err)
	}
	ret := make([]Lore, 0)
	var (
		userId  string
		message string
	)
	for rows.Next() {
		if err := rows.Scan(&userId, &message); err != nil {
			log.Fatal(err)
		}
		lore := Lore{UserID: userId, Message: message}
		ret = append(ret, lore)

	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return ret
}

func (p *PostgresClient) LoreForUser(userId string) []Lore {
	sqlStatement := `
    SELECT message FROM lores WHERE user_id IN ($1)`
	rows, err := p.Db.Query(sqlStatement, userId)
	defer rows.Close()
	if err != nil {
		panic(err)
	}
	ret := make([]Lore, 0)
	var (
		message string
	)
	for rows.Next() {
		if err := rows.Scan(&message); err != nil {
			log.Fatal(err)
		}
		lore := Lore{UserID: userId, Message: message}
		ret = append(ret, lore)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return ret
}

func (p *PostgresClient) SearchLore(query string) []Lore {
	sqlStatement := `
    SELECT user_id, message FROM lores WHERE message LIKE '%' || $1 || '%'`
	rows, err := p.Db.Query(sqlStatement, query)
	defer rows.Close()
	if err != nil {
		panic(err)
	}
	ret := make([]Lore, 0)
	var (
		userId  string
		message string
	)
	for rows.Next() {
		if err := rows.Scan(&userId, &message); err != nil {
			log.Fatal(err)
		}
		lore := Lore{UserID: userId, Message: message}
		ret = append(ret, lore)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return ret
}

func (p *PostgresClient) UpvoteLore(userId string, message string) {
	sqlStatement := `
    UPDATE lores 
       SET score = score + 1 
     WHERE message IN ($1) and user_id in ($2)`
	_, err := p.Db.Query(sqlStatement, message, userId)
	if err != nil {
		panic(err)
	}
}

func (p *PostgresClient) LoreExists(message string, user_id string) bool {
	sqlStatement := `
    SELECT COUNT(*) FROM lores WHERE message IN ($1) and user_id in ($2)`
	rows, err := p.Db.Query(sqlStatement, message, user_id)
	defer rows.Close()
	if err != nil {
		panic(err)
	}
	rows.Next()
	var count int
	if err := rows.Scan(&count); err != nil {
		panic(err)
	}
	return count > 0
}

func (p *PostgresClient) InsertLore(user_id string, content string) {
	sqlStatement := `  
  INSERT INTO lores (user_id, message, score)
  VALUES ($1, $2, $3)`
	_, err := p.Db.Exec(sqlStatement, user_id, content, 0)
	if err != nil {
		panic(err)
	}
}

func NewPostgresClient(pgHost string, pgPort int, pgUser string,
	pgPassword string, pgDbname string) *PostgresClient {
	p := new(PostgresClient)
	p.Host = pgHost
	p.Port = pgPort
	p.User = pgUser
	p.Password = pgPassword
	p.Dbname = pgDbname
	p.Db = p.GetDB()
	p.Db.SetMaxOpenConns(50)
	return p
}
