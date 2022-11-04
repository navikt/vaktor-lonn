-- name: ListBeredskapsvakter :many
SELECT *
FROM beredskapsvakt
ORDER BY ident;

-- name: GetPlan :one
SELECT *
FROM beredskapsvakt
WHERE id = $1;

-- name: CreatePlan :exec
INSERT INTO beredskapsvakt
    ("id", "ident", "plan", "period_begin", "period_end")
VALUES ($1, $2, $3, $4, $5);

-- name: UpdatePlan :exec
UPDATE beredskapsvakt
SET "ident"        = @ident,
    "plan"         = @plan,
    "period_begin" = @period_begin,
    "period_end"   = @period_end
WHERE id = @id;

-- name: DeletePlan :exec
DELETE
FROM beredskapsvakt
WHERE id = $1;