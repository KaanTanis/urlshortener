package service

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/kaantanis/urlshortener/internal/model"
	"github.com/kaantanis/urlshortener/internal/repository"
	"github.com/matoous/go-nanoid/v2"
)

var (
	ErrInvalidURL  = errors.New("url must be a valid absolute http or https URL")
	ErrURLTooLong  = errors.New("url exceeds max length of 2048")
	ErrCodeMissing = errors.New("code is required")
)

type URLService struct {
	urlRepo  *repository.URLRepository
	logRepo  *repository.VisitLogRepository
	baseURL  string
	codeSize int
}

func NewURLService(
	urlRepo *repository.URLRepository,
	logRepo *repository.VisitLogRepository,
	baseURL string,
	codeSize int,
) *URLService {
	return &URLService{
		urlRepo:  urlRepo,
		logRepo:  logRepo,
		baseURL:  strings.TrimRight(baseURL, "/"),
		codeSize: codeSize,
	}
}

func (s *URLService) ValidateOriginalURL(input string) error {
	if len(input) == 0 {
		return ErrInvalidURL
	}
	if len(input) > 2048 {
		return ErrURLTooLong
	}

	parsed, err := url.ParseRequestURI(input)
	if err != nil || parsed == nil {
		return ErrInvalidURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidURL
	}
	if parsed.Host == "" {
		return ErrInvalidURL
	}

	return nil
}

func (s *URLService) CreateShortURL(original string) (model.URL, string, error) {
	if err := s.ValidateOriginalURL(original); err != nil {
		return model.URL{}, "", err
	}

	const attemptsPerLength = 10
	const maxCodeLength = 32

	currentLength := s.codeSize
	for currentLength <= maxCodeLength {
		for range attemptsPerLength {
			code, err := gonanoid.New(currentLength)
			if err != nil {
				return model.URL{}, "", fmt.Errorf("generate nanoid: %w", err)
			}

			created, err := s.urlRepo.Create(code, original)
			if err != nil {
				if errors.Is(err, repository.ErrCodeExists) {
					continue
				}
				return model.URL{}, "", err
			}

			if currentLength != s.codeSize {
				slog.Info("code length increased due to collisions", "new_length", currentLength)
			}
			return created, s.baseURL + "/" + created.Code, nil
		}

		currentLength++
	}

	return model.URL{}, "", errors.New("failed to generate unique code after expanding length")
}

func (s *URLService) ResolveByCode(code string) (model.URL, error) {
	if strings.TrimSpace(code) == "" {
		return model.URL{}, ErrCodeMissing
	}
	return s.urlRepo.FindByCode(code)
}

func (s *URLService) IncrementHitCount(code string) error {
	return s.urlRepo.IncrementHitCount(code)
}

func (s *URLService) LogVisitAsync(entry model.VisitLog) {
	go func() {
		if err := s.logRepo.Create(entry); err != nil {
			slog.Error("async visit log insert failed", "code", entry.Code, "error", err)
		}
	}()
}

func (s *URLService) GetStats(code string) (model.URL, []model.VisitLog, error) {
	urlRow, err := s.ResolveByCode(code)
	if err != nil {
		return model.URL{}, nil, err
	}

	recent, err := s.logRepo.FindRecentByCode(code, 20)
	if err != nil {
		return model.URL{}, nil, err
	}

	return urlRow, recent, nil
}

func IsNotFoundError(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
