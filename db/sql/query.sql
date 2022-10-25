SELECT id, driver, kind, start_time, end_time, distance
FROM rides
WHERE
    start_time >= $1
    AND
    end_time < $2
;
