-- INIT
CREATE TABLE IF NOT EXISTS mio_configs (
    node    VARCHAR(128) NOT NULL, 
    k       VARCHAR(64) NOT NULL, 
    s       VARCHAR(64) NOT NULL,
    i       INTEGER NOT NULL,
    b       BLOB,
    CONSTRAINT pk_safe_key PRIMARY KEY(node,k)
);

-- MIO_GET_CONFIG
SELECT s, i, b FROM mio_configs WHERE node=:node AND k=:key

-- MIO_SET_CONFIG
INSERT INTO mio_configs(node,k,s,i,b) VALUES(:node,:key,:s,:i,:b)
	ON CONFLICT(node,k) DO UPDATE SET s=:s,i=:i,b=:b
	WHERE node=:node AND k=:key

-- MIO_DEL_CONFIG
DELETE FROM mio_configs WHERE node=:node

-- MIO_LIST_CONFIG
SELECT k FROM mio_configs WHERE node=:node

-- INIT
CREATE TABLE IF NOT EXISTS mio_files (
    safeID          VARCHAR(256)    NOT NULL,
    name            VARCHAR(256)    NOT NULL,
    dir             VARCHAR(4096)   NOT NULL,
    id              INTEGER         NOT NULL,
    creator         VARCHAR(256)    NOT NULL,
    groupName       VARCHAR(256)    NOT NULL,
    tags            VARCHAR(4096)   NOT NULL,
    localPath       VARCHAR(4096)   NOT NULL,
    encryptionKey   VARCHAR(256)    NOT NULL,
    modTime         INTEGER         NOT NULL,
    size            INTEGER         NOT NULL,
    attributes      BLOB,
    PRIMARY KEY(safeID, name, dir, id)
);

-- INIT
CREATE INDEX IF NOT EXISTS idx_mio_files_id ON mio_files(id)

-- INIT
CREATE INDEX IF NOT EXISTS idx_mio_files_groupName ON mio_files(groupName)

-- INIT
CREATE INDEX IF NOT EXISTS idx_mio_files_tags ON mio_files(tags)

-- INIT
CREATE INDEX IF NOT EXISTS idx_mio_files_modTime ON mio_files(modTime)

-- INIT
CREATE INDEX IF NOT EXISTS idx_mio_files_name ON mio_files(name)

-- MIO_STORE_FILE
INSERT INTO mio_files(safeID,name,dir,id,creator,groupName,tags,localPath,encryptionKey,modTime,size,attributes) 
    VALUES(:safeID,:name,:dir,:id,:creator,:groupName,:tags,:localPath,:encryptionKey,:modTime,:size,:attributes)
    ON CONFLICT(safeID,name,dir,id) DO UPDATE SET creator=:creator,groupName=:groupName,tags=:tags,localPath=:localPath,encryptionKey=:encryptionKey,modTime=:modTime,size=:size,attributes=:attributes
    WHERE id=:id AND safeID=:safeID AND name=:name AND dir=:dir

-- MIO_STORE_DIR
INSERT INTO mio_files(safeID,name,dir,id,creator,groupName,tags,localPath,encryptionKey,modTime,size) 
    VALUES(:safeID,:name,:dir,'','','','','','',0,0)
    ON CONFLICT(safeID,name,dir,id) DO NOTHING

-- MIO_GET_LAST_ID
SELECT id FROM mio_files WHERE dir=:dir ORDER BY id DESC LIMIT 1

-- MIO_GET_FILES_BY_DIR
SELECT id,name,dir,groupName,tags,modTime,size,creator,attributes,localPath,encryptionKey FROM mio_files 
    WHERE dir=:dir AND safeID=:safeID
    AND (:name = '' OR name = :name)
    AND (:groupName = '' OR groupName = :groupName)
    AND (:tag = '' OR tags LIKE '% ' || :tag || ' %')
    AND (:creator = '' OR creator = :creator)
    AND (:before < 0 OR modTime < :before)
    AND (:after < 0 OR modTime > :after)
    AND (:prefix = '' OR name LIKE :prefix || '%')
    AND (:suffix = '' OR name LIKE '%' || :suffix)
    #orderBy
    LIMIT CASE WHEN :limit = 0 THEN -1 ELSE :limit END OFFSET :offset

-- MIO_GET_FILE_BY_NAME
SELECT  id,dir,groupName,tags,modTime,size,creator,attributes,localPath,encryptionKey  FROM mio_files 
    WHERE safeID=:safeID AND dir=:dir AND name=:name

-- MIO_GET_GROUP_NAME 
SELECT DISTINCT groupName FROM mio_files WHERE safeID=:safeID AND dir = :dir AND name = :name 

-- INIT
CREATE TABLE IF NOT EXISTS mio_file_async (
    safeID      VARCHAR(256)    NOT NULL,
    id          VARCHAR(256)    NOT NULL,
    deleteSrc   INTEGER         NOT NULL,
    localCopy   VARCHAR(4096)   NOT NULL,
    operation   VARCHAR(0)      NOT NULL,
    file        BLOB NOT        NULL,
    data        BLOB,
    CONSTRAINT pk_mio_file_async PRIMARY KEY(safeID,id)
);

-- MIO_INSERT_FILE_ASYNC
INSERT INTO mio_file_async(safeID,id,deleteSrc,localCopy,operation,file,data) 
    VALUES(:safeID,:id,:deleteSrc,:localCopy,:operation,:file,:data)

-- MIO_GET_FILE_ASYNC
SELECT file,data, deleteSrc, localCopy, operation FROM mio_file_async WHERE safeID=:safeID AND id=:id

-- MIO_GET_FILES_ASYNC
SELECT id,file,data, deleteSrc, localCopy, operation FROM mio_file_async WHERE safeID=:safeID

-- MIO_DEL_FILE_ASYNC
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
