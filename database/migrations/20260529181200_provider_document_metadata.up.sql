ALTER TABLE providerapplicationdocument
ADD COLUMN document_type character varying(100),
ADD COLUMN original_filename character varying(500),
ADD COLUMN content_type character varying(100),
ADD COLUMN size_bytes bigint,
ADD COLUMN upload_status character varying(50) DEFAULT 'PENDING';
