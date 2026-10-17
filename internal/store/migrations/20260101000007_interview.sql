CREATE TABLE IF NOT EXISTS interview_kit (
    id TEXT PRIMARY KEY,
    application_id TEXT NOT NULL,
    loop_json TEXT NOT NULL DEFAULT '[]',
    topics_json TEXT NOT NULL DEFAULT '[]',
    prompts_json TEXT NOT NULL DEFAULT '[]',
    questions_json TEXT NOT NULL DEFAULT '[]',
    created_at TEXT NOT NULL,
    FOREIGN KEY (application_id) REFERENCES applications(job_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_interview_kit_application_id ON interview_kit(application_id);

CREATE TABLE IF NOT EXISTS story_bank (
    id TEXT PRIMARY KEY,
    application_id TEXT NOT NULL,
    category TEXT NOT NULL,
    situation TEXT NOT NULL,
    task TEXT NOT NULL,
    action TEXT NOT NULL,
    result TEXT NOT NULL,
    learning TEXT NOT NULL,
    evidence_json TEXT NOT NULL DEFAULT '[]',
    created_at TEXT NOT NULL,
    FOREIGN KEY (application_id) REFERENCES applications(job_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_story_bank_application_id ON story_bank(application_id);

CREATE TABLE IF NOT EXISTS mock_session (
    id TEXT PRIMARY KEY,
    application_id TEXT NOT NULL,
    type TEXT NOT NULL,
    prompt TEXT NOT NULL DEFAULT '',
    started_at TEXT NOT NULL,
    finished_at TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    FOREIGN KEY (application_id) REFERENCES applications(job_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_mock_session_application_id ON mock_session(application_id);

CREATE TABLE IF NOT EXISTS mock_step (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    idx INTEGER NOT NULL,
    name TEXT NOT NULL,
    timebox_s INTEGER NOT NULL,
    output TEXT NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '',
    weak INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (session_id) REFERENCES mock_session(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_mock_step_session_id ON mock_step(session_id);

CREATE TABLE IF NOT EXISTS rubric_score (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    dimension TEXT NOT NULL,
    score INTEGER NOT NULL,
    comment TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (session_id) REFERENCES mock_session(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_rubric_score_session_id ON rubric_score(session_id);

CREATE TABLE IF NOT EXISTS topic_note (
    slug TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    tier TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS drill (
    id TEXT PRIMARY KEY,
    topic_slug TEXT NOT NULL,
    prompt TEXT NOT NULL,
    minutes INTEGER NOT NULL DEFAULT 15,
    FOREIGN KEY (topic_slug) REFERENCES topic_note(slug) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_drill_topic_slug ON drill(topic_slug);

-- Mindset (tier mindset)
INSERT OR IGNORE INTO topic_note (slug, title, body, tier) VALUES
('mindset-client-meeting', 'Interview is a client meeting', 'Treat the loop as scoping work with a future teammate, not passing a test. Ask what success looks like for the role.', 'mindset'),
('mindset-ask-questions', 'Do not be afraid to ask', 'Clarifying questions show judgment. Silence while guessing hurts you more than one good question.', 'mindset'),
('mindset-think-aloud', 'Think out loud', 'Share reasoning as you go. The interviewer cannot grade a black box.', 'mindset'),
('mindset-no-prior-use', 'Prior production use is optional', 'You can design with technologies you have studied but not shipped. Ground claims in tradeoffs, not resume keywords.', 'mindset'),
('mindset-defend-decisions', 'Defend your decisions', 'State assumptions, compare options, and know what would change your mind.', 'mindset'),
('mindset-not-perfect', 'Do not wait for perfect', 'Ship a workable design, then iterate. Perfectionism burns the clock.', 'mindset'),
('mindset-can-work-together', 'Passing means they can work with you', 'Signals: clear communication, collaboration, and honest limits beat flawless but silent answers.', 'mindset');

-- Reference topics (tier reference)
INSERT OR IGNORE INTO topic_note (slug, title, body, tier) VALUES
('scalability', 'Scalability', 'Grow capacity by splitting work, caching, and async boundaries. Watch hot keys and coordination cost.', 'reference'),
('latency-vs-throughput', 'Latency vs throughput', 'Latency is time per request; throughput is work per second. Optimizing one can hurt the other.', 'reference'),
('cap-theorem', 'CAP theorem', 'Under partition, choose consistency or availability. Most systems pick partial availability with tunable consistency.', 'reference'),
('availability', 'Availability', 'Measure uptime and graceful degradation. Redundancy, health checks, and failover paths matter.', 'reference'),
('load-balancers', 'Load balancers', 'Distribute traffic with L4/L7 balancers. Sticky sessions, health probes, and TLS termination are common concerns.', 'reference'),
('caching', 'Caching', 'Cache at browser, CDN, app, and database layers. Define TTL, invalidation, and stampede protection.', 'reference'),
('databases', 'Databases', 'Pick SQL vs NoSQL by access patterns. Indexing, replication, and backup/restore are interview staples.', 'reference'),
('message-queues', 'Message queues', 'Decouple producers and consumers. Consider ordering, duplicates, poison messages, and backpressure.', 'reference'),
('dns-cdn', 'DNS and CDN', 'DNS routes users; CDNs cache static assets at the edge. Both affect latency and resilience.', 'reference'),
('rate-limiting', 'Rate limiting', 'Protect APIs with token bucket or leaky bucket limits. Return clear errors and idempotent retries.', 'reference'),
('consistent-hashing', 'Consistent hashing', 'Minimize remapping when nodes join or leave. Used in caches and distributed stores.', 'reference'),
('sharding', 'Sharding', 'Split data by shard key. Rebalancing, cross-shard queries, and hotspots need explicit plans.', 'reference'),
('event-driven', 'Event-driven architecture', 'Publish domain events for loose coupling. Mind ordering, schemas, and consumer lag.', 'reference');

INSERT OR IGNORE INTO drill (id, topic_slug, prompt, minutes) VALUES
('drill-scalability', 'scalability', 'Sketch how you would scale a read-heavy API from 1k to 100k RPS without a full redesign.', 15),
('drill-latency-vs-throughput', 'latency-vs-throughput', 'Given p99 latency SLO and a throughput target, list two changes that help each and one conflict.', 15),
('drill-cap-theorem', 'cap-theorem', 'During a network partition, defend choosing availability for a social feed vs consistency for payments.', 15),
('drill-availability', 'availability', 'Design health checks and failover for a stateless service behind a load balancer.', 15),
('drill-load-balancers', 'load-balancers', 'Compare round-robin, least connections, and consistent-hash routing for a cache tier.', 15),
('drill-caching', 'caching', 'Pick cache keys and TTL for a product catalog with frequent price updates.', 15),
('drill-databases', 'databases', 'Choose primary/replica layout for a write-light, read-heavy dashboard.', 15),
('drill-message-queues', 'message-queues', 'Handle duplicate delivery and poison messages in an order-processing queue.', 15),
('drill-dns-cdn', 'dns-cdn', 'Explain request path from browser to origin with CDN cache hit and miss.', 15),
('drill-rate-limiting', 'rate-limiting', 'Design per-user and global rate limits for a public API.', 15),
('drill-consistent-hashing', 'consistent-hashing', 'Add a node to a consistent-hash ring and estimate how many keys move.', 15),
('drill-sharding', 'sharding', 'Pick a shard key for multi-tenant SaaS and explain one cross-shard query strategy.', 15),
('drill-event-driven', 'event-driven', 'Draw event flow for signup: account created, welcome email, analytics.', 15);
