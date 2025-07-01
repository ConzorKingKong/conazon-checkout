CREATE SCHEMA checkout;
CREATE TABLE checkout.checkout (
  id SERIAL PRIMARY KEY NOT NULL,
  created_at TIMESTAMP DEFAULT NOW() NOT NULL,
  updated_at TIMESTAMP DEFAULT NOW() NOT NULL,
  user_id INT NOT NULL,
  total_price MONEY NOT NULL,
  cart_item_ids INTEGER[] NOT NULL DEFAULT '{}',
  cart_snapshot JSONB NOT NULL DEFAULT '[]'::jsonb
);
INSERT INTO checkout.checkout (user_id, total_price, cart_item_ids, cart_snapshot) 
VALUES 
    (1, 59.98, ARRAY[1, 2], '[{"cart_id": 1, "product_id": 1, "quantity": 1, "unit_price": 29.99}]'::jsonb),
    (1, 89.97, ARRAY[4, 5, 6], '[{"cart_id": 2, "product_id": 2, "quantity": 1, "unit_price": 29.99}]'::jsonb);