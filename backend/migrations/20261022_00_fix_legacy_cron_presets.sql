-- Migration: Replace legacy scheduler preset labels with real cron expressions.
-- Date: 2026-10-22
--
-- Early seeds stored schedule presets such as 'EVERY_30_MINUTES' in
-- cron_expr. The scheduler feeds that column straight into cron.AddFunc with a
-- six-field (seconds) parser, so registration failed for those rows and the
-- jobs never ran. The values below match the translation the scheduler now
-- applies at load time, which keeps fresh installs and already-migrated
-- databases on the same schedule.
--
-- EVERY_100_MINUTES has no exact cron equivalent (100 minutes does not divide
-- an hour); it is normalised to a two-hour cadence.

UPDATE scheduler_job_definitions
   SET cron_expr = '0 */30 * * * *', updated_at = NOW()
 WHERE job_name = 'vip_membership_lifecycle'
   AND cron_expr = 'EVERY_30_MINUTES';

UPDATE scheduler_job_definitions
   SET cron_expr = '0 */20 * * * *', updated_at = NOW()
 WHERE job_name = 'tushare_news_incremental'
   AND cron_expr = 'EVERY_20_MINUTES';

UPDATE scheduler_job_definitions
   SET cron_expr = '0 0 */2 * * *', updated_at = NOW()
 WHERE job_name = 'doc_fast_news_incremental'
   AND cron_expr = 'EVERY_100_MINUTES';
