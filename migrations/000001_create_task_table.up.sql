CREATE TABLE task (
    id bigserial PRIMARY KEY,
    command text,
    scheduled_at timestamp(0) with time zone,

    -- Status timestamps
    picked_at timestamp(0) with time zone,
    successful_at timestamp(0) with time zone,
    failed_at timestamp(0) with time zone,

    -- Timestamps
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);
