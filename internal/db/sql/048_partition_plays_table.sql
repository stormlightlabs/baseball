-- Partition the plays table by year to improve query performance

-- Step 1: Create new partitioned table with same schema as plays
CREATE TABLE plays_partitioned (
    gid character varying(12) NOT NULL,
    event text,
    inning integer,
    top_bot integer,
    vis_home integer,
    site character varying(5),
    batteam character varying(3),
    pitteam character varying(3),
    score_v integer,
    score_h integer,
    batter character varying(8),
    pitcher character varying(8),
    lp integer,
    bat_f integer,
    bathand character varying(1),
    pithand character varying(1),
    balls integer,
    strikes integer,
    count character varying(10),
    pitches text,
    nump integer,
    pa integer,
    ab integer,
    single integer,
    double integer,
    triple integer,
    hr integer,
    sh integer,
    sf integer,
    hbp integer,
    walk integer,
    k integer,
    xi integer,
    roe integer,
    fc integer,
    othout integer,
    noout integer,
    oth integer,
    bip integer,
    bunt integer,
    ground integer,
    fly integer,
    line integer,
    iw integer,
    gdp integer,
    othdp integer,
    tp integer,
    fle integer,
    wp integer,
    pb integer,
    bk integer,
    oa integer,
    di integer,
    sb2 integer,
    sb3 integer,
    sbh integer,
    cs2 integer,
    cs3 integer,
    csh integer,
    pko1 integer,
    pko2 integer,
    pko3 integer,
    k_safe integer,
    e1 integer,
    e2 integer,
    e3 integer,
    e4 integer,
    e5 integer,
    e6 integer,
    e7 integer,
    e8 integer,
    e9 integer,
    outs_pre integer,
    outs_post integer,
    br1_pre character varying(8),
    br2_pre character varying(8),
    br3_pre character varying(8),
    br1_post character varying(8),
    br2_post character varying(8),
    br3_post character varying(8),
    lob_id1 character varying(8),
    lob_id2 character varying(8),
    lob_id3 character varying(8),
    pr1_pre character varying(8),
    pr2_pre character varying(8),
    pr3_pre character varying(8),
    pr1_post character varying(8),
    pr2_post character varying(8),
    pr3_post character varying(8),
    run_b character varying(8),
    run1 character varying(8),
    run2 character varying(8),
    run3 character varying(8),
    prun_b character varying(8),
    prun1 character varying(8),
    prun2 character varying(8),
    prun3 character varying(8),
    ur_b integer,
    ur1 integer,
    ur2 integer,
    ur3 integer,
    rbi_b integer,
    rbi1 integer,
    rbi2 integer,
    rbi3 integer,
    runs integer,
    rbi integer,
    er integer,
    tur integer,
    l1 character varying(8),
    l2 character varying(8),
    l3 character varying(8),
    l4 character varying(8),
    l5 character varying(8),
    l6 character varying(8),
    l7 character varying(8),
    l8 character varying(8),
    l9 character varying(8),
    lf1 integer,
    lf2 integer,
    lf3 integer,
    lf4 integer,
    lf5 integer,
    lf6 integer,
    lf7 integer,
    lf8 integer,
    lf9 integer,
    f2 character varying(8),
    f3 character varying(8),
    f4 character varying(8),
    f5 character varying(8),
    f6 character varying(8),
    f7 character varying(8),
    f8 character varying(8),
    f9 character varying(8),
    po0 integer,
    po1 integer,
    po2 integer,
    po3 integer,
    po4 integer,
    po5 integer,
    po6 integer,
    po7 integer,
    po8 integer,
    po9 integer,
    a1 integer,
    a2 integer,
    a3 integer,
    a4 integer,
    a5 integer,
    a6 integer,
    a7 integer,
    a8 integer,
    a9 integer,
    fseq character varying(20),
    batout1 integer,
    batout2 integer,
    batout3 integer,
    brout_b integer,
    brout1 integer,
    brout2 integer,
    brout3 integer,
    firstf integer,
    loc character varying(10),
    hittype character varying(5),
    dpopp integer,
    pivot integer,
    pn integer NOT NULL,
    umphome character varying(8),
    ump1b character varying(8),
    ump2b character varying(8),
    ump3b character varying(8),
    umplf character varying(8),
    umprf character varying(8),
    date character varying(8),
    gametype character varying(20),
    pbp character varying(10),
    PRIMARY KEY (gid, pn, date)
) PARTITION BY RANGE (date);

-- Step 2: Create partitions for each era/year

-- Pre-1914 sparse data (grouped)
CREATE TABLE plays_pre1914 PARTITION OF plays_partitioned
    FOR VALUES FROM ('00000000') TO ('19140000');

-- Federal League Era (yearly)
CREATE TABLE plays_1914 PARTITION OF plays_partitioned
    FOR VALUES FROM ('19140000') TO ('19150000');
