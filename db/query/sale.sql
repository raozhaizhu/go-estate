-- name: GetAllData :many
SELECT
        *
FROM
        daily_data;

-- name: GetDataByDay :many
SELECT
        *
FROM
        daily_data
WHERE
        date = sqlc.arg(target_date);

-- name: GetDataByPeriod :many
SELECT
        *
FROM
        daily_data
WHERE
        date >= sqlc.arg(start_date)
        AND date <= sqlc.arg(end_date);

