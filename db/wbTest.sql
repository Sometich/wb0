CREATE TABLE payments
(
    transaction   VARCHAR(64) primary key,
    request_id    VARCHAR(64),
    currency      VARCHAR(64),
    provider      VARCHAR(64),
    amount        BIGINT,
    payment_dt    BIGINT,
    bank          VARCHAR(64),
    delivery_cost BIGINT,
    goods_total   BIGINT,
    custom_fee    BIGINT

);

CREATE TABLE orders
(
    order_uid          VARCHAR primary key references payments (transaction),
    track_number       VARCHAR unique ,
    entry              VARCHAR,
    locale             VARCHAR,
    internal_signature VARCHAR,
    customer_id        VARCHAR,
    delivery_service   VARCHAR,
    shardkey           VARCHAR,
    sm_id              BIGINT,
    date_created       TIMESTAMP,
    oof_shard          VARCHAR
);

CREATE TABLE items
(
    chrt_id      BIGINT,
    track_number VARCHAR REFERENCES orders (track_number),
    price        BIGINT,
    rid          VARCHAR,
    name         VARCHAR,
    sale         BIGINT,
    size         VARCHAR,
    total_price  BIGINT,
    nm_id        BIGINT,
    brand        VARCHAR,
    status       BIGINT
);

CREATE TABLE deliveries
(
    id      SERIAL PRIMARY KEY ,
    name    VARCHAR(64),
    phone   VARCHAR(64),
    zip     VARCHAR(64),
    city    VARCHAR(64),
    address VARCHAR(64),
    region  VARCHAR(64),
    email   VARCHAR(64)
);

CREATE TABLE order_deliveries
(
    order_uid   VARCHAR(64) REFERENCES orders (order_uid),
    delivery_id BIGINT REFERENCES deliveries (id)
);
