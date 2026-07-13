CREATE TABLE public.password_reset (
    email character varying(255) NOT NULL,
    token character varying(6) NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    PRIMARY KEY (email)
);
