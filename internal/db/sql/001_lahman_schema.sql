DROP TABLE IF EXISTS "AllstarFull";
CREATE TABLE "AllstarFull" (
	"playerID" varchar(9) NOT NULL,
	"yearID" int4 NOT NULL,
	"gameNum" int4,
	"gameID" varchar(12),
	"teamID" varchar(3),
	"lgID" varchar(3),
	"GP" int4,
	"startingPos" varchar(10)
);

DROP TABLE IF EXISTS "Appearances";
CREATE TABLE "Appearances" (
	"yearID" int4 NOT NULL,
	"teamID" varchar(3) NOT NULL,
	"lgID" varchar(3),
	"playerID" varchar(9) NOT NULL,
	"G_all" int4,
	"GS" int4,
	"G_batting" int4,
	"G_defense" int4,
	"G_p" int4,
	"G_c" int4,
	"G_1b" int4,
	"G_2b" int4,
	"G_3b" int4,
	"G_ss" int4,
	"G_lf" int4,
	"G_cf" int4,
	"G_rf" int4,
	"G_of" int4,
	"G_dh" int4,
	"G_ph" int4,
	"G_pr" int4
);

DROP TABLE IF EXISTS "AwardsManagers";
CREATE TABLE "AwardsManagers" (
	"playerID" varchar(10) NOT NULL,
	"awardID" varchar(75) NOT NULL,
	"yearID" int4 NOT NULL,
	"lgID" varchar(3) NOT NULL,
	"tie" varchar(1),
	"notes" varchar(100)
);

DROP TABLE IF EXISTS "AwardsPlayers";
CREATE TABLE "AwardsPlayers" (
	"playerID" varchar(9) NOT NULL,
	"awardID" varchar(255) NOT NULL,
	"yearID" int4 NOT NULL,
	"lgID" varchar(3),
	"tie" varchar(1),
	"notes" varchar(100)
);

DROP TABLE IF EXISTS "AwardsShareManagers";
CREATE TABLE "AwardsShareManagers" (
	"awardID" varchar(25) NOT NULL,
	"yearID" int4 NOT NULL,
	"lgID" varchar(3) NOT NULL,
	"playerID" varchar(10) NOT NULL,
	"pointsWon" int4,
	"pointsMax" int4,
	"votesFirst" int4
);

DROP TABLE IF EXISTS "AwardsSharePlayers";
CREATE TABLE "AwardsSharePlayers" (
	"awardID" varchar(25) NOT NULL,
	"yearID" int4 NOT NULL,
	"lgID" varchar(3) NOT NULL,
	"playerID" varchar(9) NOT NULL,
	"pointsWon" float8,
	"pointsMax" int4,
	"votesFirst" float8
);

DROP TABLE IF EXISTS "Batting";
CREATE TABLE "Batting" (
	"playerID" varchar(9) NOT NULL,
	"yearID" int4 NOT NULL,
	"stint" int4 NOT NULL,
	"teamID" varchar(3),
	"lgID" varchar(3),
	"G" int4,
	"G_batting" int4,
	"AB" int4,
	"R" int4,
	"H" int4,
	"2B" int4,
	"3B" int4,
	"HR" int4,
	"RBI" int4,
	"SB" int4,
	"CS" int4,
	"BB" int4,
	"SO" int4,
	"IBB" int4,
	"HBP" int4,
	"SH" int4,
	"SF" int4,
	"GIDP" int4,
	"G_old" int4
);

DROP TABLE IF EXISTS "BattingPost";
CREATE TABLE "BattingPost" (
	"yearID" int4 NOT NULL,
	"round" varchar(10) NOT NULL,
	"playerID" varchar(9) NOT NULL,
	"teamID" varchar(3),
	"lgID" varchar(3),
	"G" int4,
	"AB" int4,
	"R" int4,
	"H" int4,
	"2B" int4,
	"3B" int4,
	"HR" int4,
	"RBI" int4,
	"SB" int4,
	"CS" int4,
	"BB" int4,
	"SO" int4,
	"IBB" int4,
	"HBP" int4,
	"SH" int4,
	"SF" int4,
	"GIDP" int4
);


--  Table structure for CollegePlaying

DROP TABLE IF EXISTS "CollegePlaying";
CREATE TABLE "CollegePlaying" (
	"playerID" varchar(9),
	"schoolID" varchar(15),
	"yearID" int4
);

