package models

import (
	"testing"
)

func TestCalculatePagination(t *testing.T) {
	tests := []struct {
		name       string
		totalItems int64
		page       int
		limit      int
		wantPage   int
		wantLimit  int
		wantTotalP int
	}{
		{
			name:       "zero items",
			totalItems: 0,
			page:       1,
			limit:      10,
			wantPage:   1,
			wantLimit:  10,
			wantTotalP: 0,
		},
		{
			name:       "exact match",
			totalItems: 20,
			page:       1,
			limit:      10,
			wantPage:   1,
			wantLimit:  10,
			wantTotalP: 2,
		},
		{
			name:       "partial page",
			totalItems: 25,
			page:       2,
			limit:      10,
			wantPage:   2,
			wantLimit:  10,
			wantTotalP: 3,
		},
		{
			name:       "invalid page reset to 1",
			totalItems: 25,
			page:       -1,
			limit:      10,
			wantPage:   1,
			wantLimit:  10,
			wantTotalP: 3,
		},
		{
			name:       "all items (limit <= 0)",
			totalItems: 42,
			page:       1,
			limit:      0,
			wantPage:   1,
			wantLimit:  42,
			wantTotalP: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculatePagination(tt.totalItems, tt.page, tt.limit)
			if got.Page != tt.wantPage {
				t.Errorf("got Page %d, want %d", got.Page, tt.wantPage)
			}
			if got.Limit != tt.wantLimit {
				t.Errorf("got Limit %d, want %d", got.Limit, tt.wantLimit)
			}
			if got.TotalPages != tt.wantTotalP {
				t.Errorf("got TotalPages %d, want %d", got.TotalPages, tt.wantTotalP)
			}
			if got.TotalItems != tt.totalItems {
				t.Errorf("got TotalItems %d, want %d", got.TotalItems, tt.totalItems)
			}
		})
	}
}

