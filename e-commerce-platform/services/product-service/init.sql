CREATE DATABASE IF NOT EXISTS e_commerce_products;
USE e_commerce_products;

CREATE TABLE IF NOT EXISTS products (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  price DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  inventory INT NOT NULL DEFAULT 0
);

INSERT INTO products (name, description, price, inventory) VALUES
("Sample Product 1", "A demo product", 9.99, 100),
("Sample Product 2", "Another product", 19.95, 50);
