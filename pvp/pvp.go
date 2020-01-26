package pvp

import (
	//"fmt"
	"log"

	"github.com/gocraft/dbr"
	_ "github.com/lib/pq"
)

const tableName = "pvp"

type Player struct {
	ID         string `db:"id" json:"id"`
	Username   string `db:"username" json:"username"`
	Server     string `db:"server" json:"server,omitempty"`
	IGN        string `db:"ign" json:"ign"`
	FriendCode string `db:"friendcode" json:"friendcode"`
}

type PVP struct {
	db     *dbr.Connection
	logger dbr.EventReceiver
}

func New(db *dbr.Connection, logger dbr.EventReceiver) (*PVP, error) {
	pvp := &PVP{
		db:     db,
		logger: logger,
	}
	err := pvp.createDB()
	if err != nil {
		log.Println(err)
		_ = pvp.db.Close()
		return nil, err
	}
	return pvp, nil
}

func (pvp *PVP) ListPlayers(server string) []Player {
	var results []Player

	session := pvp.db.NewSession(pvp.logger)
	session.Begin()
	session.Select("id", "username", "ign", "friendcode").
		From(tableName).
		Where("server = ?", server).
		Load(&results)
	return results
}

func (pvp *PVP) Register(user Player) error {
	session := pvp.db.NewSession(pvp.logger)
	tx, err := session.Begin()
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	_, err = tx.InsertInto(tableName).
		Columns("id", "username", "ign", "friendcode", "server").
		Record(user).
		Exec()

	if err != nil {
		// TODO figure out type
		return err
	}

	tx.Commit()
	return nil
}

func (pvp *PVP) createDB() error {
	createSQL := `
	CREATE TABLE IF NOT EXISTS pvp(
		id varchar(50) NOT NULL,
		username varchar(50) NOT NULL,
		server varchar(100) NOT NULL,
		ign varchar(100) NOT NULL,
		friendcode varchar(12) NOT NULL,
		PRIMARY KEY (id, server)
	)`
	_, err := pvp.db.Exec(createSQL)
	return err
}
