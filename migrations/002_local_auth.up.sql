-- Local password auth (dev bootstrap only; password_hash unused when DEV_MODE is off)
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);
