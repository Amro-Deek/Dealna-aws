ALTER TABLE public.transaction 
  DROP COLUMN seller_confirmed,
  DROP COLUMN buyer_confirmed,
  DROP COLUMN updated_at;
