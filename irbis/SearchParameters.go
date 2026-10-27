package irbis

type SearchParameters struct {
	Database        string
	FirstRecord     int
	Format          string
	MaxMfn          int
	MinMfn          int
	NumberOfRecords int
	Expression      string
	Sequential      string
	Filter          string
	IsUtf           bool
}

func NewSearchParameters() *SearchParameters {
	return &SearchParameters{
		FirstRecord: 1,
	}
}

// SearchOutcome is the result of an extended search: total matches on the
// server plus the page of found lines (MFN and optional formatted text).
type SearchOutcome struct {
	Total int
	Found []FoundLine
}

// SearchReadOutcome is SearchRead with total match count.
type SearchReadOutcome struct {
	Total   int
	Records []MarcRecord
}
