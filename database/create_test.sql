DROP DATABASE IF EXISTS blueprint_test;
CREATE DATABASE blueprint_test;
USE blueprint_test;

CREATE TABLE account (
    user_id INT UNSIGNED,
    username VARCHAR(16) NOT NULL,
    password BINARY(60) NOT NULL,
    PRIMARY KEY (user_id)
);

CREATE TABLE token (
    pair_id INT UNSIGNED,
    user_id INT UNSIGNED NOT NULL,
    access  CHAR(64) NOT NULL,
    refresh CHAR(64) NOT NULL,
    access_expire  BIGINT NOT NULL,
    refresh_expire BIGINT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES account(user_id),
    PRIMARY KEY (pair_id)
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
    gcs_lat  DECIMAL(10,8) NOT NULL,
    gcs_long DECIMAL(11,8) NOT NULL,
    resource_expire BIGINT NOT NULL,
    PRIMARY KEY (spawn_id)
);
