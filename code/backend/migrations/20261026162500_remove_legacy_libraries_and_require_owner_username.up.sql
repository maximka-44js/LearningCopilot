DELETE FROM libraries
WHERE owner_username IS NULL;

ALTER TABLE libraries
ALTER COLUMN owner_username SET NOT NULL;
