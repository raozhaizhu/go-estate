-- name: CreateUser :execresult
INSERT INTO
        users(
                username,
                hashed_password,
                email,
                avatar_key,
                role
        )
VALUES
        (
                sqlc.arg(username),
                sqlc.arg(hashed_password),
                sqlc.arg(email),
                CASE
                        WHEN sqlc.arg(avatar_key) = '' THEN 'default_avatar.png'
                        ELSE sqlc.arg(avatar_key)
                END,
                COALESCE(sqlc.arg(role), 1)
        );

-- name: GetUser :one
SELECT
        *
FROM
        users
WHERE
        username = ?;

-- name: UpdateUser :execresult
UPDATE
        users
SET
        hashed_password = COALESCE(sqlc.narg(hashed_password), hashed_password),
        password_changed_at = COALESCE(
                sqlc.narg(password_changed_at),
                password_changed_at
        ),
        email = COALESCE(sqlc.narg(email), email),
        avatar_key = COALESCE(sqlc.narg(avatar_key), avatar_key)
WHERE
        username = sqlc.arg(username);

-- name: DecreasePoints :execresult
UPDATE
        users
SET
        points = points - sqlc.arg(amount)
WHERE
        username = sqlc.arg(username)
        AND points >= sqlc.arg(amount)