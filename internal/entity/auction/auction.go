package auction

import (
	"context"
	"fullcycle-auction_go/internal/internal_error"
	"time"

	"github.com/google/uuid"
)

func CreateAuction(
	productName, category, description string,
	condition ProductCondition) (*Auction, *internal_error.InternalError) {
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
func (au *Auction) Validate() *internal_error.InternalError {
	if len(au.ProductName) <= 1 ||
		len(au.Category) <= 2 ||
		len(au.Description) <= 10 && (au.Condition != New &&
			au.Condition != Refurbished &&
			au.Condition != Used) {
		return internal_error.NewBadRequestError("invalid auction object")
	}

	return nil
}

type Auction struct {
	ID          string
	ProductName string
	Category    string
	Description string
	Condition   ProductCondition
	Status      AuctionStatus
	Timestamp   time.Time
}

type ProductCondition int
type AuctionStatus int

const (
	Active AuctionStatus = iota
	Completed
)

const (
	New ProductCondition = iota + 1
	Used
	Refurbished
)

type AuctionRepository interface {
	CreateAuction(
		ctx context.Context,
		auction *Auction) *internal_error.InternalError

	FindAuctions(
		ctx context.Context,
		status AuctionStatus,
		category, productName string) ([]Auction, *internal_error.InternalError)

	FindAuctionByID(
		ctx context.Context, id string) (*Auction, *internal_error.InternalError)
}
