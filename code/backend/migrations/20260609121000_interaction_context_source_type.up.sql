ALTER TABLE interaction_context_sources
    DROP CONSTRAINT IF EXISTS interaction_context_sources_source_type_check;

ALTER TABLE interaction_context_sources
    DROP CONSTRAINT IF EXISTS chk_interaction_context_sources_source_type;

ALTER TABLE interaction_context_sources
    ADD CONSTRAINT chk_interaction_context_sources_source_type
    CHECK (source_type IN ('library', 'lib_file', 'attached_file', 'web_search'));