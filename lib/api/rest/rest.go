package rest

import (
	"context"
	"crypto/sha3"
	"encoding/json"
	"fmt"
	"hash"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/blake2b"

	"github.com/snowmerak/keycl/lib/store"
	"github.com/snowmerak/keycl/lib/store/queries"
	"github.com/snowmerak/keycl/lib/util/password"
)

const (
	CookieNameToken = "k-token"
)

func defaultHash1() hash.Hash {
	return sha3.New512()
}

func defaultHash2() hash.Hash {
	h, _ := blake2b.New384(nil)
	return h
}

type API struct {
	store   *store.Store
	cliName string
}

func New(store *store.Store, cliName string) *API {
	return &API{
		store:   store,
		cliName: cliName,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool `json:"success"`
}

// Login creates a new session
// POST /api/session
func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	defer r.Body.Close()

	request := &LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token := ""
	expires := time.Now().Add(24 * time.Hour)
	responseStatus := http.StatusOK
	if err := a.store.Visit(ctx, func(ctx context.Context, q *queries.Queries) error {
		resp, err := q.GetUserPassword(ctx, request.Email)
		if err != nil {
			responseStatus = http.StatusBadRequest
			return fmt.Errorf("q.GetUserPassword: %w", err)
		}

		if resp.Deleted {
			responseStatus = http.StatusUnauthorized
			return nil
		}

		hashed := password.HashPassword(defaultHash1(), defaultHash2(), resp.Salt, request.Password)
		if hashed != resp.Hash {
			responseStatus = http.StatusUnauthorized
			return nil
		}

		token = password.HashToken(defaultHash2(), resp.Email, hashed)
		if token == "" {
			responseStatus = http.StatusInternalServerError
			return fmt.Errorf("password.HashToken: %w", err)
		}

		ns, err := q.CreateSession(ctx, queries.CreateSessionParams{
			Email: resp.Email,
			Token: token,
			ExpiresAt: pgtype.Timestamp{
				Time:  expires,
				Valid: true,
			},
		})
		if err != nil {
			responseStatus = http.StatusInternalServerError
			return fmt.Errorf("q.CreateSession: %w", err)
		}

		token = ns.Token

		return nil
	}); err != nil {
		log.Error().Err(err).Any("request", request).Msg("Failed to login")
		http.Error(w, "invalid email or password", responseStatus)
		return
	}

	ck := http.Cookie{
		Name:     CookieNameToken,
		Value:    token,
		HttpOnly: true,
		Expires:  expires,
		MaxAge:   int(expires.Sub(time.Now()).Seconds()),
		SameSite: http.SameSiteStrictMode,
	}
	switch isDev {
	case true:
		ck.Secure = false
	default:
		ck.Secure = true
	}
	w.Header().Add("Set-Cookie", ck.String())

	data, _ := json.Marshal(LoginResponse{Success: true})
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	w.WriteHeader(http.StatusCreated)
}

