-- name: ListBeredskapsvakter :many
SELECT *
FROM beredskapsvakt
ORDER BY ident;

-- name: CreatePlan :exec
INSERT INTO beredskapsvakt
    ("id", "ident", "plan", "period_begin", "period_end")
VALUES ($1, $2, $3, $4, $5);

-- name: DeletePlan :exec
DELETE
FROM beredskapsvakt
WHERE id = $1;