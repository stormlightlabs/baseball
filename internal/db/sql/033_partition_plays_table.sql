-- Partition the plays table by year to improve query performance
-- Idempotent: skips when plays is already partitioned.

DO $$
DECLARE
    part record;
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_partitioned_table pt
        INNER JOIN pg_class c ON c.oid = pt.partrelid
        WHERE c.relname = 'plays'
    ) THEN
        RAISE NOTICE 'plays is already partitioned; skipping migration 033';
        RETURN;
    END IF;

    -- Step 1: Create new partitioned table with same schema as plays
    EXECUTE $ddl$
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
        ) PARTITION BY RANGE (date)
    $ddl$;

    -- Step 2: Create partitions for each era/year
    FOR part IN
        SELECT *
        FROM (
            VALUES
                ('plays_pre1914', '00000000', '19140000'),
                ('plays_1914', '19140000', '19150000'),
                ('plays_1915', '19150000', '19160000'),
                ('plays_1916_1934', '19160000', '19350000'),
                ('plays_1935', '19350000', '19360000'),
                ('plays_1936', '19360000', '19370000'),
                ('plays_1937', '19370000', '19380000'),
                ('plays_1938', '19380000', '19390000'),
                ('plays_1939', '19390000', '19400000'),
                ('plays_1940', '19400000', '19410000'),
                ('plays_1941', '19410000', '19420000'),
                ('plays_1942', '19420000', '19430000'),
                ('plays_1943', '19430000', '19440000'),
                ('plays_1944', '19440000', '19450000'),
                ('plays_1945', '19450000', '19460000'),
                ('plays_1946', '19460000', '19470000'),
                ('plays_1947', '19470000', '19480000'),
                ('plays_1948', '19480000', '19490000'),
                ('plays_1949', '19490000', '19500000'),
                ('plays_1950_1962', '19500000', '19630000'),
                ('plays_1963_1968', '19630000', '19690000'),
                ('plays_1969_1973', '19690000', '19740000'),
                ('plays_1974_1978', '19740000', '19790000'),
                ('plays_1979_1983', '19790000', '19840000'),
                ('plays_1984_1988', '19840000', '19890000'),
                ('plays_1989_1993', '19890000', '19940000'),
                ('plays_1994', '19940000', '19950000'),
                ('plays_1995', '19950000', '19960000'),
                ('plays_1996', '19960000', '19970000'),
                ('plays_1997', '19970000', '19980000'),
                ('plays_1998', '19980000', '19990000'),
                ('plays_1999', '19990000', '20000000'),
                ('plays_2000', '20000000', '20010000'),
                ('plays_2001', '20010000', '20020000'),
                ('plays_2002', '20020000', '20030000'),
                ('plays_2003', '20030000', '20040000'),
                ('plays_2004', '20040000', '20050000'),
                ('plays_2005', '20050000', '20060000'),
                ('plays_2006', '20060000', '20070000'),
                ('plays_2007', '20070000', '20080000'),
                ('plays_2008', '20080000', '20090000'),
                ('plays_2009', '20090000', '20100000'),
                ('plays_2010', '20100000', '20110000'),
                ('plays_2011', '20110000', '20120000'),
                ('plays_2012', '20120000', '20130000'),
                ('plays_2013', '20130000', '20140000'),
                ('plays_2014', '20140000', '20150000'),
                ('plays_2015', '20150000', '20160000'),
                ('plays_2016', '20160000', '20170000'),
                ('plays_2017', '20170000', '20180000'),
                ('plays_2018', '20180000', '20190000'),
                ('plays_2019', '20190000', '20200000'),
                ('plays_2020', '20200000', '20210000'),
                ('plays_2021', '20210000', '20220000'),
                ('plays_2022', '20220000', '20230000'),
                ('plays_2023', '20230000', '20240000'),
                ('plays_2024', '20240000', '20250000'),
                ('plays_2025', '20250000', '20260000'),
                ('plays_2026', '20260000', '20270000'),
                ('plays_2027', '20270000', '20280000'),
                ('plays_2028', '20280000', '20290000'),
                ('plays_2029', '20290000', '20300000'),
                ('plays_2030', '20300000', '20310000')
        ) AS partitions(partition_name, range_start, range_end)
    LOOP
        EXECUTE format(
            'CREATE TABLE %I PARTITION OF plays_partitioned FOR VALUES FROM (%L) TO (%L)',
            part.partition_name,
            part.range_start,
            part.range_end
        );
    END LOOP;

    EXECUTE 'CREATE TABLE plays_default PARTITION OF plays_partitioned DEFAULT';

    -- Step 3: Migrate existing data.
    EXECUTE 'INSERT INTO plays_partitioned SELECT * FROM plays';

    -- Step 4: Create indexes on the partitioned table.
    -- These will automatically be created on all child partitions.
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_plays_partitioned_gid ON plays_partitioned(gid)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_plays_partitioned_gid_pn ON plays_partitioned(gid, pn)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_plays_partitioned_date ON plays_partitioned(date)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_plays_partitioned_batter ON plays_partitioned(batter)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_plays_partitioned_pitcher ON plays_partitioned(pitcher)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_plays_partitioned_batteam ON plays_partitioned(batteam)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_plays_partitioned_pitteam ON plays_partitioned(pitteam)';

    -- Step 5: Swap tables atomically. Migration runner already wraps this migration in a transaction.
    IF to_regclass('plays_old') IS NOT NULL THEN
        EXECUTE 'DROP TABLE plays_old CASCADE';
    END IF;

    EXECUTE 'ALTER TABLE plays RENAME TO plays_old';
    EXECUTE 'ALTER TABLE plays_partitioned RENAME TO plays';

    -- Step 6: Analyze new table for query planner.
    EXECUTE 'ANALYZE plays';
END
$$;
