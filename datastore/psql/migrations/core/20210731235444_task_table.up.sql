CREATE TABLE IF NOT EXISTS task (
    task_type VARCHAR (32) PRIMARY KEY,
    task_params TEXT[],
    task_queue_id VARCHAR (512)
);