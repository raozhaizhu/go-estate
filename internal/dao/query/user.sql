-- name: CreateUser :execresult
INSERT INTO
        users(
                username,
                hashed_password,
                email,
                role
        )
VALUES
        (?, ?, ?, COALESCE(sqlc.narg(role), 1));

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
        email = COALESCE(sqlc.narg(email), email)
WHERE
        username = sqlc.arg(username);