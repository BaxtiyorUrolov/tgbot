-- Create status_type enum
CREATE TYPE "status_type" AS ENUM ('in_process', 'done', 'cancel');
CREATE TYPE "admin_type" AS ENUM ('admin', 'barber');

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY NOT NULL,
    user_id BIGINT UNIQUE NOT NULL,
    name VARCHAR(30),
    phone VARCHAR(30) UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create barbers table
CREATE TABLE IF NOT EXISTS barbers (
    id BIGINT UNIQUE NOT NULL,
    name VARCHAR(30) UNIQUE NOT NULL,
    user_name VARCHAR(30),
    phone VARCHAR(13),
    admin admin_type DEFAULT 'barber'
);

-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY NOT NULL,
    order_time varchar(5),
    order_date DATE NOT NULL,
    user_id BIGINT REFERENCES users(user_id),
    barber_name VARCHAR(30) REFERENCES barbers(name),
    status status_type DEFAULT 'in_process',
    deleted_at TIMESTAMP DEFAULT NULL
);
