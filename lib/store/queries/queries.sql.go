// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: queries.sql

package queries

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const connectNode = `-- name: ConnectNode :one
UPDATE nodes SET connected = true, updated_at = now() WHERE node_id = $1 RETURNING id, cluster_id, node_id, host, port, connected, is_candidate, created_at, updated_at
`

func (q *Queries) ConnectNode(ctx context.Context, nodeID string) (Node, error) {
	row := q.db.QueryRow(ctx, connectNode, nodeID)
	var i Node
	err := row.Scan(
		&i.ID,
		&i.ClusterID,
		&i.NodeID,
		&i.Host,
		&i.Port,
		&i.Connected,
		&i.IsCandidate,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createCluster = `-- name: CreateCluster :one
INSERT INTO clusters (name, description, password) VALUES ($1, $2, $3) RETURNING id, name, description, password, created_at, updated_at
`

type CreateClusterParams struct {
	Name        string
	Description pgtype.Text
	Password    string
}

func (q *Queries) CreateCluster(ctx context.Context, arg CreateClusterParams) (Cluster, error) {
	row := q.db.QueryRow(ctx, createCluster, arg.Name, arg.Description, arg.Password)
	var i Cluster
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Password,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createNode = `-- name: CreateNode :one
INSERT INTO nodes (cluster_id, node_id, host, port) VALUES ((SELECT id FROM clusters WHERE name = $1), $2, $3, $4) RETURNING id, cluster_id, node_id, host, port, connected, is_candidate, created_at, updated_at
`

type CreateNodeParams struct {
	Name   string
	NodeID string
	Host   string
	Port   int32
}

func (q *Queries) CreateNode(ctx context.Context, arg CreateNodeParams) (Node, error) {
	row := q.db.QueryRow(ctx, createNode,
		arg.Name,
		arg.NodeID,
		arg.Host,
		arg.Port,
	)
	var i Node
	err := row.Scan(
		&i.ID,
		&i.ClusterID,
		&i.NodeID,
		&i.Host,
		&i.Port,
		&i.Connected,
		&i.IsCandidate,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createPassword = `-- name: CreatePassword :one
INSERT INTO passwords (id, salt, hash) VALUES ($1, $2, $3) RETURNING id, salt, hash
`

type CreatePasswordParams struct {
	ID   int32
	Salt string
	Hash string
}

func (q *Queries) CreatePassword(ctx context.Context, arg CreatePasswordParams) (Password, error) {
	row := q.db.QueryRow(ctx, createPassword, arg.ID, arg.Salt, arg.Hash)
	var i Password
	err := row.Scan(&i.ID, &i.Salt, &i.Hash)
	return i, err
}

const createSession = `-- name: CreateSession :one
INSERT INTO sessions (user_id, token, expires_at) VALUES ((SELECT id FROM users WHERE email = $1), $2, $3) RETURNING id, user_id, token, created_at, updated_at, expired, expires_at
`

type CreateSessionParams struct {
	Email     string
	Token     string
	ExpiresAt pgtype.Timestamp
}

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error) {
	row := q.db.QueryRow(ctx, createSession, arg.Email, arg.Token, arg.ExpiresAt)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Token,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Expired,
		&i.ExpiresAt,
	)
	return i, err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (email) VALUES ($1) RETURNING id, email, is_admin, validated, deleted, created_at, updated_at
`

func (q *Queries) CreateUser(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, createUser, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.IsAdmin,
		&i.Validated,
		&i.Deleted,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteCluster = `-- name: DeleteCluster :one
DELETE FROM clusters WHERE name = $1 RETURNING id, name, description, password, created_at, updated_at
`

func (q *Queries) DeleteCluster(ctx context.Context, name string) (Cluster, error) {
	row := q.db.QueryRow(ctx, deleteCluster, name)
	var i Cluster
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Password,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteNode = `-- name: DeleteNode :one
DELETE FROM nodes WHERE cluster_id = (SELECT id FROM clusters WHERE name = $1) AND node_id = $2 RETURNING id, cluster_id, node_id, host, port, connected, is_candidate, created_at, updated_at
`

type DeleteNodeParams struct {
	Name   string
	NodeID string
}

func (q *Queries) DeleteNode(ctx context.Context, arg DeleteNodeParams) (Node, error) {
	row := q.db.QueryRow(ctx, deleteNode, arg.Name, arg.NodeID)
	var i Node
	err := row.Scan(
		&i.ID,
		&i.ClusterID,
		&i.NodeID,
		&i.Host,
		&i.Port,
		&i.Connected,
		&i.IsCandidate,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteUser = `-- name: DeleteUser :one
DELETE FROM users WHERE email = $1 RETURNING id, email, is_admin, validated, deleted, created_at, updated_at
`

func (q *Queries) DeleteUser(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, deleteUser, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.IsAdmin,
		&i.Validated,
		&i.Deleted,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const disconnectNode = `-- name: DisconnectNode :one
UPDATE nodes SET connected = false, updated_at = now() WHERE node_id = $1 RETURNING id, cluster_id, node_id, host, port, connected, is_candidate, created_at, updated_at
`

func (q *Queries) DisconnectNode(ctx context.Context, nodeID string) (Node, error) {
	row := q.db.QueryRow(ctx, disconnectNode, nodeID)
	var i Node
	err := row.Scan(
		&i.ID,
		&i.ClusterID,
		&i.NodeID,
		&i.Host,
		&i.Port,
		&i.Connected,
		&i.IsCandidate,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const expireSession = `-- name: ExpireSession :one
UPDATE sessions SET expires_at = now(), expired = true WHERE token = $1 RETURNING id, user_id, token, created_at, updated_at, expired, expires_at
`

func (q *Queries) ExpireSession(ctx context.Context, token string) (Session, error) {
	row := q.db.QueryRow(ctx, expireSession, token)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Token,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Expired,
		&i.ExpiresAt,
	)
	return i, err
}

const getCluster = `-- name: GetCluster :one
SELECT id, name, description, password, created_at, updated_at FROM clusters WHERE name = $1
`

func (q *Queries) GetCluster(ctx context.Context, name string) (Cluster, error) {
	row := q.db.QueryRow(ctx, getCluster, name)
	var i Cluster
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Password,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getClusterNodes = `-- name: GetClusterNodes :many
SELECT id, cluster_id, node_id, host, port, connected, is_candidate, created_at, updated_at FROM nodes WHERE cluster_id = (SELECT id FROM clusters WHERE name = $1)
`

func (q *Queries) GetClusterNodes(ctx context.Context, name string) ([]Node, error) {
	rows, err := q.db.Query(ctx, getClusterNodes, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Node
	for rows.Next() {
		var i Node
		if err := rows.Scan(
			&i.ID,
			&i.ClusterID,
			&i.NodeID,
			&i.Host,
			&i.Port,
			&i.Connected,
			&i.IsCandidate,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getClusters = `-- name: GetClusters :many
SELECT id, name, description, password, created_at, updated_at FROM clusters ORDER BY name ASC LIMIT $1
`

func (q *Queries) GetClusters(ctx context.Context, limit int32) ([]Cluster, error) {
	rows, err := q.db.Query(ctx, getClusters, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Cluster
	for rows.Next() {
		var i Cluster
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Password,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getClustersByCursor = `-- name: GetClustersByCursor :many
SELECT id, name, description, password, created_at, updated_at FROM clusters WHERE name > $1 ORDER BY name ASC LIMIT $2
`

type GetClustersByCursorParams struct {
	Name  string
	Limit int32
}

func (q *Queries) GetClustersByCursor(ctx context.Context, arg GetClustersByCursorParams) ([]Cluster, error) {
	rows, err := q.db.Query(ctx, getClustersByCursor, arg.Name, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Cluster
	for rows.Next() {
		var i Cluster
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Password,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getNode = `-- name: GetNode :one
SELECT id, cluster_id, node_id, host, port, connected, is_candidate, created_at, updated_at FROM nodes WHERE node_id = $1
`

func (q *Queries) GetNode(ctx context.Context, nodeID string) (Node, error) {
	row := q.db.QueryRow(ctx, getNode, nodeID)
	var i Node
	err := row.Scan(
		&i.ID,
		&i.ClusterID,
		&i.NodeID,
		&i.Host,
		&i.Port,
		&i.Connected,
		&i.IsCandidate,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getNodeByHostPort = `-- name: GetNodeByHostPort :one
SELECT id, cluster_id, node_id, host, port, connected, is_candidate, created_at, updated_at FROM nodes WHERE cluster_id = (SELECT id FROM clusters WHERE name = $3) AND host = $1 AND port = $2
`

type GetNodeByHostPortParams struct {
	Host string
	Port int32
	Name string
}

func (q *Queries) GetNodeByHostPort(ctx context.Context, arg GetNodeByHostPortParams) (Node, error) {
	row := q.db.QueryRow(ctx, getNodeByHostPort, arg.Host, arg.Port, arg.Name)
	var i Node
	err := row.Scan(
		&i.ID,
		&i.ClusterID,
		&i.NodeID,
		&i.Host,
		&i.Port,
		&i.Connected,
		&i.IsCandidate,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getNodeByNodeID = `-- name: GetNodeByNodeID :one
SELECT id, cluster_id, node_id, host, port, connected, is_candidate, created_at, updated_at FROM nodes WHERE cluster_id = (SELECT id FROM clusters WHERE name = $2) AND node_id = $1
`

type GetNodeByNodeIDParams struct {
	NodeID string
	Name   string
}

func (q *Queries) GetNodeByNodeID(ctx context.Context, arg GetNodeByNodeIDParams) (Node, error) {
	row := q.db.QueryRow(ctx, getNodeByNodeID, arg.NodeID, arg.Name)
	var i Node
	err := row.Scan(
		&i.ID,
		&i.ClusterID,
		&i.NodeID,
		&i.Host,
		&i.Port,
		&i.Connected,
		&i.IsCandidate,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getNodes = `-- name: GetNodes :many
SELECT id, cluster_id, node_id, host, port, connected, is_candidate, created_at, updated_at FROM nodes WHERE cluster_id = (SELECT id FROM clusters WHERE name = $1) ORDER BY node_id ASC LIMIT $2
`

type GetNodesParams struct {
	Name  string
	Limit int32
}

func (q *Queries) GetNodes(ctx context.Context, arg GetNodesParams) ([]Node, error) {
	rows, err := q.db.Query(ctx, getNodes, arg.Name, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Node
	for rows.Next() {
		var i Node
		if err := rows.Scan(
			&i.ID,
			&i.ClusterID,
			&i.NodeID,
			&i.Host,
			&i.Port,
			&i.Connected,
			&i.IsCandidate,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getNodesByCursor = `-- name: GetNodesByCursor :many
SELECT id, cluster_id, node_id, host, port, connected, is_candidate, created_at, updated_at FROM nodes WHERE cluster_id = (SELECT id FROM clusters WHERE name = $1) AND node_id > $2 ORDER BY node_id ASC LIMIT $3
`

type GetNodesByCursorParams struct {
	Name   string
	NodeID string
	Limit  int32
}

func (q *Queries) GetNodesByCursor(ctx context.Context, arg GetNodesByCursorParams) ([]Node, error) {
	rows, err := q.db.Query(ctx, getNodesByCursor, arg.Name, arg.NodeID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Node
	for rows.Next() {
		var i Node
		if err := rows.Scan(
			&i.ID,
			&i.ClusterID,
			&i.NodeID,
			&i.Host,
			&i.Port,
			&i.Connected,
			&i.IsCandidate,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getSession = `-- name: GetSession :one
SELECT id, user_id, token, created_at, updated_at, expired, expires_at FROM sessions WHERE token = $1
`

func (q *Queries) GetSession(ctx context.Context, token string) (Session, error) {
	row := q.db.QueryRow(ctx, getSession, token)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Token,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Expired,
		&i.ExpiresAt,
	)
	return i, err
}

const getUser = `-- name: GetUser :one
SELECT id, email, is_admin, validated, deleted, created_at, updated_at FROM users WHERE email = $1
`

func (q *Queries) GetUser(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, getUser, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.IsAdmin,
		&i.Validated,
		&i.Deleted,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserBySession = `-- name: GetUserBySession :one
SELECT id, email, is_admin, validated, deleted, created_at, updated_at FROM users WHERE id = (SELECT user_id FROM sessions WHERE token = $1)
`

func (q *Queries) GetUserBySession(ctx context.Context, token string) (User, error) {
	row := q.db.QueryRow(ctx, getUserBySession, token)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.IsAdmin,
		&i.Validated,
		&i.Deleted,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserPassword = `-- name: GetUserPassword :one
SELECT users.id, email, is_admin, validated, deleted, created_at, updated_at, passwords.id, salt, hash FROM users JOIN passwords ON users.id = passwords.id WHERE email = $1
`

type GetUserPasswordRow struct {
	ID        int32
	Email     string
	IsAdmin   bool
	Validated bool
	Deleted   bool
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
	ID_2      int32
	Salt      string
	Hash      string
}

func (q *Queries) GetUserPassword(ctx context.Context, email string) (GetUserPasswordRow, error) {
	row := q.db.QueryRow(ctx, getUserPassword, email)
	var i GetUserPasswordRow
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.IsAdmin,
		&i.Validated,
		&i.Deleted,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ID_2,
		&i.Salt,
		&i.Hash,
	)
	return i, err
}

const setNodeCandidate = `-- name: SetNodeCandidate :one
UPDATE nodes SET is_candidate = $1, updated_at = now() WHERE node_id = $2 RETURNING id, cluster_id, node_id, host, port, connected, is_candidate, created_at, updated_at
`

type SetNodeCandidateParams struct {
	IsCandidate bool
	NodeID      string
}

func (q *Queries) SetNodeCandidate(ctx context.Context, arg SetNodeCandidateParams) (Node, error) {
	row := q.db.QueryRow(ctx, setNodeCandidate, arg.IsCandidate, arg.NodeID)
	var i Node
	err := row.Scan(
		&i.ID,
		&i.ClusterID,
		&i.NodeID,
		&i.Host,
		&i.Port,
		&i.Connected,
		&i.IsCandidate,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateCluster = `-- name: UpdateCluster :one
UPDATE clusters SET name = $1, password = $2, description = $3, updated_at = now() WHERE name = $4 RETURNING id, name, description, password, created_at, updated_at
`

type UpdateClusterParams struct {
	Name        string
	Password    string
	Description pgtype.Text
	Name_2      string
}

func (q *Queries) UpdateCluster(ctx context.Context, arg UpdateClusterParams) (Cluster, error) {
	row := q.db.QueryRow(ctx, updateCluster,
		arg.Name,
		arg.Password,
		arg.Description,
		arg.Name_2,
	)
	var i Cluster
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Password,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateNode = `-- name: UpdateNode :one
UPDATE nodes SET host = $1, port = $2, updated_at = now() WHERE node_id = $3 RETURNING id, cluster_id, node_id, host, port, connected, is_candidate, created_at, updated_at
`

type UpdateNodeParams struct {
	Host   string
	Port   int32
	NodeID string
}

func (q *Queries) UpdateNode(ctx context.Context, arg UpdateNodeParams) (Node, error) {
	row := q.db.QueryRow(ctx, updateNode, arg.Host, arg.Port, arg.NodeID)
	var i Node
	err := row.Scan(
		&i.ID,
		&i.ClusterID,
		&i.NodeID,
		&i.Host,
		&i.Port,
		&i.Connected,
		&i.IsCandidate,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateSession = `-- name: UpdateSession :one
UPDATE sessions SET expires_at = $1 WHERE token = $2 RETURNING id, user_id, token, created_at, updated_at, expired, expires_at
`

type UpdateSessionParams struct {
	ExpiresAt pgtype.Timestamp
	Token     string
}

func (q *Queries) UpdateSession(ctx context.Context, arg UpdateSessionParams) (Session, error) {
	row := q.db.QueryRow(ctx, updateSession, arg.ExpiresAt, arg.Token)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Token,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Expired,
		&i.ExpiresAt,
	)
	return i, err
}

const updateUser = `-- name: UpdateUser :one
UPDATE users SET is_admin = $1, validated = $2, deleted = $3, updated_at = now() WHERE email = $4 RETURNING id, email, is_admin, validated, deleted, created_at, updated_at
`

type UpdateUserParams struct {
	IsAdmin   bool
	Validated bool
	Deleted   bool
	Email     string
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUser,
		arg.IsAdmin,
		arg.Validated,
		arg.Deleted,
		arg.Email,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.IsAdmin,
		&i.Validated,
		&i.Deleted,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateUserPassword = `-- name: UpdateUserPassword :one
UPDATE passwords SET salt = $1, hash = $2 WHERE id = (SELECT id FROM users WHERE email = $3) RETURNING id, salt, hash
`

type UpdateUserPasswordParams struct {
	Salt  string
	Hash  string
	Email string
}

func (q *Queries) UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) (Password, error) {
	row := q.db.QueryRow(ctx, updateUserPassword, arg.Salt, arg.Hash, arg.Email)
	var i Password
	err := row.Scan(&i.ID, &i.Salt, &i.Hash)
	return i, err
}
