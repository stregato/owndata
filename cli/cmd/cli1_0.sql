-- INIT
CREATE TABLE IF NOT EXISTS chat (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    safe        TEXT    NOT NULL,
    groupName   TEXT    NOT NULL,
    message     TEXT    NOT NULL,
    createdAt   INTEGER NOT NULL,
    creatorId   TEXT    NOT NULL
);

-- INSERT_MESSAGE
INSERT INTO chat (safe, groupName, message, createdAt, creatorId) VALUES (:safe, :groupName, :message, :createdAt, :creatorId);

-- GET_MESSAGES
SELECT message, createdAt, creatorId FROM chat WHERE safe=:safe AND groupName = :groupName 
ORDER BY createdAt DESC LIMIT :limit