DROP TABLE IF EXISTS "Fielding";
CREATE TABLE "Fielding" (
	"playerID" varchar(9) NOT NULL,
	"yearID" int4 NOT NULL,
	"stint" int4 NOT NULL,
	"teamID" varchar(3),
	"lgID" varchar(3),
	"POS" varchar(3) NOT NULL,
	"G" int4,
	"GS" int4,
	"InnOuts" int4,
	"PO" int4,
	"A" int4,
	"E" int4,
	"DP" int4,
	"PB" int4,
	"WP" int4,
	"SB" int4,
	"CS" int4,
	"ZR" float8
);

DROP TABLE IF EXISTS "FieldingOF";
CREATE TABLE "FieldingOF" (
	"playerID" varchar(9) NOT NULL,
	"yearID" int4 NOT NULL,
	"stint" int4 NOT NULL,
	"Glf" int4,
	"Gcf" int4,
	"Grf" int4
);

DROP TABLE IF EXISTS "FieldingOFsplit";
CREATE TABLE "FieldingOFsplit" (
	"playerID" varchar(9) NOT NULL,
	"yearID" int4 NOT NULL,
	"stint" int4 NOT NULL,
	"teamID" varchar(3),
	"lgID" varchar(3),
	"POS" varchar(3),
	"G" int4,
	"GS" int4,
	"InnOuts" int4,
	"PO" int4,
	"A" int4,
	"E" int4,
	"DP" int4,
	"PB" int4,
	"WP" int4,
	"SB" int4,
	"CS" int4,
	"ZR" float8
);

DROP TABLE IF EXISTS "FieldingPost";
CREATE TABLE "FieldingPost" (
	"playerID" varchar(9) NOT NULL,
	"yearID" int4 NOT NULL,
	"teamID" varchar(3),
	"lgID" varchar(3),
	"round" varchar(10) NOT NULL,
	"POS" varchar(3) NOT NULL,
	"G" int4,
	"GS" int4,
	"InnOuts" int4,
	"PO" int4,
	"A" int4,
	"E" int4,
	"DP" int4,
	"TP" int4,
	"PB" int4,
	"SB" int4,
	"CS" int4
);

DROP TABLE IF EXISTS "HomeGames";
CREATE TABLE "HomeGames" (
	"year_key" int4 NOT NULL,
	"league_key" varchar(3) NOT NULL,
	"team_key" varchar(3) NOT NULL,
	"parkkey" varchar(255) NOT NULL,
	"span_first" varchar(10),
	"span_last" varchar(10),
	"games" int4,
	"openings" int4,
	"attendance" int4
);

DROP TABLE IF EXISTS "HallOfFame";
CREATE TABLE "HallOfFame" (
	"playerID" varchar(10) NOT NULL,
	"yearid" int4 NOT NULL,
	"votedBy" varchar(64) NOT NULL,
	"ballots" int4,
	"needed" int4,
	"votes" int4,
	"inducted" varchar(1),
	"category" varchar(20),
	"needed_note" varchar(150)
);

DROP TABLE IF EXISTS "Managers";
CREATE TABLE "Managers" (
	"playerID" varchar(10),
	"yearID" int4 NOT NULL,
	"teamID" varchar(3) NOT NULL,
	"lgID" varchar(3),
	"inseason" int4 NOT NULL,
	"G" int4,
	"W" int4,
	"L" int4,
	"rank" int4,
	"plyrMgr" varchar(1)
);

DROP TABLE IF EXISTS "ManagersHalf";
CREATE TABLE "ManagersHalf" (
	"playerID" varchar(10) NOT NULL,
	"yearID" int4 NOT NULL,
	"teamID" varchar(3) NOT NULL,
	"lgID" varchar(3),
	"inseason" int4,
	"half" int4 NOT NULL,
	"G" int4,
	"W" int4,
	"L" int4,
	"rank" int4
);

