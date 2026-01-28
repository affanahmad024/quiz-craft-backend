BEGIN;

CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscriber_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    educator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscribed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    notification_enabled BOOLEAN DEFAULT true,

    UNIQUE (subscriber_id, educator_id),
    CHECK (subscriber_id != educator_id) -- Can't subscribe to yourself
);

CREATE INDEX idx_subscriptions_subscriber_id ON subscriptions(subscriber_id);
CREATE INDEX idx_subscriptions_educator_id ON subscriptions(educator_id);
CREATE INDEX idx_subscriptions_subscribed_at ON subscriptions(subscribed_at DESC);

COMMENT ON TABLE subscriptions IS 'Follow/subscription relationship between users and educators';
COMMENT ON COLUMN subscriptions.notification_enabled IS 'Whether subscriber receives notifications for this educator';

-- Trigger to update educator total_subscribers count
CREATE OR REPLACE FUNCTION update_educator_subscriber_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE educator_profiles
        SET total_subscribers = total_subscribers + 1,
            updated_at = NOW()
        WHERE user_id = NEW.educator_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE educator_profiles
        SET total_subscribers = GREATEST(total_subscribers - 1, 0),
            updated_at = NOW()
        WHERE user_id = OLD.educator_id;
    END IF
;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_educator_subscriber_count
AFTER INSERT OR DELETE ON subscriptions
FOR EACH ROW EXECUTE FUNCTION update_educator_subscriber_count();

COMMENT ON FUNCTION update_educator_subscriber_count() IS 'Automatically updates educator_profiles.total_subscribers when subscriptions change';

COMMIT;
