package application

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"real-estate-sdr/internal/domain"
	"real-estate-sdr/internal/repository"
)

type fakePropertyRepository struct {
	createFunc       func(context.Context, *domain.Property) (*domain.Property, error)
	getByIDFunc      func(context.Context, uuid.UUID) (*domain.Property, error)
	listFunc         func(context.Context, int, int) ([]*domain.Property, error)
	updateFunc       func(context.Context, *domain.Property) (*domain.Property, error)
	updateStatusFunc func(context.Context, uuid.UUID, domain.PropertyStatus) (*domain.Property, error)
	countFunc        func(context.Context) (int64, error)
}

func (f *fakePropertyRepository) Create(
	ctx context.Context,
	property *domain.Property,
) (*domain.Property, error) {
	if f.createFunc != nil {
		return f.createFunc(ctx, property)
	}

	return property, nil
}

func (f *fakePropertyRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Property, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(ctx, id)
	}

	return nil, nil
}

func (f *fakePropertyRepository) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.Property, error) {
	if f.listFunc != nil {
		return f.listFunc(ctx, limit, offset)
	}

	return nil, nil
}

func (f *fakePropertyRepository) Update(
	ctx context.Context,
	property *domain.Property,
) (*domain.Property, error) {
	if f.updateFunc != nil {
		return f.updateFunc(ctx, property)
	}

	return property, nil
}

func (f *fakePropertyRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status domain.PropertyStatus,
) (*domain.Property, error) {
	if f.updateStatusFunc != nil {
		return f.updateStatusFunc(ctx, id, status)
	}

	return nil, nil
}

func (f *fakePropertyRepository) Count(
	ctx context.Context,
) (int64, error) {
	if f.countFunc != nil {
		return f.countFunc(ctx)
	}

	return 0, nil
}

func validProperty() *domain.Property {
	return &domain.Property{
		ID:            uuid.New(),
		Title:         "Apartamento Atualizado",
		Description:   "Descrição atualizada",
		Type:          domain.PropertyTypeApartment,
		Transaction:   domain.TransactionTypeSale,
		Price:         500000,
		Bedrooms:      3,
		Bathrooms:     2,
		ParkingSpaces: 2,
		Area:          130,
		Address:       "Rua Atualizada, 100",
		Neighborhood:  "Centro",
		City:          "São Paulo",
		Status:        domain.PropertyStatusAvailable,
		State:         "SP",
	}
}

