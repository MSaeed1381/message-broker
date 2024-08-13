CREATE TABLE IF NOT EXISTS message (
                                     id SERIAL PRIMARY KEY,
                                     body VARCHAR NOT NULL,
                                     createdAt TIMESTAMP,
                                     subject VARCHAR NOT NULL,
                                     expiration BIGINT
);
