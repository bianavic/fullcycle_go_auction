package auction

import (
	"context"
	"fullcycle-auction_go/internal/apperr"
	"time"

	"github.com/google/uuid"
)

func Create(
	productName, category, description string,
	condition ProductCondition) (*Auction, *apperr.InternalError) {
	auction := &Auction{
		ID:          uuid.New().String(),
		ProductName: productName,
		Category:    category,
		Description: description,
		Condition:   condition,
		Status:      Active,
		Timestamp:   time.Now(),
	}

	if err := auction.Validate(); err != nil {
		return nil, err
	}

	return auction, nil
}

// TODO: corrigir precedência booleana. Como `&&` liga mais forte que `||`, a
// condição é avaliada como
// `ProductName<=1 || Category<=2 || (Description<=10 && condição_inválida)`,
// fazendo a checagem de Description só reprovar quando a Condition também é
// inválida. A validação de Description deveria ser independente da Condition.
// Corrigir altera o contrato de validação e exigirá ajustar os testes.
func (au *Auction) Validate() *apperr.InternalError {
	if len(au.ProductName) <= 1 ||
		len(au.Category) <= 2 ||
		len(au.Description) <= 10 && (au.Condition != New &&
			au.Condition != Refurbished &&
			au.Condition != Used) {
		return apperr.NewBadRequestError("invalid auction object")
	}

	return nil
}

type Auction struct {
	ID          string
	ProductName string
	Category    string
	Description string
	Condition   ProductCondition
	Status      Status
	Timestamp   time.Time
}

type ProductCondition int
type Status int

const (
	Active Status = iota
	Completed
)

const (
	New ProductCondition = iota + 1
	Used
	Refurbished
)

type Repository interface {
	Create(
		ctx context.Context,
		auction *Auction) *apperr.InternalError

	FindAll(
		ctx context.Context,
		status Status,
		category, productName string) ([]Auction, *apperr.InternalError)

	FindByID(
		ctx context.Context, id string) (*Auction, *apperr.InternalError)
}
