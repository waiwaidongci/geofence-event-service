package terminal

import "strings"

type Filter struct {
	Status Status
	Name   string
	Tags   map[string]string
	Limit  int
	Offset int
}

func (f Filter) Matches(value Terminal) bool {
	if f.Status != "" && value.Status != f.Status {
		return false
	}
	if name := strings.ToLower(strings.TrimSpace(f.Name)); name != "" {
		candidate := strings.ToLower(value.Name + " " + value.ID)
		if !strings.Contains(candidate, name) {
			return false
		}
	}
	return MatchTags(value.Tags, f.Tags)
}

func (f Filter) Page(items []Terminal) []Terminal {
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	if offset >= len(items) {
		return []Terminal{}
	}
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}
