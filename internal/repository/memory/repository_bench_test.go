package memory

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/newmersedez/urlshort/internal/model"
)

func BenchmarkAdd(b *testing.B) {
	repo, _ := NewMemoryRepository()
	userID := uuid.New()
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("%08d", i)
		url := model.NewShortenURL(userID, id, "https://example.com/"+id)
		repo.Add(ctx, url) 
	}
}

func BenchmarkGet(b *testing.B) {
	repo, _ := NewMemoryRepository()
	userID := uuid.New()
	ctx := context.Background()

	const count = 1000
	for i := 0; i < count; i++ {
		id := fmt.Sprintf("%08d", i)
		url := model.NewShortenURL(userID, id, "https://example.com/"+id)
		repo.Add(ctx, url)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("%08d", i%count)
		repo.Get(ctx, id) 
	}
}

func BenchmarkGetList(b *testing.B) {
	repo, _ := NewMemoryRepository()
	userID := uuid.New()
	otherID := uuid.New()
	ctx := context.Background()

	const count = 500
	for i := 0; i < count; i++ {
		id := fmt.Sprintf("%08d", i)
		repo.Add(ctx, model.NewShortenURL(userID, id, "https://example.com/"+id))          
		repo.Add(ctx, model.NewShortenURL(otherID, "x"+id, "https://other.com/"+id))       
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.GetList(ctx, userID) 
	}
}
