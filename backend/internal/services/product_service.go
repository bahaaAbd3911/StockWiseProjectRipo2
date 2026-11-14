package services

import (
	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
)

type ProductService struct {
	Repository *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{
		Repository: repo,
	}
}

func (s *ProductService) GetAllProducts() ([]models.Product, error) {
	return s.Repository.GetAll()
}
func (s *ProductService) GetProductByID(id int64) (models.Product, error) {
	return s.Repository.GetByID(id)
}
func (s *ProductService) CreateProduct(product models.CreateProductRequest) (models.Product, error) {
	return s.Repository.Create(product)
}
func (s *ProductService) UpdateProduct(
	id int64,
	product models.UpdateProductRequest,
) (models.Product, error) {
	return s.Repository.Update(id, product)
}
func (s *ProductService) DeleteProduct(id int64) error {
	return s.Repository.Delete(id)
}
