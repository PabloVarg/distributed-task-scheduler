CREATE TABLE task (
    id bigserial PRIMARY KEY,
    command text,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    scheduled_at timestamp(0) with time zone,
    successful_at timestamp(0) with time zone
);
