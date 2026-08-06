-- Roll back community. Refuses while any post or report exists: what
-- was said and how the school responded is moderation history.

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM community_posts)
       OR EXISTS (SELECT 1 FROM community_reports) THEN
        RAISE EXCEPTION 'cannot roll back community while its history exists';
    END IF;
END
$$;

DROP TABLE IF EXISTS community_blocks;
DROP TRIGGER IF EXISTS community_reports_no_delete ON community_reports;
DROP TABLE IF EXISTS community_reports;
DROP TRIGGER IF EXISTS community_comments_no_delete ON community_comments;
DROP TABLE IF EXISTS community_comments;
DROP TRIGGER IF EXISTS community_posts_no_delete ON community_posts;
DROP FUNCTION IF EXISTS reject_community_delete();
DROP TABLE IF EXISTS community_posts;
