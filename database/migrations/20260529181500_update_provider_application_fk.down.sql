ALTER TABLE providerapplication DROP CONSTRAINT fk_provapp_applicant;
ALTER TABLE providerapplication ADD CONSTRAINT fk_provapp_applicant FOREIGN KEY (applicant_id) REFERENCES providerapplicant(applicant_id) ON DELETE CASCADE;
