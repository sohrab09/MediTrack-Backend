--
-- PostgreSQL database dump
--

\restrict j0QJiG2rn0egJFrMeJgK1mkLWGHOch3Tc6xql8Tbrt5qV5eCT3UrhGR7RogYAaq

-- Dumped from database version 17.6
-- Dumped by pg_dump version 17.6

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
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: medicine_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.medicine_categories (
    id integer NOT NULL,
    name character varying(150) NOT NULL,
    status smallint DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT categories_status_check CHECK ((status = ANY (ARRAY[0, 1])))
);


--
-- Name: categories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.categories_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.categories_id_seq OWNED BY public.medicine_categories.id;


--
-- Name: customers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.customers (
    id integer NOT NULL,
    customer_name character varying(150) NOT NULL,
    mobile character varying(20) NOT NULL,
    email character varying(100),
    contact_person character varying(100),
    address text,
    city character varying(50),
    state character varying(50),
    zip character varying(20),
    country character varying(50) DEFAULT 'Bangladesh'::character varying,
    opening_balance numeric(12,2) DEFAULT 0.00,
    current_balance numeric(12,2) DEFAULT 0.00,
    status smallint DEFAULT 1,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: customers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.customers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: customers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.customers_id_seq OWNED BY public.customers.id;


--
-- Name: medicine_box_sizes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.medicine_box_sizes (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    total_pcs integer NOT NULL,
    status integer DEFAULT 1,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT medicine_box_sizes_total_pcs_check CHECK ((total_pcs > 0))
);


--
-- Name: medicine_box_sizes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.medicine_box_sizes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: medicine_box_sizes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.medicine_box_sizes_id_seq OWNED BY public.medicine_box_sizes.id;


--
-- Name: medicine_leaves; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.medicine_leaves (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    qty_per_leaf integer DEFAULT 1 NOT NULL,
    description text,
    status integer DEFAULT 1,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT medicine_leaves_qty_per_leaf_check CHECK ((qty_per_leaf > 0))
);


--
-- Name: medicine_leaves_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.medicine_leaves_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: medicine_leaves_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.medicine_leaves_id_seq OWNED BY public.medicine_leaves.id;


--
-- Name: medicine_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.medicine_types (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    status integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: medicine_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.medicine_types_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: medicine_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.medicine_types_id_seq OWNED BY public.medicine_types.id;


--
-- Name: medicine_units; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.medicine_units (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    status smallint DEFAULT 1 NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: medicine_units_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.medicine_units_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: medicine_units_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.medicine_units_id_seq OWNED BY public.medicine_units.id;


--
-- Name: medicines; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.medicines (
    id integer NOT NULL,
    code character varying(30) NOT NULL,
    barcode character varying(100),
    name character varying(255) NOT NULL,
    strength character varying(100),
    generic character varying(255),
    category_id integer NOT NULL,
    type_id integer,
    box_size_id integer NOT NULL,
    unit_id integer NOT NULL,
    leaf_id integer,
    selling_price numeric(10,2) DEFAULT 0 NOT NULL,
    mrp numeric(10,2) DEFAULT 0 NOT NULL,
    current_stock integer DEFAULT 0 NOT NULL,
    minimum_stock integer DEFAULT 0 NOT NULL,
    status smallint DEFAULT 1 NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT medicines_current_stock_non_negative CHECK ((current_stock >= 0))
);


--
-- Name: medicines_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.medicines_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: medicines_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.medicines_id_seq OWNED BY public.medicines.id;


--
-- Name: menus; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.menus (
    id integer NOT NULL,
    parent_id integer,
    label character varying(255) NOT NULL,
    path character varying(255) DEFAULT NULL::character varying,
    icon character varying(100) DEFAULT NULL::character varying,
    roles text[] NOT NULL,
    sort_order integer DEFAULT 0
);


--
-- Name: purchases; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.purchases (
    id integer NOT NULL,
    invoice_no character varying(50) NOT NULL,
    supplier_id integer NOT NULL,
    purchase_date date DEFAULT CURRENT_DATE NOT NULL,
    subtotal numeric(12,2) DEFAULT 0,
    discount numeric(12,2) DEFAULT 0,
    tax numeric(12,2) DEFAULT 0,
    total_amount numeric(12,2) DEFAULT 0,
    paid_amount numeric(12,2) DEFAULT 0,
    due_amount numeric(12,2) DEFAULT 0,
    payment_status character varying(20) DEFAULT 'unpaid'::character varying,
    status integer DEFAULT 1,
    note text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


--
-- Name: purchases_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.purchases_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: purchases_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.purchases_id_seq OWNED BY public.purchases.id;


--
-- Name: suppliers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.suppliers (
    id integer NOT NULL,
    supplier_name character varying(150) NOT NULL,
    mobile character varying(20) NOT NULL,
    email character varying(100),
    contact_person character varying(100),
    address text,
    city character varying(50),
    state character varying(50),
    zip character varying(20),
    country character varying(50) DEFAULT 'Bangladesh'::character varying,
    opening_balance numeric(12,2) DEFAULT 0.00,
    current_balance numeric(12,2) DEFAULT 0.00,
    status smallint DEFAULT 1,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT suppliers_current_balance_non_negative CHECK ((current_balance >= (0)::numeric)),
    CONSTRAINT suppliers_opening_balance_non_negative CHECK ((opening_balance >= (0)::numeric))
);


--
-- Name: suppliers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.suppliers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: suppliers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.suppliers_id_seq OWNED BY public.suppliers.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id integer NOT NULL,
    firstname character varying(100) NOT NULL,
    lastname character varying(100) NOT NULL,
    phone character varying(20) NOT NULL,
    email character varying(150) NOT NULL,
    password text NOT NULL,
    role integer NOT NULL,
    status integer DEFAULT 1,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: customers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers ALTER COLUMN id SET DEFAULT nextval('public.customers_id_seq'::regclass);


--
-- Name: medicine_box_sizes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicine_box_sizes ALTER COLUMN id SET DEFAULT nextval('public.medicine_box_sizes_id_seq'::regclass);


--
-- Name: medicine_categories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicine_categories ALTER COLUMN id SET DEFAULT nextval('public.categories_id_seq'::regclass);


--
-- Name: medicine_leaves id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicine_leaves ALTER COLUMN id SET DEFAULT nextval('public.medicine_leaves_id_seq'::regclass);


--
-- Name: medicine_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicine_types ALTER COLUMN id SET DEFAULT nextval('public.medicine_types_id_seq'::regclass);


--
-- Name: medicine_units id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicine_units ALTER COLUMN id SET DEFAULT nextval('public.medicine_units_id_seq'::regclass);


--
-- Name: medicines id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicines ALTER COLUMN id SET DEFAULT nextval('public.medicines_id_seq'::regclass);


--
-- Name: purchases id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchases ALTER COLUMN id SET DEFAULT nextval('public.purchases_id_seq'::regclass);


--
-- Name: suppliers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.suppliers ALTER COLUMN id SET DEFAULT nextval('public.suppliers_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: medicine_categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicine_categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: customers customers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_pkey PRIMARY KEY (id);


--
-- Name: medicine_box_sizes medicine_box_sizes_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicine_box_sizes
    ADD CONSTRAINT medicine_box_sizes_name_key UNIQUE (name);


--
-- Name: medicine_box_sizes medicine_box_sizes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicine_box_sizes
    ADD CONSTRAINT medicine_box_sizes_pkey PRIMARY KEY (id);


--
-- Name: medicine_leaves medicine_leaves_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicine_leaves
    ADD CONSTRAINT medicine_leaves_name_key UNIQUE (name);


--
-- Name: medicine_leaves medicine_leaves_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicine_leaves
    ADD CONSTRAINT medicine_leaves_pkey PRIMARY KEY (id);


--
-- Name: medicine_types medicine_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicine_types
    ADD CONSTRAINT medicine_types_pkey PRIMARY KEY (id);


--
-- Name: medicine_units medicine_units_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicine_units
    ADD CONSTRAINT medicine_units_name_key UNIQUE (name);


--
-- Name: medicine_units medicine_units_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicine_units
    ADD CONSTRAINT medicine_units_pkey PRIMARY KEY (id);


--
-- Name: medicines medicines_barcode_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicines
    ADD CONSTRAINT medicines_barcode_key UNIQUE (barcode);


--
-- Name: medicines medicines_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicines
    ADD CONSTRAINT medicines_code_key UNIQUE (code);


--
-- Name: medicines medicines_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicines
    ADD CONSTRAINT medicines_pkey PRIMARY KEY (id);


--
-- Name: menus menus_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.menus
    ADD CONSTRAINT menus_pkey PRIMARY KEY (id);


--
-- Name: purchases purchases_invoice_no_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchases
    ADD CONSTRAINT purchases_invoice_no_key UNIQUE (invoice_no);


--
-- Name: purchases purchases_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchases
    ADD CONSTRAINT purchases_pkey PRIMARY KEY (id);


--
-- Name: suppliers suppliers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.suppliers
    ADD CONSTRAINT suppliers_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_phone_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_phone_key UNIQUE (phone);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_customers_mobile; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_customers_mobile ON public.customers USING btree (mobile);


--
-- Name: idx_medicine_category; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_medicine_category ON public.medicines USING btree (category_id);


--
-- Name: idx_medicine_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_medicine_code ON public.medicines USING btree (code);


--
-- Name: idx_medicine_generic; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_medicine_generic ON public.medicines USING btree (generic);


--
-- Name: idx_medicine_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_medicine_name ON public.medicines USING btree (name);


--
-- Name: idx_medicine_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_medicine_status ON public.medicines USING btree (status);


--
-- Name: idx_medicine_units_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_medicine_units_name ON public.medicine_units USING btree (name);


--
-- Name: idx_medicine_units_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_medicine_units_status ON public.medicine_units USING btree (status);


--
-- Name: idx_suppliers_mobile; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_suppliers_mobile ON public.suppliers USING btree (mobile);


--
-- Name: unique_category_name_lower; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX unique_category_name_lower ON public.medicine_categories USING btree (lower((name)::text));


--
-- Name: medicines trg_update_medicines_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_update_medicines_updated_at BEFORE UPDATE ON public.medicines FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: purchases fk_purchase_supplier; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.purchases
    ADD CONSTRAINT fk_purchase_supplier FOREIGN KEY (supplier_id) REFERENCES public.suppliers(id);


--
-- Name: medicines medicines_box_size_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicines
    ADD CONSTRAINT medicines_box_size_id_fkey FOREIGN KEY (box_size_id) REFERENCES public.medicine_box_sizes(id);


--
-- Name: medicines medicines_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicines
    ADD CONSTRAINT medicines_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.medicine_categories(id);


--
-- Name: medicines medicines_leaf_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicines
    ADD CONSTRAINT medicines_leaf_id_fkey FOREIGN KEY (leaf_id) REFERENCES public.medicine_leaves(id);


--
-- Name: medicines medicines_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicines
    ADD CONSTRAINT medicines_type_id_fkey FOREIGN KEY (type_id) REFERENCES public.medicine_types(id);


--
-- Name: medicines medicines_unit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.medicines
    ADD CONSTRAINT medicines_unit_id_fkey FOREIGN KEY (unit_id) REFERENCES public.medicine_units(id);


--
-- Name: menus menus_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.menus
    ADD CONSTRAINT menus_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.menus(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict j0QJiG2rn0egJFrMeJgK1mkLWGHOch3Tc6xql8Tbrt5qV5eCT3UrhGR7RogYAaq

