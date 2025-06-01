package biz

import (
	"context"
	"time"

	pb "review-service/api/review/v1"

	"github.com/go-kratos/kratos/v2/log"

	v1 "review-service/api/review/v1"
	"review-service/internal/common"
	"review-service/internal/data/model"
)

// GreeterRepo is a Greater repo.
type ReviewRepo interface {
	// 评价中台
	SaveReview(context.Context, *model.ReviewInfo) (*model.ReviewInfo, error)
	GetReviewByID(context.Context, int64) ([]*model.ReviewInfo, error)
	GetReviewByReviewID(ctx context.Context, reviewID int64) (*model.ReviewInfo, error)
	SaveReviewReplyAndUpdateReview(context.Context, *model.ReviewReplyInfo) (*model.ReviewReplyInfo, error)

	// b 端
	UpdateAppealReviewByReviewID(ctx context.Context, reviewID int64, appeal *model.ReviewAppealInfo) (int64, error)
	SaveAppealReviewAndUpdateReview(context.Context, *model.ReviewAppealInfo) (*model.ReviewAppealInfo, error)
}

// ReviewUsecase is a Review usecase.
type ReviewUsecase struct {
	repo ReviewRepo
	log  *log.Helper
}

// NewReviewUsecase new a Review usecase.
func NewReviewUsecase(repo ReviewRepo, logger log.Logger) *ReviewUsecase {
	return &ReviewUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// CreateReview creates a Review, and returns the new Review.
func (uc *ReviewUsecase) CreateReview(ctx context.Context, r *model.ReviewInfo) (*model.ReviewInfo, error) {
	uc.log.WithContext(ctx).Infof("CreateReview: %+v", r)
	// 1. 数据校验
	reviews, err := uc.repo.GetReviewByID(ctx, r.OrderID)
	if err != nil {
		return nil, v1.ErrorDbFailed("数据库查询失败: %v", err)
	}
	if len(reviews) > 0 {
		return nil, v1.ErrorOrderReviewed("订单已评价")
	}
	// 2. 生成 review ID
	r.ReviewID = time.Now().UnixNano()
	// 3. 查询订单和商品快照信息
	// 4. 拼装数据入库
	return uc.repo.SaveReview(ctx, r)
}

func (uc *ReviewUsecase) ReplyReview(ctx context.Context, r *pb.ReplyReviewRequest) (id int64, err error) {
	uc.log.WithContext(ctx).Infof("ReplyReview: %+v", r)
	// 1. 数据校验
	review, err := uc.repo.GetReviewByReviewID(ctx, r.GetReviewId())
	if err != nil {
		uc.log.WithContext(ctx).Errorf("ReplyReview: %+v", err)
		return 0, v1.ErrorDbFailed("数据库查询失败: %v", err)
	}
	if review == nil {
		uc.log.WithContext(ctx).Errorf("ReplyReview: %+v", err)
		return 0, v1.ErrorReviewNotFound("评价不存在: %d", r.GetReviewId())
	}
	// 1.1 水平越权，A商家不能回复B商家的评价
	if review.StoreID != r.GetStoreId() {
		uc.log.WithContext(ctx).Errorf("ReplyReview: %+v", err)
		return 0, v1.ErrorStoreNotMatch("商家不匹配: %d", r.GetStoreId())
	}
	// 1.2 已经回复过的评价不允许商家再次回复
	if review.HasReply == 1 {
		uc.log.WithContext(ctx).Errorf("ReplyReview: %+v", err)
		return 0, v1.ErrorReviewReplyExists("评价已回复: %d", r.GetReviewId())
	}
	// 2. 更新数据库中的数据（评价回复表新建，评价表同时更新）
	replyID := time.Now().UnixNano()
	reply := &model.ReviewReplyInfo{
		ReplyID:   replyID,
		ReviewID:  r.GetReviewId(),
		StoreID:   r.GetStoreId(),
		Content:   r.GetContent(),
		PicInfo:   r.GetPicInfo(),
		VideoInfo: r.GetVideoInfo(),
	}
	reviewReply, err := uc.repo.SaveReviewReplyAndUpdateReview(ctx, reply)
	if err != nil {
		return 0, v1.ErrorDbFailed("数据库保存失败: %v", err)
	}
	// 3. 返回
	uc.log.WithContext(ctx).Infof("ReplyReview: %+v", reviewReply)
	return reviewReply.ID, nil
}

func (uc *ReviewUsecase) AppealReview(ctx context.Context, r *pb.AppealReviewRequest) (id int64, err error) {
	uc.log.WithContext(ctx).Infof("AppealReview: %+v", r)
	// 1. 数据校验
	review, err := uc.repo.GetReviewByReviewID(ctx, r.GetReviewId())
	if err != nil {
		uc.log.WithContext(ctx).Errorf("AppealReview: %+v", err)
		return 0, v1.ErrorDbFailed("数据库查询失败: %v", err)
	}
	if review == nil {
		uc.log.WithContext(ctx).Errorf("AppealReview: %+v", err)
		return 0, v1.ErrorReviewNotFound("评价不存在: %d", r.GetReviewId())
	}
	// 1.1 水平越权，A商家不能申诉B商家的评价
	if review.StoreID != r.GetStoreId() {
		uc.log.WithContext(ctx).Errorf("AppealReview: %+v", err)
		return 0, v1.ErrorStoreNotMatch("商家不匹配: %d", r.GetStoreId())
	}
	// 1.2 已经申诉过的评价则更新
	appealID := time.Now().UnixNano()
	appeal := &model.ReviewAppealInfo{
		AppealID:  appealID,
		ReviewID:  r.GetReviewId(),
		StoreID:   r.GetStoreId(),
		Content:   r.GetContent(),
		PicInfo:   r.GetPicInfo(),
		VideoInfo: r.GetVideoInfo(),
		Status:    common.APPEAL_ING,
	}
	affected, err := uc.repo.UpdateAppealReviewByReviewID(ctx, r.GetReviewId(), appeal)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("AppealReview update failed: %+v", err)
		return 0, v1.ErrorDbFailed("数据库更新失败: %v", err)
	}
	if affected > 0 {
		uc.log.WithContext(ctx).Infof("AppealReview updated successfully: reviewId=%d", r.GetReviewId())
		return r.GetReviewId(), nil
	}

	// 2. 更新数据库中的数据（评价申诉表新建，评价表同时更新）
	appealReview, err := uc.repo.SaveAppealReviewAndUpdateReview(ctx, appeal)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("AppealReview save failed: %+v", err)
		return 0, v1.ErrorDbFailed("数据库保存失败: %v", err)
	}
	// 3. 返回
	uc.log.WithContext(ctx).Infof("AppealReview saved successfully: %+v", appealReview)
	return appealReview.ReviewID, nil
}
