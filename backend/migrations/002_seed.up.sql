-- Seed data for development

-- ============================================
-- Sports
-- ============================================
INSERT INTO sports (name, slug, icon) VALUES
    ('Football', 'football', '⚽');

-- ============================================
-- Admin user (password: admin123)
-- Hash generated with bcrypt cost 12
-- ============================================
INSERT INTO users (email, password_hash, balance, role) VALUES
    ('admin@betkz.com', '$2a$12$Md/UBS4WXxiwE8i68S4bru52PT5BwElNp6fFYK5uBKACMWQjfLWC2', 10000.00, 'admin');

-- ============================================
-- Test user (password: user1234)
-- ============================================
INSERT INTO users (email, password_hash, balance, role) VALUES
    ('user@betkz.com', '$2a$12$ilr4oGOPDcXxZS/M2SJYKePL3Z/Z0j2U9AlPNKor.XsTWp.r.0nvq', 500.00, 'user');
