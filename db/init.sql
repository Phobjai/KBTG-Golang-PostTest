-- Create table if it does not exist
CREATE TABLE  "admin_config" (
    "deduction" double precision,
    "kreceipt" double precision
);

-- Insert initial values
INSERT INTO "admin_config" ("deduction", "kreceipt")
VALUES (60000, 50000);

SELECT * FROM "admin_config"