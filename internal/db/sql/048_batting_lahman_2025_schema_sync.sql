-- Align Batting schema with Lahman 2025+ CSV layout.
-- Lahman 2025 removed legacy Batting columns G_batting and G_old.

ALTER TABLE "Batting"
    DROP COLUMN IF EXISTS "G_batting",
    DROP COLUMN IF EXISTS "G_old";

-- Lahman historical batting rows often leave trailing numeric fields blank.
-- Treat blanks as zero for stable aggregations/scan behavior.
ALTER TABLE "Batting"
    ALTER COLUMN "G" SET DEFAULT 0,
    ALTER COLUMN "AB" SET DEFAULT 0,
    ALTER COLUMN "R" SET DEFAULT 0,
    ALTER COLUMN "H" SET DEFAULT 0,
    ALTER COLUMN "2B" SET DEFAULT 0,
    ALTER COLUMN "3B" SET DEFAULT 0,
    ALTER COLUMN "HR" SET DEFAULT 0,
    ALTER COLUMN "RBI" SET DEFAULT 0,
    ALTER COLUMN "SB" SET DEFAULT 0,
    ALTER COLUMN "CS" SET DEFAULT 0,
    ALTER COLUMN "BB" SET DEFAULT 0,
    ALTER COLUMN "SO" SET DEFAULT 0,
    ALTER COLUMN "IBB" SET DEFAULT 0,
    ALTER COLUMN "HBP" SET DEFAULT 0,
    ALTER COLUMN "SH" SET DEFAULT 0,
    ALTER COLUMN "SF" SET DEFAULT 0,
    ALTER COLUMN "GIDP" SET DEFAULT 0;

UPDATE "Batting"
SET
    "G" = COALESCE("G", 0),
    "AB" = COALESCE("AB", 0),
    "R" = COALESCE("R", 0),
    "H" = COALESCE("H", 0),
    "2B" = COALESCE("2B", 0),
    "3B" = COALESCE("3B", 0),
    "HR" = COALESCE("HR", 0),
    "RBI" = COALESCE("RBI", 0),
    "SB" = COALESCE("SB", 0),
    "CS" = COALESCE("CS", 0),
    "BB" = COALESCE("BB", 0),
    "SO" = COALESCE("SO", 0),
    "IBB" = COALESCE("IBB", 0),
    "HBP" = COALESCE("HBP", 0),
    "SH" = COALESCE("SH", 0),
    "SF" = COALESCE("SF", 0),
    "GIDP" = COALESCE("GIDP", 0)
WHERE
    "G" IS NULL OR
    "AB" IS NULL OR
    "R" IS NULL OR
    "H" IS NULL OR
    "2B" IS NULL OR
    "3B" IS NULL OR
    "HR" IS NULL OR
    "RBI" IS NULL OR
    "SB" IS NULL OR
    "CS" IS NULL OR
    "BB" IS NULL OR
    "SO" IS NULL OR
    "IBB" IS NULL OR
    "HBP" IS NULL OR
    "SH" IS NULL OR
    "SF" IS NULL OR
    "GIDP" IS NULL;
