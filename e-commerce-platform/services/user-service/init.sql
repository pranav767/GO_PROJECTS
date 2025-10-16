CREATE DATABASE IF NOT EXISTS e_commerce_users;
USE e_commerce_users;

CREATE TABLE IF NOT EXISTS users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL
);

INSERT IGNORE INTO users (username, password_hash) VALUES ('demo', 'demo');
