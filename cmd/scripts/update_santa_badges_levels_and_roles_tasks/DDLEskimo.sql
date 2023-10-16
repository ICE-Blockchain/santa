-- SPDX-License-Identifier: ice License 1.0
CREATE TABLE IF NOT EXISTS users  (
                    created_at timestamp NOT NULL,
                    updated_at timestamp NOT NULL,
                    last_mining_started_at timestamp,
                    last_mining_ended_at timestamp,
                    last_ping_cooldown_ended_at timestamp,
                    hash_code bigint not null generated always as identity,
                    kyc_step_passed smallint NOT NULL DEFAULT 0,
                    kyc_step_blocked smallint NOT NULL DEFAULT 0,
                    random_referred_by BOOLEAN NOT NULL DEFAULT FALSE,
                    client_data text,
                    hidden_profile_elements text[],
                    phone_number text NOT NULL UNIQUE,
                    email text NOT NULL UNIQUE,
                    first_name text,
                    last_name text,
                    country text NOT NULL,
                    city text NOT NULL,
                    id text primary key,
                    username text NOT NULL UNIQUE,
                    profile_picture_name text NOT NULL,
                    referred_by text NOT NULL REFERENCES users(id),
                    phone_number_hash text NOT NULL UNIQUE,
                    agenda_contact_user_ids text[],
                    kyc_steps_last_updated_at timestamp[],
                    kyc_steps_created_at timestamp[],
                    mining_blockchain_account_address text NOT NULL UNIQUE,
                    blockchain_account_address text NOT NULL UNIQUE,
                    language text NOT NULL DEFAULT 'en',
                    lookup tsvector NOT NULL)
                    WITH (FILLFACTOR = 70);


-- updater
CREATE TABLE IF NOT EXISTS updated_santa_users (
                    user_id          TEXT NOT NULL PRIMARY KEY
)
