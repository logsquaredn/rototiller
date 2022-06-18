INSERT INTO storage (
    storage_id,
    customer_id,
    storage_name
) VALUES (
    $1,
    $2,
    $3
) RETURNING storage_id, storage_status, customer_id, storage_name, last_used, create_time;
