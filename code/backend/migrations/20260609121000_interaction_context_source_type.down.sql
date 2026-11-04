ALTER TABLE interaction_context_sources
    DROP CONSTRAINT IF EXISTS chk_interaction_context_sources_source_type;

ALTER TABLE interaction_context_sources
    ADD CONSTRAINT interaction_context_sources_source_type_check
    CHECK (source_type IN ('library', 'attached_file', 'web_search'));