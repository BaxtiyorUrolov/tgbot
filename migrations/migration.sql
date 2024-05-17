
CREATE TYPE "status_type" AS ENUM ('in_process', 'done', 'cancel');

CREATE TABLE IF NOT EXISTS users (
    id Serial PRIMARY KEY NOT NULL,
    user_id BIGINT UNIQUE NOT NULL,
    name VARCHAR(30),
    phone VARCHAR(30) UNIQUE,
    order_time TIMESTAMP,
    last_order_date VARCHAR(30),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE Table IF NOT EXISTS barbers (
  id UUID PRIMARY KEY NOT NULL,
  name varchar(30),
  phone varchar(13)
);

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY NOT NULL,
    order_time varchar(10),
    order_date varchar(10),
    user_id uuid [ref: > users.user_id],
    barber_id uuid [ref : > barbers.id],
    status status_type DEFAULT 'in_process',
    created_at TIMESTAMP [DEFAULT: `NOW()`]
);
