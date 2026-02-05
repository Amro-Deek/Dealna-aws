--
-- PostgreSQL database dump
--

-- Dumped from database version 16.11 (Ubuntu 16.11-0ubuntu0.24.04.1)
-- Dumped by pg_dump version 17.5

-- Started on 2026-02-04 03:31:11

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
-- TOC entry 3714 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 218 (class 1259 OID 16450)
-- Name: User; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public."User" (
    user_id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying(255) NOT NULL,
    password_hash character varying(255),
    role character varying(20) NOT NULL,
    account_status character varying(20) DEFAULT 'PENDING'::character varying NOT NULL,
    email_verified boolean DEFAULT false NOT NULL,
    posting_limit integer DEFAULT 10 NOT NULL,
    failed_login_attempts integer DEFAULT 0 NOT NULL,
    last_login_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone,
    university_id uuid NOT NULL
);


--
-- TOC entry 219 (class 1259 OID 16465)
-- Name: admin; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.admin (
    user_id uuid NOT NULL,
    admin_name character varying(255) NOT NULL
);


--
-- TOC entry 220 (class 1259 OID 16470)
-- Name: attachment; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.attachment (
    attachment_id uuid DEFAULT gen_random_uuid() NOT NULL,
    item_id uuid NOT NULL,
    file_path character varying(500) NOT NULL,
    uploaded_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- TOC entry 221 (class 1259 OID 16479)
-- Name: category; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.category (
    category_id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- TOC entry 222 (class 1259 OID 16490)
-- Name: chat; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chat (
    chat_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user1_id uuid NOT NULL,
    user2_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone
);


--
-- TOC entry 223 (class 1259 OID 16499)
-- Name: follow; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.follow (
    follower_profile_id uuid NOT NULL,
    following_profile_id uuid NOT NULL,
    followed_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- TOC entry 217 (class 1259 OID 16391)
-- Name: health_test; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.health_test (
    id integer NOT NULL,
    note text,
    created_at timestamp without time zone DEFAULT now()
);


--
-- TOC entry 216 (class 1259 OID 16390)
-- Name: health_test_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.health_test_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 3715 (class 0 OID 0)
-- Dependencies: 216
-- Name: health_test_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.health_test_id_seq OWNED BY public.health_test.id;


--
-- TOC entry 224 (class 1259 OID 16505)
-- Name: item; Type: TABLE; Schema: public; Owner: -
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


--
-- TOC entry 225 (class 1259 OID 16516)
-- Name: message; Type: TABLE; Schema: public; Owner: -
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


--
-- TOC entry 226 (class 1259 OID 16525)
-- Name: notification; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notification (
    notification_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    type character varying(50) NOT NULL,
    payload jsonb,
    is_read boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- TOC entry 227 (class 1259 OID 16535)
-- Name: profile; Type: TABLE; Schema: public; Owner: -
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


--
-- TOC entry 228 (class 1259 OID 16553)
-- Name: provider; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.provider (
    user_id uuid NOT NULL,
    business_name character varying(255) NOT NULL,
    phone_number character varying(20),
    business_type character varying(100),
    address text,
    verified_at timestamp without time zone
);


--
-- TOC entry 229 (class 1259 OID 16560)
-- Name: providerapplicant; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.providerapplicant (
    applicant_id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying(255) NOT NULL,
    email_verified boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    password_hash text,
    role text DEFAULT 'APPLICANT'::text
);


--
-- TOC entry 230 (class 1259 OID 16573)
-- Name: providerapplication; Type: TABLE; Schema: public; Owner: -
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


--
-- TOC entry 231 (class 1259 OID 16583)
-- Name: providerapplicationdocument; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.providerapplicationdocument (
    document_id uuid DEFAULT gen_random_uuid() NOT NULL,
    application_id uuid NOT NULL,
    file_path character varying(500) NOT NULL,
    uploaded_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- TOC entry 232 (class 1259 OID 16592)
-- Name: providerreview; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.providerreview (
    review_id uuid DEFAULT gen_random_uuid() NOT NULL,
    application_id uuid NOT NULL,
    reviewer_admin_id uuid NOT NULL,
    decision character varying(20) NOT NULL,
    comment text,
    reviewed_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- TOC entry 233 (class 1259 OID 16601)
-- Name: queue; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.queue (
    queue_id uuid DEFAULT gen_random_uuid() NOT NULL,
    item_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- TOC entry 234 (class 1259 OID 16610)
-- Name: queueentry; Type: TABLE; Schema: public; Owner: -
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


--
-- TOC entry 235 (class 1259 OID 16619)
-- Name: rating; Type: TABLE; Schema: public; Owner: -
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


--
-- TOC entry 236 (class 1259 OID 16631)
-- Name: report; Type: TABLE; Schema: public; Owner: -
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


--
-- TOC entry 237 (class 1259 OID 16641)
-- Name: student; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.student (
    user_id uuid NOT NULL,
    student_id character varying(50) NOT NULL,
    major character varying(100),
    academic_year integer,
    verification_status boolean DEFAULT false NOT NULL
);


--
-- TOC entry 238 (class 1259 OID 16647)
-- Name: transaction; Type: TABLE; Schema: public; Owner: -
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


--
-- TOC entry 239 (class 1259 OID 16656)
-- Name: university; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.university (
    university_id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    domain character varying(255) NOT NULL,
    status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- TOC entry 3376 (class 2604 OID 16394)
-- Name: health_test id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.health_test ALTER COLUMN id SET DEFAULT nextval('public.health_test_id_seq'::regclass);


--
-- TOC entry 3439 (class 2606 OID 16464)
-- Name: User User_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."User"
    ADD CONSTRAINT "User_email_key" UNIQUE (email);


--
-- TOC entry 3441 (class 2606 OID 16462)
-- Name: User User_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."User"
    ADD CONSTRAINT "User_pkey" PRIMARY KEY (user_id);


--
-- TOC entry 3443 (class 2606 OID 16469)
-- Name: admin admin_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin
    ADD CONSTRAINT admin_pkey PRIMARY KEY (user_id);


--
-- TOC entry 3445 (class 2606 OID 16478)
-- Name: attachment attachment_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attachment
    ADD CONSTRAINT attachment_pkey PRIMARY KEY (attachment_id);


--
-- TOC entry 3447 (class 2606 OID 16489)
-- Name: category category_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.category
    ADD CONSTRAINT category_name_key UNIQUE (name);


--
-- TOC entry 3449 (class 2606 OID 16487)
-- Name: category category_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.category
    ADD CONSTRAINT category_pkey PRIMARY KEY (category_id);


--
-- TOC entry 3451 (class 2606 OID 16496)
-- Name: chat chat_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat
    ADD CONSTRAINT chat_pkey PRIMARY KEY (chat_id);


--
-- TOC entry 3455 (class 2606 OID 16504)
-- Name: follow follow_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.follow
    ADD CONSTRAINT follow_pkey PRIMARY KEY (follower_profile_id, following_profile_id);


--
-- TOC entry 3437 (class 2606 OID 16399)
-- Name: health_test health_test_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.health_test
    ADD CONSTRAINT health_test_pkey PRIMARY KEY (id);


--
-- TOC entry 3457 (class 2606 OID 16515)
-- Name: item item_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.item
    ADD CONSTRAINT item_pkey PRIMARY KEY (item_id);


--
-- TOC entry 3459 (class 2606 OID 16524)
-- Name: message message_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message
    ADD CONSTRAINT message_pkey PRIMARY KEY (message_id);


--
-- TOC entry 3461 (class 2606 OID 16534)
-- Name: notification notification_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT notification_pkey PRIMARY KEY (notification_id);


--
-- TOC entry 3463 (class 2606 OID 16550)
-- Name: profile profile_display_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile
    ADD CONSTRAINT profile_display_name_key UNIQUE (display_name);


--
-- TOC entry 3465 (class 2606 OID 16548)
-- Name: profile profile_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile
    ADD CONSTRAINT profile_pkey PRIMARY KEY (profile_id);


--
-- TOC entry 3467 (class 2606 OID 16552)
-- Name: profile profile_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile
    ADD CONSTRAINT profile_user_id_key UNIQUE (user_id);


--
-- TOC entry 3469 (class 2606 OID 16559)
-- Name: provider provider_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.provider
    ADD CONSTRAINT provider_pkey PRIMARY KEY (user_id);


--
-- TOC entry 3471 (class 2606 OID 16572)
-- Name: providerapplicant providerapplicant_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providerapplicant
    ADD CONSTRAINT providerapplicant_email_key UNIQUE (email);


--
-- TOC entry 3473 (class 2606 OID 16570)
-- Name: providerapplicant providerapplicant_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providerapplicant
    ADD CONSTRAINT providerapplicant_pkey PRIMARY KEY (applicant_id);


--
-- TOC entry 3475 (class 2606 OID 16582)
-- Name: providerapplication providerapplication_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providerapplication
    ADD CONSTRAINT providerapplication_pkey PRIMARY KEY (application_id);


--
-- TOC entry 3477 (class 2606 OID 16591)
-- Name: providerapplicationdocument providerapplicationdocument_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providerapplicationdocument
    ADD CONSTRAINT providerapplicationdocument_pkey PRIMARY KEY (document_id);


--
-- TOC entry 3479 (class 2606 OID 16600)
-- Name: providerreview providerreview_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providerreview
    ADD CONSTRAINT providerreview_pkey PRIMARY KEY (review_id);


--
-- TOC entry 3481 (class 2606 OID 16609)
-- Name: queue queue_item_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.queue
    ADD CONSTRAINT queue_item_id_key UNIQUE (item_id);


--
-- TOC entry 3483 (class 2606 OID 16607)
-- Name: queue queue_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.queue
    ADD CONSTRAINT queue_pkey PRIMARY KEY (queue_id);


--
-- TOC entry 3485 (class 2606 OID 16616)
-- Name: queueentry queueentry_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.queueentry
    ADD CONSTRAINT queueentry_pkey PRIMARY KEY (queue_entry_id);


--
-- TOC entry 3489 (class 2606 OID 16628)
-- Name: rating rating_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rating
    ADD CONSTRAINT rating_pkey PRIMARY KEY (rating_id);


--
-- TOC entry 3491 (class 2606 OID 16630)
-- Name: rating rating_transaction_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rating
    ADD CONSTRAINT rating_transaction_id_key UNIQUE (transaction_id);


--
-- TOC entry 3493 (class 2606 OID 16640)
-- Name: report report_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report
    ADD CONSTRAINT report_pkey PRIMARY KEY (report_id);


--
-- TOC entry 3495 (class 2606 OID 16646)
-- Name: student student_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student
    ADD CONSTRAINT student_pkey PRIMARY KEY (user_id);


--
-- TOC entry 3498 (class 2606 OID 16655)
-- Name: transaction transaction_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transaction
    ADD CONSTRAINT transaction_pkey PRIMARY KEY (transaction_id);


--
-- TOC entry 3500 (class 2606 OID 16667)
-- Name: university university_domain_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.university
    ADD CONSTRAINT university_domain_key UNIQUE (domain);


--
-- TOC entry 3502 (class 2606 OID 16665)
-- Name: university university_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.university
    ADD CONSTRAINT university_pkey PRIMARY KEY (university_id);


--
-- TOC entry 3453 (class 2606 OID 16498)
-- Name: chat uq_chat_users; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat
    ADD CONSTRAINT uq_chat_users UNIQUE (user1_id, user2_id);


--
-- TOC entry 3487 (class 2606 OID 16618)
-- Name: queueentry uq_queueentry; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.queueentry
    ADD CONSTRAINT uq_queueentry UNIQUE (queue_id, buyer_id);


--
-- TOC entry 3496 (class 1259 OID 16828)
-- Name: idx_transaction_item_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transaction_item_active ON public.transaction USING btree (item_id);


--
-- TOC entry 3504 (class 2606 OID 16673)
-- Name: admin fk_admin_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin
    ADD CONSTRAINT fk_admin_user FOREIGN KEY (user_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3505 (class 2606 OID 16678)
-- Name: attachment fk_attachment_item; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attachment
    ADD CONSTRAINT fk_attachment_item FOREIGN KEY (item_id) REFERENCES public.item(item_id) ON DELETE CASCADE;


--
-- TOC entry 3506 (class 2606 OID 16683)
-- Name: chat fk_chat_user1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat
    ADD CONSTRAINT fk_chat_user1 FOREIGN KEY (user1_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3507 (class 2606 OID 16688)
-- Name: chat fk_chat_user2; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat
    ADD CONSTRAINT fk_chat_user2 FOREIGN KEY (user2_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3508 (class 2606 OID 16693)
-- Name: follow fk_follow_follower; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.follow
    ADD CONSTRAINT fk_follow_follower FOREIGN KEY (follower_profile_id) REFERENCES public.profile(profile_id) ON DELETE CASCADE;


--
-- TOC entry 3509 (class 2606 OID 16698)
-- Name: follow fk_follow_following; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.follow
    ADD CONSTRAINT fk_follow_following FOREIGN KEY (following_profile_id) REFERENCES public.profile(profile_id) ON DELETE CASCADE;


--
-- TOC entry 3510 (class 2606 OID 16703)
-- Name: item fk_item_category; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.item
    ADD CONSTRAINT fk_item_category FOREIGN KEY (category_id) REFERENCES public.category(category_id) ON DELETE SET NULL;


--
-- TOC entry 3511 (class 2606 OID 16708)
-- Name: item fk_item_owner; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.item
    ADD CONSTRAINT fk_item_owner FOREIGN KEY (owner_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3512 (class 2606 OID 16713)
-- Name: message fk_message_chat; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message
    ADD CONSTRAINT fk_message_chat FOREIGN KEY (chat_id) REFERENCES public.chat(chat_id) ON DELETE CASCADE;


--
-- TOC entry 3513 (class 2606 OID 16718)
-- Name: message fk_message_sender; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message
    ADD CONSTRAINT fk_message_sender FOREIGN KEY (sender_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3514 (class 2606 OID 16723)
-- Name: notification fk_notification_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification
    ADD CONSTRAINT fk_notification_user FOREIGN KEY (user_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3515 (class 2606 OID 16728)
-- Name: profile fk_profile_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profile
    ADD CONSTRAINT fk_profile_user FOREIGN KEY (user_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3517 (class 2606 OID 16738)
-- Name: providerapplication fk_provapp_admin; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providerapplication
    ADD CONSTRAINT fk_provapp_admin FOREIGN KEY (reviewed_by_admin_id) REFERENCES public.admin(user_id);


--
-- TOC entry 3518 (class 2606 OID 16743)
-- Name: providerapplication fk_provapp_applicant; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providerapplication
    ADD CONSTRAINT fk_provapp_applicant FOREIGN KEY (applicant_id) REFERENCES public.providerapplicant(applicant_id) ON DELETE CASCADE;


--
-- TOC entry 3519 (class 2606 OID 16748)
-- Name: providerapplication fk_provapp_university; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providerapplication
    ADD CONSTRAINT fk_provapp_university FOREIGN KEY (university_id) REFERENCES public.university(university_id);


--
-- TOC entry 3523 (class 2606 OID 16753)
-- Name: providerapplicationdocument fk_provappdoc_application; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providerapplicationdocument
    ADD CONSTRAINT fk_provappdoc_application FOREIGN KEY (application_id) REFERENCES public.providerapplication(application_id) ON DELETE CASCADE;


--
-- TOC entry 3516 (class 2606 OID 16733)
-- Name: provider fk_provider_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.provider
    ADD CONSTRAINT fk_provider_user FOREIGN KEY (user_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3520 (class 2606 OID 16845)
-- Name: providerapplication fk_providerapplication_admin; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providerapplication
    ADD CONSTRAINT fk_providerapplication_admin FOREIGN KEY (reviewed_by_admin_id) REFERENCES public.admin(user_id);


--
-- TOC entry 3521 (class 2606 OID 16835)
-- Name: providerapplication fk_providerapplication_applicant; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providerapplication
    ADD CONSTRAINT fk_providerapplication_applicant FOREIGN KEY (applicant_id) REFERENCES public.providerapplicant(applicant_id) ON DELETE CASCADE;


--
-- TOC entry 3522 (class 2606 OID 16840)
-- Name: providerapplication fk_providerapplication_university; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providerapplication
    ADD CONSTRAINT fk_providerapplication_university FOREIGN KEY (university_id) REFERENCES public.university(university_id);


--
-- TOC entry 3524 (class 2606 OID 16850)
-- Name: providerapplicationdocument fk_providerapplicationdocument_application; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providerapplicationdocument
    ADD CONSTRAINT fk_providerapplicationdocument_application FOREIGN KEY (application_id) REFERENCES public.providerapplication(application_id) ON DELETE CASCADE;


--
-- TOC entry 3525 (class 2606 OID 16855)
-- Name: providerreview fk_providerreview_admin; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providerreview
    ADD CONSTRAINT fk_providerreview_admin FOREIGN KEY (reviewer_admin_id) REFERENCES public.admin(user_id);


--
-- TOC entry 3526 (class 2606 OID 16860)
-- Name: providerreview fk_providerreview_application; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providerreview
    ADD CONSTRAINT fk_providerreview_application FOREIGN KEY (application_id) REFERENCES public.providerapplication(application_id) ON DELETE CASCADE;


--
-- TOC entry 3527 (class 2606 OID 16758)
-- Name: providerreview fk_provreview_admin; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providerreview
    ADD CONSTRAINT fk_provreview_admin FOREIGN KEY (reviewer_admin_id) REFERENCES public.admin(user_id);


--
-- TOC entry 3528 (class 2606 OID 16763)
-- Name: providerreview fk_provreview_application; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providerreview
    ADD CONSTRAINT fk_provreview_application FOREIGN KEY (application_id) REFERENCES public.providerapplication(application_id) ON DELETE CASCADE;


--
-- TOC entry 3529 (class 2606 OID 16768)
-- Name: queue fk_queue_item; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.queue
    ADD CONSTRAINT fk_queue_item FOREIGN KEY (item_id) REFERENCES public.item(item_id) ON DELETE CASCADE;


--
-- TOC entry 3530 (class 2606 OID 16773)
-- Name: queueentry fk_queueentry_buyer; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.queueentry
    ADD CONSTRAINT fk_queueentry_buyer FOREIGN KEY (buyer_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3531 (class 2606 OID 16778)
-- Name: queueentry fk_queueentry_queue; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.queueentry
    ADD CONSTRAINT fk_queueentry_queue FOREIGN KEY (queue_id) REFERENCES public.queue(queue_id) ON DELETE CASCADE;


--
-- TOC entry 3532 (class 2606 OID 16783)
-- Name: rating fk_rating_rated_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rating
    ADD CONSTRAINT fk_rating_rated_user FOREIGN KEY (rated_user_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3533 (class 2606 OID 16788)
-- Name: rating fk_rating_rater; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rating
    ADD CONSTRAINT fk_rating_rater FOREIGN KEY (rater_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3534 (class 2606 OID 16793)
-- Name: rating fk_rating_transaction; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rating
    ADD CONSTRAINT fk_rating_transaction FOREIGN KEY (transaction_id) REFERENCES public.transaction(transaction_id) ON DELETE CASCADE;


--
-- TOC entry 3535 (class 2606 OID 16798)
-- Name: report fk_report_reported_item; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report
    ADD CONSTRAINT fk_report_reported_item FOREIGN KEY (reported_item_id) REFERENCES public.item(item_id) ON DELETE CASCADE;


--
-- TOC entry 3536 (class 2606 OID 16803)
-- Name: report fk_report_reported_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report
    ADD CONSTRAINT fk_report_reported_user FOREIGN KEY (reported_user_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3537 (class 2606 OID 16808)
-- Name: report fk_report_reporter; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report
    ADD CONSTRAINT fk_report_reporter FOREIGN KEY (reporter_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3538 (class 2606 OID 16813)
-- Name: student fk_student_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.student
    ADD CONSTRAINT fk_student_user FOREIGN KEY (user_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3539 (class 2606 OID 16818)
-- Name: transaction fk_transaction_buyer; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transaction
    ADD CONSTRAINT fk_transaction_buyer FOREIGN KEY (buyer_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3540 (class 2606 OID 16823)
-- Name: transaction fk_transaction_item; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transaction
    ADD CONSTRAINT fk_transaction_item FOREIGN KEY (item_id) REFERENCES public.item(item_id) ON DELETE CASCADE;


--
-- TOC entry 3541 (class 2606 OID 16829)
-- Name: transaction fk_transaction_seller; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transaction
    ADD CONSTRAINT fk_transaction_seller FOREIGN KEY (seller_id) REFERENCES public."User"(user_id) ON DELETE CASCADE;


--
-- TOC entry 3503 (class 2606 OID 16668)
-- Name: User fk_user_university; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."User"
    ADD CONSTRAINT fk_user_university FOREIGN KEY (university_id) REFERENCES public.university(university_id);


-- Completed on 2026-02-04 03:31:24

--
-- PostgreSQL database dump complete
--

