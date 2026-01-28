BEGIN;

CREATE TABLE IF NOT EXISTS educator_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    bio TEXT,
    tagline VARCHAR(255),
    expertise_areas TEXT[], -- Array of expertise tags
    credentials TEXT, -- JSON array of credentials/certifications
    social_links JSONB DEFAULT '{}'::jsonb, -- LinkedIn, Twitter, website, etc.
    profile_banner_url TEXT,
    is_featured BOOLEAN DEFAULT false,
    is_verified BOOLEAN DEFAULT false, -- Platform verified badge
    total_subscribers INT DEFAULT 0,
    total_quizzes INT DEFAULT 0,
    average_rating DECIMAL(3,2) DEFAULT 0.0, -- Future: rating system
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_educator_profiles_user_id ON educator_profiles(user_id);
CREATE INDEX idx_educator_profiles_featured ON educator_profiles(is_featured) WHERE is_featured = true;
CREATE INDEX idx_educator_profiles_expertise ON educator_profiles USING GIN (expertise_areas);

COMMENT ON TABLE educator_profiles IS 'Extended profiles for users with educator role';
COMMENT ON COLUMN educator_profiles.expertise_areas IS 'Array of expertise topics/tags';
COMMENT ON COLUMN educator_profiles.social_links IS 'JSON object with social media links';
COMMENT ON COLUMN educator_profiles.is_featured IS 'Featured educators shown prominently';
COMMENT ON COLUMN educator_profiles.is_verified IS 'Platform-verified educator badge';

COMMIT;
