-- Win Expectancy Historical Table
-- Stores historical win probabilities for all game states to enable accurate leverage index calculations

CREATE TABLE IF NOT EXISTS win_expectancy_historical (
    id SERIAL PRIMARY KEY,

    -- Game state dimensions
    inning INT NOT NULL,
    is_bottom BOOLEAN NOT NULL,
    outs INT NOT NULL,
    runners_state VARCHAR(8) NOT NULL,
    score_diff INT NOT NULL,

    -- Win probability data
    win_probability DECIMAL(5,4) NOT NULL,
    sample_size INT NOT NULL,

    start_year INT,
    end_year INT,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT valid_inning CHECK (inning >= 1 AND inning <= 9),
    CONSTRAINT valid_outs CHECK (outs >= 0 AND outs <= 2),
    CONSTRAINT valid_score_diff CHECK (score_diff >= -11 AND score_diff <= 11),
    CONSTRAINT valid_win_prob CHECK (win_probability >= 0 AND win_probability <= 1),
    CONSTRAINT valid_runners_state CHECK (runners_state ~ '^[_123]{3}$'),

    UNIQUE (inning, is_bottom, outs, runners_state, score_diff, start_year, end_year)
);

CREATE INDEX idx_we_lookup ON win_expectancy_historical(
    inning, is_bottom, outs, runners_state, score_diff
);

CREATE INDEX idx_we_era ON win_expectancy_historical(
    start_year, end_year
);

COMMENT ON TABLE win_expectancy_historical IS 'Historical win expectancy probabilities by game state for leverage index calculations';
COMMENT ON COLUMN win_expectancy_historical.runners_state IS 'Base state encoded as 3-char string: ___ = empty, 1__ = runner on first, _2_ = runner on second, __3 = runner on third, 12_ = first and second, 1_3 = first and third, _23 = second and third, 123 = bases loaded';
COMMENT ON COLUMN win_expectancy_historical.score_diff IS 'Score differential from batting team perspective, capped at ±11';
COMMENT ON COLUMN win_expectancy_historical.win_probability IS 'Probability that the home team wins from this game state';
COMMENT ON COLUMN win_expectancy_historical.sample_size IS 'Number of historical game states used to calculate this probability';
