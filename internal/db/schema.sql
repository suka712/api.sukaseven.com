CREATE TABLE otps (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL,
  otp TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE sessions (
  token UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL
);