package service

import (
	"context"
	"fmt"
	"time"

	"github.com/faust8888/GophKeeper/internal/client"
	"github.com/faust8888/GophKeeper/internal/client/model"
	"github.com/faust8888/GophKeeper/internal/client/repository"
	pb "github.com/faust8888/GophKeeper/internal/pkg/proto/gophkeeperpb"
	"github.com/faust8888/GophKeeper/internal/pkg/security"
)

// private implementation of UserService
type userService struct {
	client  client.Client
	repo    repository.Repository
	session *model.Session
}

func newUserService(repo repository.Repository, client client.Client) userService {
	session := repo.GetSession()
	return userService{client: client, repo: repo, session: session}
}

func (s *userService) Register(ctx context.Context, req *pb.RegisterUserRequest) error {
	if req.Login == "" || len(req.Password) == 0 {
		return ErrRequiredArgumentIsMissing
	}

	_, err := s.client.RegisterUser(ctx, req)
	if err != nil {
		return fmt.Errorf("server failed to sign up req: %w", err)
	}

	return nil
}

func (s *userService) Login(ctx context.Context, req *pb.LoginRequest) error {
	if req.Login == "" || req.Password == "" {
		return ErrRequiredArgumentIsMissing
	}

	res, err := s.client.Login(ctx, req)
	if err != nil {
		return fmt.Errorf("server failed to sign in req: %w", err)
	}

	session := &model.Session{
		Token:  res.Token,
		UserID: res.UserId,
		Salt:   security.GenerateSalt(),
	}

	if err = s.repo.SaveSession(session, time.Now().Add(12*time.Hour)); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}
