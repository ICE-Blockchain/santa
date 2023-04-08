-- SPDX-License-Identifier: ice License 1.0
--************************************************************************************************************************************
-- task_progress
CREATE TABLE IF NOT EXISTS task_progress (
                        completed_tasks             TEXT,
                        pseudo_completed_tasks      TEXT,
                        user_id                     TEXT NOT NULL PRIMARY KEY,
                        twitter_user_handle         TEXT,
                        telegram_user_handle        TEXT,
                        friends_invited             BIGINT NOT NULL DEFAULT 0,
                        username_set                BOOLEAN,
                        profile_picture_set         BOOLEAN,
                        mining_started              BOOLEAN
                    );
--************************************************************************************************************************************
-- referrals
CREATE TABLE IF NOT EXISTS referrals (
                     user_id        TEXT NOT NULL PRIMARY KEY,
                     referred_by    TEXT NOT NULL
                 );
CREATE INDEX IF NOT EXISTS referrals_referred_by_ix ON referrals (referred_by);