-- INIT 
CREATE TABLE db_test (
  msg VARCHAR(255) NOT NULL,
  cnt INT DEFAULT 0,
  ratio FLOAT DEFAULT 0.0,
  bin BLOB,
  PRIMARY KEY (msg)
)

-- INSERT_TEST_DATA
INSERT INTO db_test(msg, cnt, ratio, bin) VALUES (:msg, :cnt, :ratio, :bin)

-- SELECT_TEST_DATA
SELECT msg, cnt, ratio, bin FROM db_test
