BEGIN;

DROP TRIGGER IF EXISTS trg_update_educator_subscriber_count ON subscriptions;
DROP FUNCTION IF EXISTS update_educator_subscriber_count();
DROP TABLE IF EXISTS subscriptions CASCADE;

COMMIT;
