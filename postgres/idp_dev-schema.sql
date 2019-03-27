CREATE database "idp_dev";
CREATE database "hydra";
GRANT ALL PRIVILEGES ON database hydra TO idp;
GRANT ALL PRIVILEGES ON database idp_dev TO idp;
--
-- PostgreSQL database dump
--

-- Dumped from database version 11.2 (Debian 11.2-1.pgdg90+1)
-- Dumped by pg_dump version 11.2 (Debian 11.2-1.pgdg90+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: log_last_changes(); Type: FUNCTION; Schema: public; Owner: idp
--

CREATE FUNCTION public.log_last_changes() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN UPDATE last_update SET tstamp = now() WHERE id = 1; RETURN NEW; END; $$;


ALTER FUNCTION public.log_last_changes() OWNER TO idp;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: ab_users; Type: TABLE; Schema: public; Owner: idp
--

CREATE TABLE public.ab_users (
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    username text NOT NULL,
    email text NOT NULL,
    password text NOT NULL,
    confirm_selector text,
    confirm_verifier text,
    confirmed boolean,
    attempt_count integer,
    last_attempt timestamp with time zone,
    locked timestamp with time zone,
    recover_selector text,
    recover_verifier text,
    recover_token_expiry timestamp with time zone,
    oauth_uid text,
    oauth_provider text,
    oauth_access_token text,
    oauth_refresh_token text,
    oauth_expiry timestamp with time zone,
    totp_secret_key text,
    sms_phone_number text,
    sms_seed_phone_number text,
    recovery_codes text
);


ALTER TABLE public.ab_users OWNER TO idp;

--
-- Name: ab_users_id_seq; Type: SEQUENCE; Schema: public; Owner: idp
--

CREATE SEQUENCE public.ab_users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ab_users_id_seq OWNER TO idp;

--
-- Name: ab_users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: idp
--

ALTER SEQUENCE public.ab_users_id_seq OWNED BY public.ab_users.id;


--
-- Name: casbin_rules; Type: TABLE; Schema: public; Owner: idp
--

CREATE TABLE public.casbin_rules (
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    p_type character varying(100),
    v0 character varying(100),
    v1 character varying(100),
    v2 character varying(100),
    v3 character varying(100),
    v4 character varying(100),
    v5 character varying(100)
);


ALTER TABLE public.casbin_rules OWNER TO idp;

--
-- Name: casbin_rules_id_seq; Type: SEQUENCE; Schema: public; Owner: idp
--

CREATE SEQUENCE public.casbin_rules_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.casbin_rules_id_seq OWNER TO idp;

--
-- Name: casbin_rules_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: idp
--

ALTER SEQUENCE public.casbin_rules_id_seq OWNED BY public.casbin_rules.id;


--
-- Name: last_update; Type: TABLE; Schema: public; Owner: idp
--

CREATE TABLE public.last_update (
    id integer NOT NULL,
    tstamp timestamp without time zone DEFAULT now()
);


ALTER TABLE public.last_update OWNER TO idp;

--
-- Name: last_update_id_seq; Type: SEQUENCE; Schema: public; Owner: idp
--

CREATE SEQUENCE public.last_update_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.last_update_id_seq OWNER TO idp;

--
-- Name: last_update_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: idp
--

ALTER SEQUENCE public.last_update_id_seq OWNED BY public.last_update.id;


--
-- Name: qor_admin_settings; Type: TABLE; Schema: public; Owner: idp
--

CREATE TABLE public.qor_admin_settings (
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    key text,
    resource text,
    user_id text,
    value text
);


ALTER TABLE public.qor_admin_settings OWNER TO idp;

--
-- Name: qor_admin_settings_id_seq; Type: SEQUENCE; Schema: public; Owner: idp
--

CREATE SEQUENCE public.qor_admin_settings_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.qor_admin_settings_id_seq OWNER TO idp;

--
-- Name: qor_admin_settings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: idp
--

ALTER SEQUENCE public.qor_admin_settings_id_seq OWNED BY public.qor_admin_settings.id;


--
-- Name: ab_users id; Type: DEFAULT; Schema: public; Owner: idp
--

ALTER TABLE ONLY public.ab_users ALTER COLUMN id SET DEFAULT nextval('public.ab_users_id_seq'::regclass);


--
-- Name: casbin_rules id; Type: DEFAULT; Schema: public; Owner: idp
--

ALTER TABLE ONLY public.casbin_rules ALTER COLUMN id SET DEFAULT nextval('public.casbin_rules_id_seq'::regclass);


--
-- Name: last_update id; Type: DEFAULT; Schema: public; Owner: idp
--

ALTER TABLE ONLY public.last_update ALTER COLUMN id SET DEFAULT nextval('public.last_update_id_seq'::regclass);


--
-- Name: qor_admin_settings id; Type: DEFAULT; Schema: public; Owner: idp
--

ALTER TABLE ONLY public.qor_admin_settings ALTER COLUMN id SET DEFAULT nextval('public.qor_admin_settings_id_seq'::regclass);


--
-- Name: ab_users ab_users_pkey; Type: CONSTRAINT; Schema: public; Owner: idp
--

ALTER TABLE ONLY public.ab_users
    ADD CONSTRAINT ab_users_pkey PRIMARY KEY (id);


--
-- Name: casbin_rules casbin_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: idp
--

ALTER TABLE ONLY public.casbin_rules
    ADD CONSTRAINT casbin_rules_pkey PRIMARY KEY (id);


--
-- Name: qor_admin_settings qor_admin_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: idp
--

ALTER TABLE ONLY public.qor_admin_settings
    ADD CONSTRAINT qor_admin_settings_pkey PRIMARY KEY (id);


--
-- Name: idx_ab_users_deleted_at; Type: INDEX; Schema: public; Owner: idp
--

CREATE INDEX idx_ab_users_deleted_at ON public.ab_users USING btree (deleted_at);


--
-- Name: idx_casbin_rules_deleted_at; Type: INDEX; Schema: public; Owner: idp
--

CREATE INDEX idx_casbin_rules_deleted_at ON public.casbin_rules USING btree (deleted_at);


--
-- Name: idx_qor_admin_settings_deleted_at; Type: INDEX; Schema: public; Owner: idp
--

CREATE INDEX idx_qor_admin_settings_deleted_at ON public.qor_admin_settings USING btree (deleted_at);


--
-- Name: uix_ab_users_email; Type: INDEX; Schema: public; Owner: idp
--

CREATE UNIQUE INDEX uix_ab_users_email ON public.ab_users USING btree (email);


--
-- Name: casbin_rules last_change; Type: TRIGGER; Schema: public; Owner: idp
--

CREATE TRIGGER last_change AFTER INSERT OR DELETE OR UPDATE ON public.casbin_rules FOR EACH STATEMENT EXECUTE PROCEDURE public.log_last_changes();


--
-- PostgreSQL database dump complete
--

INSERT INTO public.ab_users (id, created_at, updated_at, deleted_at, username, email, password, confirm_selector, confirm_verifier, confirmed, attempt_count, last_attempt, locked, recover_selector, recover_verifier, recover_token_expiry, oauth_uid, oauth_provider, oauth_access_token, oauth_refresh_token, oauth_expiry, totp_secret_key, sms_phone_number, sms_seed_phone_number, recovery_codes) VALUES (5, '2019-02-21 15:20:07.975000', '2019-02-21 15:20:10.125000', null, 'vmalyshev', 'vadym.malyshev@gmail.com', '$2a$10$lmWdGp8ZJsFz5wJ9X8fi7uZ95XTC6zcx/trmd/TBuR3znx6.egrVC', null, null, true, null, null, null, null, null, null, null, null, null, null, null, null, null, null, null);
INSERT INTO public.ab_users (id, created_at, updated_at, deleted_at, username, email, password, confirm_selector, confirm_verifier, confirmed, attempt_count, last_attempt, locked, recover_selector, recover_verifier, recover_token_expiry, oauth_uid, oauth_provider, oauth_access_token, oauth_refresh_token, oauth_expiry, totp_secret_key, sms_phone_number, sms_seed_phone_number, recovery_codes) VALUES (999, '2019-02-22 10:37:08.784409', '2019-02-22 10:37:08.784409', null, 'admin', 'admin@hiveon.net', '$2a$10$lmWdGp8ZJsFz5wJ9X8fi7uZ95XTC6zcx/trmd/TBuR3znx6.egrVC', null, null, true, null, null, null, null, null, null, null, null, null, null, null, null, null, null, null);
INSERT INTO public.ab_users (id, created_at, updated_at, deleted_at, username, email, password, confirm_selector, confirm_verifier, confirmed, attempt_count, last_attempt, locked, recover_selector, recover_verifier, recover_token_expiry, oauth_uid, oauth_provider, oauth_access_token, oauth_refresh_token, oauth_expiry, totp_secret_key, sms_phone_number, sms_seed_phone_number, recovery_codes) VALUES (4, '2019-02-22 15:20:57.353500', '2019-02-22 15:20:57.353500', null, 'test', 'test@test.com', '$2a$10$lmWdGp8ZJsFz5wJ9X8fi7uZ95XTC6zcx/trmd/TBuR3znx6.egrVC', null, null, null, null, null, null, null, null, null, null, null, null, null, null, null, null, null, null);
INSERT INTO public.ab_users (id, created_at, updated_at, deleted_at, username, email, password, confirm_selector, confirm_verifier, confirmed, attempt_count, last_attempt, locked, recover_selector, recover_verifier, recover_token_expiry, oauth_uid, oauth_provider, oauth_access_token, oauth_refresh_token, oauth_expiry, totp_secret_key, sms_phone_number, sms_seed_phone_number, recovery_codes) VALUES (1, '2019-02-19 12:48:51.274852', '2019-02-22 15:24:40.543033', null, 'malysheff', 'malysheff@ukr.net', '$2a$10$lmWdGp8ZJsFz5wJ9X8fi7uZ95XTC6zcx/trmd/TBuR3znx6.egrVC', '', '', true, 0, '0001-01-01 00:00:00.000000', '0001-01-01 00:00:00.000000', '', '', '0001-01-01 00:00:00.000000', '', '', '', '', '0001-01-01 00:00:00.000000', '', '', '', '');
INSERT INTO public.ab_users (id, created_at, updated_at, deleted_at, username, email, password, confirm_selector, confirm_verifier, confirmed, attempt_count, last_attempt, locked, recover_selector, recover_verifier, recover_token_expiry, oauth_uid, oauth_provider, oauth_access_token, oauth_refresh_token, oauth_expiry, totp_secret_key, sms_phone_number, sms_seed_phone_number, recovery_codes) VALUES (2, '2019-02-21 14:17:37.713336', '2019-02-22 15:26:13.245852', null, '', 'a@a.com', '$2a$10$xzVWBNdmGtyH0UZ4vOMrf.Z2h4FUiUCaUMJ5Pag4DW3aPWudrUx1O', '', '', true, 0, '0001-01-01 00:00:00.000000', '0001-01-01 00:00:00.000000', '', '', '0001-01-01 00:00:00.000000', '', '', '', '', '0001-01-01 00:00:00.000000', '', '', '', '');
INSERT INTO public.ab_users (id, created_at, updated_at, deleted_at, username, email, password, confirm_selector, confirm_verifier, confirmed, attempt_count, last_attempt, locked, recover_selector, recover_verifier, recover_token_expiry, oauth_uid, oauth_provider, oauth_access_token, oauth_refresh_token, oauth_expiry, totp_secret_key, sms_phone_number, sms_seed_phone_number, recovery_codes) VALUES (3, '2019-02-21 14:18:39.012692', '2019-02-22 15:26:26.215404', null, '', 'vadym.malyshev@yahoo.com', '$2a$10$SigcZVyLv0gCOXGv3EAt.uA3GjoLWBPeTZVaeVJHWbElEMsRkate.', '', '', false, 0, '0001-01-01 00:00:00.000000', '0001-01-01 00:00:00.000000', '', '', '0001-01-01 00:00:00.000000', '', '', '', '', '0001-01-01 00:00:00.000000', '', '', '', '');