-- INIT
CREATE TABLE IF NOT EXISTS chat (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    safe        TEXT    NOT NULL,
    groupName   TEXT    NOT NULL,
    message     TEXT    NOT NULL,
    createdAt   INTEGER NOT NULL,
    creatorId   TEXT    NOT NULL,
    contentType TEXT    NOT NULL
);

-- INIT
CREATE INDEX IF NOT EXISTS chat_safe_groupName_createdAt ON chat (safe, groupName, createdAt);

-- INSERT_MESSAGE
INSERT INTO chat (safe, groupName, message, createdAt, creatorId, contentType) VALUES (:safe, :groupName, :message, :createdAt, :creatorId, :contentType);

-- GET_MESSAGES
SELECT message, createdAt, creatorId, contentType FROM chat WHERE safe=:safe AND groupName = :groupName
ORDER BY createdAt DESC LIMIT :limit

