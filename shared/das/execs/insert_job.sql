INSERT INTO job (
    job_id,
    task_type,
    job_params
) VALUES (
    $1,
    $2,
    $3
) RETURNING job_id, task_type, job_status, job_error, start_time, end_time, job_params;
