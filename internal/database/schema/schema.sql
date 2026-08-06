-- Final schema state after all migrations
-- This is used by sqlc for type generation

CREATE TABLE setting (
	id     TEXT PRIMARY KEY NOT NULL,
	key    TEXT UNIQUE NOT NULL,
	value  TEXT DEFAULT NULL
);

CREATE TABLE session (
	key      TEXT UNIQUE NOT NULL,
	value    TEXT DEFAULT NULL,
	expires  INTEGER
);

CREATE TABLE subdomain (
	id    TEXT PRIMARY KEY NOT NULL,
	name  TEXT UNIQUE NOT NULL,
	desc  TEXT DEFAULT NULL
);

CREATE TABLE page (
	id 				TEXT PRIMARY KEY NOT NULL,
	name 			TEXT NOT NULL,
	slug 			TEXT UNIQUE NOT NULL,
	content 	TEXT DEFAULT NULL,
	position  TEXT NOT NULL CHECK (position == 'header' OR position == 'footer'),
	active    BOOLEAN DEFAULT FALSE NOT NULL,
	created 	TIMESTAMP DEFAULT (datetime('now')),
	updated 	TIMESTAMP
);

CREATE TABLE product (
	id         TEXT PRIMARY KEY NOT NULL,
	name       TEXT NOT NULL,
	desc       TEXT NOT NULL,
	slug       TEXT UNIQUE NOT NULL,
	amount     NUMERIC NOT NULL,
	metadata   JSON DEFAULT '[]' NOT NULL,
	attribute  JSON DEFAULT '[]' NOT NULL,
	digital    TEXT CHECK (digital == 'file' OR digital == 'data' OR digital == 'api'),
	active     BOOLEAN DEFAULT TRUE NOT NULL,
	deleted    BOOLEAN DEFAULT FALSE NOT NULL,
	created    TIMESTAMP DEFAULT (datetime('now')),
	updated    TIMESTAMP
);

CREATE TABLE digital_file (
	id            TEXT PRIMARY KEY NOT NULL,
	product_id    TEXT NOT NULL,
	name          TEXT NOT NULL,
	ext           TEXT NOT NULL,
	orig_name     TEXT NOT NULL,
	FOREIGN KEY (product_id) REFERENCES product(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE digital_data (
	id            TEXT PRIMARY KEY NOT NULL,
	product_id    TEXT NOT NULL,
	content       TEXT NOT NULL,
	cart_id       TEXT DEFAULT NULL,
	FOREIGN KEY (product_id) REFERENCES product(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE product_image (
	id          TEXT PRIMARY KEY NOT NULL,
	product_id  TEXT NOT NULL,
	name        TEXT NOT NULL,
	ext         TEXT NOT NULL,
	orig_name   TEXT NOT NULL,
	FOREIGN KEY (product_id) REFERENCES product(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE cart (
	id              TEXT PRIMARY KEY NOT NULL,
	email           TEXT DEFAULT NULL,
	amount_total    NUMERIC NOT NULL,
	currency        TEXT NOT NULL,
	payment_id      TEXT DEFAULT NULL,
	payment_status  TEXT DEFAULT NULL,
	payment_system  TEXT NOT NULL DEFAULT '',
	created         TIMESTAMP DEFAULT (datetime('now')),
	updated         TIMESTAMP
);

CREATE TABLE cart_product (
	id          TEXT PRIMARY KEY NOT NULL,
	cart_id     TEXT NOT NULL,
	product_id  TEXT NOT NULL,
	quantity    NUMERIC DEFAULT NULL,
	amount      NUMERIC DEFAULT NULL,
	FOREIGN KEY (cart_id) REFERENCES cart(id) ON UPDATE CASCADE ON DELETE CASCADE,
	FOREIGN KEY (product_id) REFERENCES product(id) ON UPDATE CASCADE ON DELETE CASCADE
);
