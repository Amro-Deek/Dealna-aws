-- name: CreateProfile :exec
INSERT INTO profile (
    user_id,
    display_name
) VALUES (
    $1, $2
);