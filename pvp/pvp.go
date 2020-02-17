package pvp

import (
	"fmt"

	"github.com/gocraft/dbr/v2"
	_ "github.com/lib/pq"
)

type Player struct {
	ID         string   `db:"id" json:"id"`
	Username   string   `db:"username" json:"username"`
	Server     string   `db:"server" json:"server,omitempty"`
	Servers    []string `json:"servers,omitempty"`
	IGN        string   `db:"ign" json:"ign"`
	FriendCode string   `db:"friendcode" json:"friendcode"`
	EggUltra   bool     `db:"egg_for_ultra" json:"egg_for_ultra"`
}

func (p *Player) ToString() string {
	eggUltra := "N"
	if p.EggUltra {
		eggUltra = "Y"
	}
	return fmt.Sprintf("**FC**: %s | **IGN**: %s | **User**: %s | **Egg for Ultra?**: %s", p.FriendCode, p.IGN, p.Username, eggUltra)
}

type Friendship struct {
	User   string `db:"user_id" json:"user"`
	Friend string `db:"friend_id" json:"friend"`
}

type PVP struct {
	db     *dbr.Connection
	logger dbr.EventReceiver
}

func New(db *dbr.Connection, logger dbr.EventReceiver) *PVP {
	return &PVP{
		db:     db,
		logger: logger,
	}
}

func (pvp *PVP) ListPlayers(server string) []Player {
	var results []Player

	session := pvp.db.NewSession(pvp.logger)
	session.Begin()
	session.Select("pvp_user.id", "username", "ign", "friendcode", "egg_for_ultra").
		From("pvp_user").
		Join("pvp_user_server", "pvp_user.id=pvp_user_server.user_id").
		Where("server = ?", server).
		Load(&results)
	return results
}

func (pvp *PVP) CreateUser(user *Player) error {
	session := pvp.db.NewSession(pvp.logger)
	tx, err := session.Begin()
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	_, err = tx.InsertInto("pvp_user").
		Columns("id", "username", "ign", "friendcode", "egg_for_ultra").
		Record(user).
		Exec()

	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func (pvp *PVP) RegisterUser(user *Player) error {
	session := pvp.db.NewSession(pvp.logger)
	tx, err := session.Begin()
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()

	_, err = tx.InsertInto("pvp_user_server").
		Pair("user_id", user.ID).
		Pair("server", user.Server).
		Exec()

	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func (pvp *PVP) GetPlayer(ID string) (*Player, error) {
	session := pvp.db.NewSession(pvp.logger)
	tx, err := session.Begin()
	if err != nil {
		return nil, err
	}

	var p Player
	err = tx.Select("id", "username", "ign", "friendcode", "egg_for_ultra").
		From("pvp_user").
		Where("id = ?", ID).
		LoadOne(&p)

	if err != nil {
		return nil, err
	}

	servers, err := tx.Select("server").
		From("pvp_user_server").
		Where("user_id = ?", p.ID).
		ReturnStrings()

	if err != nil {
		return &p, err
	}

	p.Servers = servers
	return &p, nil
}

func (pvp *PVP) AddFriend(f Friendship) error {
	if f.User > f.Friend {
		sorted := Friendship{
			User:   f.Friend,
			Friend: f.User,
		}
		f = sorted
	}

	session := pvp.db.NewSession(pvp.logger)
	tx, err := session.Begin()
	if err != nil {
		return err
	}
	defer tx.RollbackUnlessCommitted()
	_, err = tx.InsertInto("pvp_user_friendship").
		Columns("user_id", "friend_id").
		Record(&f).
		Exec()

	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (pvp *PVP) GetFriends(user string) ([]Player, error) {
	var friends []Player

	session := pvp.db.NewSession(pvp.logger)
	session.Begin()
	_, err := session.Select("id", "username", "ign", "friendcode", "egg_for_ultra").
		From("pvp_user").
		Join(dbr.UnionAll(
			dbr.Select(dbr.I("user_id").As("uid"), dbr.I("friend_id").As("fid")).From("pvp_user_friendship"),
			dbr.Select(dbr.I("friend_id").As("uid"), dbr.I("user_id").As("fid")).From("pvp_user_friendship"),
		).As("puf"), "pvp_user.id = puf.fid").
		Where("puf.uid = ?", user).
		Load(&friends)

	return friends, err
}
