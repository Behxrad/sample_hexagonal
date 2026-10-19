CREATE SCHEMA IF NOT EXISTS inventory;

CREATE SEQUENCE inventory.product_id_seq
    START WITH 1
    INCREMENT BY 1
    MINVALUE 1;

CREATE TABLE inventory.products
(
    id    VARCHAR(32) PRIMARY KEY DEFAULT nextval('inventory.product_id_seq')::TEXT,
    name  VARCHAR(255) NOT NULL,
    count INTEGER      NOT NULL,
    price BIGINT       NOT NULL
);