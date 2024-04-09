CREATE SCHEMA checkout;
CREATE TABLE checkout.checkout (
  id SERIAL PRIMARY KEY NOT NULL,
  created_at TIMESTAMP DEFAULT NOW() NOT NULL,
  updated_at TIMESTAMP DEFAULT NOW() NOT NULL,
  user_id INT NOT NULL,
  total_price MONEY NOT NULL,
  billing_status VARCHAR(255) NOT NULL,
  shipping_status VARCHAR(255) NOT NULL,
  tracking_number VARCHAR(255)
);
insert into checkout.checkout (user_id, total_price, billing_status, shipping_status, tracking_number) values (1, 120.00, 'paid', 'delivered', '1234567890');