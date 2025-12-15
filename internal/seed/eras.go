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
			Name: "1970s", ShortName: "1970s",
			StartYear: 1970, EndYear: 1979,
			Notes: "Expansion era and free agency begins",
		},
		{
			Name: "1980s", ShortName: "1980s",
			StartYear: 1980, EndYear: 1989,
			Notes: "Rise of power hitting and offensive explosion",
		},
		{
			Name: "Steroid Era", ShortName: "steroid",
			StartYear: 1990, EndYear: 2010,
			Notes: "Enhanced performance and home run records",
		},
		{
			Name: "Modern Era", ShortName: "modern",
			StartYear: 2011, EndYear: 2025,
			Notes: "Analytics-driven baseball and pitch clock",
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
