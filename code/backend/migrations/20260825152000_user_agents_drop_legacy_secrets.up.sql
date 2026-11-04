ALTER TABLE user_agents
	DROP COLUMN IF EXISTS secret_ciphertext,
	DROP COLUMN IF EXISTS secret_nonce,
	DROP COLUMN IF EXISTS secret_key_version;
