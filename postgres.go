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
	DB       *sql.DB
}

func (p *PostgresClient) GetDB() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.User, p.Password, p.Dbname)
	DB, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = DB.Ping()
	if err != nil {
		panic(err)
	}
	return DB
}

type Lore struct {
	userID  string
	Message string
	Score   int
}

func (p *PostgresClient) RecentLore() []Lore {
	sqlStatement := `
    SELECT user_id, message, score FROM lores ORDER BY timestamp_added DESC LIMIT 3`
	rows, err := p.DB.Query(sqlStatement)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	ret := make([]Lore, 0)

	var l Lore
	for rows.Next() {
		if err := rows.Scan(&l.userID, &l.Message, &l.Score); err != nil {
			log.Fatal(err)
		}
		ret = append(ret, l)

	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return ret
}

func (p *PostgresClient) RandomLore() []Lore {
	sqlStatement := `
    SELECT user_id, message, score FROM lores ORDER BY RANDOM() LIMIT 1`
	rows, err := p.DB.Query(sqlStatement)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	ret := make([]Lore, 0)

	var l Lore
	for rows.Next() {
		if err := rows.Scan(&l.userID, &l.Message, &l.Score); err != nil {
			log.Fatal(err)
		}
		ret = append(ret, l)

	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return ret
}

func (p *PostgresClient) TopLore() []Lore {
	sqlStatement := `
    SELECT user_id, message, score FROM lores ORDER BY score DESC LIMIT 3`
	rows, err := p.DB.Query(sqlStatement)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	ret := make([]Lore, 0)

	var l Lore
	for rows.Next() {
		if err := rows.Scan(&l.userID, &l.Message, &l.Score); err != nil {
			log.Fatal(err)
		}
		ret = append(ret, l)

	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return ret
}

func (p *PostgresClient) LoreForUser(userID string) []Lore {
	sqlStatement := `
    SELECT message, score FROM lores WHERE user_id IN ($1)`
	rows, err := p.DB.Query(sqlStatement, userID)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	ret := make([]Lore, 0)

	var l Lore
	for rows.Next() {
		if err := rows.Scan(&l.userID, &l.Score); err != nil {
			log.Fatal(err)
		}
		ret = append(ret, l)

	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return ret
}

func (p *PostgresClient) SearchLore(query string) []Lore {
	sqlStatement := `
    SELECT user_id, message, score FROM lores WHERE message LIKE '%' || $1 || '%'`
	rows, err := p.DB.Query(sqlStatement, query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	ret := make([]Lore, 0)

	var l Lore
	for rows.Next() {
		if err := rows.Scan(&l.userID, &l.Message, &l.Score); err != nil {
			log.Fatal(err)
		}
		ret = append(ret, l)

	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return ret
}

func (p *PostgresClient) UpvoteLore(userID string, message string) {
	sqlStatement := `
    UPDATE lores 
       SET score = score + 1 
     WHERE message IN ($1) and user_id in ($2)`
	_, err := p.DB.Query(sqlStatement, message, userID)
	if err != nil {
		panic(err)
	}
}

func (p *PostgresClient) LoreExists(message string, user_id string) bool {
	sqlStatement := `
    SELECT COUNT(*) FROM lores WHERE message IN ($1) and user_id in ($2)`
	rows, err := p.DB.Query(sqlStatement, message, user_id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

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
	_, err := p.DB.Exec(sqlStatement, user_id, content, 0)
	if err != nil {
		panic(err)
	}
}

func NewPostgresClient(p *PostgresClient) *PostgresClient {
	p.DB = p.GetDB()
	p.DB.SetMaxOpenConns(50)
	return p
}
