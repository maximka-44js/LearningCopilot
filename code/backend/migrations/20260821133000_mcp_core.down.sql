DROP INDEX IF EXISTS idx_mcp_webhook_nonces_expires;
DROP INDEX IF EXISTS idx_mcp_events_audit_event_type_created;
DROP INDEX IF EXISTS idx_mcp_events_audit_connection_created;
DROP INDEX IF EXISTS idx_mcp_dispatch_commands_connection_created;
DROP INDEX IF EXISTS idx_mcp_dispatch_commands_state_retry;
DROP INDEX IF EXISTS idx_mcp_runtime_state_lifecycle;
DROP INDEX IF EXISTS ux_mcp_gateway_bindings_gateway_connection;
DROP INDEX IF EXISTS idx_mcp_connections_owner_mode_enabled;
DROP INDEX IF EXISTS ux_mcp_connections_owner_name;

DROP TABLE IF EXISTS mcp_webhook_nonces;
DROP TABLE IF EXISTS mcp_events_audit;
DROP TABLE IF EXISTS mcp_dispatch_commands;
DROP TABLE IF EXISTS mcp_runtime_state;
DROP TABLE IF EXISTS mcp_gateway_bindings;
DROP TABLE IF EXISTS mcp_command_configs;
DROP TABLE IF EXISTS mcp_url_configs;
DROP TABLE IF EXISTS mcp_connections;
