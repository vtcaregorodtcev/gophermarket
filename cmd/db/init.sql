-- Create the users table if it doesn't exist
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    balance DECIMAL(10, 2) NOT NULL DEFAULT 0.00
);

-- Create the orders table if it doesn't exist
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    number VARCHAR(255) NOT NULL UNIQUE,
    FOREIGN KEY (user_id) REFERENCES users (id)
);

-- Insert a couple of users
INSERT INTO
    users (login, password)
SELECT
    'user1',
    'password1'
WHERE
    NOT EXISTS (
        SELECT
            1
        FROM
            users
        WHERE
            login = 'user1'
    )
UNION
ALL
SELECT
    'user2',
    'password2'
WHERE
    NOT EXISTS (
        SELECT
            1
        FROM
            users
        WHERE
            login = 'user2'
    );

-- Insert orders for the users
INSERT INTO
    orders (user_id, number)
SELECT
    u.id,
    'order1'
FROM
    users u
WHERE
    u.login = 'user1'
    AND NOT EXISTS (
        SELECT
            1
        FROM
            orders o
        WHERE
            o.user_id = u.id
            AND o.number = 'order1'
    )
UNION
ALL
SELECT
    u.id,
    'order2'
FROM
    users u
WHERE
    u.login = 'user1'
    AND NOT EXISTS (
        SELECT
            1
        FROM
            orders o
        WHERE
            o.user_id = u.id
            AND o.number = 'order2'
    )
UNION
ALL
SELECT
    u.id,
    'order3'
FROM
    users u
WHERE
    u.login = 'user2'
    AND NOT EXISTS (
        SELECT
            1
        FROM
            orders o
        WHERE
            o.user_id = u.id
            AND o.number = 'order3'
    );