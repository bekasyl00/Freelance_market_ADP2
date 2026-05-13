package grpc

import (
	"context"
	"database/sql"
	"errors"

	"user_service/internal/domain"
	"user_service/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	proto.UnimplementedUserServiceServer
	usecase domain.UserUsecase
}

func NewUserHandler(usecase domain.UserUsecase) *UserHandler {
	return &UserHandler{usecase: usecase}
}

func (h *UserHandler) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	role, ok := protoRoleToDomain(req.GetRole())
	if !ok {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidRole.Error())
	}

	id, err := h.usecase.Register(ctx, req.GetUsername(), req.GetEmail(), req.GetPassword(), role)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRole) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "register: %v", err)
	}
	return &proto.RegisterResponse{Id: id, Message: "User created successfully"}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	token, err := h.usecase.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "login: %v", err)
	}
	return &proto.LoginResponse{Token: token}, nil
}

func (h *UserHandler) GetProfile(ctx context.Context, req *proto.GetProfileRequest) (*proto.UserResponse, error) {
	u, err := h.usecase.GetProfile(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "get profile: %v", err)
	}
	return &proto.UserResponse{User: domainUserToProto(u)}, nil
}

func (h *UserHandler) UpdateSkills(ctx context.Context, req *proto.UpdateSkillsRequest) (*proto.UserResponse, error) {
	if err := h.usecase.UpdateSkills(ctx, req.GetId(), req.GetSkills()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		if errors.Is(err, domain.ErrOnlyFreelancerSkills) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "update skills: %v", err)
	}

	u, err := h.usecase.GetProfile(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "reload profile: %v", err)
	}
	return &proto.UserResponse{User: domainUserToProto(u)}, nil
}
