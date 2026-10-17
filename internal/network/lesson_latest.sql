-- Each lesson's latest note: where it stands now.
SELECT DISTINCT ON (lesson) lesson, status, note, noted_by, noted_at
FROM lesson_note
ORDER BY lesson, lesson_note_id DESC
