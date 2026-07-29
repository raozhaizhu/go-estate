-- name: GetUserQueryRecord :execresult
SELECT
        *
FROM
        `user_query_records`
WHERE
        username = sqlc.arg(username)
        AND query_type = sqlc.arg(query_type)
        AND start_date = sqlc.arg(start_date)
        AND end_date = sqlc.arg(end_date)