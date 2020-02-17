ALTER TABLE pvp_user_server
ADD CONSTRAINT pvp_user_server_user_fk
FOREIGN KEY (user_id) REFERENCES pvp_user(id)
ON DELETE CASCADE;
