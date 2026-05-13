package grpc

import (
	"user_service/internal/domain"
	"user_service/proto"
)

func domainUserToProto(u *domain.User) *proto.User {
	if u == nil {
		return nil
	}
	return &proto.User{
		Id:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Role:     domainRoleToProto(u.Role),
		Skills:   append([]string(nil), u.Skills...),
		Rating:   u.Rating,
	}
}

func domainRoleToProto(r domain.Role) proto.Role {
	switch r {
	case domain.RoleAdmin:
		return proto.Role_ADMIN
	case domain.RoleClient:
		return proto.Role_CLIENT
	case domain.RoleFreelancer:
		return proto.Role_FREELANCER
	default:
		return proto.Role_UNKNOWN
	}
}

func protoRoleToDomain(r proto.Role) (domain.Role, bool) {
	switch r {
	case proto.Role_ADMIN:
		return domain.RoleAdmin, true
	case proto.Role_CLIENT:
		return domain.RoleClient, true
	case proto.Role_FREELANCER:
		return domain.RoleFreelancer, true
	default:
		return "", false
	}
}
