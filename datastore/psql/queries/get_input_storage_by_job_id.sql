SELECT s.storage_id, s.customer_id, s.storage_name, s.last_used FROM storage s INNER JOIN job j ON s.storage_id = j.input_id WHERE j.job_id = $1;
