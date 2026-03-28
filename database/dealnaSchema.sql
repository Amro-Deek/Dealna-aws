--
-- PostgreSQL database dump
--

-- Dumped from database version 16.13 (Ubuntu 16.13-0ubuntu0.24.04.1)
-- Dumped by pg_dump version 17.5

-- Started on 2026-03-26 14:52:04

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 2 (class 3079 OID 16400)
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- TOC entry 3710 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 218 (class 1259 OID 16450)
-- Name: User; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public."User" (
    user_id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying(255) NOT NULL,
    role character varying(20) NOT NULL,
    account_status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    email_verified boolean DEFAULT true NOT NULL,
    posting_limit integer DEFAULT 10 NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone,
    university_id uuid NOT NULL,
    keycloak_sub uuid
);


ALTER TABLE public."User" OWNER TO dealna_user;

--
-- TOC entry 219 (class 1259 OID 16465)
-- Name: admin; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.admin (
    user_id uuid NOT NULL,
    admin_name character varying(255) NOT NULL
);


ALTER TABLE public.admin OWNER TO dealna_user;

--
-- TOC entry 220 (class 1259 OID 16470)
-- Name: attachment; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.attachment (
    attachment_id uuid DEFAULT gen_random_uuid() NOT NULL,
    item_id uuid NOT NULL,
    file_path character varying(500) NOT NULL,
    uploaded_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.attachment OWNER TO dealna_user;

--
-- TOC entry 221 (class 1259 OID 16479)
-- Name: category; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.category (
    category_id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.category OWNER TO dealna_user;

--
-- TOC entry 222 (class 1259 OID 16490)
-- Name: chat; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.chat (
    chat_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user1_id uuid NOT NULL,
    user2_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone
);


ALTER TABLE public.chat OWNER TO dealna_user;

--
-- TOC entry 223 (class 1259 OID 16499)
-- Name: follow; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.follow (
    follower_profile_id uuid NOT NULL,
    following_profile_id uuid NOT NULL,
    followed_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.follow OWNER TO dealna_user;

--
-- TOC entry 217 (class 1259 OID 16391)
-- Name: health_test; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.health_test (
    id integer NOT NULL,
    note text,
    created_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.health_test OWNER TO postgres;

--
-- TOC entry 216 (class 1259 OID 16390)
-- Name: health_test_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.health_test_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.health_test_id_seq OWNER TO postgres;

--
-- TOC entry 3712 (class 0 OID 0)
-- Dependencies: 216
-- Name: health_test_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.health_test_id_seq OWNED BY public.health_test.id;


--
-- TOC entry 224 (class 1259 OID 16505)
-- Name: item; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.item (
    item_id uuid DEFAULT gen_random_uuid() NOT NULL,
    owner_id uuid NOT NULL,
    category_id uuid,
    title character varying(255) NOT NULL,
    description text,
    price numeric(10,2) NOT NULL,
    pickup_location text,
    item_status character varying(20) DEFAULT 'AVAILABLE'::character varying NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone
);


ALTER TABLE public.item OWNER TO dealna_user;

--
-- TOC entry 225 (class 1259 OID 16516)
-- Name: message; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.message (
    message_id uuid DEFAULT gen_random_uuid() NOT NULL,
    chat_id uuid NOT NULL,
    sender_id uuid NOT NULL,
    encrypted_content text NOT NULL,
    sent_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    read_at timestamp without time zone,
    edited_at timestamp without time zone,
    deleted_at timestamp without time zone
);


ALTER TABLE public.message OWNER TO dealna_user;

--
-- TOC entry 226 (class 1259 OID 16525)
-- Name: notification; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.notification (
    notification_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    type character varying(50) NOT NULL,
    payload jsonb,
    is_read boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.notification OWNER TO dealna_user;

--
-- TOC entry 227 (class 1259 OID 16535)
-- Name: profile; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.profile (
    profile_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    display_name character varying(50),
    bio text,
    profile_picture_url character varying(500),
    rating_count integer DEFAULT 0 NOT NULL,
    sold_items_count integer DEFAULT 0 NOT NULL,
    total_reviews_count integer DEFAULT 0 NOT NULL,
    follower_count integer DEFAULT 0 NOT NULL,
    following_count integer DEFAULT 0 NOT NULL,
    display_name_last_changed_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone
);


ALTER TABLE public.profile OWNER TO dealna_user;

--
-- TOC entry 228 (class 1259 OID 16553)
-- Name: provider; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.provider (
    user_id uuid NOT NULL,
    business_name character varying(255) NOT NULL,
    phone_number character varying(20),
    business_type character varying(100),
    address text,
    verified_at timestamp without time zone
);


ALTER TABLE public.provider OWNER TO dealna_user;

--
-- TOC entry 229 (class 1259 OID 16560)
-- Name: providerapplicant; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.providerapplicant (
    applicant_id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying(255) NOT NULL,
    email_verified boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    password_hash text,
    role text DEFAULT 'APPLICANT'::text
);


ALTER TABLE public.providerapplicant OWNER TO dealna_user;

--
-- TOC entry 230 (class 1259 OID 16573)
-- Name: providerapplication; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.providerapplication (
    application_id uuid DEFAULT gen_random_uuid() NOT NULL,
    applicant_id uuid NOT NULL,
    university_id uuid NOT NULL,
    business_name character varying(255) NOT NULL,
    phone_number character varying(20),
    business_type character varying(100),
    address text,
    status character varying(30) DEFAULT 'EMAIL_VERIFIED'::character varying NOT NULL,
    submitted_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    reviewed_at timestamp without time zone,
    admin_comment text,
    reviewed_by_admin_id uuid
);


ALTER TABLE public.providerapplication OWNER TO dealna_user;

--
-- TOC entry 231 (class 1259 OID 16583)
-- Name: providerapplicationdocument; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.providerapplicationdocument (
    document_id uuid DEFAULT gen_random_uuid() NOT NULL,
    application_id uuid NOT NULL,
    file_path character varying(500) NOT NULL,
    uploaded_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.providerapplicationdocument OWNER TO dealna_user;

--
-- TOC entry 232 (class 1259 OID 16592)
-- Name: providerreview; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.providerreview (
    review_id uuid DEFAULT gen_random_uuid() NOT NULL,
    application_id uuid NOT NULL,
    reviewer_admin_id uuid NOT NULL,
    decision character varying(20) NOT NULL,
    comment text,
    reviewed_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.providerreview OWNER TO dealna_user;

--
-- TOC entry 233 (class 1259 OID 16601)
-- Name: queue; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.queue (
    queue_id uuid DEFAULT gen_random_uuid() NOT NULL,
    item_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.queue OWNER TO dealna_user;

--
-- TOC entry 234 (class 1259 OID 16610)
-- Name: queueentry; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.queueentry (
    queue_entry_id uuid DEFAULT gen_random_uuid() NOT NULL,
    queue_id uuid NOT NULL,
    buyer_id uuid NOT NULL,
    "position" integer NOT NULL,
    joined_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expired_at timestamp without time zone,
    notified_at timestamp without time zone
);


ALTER TABLE public.queueentry OWNER TO dealna_user;

--
-- TOC entry 235 (class 1259 OID 16619)
-- Name: rating; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.rating (
    rating_id uuid DEFAULT gen_random_uuid() NOT NULL,
    transaction_id uuid NOT NULL,
    rater_id uuid NOT NULL,
    rated_user_id uuid NOT NULL,
    stars integer NOT NULL,
    comment text,
    is_frozen boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.rating OWNER TO dealna_user;

--
-- TOC entry 236 (class 1259 OID 16631)
-- Name: report; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.report (
    report_id uuid DEFAULT gen_random_uuid() NOT NULL,
    reporter_id uuid NOT NULL,
    reported_user_id uuid,
    reported_item_id uuid,
    reason character varying(100) NOT NULL,
    description text,
    status character varying(20) DEFAULT 'PENDING'::character varying NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    resolved_at timestamp without time zone
);


ALTER TABLE public.report OWNER TO dealna_user;

--
-- TOC entry 240 (class 1259 OID 16865)
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


ALTER TABLE public.schema_migrations OWNER TO dealna_user;

--
-- TOC entry 237 (class 1259 OID 16641)
-- Name: student; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.student (
    user_id uuid NOT NULL,
    student_id character varying(50) NOT NULL,
    major character varying(100),
    academic_year integer
);


ALTER TABLE public.student OWNER TO dealna_user;

--
-- TOC entry 241 (class 1259 OID 16974)
-- Name: student_pre_registration; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.student_pre_registration (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying(255) NOT NULL,
    token uuid NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    used_at timestamp without time zone,
    resend_count integer DEFAULT 0 NOT NULL,
    resend_window_start timestamp without time zone,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    verified_at timestamp without time zone
);


ALTER TABLE public.student_pre_registration OWNER TO dealna_user;

--
-- TOC entry 238 (class 1259 OID 16647)
-- Name: transaction; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.transaction (
    transaction_id uuid DEFAULT gen_random_uuid() NOT NULL,
    item_id uuid NOT NULL,
    buyer_id uuid NOT NULL,
    seller_id uuid NOT NULL,
    transaction_status character varying(20) DEFAULT 'PENDING'::character varying NOT NULL,
    is_flagged boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    completed_at timestamp without time zone,
    deleted_at timestamp without time zone
);


ALTER TABLE public.transaction OWNER TO dealna_user;

--
-- TOC entry 239 (class 1259 OID 16656)
-- Name: university; Type: TABLE; Schema: public; Owner: dealna_user
--

CREATE TABLE public.university (
    university_id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    domain character varying(255) NOT NULL,
    status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.university OWNER TO dealna_user;

--
-- TOC entry 3384 (class 2604 OID 16394)
-- Name: health_test id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.health_test ALTER COLUMN id SET DEFAULT nextval('public.health_test_id_seq'::regclass);


--
-- TOC entry 3448 (class 2606 OID 16464)
-- Name: User User_email_key; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public."User"
    ADD CONSTRAINT "User_email_key" UNIQUE (email);


--
-- TOC entry 3450 (class 2606 OID 17001)
-- Name: User User_keycloak_sub_key; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public."User"
    ADD CONSTRAINT "User_keycloak_sub_key" UNIQUE (keycloak_sub);


--
-- TOC entry 3452 (class 2606 OID 16462)
-- Name: User User_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public."User"
    ADD CONSTRAINT "User_pkey" PRIMARY KEY (user_id);


--
-- TOC entry 3454 (class 2606 OID 16469)
-- Name: admin admin_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.admin
    ADD CONSTRAINT admin_pkey PRIMARY KEY (user_id);


--
-- TOC entry 3456 (class 2606 OID 16478)
-- Name: attachment attachment_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.attachment
    ADD CONSTRAINT attachment_pkey PRIMARY KEY (attachment_id);


--
-- TOC entry 3458 (class 2606 OID 16489)
-- Name: category category_name_key; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.category
    ADD CONSTRAINT category_name_key UNIQUE (name);


--
-- TOC entry 3460 (class 2606 OID 16487)
-- Name: category category_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.category
    ADD CONSTRAINT category_pkey PRIMARY KEY (category_id);


--
-- TOC entry 3462 (class 2606 OID 16496)
-- Name: chat chat_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.chat
    ADD CONSTRAINT chat_pkey PRIMARY KEY (chat_id);


--
-- TOC entry 3466 (class 2606 OID 16504)
-- Name: follow follow_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.follow
    ADD CONSTRAINT follow_pkey PRIMARY KEY (follower_profile_id, following_profile_id);


--
-- TOC entry 3446 (class 2606 OID 16399)
-- Name: health_test health_test_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.health_test
    ADD CONSTRAINT health_test_pkey PRIMARY KEY (id);


--
-- TOC entry 3468 (class 2606 OID 16515)
-- Name: item item_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.item
    ADD CONSTRAINT item_pkey PRIMARY KEY (item_id);


--
-- TOC entry 3470 (class 2606 OID 16524)
-- Name: message message_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.message
    ADD CONSTRAINT message_pkey PRIMARY KEY (message_id);


--
-- TOC entry 3472 (class 2606 OID 16534)
-- Name: notification notification_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT notification_pkey PRIMARY KEY (notification_id);


--
-- TOC entry 3474 (class 2606 OID 16550)
-- Name: profile profile_display_name_key; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.profile
    ADD CONSTRAINT profile_display_name_key UNIQUE (display_name);


--
-- TOC entry 3476 (class 2606 OID 16548)
-- Name: profile profile_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.profile
    ADD CONSTRAINT profile_pkey PRIMARY KEY (profile_id);


--
-- TOC entry 3478 (class 2606 OID 16552)
-- Name: profile profile_user_id_key; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.profile
    ADD CONSTRAINT profile_user_id_key UNIQUE (user_id);


--
-- TOC entry 3480 (class 2606 OID 16559)
-- Name: provider provider_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.provider
    ADD CONSTRAINT provider_pkey PRIMARY KEY (user_id);


--
-- TOC entry 3482 (class 2606 OID 16572)
-- Name: providerapplicant providerapplicant_email_key; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.providerapplicant
    ADD CONSTRAINT providerapplicant_email_key UNIQUE (email);


--
-- TOC entry 3484 (class 2606 OID 16570)
-- Name: providerapplicant providerapplicant_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.providerapplicant
    ADD CONSTRAINT providerapplicant_pkey PRIMARY KEY (applicant_id);


--
-- TOC entry 3486 (class 2606 OID 16582)
-- Name: providerapplication providerapplication_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.providerapplication
    ADD CONSTRAINT providerapplication_pkey PRIMARY KEY (application_id);


--
-- TOC entry 3488 (class 2606 OID 16591)
-- Name: providerapplicationdocument providerapplicationdocument_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.providerapplicationdocument
    ADD CONSTRAINT providerapplicationdocument_pkey PRIMARY KEY (document_id);


--
-- TOC entry 3490 (class 2606 OID 16600)
-- Name: providerreview providerreview_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.providerreview
    ADD CONSTRAINT providerreview_pkey PRIMARY KEY (review_id);


--
-- TOC entry 3492 (class 2606 OID 16609)
-- Name: queue queue_item_id_key; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.queue
    ADD CONSTRAINT queue_item_id_key UNIQUE (item_id);


--
-- TOC entry 3494 (class 2606 OID 16607)
-- Name: queue queue_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.queue
    ADD CONSTRAINT queue_pkey PRIMARY KEY (queue_id);


--
-- TOC entry 3496 (class 2606 OID 16616)
-- Name: queueentry queueentry_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.queueentry
    ADD CONSTRAINT queueentry_pkey PRIMARY KEY (queue_entry_id);


--
-- TOC entry 3500 (class 2606 OID 16628)
-- Name: rating rating_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.rating
    ADD CONSTRAINT rating_pkey PRIMARY KEY (rating_id);


--
-- TOC entry 3502 (class 2606 OID 16630)
-- Name: rating rating_transaction_id_key; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.rating
    ADD CONSTRAINT rating_transaction_id_key UNIQUE (transaction_id);


--
-- TOC entry 3504 (class 2606 OID 16640)
-- Name: report report_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.report
    ADD CONSTRAINT report_pkey PRIMARY KEY (report_id);


--
-- TOC entry 3515 (class 2606 OID 16869)
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- TOC entry 3506 (class 2606 OID 16646)
-- Name: student student_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.student
    ADD CONSTRAINT student_pkey PRIMARY KEY (user_id);


--
-- TOC entry 3517 (class 2606 OID 16983)
-- Name: student_pre_registration student_pre_registration_email_key; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.student_pre_registration
    ADD CONSTRAINT student_pre_registration_email_key UNIQUE (email);


--
-- TOC entry 3519 (class 2606 OID 16981)
-- Name: student_pre_registration student_pre_registration_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.student_pre_registration
    ADD CONSTRAINT student_pre_registration_pkey PRIMARY KEY (id);


--
-- TOC entry 3521 (class 2606 OID 16985)
-- Name: student_pre_registration student_pre_registration_token_key; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.student_pre_registration
    ADD CONSTRAINT student_pre_registration_token_key UNIQUE (token);


--
-- TOC entry 3509 (class 2606 OID 16655)
-- Name: transaction transaction_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.transaction
    ADD CONSTRAINT transaction_pkey PRIMARY KEY (transaction_id);


--
-- TOC entry 3511 (class 2606 OID 16667)
-- Name: university university_domain_key; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.university
    ADD CONSTRAINT university_domain_key UNIQUE (domain);


--
-- TOC entry 3513 (class 2606 OID 16665)
-- Name: university university_pkey; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.university
    ADD CONSTRAINT university_pkey PRIMARY KEY (university_id);


--
-- TOC entry 3464 (class 2606 OID 16498)
-- Name: chat uq_chat_users; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.chat
    ADD CONSTRAINT uq_chat_users UNIQUE (user1_id, user2_id);


--
-- TOC entry 3498 (class 2606 OID 16618)
-- Name: queueentry uq_queueentry; Type: CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.queueentry
    ADD CONSTRAINT uq_queueentry UNIQUE (queue_id, buyer_id);


--
-- TOC entry 3507 (class 1259 OID 16828)
-- Name: idx_transaction_item_active; Type: INDEX; Schema: public; Owner: dealna_user
--

CREATE INDEX idx_transaction_item_active ON public.transaction USING btree (item_id);


--
-- TOC entry 3523 (class 2606 OID 16673)
-- Name: admin fk_admin_user; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.admin
    ADD CONSTRAINT fk_admin_user FOREIGN KEY (user_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3524 (class 2606 OID 16678)
-- Name: attachment fk_attachment_item; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.attachment
    ADD CONSTRAINT fk_attachment_item FOREIGN KEY (item_id) REFERENCES public.item(item_id) ON DELETE CASCADE;


--
-- TOC entry 3525 (class 2606 OID 16683)
-- Name: chat fk_chat_user1; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.chat
    ADD CONSTRAINT fk_chat_user1 FOREIGN KEY (user1_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3526 (class 2606 OID 16688)
-- Name: chat fk_chat_user2; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.chat
    ADD CONSTRAINT fk_chat_user2 FOREIGN KEY (user2_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3527 (class 2606 OID 16693)
-- Name: follow fk_follow_follower; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.follow
    ADD CONSTRAINT fk_follow_follower FOREIGN KEY (follower_profile_id) REFERENCES public.profile(profile_id) ON DELETE CASCADE;


--
-- TOC entry 3528 (class 2606 OID 16698)
-- Name: follow fk_follow_following; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.follow
    ADD CONSTRAINT fk_follow_following FOREIGN KEY (following_profile_id) REFERENCES public.profile(profile_id) ON DELETE CASCADE;


--
-- TOC entry 3529 (class 2606 OID 16703)
-- Name: item fk_item_category; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.item
    ADD CONSTRAINT fk_item_category FOREIGN KEY (category_id) REFERENCES public.category(category_id) ON DELETE SET NULL;


--
-- TOC entry 3530 (class 2606 OID 16708)
-- Name: item fk_item_owner; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.item
    ADD CONSTRAINT fk_item_owner FOREIGN KEY (owner_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3531 (class 2606 OID 16713)
-- Name: message fk_message_chat; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.message
    ADD CONSTRAINT fk_message_chat FOREIGN KEY (chat_id) REFERENCES public.chat(chat_id) ON DELETE CASCADE;


--
-- TOC entry 3532 (class 2606 OID 16718)
-- Name: message fk_message_sender; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.message
    ADD CONSTRAINT fk_message_sender FOREIGN KEY (sender_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3533 (class 2606 OID 16723)
-- Name: notification fk_notification_user; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT fk_notification_user FOREIGN KEY (user_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3534 (class 2606 OID 16728)
-- Name: profile fk_profile_user; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.profile
    ADD CONSTRAINT fk_profile_user FOREIGN KEY (user_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3536 (class 2606 OID 16738)
-- Name: providerapplication fk_provapp_admin; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.providerapplication
    ADD CONSTRAINT fk_provapp_admin FOREIGN KEY (reviewed_by_admin_id) REFERENCES public.admin(user_id);


--
-- TOC entry 3537 (class 2606 OID 16743)
-- Name: providerapplication fk_provapp_applicant; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.providerapplication
    ADD CONSTRAINT fk_provapp_applicant FOREIGN KEY (applicant_id) REFERENCES public.providerapplicant(applicant_id) ON DELETE CASCADE;


--
-- TOC entry 3538 (class 2606 OID 16748)
-- Name: providerapplication fk_provapp_university; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.providerapplication
    ADD CONSTRAINT fk_provapp_university FOREIGN KEY (university_id) REFERENCES public.university(university_id);


--
-- TOC entry 3542 (class 2606 OID 16753)
-- Name: providerapplicationdocument fk_provappdoc_application; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.providerapplicationdocument
    ADD CONSTRAINT fk_provappdoc_application FOREIGN KEY (application_id) REFERENCES public.providerapplication(application_id) ON DELETE CASCADE;


--
-- TOC entry 3535 (class 2606 OID 16733)
-- Name: provider fk_provider_user; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.provider
    ADD CONSTRAINT fk_provider_user FOREIGN KEY (user_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3539 (class 2606 OID 16845)
-- Name: providerapplication fk_providerapplication_admin; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.providerapplication
    ADD CONSTRAINT fk_providerapplication_admin FOREIGN KEY (reviewed_by_admin_id) REFERENCES public.admin(user_id);


--
-- TOC entry 3540 (class 2606 OID 16835)
-- Name: providerapplication fk_providerapplication_applicant; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.providerapplication
    ADD CONSTRAINT fk_providerapplication_applicant FOREIGN KEY (applicant_id) REFERENCES public.providerapplicant(applicant_id) ON DELETE CASCADE;


--
-- TOC entry 3541 (class 2606 OID 16840)
-- Name: providerapplication fk_providerapplication_university; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.providerapplication
    ADD CONSTRAINT fk_providerapplication_university FOREIGN KEY (university_id) REFERENCES public.university(university_id);


--
-- TOC entry 3543 (class 2606 OID 16850)
-- Name: providerapplicationdocument fk_providerapplicationdocument_application; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.providerapplicationdocument
    ADD CONSTRAINT fk_providerapplicationdocument_application FOREIGN KEY (application_id) REFERENCES public.providerapplication(application_id) ON DELETE CASCADE;


--
-- TOC entry 3544 (class 2606 OID 16855)
-- Name: providerreview fk_providerreview_admin; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.providerreview
    ADD CONSTRAINT fk_providerreview_admin FOREIGN KEY (reviewer_admin_id) REFERENCES public.admin(user_id);


--
-- TOC entry 3545 (class 2606 OID 16860)
-- Name: providerreview fk_providerreview_application; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.providerreview
    ADD CONSTRAINT fk_providerreview_application FOREIGN KEY (application_id) REFERENCES public.providerapplication(application_id) ON DELETE CASCADE;


--
-- TOC entry 3546 (class 2606 OID 16758)
-- Name: providerreview fk_provreview_admin; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.providerreview
    ADD CONSTRAINT fk_provreview_admin FOREIGN KEY (reviewer_admin_id) REFERENCES public.admin(user_id);


--
-- TOC entry 3547 (class 2606 OID 16763)
-- Name: providerreview fk_provreview_application; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.providerreview
    ADD CONSTRAINT fk_provreview_application FOREIGN KEY (application_id) REFERENCES public.providerapplication(application_id) ON DELETE CASCADE;


--
-- TOC entry 3548 (class 2606 OID 16768)
-- Name: queue fk_queue_item; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.queue
    ADD CONSTRAINT fk_queue_item FOREIGN KEY (item_id) REFERENCES public.item(item_id) ON DELETE CASCADE;


--
-- TOC entry 3549 (class 2606 OID 16773)
-- Name: queueentry fk_queueentry_buyer; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.queueentry
    ADD CONSTRAINT fk_queueentry_buyer FOREIGN KEY (buyer_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3550 (class 2606 OID 16778)
-- Name: queueentry fk_queueentry_queue; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.queueentry
    ADD CONSTRAINT fk_queueentry_queue FOREIGN KEY (queue_id) REFERENCES public.queue(queue_id) ON DELETE CASCADE;


--
-- TOC entry 3551 (class 2606 OID 16783)
-- Name: rating fk_rating_rated_user; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.rating
    ADD CONSTRAINT fk_rating_rated_user FOREIGN KEY (rated_user_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3552 (class 2606 OID 16788)
-- Name: rating fk_rating_rater; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.rating
    ADD CONSTRAINT fk_rating_rater FOREIGN KEY (rater_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3553 (class 2606 OID 16793)
-- Name: rating fk_rating_transaction; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.rating
    ADD CONSTRAINT fk_rating_transaction FOREIGN KEY (transaction_id) REFERENCES public.transaction(transaction_id) ON DELETE CASCADE;


--
-- TOC entry 3554 (class 2606 OID 16798)
-- Name: report fk_report_reported_item; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.report
    ADD CONSTRAINT fk_report_reported_item FOREIGN KEY (reported_item_id) REFERENCES public.item(item_id) ON DELETE CASCADE;


--
-- TOC entry 3555 (class 2606 OID 16803)
-- Name: report fk_report_reported_user; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.report
    ADD CONSTRAINT fk_report_reported_user FOREIGN KEY (reported_user_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3556 (class 2606 OID 16808)
-- Name: report fk_report_reporter; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.report
    ADD CONSTRAINT fk_report_reporter FOREIGN KEY (reporter_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3557 (class 2606 OID 16813)
-- Name: student fk_student_user; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.student
    ADD CONSTRAINT fk_student_user FOREIGN KEY (user_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3558 (class 2606 OID 16818)
-- Name: transaction fk_transaction_buyer; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.transaction
    ADD CONSTRAINT fk_transaction_buyer FOREIGN KEY (buyer_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3559 (class 2606 OID 16823)
-- Name: transaction fk_transaction_item; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.transaction
    ADD CONSTRAINT fk_transaction_item FOREIGN KEY (item_id) REFERENCES public.item(item_id) ON DELETE CASCADE;


--
-- TOC entry 3560 (class 2606 OID 16829)
-- Name: transaction fk_transaction_seller; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public.transaction
    ADD CONSTRAINT fk_transaction_seller FOREIGN KEY (seller_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3522 (class 2606 OID 16668)
-- Name: User fk_user_university; Type: FK CONSTRAINT; Schema: public; Owner: dealna_user
--

ALTER TABLE ONLY public."User"
    ADD CONSTRAINT fk_user_university FOREIGN KEY (university_id) REFERENCES public.university(university_id);


--
-- TOC entry 3709 (class 0 OID 0)
-- Dependencies: 6
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: pg_database_owner
--

GRANT ALL ON SCHEMA public TO dealna_user;


--
-- TOC entry 3711 (class 0 OID 0)
-- Dependencies: 217
-- Name: TABLE health_test; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,INSERT,REFERENCES,DELETE,TRIGGER,TRUNCATE,UPDATE ON TABLE public.health_test TO dealna_user;


--
-- TOC entry 3713 (class 0 OID 0)
-- Dependencies: 216
-- Name: SEQUENCE health_test_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.health_test_id_seq TO dealna_user;


--
-- TOC entry 2172 (class 826 OID 16440)
-- Name: DEFAULT PRIVILEGES FOR SEQUENCES; Type: DEFAULT ACL; Schema: public; Owner: postgres
--

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT ALL ON SEQUENCES TO dealna_user;


--
-- TOC entry 2171 (class 826 OID 16439)
-- Name: DEFAULT PRIVILEGES FOR TABLES; Type: DEFAULT ACL; Schema: public; Owner: postgres
--

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT SELECT,INSERT,REFERENCES,DELETE,TRIGGER,TRUNCATE,UPDATE ON TABLES TO dealna_user;


-- Completed on 2026-03-26 14:52:17

--
-- PostgreSQL database dump complete
--

