INSERT INTO
        `users`(
                username,
                hashed_password,
                email,
                role
        )
VALUES
        (
                'Bob',
                -- 密码:123456
                '$2a$10$.FSgaTYBmJlxmiJAvy3haOPbYN3lExDljBm.SgoE8agZbLBicjx1m',
                'bob@example.com',
                1
        ),
        (
                'Alice',
                '$2a$10$.FSgaTYBmJlxmiJAvy3haOPbYN3lExDljBm.SgoE8agZbLBicjx1m',
                'alice@example.com',
                2
        ),
        (
                'Admin',
                '$2a$10$.FSgaTYBmJlxmiJAvy3haOPbYN3lExDljBm.SgoE8agZbLBicjx1m',
                'admin@example.com',
                3
        );