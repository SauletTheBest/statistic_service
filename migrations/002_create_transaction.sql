CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount NUMERIC(14,2) NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('income', 'expense')),
    category TEXT,
    category_id UUID REFERENCES categories(id),
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    wallet_id UUID REFERENCES wallets(id)
);
