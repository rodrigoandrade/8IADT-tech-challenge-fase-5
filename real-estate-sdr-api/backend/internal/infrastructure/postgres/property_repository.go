package postgres

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	generated "real-estate-sdr/internal/infrastructure/postgres/generated"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"real-estate-sdr/internal/domain"
	"real-estate-sdr/internal/repository"
)

type PropertyRepository struct {
	queries *generated.Queries
}

func NewPropertyRepository(
	queries *generated.Queries,
) *PropertyRepository {
	return &PropertyRepository{
		queries: queries,
	}
}

func (r *PropertyRepository) Create(
	ctx context.Context,
	property *domain.Property,
) (*domain.Property, error) {
	if property == nil {
		return nil, repository.ErrInvalidProperty
	}

	result, err := r.queries.CreateProperty(
		ctx,
		generated.CreatePropertyParams{
			Title:           property.Title,
			Description:     property.Description,
			PropertyType:    string(property.Type),
			TransactionType: string(property.Transaction),
			Price:           numericFromFloat64(property.Price),
			Bedrooms:        int32(property.Bedrooms),
			Bathrooms:       int32(property.Bathrooms),
			ParkingSpaces:   int32(property.ParkingSpaces),
			Area:            numericFromFloat64(property.Area),
			City:            property.City,
			Neighborhood:    property.Neighborhood,
			Address:         property.Address,
			Available:       property.Status == domain.PropertyStatusAvailable,
		},
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repository.ErrPropertyAlreadyExists
		}

		return nil, fmt.Errorf(
			"create property: %w",
			err,
		)
	}

	return mapProperty(result)
}

func (r *PropertyRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Property, error) {
	result, err := r.queries.GetPropertyByID(
		ctx,
		id,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrPropertyNotFound
		}

		return nil, fmt.Errorf(
			"get property by id: %w",
			err,
		)
	}

	return mapProperty(result)
}

func (r *PropertyRepository) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.Property, error) {
	results, err := r.queries.ListProperties(
		ctx,
		generated.ListPropertiesParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list properties: %w",
			err,
		)
	}

	properties := make(
		[]*domain.Property,
		0,
		len(results),
	)

	for _, result := range results {
		property, err := mapProperty(result)
		if err != nil {
			return nil, err
		}

		properties = append(
			properties,
			property,
		)
	}

	return properties, nil
}

func (r *PropertyRepository) Update(
	ctx context.Context,
	property *domain.Property,
) (*domain.Property, error) {
	if property == nil {
		return nil, repository.ErrInvalidProperty
	}

	result, err := r.queries.UpdateProperty(
		ctx,
		generated.UpdatePropertyParams{
			ID:              property.ID,
			Title:           property.Title,
			Description:     property.Description,
			PropertyType:    string(property.Type),
			TransactionType: string(property.Transaction),
			Price:           numericFromFloat64(property.Price),
			Bedrooms:        int32(property.Bedrooms),
			Bathrooms:       int32(property.Bathrooms),
			ParkingSpaces:   int32(property.ParkingSpaces),
			Area:            numericFromFloat64(property.Area),
			City:            property.City,
			Neighborhood:    property.Neighborhood,
			Address:         property.Address,
			Available:       property.Status == domain.PropertyStatusAvailable,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrPropertyNotFound
		}

		if isUniqueViolation(err) {
			return nil, repository.ErrPropertyAlreadyExists
		}

		return nil, fmt.Errorf(
			"update property: %w",
			err,
		)
	}

	return mapProperty(result)
}

func (r *PropertyRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status domain.PropertyStatus,
) (*domain.Property, error) {
	available := status == domain.PropertyStatusAvailable

	result, err := r.queries.UpdatePropertyAvailability(
		ctx,
		generated.UpdatePropertyAvailabilityParams{
			ID:        id,
			Available: available,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrPropertyNotFound
		}

		return nil, fmt.Errorf(
			"update property status: %w",
			err,
		)
	}

	property, err := mapProperty(result)
	if err != nil {
		return nil, err
	}

	property.Status = status

	return property, nil
}

func (r *PropertyRepository) Count(
	ctx context.Context,
) (int64, error) {
	count, err := r.queries.CountProperties(ctx)
	if err != nil {
		return 0, fmt.Errorf(
			"count properties: %w",
			err,
		)
	}

	return count, nil
}

func mapProperty(
	result generated.Property,
) (*domain.Property, error) {
	price, err := float64FromNumeric(result.Price)
	if err != nil {
		return nil, fmt.Errorf(
			"map property price: %w",
			err,
		)
	}

	area, err := float64FromNumeric(result.Area)
	if err != nil {
		return nil, fmt.Errorf(
			"map property area: %w",
			err,
		)
	}

	status := domain.PropertyStatusInactive

	if result.Available {
		status = domain.PropertyStatusAvailable
	}

	return &domain.Property{
		ID:            result.ID,
		Title:         result.Title,
		Description:   result.Description,
		Type:          domain.PropertyType(result.PropertyType),
		Transaction:   domain.TransactionType(result.TransactionType),
		Price:         price,
		Bedrooms:      int(result.Bedrooms),
		Bathrooms:     int(result.Bathrooms),
		ParkingSpaces: int(result.ParkingSpaces),
		Area:          area,
		City:          result.City,
		Neighborhood:  result.Neighborhood,
		Address:       result.Address,
		Status:        status,
		CreatedAt:     result.CreatedAt.Time,
		UpdatedAt:     result.UpdatedAt.Time,
	}, nil
}

func numericFromFloat64(
	value float64,
) pgtype.Numeric {
	return pgtype.Numeric{
		Int:              big.NewInt(int64(value * 100)),
		Exp:              -2,
		Valid:            true,
		NaN:              false,
		InfinityModifier: pgtype.Finite,
	}
}

func float64FromNumeric(
	value pgtype.Numeric,
) (float64, error) {
	if !value.Valid {
		return 0, nil
	}

	if value.NaN {
		return 0, fmt.Errorf("numeric value is NaN")
	}

	if value.InfinityModifier != pgtype.Finite {
		return 0, fmt.Errorf("numeric value is infinite")
	}

	if value.Int == nil {
		return 0, nil
	}

	rat := new(big.Rat).SetInt(value.Int)

	if value.Exp != 0 {
		if value.Exp > 0 {
			rat.Mul(
				rat,
				new(big.Rat).SetInt(
					new(big.Int).Exp(
						big.NewInt(10),
						big.NewInt(int64(value.Exp)),
						nil,
					),
				),
			)
		} else {
			rat.Quo(
				rat,
				new(big.Rat).SetInt(
					new(big.Int).Exp(
						big.NewInt(10),
						big.NewInt(int64(-value.Exp)),
						nil,
					),
				),
			)
		}
	}

	result, err := strconv.ParseFloat(
		rat.FloatString(10),
		64,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"convert numeric to float64: %w",
			err,
		)
	}

	return result, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}
