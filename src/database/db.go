package database

import (
	"awesomeProject/game"
	"context"
	"database/sql"
	_ "github.com/genjidb/genji/driver"
)

type KulkiDatabase struct {
	db *sql.DB
}

func (k *KulkiDatabase) Close() error {
	return k.db.Close()
}

func SetupDb() *KulkiDatabase {

	db, err := sql.Open("genji", ":memory:")
	//var db *sql.DB
	//var err error
	//db, err = genji.Open(":memory:")
	if err != nil {
		panic(err)
	}
	var ctx = context.TODO()

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS user (email, password);")
	if err != nil {
		panic(err)
	}

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS game (email, board, number_of_colors, next_ball_colors, score, world_depth, world_width);")
	if err != nil {
		panic(err)
	}

	_, err = db.ExecContext(ctx, "CREATE UNIQUE INDEX IF NOT EXISTS idx_email ON user (email);")
	if err != nil {
		panic(err)
	}

	_, err = db.ExecContext(ctx, "CREATE UNIQUE INDEX IF NOT EXISTS idx_game ON game (email);")
	if err != nil {
		panic(err)
	}

	return &KulkiDatabase{db: db}
}

func (k *KulkiDatabase) upsert(data []interface{}) error {

	var ctx = context.TODO()

	_, err := k.db.ExecContext(ctx, data[0].(string), data[1:]...)

	return err
}

func (k *KulkiDatabase) SetBoard(email *EmailUser, board *game.Board) error {
	return k.upsert([]interface{}{"INSERT INTO game (email, board, number_of_colors, next_ball_colors, score, world_depth, world_width) VALUES (?,?,?,?,?,?,?);",
		email.Email,
		board.Memento(),
		board.NumberOfColors,
		board.NextBallColorsMemento(),
		board.Score,
		board.Depth(),
		board.Width(),
	})
}

func (k *KulkiDatabase) AddUser(user *User) error {

	return k.upsert([]interface{}{"INSERT INTO user (email, password) VALUES (?,?)", user.Email, user.Password})
}

func (k *KulkiDatabase) UpdateBoard(email *EmailUser, board *game.Board) error {
	return k.upsert([]interface{}{"UPDATE game SET board = ?, next_ball_colors = ?, score = ? WHERE email = ?",
		board.Memento(),
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

	row := k.db.QueryRow("select board, number_of_colors, next_ball_colors, score, world_depth, world_width from game where email = ?", email.Email)

	error := row.Scan(&record.memento, &record.colors, &record.nextColors, &record.score, &record.depth, &record.width)

	if error != nil {
		return nil, nil, error
	}

	board, commands := game.ReviveBoard(record.width, record.depth, record.colors, record.memento, record.nextColors, record.score)

	return board, commands, nil
}

func (k *KulkiDatabase) GetUser(email string) (*User, error) {

	var user = User{EmailUser{Email: ""}, PasswordUser{Password: ""}}

	row := k.db.QueryRow("select email, password from user where email = ?", email)

	err := row.Scan(&user.EmailUser.Email, &user.PasswordUser.Password)

	return &user, err
}

func (k *KulkiDatabase) DeleteBoard(user *EmailUser) error {
	return k.upsert([]interface{}{"DELETE FROM game WHERE email = ?", user.Email})
}
