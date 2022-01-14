BEGIN;

ALTER TABLE todos RENAME COLUMN discription TO description;

COMMIT;