// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: queries.sql

package queries

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createCluster = `-- name: CreateCluster :one
INSERT INTO clusters (name, password) VALUES ($1, $2) RETURNING id, name, password, created_at, updated_at
`

type CreateClusterParams struct {
	Name     string
	Password string
}

func (q *Queries) CreateCluster(ctx context.Context, arg CreateClusterParams) (Cluster, error) {
	row := q.db.QueryRow(ctx, createCluster, arg.Name, arg.Password)
	var i Cluster
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Password,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createNode = `-- name: CreateNode :one
INSERT INTO nodes (cluster_id, node_id, host, port) VALUES ((SELECT id FROM clusters WHERE name = $1), $2, $3, $4) RETURNING id, cluster_id, node_id, host, port, created_at, updated_at
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
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createPassword = `-- name: CreatePassword :one
INSERT INTO passwords (salt, hash) VALUES ($1, $2) RETURNING id, salt, hash
`

type CreatePasswordParams struct {
	Salt string
	Hash string
}

func (q *Queries) CreatePassword(ctx context.Context, arg CreatePasswordParams) (Password, error) {
	row := q.db.QueryRow(ctx, createPassword, arg.Salt, arg.Hash)
	var i Password
	err := row.Scan(&i.ID, &i.Salt, &i.Hash)
	return i, err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (email) VALUES ($1) RETURNING id, email, created_at, updated_at
`

func (q *Queries) CreateUser(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, createUser, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getCluster = `-- name: GetCluster :one
SELECT id, name, password, created_at, updated_at FROM clusters WHERE name = $1
`

func (q *Queries) GetCluster(ctx context.Context, name string) (Cluster, error) {
	row := q.db.QueryRow(ctx, getCluster, name)
	var i Cluster
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Password,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getClusterNodes = `-- name: GetClusterNodes :many
SELECT id, cluster_id, node_id, host, port, created_at, updated_at FROM nodes WHERE cluster_id = (SELECT id FROM clusters WHERE name = $1)
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
SELECT id, cluster_id, node_id, host, port, created_at, updated_at FROM nodes WHERE node_id = $1
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
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUser = `-- name: GetUser :one
SELECT id, email, created_at, updated_at FROM users WHERE email = $1
`

func (q *Queries) GetUser(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, getUser, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserPassword = `-- name: GetUserPassword :one
SELECT users.id, email, created_at, updated_at, passwords.id, salt, hash FROM users JOIN passwords ON users.id = passwords.id WHERE email = $1
`

type GetUserPasswordRow struct {
	ID        int32
	Email     string
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
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ID_2,
		&i.Salt,
		&i.Hash,
	)
	return i, err
}

const updateCluster = `-- name: UpdateCluster :one
UPDATE clusters SET name = $1, password = $2, updated_at = now() WHERE name = $3 RETURNING id, name, password, created_at, updated_at
`

type UpdateClusterParams struct {
	Name     string
	Password string
	Name_2   string
}

func (q *Queries) UpdateCluster(ctx context.Context, arg UpdateClusterParams) (Cluster, error) {
	row := q.db.QueryRow(ctx, updateCluster, arg.Name, arg.Password, arg.Name_2)
	var i Cluster
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Password,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateNode = `-- name: UpdateNode :one
UPDATE nodes SET host = $1, port = $2, updated_at = now() WHERE node_id = $3 RETURNING id, cluster_id, node_id, host, port, created_at, updated_at
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
