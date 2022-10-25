SELECT id, driver, kind, start_time, end_time, distance
FROM rides
WHERE id = $1
;
