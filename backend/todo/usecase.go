package todo

import (
	"context"
	"errors"
	"time"

	"todo/shared/transaction"
)

type Usecase interface {
	GetTodosByUserID(ctx context.Context, userID int) ([]*Todo, error)
	CreateTodo(ctx context.Context, userID int, input CreateInput) (*Todo, error)
	UpdateTodo(ctx context.Context, userID int, todoID int, input UpdateInput) (*Todo, error)
	DeleteTodo(ctx context.Context, userID int, todoID int) error
}

type CreateInput struct {
	Title   string
	Content *string
	DueDate *time.Time
}

type UpdateInput struct {
	Title       string
	Content     string
	DueDate     *time.Time
	IsCompleted bool
}

type usecase struct {
	txManager transaction.Manager
	repo      Repository
}

func NewUsecase(txManager transaction.Manager, repo Repository) Usecase {
	return &usecase{
		txManager: txManager,
		repo:      repo,
	}
}

func (u *usecase) GetTodosByUserID(ctx context.Context, userID int) ([]*Todo, error) {
	return u.repo.FindByUserID(ctx, userID)
}

func (u *usecase) CreateTodo(ctx context.Context, userID int, input CreateInput) (*Todo, error) {
	todo := &Todo{
		UserID:      userID,
		Title:       input.Title,
		Content:     input.Content,
		DueDate:     input.DueDate,
		IsCompleted: false,
	}

	err := u.txManager.Do(ctx, func(ctx context.Context) error {
		return u.repo.Create(ctx, todo)
	})
	if err != nil {
		return nil, err
	}

	return todo, nil
}

func (u *usecase) UpdateTodo(ctx context.Context, userID int, todoID int, input UpdateInput) (*Todo, error) {
	var todo *Todo

	err := u.txManager.Do(ctx, func(ctx context.Context) error {
		var err error
		todo, err = u.repo.FindByID(ctx, todoID)
		if err != nil {
			return err
		}

		if todo.UserID != userID {
			return errors.New("forbidden")
		}

		todo.Title = input.Title
		todo.Content = &input.Content
		todo.DueDate = input.DueDate
		todo.IsCompleted = input.IsCompleted

		return u.repo.Update(ctx, todo)
	})
	if err != nil {
		return nil, err
	}

	return todo, nil
}

func (u *usecase) DeleteTodo(ctx context.Context, userID int, todoID int) error {
	return u.txManager.Do(ctx, func(ctx context.Context) error {
		todo, err := u.repo.FindByID(ctx, todoID)
		if err != nil {
			return err
		}

		if todo.UserID != userID {
			return errors.New("forbidden")
		}

		return u.repo.Delete(ctx, todoID)
	})
}