func TestPropertyService_Create(t *testing.T) {
	ctx := context.Background()

	property := validProperty()

	repositoryMock := &fakePropertyRepository{
		createFunc: func(
			ctx context.Context,
			property *domain.Property,
		) (*domain.Property, error) {
			return property, nil
		},
	}

	service := NewPropertyService(repositoryMock)

	createdProperty, err := service.Create(
		ctx,
		property,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if createdProperty == nil {
		t.Fatal("expected property, got nil")
	}

	if createdProperty.Title != property.Title {
		t.Errorf(
			"expected title %q, got %q",
			property.Title,
			createdProperty.Title,
		)
	}
}

func TestPropertyService_Create_InvalidProperty(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		property *domain.Property
	}{
		{
			name:     "nil property",
			property: nil,
		},
		{
			name: "missing title",
			property: func() *domain.Property {
				property := validProperty()
				property.Title = ""
				return property
			}(),
		},
		{
			name: "missing type",
			property: func() *domain.Property {
				property := validProperty()
				property.Type = ""
				return property
			}(),
		},
		{
			name: "invalid type",
			property: func() *domain.Property {
				property := validProperty()
				property.Type = "INVALID"
				return property
			}(),
		},
		{
			name: "missing transaction type",
			property: func() *domain.Property {
				property := validProperty()
				property.Transaction = ""
				return property
			}(),
		},
		{
			name: "invalid transaction type",
			property: func() *domain.Property {
				property := validProperty()
				property.Transaction = "INVALID"
				return property
			}(),
		},
		{
			name: "invalid price",
			property: func() *domain.Property {
				property := validProperty()
				property.Price = 0
				return property
			}(),
		},
		{
			name: "invalid area",
			property: func() *domain.Property {
				property := validProperty()
				property.Area = 0
				return property
			}(),
		},
		{
			name: "negative bedrooms",
			property: func() *domain.Property {
				property := validProperty()
				property.Bedrooms = -1
				return property
			}(),
		},
		{
			name: "negative bathrooms",
			property: func() *domain.Property {
				property := validProperty()
				property.Bathrooms = -1
				return property
			}(),
		},
		{
			name: "negative parking spaces",
			property: func() *domain.Property {
				property := validProperty()
				property.ParkingSpaces = -1
				return property
			}(),
		},
		{
			name: "missing city",
			property: func() *domain.Property {
				property := validProperty()
				property.City = ""
				return property
			}(),
		},
		//{
		//	name: "missing state",
		//	property: func() *domain.Property {
		//		property := validProperty()
		//		property.State = ""
		//		return property
		//	}(),
		//},
		{
			name: "invalid status",
			property: func() *domain.Property {
				property := validProperty()
				property.Status = "INVALID"
				return property
			}(),
		},
	}

	service := NewPropertyService(
		&fakePropertyRepository{},
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.Create(
				ctx,
				test.property,
			)

			if err == nil {
				t.Fatal(
					"expected validation error, got nil",
				)
			}

			if !errors.Is(
				err,
				domain.ErrInvalidProperty,
			) &&
				!errors.Is(
					err,
					domain.ErrInvalidPropertyType,
				) &&
				!errors.Is(
					err,
					domain.ErrInvalidTransactionType,
				) &&
				!errors.Is(
					err,
					domain.ErrInvalidPropertyStatus,
				) {
				t.Errorf(
					"expected property validation error, got %v",
					err,
				)
			}
		})
	}
}

