-- The ids of Rhea's own rules: those no company owns. A rule's id keeps its
-- company through every version (rule_company_fixed), so one row says it.
SELECT DISTINCT rule_id FROM rule WHERE COALESCE(company, '') = ''