DROP TABLE IF EXISTS "People";
CREATE TABLE "People" (
	"ID" int4,
	"playerID" varchar(10) NOT NULL,
	"birthYear" int4,
	"birthMonth" int4,
	"birthDay" int4,
	"birthCity" varchar(50),
	"birthCountry" varchar(50),
	"birthState" varchar(50),
	"deathYear" int4,
	"deathMonth" int4,
	"deathDay" int4,
	"deathCountry" varchar(50),
	"deathState" varchar(50),
	"deathCity" varchar(50),
	"nameFirst" varchar(50),
	"nameLast" varchar(50),
	"nameGiven" varchar(255),
	"weight" int4,
	"height" int4,
	"bats" varchar(1),
	"throws" varchar(1),
	"debut" varchar(10),
	"bbrefID" varchar(9),
	"finalGame" varchar(10),
	"retroID" varchar(9)
);

DROP TABLE IF EXISTS "Parks";
CREATE TABLE "Parks" (
	"ID" int4,
	"parkalias" varchar(255),
	"parkkey" varchar(255),
	"parkname" varchar(255),
	"city" varchar(255),
	"state" varchar(255),
	"country" varchar(255)
);

DROP TABLE IF EXISTS "Pitching";
CREATE TABLE "Pitching" (
	"playerID" varchar(9) NOT NULL,
	"yearID" int4 NOT NULL,
	"stint" int4 NOT NULL,
	"teamID" varchar(3),
	"lgID" varchar(3),
	"W" int4,
	"L" int4,
	"G" int4,
	"GS" int4,
	"CG" int4,
	"SHO" int4,
	"SV" int4,
	"IPouts" int4,
	"H" int4,
	"ER" int4,
	"HR" int4,
	"BB" int4,
	"SO" int4,
	"BAOpp" float8,
	"ERA" float8,
	"IBB" int4,
	"WP" int4,
	"HBP" int4,
	"BK" int4,
	"BFP" int4,
	"GF" int4,
	"R" int4,
	"SH" int4,
	"SF" int4,
	"GIDP" int4
);

DROP TABLE IF EXISTS "PitchingPost";
CREATE TABLE "PitchingPost" (
	"playerID" varchar(9) NOT NULL,
	"yearID" int4 NOT NULL,
	"round" varchar(10) NOT NULL,
	"teamID" varchar(3),
	"lgID" varchar(3),
	"W" int4,
	"L" int4,
	"G" int4,
	"GS" int4,
	"CG" int4,
	"SHO" int4,
	"SV" int4,
	"IPouts" int4,
	"H" int4,
	"ER" int4,
	"HR" int4,
	"BB" int4,
	"SO" int4,
	"BAOpp" float8,
	"ERA" float8,
	"IBB" int4,
	"WP" int4,
	"HBP" int4,
	"BK" int4,
	"BFP" int4,
	"GF" int4,
	"R" int4,
	"SH" int4,
	"SF" int4,
	"GIDP" int4
);

DROP TABLE IF EXISTS "Salaries";
CREATE TABLE "Salaries" (
	"yearID" int4 NOT NULL,
	"teamID" varchar(3) NOT NULL,
	"lgID" varchar(3) NOT NULL,
	"playerID" varchar(9) NOT NULL,
	"salary" float8
);

DROP TABLE IF EXISTS "Schools";
CREATE TABLE "Schools" (
	"schoolID" varchar(15) NOT NULL,
	"name_full" varchar(255),
	"city" varchar(55),
	"state" varchar(55),
	"country" varchar(55)
);

DROP TABLE IF EXISTS "SeriesPost";
CREATE TABLE "SeriesPost" (
	"yearID" int4 NOT NULL,
	"round" varchar(5) NOT NULL,
	"teamIDwinner" varchar(3),
	"lgIDwinner" varchar(3),
	"teamIDloser" varchar(3),
	"lgIDloser" varchar(3),
	"wins" int4,
	"losses" int4,
	"ties" int4
);

DROP TABLE IF EXISTS "Teams";
CREATE TABLE "Teams" (
	"yearID" int4 NOT NULL,
	"lgID" varchar(3) NOT NULL,
	"teamID" varchar(3) NOT NULL,
	"franchID" varchar(3),
	"divID" varchar(1),
	"Rank" int4,
	"G" int4,
	"Ghome" int4,
	"W" int4,
	"L" int4,
	"DivWin" varchar(1),
	"WCWin" varchar(1),
	"LgWin" varchar(1),
	"WSWin" varchar(1),
	"R" int4,
	"AB" int4,
	"H" int4,
	"2B" int4,
	"3B" int4,
	"HR" int4,
	"BB" int4,
	"SO" int4,
	"SB" int4,
	"CS" int4,
	"HBP" int4,
	"SF" int4,
	"RA" int4,
	"ER" int4,
	"ERA" float8,
	"CG" int4,
	"SHO" int4,
	"SV" int4,
	"IPouts" int4,
	"HA" int4,
	"HRA" int4,
	"BBA" int4,
	"SOA" int4,
	"E" int4,
	"DP" int4,
	"FP" float8,
	"name" varchar(50),
	"park" varchar(255),
	"attendance" int4,
	"BPF" int4,
	"PPF" int4,
	"teamIDBR" varchar(3),
	"teamIDlahman45" varchar(3),
	"teamIDretro" varchar(3)
);

