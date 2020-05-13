CREATE TABLE items (
    id INT NOT NULL,
    label VARCHAR(255) NOT NULL,
    total INT,

    PRIMARY KEY (id, label)
);