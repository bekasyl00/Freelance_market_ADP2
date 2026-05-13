package grpc

import (
	"context"
	"errors"

	"job_service/internal/domain"
	"job_service/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type JobHandler struct {
	proto.UnimplementedJobServiceServer
	uc domain.JobUsecase
}

func NewJobHandler(uc domain.JobUsecase) *JobHandler {
	return &JobHandler{uc: uc}
}

func (h *JobHandler) CreateJob(ctx context.Context, req *proto.CreateJobRequest) (*proto.CreateJobResponse, error) {
	j, err := h.uc.CreateJob(ctx, req.GetClientId(), req.GetTitle(), req.GetDescription(), req.GetBudgetCents())
	if err != nil {
		return nil, mapErr(err)
	}
	return &proto.CreateJobResponse{Job: domainJobToProto(j)}, nil
}

func (h *JobHandler) ListJobs(ctx context.Context, req *proto.ListJobsRequest) (*proto.ListJobsResponse, error) {
	jobs, next, err := h.uc.ListJobs(ctx, req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*proto.Job, 0, len(jobs))
	for i := range jobs {
		out = append(out, domainJobToProto(&jobs[i]))
	}
	return &proto.ListJobsResponse{Jobs: out, NextPageToken: next}, nil
}

func (h *JobHandler) ApplyToJob(ctx context.Context, req *proto.ApplyToJobRequest) (*proto.ApplyToJobResponse, error) {
	id, err := h.uc.ApplyToJob(ctx, req.GetJobId(), req.GetFreelancerId(), req.GetCoverLetter(), req.GetBidCents())
	if err != nil {
		return nil, mapErr(err)
	}
	return &proto.ApplyToJobResponse{ApplicationId: id}, nil
}

func (h *JobHandler) GetJobDetails(ctx context.Context, req *proto.GetJobDetailsRequest) (*proto.JobDetailsResponse, error) {
	j, n, err := h.uc.GetJobDetails(ctx, req.GetJobId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &proto.JobDetailsResponse{Job: domainJobToProto(j), ApplicationsCount: n}, nil
}

func (h *JobHandler) CompleteJob(ctx context.Context, req *proto.CompleteJobRequest) (*proto.CompleteJobResponse, error) {
	j, err := h.uc.CompleteJob(ctx, req.GetJobId(), req.GetClientId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &proto.CompleteJobResponse{Job: domainJobToProto(j)}, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrJobNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrJobNotOpen):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrDuplicateProposal):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrForbiddenClient):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, domain.ErrJobAlreadyTerminal):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}
