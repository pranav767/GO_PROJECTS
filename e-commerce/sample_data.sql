-- SQL to create and populate products and cart tables for testing

CREATE TABLE products (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  price DECIMAL(10,2) NOT NULL,
  inventory INT NOT NULL
);

CREATE TABLE cart (
  id INT AUTO_INCREMENT PRIMARY KEY,
  userid INT NOT NULL,
  items JSON NOT NULL,
  FOREIGN KEY (userid) REFERENCES users(id)
);

-- Sample products
INSERT INTO products (name, description, price, inventory) VALUES
('Laptop', 'A fast laptop', 999.99, 10),
('Phone', 'A smart phone', 499.99, 20),
('Headphones', 'Noise cancelling', 199.99, 15);
