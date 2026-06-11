ALTER TABLE "User"
DROP COLUMN IF EXISTS total_ratings,
DROP COLUMN IF EXISTS sum_ratings,
DROP COLUMN IF EXISTS bayesian_rating;
