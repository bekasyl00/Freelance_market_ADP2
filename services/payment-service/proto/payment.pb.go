package proto

type EscrowStatus int32

const (
	EscrowStatus_ESCROW_STATUS_UNSPECIFIED EscrowStatus = 0
	EscrowStatus_ESCROW_STATUS_HELD        EscrowStatus = 1
	EscrowStatus_ESCROW_STATUS_RELEASED    EscrowStatus = 2
	EscrowStatus_ESCROW_STATUS_REFUNDED    EscrowStatus = 3
	EscrowStatus_ESCROW_STATUS_CANCELLED   EscrowStatus = 4
)

type TransactionType int32

const (
	TransactionType_TRANSACTION_TYPE_UNSPECIFIED    TransactionType = 0
	TransactionType_TRANSACTION_TYPE_DEPOSIT        TransactionType = 1
	TransactionType_TRANSACTION_TYPE_ESCROW_HOLD    TransactionType = 2
	TransactionType_TRANSACTION_TYPE_ESCROW_RELEASE TransactionType = 3
	TransactionType_TRANSACTION_TYPE_REFUND         TransactionType = 4
	TransactionType_TRANSACTION_TYPE_WITHDRAWAL     TransactionType = 5
)

type TransactionStatus int32

const (
	TransactionStatus_TRANSACTION_STATUS_UNSPECIFIED TransactionStatus = 0
	TransactionStatus_TRANSACTION_STATUS_PENDING     TransactionStatus = 1
	TransactionStatus_TRANSACTION_STATUS_COMPLETED   TransactionStatus = 2
	TransactionStatus_TRANSACTION_STATUS_FAILED      TransactionStatus = 3
	TransactionStatus_TRANSACTION_STATUS_CANCELLED   TransactionStatus = 4
)

type PaymentAccount struct {
	Id             string
	UserId         string
	AvailableCents int64
	EscrowCents    int64
	Currency       string
	CreatedAtUnix  int64
	UpdatedAtUnix  int64
}

type Escrow struct {
	Id             string
	JobId          string
	ClientId       string
	FreelancerId   string
	AmountCents    int64
	Currency       string
	Status         EscrowStatus
	HeldAtUnix     int64
	ReleasedAtUnix int64
	RefundedAtUnix int64
}

type Transaction struct {
	Id                string
	UserId            string
	JobId             string
	EscrowId          string
	Type              TransactionType
	AmountCents       int64
	Currency          string
	Status            TransactionStatus
	Provider          string
	ProviderReference string
	CreatedAtUnix     int64
}

type CreateAccountRequest struct {
	UserId   string
	Currency string
}

func (x *CreateAccountRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}
func (x *CreateAccountRequest) GetCurrency() string {
	if x != nil {
		return x.Currency
	}
	return ""
}

type GetAccountRequest struct {
	UserId string
}

func (x *GetAccountRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type DepositRequest struct {
	UserId            string
	AmountCents       int64
	Currency          string
	Provider          string
	ProviderReference string
}

func (x *DepositRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}
func (x *DepositRequest) GetAmountCents() int64 {
	if x != nil {
		return x.AmountCents
	}
	return 0
}
func (x *DepositRequest) GetCurrency() string {
	if x != nil {
		return x.Currency
	}
	return ""
}
func (x *DepositRequest) GetProvider() string {
	if x != nil {
		return x.Provider
	}
	return ""
}
func (x *DepositRequest) GetProviderReference() string {
	if x != nil {
		return x.ProviderReference
	}
	return ""
}

type CreateEscrowRequest struct {
	JobId        string
	ClientId     string
	FreelancerId string
	AmountCents  int64
	Currency     string
}

func (x *CreateEscrowRequest) GetJobId() string {
	if x != nil {
		return x.JobId
	}
	return ""
}
func (x *CreateEscrowRequest) GetClientId() string {
	if x != nil {
		return x.ClientId
	}
	return ""
}
func (x *CreateEscrowRequest) GetFreelancerId() string {
	if x != nil {
		return x.FreelancerId
	}
	return ""
}
func (x *CreateEscrowRequest) GetAmountCents() int64 {
	if x != nil {
		return x.AmountCents
	}
	return 0
}
func (x *CreateEscrowRequest) GetCurrency() string {
	if x != nil {
		return x.Currency
	}
	return ""
}

type ReleaseEscrowRequest struct {
	EscrowId    string
	RequesterId string
}

func (x *ReleaseEscrowRequest) GetEscrowId() string {
	if x != nil {
		return x.EscrowId
	}
	return ""
}
func (x *ReleaseEscrowRequest) GetRequesterId() string {
	if x != nil {
		return x.RequesterId
	}
	return ""
}

type RefundEscrowRequest struct {
	EscrowId    string
	RequesterId string
}

func (x *RefundEscrowRequest) GetEscrowId() string {
	if x != nil {
		return x.EscrowId
	}
	return ""
}
func (x *RefundEscrowRequest) GetRequesterId() string {
	if x != nil {
		return x.RequesterId
	}
	return ""
}

type ListTransactionsRequest struct {
	UserId string
	Limit  int32
	Offset int64
}

func (x *ListTransactionsRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}
func (x *ListTransactionsRequest) GetLimit() int32 {
	if x != nil {
		return x.Limit
	}
	return 0
}
func (x *ListTransactionsRequest) GetOffset() int64 {
	if x != nil {
		return x.Offset
	}
	return 0
}

type PaymentAccountResponse struct {
	Account *PaymentAccount
}

type TransactionResponse struct {
	Transaction *Transaction
	Account     *PaymentAccount
}

type EscrowResponse struct {
	Escrow *Escrow
}

type ListTransactionsResponse struct {
	Transactions []*Transaction
}
