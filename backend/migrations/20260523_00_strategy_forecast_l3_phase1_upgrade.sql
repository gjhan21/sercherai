ALTER TABLE strategy_forecast_l3_reports
    ADD COLUMN headline_verdict TEXT NULL AFTER version,
    ADD COLUMN state_assessment_json JSON NULL AFTER primary_scenario,
    ADD COLUMN dimension_evidence_json JSON NULL AFTER state_assessment_json,
    ADD COLUMN scenario_assessment_json JSON NULL AFTER dimension_evidence_json,
    ADD COLUMN validation_review_json JSON NULL AFTER scenario_assessment_json;
