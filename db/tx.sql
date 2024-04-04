-- INIT
CREATE TABLE IF NOT EXISTS mio_tx (
    id TEXT, 
    storeUrl TEXT, 
    groupName TEXT, 
    version REAL, 
    consumed INTEGER,
    PRIMARY KEY(id)
)

-- MIO_STORE_TX
INSERT INTO mio_tx (id, storeUrl, groupName, version, consumed) VALUES (:id, :storeUrl, :groupName, :version, :consumed)
ON CONFLICT(id) DO UPDATE SET consumed = :consumed

-- MIO_GET_LAST_TX
SELECT id, groupName, version, consumed FROM mio_tx WHERE storeUrl = :storeUrl AND groupName = :groupName ORDER BY version DESC LIMIT 1
