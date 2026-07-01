-- name: CreateSession :exec
INSERT INTO
        `sessions`(
                id,
                username,
                device_id,
                user_agent,
                client_ip,
                is_blocked,
                expires_at
        )
VALUES
        (?, ?, ?, ?, ?, ?, ?);

-- name: GetSession :one
SELECT
        *
FROM
        `sessions`
WHERE
        id = ?
LIMIT
        1;

-- name: GetActiveSessionIDsByUserDevice :many
SELECT
        id
FROM
        `sessions`
WHERE
        username = sqlc.arg(username)
        AND device_id = sqlc.arg(device_id)
        AND is_blocked = false;

-- name: BlockSessionsByIDs :exec
UPDATE
        `sessions`
SET
        is_blocked = TRUE
WHERE
        id IN (sqlc.slice('ids'));

-- name: BlockAllUserSessions :exec
UPDATE
        `sessions`
SET
        is_blocked = TRUE
WHERE
        username = ?;

-- -- name: BlockSession :exec
-- UPDATE
--         `sessions`
-- SET
--         is_blocked = TRUE
-- WHERE
--         id = ?;
-- -- name: DeleteExpiredSessions :exec
-- DELETE FROM
--         `sessions`
-- WHERE
--         expires_at < NOW() - INTERVAL 7 DAY;
-- -- name: BlockUserDeviceSession :exec
-- UPDATE
--         `sessions`
-- SET
--         is_blocked = TRUE
-- WHERE
--         username = ?
--         AND device_id = ?;
-- -- name: GetSessionIDsByUserDevice :many
-- SELECT
--         id
-- FROM
--         `sessions`
-- WHERE
--         username = sqlc.arg(username)
--         AND device_id = sqlc.arg(device_id);