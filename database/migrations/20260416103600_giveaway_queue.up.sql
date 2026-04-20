CREATE TABLE IF NOT EXISTS queue_entry (
    entry_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    item_id uuid NOT NULL REFERENCES item(item_id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES "User"(user_id) ON DELETE CASCADE,
    joined_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    entry_status character varying(20) DEFAULT 'WAITING' NOT NULL,
    turn_started_at timestamp without time zone,
    UNIQUE(item_id, user_id)
);

CREATE TABLE IF NOT EXISTS purchase_request (
    request_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    item_id uuid NOT NULL REFERENCES item(item_id) ON DELETE CASCADE,
    buyer_id uuid NOT NULL REFERENCES "User"(user_id) ON DELETE CASCADE,
    status character varying(20) DEFAULT 'PENDING' NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS transaction (
    transaction_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    item_id uuid NOT NULL REFERENCES item(item_id) ON DELETE CASCADE,
    buyer_id uuid NOT NULL REFERENCES "User"(user_id),
    seller_id uuid NOT NULL REFERENCES "User"(user_id),
    status character varying(20) DEFAULT 'PENDING' NOT NULL,
    seller_confirmed boolean DEFAULT false NOT NULL,
    buyer_confirmed boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE(item_id)
);
