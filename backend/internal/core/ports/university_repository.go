package ports
import (
	"context"
	domain "github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)
type IUniversityRepository interface {
	GetByDomain(ctx context.Context, domain string) (*domain.University, error)
	//GetAllUniversities() ([]domain.University, error)
}