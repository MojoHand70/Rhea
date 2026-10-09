package store_test

import (
	"context"
	"fmt"
	"testing"

	"rhea/internal/core"
	"rhea/internal/store"
)

func BenchmarkStoreRoundTrips(b *testing.B) {
	ctx := context.Background()
	s, cleanup, err := store.Throwaway(ctx, "rhea_bench")
	if err != nil {
		b.Skip(err)
	}
	defer cleanup()
	ot := core.ObjectType{Name: "stock_movement", Version: 1, Domain: "w", Fields: []core.FieldDef{
		{Name: "item", Type: "string"}, {Name: "qty", Type: "int"}}}
	if err := s.InsertObjectType(ctx, ot); err != nil {
		b.Fatal(err)
	}
	tx, _ := s.Pool.Begin(ctx)
	for i := 0; i < 1000; i++ {
		id := fmt.Sprintf("stock_movement-%d", i)
		if err := store.InsertObject(ctx, tx, core.Object{ID: id, Type: "stock_movement", TypeVersion: 1,
			State: map[string]any{"item": "item-1", "qty": i}, SourceEventID: int64(i + 1), RuleID: "r", RuleVersion: 1}); err != nil {
			b.Fatal(err)
		}
	}
	tx.Commit(ctx)
	b.Run("GetObjectType", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := s.GetObjectType(ctx, "stock_movement"); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("GetObject", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := s.GetObject(ctx, fmt.Sprintf("stock_movement-%d", i%1000)); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("FindObjectIDsByField_1000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ids, err := s.FindObjectIDsByField(ctx, "stock_movement", "item", "item-1")
			if err != nil || len(ids) != 1000 {
				b.Fatal(err, len(ids))
			}
		}
	})
	b.Run("ObjectsByType_1000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			objs, err := s.ObjectsByType(ctx, "stock_movement")
			if err != nil || len(objs) != 1000 {
				b.Fatal(err, len(objs))
			}
		}
	})
}