func TestPropertyService_Create_RepositoryError(t *testing.T) {
	ctx := context.Background()

	repositoryError := errors.New("database error")

	repositoryMock := &fakePropertyRepository{
		createFunc: func(
			ctx context.Context,
			property *domain.Property,
		) (*domain.Property, error) {
			return nil, repositoryError
		},
	}

	service := NewPropertyService(repositoryMock)

	_, err := service.Create(
		ctx,
		validProperty(),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, repositoryError) {
		t.Errorf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestPropertyService_GetByID(t *testing.T) {
	ctx := context.Background()

	property := validProperty()

	repositoryMock := &fakePropertyRepository{
		getByIDFunc: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Property, error) {
			return property, nil
		},
	}

	service := NewPropertyService(repositoryMock)

	result, err := service.GetByID(
		ctx,
		property.ID,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result == nil {
		t.Fatal("expected property, got nil")
	}

	if result.ID != property.ID {
		t.Errorf(
			"expected id %s, got %s",
			property.ID,
			result.ID,
		)
	}
}

func TestPropertyService_GetByID_InvalidID(t *testing.T) {
	ctx := context.Background()

	service := NewPropertyService(
		&fakePropertyRepository{},
	)

	_, err := service.GetByID(
		ctx,
		uuid.Nil,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(
		err,
		domain.ErrInvalidProperty,
	) {
		t.Errorf(
			"expected ErrInvalidProperty, got %v",
			err,
		)
	}
}

func TestPropertyService_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()

	repositoryMock := &fakePropertyRepository{
		getByIDFunc: func(
			ctx context.Context,
			id uuid.UUID,
		) (*domain.Property, error) {
			return nil, repository.ErrPropertyNotFound
		},
	}

	service := NewPropertyService(repositoryMock)

	_, err := service.GetByID(
		ctx,
		uuid.New(),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(
		err,
		repository.ErrPropertyNotFound,
	) {
		t.Errorf(
			"expected ErrPropertyNotFound, got %v",
			err,
		)
	}
}

func TestPropertyService_List(t *testing.T) {
	ctx := context.Background()

	properties := []*domain.Property{
		validProperty(),
		validProperty(),
	}

	var receivedLimit int
	var receivedOffset int

	repositoryMock := &fakePropertyRepository{
		listFunc: func(
			ctx context.Context,
			limit int,
			offset int,
		) ([]*domain.Property, error) {
			receivedLimit = limit
			receivedOffset = offset

			return properties, nil
		},
	}

	service := NewPropertyService(repositoryMock)

	result, err := service.List(
		ctx,
		50,
		10,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(result) != 2 {
		t.Fatalf(
			"expected 2 properties, got %d",
			len(result),
		)
	}

	if receivedLimit != 50 {
		t.Errorf(
			"expected limit 50, got %d",
			receivedLimit,
		)
	}

	if receivedOffset != 10 {
		t.Errorf(
			"expected offset 10, got %d",
			receivedOffset,
		)
	}
}

func TestPropertyService_List_DefaultLimit(t *testing.T) {
	ctx := context.Background()

	var receivedLimit int
	var receivedOffset int

	repositoryMock := &fakePropertyRepository{
		listFunc: func(
			ctx context.Context,
			limit int,
			offset int,
		) ([]*domain.Property, error) {
			receivedLimit = limit
			receivedOffset = offset

			return []*domain.Property{}, nil
		},
	}

	service := NewPropertyService(repositoryMock)

	_, err := service.List(
		ctx,
		0,
		-10,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if receivedLimit != 20 {
		t.Errorf(
			"expected default limit 20, got %d",
			receivedLimit,
		)
	}

	if receivedOffset != 0 {
		t.Errorf(
			"expected offset 0, got %d",
			receivedOffset,
		)
	}
}

func TestPropertyService_List_MaxLimit(t *testing.T) {
	ctx := context.Background()

	var receivedLimit int

	repositoryMock := &fakePropertyRepository{
		listFunc: func(
			ctx context.Context,
			limit int,
			offset int,
		) ([]*domain.Property, error) {
			receivedLimit = limit

			return []*domain.Property{}, nil
		},
	}

	service := NewPropertyService(repositoryMock)

	_, err := service.List(
		ctx,
		500,
		0,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if receivedLimit != 100 {
		t.Errorf(
			"expected max limit 100, got %d",
			receivedLimit,
		)
	}
}

func TestPropertyService_List_RepositoryError(t *testing.T) {
	ctx := context.Background()

	repositoryError := errors.New("database error")

	repositoryMock := &fakePropertyRepository{
		listFunc: func(
			ctx context.Context,
			limit int,
			offset int,
		) ([]*domain.Property, error) {
			return nil, repositoryError
		},
	}

	service := NewPropertyService(repositoryMock)

	_, err := service.List(
		ctx,
		20,
		0,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, repositoryError) {
		t.Errorf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestPropertyService_Update(t *testing.T) {
	ctx := context.Background()

	property := validProperty()

	repositoryMock := &fakePropertyRepository{
		updateFunc: func(
			ctx context.Context,
			property *domain.Property,
		) (*domain.Property, error) {
			return property, nil
		},
	}

	service := NewPropertyService(repositoryMock)

	result, err := service.Update(
		ctx,
		property,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result == nil {
		t.Fatal("expected property, got nil")
	}

	if result.ID != property.ID {
		t.Errorf(
			"expected id %s, got %s",
			property.ID,
			result.ID,
		)
	}
}

func TestPropertyService_Update_InvalidID(t *testing.T) {
	ctx := context.Background()

	property := validProperty()
	property.ID = uuid.Nil

	service := NewPropertyService(
		&fakePropertyRepository{},
	)

	_, err := service.Update(
		ctx,
		property,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(
		err,
		domain.ErrInvalidProperty,
	) {
		t.Errorf(
			"expected ErrInvalidProperty, got %v",
			err,
		)
	}
}

func TestPropertyService_Update_InvalidProperty(t *testing.T) {
	ctx := context.Background()

	property := validProperty()
	property.Title = ""

	service := NewPropertyService(
		&fakePropertyRepository{},
	)

	_, err := service.Update(
		ctx,
		property,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(
		err,
		domain.ErrInvalidProperty,
	) {
		t.Errorf(
			"expected ErrInvalidProperty, got %v",
			err,
		)
	}
}

func TestPropertyService_Update_NotFound(t *testing.T) {
	ctx := context.Background()

	repositoryMock := &fakePropertyRepository{
		updateFunc: func(
			ctx context.Context,
			property *domain.Property,
		) (*domain.Property, error) {
			return nil, repository.ErrPropertyNotFound
		},
	}

	service := NewPropertyService(repositoryMock)

	_, err := service.Update(
		ctx,
		validProperty(),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(
		err,
		repository.ErrPropertyNotFound,
	) {
		t.Errorf(
			"expected ErrPropertyNotFound, got %v",
			err,
		)
	}
}

func TestPropertyService_UpdateStatus(t *testing.T) {
	ctx := context.Background()

	property := validProperty()

	repositoryMock := &fakePropertyRepository{
		updateStatusFunc: func(
			ctx context.Context,
			id uuid.UUID,
			status domain.PropertyStatus,
		) (*domain.Property, error) {
			property.Status = status
			return property, nil
		},
	}

	service := NewPropertyService(repositoryMock)

	result, err := service.UpdateStatus(
		ctx,
		property.ID,
		domain.PropertyStatusSold,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result == nil {
		t.Fatal("expected property, got nil")
	}

	if result.Status != domain.PropertyStatusSold {
		t.Errorf(
			"expected status %q, got %q",
			domain.PropertyStatusSold,
			result.Status,
		)
	}
}

func TestPropertyService_UpdateStatus_InvalidInput(t *testing.T) {
	ctx := context.Background()

	service := NewPropertyService(
		&fakePropertyRepository{},
	)

	tests := []struct {
		name   string
		id     uuid.UUID
		status domain.PropertyStatus
	}{
		{
			name:   "invalid id",
			id:     uuid.Nil,
			status: domain.PropertyStatusAvailable,
		},
		{
			name:   "invalid status",
			id:     uuid.New(),
			status: "INVALID",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.UpdateStatus(
				ctx,
				test.id,
				test.status,
			)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !errors.Is(
				err,
				domain.ErrInvalidProperty,
			) &&
				!errors.Is(
					err,
					domain.ErrInvalidPropertyStatus,
				) {
				t.Errorf(
					"unexpected error: %v",
					err,
				)
			}
		})
	}
}

func TestPropertyService_UpdateStatus_NotFound(t *testing.T) {
	ctx := context.Background()

	repositoryMock := &fakePropertyRepository{
		updateStatusFunc: func(
			ctx context.Context,
			id uuid.UUID,
			status domain.PropertyStatus,
		) (*domain.Property, error) {
			return nil, repository.ErrPropertyNotFound
		},
	}

	service := NewPropertyService(repositoryMock)

	_, err := service.UpdateStatus(
		ctx,
		uuid.New(),
		domain.PropertyStatusSold,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(
		err,
		repository.ErrPropertyNotFound,
	) {
		t.Errorf(
			"expected ErrPropertyNotFound, got %v",
			err,
		)
	}
}

func TestPropertyService_Count(t *testing.T) {
	ctx := context.Background()

	repositoryMock := &fakePropertyRepository{
		countFunc: func(
			ctx context.Context,
		) (int64, error) {
			return 15, nil
		},
	}

	service := NewPropertyService(repositoryMock)

	count, err := service.Count(ctx)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if count != 15 {
		t.Errorf(
			"expected count 15, got %d",
			count,
		)
	}
}

func TestPropertyService_Count_RepositoryError(t *testing.T) {
	ctx := context.Background()

	repositoryError := errors.New("database error")

	repositoryMock := &fakePropertyRepository{
		countFunc: func(
			ctx context.Context,
		) (int64, error) {
			return 0, repositoryError
		},
	}

	service := NewPropertyService(repositoryMock)

	_, err := service.Count(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, repositoryError) {
		t.Errorf(
			"expected repository error, got %v",
			err,
		)
	}
}
