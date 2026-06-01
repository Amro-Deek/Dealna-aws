-- name: CreateProviderApplication :one
INSERT INTO providerapplication (
    applicant_id,
    university_id,
    business_name,
    phone_number,
    business_type,
    address,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: UpdateProviderApplicationStatus :exec
UPDATE providerapplication
SET status = $2
WHERE application_id = $1;

-- name: UpdateProviderApplication :one
UPDATE providerapplication
SET university_id = $2,
    business_name = $3,
    phone_number = $4,
    business_type = $5,
    address = $6
WHERE applicant_id = $1
RETURNING *;

-- name: UpdateProviderApplicationReview :exec
UPDATE providerapplication
SET status = $2,
    reviewed_at = CURRENT_TIMESTAMP,
    reviewed_by_admin_id = $3,
    admin_comment = $4
WHERE application_id = $1;

-- name: GetProviderApplicationByApplicantID :one
SELECT * FROM providerapplication
WHERE applicant_id = $1;

-- name: CreateProviderApplicationDocument :one
INSERT INTO providerapplicationdocument (
    application_id,
    file_path,
    document_type,
    original_filename,
    content_type,
    size_bytes,
    upload_status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetProviderApplicationDocuments :many
SELECT * FROM providerapplicationdocument
WHERE application_id = $1;

-- name: CreateProvider :one
INSERT INTO provider (
    user_id,
    business_name,
    phone_number,
    business_type,
    address,
    verified_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;
