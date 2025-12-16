package seed

import "fmt"

// Era represents a named period in baseball history with associated year range
type Era struct {
	Name      string
	ShortName string
	StartYear int
	EndYear   int
	Notes     string
}

// GetAllEras returns all defined baseball eras in chronological order
func GetAllEras() []Era {
	return []Era{
		{
			Name: "Federal League Era", ShortName: "fed",
			StartYear: 1914, EndYear: 1915,
			Notes: "Federal League games (third major league)",
		},
		{
			Name: "Negro Leagues Era", ShortName: "nlg",
			StartYear: 1935, EndYear: 1949,
			Notes: "Negro Leagues games available in Retrosheet",
		},
		{
			Name: "Baby Boomer Era", ShortName: "boomer",
			StartYear: 1950, EndYear: 1962,
			Notes: "Post-war expansion and integration",
		},
		{
			Name: "Pitcher Era", ShortName: "pitcher",
			StartYear: 1963, EndYear: 1968,
			Notes: "Dominant pitching before mound lowering",
		},
		{
			Name: "Turf Time", ShortName: "turf",
			StartYear: 1969, EndYear: 1993,
			Notes: "Artificial turf, expansion, and free agency",
		},
		{
			Name: "Steroid Era", ShortName: "steroid",
			StartYear: 1994, EndYear: 2004,
			Notes: "Enhanced performance and home run records",
		},
		{
			Name: "Moneyball Era", ShortName: "moneyball",
			StartYear: 2005, EndYear: 2012,
			Notes: "Analytics revolution and sabermetrics adoption",
		},
		{
			Name: "Statcast Era", ShortName: "statcast",
			StartYear: 2013, EndYear: 2019,
			Notes: "Launch angle revolution and tracking data",
		},
		{
			Name: "Modern Era", ShortName: "modern",
			StartYear: 2020, EndYear: 2025,
			Notes: "Pitch clock and pace of play changes",
		},
	}
}

// GetEra returns an era by its short name
func GetEra(shortName string) *Era {
	for _, era := range GetAllEras() {
		if era.ShortName == shortName {
			return &era
		}
	}
	return nil
}

// Years returns all years in the era range (inclusive)
func (e Era) Years() []int {
	years := make([]int, 0, e.EndYear-e.StartYear+1)
	for y := e.StartYear; y <= e.EndYear; y++ {
		years = append(years, y)
	}
	return years
}

// GetYearsForEras returns all years covered by the specified eras
func GetYearsForEras(eraNames []string) []int {
	yearSet := make(map[int]bool)

	for _, name := range eraNames {
		era := GetEra(name)
		if era != nil {
			for _, year := range era.Years() {
				yearSet[year] = true
			}
		}
	}

	years := make([]int, 0, len(yearSet))
	for year := range yearSet {
		years = append(years, year)
	}

	for i := 0; i < len(years); i++ {
		for j := i + 1; j < len(years); j++ {
			if years[i] > years[j] {
				years[i], years[j] = years[j], years[i]
			}
		}
	}

	return years
}

// GetErasForYear returns all eras that include the specified year
func GetErasForYear(year int) []Era {
	var eras []Era
	for _, era := range GetAllEras() {
		if year >= era.StartYear && year <= era.EndYear {
			eras = append(eras, era)
		}
	}
	return eras
}

// ListEras returns a formatted list of all eras for display
func ListEras() string {
	var result string
	for _, era := range GetAllEras() {
		result += fmt.Sprintf("%s (%s): %d-%d - %s\n", era.Name, era.ShortName, era.StartYear, era.EndYear, era.Notes)
	}
	return result
}

// DateRange represents a date range in YYYYMMDD format
type DateRange struct {
	From string
	To   string
}

// LeagueDateRanges maps league codes to their historical active date ranges
// This enables partition pruning for league-specific queries
// Modern leagues (AL, NL) are intentionally omitted - they span the entire dataset (1876-2025)
var LeagueDateRanges = map[string]DateRange{
	// 19th Century Leagues & Early 20th Century Leagues
	"UA": {From: "18840101", To: "18841231"}, // Union Association (1884)
	"PL": {From: "18900101", To: "18901231"}, // Players League (1890)
	"AA": {From: "18820101", To: "18911231"}, // American Association (1882-1891)
	"FL": {From: "19140101", To: "19151231"}, // Federal League (1914-1915)

	// Negro Leagues
	"NAL": {From: "19350101", To: "19491231"}, // Negro American League
	"NNL": {From: "19350101", To: "19491231"}, // Negro National League
	"NN2": {From: "19350101", To: "19491231"}, // Negro National League (second)
	"ECL": {From: "19350101", To: "19491231"}, // East Coast League
	"ANL": {From: "19350101", To: "19491231"}, // American Negro League
	"EWL": {From: "19350101", To: "19491231"}, // East-West League
	"NSL": {From: "19350101", To: "19491231"}, // Negro Southern League
	"IND": {From: "19350101", To: "19491231"}, // Independent Negro League teams
}

// GetLeagueDateRange returns the active date range for a set of leagues
// This enables partition pruning by adding implicit date filters to the query
func GetLeagueDateRange(leagues []string) *DateRange {
	if len(leagues) == 0 {
		return nil
	}

	for _, league := range leagues {
		if league == "AL" || league == "NL" {
			return nil
		}
	}

	if len(leagues) == 1 {
		if dateRange, exists := LeagueDateRanges[leagues[0]]; exists {
			return &dateRange
		}
		return nil
	}

	var firstRange *DateRange
	for _, league := range leagues {
		dateRange, exists := LeagueDateRanges[league]
		if !exists {
			return nil
		}

		if firstRange == nil {
			firstRange = &dateRange
		} else if dateRange.From != firstRange.From || dateRange.To != firstRange.To {
			return nil
		}
	}

	return firstRange
}
