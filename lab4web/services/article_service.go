package services

import (
	"goapp/models"
	"goapp/repository"

	"gorm.io/gorm"
)

// ArticleService обрабатывает логику CRUD для статей.
type ArticleService struct {
	repo *repository.ArticleRepo
}

func NewArticleService(db *gorm.DB) *ArticleService {
	return &ArticleService{
		repo: repository.NewArticleRepo(db),
	}
}

// Создание статьи от лица пользователя userID
func (s *ArticleService) CreateArticle(input *models.Article) (*models.Article, error) {
	if err := s.repo.Create(input); err != nil {
		return nil, err
	}
	return input, nil
}

// Получение статьи по ID
func (s *ArticleService) GetArticleByID(id string) (*models.Article, error) {
	return s.repo.FindByID(id)
}

// Получение списка статей пользователя
func (s *ArticleService) GetArticlesByUser(userID string) ([]models.Article, error) {
	list, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// Редактирование статьи (проверка владения — на уровне контроллера)
func (s *ArticleService) UpdateArticle(article *models.Article, data map[string]interface{}) (*models.Article, error) {
	// Заполняем поля для обновления
	if title, ok := data["title"].(string); ok {
		article.Title = title
	}
	if content, ok := data["content"].(string); ok {
		article.Content = content
	}
	if status, ok := data["status"].(string); ok {
		article.Status = status
	}
	if err := s.repo.Update(article); err != nil {
		return nil, err
	}
	return article, nil
}

// Удаление статьи (soft delete)
func (s *ArticleService) DeleteArticle(article *models.Article) error {
	return s.repo.Delete(article)
}
