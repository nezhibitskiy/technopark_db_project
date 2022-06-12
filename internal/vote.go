package internal

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

func (s *Service) FindUserVote(data Vote) (int, error) {
	vote := 0
	err := s.dbPool.QueryRow(context.Background(), "SELECT id FROM votes WHERE author = $1",
		&data.Nickname).Scan(&vote)
	if err != nil {
		return 0, err
	}
	return vote, nil
}

func (s *Service) ThreadVote() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		data := Vote{}

		slug := ctx.Param("slug_or_id")
		err := ctx.Bind(&data)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, err)
		}
		user, err := s.searchUsersByNickname(data.Nickname)
		if err != nil && err != pgx.ErrNoRows {
			return ctx.JSON(http.StatusInternalServerError, err)
		}
		if user == nil {
			return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find user with nickname: " + data.Nickname})
		}

		var thread *Thread
		id, err := strconv.Atoi(slug)
		if err != nil {
			thread, err = s.GetThreadByIDorSlug(slug, 0)
			if err != nil && err != pgx.ErrNoRows {
				return ctx.JSON(http.StatusInternalServerError, err)
			}
			if thread == nil {
				return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find thread by slug: " + slug})
			}
		} else {
			thread, err = s.GetThreadByIDorSlug("", uint(id))
			if err != nil && err != pgx.ErrNoRows {
				return ctx.JSON(http.StatusInternalServerError, err)
			}
			if thread == nil {
				return ctx.JSON(http.StatusNotFound, ResponseError{Message: "Can't find thread by id: " + slug})
			}

		}

		voteID, err := s.FindUserVote(data)
		if err != nil && err != pgx.ErrNoRows {
			return ctx.JSON(http.StatusInternalServerError, err)
		}
		if err == pgx.ErrNoRows {
			_, err = s.dbPool.Exec(context.Background(), "INSERT INTO votes(thread_id, author, value) VALUES($1, $2, $3);",
				&thread.Id, &data.Nickname, &data.Voice)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, err)
			}
		} else {
			_, err = s.dbPool.Exec(context.Background(), "UPDATE votes SET value = $1 WHERE id = $2;",
				&data.Voice, &voteID)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, err)
			}
		}

		err = s.FillThreadVotes(thread)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, err)
		}

		return ctx.JSON(http.StatusOK, &thread)
	}
}
