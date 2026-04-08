package postgres

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func toUUID(id string) pgtype.UUID {
	var u pgtype.UUID
	_ = u.Scan(id)
	return u
}

func fromUUID(u pgtype.UUID) string {
	return u.String()
}

func toTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:  t,
		Valid: true,
	}
}

func toNullableTimestamp(t *time.Time) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{
		Time:  *t,
		Valid: true,
	}
}

func fromTimestamp(t pgtype.Timestamp) time.Time {
	return t.Time
}

func fromNullableTimestamp(t pgtype.Timestamp) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
func toText(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  true,
	}
}

func toNullableText(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *v, Valid: true}
}

func toNullableInt32(v *int) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: int32(*v), Valid: true}
}

func uuidToString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return id.String()
}

func fromNullableText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func fromNullableTime(t pgtype.Timestamp) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func fromNullableInt32(v pgtype.Int4) int32 {
	if !v.Valid {
		return 0
	}
	return v.Int32
}

func fromNullableBool(v pgtype.Bool) bool {
	if !v.Valid {
		return false
	}
	return v.Bool
}

func toNullableTimePtr(t *time.Time) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: *t, Valid: true}
}

func toNullableInt32Ptr(v *int) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: int32(*v), Valid: true}
}
