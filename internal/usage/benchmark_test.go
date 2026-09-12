package usage

import (
	"encoding/json"
	"testing"
	"time"
)

func makeSampleReport() Report {
	now := time.Now()
	return Report{
		Agent:     "agy",
		Profile:   "work",
		Status:    StatusOK,
		Summary:   "5h: 82%, wk: 90%",
		FetchedAt: now,
		Windows: []LimitWindow{
			{
				Category:     "Claude & GPT",
				Name:         "Five Hour Limit Remaining",
				RemainingPct: 82,
				ResetsAt:     now.Add(2 * time.Hour),
			},
			{
				Category:     "Claude & GPT",
				Name:         "Weekly Limit Remaining",
				RemainingPct: 90,
				ResetsAt:     now.Add(5 * 24 * time.Hour),
			},
			{
				Category:     "Gemini",
				Name:         "Daily Pro Requests Remaining",
				RemainingPct: 55,
				ResetsAt:     now.Add(12 * time.Hour),
			},
		},
	}
}

func BenchmarkCacheStore_Get_Hit(b *testing.B) {
	tmpDir := b.TempDir()
	store := NewCacheStore(tmpDir, 1*time.Hour)
	rep := makeSampleReport()
	_ = store.Put(rep)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cached, ok := store.Get("agy", "work")
		if !ok || len(cached.Windows) != 3 {
			b.Fatalf("expected hit with 3 windows")
		}
	}
}

func BenchmarkCacheStore_Get_Miss(b *testing.B) {
	tmpDir := b.TempDir()
	store := NewCacheStore(tmpDir, 1*time.Hour)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, ok := store.Get("agy", "nonexistent")
		if ok {
			b.Fatalf("expected miss")
		}
	}
}

func BenchmarkCacheStore_Put(b *testing.B) {
	tmpDir := b.TempDir()
	store := NewCacheStore(tmpDir, 1*time.Hour)
	rep := makeSampleReport()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := store.Put(rep); err != nil {
			b.Fatalf("Put failed: %v", err)
		}
	}
}

func BenchmarkDiskCache_Marshal(b *testing.B) {
	d := diskCacheData{
		Version:   "1",
		UpdatedAt: time.Now(),
		Reports: map[string]Report{
			"agy:work":     makeSampleReport(),
			"agy:personal": makeSampleReport(),
			"gemini:code":  makeSampleReport(),
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		data, err := json.MarshalIndent(d, "", "  ")
		if err != nil || len(data) == 0 {
			b.Fatalf("MarshalIndent failed: %v", err)
		}
	}
}

func BenchmarkDiskCache_Unmarshal(b *testing.B) {
	d := diskCacheData{
		Version:   "1",
		UpdatedAt: time.Now(),
		Reports: map[string]Report{
			"agy:work":     makeSampleReport(),
			"agy:personal": makeSampleReport(),
			"gemini:code":  makeSampleReport(),
		},
	}
	data, _ := json.MarshalIndent(d, "", "  ")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var decoded diskCacheData
		if err := json.Unmarshal(data, &decoded); err != nil {
			b.Fatalf("Unmarshal failed: %v", err)
		}
	}
}

func BenchmarkFormatDuration(b *testing.B) {
	durations := []time.Duration{
		0,
		45 * time.Minute,
		2*time.Hour + 30*time.Minute,
		3*24*time.Hour + 4*time.Hour,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d := durations[i%len(durations)]
		s := FormatDuration(d)
		if s == "" {
			b.Fatalf("empty duration string")
		}
	}
}

func BenchmarkFormatWindowSummary(b *testing.B) {
	win := LimitWindow{
		Name:         "Five Hour Limit Remaining",
		RemainingPct: 82,
		ResetsIn:     2*time.Hour + 15*time.Minute,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := FormatWindowSummary(&win)
		if s == "" {
			b.Fatalf("empty window summary")
		}
	}
}

func BenchmarkModelGroups(b *testing.B) {
	rep := makeSampleReport()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		groups := rep.ModelGroups()
		if len(groups) != 2 {
			b.Fatalf("expected 2 groups, got %d", len(groups))
		}
	}
}
