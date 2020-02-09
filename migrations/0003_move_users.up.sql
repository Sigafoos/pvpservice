BEGIN;
	ALTER TABLE pvp RENAME TO pvp_user;

	CREATE TABLE IF NOT EXISTS pvp_user_server(
		user_id varchar(50) NOT NULL,
		server varchar(100) NOT NULL,
		PRIMARY KEY(user_id, server)
	);

	INSERT INTO pvp_user_server(user_id, server)
	SELECT id, server FROM pvp_user;

	ALTER TABLE pvp_user DROP COLUMN server;
	ALTER TABLE pvp_user ADD PRIMARY KEY (id);

	COMMIT;
