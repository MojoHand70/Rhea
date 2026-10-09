-- A bundle's latest rejection, as the reject_bundle door recorded it:
-- the reason, who refused, and the business date.
SELECT payload->>'reason', payload->>'rejected_by', to_char(occurred_at, 'YYYY-MM-DD')
FROM event
WHERE event_type = 'bundle.rejected' AND payload->>'bundle_id' = $1
ORDER BY event_id DESC
LIMIT 1
