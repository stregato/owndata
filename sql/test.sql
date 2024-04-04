-- INIT
CREATE TABLE Users (
    security.UserId INTEGER PRIMARY KEY AUTOINCREMENT,
    Username TEXT NOT NULL,
    Email TEXT NOT NULL UNIQUE,
    RegistrationDate INT NOT NULL
);

-- INIT
CREATE INDEX idx_username ON Users(Username);

-- INIT
CREATE TABLE Posts (
    PostID INTEGER PRIMARY KEY AUTOINCREMENT,
    security.UserId INTEGER NOT NULL,
    Title TEXT NOT NULL,
    Content BLOB NOT NULL,
    PostDate INT NOT NULL,
    FOREIGN KEY (security.UserId) REFERENCES Users(security.UserId)
);

-- INIT
CREATE INDEX idx_postdate ON Posts(PostDate);

-- INSERT_USER
INSERT INTO Users (Username, Email, RegistrationDate) VALUES (:username, :email, :registrationDate);

-- INSERT_POST
INSERT INTO Posts (security.UserId, Title, Content, PostDate) VALUES (:security.UserId, :title, :content, :postDate);

-- SELECT_POSTS
SELECT Posts.Title, Posts.Content, Posts.PostDate
FROM Posts
INNER JOIN Users ON Posts.security.UserId = Users.security.UserId
WHERE Users.Username = :username;

-- UPDATE_POSTS
UPDATE Posts
SET Content = :content
WHERE PostID = :postId;
