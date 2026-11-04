ALTER TABLE generations
	ADD COLUMN IF NOT EXISTS stop_reason TEXT;
