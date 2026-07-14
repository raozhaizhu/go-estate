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
        DATE(date) = DATE(sqlc.arg(target_date));

-- name: GetDataByPeriod :many
SELECT
        *
FROM
        daily_data
WHERE
        DATE(date) >= DATE(sqlc.arg(start_date))
        AND DATE(date) <= DATE(sqlc.arg(end_date));