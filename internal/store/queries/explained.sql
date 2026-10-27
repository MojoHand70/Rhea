-- Which of these events a rule has acted on since: those some event names as
-- its cause — the waiting facts the executor remembers, asked again.
SELECT DISTINCT cause_event_id FROM event WHERE cause_event_id = ANY($1)
