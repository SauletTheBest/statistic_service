CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Создание таблицы для участников кошелька (связь многие-ко-многим)
CREATE TABLE IF NOT EXISTS wallet_members (
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('admin', 'member')), -- 'admin' может управлять участниками, 'member' - только транзакциями
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (wallet_id, user_id) -- Составной первичный ключ
);

-- Добавление колонки wallet_id в таблицу транзакций
-- Эта колонка может быть NULL, так как у пользователя могут быть личные транзакции, не привязанные к кошельку.
ALTER TABLE transactions
ADD COLUMN IF NOT EXISTS wallet_id UUID REFERENCES wallets(id) ON DELETE SET NULL;

-- Добавляем индекс для ускорения выборок по wallet_id
CREATE INDEX IF NOT EXISTS idx_transactions_wallet_id ON transactions(wallet_id);