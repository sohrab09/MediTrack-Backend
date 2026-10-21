CREATE TABLE
    IF NOT EXISTS banks (
        id SERIAL PRIMARY KEY,
        bank_name VARCHAR(150) NOT NULL,
        account_holder_name VARCHAR(150) NOT NULL,
        account_number VARCHAR(100) UNIQUE NOT NULL,
        account_type VARCHAR(50) NOT NULL, -- যেমন: Saving, Current ইত্যাদি
        branch_name VARCHAR(150) NOT NULL,
        routing_number VARCHAR(50) NOT NULL,
        opening_balance NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
        current_balance NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
        status INT DEFAULT 1, -- 1 = Active, 0 = Inactive
        created_at TIMESTAMP
        WITH
            TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP
        WITH
            TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );