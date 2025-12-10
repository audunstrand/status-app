-- Restore poll_schedule column
ALTER TABLE projections.teams ADD COLUMN poll_schedule VARCHAR(100);
