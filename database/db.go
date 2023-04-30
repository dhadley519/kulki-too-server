package database

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"kulki/game"
	"os"
)

type KulkiDatabase struct {
	db *sql.DB
}

func (k *KulkiDatabase) Close() error {
	return k.db.Close()
}

func SetupDb(uid string, pwd string, host string) *KulkiDatabase {

	db, err := sql.Open("postgres", fmt.Sprintf("postgresql://%s:%s@%s:5432?sslmode=disable", uid, pwd, host))
	//db, err := sql.Open("genji", ":memory:")
	//var db *sql.DB
	//var err error
	//db, err = genji.Open(":memory:")
	if err != nil {
		panic(err)
	}
	var ctx = context.TODO()

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS user_tbl (email text, passw text)")
	if err != nil {
		panic(err)
	}

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS game_tbl (email text, board text, number_of_colors integer, next_ball_colors text, score integer, world_depth integer, world_width integer)")
	if err != nil {
		panic(err)
	}

	_, err = db.ExecContext(ctx, "CREATE UNIQUE INDEX IF NOT EXISTS idx_email ON user_tbl (email)")
	if err != nil {
		panic(err)
	}

	_, err = db.ExecContext(ctx, "CREATE UNIQUE INDEX IF NOT EXISTS idx_game ON game_tbl (email)")
	if err != nil {
		panic(err)
	}

	return &KulkiDatabase{db: db}
}

//func (k *KulkiDatabase) upsert(data []interface{}) error {
//
//	var ctx = context.TODO()
//
//	_, err := k.db.ExecContext(ctx, data[0].(string), data[1:]...)
//
//	return err
//}

func (k *KulkiDatabase) SetBoard(email *EmailUser, board *game.Board) error {

	data := []interface{}{
		email.Email,
		board.Memento(),
		board.NumberOfColors,
		board.NextBallColorsMemento(),
		board.Score,
		board.Depth(),
		board.Width()}

	return k.upsert("INSERT INTO game_tbl (email, board, number_of_colors, next_ball_colors, score, world_depth, world_width) VALUES ($1,$2,$3,$4,$5,$6,$7)", data)
}

func (k *KulkiDatabase) upsert(squeal string, data []any) error {

	stmt, err := k.db.Prepare(squeal)
	if err != nil {
		return err
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			_, _ = os.Stderr.WriteString(err.Error())
		}
	}(stmt)

	_, execErr := stmt.Exec(data...)
	return execErr
}

func (k *KulkiDatabase) AddUser(user *User) error {
	return k.upsert("INSERT INTO user_tbl (email, passw) VALUES ($1, $2)", []interface{}{user.Email, user.Password})
}

func (k *KulkiDatabase) UpdateBoard(email *EmailUser, board *game.Board) error {
	return k.upsert("UPDATE game_tbl SET board = $1, next_ball_colors = $2, score = $3 WHERE email = $4",
		[]interface{}{board.Memento(),
			board.NextBallColorsMemento(),
			board.Score,
			email.Email,
		})
}

type boardRecord struct {
	memento    string
	colors     int
	nextColors string
	score      int
	depth      int
	width      int
}

func (k *KulkiDatabase) GetBoard(email *EmailUser) (*game.Board, []*game.BoardCommand, error) {

	var record = boardRecord{memento: "", colors: 0, nextColors: "", score: 0, depth: 0, width: 0}

	row := k.db.QueryRow("select board, number_of_colors, next_ball_colors, score, world_depth, world_width from game_tbl where email = $1", email.Email)

	err := row.Scan(&record.memento, &record.colors, &record.nextColors, &record.score, &record.depth, &record.width)

	if err != nil {
		return nil, nil, err
	}

	board, commands := game.ReviveBoard(record.width, record.depth, record.colors, record.memento, record.nextColors, record.score)

	return board, commands, nil
}

func (k *KulkiDatabase) GetUser(email string) (*User, error) {

	var user = User{EmailUser{Email: ""}, PasswordUser{Password: ""}}

	row := k.db.QueryRow("select email, passw from user_tbl where email = $1", email)

	err := row.Scan(&user.EmailUser.Email, &user.PasswordUser.Password)

	return &user, err
}

func (k *KulkiDatabase) DeleteBoard(user *EmailUser) error {
	return k.upsert("DELETE FROM game_tbl WHERE email = $1", []interface{}{user.Email})
}
