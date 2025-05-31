package data

import (
	"context"

	v1 "review-service/api/review/v1"
	"review-service/internal/biz"
	"review-service/internal/data/model"
	"review-service/internal/data/query"

	"github.com/go-kratos/kratos/v2/log"
)

type reviewRepo struct {
	data *Data
	log  *log.Helper
}

func NewReviewRepo(data *Data, logger log.Logger) biz.ReviewRepo {
	return &reviewRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *reviewRepo) SaveReview(ctx context.Context, review *model.ReviewInfo) (*model.ReviewInfo, error) {
	r.log.Infof("SaveReview req %+v", review)
	err := r.data.query.ReviewInfo.WithContext(ctx).Save(review)
	if err != nil {
		r.log.Errorf("SaveReview err %+v", err)
		return nil, err
	}
	return review, err
}

func (r *reviewRepo) GetReviewByID(ctx context.Context, orderID int64) ([]*model.ReviewInfo, error) {
	r.log.Infof("GetReviewByID req %+v", orderID)
	reviews, err := r.data.query.ReviewInfo.
		WithContext(ctx).
		Where(r.data.query.ReviewInfo.OrderID.Eq(orderID)).Find()
	return reviews, err
}

func (r *reviewRepo) GetReviewByReviewID(ctx context.Context, reviewID int64) (*model.ReviewInfo, error) {
	r.log.Infof("GetReviewByReviewID req %+v", reviewID)
	review, err := r.data.query.ReviewInfo.
		WithContext(ctx).
		Where(r.data.query.ReviewInfo.ReviewID.Eq(reviewID)).First()
	return review, err
}

func (r *reviewRepo) SaveReviewReplyAndUpdateReview(ctx context.Context, reply *model.ReviewReplyInfo) (*model.ReviewReplyInfo, error) {
	r.log.Infof("SaveReviewReply req %+v", reply)
	err := r.data.query.Transaction(func(tx *query.Query) error {
		// 1. 回复表入库
		err := tx.ReviewReplyInfo.
			WithContext(ctx).
			Save(reply)
		if err != nil {
			r.log.Errorf("SaveReviewReply err %+v", err)
			return err
		}
		// 2. 评价表更新
		resp, err := tx.ReviewInfo.
			WithContext(ctx).
			Where(tx.ReviewInfo.ReviewID.Eq(reply.ReviewID)).
			Update(tx.ReviewInfo.HasReply, 1)
		if err != nil {
			r.log.Errorf("SaveReviewReply err %+v", err)
			return err
		}
		if resp.RowsAffected == 0 {
			r.log.Errorf("SaveReviewReply err %+v", err)
			return v1.ErrorReviewNotFound("评价不存在")
		}
		return nil
	})
	if err != nil {
		r.log.Errorf("SaveReviewReply err %+v", err)
		return nil, err
	}
	return reply, nil
}
