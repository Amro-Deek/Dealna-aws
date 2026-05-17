-- Remove "Free / Giveaways" category and any "Test" category from the database.
-- Items assigned to these categories will have their category set to NULL (FK is ON DELETE SET NULL).
DELETE FROM public.category
WHERE name IN ('Free / Giveaways', 'Test', 'Test Category');