// Logout deletes the session
// DELETE /api/session
func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	defer r.Body.Close()

	ck, err := r.Cookie(CookieNameToken)
	if err != nil {
		http.Error(w, "no token", http.StatusBadRequest)
		return
	}

	if err := a.store.Visit(ctx, func(ctx context.Context, q *queries.Queries) error {
		if _, err := q.ExpireSession(ctx, ck.Value); err != nil {
			return fmt.Errorf("q.ExpireSession: %w", err)
		}

		return nil
	}); err != nil {
		log.Error().Err(err).Msg("Failed to logout")
		http.Error(w, "failed to logout", http.StatusInternalServerError)
		return
	}

	ck.Expires = time.Now().Add(-time.Hour)
	ck.MaxAge = -1
	w.Header().Add("Set-Cookie", ck.String())

	w.WriteHeader(http.StatusOK)
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateUser creates a new user
// POST /api/user
func (a *API) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	defer r.Body.Close()

	request := &CreateUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	responseStatus := http.StatusOK
	if err := a.store.VisitTx(ctx, func(ctx context.Context, q *queries.Queries) error {
		createdUser, err := q.CreateUser(ctx, request.Email)
		if err != nil {
			responseStatus = http.StatusBadRequest
			return fmt.Errorf("q.CreateUser: %w", err)
		}

		salt := password.GenerateSalt()
		hashed := password.HashPassword(defaultHash1(), defaultHash2(), salt, request.Password)
		if _, err := q.CreatePassword(ctx, queries.CreatePasswordParams{
			ID:   createdUser.ID,
			Hash: hashed,
			Salt: salt,
		}); err != nil {
			responseStatus = http.StatusInternalServerError
			return fmt.Errorf("q.CreatePassword: %w", err)
		}

		return nil
	}); err != nil {
		log.Error().Err(err).Any("request", request).Msg("Failed to create user")
		http.Error(w, "failed to create user", responseStatus)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// DeleteUser deletes the user
// DELETE /api/user?email=email
func (a *API) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	defer r.Body.Close()

	ck, err := r.Cookie(CookieNameToken)
	if err != nil {
		http.Error(w, "no token", http.StatusBadRequest)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "no email", http.StatusBadRequest)
		return
	}

	responseStatus := http.StatusOK
	if err := a.store.VisitTx(ctx, func(ctx context.Context, q *queries.Queries) error {
		session, err := q.GetUserBySession(ctx, ck.Value)
		if err != nil {
			responseStatus = http.StatusUnauthorized
			return fmt.Errorf("q.GetSession: %w", err)
		}

		if session.Email != email || !session.IsAdmin {
			responseStatus = http.StatusUnauthorized
			return fmt.Errorf("q.GetSession: %w", err)
		}

		if _, err := q.DeleteUser(ctx, ck.Value); err != nil {
			responseStatus = http.StatusInternalServerError
			return fmt.Errorf("q.DeleteUser: %w", err)
		}

		return nil
	}); err != nil {
		log.Error().Err(err).Msg("Failed to delete user")
		http.Error(w, "failed to delete user", responseStatus)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ActivateUser activates the user
// GET /api/user?email=email
func (a *API) ActivateUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	ck, err := r.Cookie(CookieNameToken)
	if err != nil {
		http.Error(w, "no token", http.StatusBadRequest)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "no email", http.StatusBadRequest)
		return
	}

	responseStatus := http.StatusOK
	if err := a.store.Visit(ctx, func(ctx context.Context, q *queries.Queries) error {
		session, err := q.GetUserBySession(ctx, ck.Value)
		if err != nil {
			responseStatus = http.StatusUnauthorized
			return fmt.Errorf("q.GetSession: %w", err)
		}

		if !session.IsAdmin {
			responseStatus = http.StatusUnauthorized
			return fmt.Errorf("q.GetSession: %w", err)
		}

		userInfo, err := q.GetUser(ctx, email)
		if err != nil || userInfo.Deleted {
			responseStatus = http.StatusNotFound
			return fmt.Errorf("q.GetUser: %w", err)
		}

		if _, err := q.UpdateUser(ctx, queries.UpdateUserParams{
			Email:     email,
			Validated: true,
			IsAdmin:   userInfo.IsAdmin,
			Deleted:   userInfo.Deleted,
		}); err != nil {
			responseStatus = http.StatusInternalServerError
			return fmt.Errorf("q.UpdateUser: %w", err)
		}

		return nil
	}); err != nil {
		log.Error().Err(err).Msg("Failed to activate user")
		http.Error(w, "failed to activate user", responseStatus)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// PromoteUser promotes the user
// PATCH /api/user/promotion?email=email
func (a *API) PromoteUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	ck, err := r.Cookie(CookieNameToken)
	if err != nil {
		http.Error(w, "no token", http.StatusBadRequest)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "no email", http.StatusBadRequest)
		return
	}

	responseStatus := http.StatusOK
	if err := a.store.Visit(ctx, func(ctx context.Context, q *queries.Queries) error {
		session, err := q.GetUserBySession(ctx, ck.Value)
		if err != nil {
			responseStatus = http.StatusUnauthorized
			return fmt.Errorf("q.GetSession: %w", err)
		}

		if !session.IsAdmin {
			responseStatus = http.StatusUnauthorized
			return fmt.Errorf("q.GetSession: %w", err)
		}

		userInfo, err := q.GetUser(ctx, email)
		if err != nil || userInfo.Deleted {
			responseStatus = http.StatusNotFound
			return fmt.Errorf("q.GetUser: %w", err)
		}

		if _, err := q.UpdateUser(ctx, queries.UpdateUserParams{
			Email:     email,
			Validated: userInfo.Validated,
			IsAdmin:   true,
			Deleted:   userInfo.Deleted,
		}); err != nil {
			responseStatus = http.StatusInternalServerError
			return fmt.Errorf("q.UpdateUser: %w", err)
		}

		return nil
	}); err != nil {
		log.Error().Err(err).Msg("Failed to promote user")
		http.Error(w, "failed to promote user", responseStatus)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DemoteUser demotes the user
// PATCH /api/user/demotion?email=email
func (a *API) DemoteUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	ck, err := r.Cookie(CookieNameToken)
	if err != nil {
		http.Error(w, "no token", http.StatusBadRequest)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "no email", http.StatusBadRequest)
		return
	}

	responseStatus := http.StatusOK
	if err := a.store.Visit(ctx, func(ctx context.Context, q *queries.Queries) error {
		session, err := q.GetUserBySession(ctx, ck.Value)
		if err != nil {
			responseStatus = http.StatusUnauthorized
			return fmt.Errorf("q.GetSession: %w", err)
		}

		if !session.IsAdmin {
			responseStatus = http.StatusUnauthorized
			return fmt.Errorf("q.GetSession: %w", err)
		}

		userInfo, err := q.GetUser(ctx, email)
		if err != nil || userInfo.Deleted {
			responseStatus = http.StatusNotFound
			return fmt.Errorf("q.GetUser: %w", err)
		}

		if _, err := q.UpdateUser(ctx, queries.UpdateUserParams{
			Email:     email,
			Validated: userInfo.Validated,
			IsAdmin:   false,
			Deleted:   userInfo.Deleted,
		}); err != nil {
			responseStatus = http.StatusInternalServerError
			return fmt.Errorf("q.UpdateUser: %w", err)
		}

		return nil
	}); err != nil {
		log.Error().Err(err).Msg("Failed to demote user")
		http.Error(w, "failed to demote user", responseStatus)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type CreateClusterRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Password    string `json:"password"`
}

