-- name: GetUser :one
SELECT * FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (email) VALUES ($1) RETURNING *;

-- name: UpdateUser :one
UPDATE users SET is_admin = $1, validated = $2, deleted = $3, updated_at = now() WHERE email = $4 RETURNING *;

-- name: DeleteUser :one
DELETE FROM users WHERE email = $1 RETURNING *;

-- name: CreatePassword :one
INSERT INTO passwords (id, salt, hash) VALUES ($1, $2, $3) RETURNING *;

-- name: GetUserPassword :one
SELECT * FROM users JOIN passwords ON users.id = passwords.id WHERE email = $1;

-- name: UpdateUserPassword :one
UPDATE passwords SET salt = $1, hash = $2 WHERE id = (SELECT id FROM users WHERE email = $3) RETURNING *;

-- name: CreateSession :one
INSERT INTO sessions (user_id, token, expires_at) VALUES ((SELECT id FROM users WHERE email = $1), $2, $3) RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions WHERE token = $1;

-- name: GetUserBySession :one
SELECT * FROM users WHERE id = (SELECT user_id FROM sessions WHERE token = $1);

-- name: UpdateSession :one
UPDATE sessions SET expires_at = $1 WHERE token = $2 RETURNING *;

-- name: ExpireSession :one
UPDATE sessions SET expires_at = now(), expired = true WHERE token = $1 RETURNING *;

-- name: CreateCluster :one
INSERT INTO clusters (name, description, password) VALUES ($1, $2, $3) RETURNING *;

-- name: GetCluster :one
SELECT * FROM clusters WHERE name = $1;

-- name: GetClusters :many
SELECT * FROM clusters ORDER BY name ASC LIMIT $1;

-- name: GetClustersByCursor :many
SELECT * FROM clusters WHERE name > $1 ORDER BY name ASC LIMIT $2;

-- name: UpdateCluster :one
UPDATE clusters SET name = $1, password = $2, description = $3, updated_at = now() WHERE name = $4 RETURNING *;

-- name: GetClusterNodes :many
SELECT * FROM nodes WHERE cluster_id = (SELECT id FROM clusters WHERE name = $1);

-- name: DeleteCluster :one
DELETE FROM clusters WHERE name = $1 RETURNING *;

-- name: CreateNode :one
INSERT INTO nodes (cluster_id, node_id, host, port) VALUES ((SELECT id FROM clusters WHERE name = $1), $2, $3, $4) RETURNING *;

-- name: GetNode :one
SELECT * FROM nodes WHERE node_id = $1;

-- name: GetNodeByHostPort :one
SELECT * FROM nodes WHERE cluster_id = (SELECT id FROM clusters WHERE name = $3) AND host = $1 AND port = $2;

-- name: GetNodeByNodeID :one
SELECT * FROM nodes WHERE cluster_id = (SELECT id FROM clusters WHERE name = $2) AND node_id = $1;

-- name: GetNodes :many
SELECT * FROM nodes WHERE cluster_id = (SELECT id FROM clusters WHERE name = $1) ORDER BY node_id ASC LIMIT $2;

-- name: GetNodesByCursor :many
SELECT * FROM nodes WHERE cluster_id = (SELECT id FROM clusters WHERE name = $1) AND node_id > $2 ORDER BY node_id ASC LIMIT $3;

-- name: UpdateNode :one
UPDATE nodes SET host = $1, port = $2, updated_at = now() WHERE node_id = $3 RETURNING *;

-- name: ConnectNode :one
UPDATE nodes SET connected = true, updated_at = now() WHERE node_id = $1 RETURNING *;

-- name: DisconnectNode :one
UPDATE nodes SET connected = false, updated_at = now() WHERE node_id = $1 RETURNING *;

-- name: SetNodeCandidate :one
UPDATE nodes SET is_candidate = $1, updated_at = now() WHERE node_id = $2 RETURNING *;

-- name: DeleteNode :one
DELETE FROM nodes WHERE cluster_id = (SELECT id FROM clusters WHERE name = $1) AND node_id = $2 RETURNING *;