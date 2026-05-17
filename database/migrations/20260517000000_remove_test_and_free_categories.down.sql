-- Re-insert the categories that were removed (rollback).
INSERT INTO public.category (name, description)
VALUES
  ('Free / Giveaways', 'Exchange items, donations, "تبرع"')
ON CONFLICT (name) DO NOTHING;
