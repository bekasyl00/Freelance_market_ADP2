# Freelance Market REST API

Base URL for local development:

```text
http://localhost:8088/api
```

## Health

```http
GET /health
```

Returns:

```json
{ "status": "ok" }
```

## Dashboard

```http
GET /api/summary
```

Returns:

```json
{
  "activeJobs": 4,
  "escrowBalance": 3100,
  "proposals": 3,
  "rating": 4.7
}
```

## Jobs

```http
GET /api/jobs
```

```http
POST /api/jobs
Content-Type: application/json

{
  "title": "Write website copy",
  "description": "Create homepage copy and onboarding text.",
  "budget": 500,
  "deadline": "2026-06-20",
  "skills": ["Copywriting", "SEO"]
}
```

```http
POST /api/jobs/{jobId}/apply
Content-Type: application/json

{
  "coverLetter": "I can deliver this with clear milestones.",
  "bid": 500,
  "estimatedDays": 7
}
```

## Profile

```http
GET /api/profile
```

```http
PUT /api/profile/skills
Content-Type: application/json

{
  "userId": "optional-user-id",
  "skills": ["Web Design", "Frontend", "SEO"]
}
```

## Payments

```http
GET /api/payments
```

```http
POST /api/payments/deposit
Content-Type: application/json

{
  "amount": 250
}
```

```http
POST /api/payments/escrows
Content-Type: application/json

{
  "jobId": "job-id",
  "amount": 500
}
```

```http
POST /api/payments/escrows/{escrowId}/release
```

## Notes

- The current gateway uses demo users when `clientId`, `freelancerId`, or `userId` are omitted.
- Money is sent as normal decimal values in HTTP JSON and stored as integer cents in PostgreSQL.
- Frontend should use only this API. Direct table access is reserved for backend repositories.
