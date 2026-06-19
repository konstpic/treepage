-- Multilingual full-text search: index english + russian in search_vector.
CREATE OR REPLACE FUNCTION documents_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('russian', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.content, '')), 'B') ||
        setweight(to_tsvector('russian', COALESCE(NEW.content, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(array_to_string(NEW.tags, ' '), '')), 'C') ||
        setweight(to_tsvector('russian', COALESCE(array_to_string(NEW.tags, ' '), '')), 'C') ||
        setweight(to_tsvector('english', COALESCE(NEW.author_name, '')), 'D') ||
        setweight(to_tsvector('russian', COALESCE(NEW.author_name, '')), 'D');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Refresh existing documents so search_vector includes Russian tokens.
UPDATE documents SET content = content WHERE content IS NOT NULL;
