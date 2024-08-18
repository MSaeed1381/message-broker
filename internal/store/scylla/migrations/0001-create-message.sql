CREATE KEYSPACE message_broker WITH replication = {'class': 'NetworkTopologyStrategy', 'replication_factor' : 1};


CREATE TABLE IF NOT EXISTS message_broker.message (
                                       id BIGINT PRIMARY KEY,
                                       body TEXT,
                                       createdAt TIMESTAMP,
                                       subject TEXT,
                                       expiration BIGINT
);

