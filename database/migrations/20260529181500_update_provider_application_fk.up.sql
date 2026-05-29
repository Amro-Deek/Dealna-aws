ALTER TABLE providerapplication DROP CONSTRAINT IF EXISTS fk_provapp_applicant;
ALTER TABLE providerapplication DROP CONSTRAINT IF EXISTS fk_providerapplication_applicant;
ALTER TABLE providerapplication ADD CONSTRAINT fk_provapp_applicant FOREIGN KEY (applicant_id) REFERENCES "User"(user_id) ON DELETE CASCADE;
