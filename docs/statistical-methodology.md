
# Statistical Methodology

## Advanced Stats Implementation

Our advanced statistics endpoints (`/v1/players/{id}/stats/batting/advanced`, `/v1/players/{id}/stats/war`, etc.) use industry-standard formulas with some simplifications. Here's how our calculations compare to popular sources:

### Batting Stats (wOBA, wRC+)

**Implementation**: Uses FanGraphs constants and formulas

- **wOBA**: Matches FanGraphs exactly (same weights and scale)
- **wRC+**: Matches FanGraphs methodology (park and league adjusted)
- **Data source**: FanGraphs Guts! constants (1871-2025)

**Expected accuracy**: Should closely match FanGraphs values

### WAR (Wins Above Replacement)

**Implementation**: Hybrid approach using Lahman + FanGraphs constants

Our WAR calculation differs from both FanGraphs (fWAR) and Baseball Reference (bWAR):

**Components:**

1. **Batting Runs**: wRAA (weighted runs above average) - Matches FanGraphs ✓
2. **Base Running**: $wSB$ only ($SB × run_{sb} + CS × run_{cs}$)
   - **FanGraphs**: Includes UBR (ultimate base running) from play-by-play
   - **Baseball Reference**: More comprehensive base running metrics
   - **Our approach**: Simplified - stolen bases only
3. **Fielding Runs**: Range factor vs league average
   - **FanGraphs**: Uses UZR (Ultimate Zone Rating)
   - **Baseball Reference**: Uses Total Zone and DRS
   - **Our approach**: Simplified - (Player RF - League Avg RF) × Games × 0.1
   - **Impact**: Our fielding values will differ significantly from both sources
4. **Positional Adjustment**: Uses FanGraphs standard adjustments ✓
5. **Replacement Level**: -20 runs per 600 PA (simplified)
6. **Runs per Win**: Uses FanGraphs season-specific r_w values ✓

**Expected differences:**

- Our WAR will generally be **close but not identical** to fWAR/bWAR
- Largest variance in fielding component (we use simplified range factor)
- Base running component less comprehensive (no UBR/extra base advancement)
- For players with exceptional defense or base running, expect 0.5-2.0 WAR difference

## Fielding Stats

**Implementation**: Range factor methodology

- **Formula**: $\frac{(PO + A)}{Games}$, compared to league average at position
- **Runs conversion**: 0.1 runs per play above/below average
- **FanGraphs**: Uses UZR (requires batted ball tracking data)
- **Baseball Reference**: Uses Total Zone Rating

**Expected accuracy**: Directionally correct but less precise than UZR/DRS

## Base Running Stats

**Implementation**: wSB (weighted stolen bases)

- **Formula**: (SB × run_sb) + (CS × run_cs) using FanGraphs constants
- **FanGraphs**: Also includes UBR (extra bases, advancement, outs on bases)
- **Baseball Reference**: Comprehensive base running runs

**Expected accuracy**: Captures ~60-70% of total base running value (stolen bases only)

## Data Coverage

- **wOBA/wRC+ constants**: 1871-2025 (155 years from FanGraphs)
- **Park factors**: 2016-2025 (FanGraphs 5-year regressed)
- **League constants**: Currently 2023-2024 (expandable to full Lahman range)
- **WAR calculations**: Available for any Lahman season (1871-2024)

[FanGraphs](https://www.fangraphs.com) and [Baseball Reference](https://www.baseball-reference.com) are still better.
