DROP DATABASE IF EXISTS blueprint;
CREATE DATABASE blueprint;
USE blueprint;

-- TODO: Change password type when salted hash is implemented
CREATE TABLE account (
    user_id INT UNSIGNED,
    username VARCHAR(16) NOT NULL,
    password BINARY(60) NOT NULL,
    PRIMARY KEY (user_id)
);

-- TODO: Change token type when salted hash is implemented
CREATE TABLE token (
    user_id INT UNSIGNED,
    access  CHAR(32) NOT NULL,
    refresh CHAR(32) NOT NULL,
    access_expire  INT(11) UNSIGNED NOT NULL,
    refresh_expire INT(11) UNSIGNED NOT NULL,
    FOREIGN KEY (user_id) REFERENCES account(user_id),
    PRIMARY KEY (user_id)
);

CREATE TABLE inventory (
    user_id  INT UNSIGNED,
    item_id  INT UNSIGNED,
    quantity INT UNSIGNED NOT NULL,
    FOREIGN KEY (user_id) REFERENCES account(user_id),
    PRIMARY KEY (user_id, item_id)
);

CREATE TABLE resources (
    spawn_id INT UNSIGNED,
    item_id  INT UNSIGNED NOT NULL,
    gcs_lat  DECIMAL NOT NULL,
    gcs_long DECIMAL NOT NULL,
    resource_expire INT(11) UNSIGNED NOT NULL,
    PRIMARY KEY (spawn_id)
);