// CreateCluster creates a new cluster
// POST /api/cluster
func (a *API) CreateCluster(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	defer r.Body.Close()

	ck, err := r.Cookie(CookieNameToken)
	if err != nil {
		http.Error(w, "no token", http.StatusBadRequest)
		return
	}

	request := &CreateClusterRequest{}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	responseStatus := http.StatusOK
	if err := a.store.VisitTx(ctx, func(ctx context.Context, q *queries.Queries) error {
		_, err := q.GetSession(ctx, ck.Value)
		if err != nil {
			responseStatus = http.StatusUnauthorized
			return fmt.Errorf("q.GetSession: %w", err)
		}

		_, err = q.CreateCluster(ctx, queries.CreateClusterParams{
			Name: request.Name,
		})
		if err != nil {
			responseStatus = http.StatusBadRequest
			return fmt.Errorf("q.CreateCluster: %w", err)
		}

		return nil
	}); err != nil {
		log.Error().Err(err).Any("request", request).Msg("Failed to create cluster")
		http.Error(w, "failed to create cluster", responseStatus)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type GetClusterRequest struct {
	Name string `json:"name"`
}

type GetClusterResponse struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GetCluster returns the cluster information
// GET /api/cluster?name=cluster_name
func (a *API) GetCluster(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	defer r.Body.Close()

	ck, err := r.Cookie(CookieNameToken)
	if err != nil {
		http.Error(w, "no token", http.StatusBadRequest)
		return
	}

	request := &GetClusterRequest{
		Name: r.URL.Query().Get("name"),
	}

	response := &GetClusterResponse{}
	responseStatus := http.StatusOK
	if err := a.store.Visit(ctx, func(ctx context.Context, q *queries.Queries) error {
		_, err := q.GetSession(ctx, ck.Value)
		if err != nil {
			responseStatus = http.StatusUnauthorized
			return fmt.Errorf("q.GetSession: %w", err)
		}

		resp, err := q.GetCluster(ctx, request.Name)
		if err != nil {
			responseStatus = http.StatusNotFound
			return fmt.Errorf("q.GetCluster: %w", err)
		}

		response.Name = resp.Name
		response.Description = resp.Description.String
		response.CreatedAt = resp.CreatedAt.Time
		response.UpdatedAt = resp.UpdatedAt.Time

		return nil
	}); err != nil {
		log.Error().Err(err).Any("request", request).Msg("Failed to get cluster")
		http.Error(w, "failed to get cluster", responseStatus)
		return
	}

	data, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	w.WriteHeader(http.StatusOK)
}

type GetClustersRequest struct {
	Count  int32  `json:"count"`
	Cursor string `json:"cursor"`
}

type GetClustersResponse struct {
	Clusters []GetClusterResponse `json:"clusters"`
}

// GetClusters returns the list of clusters
// GET /api/clusters?count=10&cursor=cursor
func (a *API) GetClusters(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	defer r.Body.Close()

	ck, err := r.Cookie(CookieNameToken)
	if err != nil {
		http.Error(w, "no token", http.StatusBadRequest)
		return
	}

	request := &GetClustersRequest{
		Count:  10,
		Cursor: r.URL.Query().Get("cursor"),
	}
	if r.URL.Query().Get("count") != "" {
		request.Count = 10
	}

	response := make([]GetClusterResponse, 0)
	responseStatus := http.StatusOK
	if err := a.store.Visit(ctx, func(ctx context.Context, q *queries.Queries) error {
		_, err := q.GetSession(ctx, ck.Value)
		if err != nil {
			responseStatus = http.StatusUnauthorized
			return fmt.Errorf("q.GetSession: %w", err)
		}

		resp, err := ([]queries.Cluster)(nil), error(nil)
		switch len(request.Cursor) {
		case 0:
			resp, err = q.GetClusters(ctx, request.Count)
		default:
			resp, err = q.GetClustersByCursor(ctx, queries.GetClustersByCursorParams{
				Name:  request.Cursor,
				Limit: request.Count,
			})
		}
		if err != nil {
			responseStatus = http.StatusInternalServerError
			return fmt.Errorf("q.GetClusters: %w", err)
		}

		for _, r := range resp {
			response = append(response, GetClusterResponse{
				Name:        r.Name,
				Description: r.Description.String,
				CreatedAt:   r.CreatedAt.Time,
				UpdatedAt:   r.UpdatedAt.Time,
			})
		}

		return nil
	}); err != nil {
		log.Error().Err(err).Any("request", request).Msg("Failed to get clusters")
		http.Error(w, "failed to get clusters", responseStatus)
		return
	}

	data, _ := json.Marshal(&GetClustersResponse{Clusters: response})
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	w.WriteHeader(http.StatusOK)
}

type UpdateClusterRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Password    *string `json:"password,omitempty"`
}

// UpdateCluster updates the cluster information
// PUT /api/cluster?name=cluster_name
func (a *API) UpdateCluster(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	defer r.Body.Close()

	ck, err := r.Cookie(CookieNameToken)
	if err != nil {
		http.Error(w, "no token", http.StatusBadRequest)
		return
	}

	request := &UpdateClusterRequest{
		Name: r.URL.Query().Get("name"),
	}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	responseStatus := http.StatusOK
	if err := a.store.VisitTx(ctx, func(ctx context.Context, q *queries.Queries) error {
		_, err := q.GetSession(ctx, ck.Value)
		if err != nil {
			responseStatus = http.StatusUnauthorized
			return fmt.Errorf("q.GetSession: %w", err)
		}

		origin, err := q.GetCluster(ctx, request.Name)
		if err != nil {
			responseStatus = http.StatusNotFound
			return fmt.Errorf("q.GetCluster: %w", err)
		}

		if request.Description == nil {
			request.Description = &origin.Description.String
		}
		if request.Password == nil {
			request.Password = &origin.Password
		}

		if _, err := q.UpdateCluster(ctx, queries.UpdateClusterParams{
			Name: request.Name,
			Description: pgtype.Text{
				String: *request.Description,
				Valid:  true,
			},
			Password: *request.Password,
		}); err != nil {
			responseStatus = http.StatusInternalServerError
			return fmt.Errorf("q.UpdateCluster: %w", err)
		}

		return nil
	}); err != nil {
		log.Error().Err(err).Any("request", request).Msg("Failed to update cluster")
		http.Error(w, "failed to update cluster", responseStatus)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type DeleteClusterRequest struct {
	Name string `json:"name"`
}

// DeleteCluster deletes the cluster
// DELETE /api/cluster?name=cluster_name
func (a *API) DeleteCluster(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	defer r.Body.Close()

	ck, err := r.Cookie(CookieNameToken)
	if err != nil {
		http.Error(w, "no token", http.StatusBadRequest)
		return
	}

	request := &DeleteClusterRequest{
		Name: r.URL.Query().Get("name"),
	}

	responseStatus := http.StatusOK
	if err := a.store.VisitTx(ctx, func(ctx context.Context, q *queries.Queries) error {
		_, err := q.GetSession(ctx, ck.Value)
		if err != nil {
			responseStatus = http.StatusUnauthorized
			return fmt.Errorf("q.GetSession: %w", err)
		}

		if _, err := q.DeleteCluster(ctx, request.Name); err != nil {
			responseStatus = http.StatusInternalServerError
			return fmt.Errorf("q.DeleteCluster: %w", err)
		}

		return nil
	}); err != nil {
		log.Error().Err(err).Any("request", request).Msg("Failed to delete cluster")
		http.Error(w, "failed to delete cluster", responseStatus)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type CreateNodeRequest struct {
	Name        string `json:"name"`
	ClusterName string `json:"cluster_name"`
	NodeID      string `json:"node_id"`
	Host        string `json:"host"`
	Port        int32  `json:"port"`
}

// CreateNode creates a new node
// POST /api/node
func (a *API) CreateNode(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	defer r.Body.Close()

	ck, err := r.Cookie(CookieNameToken)
	if err != nil {
		http.Error(w, "no token", http.StatusBadRequest)
		return
	}

	request := &CreateNodeRequest{}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	responseStatus := http.StatusOK
	if err := a.store.VisitTx(ctx, func(ctx context.Context, q *queries.Queries) error {
		_, err := q.GetSession(ctx, ck.Value)
		if err != nil {
			responseStatus = http.StatusUnauthorized
			return fmt.Errorf("q.GetSession: %w", err)
		}

		_, err = q.GetCluster(ctx, request.ClusterName)
		if err != nil {
			responseStatus = http.StatusNotFound
			return fmt.Errorf("q.GetCluster: %w", err)
		}

		_, err = q.GetNodeByHostPort(ctx, queries.GetNodeByHostPortParams{
			Host: request.Host,
			Port: request.Port,
		})
		if err == nil {
			responseStatus = http.StatusConflict
			return fmt.Errorf("q.GetNodeByHostPort: %w", err)
		}

		_, err = q.GetNodeByNodeID(ctx, queries.GetNodeByNodeIDParams{
			NodeID: request.NodeID,
		})
		if err == nil {
			responseStatus = http.StatusConflict
			return fmt.Errorf("q.GetNodeByNodeID: %w", err)
		}

		_, err = q.CreateNode(ctx, queries.CreateNodeParams{
			Name:   request.ClusterName,
			NodeID: request.NodeID,
			Host:   request.Host,
			Port:   request.Port,
		})
		if err != nil {
			responseStatus = http.StatusBadRequest
			return fmt.Errorf("q.CreateNode: %w", err)
		}

		return nil
	}); err != nil {
		log.Error().Err(err).Any("request", request).Msg("Failed to create node")
		http.Error(w, "failed to create node", responseStatus)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type GetNodeRequest struct {
	ClusterName string `json:"cluster_name"`
	NodeID      string `json:"node_id"`
	Host        string `json:"host"`
	Port        int32  `json:"port"`
}

type GetNodeResponse struct {
	ClusterName string    `json:"name"`
	NodeID      string    `json:"node_id"`
	Host        string    `json:"host"`
	Port        int32     `json:"port"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GetNode returns the node information
// GET /api/node?cluster_name=cluster_name&host=host&port=port
// GET /api/node?cluster_name=cluster_name&node_id=node_id
func (a *API) GetNode(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	defer r.Body.Close()

	ck, err := r.Cookie(CookieNameToken)
	if err != nil {
		http.Error(w, "no token", http.StatusBadRequest)
		return
	}

	request := &GetNodeRequest{
		ClusterName: r.URL.Query().Get("cluster_name"),
		NodeID:      r.URL.Query().Get("node_id"),
		Host:        r.URL.Query().Get("host"),
	}
	if r.URL.Query().Get("port") != "" {
		request.Port = 0
	}

	response := &GetNodeResponse{}
	responseStatus := http.StatusOK
	if err := a.store.Visit(ctx, func(ctx context.Context, q *queries.Queries) error {
		_, err := q.GetSession(ctx, ck.Value)
		if err != nil {
			responseStatus = http.StatusUnauthorized
			return fmt.Errorf("q.GetSession: %w", err)
		}

		resp, err := queries.Node{}, error(nil)
		switch request.Port {
		case 0:
			resp, err = q.GetNodeByNodeID(ctx, queries.GetNodeByNodeIDParams{
				NodeID: request.NodeID,
			})
		default:
			resp, err = q.GetNodeByHostPort(ctx, queries.GetNodeByHostPortParams{
				Host: request.Host,
				Port: request.Port,
			})
		}
		if err != nil {
			responseStatus = http.StatusNotFound
			return fmt.Errorf("q.GetNode: %w", err)
		}

		clusterInfo, err := q.GetCluster(ctx, request.ClusterName)
		if err != nil {
			responseStatus = http.StatusNotFound
			return fmt.Errorf("q.GetCluster: %w", err)
		}

		response.ClusterName = clusterInfo.Name
		response.NodeID = resp.NodeID
		response.Host = resp.Host
		response.Port = resp.Port
		response.CreatedAt = resp.CreatedAt.Time
		response.UpdatedAt = resp.UpdatedAt.Time

		return nil
	}); err != nil {
		log.Error().Err(err).Any("request", request).Msg("Failed to get node")
		http.Error(w, "failed to get node", responseStatus)
		return
	}

	data, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	w.WriteHeader(http.StatusOK)
}

type GetNodesRequest struct {
	ClusterName string `json:"cluster_name"`
	Count       int32  `json:"count"`
	Cursor      string `json:"cursor"`
}

type GetNodesResponse struct {
	Nodes []GetNodeResponse `json:"nodes"`
}

// GetNodes returns the list of nodes
// GET /api/nodes?cluster_name=cluster_name&count=10&cursor=cursor
func (a *API) GetNodes(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	defer r.Body.Close()

	ck, err := r.Cookie(CookieNameToken)
	if err != nil {
		http.Error(w, "no token", http.StatusBadRequest)
		return
	}

	request := &GetNodesRequest{
		ClusterName: r.URL.Query().Get("cluster_name"),
		Count:       10,
		Cursor:      r.URL.Query().Get("cursor"),
	}
	if r.URL.Query().Get("count") != "" {
		request.Count = 10
	}

	response := make([]GetNodeResponse, 0)
	responseStatus := http.StatusOK
	if err := a.store.Visit(ctx, func(ctx context.Context, q *queries.Queries) error {
		_, err := q.GetSession(ctx, ck.Value)
		if err != nil {
			responseStatus = http.StatusUnauthorized
			return fmt.Errorf("q.GetSession: %w", err)
		}

		_, err = q.GetCluster(ctx, request.ClusterName)
		if err != nil {
			responseStatus = http.StatusNotFound
			return fmt.Errorf("q.GetCluster: %w", err)
		}

		resp, err := ([]queries.Node)(nil), error(nil)
		switch len(request.Cursor) {
		case 0:
			resp, err = q.GetNodes(ctx, queries.GetNodesParams{
				Name:  request.ClusterName,
				Limit: request.Count,
			})
		default:
			resp, err = q.GetNodesByCursor(ctx, queries.GetNodesByCursorParams{
				Name:   request.ClusterName,
				NodeID: request.Cursor,
				Limit:  request.Count,
			})
		}
		if err != nil {
			responseStatus = http.StatusInternalServerError
			return fmt.Errorf("q.GetNodes: %w", err)
		}

		for _, r := range resp {
			response = append(response, GetNodeResponse{
				ClusterName: request.ClusterName,
				NodeID:      r.NodeID,
				Host:        r.Host,
				Port:        r.Port,
				CreatedAt:   r.CreatedAt.Time,
				UpdatedAt:   r.UpdatedAt.Time,
			})
		}

		return nil
	}); err != nil {
		log.Error().Err(err).Any("request", request).Msg("Failed to get nodes")
		http.Error(w, "failed to get nodes", responseStatus)
		return
	}

	data, _ := json.Marshal(&GetNodesResponse{Nodes: response})
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	w.WriteHeader(http.StatusOK)
}

type DeleteNodeRequest struct {
	ClusterName string `json:"cluster_name"`
	NodeID      string `json:"node_id"`
}

// DeleteNode deletes the node
// DELETE /api/node?cluster_name=cluster_name&node_id=node_id
func (a *API) DeleteNode(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	defer r.Body.Close()

	ck, err := r.Cookie(CookieNameToken)
	if err != nil {
		http.Error(w, "no token", http.StatusBadRequest)
		return
	}

	request := &DeleteNodeRequest{
		ClusterName: r.URL.Query().Get("cluster_name"),
		NodeID:      r.URL.Query().Get("node_id"),
	}

	responseStatus := http.StatusOK
	if err := a.store.VisitTx(ctx, func(ctx context.Context, q *queries.Queries) error {
		_, err := q.GetSession(ctx, ck.Value)
		if err != nil {
			responseStatus = http.StatusUnauthorized
			return fmt.Errorf("q.GetSession: %w", err)
		}

		_, err = q.GetCluster(ctx, request.ClusterName)
		if err != nil {
			responseStatus = http.StatusNotFound
			return fmt.Errorf("q.GetCluster: %w", err)
		}

		_, err = q.GetNodeByNodeID(ctx, queries.GetNodeByNodeIDParams{
			NodeID: request.NodeID,
			Name:   request.ClusterName,
		})
		if err != nil {
			responseStatus = http.StatusNotFound
			return fmt.Errorf("q.GetNodeByNodeID: %w", err)
		}

		if _, err := q.DeleteNode(ctx, queries.DeleteNodeParams{
			Name:   request.ClusterName,
			NodeID: request.NodeID,
		}); err != nil {
			responseStatus = http.StatusInternalServerError
			return fmt.Errorf("q.DeleteNode: %w", err)
		}

		return nil
	}); err != nil {
		log.Error().Err(err).Any("request", request).Msg("Failed to delete node")
		http.Error(w, "failed to delete node", responseStatus)
		return
	}

	w.WriteHeader(http.StatusOK)
}
