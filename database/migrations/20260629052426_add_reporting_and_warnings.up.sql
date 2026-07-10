DROP TABLE IF EXISTS report;

CREATE TYPE report_entity_type AS ENUM ('USER', 'ITEM', 'CHAT');
CREATE TYPE report_type AS ENUM ('SPAM', 'INAPPROPRIATE', 'FRAUD', 'OTHER');
CREATE TYPE report_status AS ENUM ('PENDING', 'RESOLVED', 'DISMISSED');

CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id UUID NOT NULL,
    reported_entity_id UUID NOT NULL,
    entity_type report_entity_type NOT NULL,
    type report_type NOT NULL,
    description TEXT,
    attachment_url VARCHAR(1024),
    status report_status NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE user_warnings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    admin_id UUID NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
