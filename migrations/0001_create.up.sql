CREATE TABLE IF NOT EXISTS pvp(
	id varchar(50) NOT NULL,
	username varchar(50) NOT NULL,
	server varchar(100) NOT NULL,
	ign varchar(100) NOT NULL,
	friendcode varchar(12) NOT NULL,
	PRIMARY KEY (id, server)
);
