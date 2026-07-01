-- SET
--         NAMES utf8mb4;
-- SELECT
--         *
-- FROM
--         daily_data
-- SELECT
--         *
-- FROM
--         users
-- WHERE
--         username = 'Wrong'
-- SHOW ENGINE INNODB STATUS;
SHOW INDEX
FROM
        `users`
WHERE
        Column_name = 'username';