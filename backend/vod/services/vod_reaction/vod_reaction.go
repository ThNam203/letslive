package vodreaction

import (
	"context"
	"fmt"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"
	"sen1or/letslive/vod/dto"
	"sen1or/letslive/vod/utils"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VODReactionService struct {
	reactionRepo domains.VODReactionRepository
	vodRepo      domains.VODRepository
	dbPool       *pgxpool.Pool
}

func NewVODReactionService(reactionRepo domains.VODReactionRepository, vodRepo domains.VODRepository, dbPool *pgxpool.Pool) *VODReactionService {
	return &VODReactionService{
		reactionRepo: reactionRepo,
		vodRepo:      vodRepo,
		dbPool:       dbPool,
	}
}

// getReactableVOD loads the VOD, hiding private ones from everyone but the owner.
func (s *VODReactionService) getReactableVOD(ctx context.Context, vodId uuid.UUID, userId uuid.UUID) (*domains.VOD, error) {
	vod, err := s.vodRepo.GetById(ctx, vodId)
	if err != nil {
		return nil, err
	}
	if vod.Visibility != domains.VODPublicVisibility && vod.UserId != userId {
		return nil, domains.ErrVODNotFound
	}
	return vod, nil
}

func (s *VODReactionService) GetMyReaction(ctx context.Context, vodId uuid.UUID, userId uuid.UUID) (*dto.VODReactionResponseDTO, error) {
	vod, err := s.getReactableVOD(ctx, vodId, userId)
	if err != nil {
		return nil, err
	}

	reaction, err := s.reactionRepo.GetReaction(ctx, vodId, userId)
	if err != nil {
		return nil, err
	}

	return &dto.VODReactionResponseDTO{LikeCount: vod.LikeCount, Reaction: reaction}, nil
}

// SetReaction likes or dislikes a VOD, switching an existing opposite reaction.
// Setting the reaction the user already has is a no-op.
func (s *VODReactionService) SetReaction(ctx context.Context, data dto.SetVODReactionRequestDTO, vodId uuid.UUID, userId uuid.UUID) (*dto.VODReactionResponseDTO, error) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, fmt.Errorf("%w: %w", domains.ErrInvalidInput, err)
	}
	reaction := domains.VODReactionType(data.Reaction)

	if _, err := s.getReactableVOD(ctx, vodId, userId); err != nil {
		return nil, err
	}

	tx, txErr := s.dbPool.Begin(ctx)
	if txErr != nil {
		logger.Errorf(ctx, "failed to begin tx [setvodreaction: %v]", txErr)
		return nil, domains.ErrDatabaseIssue
	}
	defer tx.Rollback(ctx)

	txRepo := s.reactionRepo.WithTx(tx)

	var previous *domains.VODReactionType
	inserted := false
	// a concurrent remove can delete the row between the insert and the
	// select, so retry once; each statement sees the latest committed state
	for attempt := 0; attempt < 2 && !inserted && previous == nil; attempt++ {
		var err error
		inserted, err = txRepo.InsertReaction(ctx, vodId, userId, reaction)
		if err != nil {
			return nil, err
		}
		if inserted {
			break
		}

		previous, err = txRepo.GetReactionForUpdate(ctx, vodId, userId)
		if err != nil {
			return nil, err
		}
	}
	if !inserted && previous == nil {
		logger.Errorf(ctx, "vod reaction row kept disappearing [setvodreaction vod=%s user=%s]", vodId, userId)
		return nil, domains.ErrDatabaseIssue
	}

	likeDelta, dislikeDelta := reactionDelta(reaction, 1)
	if previous != nil {
		if *previous == reaction {
			likeDelta, dislikeDelta = 0, 0
		} else {
			if err := txRepo.UpdateReaction(ctx, vodId, userId, reaction); err != nil {
				return nil, err
			}
			prevLike, prevDislike := reactionDelta(*previous, -1)
			likeDelta += prevLike
			dislikeDelta += prevDislike
		}
	}

	likeCount, err := txRepo.AdjustCounts(ctx, vodId, likeDelta, dislikeDelta)
	if err != nil {
		return nil, err
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		logger.Errorf(ctx, "failed to commit tx [setvodreaction: %v]", commitErr)
		return nil, domains.ErrDatabaseIssue
	}

	return &dto.VODReactionResponseDTO{LikeCount: likeCount, Reaction: &reaction}, nil
}

// RemoveReaction clears the user's like or dislike. Removing nothing is a no-op.
func (s *VODReactionService) RemoveReaction(ctx context.Context, vodId uuid.UUID, userId uuid.UUID) (*dto.VODReactionResponseDTO, error) {
	if _, err := s.getReactableVOD(ctx, vodId, userId); err != nil {
		return nil, err
	}

	tx, txErr := s.dbPool.Begin(ctx)
	if txErr != nil {
		logger.Errorf(ctx, "failed to begin tx [removevodreaction: %v]", txErr)
		return nil, domains.ErrDatabaseIssue
	}
	defer tx.Rollback(ctx)

	txRepo := s.reactionRepo.WithTx(tx)

	removed, err := txRepo.DeleteReaction(ctx, vodId, userId)
	if err != nil {
		return nil, err
	}

	var likeDelta, dislikeDelta int64
	if removed != nil {
		likeDelta, dislikeDelta = reactionDelta(*removed, -1)
	}

	likeCount, err := txRepo.AdjustCounts(ctx, vodId, likeDelta, dislikeDelta)
	if err != nil {
		return nil, err
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		logger.Errorf(ctx, "failed to commit tx [removevodreaction: %v]", commitErr)
		return nil, domains.ErrDatabaseIssue
	}

	return &dto.VODReactionResponseDTO{LikeCount: likeCount, Reaction: nil}, nil
}

func reactionDelta(reaction domains.VODReactionType, step int64) (likeDelta int64, dislikeDelta int64) {
	if reaction == domains.VODReactionLike {
		return step, 0
	}
	return 0, step
}
