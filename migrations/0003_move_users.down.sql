BEGIN
	ALTER TABLE pvp_user ADD COLUMN server varchar(50) NOT NULL;

	UPDATE pvp_user pu
	SET pu.server=pus.server
	FROM pvp_user_server pus
	WHERE pu.id=pus.user_id;

	DROP TABLE pvp_user_server;

	COMMIT;
