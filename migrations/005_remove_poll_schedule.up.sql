-- Remove poll_schedule column from teams table (no longer needed)
ALTER TABLE projections.teams DROP COLUMN IF EXISTS poll_schedule;