DROP TABLE IF EXISTS "TeamsFranchises";
CREATE TABLE "TeamsFranchises" (
	"franchID" varchar(3) NOT NULL,
	"franchName" varchar(50),
	"active" varchar(3),
	"NAassoc" varchar(3)
);

DROP TABLE IF EXISTS "TeamsHalf";
CREATE TABLE "TeamsHalf" (
	"yearID" int4 NOT NULL,
	"lgID" varchar(3) NOT NULL,
	"teamID" varchar(3) NOT NULL,
	"Half" varchar(1) NOT NULL,
	"divID" varchar(1),
	"DivWin" varchar(1),
	"Rank" int4,
	"G" int4,
	"W" int4,
	"L" int4
);

CREATE INDEX "AllstarFull_playerID_idx" ON "AllstarFull" ("playerID");
CREATE INDEX "AllstarFull_yearID_idx" ON "AllstarFull" ("yearID");
ALTER TABLE "Appearances" ADD PRIMARY KEY ("yearID", "teamID", "playerID") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "AwardsManagers" ADD PRIMARY KEY ("yearID", "awardID", "lgID", "playerID") NOT DEFERRABLE INITIALLY IMMEDIATE;

CREATE INDEX "AwardsPlayers_playerID_idx" ON "AwardsPlayers" ("playerID");
CREATE INDEX "AwardsPlayers_awardID_idx" ON "AwardsPlayers" ("awardID");
CREATE INDEX "AwardsPlayers_yearID_idx" ON "AwardsPlayers" ("yearID");
ALTER TABLE "AwardsShareManagers" ADD PRIMARY KEY ("awardID", "yearID", "lgID", "playerID") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "AwardsSharePlayers" ADD PRIMARY KEY ("awardID", "yearID", "lgID", "playerID") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "Batting" ADD PRIMARY KEY ("playerID", "yearID", "stint") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "BattingPost" ADD PRIMARY KEY ("yearID", "round", "playerID") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "Fielding" ADD PRIMARY KEY ("playerID", "yearID", "stint", "POS") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "FieldingOF" ADD PRIMARY KEY ("playerID", "yearID", "stint") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "FieldingOFsplit" ADD PRIMARY KEY ("playerID", "yearID", "stint", "POS") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "FieldingPost" ADD PRIMARY KEY ("playerID", "yearID", "round", "POS") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "HomeGames" ADD PRIMARY KEY ("year_key", "league_key", "team_key", "parkkey") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "HallOfFame" ADD PRIMARY KEY ("playerID", "yearid", "votedBy") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "Managers" ADD PRIMARY KEY ("yearID", "teamID", "inseason") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "ManagersHalf" ADD PRIMARY KEY ("yearID", "teamID", "playerID", "half") NOT DEFERRABLE INITIALLY IMMEDIATE;

CREATE INDEX "Parks_parkkey_idx" ON "Parks" ("parkkey");
ALTER TABLE "People" ADD PRIMARY KEY ("playerID") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "Pitching" ADD PRIMARY KEY ("playerID", "yearID", "stint") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "PitchingPost" ADD PRIMARY KEY ("playerID", "yearID", "round") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "Salaries" ADD PRIMARY KEY ("yearID", "teamID", "lgID", "playerID") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "Schools" ADD PRIMARY KEY ("schoolID") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "SeriesPost" ADD PRIMARY KEY ("yearID", "round") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "Teams" ADD PRIMARY KEY ("yearID", "lgID", "teamID") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "TeamsFranchises" ADD PRIMARY KEY ("franchID") NOT DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "TeamsHalf" ADD PRIMARY KEY ("yearID", "teamID", "lgID", "Half") NOT DEFERRABLE INITIALLY IMMEDIATE;
