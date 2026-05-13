package grpc

import (
	"job_service/internal/domain"
	"job_service/proto"
)

func domainJobToProto(j *domain.Job) *proto.Job {
	if j == nil {
		return nil
	}
	return &proto.Job{
		Id:            j.ID,
		ClientId:      j.ClientID,
		Title:         j.Title,
		Description:   j.Description,
		BudgetCents:   j.BudgetCents,
		Status:        domainStatusToProto(j.Status),
		CreatedAtUnix: j.CreatedAtUnix,
	}
}

func domainStatusToProto(s domain.JobStatus) proto.JobStatus {
	switch s {
	case domain.JobStatusOpen:
		return proto.JobStatus_JOB_STATUS_OPEN
	case domain.JobStatusInProgress:
		return proto.JobStatus_JOB_STATUS_IN_PROGRESS
	case domain.JobStatusCompleted:
		return proto.JobStatus_JOB_STATUS_COMPLETED
	case domain.JobStatusCancelled:
		return proto.JobStatus_JOB_STATUS_CANCELLED
	default:
		return proto.JobStatus_JOB_STATUS_UNSPECIFIED
	}
}
