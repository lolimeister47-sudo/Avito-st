package httpadapter

import (
	"encoding/json"
	"net/http"

	"prservice/internal/adapter/http/api"
	"prservice/internal/domain"
	"prservice/internal/usecase"
)

type Server struct {
	teamSvc *usecase.TeamService
	userSvc *usecase.UserService
	prSvc   *usecase.PRService
	prRepo  domain.PRRepository
}

func NewServer(
	team *usecase.TeamService,
	user *usecase.UserService,
	pr *usecase.PRService,
	prRepo domain.PRRepository,
) *Server {
	return &Server{
		teamSvc: team,
		userSvc: user,
		prSvc:   pr,
		prRepo:  prRepo,
	}
}

// ======== /team/add (POST) ========

func (s *Server) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	var req api.Team
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	dTeam := domain.Team{
		Name:    domain.TeamName(req.TeamName),
		Members: make([]domain.TeamMember, 0, len(req.Members)),
	}
	for _, m := range req.Members {
		dTeam.Members = append(dTeam.Members, domain.TeamMember{
			UserID:   domain.UserID(m.UserId),
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	res, err := s.teamSvc.AddTeam(r.Context(), dTeam)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := api.Team{
		TeamName: string(res.Name),
		Members:  make([]api.TeamMember, 0, len(res.Members)),
	}
	for _, m := range res.Members {
		resp.Members = append(resp.Members, api.TeamMember{
			UserId:   string(m.UserID),
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	writeJSON(w, http.StatusCreated, struct {
		Team api.Team `json:"team"`
	}{Team: resp})
}

// ======== /team/get (GET) ========

func (s *Server) GetTeamGet(w http.ResponseWriter, r *http.Request, params api.GetTeamGetParams) {
	res, err := s.teamSvc.GetTeam(r.Context(), domain.TeamName(params.TeamName))
	if err != nil {
		writeError(w, err)
		return
	}

	resp := api.Team{
		TeamName: string(res.Name),
		Members:  make([]api.TeamMember, 0, len(res.Members)),
	}
	for _, m := range res.Members {
		resp.Members = append(resp.Members, api.TeamMember{
			UserId:   string(m.UserID),
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// ======== /users/setIsActive (POST) ========

func (s *Server) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserId   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	u, err := s.userSvc.SetIsActive(r.Context(), domain.UserID(req.UserId), req.IsActive)
	if err != nil {
		writeError(w, err)
		return
	}

	respUser := api.User{
		UserId:   string(u.ID),
		Username: u.Username,
		TeamName: string(u.TeamName),
		IsActive: u.IsActive,
	}

	writeJSON(w, http.StatusOK, struct {
		User api.User `json:"user"`
	}{User: respUser})
}

// ======== /pullRequest/create (POST) ========

func (s *Server) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestId   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorId        string `json:"author_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	dPR := domain.PullRequest{
		ID:       domain.PullRequestID(req.PullRequestId),
		Name:     req.PullRequestName,
		AuthorID: domain.UserID(req.AuthorId),
	}

	pr, err := s.prSvc.CreatePR(r.Context(), dPR)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := mapPRToAPI(pr)
	writeJSON(w, http.StatusCreated, struct {
		Pr api.PullRequest `json:"pr"`
	}{Pr: resp})
}

// ======== /pullRequest/merge (POST) ========

func (s *Server) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestId string `json:"pull_request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	pr, err := s.prSvc.Merge(r.Context(), domain.PullRequestID(req.PullRequestId))
	if err != nil {
		writeError(w, err)
		return
	}

	resp := mapPRToAPI(pr)
	writeJSON(w, http.StatusOK, struct {
		Pr api.PullRequest `json:"pr"`
	}{Pr: resp})
}

// ======== /pullRequest/reassign (POST) ========

func (s *Server) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestId string `json:"pull_request_id"`
		OldUserId     string `json:"old_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	pr, newRev, err := s.prSvc.ReassignReviewer(
		r.Context(),
		domain.PullRequestID(req.PullRequestId),
		domain.UserID(req.OldUserId),
	)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := mapPRToAPI(pr)
	writeJSON(w, http.StatusOK, struct {
		Pr         api.PullRequest `json:"pr"`
		ReplacedBy string          `json:"replaced_by"`
	}{Pr: resp, ReplacedBy: string(newRev)})
}

// ======== /users/getReview (GET) ========

func (s *Server) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params api.GetUsersGetReviewParams) {
	prs, err := s.prSvc.ListByReviewer(r.Context(), domain.UserID(params.UserId))
	if err != nil {
		writeError(w, err)
		return
	}

	resp := struct {
		UserId       string                 `json:"user_id"`
		PullRequests []api.PullRequestShort `json:"pull_requests"`
	}{
		UserId:       params.UserId,
		PullRequests: make([]api.PullRequestShort, 0, len(prs)),
	}

	for _, pr := range prs {
		resp.PullRequests = append(resp.PullRequests, api.PullRequestShort{
			PullRequestId:   string(pr.ID),
			PullRequestName: pr.Name,
			AuthorId:        string(pr.AuthorID),
			Status:         api.PullRequestShortStatus(pr.Status),
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// ======== helpers ========

func mapPRToAPI(pr *domain.PullRequest) api.PullRequest {
	resp := api.PullRequest{
		PullRequestId:   string(pr.ID),
		PullRequestName: pr.Name,
		AuthorId:        string(pr.AuthorID),
		Status:         api.PullRequestStatus(pr.Status),
		AssignedReviewers: func() []string {
			res := make([]string, len(pr.AssignedReviewers))
			for i, id := range pr.AssignedReviewers {
				res[i] = string(id)
			}
			return res
		}(),
		CreatedAt: pr.CreatedAt,
		MergedAt:  pr.MergedAt,
	}
	return resp
}
