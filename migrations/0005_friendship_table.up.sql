CREATE TABLE IF NOT EXISTS pvp_user_friendship(
	user_id varchar(50) NOT NULL REFERENCES pvp_user(id) ON DELETE CASCADE,
	friend_id varchar(50) NOT NULL REFERENCES pvp_user(id) ON DELETE CASCADE,
	PRIMARY KEY (user_id, friend_id),
	CHECK (user_id < friend_id)
);
