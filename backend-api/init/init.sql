DROP TABLE IF EXISTS items;

CREATE TABLE items (
  id integer AUTO_INCREMENT,
  name varchar(255),
  desctiption varchar(255),
  amount integer,
  primary key(id)
);

INSERT INTO items (name, desctiption, amount)
VALUES
  ('toy', 'test-toy', 2000);

INSERT INTO items (name, desctiption, amount)
VALUES
  ('game', 'test-game', 6000);