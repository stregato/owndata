-- INIT
CREATE TABLE IF NOT EXISTS mio_configs (
    node    VARCHAR(128) NOT NULL, 
    k       VARCHAR(64) NOT NULL, 
    s       VARCHAR(64) NOT NULL,
    i       INTEGER NOT NULL,
    b       BLOB,
    CONSTRAINT pk_safe_key PRIMARY KEY(node,k)
);

-- GET_CONFIG
SELECT s, i, b FROM mio_configs WHERE node=:node AND k=:key

-- SET_CONFIG
INSERT INTO mio_configs(node,k,s,i,b) VALUES(:node,:key,:s,:i,:b)
	ON CONFLICT(node,k) DO UPDATE SET s=:s,i=:i,b=:b
	WHERE node=:node AND k=:key

-- DEL_CONFIG
DELETE FROM mio_configs WHERE node=:node

-- LIST_CONFIG
SELECT k FROM mio_configs WHERE node=:node

-- INIT
CREATE TABLE IF NOT EXISTS mio_files (
    id VARCHAR(256) PRIMARY KEY,
    safeID      VARCHAR(256),
    name        VARCHAR(256),
    dir         VARCHAR(4096),
    creator     VARCHAR(256),
    groupName   VARCHAR(256),
    tags        VARCHAR(4096),
    localPath   VARCHAR(4096),
    encryptionKey VARCHAR(256),
    modTime     INTEGER,
    size        INTEGER,
    attributes  BLOB
);

-- INIT
CREATE INDEX IF NOT EXISTS idx_mio_files_dir ON mio_files(dir)

-- INIT
CREATE INDEX IF NOT EXISTS idx_mio_files_groupName ON mio_files(groupName)

-- INIT
CREATE INDEX IF NOT EXISTS idx_mio_files_tags ON mio_files(tags)

-- INIT
CREATE INDEX IF NOT EXISTS idx_mio_files_modTime ON mio_files(modTime)

-- INIT
CREATE INDEX IF NOT EXISTS idx_mio_files_name ON mio_files(name)

-- INSERT_FILE
INSERT INTO mio_files(id,safeID,name,dir,creator,groupName,tags,localPath,encryptionKey,modTime,size,attributes) 
    VALUES(:id,:safeID,:name,:dir,:creator,:groupName,:tags,:localPath,:encryptionKey,:modTime,:size,:attributes)
    ON CONFLICT(id) DO UPDATE SET safeID=:safeID,name=:name,dir=:dir,creator=:creator,groupName=:groupName,tags=:tags,localPath=:localPath,encryptionKey=:encryptionKey,modTime=:modTime,size=:size,attributes=:attributes
    WHERE id=:id

-- GET_LAST_ID
SELECT id FROM mio_files WHERE dir=:dir ORDER BY id DESC LIMIT 1

-- GET_FILES_BY_DIR
SELECT id,name,dir,groupName,tags,modTime,size,creator,attributes,localPath,encryptionKey FROM mio_files 
    WHERE dir=:dir AND safeID=:safeID
    AND (:name = '' OR name = :name)
    AND (:groupName = '' OR groupName = :groupName)
    AND (:tag = '' OR tag LIKE '% ' || :tag || ' %')
    AND (:creator = '' OR creator = :creator)
    AND (:before < 0 OR modTime < :before)
    AND (:after < 0 OR modTime > :after)
    AND (:prefix = '' OR name LIKE :prefix || '%')
    AND (:suffix = '' OR name LIKE '%' || :suffix)
    #ORDER_BY
    LIMIT CASE WHEN :limit = 0 THEN -1 ELSE :limit END OFFSET :offset

-- GET_FILE_BY_NAME
SELECT id,size,encryptionKey FROM mio_files 
    WHERE safeID=:safeID AND dir=:dir AND name=:name

-- GET_GROUP_NAME 
SELECT DISTINCT groupName FROM mio_files WHERE safeID=:safeID AND dir = :dir AND name = :name 

-- INIT
CREATE TABLE IF NOT EXISTS mio_file_async (
    safeID      VARCHAR(256)    NOT NULL,
    id          VARCHAR(256)    NOT NULL,
    deleteSrc   INTEGER         NOT NULL,
    localPath   VARCHAR(4096)   NOT NULL,
    operation   VARCHAR(0)      NOT NULL,
    file        BLOB NOT        NULL,
    data        BLOB,
    CONSTRAINT pk_mio_file_async PRIMARY KEY(safeID,id)
);

-- INSERT_FILE_ASYNC
INSERT INTO mio_file_async(safeID,id,deleteSrc,localPath,operation,file,data) 
    VALUES(:safeID,:id,:deleteSrc,:localPath,:operation,:file,:data)

-- GET_FILE_ASYNC
SELECT file,data, deleteSrc, localPath, operation FROM mio_file_async WHERE safeID=:safeID AND id=:id

-- GET_FILES_ASYNC
SELECT id,file,data, deleteSrc, localPath, operation FROM mio_file_async WHERE safeID=:safeID

-- DEL_FILE_ASYNC
DELETE FROM mio_file_async WHERE safeID=:safeID AND id=:id

-- INIT
CREATE TABLE IF NOT EXISTS mio_tx (
    safeID      TEXT, 
    groupName   TEXT, 
    kind        TEXT,
    id      TEXT,
    CONSTRAINT pk_mio_tx_key PRIMARY KEY(safeID,groupName,kind)
)

-- MIO_STORE_TX
INSERT INTO mio_tx (safeID,groupName,kind,id) VALUES(:safeID,:groupName,:kind,:id)
    ON CONFLICT(safeID,groupName,kind) DO UPDATE SET id=:id
    WHERE safeID=:safeID AND groupName=:groupName AND kind=:kind

-- MIO_GET_TX
SELECT kind, id FROM mio_tx WHERE safeID = :safeID AND groupName = :groupName

-- MIO_DEL_TX_KIND
DELETE FROM mio_tx WHERE safeID = :safeID AND groupName = :groupName AND kind = :kind
