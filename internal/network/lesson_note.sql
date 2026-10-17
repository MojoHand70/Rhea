-- The team's note on one lesson. $1 lesson key, $2 status, $3 note, $4 who.
INSERT INTO lesson_note (lesson, status, note, noted_by) VALUES (lower(btrim($1)), $2, $3, $4)
RETURNING noted_at
