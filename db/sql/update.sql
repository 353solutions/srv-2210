UPDATE rides
SET
    driver = $2,
    kind = $3,
    start_time = $4,
    end_time = $5,
    distance = $6
WHERE
    id = $1
;
