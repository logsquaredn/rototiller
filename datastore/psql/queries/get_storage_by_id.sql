SELECT storage_id, customer_id, storage_name, last_used, create_time  FROM storage WHERE storage_id = $1;
