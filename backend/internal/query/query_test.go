package query

import "testing"

func TestNormalizeSortDefaultsAndValidates(t *testing.T) {
	allowed := map[string]string{
		"created_at": "created_at",
		"name":       "name",
	}

	sort, err := NormalizeSort("", "", allowed, Sort{By: "created_at", Order: Descending})
	if err != nil {
		t.Fatalf("normalize default sort: %v", err)
	}
	if sort.By != "created_at" || sort.Order != Descending {
		t.Fatalf("unexpected default sort: %+v", sort)
	}

	sort, err = NormalizeSort("name", "ASC", allowed, Sort{By: "created_at", Order: Descending})
	if err != nil {
		t.Fatalf("normalize explicit sort: %v", err)
	}
	if sort.By != "name" || sort.Order != Ascending {
		t.Fatalf("unexpected explicit sort: %+v", sort)
	}

	if _, err := NormalizeSort("password", "asc", allowed, Sort{By: "created_at", Order: Descending}); err == nil {
		t.Fatal("expected invalid sort_by error")
	}

	if _, err := NormalizeSort("name", "sideways", allowed, Sort{By: "created_at", Order: Descending}); err == nil {
		t.Fatal("expected invalid sort_order error")
	}
}

func TestOrderClauseUsesAllowlistedColumn(t *testing.T) {
	allowed := map[string]string{"created_at": "users.created_at"}
	got := OrderClause(Sort{By: "created_at", Order: Descending}, allowed)
	if got != "users.created_at DESC" {
		t.Fatalf("unexpected order clause: %q", got)
	}
}
