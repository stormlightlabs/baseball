-- Create player_relatives table for family relationships
-- Data source: Retrosheet relatives.csv

CREATE TABLE IF NOT EXISTS player_relatives (
    player_id_1 VARCHAR(10) NOT NULL,
    relation_type VARCHAR(50) NOT NULL,
    player_id_2 VARCHAR(10) NOT NULL,
    PRIMARY KEY (player_id_1, player_id_2, relation_type)
);

CREATE INDEX IF NOT EXISTS idx_player_relatives_player1 ON player_relatives(player_id_1);
CREATE INDEX IF NOT EXISTS idx_player_relatives_player2 ON player_relatives(player_id_2);
CREATE INDEX IF NOT EXISTS idx_player_relatives_type ON player_relatives(relation_type);

COMMENT ON TABLE player_relatives IS 'Family relationships between players from Retrosheet biodata';
COMMENT ON COLUMN player_relatives.player_id_1 IS 'First player Retrosheet ID';
COMMENT ON COLUMN player_relatives.relation_type IS 'Type of relationship (Brother, Father, Uncle, etc.)';
COMMENT ON COLUMN player_relatives.player_id_2 IS 'Second player Retrosheet ID';
