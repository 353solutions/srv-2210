CREATE TABLE IF NOT EXISTS rides (
    id TEXT PRIMARY KEY,
    driver TEXT NOT NULL,
    kind TEXT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    distance FLOAT
);

CREATE INDEX IF NOT EXISTS rides_start ON rides(start_time);
CREATE INDEX IF NOT EXISTS rides_end ON rides(end_time);

