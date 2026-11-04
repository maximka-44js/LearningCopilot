CREATE TABLE IF NOT EXISTS chat_mcp_connections (
	owner_id UUID NOT NULL,
	chat_id UUID NOT NULL,
	connection_id UUID NOT NULL,
	attached_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT fk_chat_mcp_connections_chat
		FOREIGN KEY (chat_id)
		REFERENCES chats(id)
		ON DELETE CASCADE,

	CONSTRAINT ux_chat_mcp_connections
		UNIQUE (owner_id, chat_id, connection_id)
);

CREATE INDEX IF NOT EXISTS idx_chat_mcp_connections_chat
ON chat_mcp_connections (chat_id, attached_at DESC);

CREATE INDEX IF NOT EXISTS idx_chat_mcp_connections_owner
ON chat_mcp_connections (owner_id, attached_at DESC);
