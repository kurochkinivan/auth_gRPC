ALTER TABLE users 
    ADD COLUMN created_at TIMESTAMPTZ 
        DEFAULT now() NOT NULL;

ALTER TABLE users 
    ADD COLUMN updated_at TIMESTAMPTZ 
        DEFAULT now() NOT NULL;