CREATE TABLE plays_1915 PARTITION OF plays_partitioned
    FOR VALUES FROM ('19150000') TO ('19160000');

-- Inter-era sparse (grouped)
CREATE TABLE plays_1916_1934 PARTITION OF plays_partitioned
    FOR VALUES FROM ('19160000') TO ('19350000');

-- Negro Leagues Era (yearly - 1935-1949)
CREATE TABLE plays_1935 PARTITION OF plays_partitioned FOR VALUES FROM ('19350000') TO ('19360000');
CREATE TABLE plays_1936 PARTITION OF plays_partitioned FOR VALUES FROM ('19360000') TO ('19370000');
CREATE TABLE plays_1937 PARTITION OF plays_partitioned FOR VALUES FROM ('19370000') TO ('19380000');
CREATE TABLE plays_1938 PARTITION OF plays_partitioned FOR VALUES FROM ('19380000') TO ('19390000');
CREATE TABLE plays_1939 PARTITION OF plays_partitioned FOR VALUES FROM ('19390000') TO ('19400000');
CREATE TABLE plays_1940 PARTITION OF plays_partitioned FOR VALUES FROM ('19400000') TO ('19410000');
CREATE TABLE plays_1941 PARTITION OF plays_partitioned FOR VALUES FROM ('19410000') TO ('19420000');
CREATE TABLE plays_1942 PARTITION OF plays_partitioned FOR VALUES FROM ('19420000') TO ('19430000');
CREATE TABLE plays_1943 PARTITION OF plays_partitioned FOR VALUES FROM ('19430000') TO ('19440000');
CREATE TABLE plays_1944 PARTITION OF plays_partitioned FOR VALUES FROM ('19440000') TO ('19450000');
CREATE TABLE plays_1945 PARTITION OF plays_partitioned FOR VALUES FROM ('19450000') TO ('19460000');
CREATE TABLE plays_1946 PARTITION OF plays_partitioned FOR VALUES FROM ('19460000') TO ('19470000');
CREATE TABLE plays_1947 PARTITION OF plays_partitioned FOR VALUES FROM ('19470000') TO ('19480000');
CREATE TABLE plays_1948 PARTITION OF plays_partitioned FOR VALUES FROM ('19480000') TO ('19490000');
CREATE TABLE plays_1949 PARTITION OF plays_partitioned FOR VALUES FROM ('19490000') TO ('19500000');

-- Baby Boomer Era (era-grouped - 1950-1962)
CREATE TABLE plays_1950_1962 PARTITION OF plays_partitioned
    FOR VALUES FROM ('19500000') TO ('19630000');

-- Pitcher Era (era-grouped - 1963-1968)
CREATE TABLE plays_1963_1968 PARTITION OF plays_partitioned
    FOR VALUES FROM ('19630000') TO ('19690000');

-- Turf Time (5-year chunks - 1969-1993)
CREATE TABLE plays_1969_1973 PARTITION OF plays_partitioned FOR VALUES FROM ('19690000') TO ('19740000');
CREATE TABLE plays_1974_1978 PARTITION OF plays_partitioned FOR VALUES FROM ('19740000') TO ('19790000');
CREATE TABLE plays_1979_1983 PARTITION OF plays_partitioned FOR VALUES FROM ('19790000') TO ('19840000');
CREATE TABLE plays_1984_1988 PARTITION OF plays_partitioned FOR VALUES FROM ('19840000') TO ('19890000');
CREATE TABLE plays_1989_1993 PARTITION OF plays_partitioned FOR VALUES FROM ('19890000') TO ('19940000');

-- Steroid Era (yearly - 1994-2004)
CREATE TABLE plays_1994 PARTITION OF plays_partitioned FOR VALUES FROM ('19940000') TO ('19950000');
CREATE TABLE plays_1995 PARTITION OF plays_partitioned FOR VALUES FROM ('19950000') TO ('19960000');
CREATE TABLE plays_1996 PARTITION OF plays_partitioned FOR VALUES FROM ('19960000') TO ('19970000');
CREATE TABLE plays_1997 PARTITION OF plays_partitioned FOR VALUES FROM ('19970000') TO ('19980000');
CREATE TABLE plays_1998 PARTITION OF plays_partitioned FOR VALUES FROM ('19980000') TO ('19990000');
CREATE TABLE plays_1999 PARTITION OF plays_partitioned FOR VALUES FROM ('19990000') TO ('20000000');
CREATE TABLE plays_2000 PARTITION OF plays_partitioned FOR VALUES FROM ('20000000') TO ('20010000');
CREATE TABLE plays_2001 PARTITION OF plays_partitioned FOR VALUES FROM ('20010000') TO ('20020000');
CREATE TABLE plays_2002 PARTITION OF plays_partitioned FOR VALUES FROM ('20020000') TO ('20030000');
CREATE TABLE plays_2003 PARTITION OF plays_partitioned FOR VALUES FROM ('20030000') TO ('20040000');
CREATE TABLE plays_2004 PARTITION OF plays_partitioned FOR VALUES FROM ('20040000') TO ('20050000');

