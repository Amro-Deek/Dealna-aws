ALTER TABLE public.transaction 
  ADD COLUMN seller_confirmed boolean DEFAULT false NOT NULL,
  ADD COLUMN buyer_confirmed boolean DEFAULT false NOT NULL,
  ADD COLUMN updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL;
