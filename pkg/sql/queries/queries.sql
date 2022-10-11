-- name: ListBeredskapsvakter :many
SELECT *
FROM beredskapsvakt
ORDER BY ident;

-- name: CreatePlan :exec
INSERT INTO beredskapsvakt
    ("id", "ident", "plan")
VALUES ($1, $2, $3);

-- name: DeletePlan :exec
DELETE
FROM beredskapsvakt
WHERE id = $1;