-- Moneyball Era (yearly - 2005-2012)
CREATE TABLE plays_2005 PARTITION OF plays_partitioned FOR VALUES FROM ('20050000') TO ('20060000');
CREATE TABLE plays_2006 PARTITION OF plays_partitioned FOR VALUES FROM ('20060000') TO ('20070000');
CREATE TABLE plays_2007 PARTITION OF plays_partitioned FOR VALUES FROM ('20070000') TO ('20080000');
CREATE TABLE plays_2008 PARTITION OF plays_partitioned FOR VALUES FROM ('20080000') TO ('20090000');
CREATE TABLE plays_2009 PARTITION OF plays_partitioned FOR VALUES FROM ('20090000') TO ('20100000');
CREATE TABLE plays_2010 PARTITION OF plays_partitioned FOR VALUES FROM ('20100000') TO ('20110000');
CREATE TABLE plays_2011 PARTITION OF plays_partitioned FOR VALUES FROM ('20110000') TO ('20120000');
CREATE TABLE plays_2012 PARTITION OF plays_partitioned FOR VALUES FROM ('20120000') TO ('20130000');

-- Statcast Era (yearly - 2013-2019)
CREATE TABLE plays_2013 PARTITION OF plays_partitioned FOR VALUES FROM ('20130000') TO ('20140000');
CREATE TABLE plays_2014 PARTITION OF plays_partitioned FOR VALUES FROM ('20140000') TO ('20150000');
CREATE TABLE plays_2015 PARTITION OF plays_partitioned FOR VALUES FROM ('20150000') TO ('20160000');
CREATE TABLE plays_2016 PARTITION OF plays_partitioned FOR VALUES FROM ('20160000') TO ('20170000');
CREATE TABLE plays_2017 PARTITION OF plays_partitioned FOR VALUES FROM ('20170000') TO ('20180000');
CREATE TABLE plays_2018 PARTITION OF plays_partitioned FOR VALUES FROM ('20180000') TO ('20190000');
CREATE TABLE plays_2019 PARTITION OF plays_partitioned FOR VALUES FROM ('20190000') TO ('20200000');

-- Modern Era (yearly - 2020-2030)
CREATE TABLE plays_2020 PARTITION OF plays_partitioned FOR VALUES FROM ('20200000') TO ('20210000');
CREATE TABLE plays_2021 PARTITION OF plays_partitioned FOR VALUES FROM ('20210000') TO ('20220000');
CREATE TABLE plays_2022 PARTITION OF plays_partitioned FOR VALUES FROM ('20220000') TO ('20230000');
CREATE TABLE plays_2023 PARTITION OF plays_partitioned FOR VALUES FROM ('20230000') TO ('20240000');
CREATE TABLE plays_2024 PARTITION OF plays_partitioned FOR VALUES FROM ('20240000') TO ('20250000');
CREATE TABLE plays_2025 PARTITION OF plays_partitioned FOR VALUES FROM ('20250000') TO ('20260000');
CREATE TABLE plays_2026 PARTITION OF plays_partitioned FOR VALUES FROM ('20260000') TO ('20270000');
CREATE TABLE plays_2027 PARTITION OF plays_partitioned FOR VALUES FROM ('20270000') TO ('20280000');
CREATE TABLE plays_2028 PARTITION OF plays_partitioned FOR VALUES FROM ('20280000') TO ('20290000');
CREATE TABLE plays_2029 PARTITION OF plays_partitioned FOR VALUES FROM ('20290000') TO ('20300000');
CREATE TABLE plays_2030 PARTITION OF plays_partitioned FOR VALUES FROM ('20300000') TO ('20310000');

CREATE TABLE plays_default PARTITION OF plays_partitioned DEFAULT;

INSERT INTO plays_partitioned SELECT * FROM plays;

-- Step 4: Create indexes on the partitioned table
-- These will automatically be created on all child partitions
CREATE INDEX idx_plays_partitioned_gid ON plays_partitioned(gid);
CREATE INDEX idx_plays_partitioned_gid_pn ON plays_partitioned(gid, pn);

-- Date index (critical for partition pruning)
CREATE INDEX idx_plays_partitioned_date ON plays_partitioned(date);

CREATE INDEX idx_plays_partitioned_batter ON plays_partitioned(batter);
CREATE INDEX idx_plays_partitioned_pitcher ON plays_partitioned(pitcher);
CREATE INDEX idx_plays_partitioned_batteam ON plays_partitioned(batteam);
CREATE INDEX idx_plays_partitioned_pitteam ON plays_partitioned(pitteam);

-- Step 5: Swap tables atomically
BEGIN;
    ALTER TABLE plays RENAME TO plays_old;
    ALTER TABLE plays_partitioned RENAME TO plays;
COMMIT;

-- Step 6: Analyze new table for query planner
ANALYZE plays;
