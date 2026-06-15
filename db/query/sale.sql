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
        date = $1;

-- name: GetDataByPeriod :many
SELECT
        *
FROM
        daily_data
WHERE
        date >= $1
        AND date <= $2;

