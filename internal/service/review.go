package service

import (
	"context"

	pb "review-service/api/review/v1"
	"review-service/internal/biz"
	"review-service/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
)

type ReviewService struct {
	pb.UnimplementedReviewServer

	uc  *biz.ReviewUsecase
	log *log.Helper
}

func NewReviewService(uc *biz.ReviewUsecase, logger log.Logger) *ReviewService {
	return &ReviewService{uc: uc, log: log.NewHelper(logger)}
}

func (s *ReviewService) CreateReview(ctx context.Context, req *pb.CreateReviewRequest) (*pb.CreateReviewReply, error) {
	s.log.Infof("CreateReview: %+v", req)
	resp, err := s.uc.CreateReview(ctx, &model.ReviewInfo{
		UserID:       req.UserId,
		OrderID:      req.OrderId,
		StoreID:      req.StoreId,
		Score:        req.Score,
		ServiceScore: req.ServiceScore,
		ExpressScore: req.ExpressScore,
		Content:      req.Content,
		PicInfo:      req.PicInfo,
		VideoInfo:    req.VideoInfo,
		Anonymous:    req.Anonymous,
		Status:       0,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateReviewReply{
		Id: resp.ID,
	}, nil
}

func (s *ReviewService) ReplyReview(ctx context.Context, req *pb.ReplyReviewRequest) (*pb.ReplyReviewResp, error) {
	s.log.Infof("ReplyReview function called: %+v", req)
	id, err := s.uc.ReplyReview(ctx, req)
	if err != nil {
		return nil, err
	}
	s.log.Infof("ReplyReview success: %+v", id)
	return &pb.ReplyReviewResp{Id: id}, nil
}

func (s *ReviewService) AppealReview(ctx context.Context, req *pb.AppealReviewRequest) (*pb.AppealReviewReply, error) {
	s.log.Infof("AppealReview function called: %+v", req)
	id, err := s.uc.AppealReview(ctx, req)
	if err != nil {
		return nil, err
	}
	s.log.Infof("AppealReview success: %+v", id)
	return &pb.AppealReviewReply{Id: id}, nil
}

func (s *ReviewService) OperationAppealReview(ctx context.Context, req *pb.OperationAppealReviewRequest) (*pb.OperationAppealReviewReply, error) {
	s.log.Infof("OperationAppealReview function called: %+v", req)
	id, err := s.uc.OperationAppealReview(ctx, req)
	if err != nil {
		s.log.Errorf("OperationAppealReview failed: %+v", err)
		return nil, err
	}
	s.log.Infof("OperationAppealReview success: %+v", id)
	return &pb.OperationAppealReviewReply{Id: id}, nil
}
