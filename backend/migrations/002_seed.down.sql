-- Rollback seed data
DELETE FROM market_pools;
DELETE FROM odds;
DELETE FROM markets;
DELETE FROM events;
DELETE FROM sports;
DELETE FROM users WHERE email IN ('admin@betkz.com', 'user@betkz.com');
